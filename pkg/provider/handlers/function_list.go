package handlers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
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
	secrets     []string
	envVars     map[string]string
	envProcess  string
	memoryLimit int64
	createdAt   time.Time
}

// ListFunctions returns a map of all functions with running tasks on namespace
func ListFunctions(client *containerd.Client, namespace string) (map[string]*Function, error) {

	// Check if namespace exists, and it has the openfaas label
	valid, err := validNamespace(client.NamespaceService(), namespace)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, errors.New("namespace not valid")
	}

	ctx := namespaces.WithNamespace(context.Background(), namespace)
	functions := make(map[string]*Function)

	containers, err := client.Containers(ctx)
	if err != nil {
		return functions, err
	}

	for _, c := range containers {
		name := c.ID()
		f, err := GetFunction(client, name, namespace)
		if err != nil {
			log.Printf("skipping %s, error: %s", name, err)
		} else {
			functions[name] = &f
		}
	}

	return functions, nil
}
