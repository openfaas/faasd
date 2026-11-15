package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
)

func TestProtectNamespaceHandlers(t *testing.T) {
	credentials := &auth.BasicAuthCredentials{
		User:     "admin",
		Password: "correct-horse-battery-staple",
	}

	tests := []struct {
		name    string
		method  string
		path    string
		handler func(*types.FaaSHandlers) http.HandlerFunc
	}{
		{
			name:   "list namespaces",
			method: http.MethodGet,
			path:   "/system/namespaces",
			handler: func(handlers *types.FaaSHandlers) http.HandlerFunc {
				return handlers.ListNamespaces
			},
		},
		{
			name:   "mutate namespace",
			method: http.MethodPost,
			path:   "/system/namespace/",
			handler: func(handlers *types.FaaSHandlers) http.HandlerFunc {
				return handlers.MutateNamespace
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := 0
			stub := func(w http.ResponseWriter, _ *http.Request) {
				called++
				w.WriteHeader(http.StatusAccepted)
			}

			handlers := types.FaaSHandlers{
				ListNamespaces:  stub,
				MutateNamespace: stub,
			}
			protectNamespaceHandlers(&handlers, credentials)

			unauthenticated := httptest.NewRequest(tc.method, tc.path, nil)
			unauthenticatedResponse := httptest.NewRecorder()
			tc.handler(&handlers)(unauthenticatedResponse, unauthenticated)

			if unauthenticatedResponse.Code != http.StatusUnauthorized {
				t.Fatalf("want unauthenticated status %d, got %d", http.StatusUnauthorized, unauthenticatedResponse.Code)
			}
			if called != 0 {
				t.Fatalf("want protected handler not to be called, got %d calls", called)
			}

			authenticated := httptest.NewRequest(tc.method, tc.path, nil)
			authenticated.SetBasicAuth(credentials.User, credentials.Password)
			authenticatedResponse := httptest.NewRecorder()
			tc.handler(&handlers)(authenticatedResponse, authenticated)

			if authenticatedResponse.Code != http.StatusAccepted {
				t.Fatalf("want authenticated status %d, got %d", http.StatusAccepted, authenticatedResponse.Code)
			}
			if called != 1 {
				t.Fatalf("want protected handler to be called once, got %d calls", called)
			}
		})
	}
}

func TestProtectNamespaceHandlersAllowsNilHandlers(t *testing.T) {
	handlers := types.FaaSHandlers{}
	protectNamespaceHandlers(&handlers, &auth.BasicAuthCredentials{})
}
