package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/types"
)

func TestProviderNamespaceRoutesRequireBasicAuth(t *testing.T) {
	const (
		user     = "admin"
		password = "correct-horse-battery-staple"
	)

	secretMountPath := t.TempDir()
	for name, value := range map[string]string{
		"basic-auth-user":     user,
		"basic-auth-password": password,
	} {
		if err := os.WriteFile(filepath.Join(secretMountPath, name), []byte(value), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	var listCalls atomic.Int32
	var mutateCalls atomic.Int32
	handlers := types.FaaSHandlers{
		ListNamespaces: func(w http.ResponseWriter, _ *http.Request) {
			listCalls.Add(1)
			w.WriteHeader(http.StatusAccepted)
		},
		MutateNamespace: func(w http.ResponseWriter, _ *http.Request) {
			mutateCalls.Add(1)
			w.WriteHeader(http.StatusCreated)
		},
	}
	config := types.FaaSConfig{
		EnableBasicAuth: true,
		SecretMountPath: secretMountPath,
	}

	server := httptest.NewServer(bootstrap.MakeHandler(&handlers, &config))
	t.Cleanup(server.Close)
	client := server.Client()
	baseURL := server.URL

	tests := []struct {
		name              string
		method            string
		path              string
		authenticatedCode int
		calls             *atomic.Int32
	}{
		{
			name:              "list namespaces",
			method:            http.MethodGet,
			path:              "/system/namespaces",
			authenticatedCode: http.StatusAccepted,
			calls:             &listCalls,
		},
		{
			name:              "mutate namespace",
			method:            http.MethodPost,
			path:              "/system/namespace/",
			authenticatedCode: http.StatusCreated,
			calls:             &mutateCalls,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request, err := http.NewRequest(tc.method, baseURL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("want unauthenticated status %d, got %d", http.StatusUnauthorized, response.StatusCode)
			}
			if got := tc.calls.Load(); got != 0 {
				t.Fatalf("want protected handler not to be called, got %d calls", got)
			}

			request, err = http.NewRequest(tc.method, baseURL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			request.SetBasicAuth(user, password)
			response, err = client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != tc.authenticatedCode {
				t.Fatalf("want authenticated status %d, got %d", tc.authenticatedCode, response.StatusCode)
			}
			if got := tc.calls.Load(); got != 1 {
				t.Fatalf("want protected handler to be called once, got %d calls", got)
			}
		})
	}
}
