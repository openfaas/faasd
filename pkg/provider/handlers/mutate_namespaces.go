package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/containerd/containerd"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
)

func MakeMutateNamespace(client *containerd.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodPost:
			createNamespace(client, w, r)
		case http.MethodGet:
			getNamespace(client, w, r)
		case http.MethodDelete:
			deleteNamespace(client, w, r)
		case http.MethodPut:
			updateNamespace(client, w, r)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func updateNamespace(client *containerd.Client, w http.ResponseWriter, r *http.Request) {
	req, err := parseNamespaceRequest(r)
	if err != nil {
		http.Error(w, err.Error(), err.(*HttpError).Status)
		return
	}

	namespaceExists, err := namespaceExists(r.Context(), client, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !namespaceExists {
		http.Error(w, fmt.Sprintf("namespace %s not found", req.Name), http.StatusNotFound)
		return
	}

	originalLabels, err := client.NamespaceService().Labels(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !hasOpenFaaSLabel(originalLabels) {
		http.Error(w, fmt.Sprintf("namespace %s is not an openfaas namespace", req.Name), http.StatusBadRequest)
		return
	}

	var exclusions []string

	// build exclusions
	for key, _ := range originalLabels {
		if _, ok := req.Labels[key]; !ok {
			exclusions = append(exclusions, key)
		}
	}

	// Call SetLabel with empty string if label is to be removed
	for _, key := range exclusions {
		if err := client.NamespaceService().SetLabel(r.Context(), req.Name, key, ""); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Now add the new labels
	for key, value := range req.Labels {
		if err := client.NamespaceService().SetLabel(r.Context(), req.Name, key, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func deleteNamespace(client *containerd.Client, w http.ResponseWriter, r *http.Request) {
	req, err := parseNamespaceRequest(r)
	if err != nil {
		http.Error(w, err.Error(), err.(*HttpError).Status)
		return
	}

	if err := client.NamespaceService().Delete(r.Context(), req.Name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("namespace %s not found", req.Name), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func namespaceExists(ctx context.Context, client *containerd.Client, name string) (bool, error) {
	ns, err := client.NamespaceService().List(ctx)
	if err != nil {
		return false, err
	}

	found := false
	for _, namespace := range ns {
		if namespace == name {
			found = true
			break
		}
	}

	return found, nil
}

func getNamespace(client *containerd.Client, w http.ResponseWriter, r *http.Request) {
	req, err := parseNamespaceRequest(r)
	if err != nil {
		http.Error(w, err.Error(), err.(*HttpError).Status)
		return
	}

	namespaceExists, err := namespaceExists(r.Context(), client, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !namespaceExists {
		http.Error(w, fmt.Sprintf("namespace %s not found", req.Name), http.StatusNotFound)
		return
	}

	labels, err := client.NamespaceService().Labels(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !hasOpenFaaSLabel(labels) {
		http.Error(w, fmt.Sprintf("namespace %s not found", req.Name), http.StatusNotFound)
		return
	}

	res := types.FunctionNamespace{
		Name:   req.Name,
		Labels: labels,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Get Namespace error: %s", err)
	}
}

func createNamespace(client *containerd.Client, w http.ResponseWriter, r *http.Request) {
	req, err := parseNamespaceRequest(r)
	if err != nil {
		http.Error(w, err.Error(), err.(*HttpError).Status)
		return
	}

	// Check if namespace exists, and it has the openfaas label
	namespaces, err := client.NamespaceService().List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	found := false
	for _, namespace := range namespaces {
		if namespace == req.Name {
			found = true
			break
		}
	}

	if found {
		http.Error(w, fmt.Sprintf("namespace %s already exists", req.Name), http.StatusConflict)
		return
	}

	if err := client.NamespaceService().Create(r.Context(), req.Name, req.Labels); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// getNamespace returns a namespace object or an error
func parseNamespaceRequest(r *http.Request) (types.FunctionNamespace, error) {
	var req types.FunctionNamespace

	vars := mux.Vars(r)
	namespaceInPath := vars["name"]

	if r.Method == http.MethodGet {
		if namespaceInPath == "" {
			return req, &HttpError{
				Err:    fmt.Errorf("namespace not specified in URL"),
				Status: http.StatusBadRequest,
			}
		}

		return types.FunctionNamespace{
			Name: namespaceInPath,
		}, nil
	}

	body, _ := io.ReadAll(r.Body)

	if err := json.Unmarshal(body, &req); err != nil {
		return req, &HttpError{
			Err:    fmt.Errorf("error parsing request body: %s", err.Error()),
			Status: http.StatusBadRequest,
		}
	}

	if r.Method != http.MethodPost {
		if namespaceInPath == "" {
			return req, &HttpError{
				Err:    fmt.Errorf("namespace not specified in URL"),
				Status: http.StatusBadRequest,
			}
		}
		if req.Name != namespaceInPath {
			return req, &HttpError{
				Err:    fmt.Errorf("namespace in request body does not match namespace in URL"),
				Status: http.StatusBadRequest,
			}
		}
	}

	if req.Name == "" {
		return req, &HttpError{
			Err:    fmt.Errorf("namespace not specified in request body"),
			Status: http.StatusBadRequest,
		}
	}

	if ok := hasOpenFaaSLabel(req.Labels); !ok {
		return req, &HttpError{
			Err:    fmt.Errorf("request does not have openfaas=1 label"),
			Status: http.StatusBadRequest,
		}
	}

	return req, nil
}

func hasOpenFaaSLabel(labels map[string]string) bool {
	if v, ok := labels["openfaas"]; ok && v == "1" {
		return true
	}

	return false
}

type HttpError struct {
	Err    error
	Status int
}

func (e *HttpError) Error() string {
	return e.Err.Error()
}
