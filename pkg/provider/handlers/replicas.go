package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/containerd/containerd"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
)

func MakeReplicaReaderHandler(client *containerd.Client) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]

		if f, err := GetFunction(client, functionName); err == nil {
			found := types.FunctionStatus{
				Name:              functionName,
				AvailableReplicas: uint64(f.replicas),
				Replicas:          uint64(f.replicas),
				Namespace:         f.namespace,
				Labels:            &f.labels,
				Annotations:       &f.annotations,
			}

			functionBytes, _ := json.Marshal(found)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(functionBytes)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
