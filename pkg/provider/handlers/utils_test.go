package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	faasd "github.com/openfaas/faasd/pkg"
)

func Test_getRequestNamespace(t *testing.T) {
	tables := []struct {
		name              string
		requestNamespace  string
		expectedNamespace string
	}{
		{name: "RequestNamespace is not provided", requestNamespace: "", expectedNamespace: faasd.DefaultFunctionNamespace},
		{name: "RequestNamespace is provided", requestNamespace: "user-namespace", expectedNamespace: "user-namespace"},
	}

	for _, tc := range tables {
		t.Run(tc.name, func(t *testing.T) {
			actualNamespace := getRequestNamespace(tc.requestNamespace)
			if actualNamespace != tc.expectedNamespace {
				t.Errorf("Want: %s, got: %s", actualNamespace, tc.expectedNamespace)
			}
		})
	}
}

func Test_getNamespaceSecretMountPath(t *testing.T) {
	userSecretPath := "/var/openfaas/secrets"
	tables := []struct {
		name               string
		requestNamespace   string
		expectedSecretPath string
		expError           bool
	}{
		{name: "Default Namespace is provided", requestNamespace: faasd.DefaultFunctionNamespace, expectedSecretPath: "/var/openfaas/secrets/" + faasd.DefaultFunctionNamespace},
		{name: "User Namespace is provided", requestNamespace: "user-namespace", expectedSecretPath: "/var/openfaas/secrets/user-namespace"},
		{name: "Namespace with traversal is rejected", requestNamespace: "../../../../tmp/escape", expError: true},
	}

	for _, tc := range tables {
		t.Run(tc.name, func(t *testing.T) {
			actualNamespace, err := getNamespaceSecretMountPath(userSecretPath, tc.requestNamespace)
			if tc.expError {
				if err == nil {
					t.Fatalf("expected error for namespace: %s, got nil", tc.requestNamespace)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if actualNamespace != tc.expectedSecretPath {
				t.Errorf("Want: %s, got: %s", actualNamespace, tc.expectedSecretPath)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	tables := []struct {
		name      string
		namespace string
		expError  bool
	}{
		{name: "Default namespace is valid", namespace: faasd.DefaultFunctionNamespace},
		{name: "User namespace is valid", namespace: "user-namespace"},
		{name: "Namespace with a dot is valid", namespace: "user.namespace"},
		{name: "Empty namespace is invalid", namespace: "", expError: true},
		{name: "Namespace with a path separator is invalid", namespace: "a/b", expError: true},
		{name: "Namespace with traversal is invalid", namespace: "../..", expError: true},
		{name: "Namespace with deep traversal is invalid", namespace: "../../../../tmp/escape", expError: true},
		{name: "Namespace with a backslash is invalid", namespace: "..\\..\\tmp", expError: true},
		{name: "Namespace with a leading dot is invalid", namespace: ".hidden", expError: true},
		{name: "Namespace with a trailing dash is invalid", namespace: "user-", expError: true},
		{name: "Namespace with spaces is invalid", namespace: "user namespace", expError: true},
	}

	for _, tc := range tables {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNamespace(tc.namespace)
			if tc.expError {
				if err == nil {
					t.Fatalf("expected error for namespace: %q, got nil", tc.namespace)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for namespace %q: %s", tc.namespace, err)
			}
		})
	}
}

func Test_readNamespaceFromQuery(t *testing.T) {
	tables := []struct {
		name              string
		queryNamespace    string
		expectedNamespace string
	}{
		{name: "No Namespace is provided", queryNamespace: "", expectedNamespace: ""},
		{name: "User Namespace is provided", queryNamespace: "user-namespace", expectedNamespace: "user-namespace"},
	}

	for _, tc := range tables {
		t.Run(tc.name, func(t *testing.T) {

			url := fmt.Sprintf("/test?namespace=%s", tc.queryNamespace)
			r := httptest.NewRequest(http.MethodGet, url, nil)

			actualNamespace := readNamespaceFromQuery(r)
			if actualNamespace != tc.expectedNamespace {
				t.Errorf("Want: %s, got: %s", actualNamespace, tc.expectedNamespace)
			}
		})
	}
}
