package cmd

import "time"

func checkAllNodesCache() {
	var requests []string
	timerBatch := time.NewTimer(batchTime * time.Millisecond)
	for {
		select {
		case req := <-hashRequests:
			requests = append(requests, req)
		case <-timerBatch:
			requests = requests[:0]
		default:
		}

	}
}

func sendAllCacheRequest() {

}
