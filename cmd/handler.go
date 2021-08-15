package cmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/httputil"
	"github.com/openfaas/faas-provider/types"
	pb "github.com/openfaas/faasd/proto/agent"
	"google.golang.org/grpc"
)

type BaseURLResolver interface {
	Resolve(functionName string) (url.URL, error)
}

const (
	//address     = "localhost:50051"
	//defaultName = "world"
	defaultContentType    = "text/plain"
	MaxCacheItem          = 2
	MaxAgentFunctionCache = 5
	MaxClientLoad         = 6
	UseCache              = false
	UseLoadBalancerCache  = true
	batchTime             = 100
)

type Agent struct {
	Id      uint
	Address string
	Loads   uint
}

type CacheChecking struct {
	sReqHash string
	message  string
	response []byte
}

var ageantAddresses []Agent
var ageantLoad []uint

//var Cache *cache.Cache
var Cache *lru.Cache
var CacheAgent *lru.Cache
var mutex sync.Mutex
var mutexAgent sync.Mutex
var cacheHit uint
var cacheMiss uint
var loadMiss uint64
var hashRequests = make(chan string, 100)
var hashRequestsResult = make(chan CacheChecking, 100)

func initHandler() {
	cacheHit = 0
	cacheMiss = 0
	loadMiss = 0

	Cache = lru.New(MaxCacheItem)
	CacheAgent = lru.New(MaxAgentFunctionCache)
	ageantAddresses = append(ageantAddresses, Agent{Id: 0, Address: "localhost:50061"})
	ageantAddresses = append(ageantAddresses, Agent{Id: 1, Address: "localhost:50062"})
	ageantAddresses = append(ageantAddresses, Agent{Id: 2, Address: "localhost:50063"})
	ageantAddresses = append(ageantAddresses, Agent{Id: 3, Address: "localhost:50064"})
}

func NewHandlerFunc(config types.FaaSConfig, resolver BaseURLResolver) http.HandlerFunc {
	log.Println("Mohammad NewHandlerFunc")
	if resolver == nil {
		panic("NewHandlerFunc: empty proxy handler resolver, cannot be nil")
	}

	//proxyClient := proxy.NewProxyClientFromConfig(config)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet:

			pathVars := mux.Vars(r)
			functionName := pathVars["name"]
			if functionName == "" {
				httputil.Errorf(w, http.StatusBadRequest, "missing function name")
				return
			}

			exteraPath := pathVars["params"]

			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("read request bodey error :", err.Error())
			}
			log.Println("Mohammad RequestURI: ", r.RequestURI, ", inputs:", string(bodyBytes))
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			sReqHash := hash(append([]byte(functionName), bodyBytes...))

			//*********** cache  ******************
			if UseCache {
				mutex.Lock()
				response, found := Cache.Get(sReqHash)
				mutex.Unlock()
				if found {
					log.Println("Mohammad founded in cache  functionName: ", functionName)
					res, err := unserializeReq(response.([]byte), r)
					if err != nil {
						log.Println("Mohammad unserialize res: ", err.Error())
						httputil.Errorf(w, http.StatusInternalServerError, "Can't unserialize res: %s.", functionName)
						return
					}

					clientHeader := w.Header()
					copyHeaders(clientHeader, &res.Header)
					w.Header().Set("Content-Type", getContentType(r.Header, res.Header))

					w.WriteHeader(res.StatusCode)
					io.Copy(w, res.Body)
					return
				}
			}

			sReq, err := captureRequestData(r)
			if err != nil {
				httputil.Errorf(w, http.StatusInternalServerError, "Can't captureRequestData for: %s.", functionName)
				return
			}

			//proxy.ProxyRequest(w, r, proxyClient, resolver)
			agentRes, err := loadBalancer(functionName, exteraPath, sReq, sReqHash)
			if err != nil {
				httputil.Errorf(w, http.StatusInternalServerError, "Can't reach service for: %s.", functionName)
				return
			}

			//log.Println("Mohammad add to cache sReqHash:", sReqHash)
			if UseCache {
				mutex.Lock()
				Cache.Add(sReqHash, agentRes.Response)
				mutex.Unlock()
			}

			res, err := unserializeReq(agentRes.Response, r)
			if err != nil {
				log.Println("Mohammad unserialize res: ", err.Error())
				httputil.Errorf(w, http.StatusInternalServerError, "Can't unserialize res: %s.", functionName)
				return
			}

			clientHeader := w.Header()
			copyHeaders(clientHeader, &res.Header)
			w.Header().Set("Content-Type", getContentType(r.Header, res.Header))

			w.WriteHeader(res.StatusCode)
			io.Copy(w, res.Body)

			//w.WriteHeader(http.StatusOK)
			//_, _ =w.Write(agentRes.Response)
			//io.Copy(w, r.Response)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
	//return proxy.NewHandlerFunc(config, resolver)
}

