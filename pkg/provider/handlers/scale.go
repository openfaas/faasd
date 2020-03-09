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
	faasd "github.com/openfaas/faasd/pkg"
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

		status, err := ScaleFunction(client, cni, req.ServiceName, req.Replicas)
		if status != http.StatusOK {
			http.Error(w, err.Error(), status)
		}
		return
	}
}

func ScaleFunction(client *containerd.Client, cni gocni.CNI, name string, replicas uint64) (int, error) {
	if _, err := GetFunction(client, name); err != nil {
		msg := fmt.Errorf("service %s not found", name)
		log.Printf("[Scale] %s\n", msg)
		return http.StatusNotFound, msg
	}

	ctx := namespaces.WithNamespace(context.Background(), faasd.FunctionNamespace)

	ctr, ctrErr := client.LoadContainer(ctx, name)
	if ctrErr != nil {
		msg := fmt.Errorf("cannot load service %s, error: %s", name, ctrErr)
		log.Printf("[Scale] %s\n", msg)
		return http.StatusNotFound, msg
	}

	taskExists := true
	task, taskErr := ctr.Task(ctx, nil)
	if taskErr != nil {
		msg := fmt.Errorf("cannot load task for service %s, error: %s", name, taskErr)
		log.Printf("[Scale] %s\n", msg)
		taskExists = false
	}

	if replicas > 0 {
		if taskExists {
			if status, statusErr := task.Status(ctx); statusErr == nil {
				if status.Status == containerd.Paused {
					if resumeErr := task.Resume(ctx); resumeErr != nil {
						log.Printf("[Scale] error resuming task %s, error: %s\n", name, resumeErr)
						return http.StatusBadRequest, resumeErr
					}
				}
			}
		} else {
			deployErr := createTask(ctx, client, ctr, cni)
			if deployErr != nil {
				log.Printf("[Scale] error deploying %s, error: %s\n", name, deployErr)
				return http.StatusBadRequest, deployErr
			}
			return http.StatusOK, nil
		}
	} else {
		if taskExists {
			if status, statusErr := task.Status(ctx); statusErr == nil {
				if status.Status == containerd.Running {
					if pauseErr := task.Pause(ctx); pauseErr != nil {
						log.Printf("[Scale] error pausing task %s, error: %s\n", name, pauseErr)
						return http.StatusBadRequest, pauseErr
					}
				}
			}
		}
	}
	return http.StatusOK, nil
}
