package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/openfaas/faasd/pkg/cninetwork"
	"github.com/openfaas/faasd/pkg/service"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	gocni "github.com/containerd/go-cni"
	"github.com/openfaas/faas-provider/types"
)

func MakeUpdateHandler(client *containerd.Client, cni gocni.CNI, secretMountPath string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "expected a body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		log.Printf("[Update] request: %s\n", string(body))

		req := types.FunctionDeployment{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			log.Printf("[Update] error parsing input: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}
		name := req.Service

		function, err := GetFunction(client, name)
		if err != nil {
			msg := fmt.Sprintf("service %s not found", name)
			log.Printf("[Update] %s\n", msg)
			http.Error(w, msg, http.StatusNotFound)
			return
		}

		err = validateSecrets(secretMountPath, req.Secrets)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		ctx := namespaces.WithNamespace(context.Background(), FunctionNamespace)
		if function.replicas != 0 {
			err = cninetwork.DeleteCNINetwork(ctx, cni, client, name)
			if err != nil {
				log.Printf("[Update] error removing CNI network for %s, %s\n", name, err)
			}
		}

		containerErr := service.Remove(ctx, client, name)
		if containerErr != nil {
			log.Printf("[Update] error removing %s, %s\n", name, containerErr)
			http.Error(w, containerErr.Error(), http.StatusInternalServerError)
			return
		}

		deployErr := deploy(ctx, req, client, cni, secretMountPath)
		if deployErr != nil {
			log.Printf("[Update] error deploying %s, error: %s\n", name, deployErr)
			http.Error(w, deployErr.Error(), http.StatusBadRequest)
			return
		}
	}

}
