package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faasd/pkg"
	provider "github.com/openfaas/faasd/pkg/provider"
)

func Test_parseSecret(t *testing.T) {
	cases := []struct {
		name      string
		payload   string
		expError  string
		expSecret types.Secret
	}{
		{
			name:      "no error when name is valid without extention and with no traversal",
			payload:   `{"name": "authorized_keys", "value": "foo"}`,
			expSecret: types.Secret{Name: "authorized_keys", Value: "foo"},
		},
		{
			name:      "no error when name is valid and parses RawValue correctly",
			payload:   `{"name": "authorized_keys", "rawValue": "YmFy"}`,
			expSecret: types.Secret{Name: "authorized_keys", RawValue: []byte("bar")},
		},
		{
			name:      "no error when name is valid with dot and with no traversal",
			payload:   `{"name": "authorized.keys", "value": "foo"}`,
			expSecret: types.Secret{Name: "authorized.keys", Value: "foo"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.payload)
			r := httptest.NewRequest(http.MethodPost, "/", reader)
			secret, err := parseSecret(r)
			if err != nil && tc.expError == "" {
				t.Fatalf("unexpected error: %s", err)
				return
			}

			if tc.expError != "" {
				if err == nil {
					t.Fatalf("expected error: %s, got nil", tc.expError)
				}
				if err.Error() != tc.expError {
					t.Fatalf("expected error: %s, got: %s", tc.expError, err)
				}

				return
			}

			if !reflect.DeepEqual(secret, tc.expSecret) {
				t.Fatalf("expected secret: %+v, got: %+v", tc.expSecret, secret)
			}
		})
	}
}

func TestSecretCreation(t *testing.T) {
	mountPath, err := os.MkdirTemp("", "test_secret_creation")
	if err != nil {
		t.Fatalf("unexpected error while creating temp directory: %s", err)
	}

	defer os.RemoveAll(mountPath)

	handler := MakeSecretHandler(nil, mountPath)

	cases := []struct {
		name       string
		verb       string
		payload    string
		status     int
		secretPath string
		secret     string
		err        string
	}{
		{
			name:    "returns error when the name contains a traversal",
			verb:    http.MethodPost,
			payload: `{"name": "/root/.ssh/authorized_keys", "value": "foo"}`,
			status:  http.StatusBadRequest,
			err:     "directory traversal found in name\n",
		},
		{
			name:    "returns error when the name contains a traversal",
			verb:    http.MethodPost,
			payload: `{"name": "..", "value": "foo"}`,
			status:  http.StatusBadRequest,
			err:     "directory traversal found in name\n",
		},
		{
			name:    "empty request returns a validation error",
			verb:    http.MethodPost,
			payload: `{}`,
			status:  http.StatusBadRequest,
			err:     "non-empty name is required\n",
		},
		{
			name:       "can create secret from string",
			verb:       http.MethodPost,
			payload:    `{"name": "foo", "value": "bar"}`,
			status:     http.StatusOK,
			secretPath: "/openfaas-fn/foo",
			secret:     "bar",
		},
		{
			name:       "can create secret from raw value",
			verb:       http.MethodPost,
			payload:    `{"name": "foo", "rawValue": "YmFy"}`,
			status:     http.StatusOK,
			secretPath: "/openfaas-fn/foo",
			secret:     "bar",
		},
		{
			name:       "can create secret in non-default namespace from raw value",
			verb:       http.MethodPost,
			payload:    `{"name": "pity", "rawValue": "dGhlIGZvbw==", "namespace": "a-team"}`,
			status:     http.StatusOK,
			secretPath: "/a-team/pity",
			secret:     "the foo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.verb, "http://example.com/foo", strings.NewReader(tc.payload))
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tc.status {
				t.Logf("response body: %s", w.Body.String())
				t.Fatalf("expected status: %d, got: %d", tc.status, resp.StatusCode)
			}

			if resp.StatusCode != http.StatusOK && w.Body.String() != tc.err {
				t.Fatalf("expected error message: %q, got %q", tc.err, w.Body.String())

			}

			if tc.secretPath != "" {
				data, err := os.ReadFile(filepath.Join(mountPath, tc.secretPath))
				if err != nil {
					t.Fatalf("can not read the secret from disk: %s", err)
				}

				if string(data) != tc.secret {
					t.Fatalf("expected secret value: %s, got %s", tc.secret, string(data))
				}
			}
		})
	}
}

func TestListSecrets(t *testing.T) {
	mountPath, err := os.MkdirTemp("", "test_secret_creation")
	if err != nil {
		t.Fatalf("unexpected error while creating temp directory: %s", err)
	}

	defer os.RemoveAll(mountPath)

	cases := []struct {
		name       string
		verb       string
		namespace  string
		labels     map[string]string
		status     int
		secretPath string
		secret     string
		err        string
		expected   []types.Secret
	}{
		{
			name:       "Get empty secret list for default namespace having no secret",
			verb:       http.MethodGet,
			status:     http.StatusOK,
			secretPath: "/test-fn/foo",
			secret:     "bar",
			expected:   make([]types.Secret, 0),
		},
		{
			name:       "Get empty secret list for non-default namespace having no secret",
			verb:       http.MethodGet,
			status:     http.StatusOK,
			secretPath: "/test-fn/foo",
			secret:     "bar",
			expected:   make([]types.Secret, 0),
			namespace:  "other-ns",
			labels: map[string]string{
				pkg.NamespaceLabel: "true",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			labelStore := provider.NewFakeLabeller(tc.labels)

			handler := MakeSecretHandler(labelStore, mountPath)

			path := "http://example.com/foo"
			if len(tc.namespace) > 0 {
				path = path + fmt.Sprintf("?namespace=%s", tc.namespace)
			}
			req := httptest.NewRequest(tc.verb, path, nil)
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tc.status {
				t.Fatalf("want status: %d, but got: %d", tc.status, resp.StatusCode)
			}

			if resp.StatusCode != http.StatusOK && w.Body.String() != tc.err {
				t.Fatalf("want error message: %q, but got %q", tc.err, w.Body.String())
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("can't read response of list %v", err)
			}

			var res []types.Secret
			err = json.Unmarshal(body, &res)

			if err != nil {
				t.Fatalf("unable to unmarshal %q, error: %v", string(body), err)
			}

			if !reflect.DeepEqual(res, tc.expected) {
				t.Fatalf("want response: %v, but got: %v", tc.expected, res)
			}
		})
	}
}
