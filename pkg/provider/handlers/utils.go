package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/containerd/containerd/identifiers"
	"github.com/openfaas/faasd/pkg"
	faasd "github.com/openfaas/faasd/pkg"
	provider "github.com/openfaas/faasd/pkg/provider"
)

func getRequestNamespace(namespace string) string {

	if len(namespace) > 0 {
		return namespace
	}
	return faasd.DefaultFunctionNamespace
}

func readNamespaceFromQuery(r *http.Request) string {
	q := r.URL.Query()
	return q.Get("namespace")
}

func getNamespaceSecretMountPath(userSecretPath string, namespace string) (string, error) {
	mountPath := path.Join(userSecretPath, namespace)

	cleanedRoot := path.Clean(userSecretPath)
	cleanedPath := path.Clean(mountPath)
	if cleanedPath != cleanedRoot && !strings.HasPrefix(cleanedPath, cleanedRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("namespace %q escapes the secrets root %q", namespace, userSecretPath)
	}

	return mountPath, nil
}

func getSecretMountPath(namespaceSecretPath string, secret string) (string, error) {
	if err := validateSecretName(secret); err != nil {
		return "", err
	}

	secretPath := path.Join(namespaceSecretPath, secret)
	cleanedRoot := path.Clean(namespaceSecretPath)
	cleanedPath := path.Clean(secretPath)
	if cleanedPath == cleanedRoot || !strings.HasPrefix(cleanedPath, cleanedRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("secret %q escapes the namespace secrets directory %q", secret, namespaceSecretPath)
	}

	return secretPath, nil
}

func validateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("non-empty namespace is required")
	}

	if err := identifiers.Validate(namespace); err != nil {
		return fmt.Errorf("invalid namespace %q: %s", namespace, err.Error())
	}

	return nil
}

// validNamespace indicates whether the namespace is eligable to be
// used for OpenFaaS functions.
func validNamespace(store provider.Labeller, namespace string) (bool, error) {
	if namespace == faasd.DefaultFunctionNamespace {
		return true, nil
	}

	labels, err := store.Labels(context.Background(), namespace)
	if err != nil {
		return false, err
	}

	// check for true to keep it backward compatible
	if value, found := labels[pkg.NamespaceLabel]; found && (value == "true" || value == "1") {
		return true, nil
	}

	return false, nil
}
