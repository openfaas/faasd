package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/openfaas/faasd/pkg/cninetwork"

	faasd "github.com/openfaas/faasd/pkg"
)

type Function struct {
	name        string
	namespace   string
	image       string
	pid         uint32
	replicas    int
	IP          string
	labels      map[string]string
	annotations map[string]string
}

// ListFunctions returns a map of all functions with running tasks on namespace
func ListFunctions(client *containerd.Client) (map[string]*Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), faasd.FunctionNamespace)
	functions := make(map[string]*Function)

	containers, err := client.Containers(ctx)
	if err != nil {
		return functions, err
	}

	for _, c := range containers {
		name := c.ID()
		f, err := GetFunction(client, name)
		if err != nil {
			log.Printf("error getting function %s: ", name)
			return functions, err
		}
		functions[name] = &f
	}

	return functions, nil
}

// GetFunction returns a function that matches name
func GetFunction(client *containerd.Client, name string) (Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), faasd.FunctionNamespace)
	fn := Function{}

	c, err := client.LoadContainer(ctx, name)
	if err != nil {
		return Function{}, fmt.Errorf("unable to find function: %s, error %s", name, err)
	}

	image, err := c.Image(ctx)
	if err != nil {
		return fn, err
	}

	containerName := c.ID()
	allLabels, labelErr := c.Labels(ctx)

	if labelErr != nil {
		log.Printf("cannot list container %s labels: %s", containerName, labelErr.Error())
	}

	labels, annotations := buildLabelsAndAnnotations(allLabels)

	fn.name = containerName
	fn.namespace = faasd.FunctionNamespace
	fn.image = image.Name()
	fn.labels = labels
	fn.annotations = annotations

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
			fn.pid = task.Pid()

			// Get container IP address
			ip, err := cninetwork.GetIPfromPID(int(task.Pid()))
			if err != nil {
				return Function{}, err
			}
			fn.IP = ip.String()
		}
	} else {
		replicas = 0
	}

	fn.replicas = replicas
	return fn, nil
}

// buildLabelsAndAnnotations returns a separated list with labels first,
// followed by annotations by checking each key of ctrLabels for a prefix.
func buildLabelsAndAnnotations(ctrLabels map[string]string) (map[string]string, map[string]string) {
	labels := make(map[string]string)
	annotations := make(map[string]string)

	for k, v := range ctrLabels {
		if strings.HasPrefix(k, annotationLabelPrefix) {
			annotations[strings.TrimPrefix(k, annotationLabelPrefix)] = v
		} else {
			labels[k] = v
		}
	}

	return labels, annotations
}
