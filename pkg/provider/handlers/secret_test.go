package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas-provider/types"
)

func Test_parseSecretValidName(t *testing.T) {

	s := types.Secret{Name: "authorized_keys"}
	body, _ := json.Marshal(s)
	reader := bytes.NewReader(body)
	r := httptest.NewRequest(http.MethodPost, "/", reader)
	_, err := parseSecret(r)

	if err != nil {
		t.Fatalf("secret name is valid with no traversal characters")
	}
}

func Test_parseSecretValidNameWithDot(t *testing.T) {

	s := types.Secret{Name: "authorized.keys"}
	body, _ := json.Marshal(s)
	reader := bytes.NewReader(body)
	r := httptest.NewRequest(http.MethodPost, "/", reader)
	_, err := parseSecret(r)

	if err != nil {
		t.Fatalf("secret name is valid with no traversal characters")
	}
}

func Test_parseSecretWithTraversalWithSlash(t *testing.T) {

	s := types.Secret{Name: "/root/.ssh/authorized_keys"}
	body, _ := json.Marshal(s)
	reader := bytes.NewReader(body)
	r := httptest.NewRequest(http.MethodPost, "/", reader)
	_, err := parseSecret(r)

	if err == nil {
		t.Fatalf("secret name should fail due to path traversal")
	}
}

func Test_parseSecretWithTraversalWithDoubleDot(t *testing.T) {

	s := types.Secret{Name: ".."}
	body, _ := json.Marshal(s)
	reader := bytes.NewReader(body)
	r := httptest.NewRequest(http.MethodPost, "/", reader)
	_, err := parseSecret(r)

	if err == nil {
		t.Fatalf("secret name should fail due to path traversal")
	}
}
