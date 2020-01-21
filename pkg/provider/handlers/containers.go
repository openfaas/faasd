package handlers

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

type Function struct {
	name      string
	namespace string
	image     string
	pid       uint32
	replicas  int
	IP        string
}

const (
	// FunctionNamespace is the containerd namespace functions are created
	FunctionNamespace = "openfaas-fn"
)

// ListFunctions returns a map of all functions with running tasks on namespace
func ListFunctions(client *containerd.Client) (map[string]Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), FunctionNamespace)
	functions := make(map[string]Function)

	containers, _ := client.Containers(ctx)
	for _, k := range containers {
		name := k.ID()
		functions[name], _ = GetFunction(client, name)
	}
	return functions, nil
}

// GetFunction returns a function that matches name
func GetFunction(client *containerd.Client, name string) (Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), FunctionNamespace)
	c, err := client.LoadContainer(ctx, name)

	if err == nil {

		image, _ := c.Image(ctx)
		f := Function{
			name:      c.ID(),
			namespace: FunctionNamespace,
			image:     image.Name(),
		}

		replicas := 0
		task, err := c.Task(ctx, nil)
		if err == nil {
			// Task for container exists
			svc, err := task.Status(ctx)
			if err != nil {
				return Function{}, fmt.Errorf("unable to get task status for container: %s %s", name, err)
			}
			if svc.Status == "running" {
				replicas = 1
				f.pid = task.Pid()
				// Get container IP address
				ip, _ := GetIPfromPID(int(task.Pid()))
				f.IP = ip.String()
			}
		} else {
			replicas = 0
		}

		f.replicas = replicas
		return f, nil

	}
	return Function{}, fmt.Errorf("unable to find function: %s, error %s", name, err)
}
