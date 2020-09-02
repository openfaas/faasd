package pkg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"time"
)

// NewProxy creates a HTTP proxy to expose the gateway container
// from OpenFaaS to the host
func NewProxy(port int, timeout time.Duration) *Proxy {

	return &Proxy{
		Port:    port,
		Timeout: timeout,
	}
}

// Proxy for exposing a private container
type Proxy struct {
	Timeout time.Duration
	Port    int
}

// Start listening and forwarding HTTP to the host
func (p *Proxy) Start(proxyChan chan string, done chan bool, proxyName string) error {
	tcp := p.Port

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	ps := proxyState{
		Host: "",
	}

	ps.Host = <-proxyChan

	log.Printf("Starting faasd proxy on %d\n", tcp)

	fmt.Printf("%s: %s\n", proxyName, ps.Host)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcp),
		ReadTimeout:    p.Timeout,
		WriteTimeout:   p.Timeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler:        http.HandlerFunc(makeProxy(&ps)),
	}

	go func() {
		log.Printf("[proxy] Begin listen on %d\n", p.Port)
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Error ListenAndServe: %v", err)
		}
	}()

	log.Println("[proxy] Wait for done")
	<-done
	log.Println("[proxy] Done received")
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("[proxy] Error in Shutdown: %v", err)
	}

	return nil
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		destination[k] = vClone
	}
}

type proxyState struct {
	Host string
}

func makeProxy(ps *proxyState) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		query := ""
		if len(r.URL.RawQuery) > 0 {
			query = "?" + r.URL.RawQuery
		}

		upstream := fmt.Sprintf("http://%s%s%s", ps.Host, r.URL.Path, query)
		fmt.Printf("[faasd] proxy: %s\n", upstream)

		if r.Body != nil {
			defer r.Body.Close()
		}

		wrapper := ioutil.NopCloser(r.Body)
		upReq, upErr := http.NewRequest(r.Method, upstream, wrapper)

		copyHeaders(upReq.Header, &r.Header)

		if upErr != nil {
			log.Println(upErr)

			http.Error(w, upErr.Error(), http.StatusInternalServerError)
			return
		}

		upRes, upResErr := http.DefaultClient.Do(upReq)

		if upResErr != nil {
			log.Println(upResErr)

			http.Error(w, upResErr.Error(), http.StatusInternalServerError)
			return
		}

		copyHeaders(w.Header(), &upRes.Header)

		w.WriteHeader(upRes.StatusCode)
		io.Copy(w, upRes.Body)
	}
}
