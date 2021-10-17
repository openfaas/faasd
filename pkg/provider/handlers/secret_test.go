package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/openfaas/faas-provider/types"
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
