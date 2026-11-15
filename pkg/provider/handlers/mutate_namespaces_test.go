package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containerd/containerd/namespaces"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
)

type fakeNamespaceStore struct {
	namespaces map[string]map[string]string
	deleted    []string
}

func newFakeNamespaceStore() *fakeNamespaceStore {
	return &fakeNamespaceStore{
		namespaces: map[string]map[string]string{},
	}
}

func (f *fakeNamespaceStore) Create(_ context.Context, namespace string, labels map[string]string) error {
	f.namespaces[namespace] = labels
	return nil
}

func (f *fakeNamespaceStore) Labels(_ context.Context, namespace string) (map[string]string, error) {
	labels, ok := f.namespaces[namespace]
	if !ok {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}
	return labels, nil
}

func (f *fakeNamespaceStore) SetLabel(_ context.Context, namespace, key, value string) error {
	if _, ok := f.namespaces[namespace]; !ok {
		return fmt.Errorf("namespace %s not found", namespace)
	}
	if value == "" {
		delete(f.namespaces[namespace], key)
	} else {
		f.namespaces[namespace][key] = value
	}
	return nil
}

func (f *fakeNamespaceStore) List(context.Context) ([]string, error) {
	names := []string{}
	for name := range f.namespaces {
		names = append(names, name)
	}
	return names, nil
}

func (f *fakeNamespaceStore) Delete(_ context.Context, namespace string, _ ...namespaces.DeleteOpts) error {
	if _, ok := f.namespaces[namespace]; !ok {
		return fmt.Errorf("namespace %s not found", namespace)
	}
	delete(f.namespaces, namespace)
	f.deleted = append(f.deleted, namespace)
	return nil
}

func TestDeleteNamespaceRequiresPersistedOpenFaaSLabel(t *testing.T) {
	store := newFakeNamespaceStore()

	store.Create(context.Background(), "containerd-plain", nil)
	store.Create(context.Background(), "foreign-ns", map[string]string{"team": "platform"})
	store.Create(context.Background(), "wrong-value", map[string]string{"openfaas": "0"})
	store.Create(context.Background(), "openfaas-ns", map[string]string{"openfaas": "1"})
	store.Create(context.Background(), "legacy-openfaas-ns", map[string]string{"openfaas": "true"})

	tests := []struct {
		name      string
		namespace string
		status    int
		deleted   bool
	}{
		{name: "unlabelled namespace is not deleted", namespace: "containerd-plain", status: http.StatusBadRequest},
		{name: "foreign-labelled namespace is not deleted", namespace: "foreign-ns", status: http.StatusBadRequest},
		{name: "wrong OpenFaaS label is not deleted", namespace: "wrong-value", status: http.StatusBadRequest},
		{name: "missing namespace is not deleted", namespace: "missing", status: http.StatusNotFound},
		{name: "OpenFaaS namespace is deleted", namespace: "openfaas-ns", status: http.StatusAccepted, deleted: true},
		{name: "legacy OpenFaaS namespace is deleted", namespace: "legacy-openfaas-ns", status: http.StatusAccepted, deleted: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(types.FunctionNamespace{
				Name:   tc.namespace,
				Labels: map[string]string{"openfaas": "1"},
			})
			if err != nil {
				t.Fatalf("marshal namespace request: %s", err)
			}

			req := httptest.NewRequest(http.MethodDelete, "/system/namespace/"+tc.namespace, strings.NewReader(string(payload)))
			req = mux.SetURLVars(req, map[string]string{"name": tc.namespace})
			response := httptest.NewRecorder()

			deleteNamespace(store, response, req)

			if response.Code != tc.status {
				t.Fatalf("want status %d, got %d: %s", tc.status, response.Code, response.Body.String())
			}

			wasDeleted := false
			for _, namespace := range store.deleted {
				if namespace == tc.namespace {
					wasDeleted = true
					break
				}
			}
			if wasDeleted != tc.deleted {
				t.Fatalf("want deleted %t, got %t", tc.deleted, wasDeleted)
			}
		})
	}
}
