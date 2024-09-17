package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/containerd/containerd"
	"github.com/openfaas/faasd/pkg"
	faasd "github.com/openfaas/faasd/pkg"
)

func MakeNamespacesLister(client *containerd.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		list := ListNamespaces(client)
		body, _ := json.Marshal(list)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func ListNamespaces(client *containerd.Client) []string {
	set := []string{faasd.DefaultFunctionNamespace}

	store := client.NamespaceService()

	namespaces, err := store.List(context.Background())
	if err != nil {
		log.Printf("Error listing namespaces: %s", err.Error())
		return set
	}

	for _, namespace := range namespaces {
		labels, err := store.Labels(context.Background(), namespace)
		if err != nil {
			log.Printf("Error listing label for namespace %s: %s", namespace, err.Error())
			continue
		}

		if _, found := labels[pkg.NamespaceLabel]; found {
			set = append(set, namespace)
		}

	}

	if len(set) == 0 {
		set = append(set, faasd.DefaultFunctionNamespace)
	}

	return set
}

func findNamespace(target string, items []string) bool {
	for _, n := range items {
		if n == target {
			return true
		}
	}
	return false
}
