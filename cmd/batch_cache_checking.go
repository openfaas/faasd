package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/openfaas/faasd/proto/agent"
	"google.golang.org/grpc"
)

func sendAllCacheRequest(agentID uint8, address string, requests []string, resChan []chan pb.TaskResponse) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("sendAllCacheRequest, did not connect: %v", err.Error())
		return
	}
	defer conn.Close()
	c := pb.NewTasksRequestClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := c.TaskAssign(ctx, &pb.TaskRequest{RequestHashes: requests})
	if err != nil {
		log.Printf("sendAllCacheRequest, could not TaskAssign: %v", err)
		return
	}

	for i, res := range r.Responses {
		if FileCaching {
			if len(res) > 0 {
				resChan[i] <- pb.TaskResponse{Response: append([]byte{}, agentID)}
			} else {
				resChan[i] <- pb.TaskResponse{Response: make([]byte, 0)}
			}
			continue
		}
		resChan[i] <- pb.TaskResponse{Response: res}
	}

}

func checkAllNodesCache() {
	var requests []string
	var resChan []chan pb.TaskResponse
	timerBatch := time.NewTicker(batchTime * time.Millisecond)
	for {
		select {
		case cacheReq := <-hashRequests:
			requests = append(requests, cacheReq.sReqHash)
			resChan = append(resChan, cacheReq.resultChan)
		case <-timerBatch.C:

			if len(requests) > 0 {
				fmt.Println("Send batch requests")
				go sendAllCacheRequest(0, ageantAddresses[0].Address, requests, resChan)
				go sendAllCacheRequest(1, ageantAddresses[1].Address, requests, resChan)
				go sendAllCacheRequest(2, ageantAddresses[2].Address, requests, resChan)
				go sendAllCacheRequest(3, ageantAddresses[3].Address, requests, resChan)
				requests = requests[:0]
				resChan = resChan[:0]
			}
			// default:
			// 	t
		}

	}
}
