package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	gocni "github.com/containerd/go-cni"
	"github.com/openfaas/faas-provider/types"
)

func MakeReplicaUpdateHandler(client *containerd.Client, cni gocni.CNI) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "expected a body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		log.Printf("[Scale] request: %s\n", string(body))

		req := types.ScaleServiceRequest{}
		err := json.Unmarshal(body, &req)

		if err != nil {
			log.Printf("[Scale] error parsing input: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		name := req.ServiceName

		if _, err := GetFunction(client, name); err != nil {
			msg := fmt.Sprintf("service %s not found", name)
			log.Printf("[Scale] %s\n", msg)
			http.Error(w, msg, http.StatusNotFound)
			return
		}

		ctx := namespaces.WithNamespace(context.Background(), FunctionNamespace)

		ctr, ctrErr := client.LoadContainer(ctx, name)
		if ctrErr != nil {
			msg := fmt.Sprintf("cannot load service %s, error: %s", name, ctrErr)
			log.Printf("[Scale] %s\n", msg)
			http.Error(w, msg, http.StatusNotFound)
			return
		}

		taskExists := true
		task, taskErr := ctr.Task(ctx, nil)
		if taskErr != nil {
			msg := fmt.Sprintf("cannot load task for service %s, error: %s", name, taskErr)
			log.Printf("[Scale] %s\n", msg)
			taskExists = false
		}

		if req.Replicas > 0 {
			if taskExists {
				if status, statusErr := task.Status(ctx); statusErr == nil {
					if status.Status == containerd.Paused {
						if resumeErr := task.Resume(ctx); resumeErr != nil {
							log.Printf("[Scale] error resuming task %s, error: %s\n", name, resumeErr)
							http.Error(w, resumeErr.Error(), http.StatusBadRequest)
						}
					}
				}
			} else {
				deployErr := createTask(ctx, client, ctr, cni)
				if deployErr != nil {
					log.Printf("[Scale] error deploying %s, error: %s\n", name, deployErr)
					http.Error(w, deployErr.Error(), http.StatusBadRequest)
					return
				}
				return
			}
		} else {
			if taskExists {
				if status, statusErr := task.Status(ctx); statusErr == nil {
					if status.Status == containerd.Running {
						if pauseErr := task.Pause(ctx); pauseErr != nil {
							log.Printf("[Scale] error pausing task %s, error: %s\n", name, pauseErr)
							http.Error(w, pauseErr.Error(), http.StatusBadRequest)
						}
					}
				}
			}
		}

	}

}
