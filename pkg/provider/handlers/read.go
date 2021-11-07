package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/containerd/containerd"
	"github.com/openfaas/faas-provider/types"
)

func MakeReadHandler(client *containerd.Client) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		lookupNamespace := getRequestNamespace(readNamespaceFromQuery(r))
		// Check if namespace exists, and it has the openfaas label
		valid, err := validNamespace(client.NamespaceService(), lookupNamespace)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !valid {
			http.Error(w, "namespace not valid", http.StatusBadRequest)
			return
		}

		res := []types.FunctionStatus{}
		fns, err := ListFunctions(client, lookupNamespace)
		if err != nil {
			log.Printf("[Read] error listing functions. Error: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, fn := range fns {
			annotations := &fn.annotations
			labels := &fn.labels
			memory := resource.NewQuantity(fn.memoryLimit, resource.BinarySI)
			res = append(res, types.FunctionStatus{
				Name:        fn.name,
				Image:       fn.image,
				Replicas:    uint64(fn.replicas),
				Namespace:   fn.namespace,
				Labels:      labels,
				Annotations: annotations,
				Secrets:     fn.secrets,
				EnvVars:     fn.envVars,
				EnvProcess:  fn.envProcess,
				Limits:      &types.FunctionResources{Memory: memory.String()},
				CreatedAt:   fn.createdAt,
			})
		}

		body, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}
