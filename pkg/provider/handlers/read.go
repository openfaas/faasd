package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/containerd/containerd"
	"github.com/openfaas/faas-provider/types"
)

func MakeReadHandler(client *containerd.Client) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		res := []types.FunctionStatus{}
		funcs, err := ListFunctions(client)
		if err != nil {
			log.Printf("[Read] error listing functions. Error: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, function := range funcs {

			res = append(res, types.FunctionStatus{
				Name:        function.name,
				Image:       function.image,
				Replicas:    uint64(function.replicas),
				Namespace:   function.namespace,
				Labels:      &function.labels,
				Annotations: &function.annotations,
			})
		}

		body, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)

	}
}
