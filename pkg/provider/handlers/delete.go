package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	gocni "github.com/containerd/go-cni"
	"github.com/openfaas/faas/gateway/requests"

	cninetwork "github.com/openfaas/faasd/pkg/cninetwork"
	"github.com/openfaas/faasd/pkg/service"
)

func MakeDeleteHandler(client *containerd.Client, cni gocni.CNI) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "expected a body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		log.Printf("[Delete] request: %s\n", string(body))

		req := requests.DeleteFunctionRequest{}
		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("[Delete] error parsing input: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

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

		name := req.FunctionName

		function, err := GetFunction(client, name, lookupNamespace)
		if err != nil {
			msg := fmt.Sprintf("function: %s not found", name)
			log.Printf("[Delete] %s\n", msg)
			http.Error(w, msg, http.StatusNotFound)
			return
		}

		ctx := namespaces.WithNamespace(context.Background(), lookupNamespace)

		// TODO: this needs to still happen if the task is paused
		if function.replicas != 0 {
			err = cninetwork.DeleteCNINetwork(ctx, cni, client, name)
			if err != nil {
				log.Printf("[Delete] error removing CNI network for: %s, %s\n", name, err)
			}
		}

		killTimeout := getKillTimeout(function.envVars)

		if err := service.Remove(ctx, client, name, killTimeout); err != nil {
			log.Printf("[Delete] error removing %s, %s\n", name, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("[Delete] deleted: %s\n", name)
	}
}

func getKillTimeout(envs map[string]string) time.Duration {
	killTimeout := time.Second * 5
	if v, ok := envs["healthcheck_interval"]; ok {
		dur, err := time.ParseDuration(v)
		if err == nil {
			killTimeout = dur
		}
	}
	return killTimeout
}