func loadBalancer(RequestURI string, exteraPath string, sReq []byte, sReqHash string) (*pb.TaskResponse, error) {
	hashRequests <- sReqHash
	stopWaitting := false
	for {
		select {
		case res := <-hashRequestsResult:
			if res.sReqHash == sReqHash {
				if len(res.response) > 0 {
					return &pb.TaskResponse{Message: "OK", Response: res.response}, nil
				}
				stopWaitting = true
			}
		default:
		}
		if stopWaitting {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	var agentId uint32
	if UseLoadBalancerCache {
		mutexAgent.Lock()
		value, found := CacheAgent.Get(RequestURI)
		mutexAgent.Unlock()
		if found {
			agentId = value.(uint32)
			if ageantAddresses[agentId].Loads < MaxClientLoad {
				mutexAgent.Lock()
				ageantAddresses[agentId].Loads++
				cacheHit++
				mutexAgent.Unlock()
				log.Printf("sendToAgent due to Cache cacheHit: %v, address: %v,  RequestURI :%s", cacheHit, ageantAddresses[agentId].Address, RequestURI)
				return sendToAgent(ageantAddresses[agentId].Address, RequestURI, exteraPath, sReq, agentId)
			}
			atomic.AddUint64(&loadMiss, 1)
		}
	}

	mutexAgent.Lock()
	for i := 0; i < len(ageantAddresses); i++ {
		agentId = uint32(rand.Int31n(int32(len(ageantAddresses))))
		if ageantAddresses[agentId].Loads < MaxClientLoad {
			break
		}
	}
	if UseLoadBalancerCache {
		CacheAgent.Add(RequestURI, agentId)
		cacheMiss++
	}
	ageantAddresses[agentId].Loads++
	log.Printf("sendToAgent loadMiss: %v, cacheMiss: %v, address: %v,  RequestURI :%s", loadMiss, cacheMiss, ageantAddresses[agentId].Address, RequestURI)
	mutexAgent.Unlock()
	return sendToAgent(ageantAddresses[agentId].Address, RequestURI, exteraPath, sReq, agentId)

}

func sendToAgent(address string, RequestURI string, exteraPath string, sReq []byte, agentId uint32) (*pb.TaskResponse, error) {
	// log.Printf("sendToAgent address: %v,  RequestURI :%s", address, RequestURI)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect: %v", err)
		mutexAgent.Lock()
		ageantAddresses[agentId].Loads--
		mutexAgent.Unlock()
		return nil, err
	}
	defer conn.Close()
	c := pb.NewTasksRequestClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	r, err := c.TaskAssign(ctx, &pb.TaskRequest{FunctionName: RequestURI, ExteraPath: exteraPath, SerializeReq: sReq})
	if err != nil {
		log.Printf("could not TaskAssign: %v", err)
		mutexAgent.Lock()
		ageantAddresses[agentId].Loads--
		mutexAgent.Unlock()
		return nil, err
	}
	// log.Printf("Response Message: %s", r.Message)
	mutexAgent.Lock()
	ageantAddresses[agentId].Loads--
	mutexAgent.Unlock()
	return r, err

}

func captureRequestData(req *http.Request) ([]byte, error) {
	var b = &bytes.Buffer{} // holds serialized representation
	//var tmp *http.Request
	var err error
	if err = req.Write(b); err != nil { // serialize request to HTTP/1.1 wire format
		return nil, err
	}
	//var reqSerialize []byte

	return b.Bytes(), nil
	//r := bufio.NewReader(b)
	//if tmp, err = http.ReadRequest(r); err != nil { // deserialize request
	//	return nil,err
	//}
	//*req = *tmp // replace original request structure
	//return nil
}

func unserializeReq(sReq []byte, req *http.Request) (*http.Response, error) {
	b := bytes.NewBuffer(sReq)
	r := bufio.NewReader(b)
	res, err := http.ReadResponse(r, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		// vClone := make([]string, len(v))
		// var vClone []string
		vClone := v
		// copy(vClone[:], v)

		destination[k] = vClone
	}
}

// getContentType resolves the correct Content-Type for a proxied function.
func getContentType(request http.Header, proxyResponse http.Header) (headerContentType string) {
	responseHeader := proxyResponse.Get("Content-Type")
	requestHeader := request.Get("Content-Type")

	if len(responseHeader) > 0 {
		headerContentType = responseHeader
	} else if len(requestHeader) > 0 {
		headerContentType = requestHeader
	} else {
		headerContentType = defaultContentType
	}

	return headerContentType
}

func hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
