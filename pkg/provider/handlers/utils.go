package handlers

import (
	"context"
	"github.com/containerd/containerd"
	"net/http"
	"path"

	faasd "github.com/openfaas/faasd/pkg"
)

func getRequestNamespace(namespace string) string {

	if len(namespace) > 0 {
		return namespace
	}
	return faasd.FunctionNamespace
}

func readNamespaceFromQuery(r *http.Request) string {
	q := r.URL.Query()
	return q.Get("namespace")
}

func getNamespaceSecretMountPath(userSecretPath string, namespace string) string {
	return path.Join(userSecretPath, namespace)
}

func validateNamespace(client *containerd.Client, namespace string) (bool, error) {
	if namespace == faasd.FunctionNamespace {
		return true, nil
	}

	store := client.NamespaceService()
	labels, err := store.Labels(context.Background(), namespace)
	if err != nil {
		return false, err
	}

	value, found := labels["openfaas"]

	if found {
		if value == "true" {
			return true, nil
		}
	}

	return false, nil
}
