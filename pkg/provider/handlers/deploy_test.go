package handlers

import (
	"fmt"
	"reflect"
	"testing"

	"os"
	"path/filepath"
	"strings"

	"github.com/openfaas/faas-provider/types"
)

func Test_BuildLabels_WithAnnotations(t *testing.T) {
	// Test each combination of nil/non-nil annotation + label
	tables := []struct {
		name       string
		label      map[string]string
		annotation map[string]string
		result     map[string]string
	}{
		{"Empty label and annotations returns empty table map", nil, nil, map[string]string{}},
		{
			"Label with empty annotation returns valid map",
			map[string]string{"L1": "V1"},
			nil,
			map[string]string{"L1": "V1"}},
		{
			"Annotation with empty label returns valid map",
			nil,
			map[string]string{"A1": "V2"},
			map[string]string{fmt.Sprintf("%sA1", annotationLabelPrefix): "V2"}},
		{
			"Label and annotation provided returns valid combined map",
			map[string]string{"L1": "V1"},
			map[string]string{"A1": "V2"},
			map[string]string{
				"L1": "V1",
				fmt.Sprintf("%sA1", annotationLabelPrefix): "V2",
			},
		},
	}

	for _, tc := range tables {

		t.Run(tc.name, func(t *testing.T) {
			request := &types.FunctionDeployment{
				Labels:      &tc.label,
				Annotations: &tc.annotation,
			}

			val, err := buildLabels(request)

			if err != nil {
				t.Fatalf("want: no error got: %v", err)
			}

			if !reflect.DeepEqual(val, tc.result) {
				t.Errorf("Want: %s, got: %s", val, tc.result)
			}
		})
	}
}

func Test_BuildLabels_WithAnnotationCollision(t *testing.T) {
	request := &types.FunctionDeployment{
		Labels: &map[string]string{
			"function_name": "echo",
			fmt.Sprintf("%scurrent-time", annotationLabelPrefix): "Wed 25 Jul 06:41:43 BST 2018",
		},
		Annotations: &map[string]string{"current-time": "Wed 25 Jul 06:41:43 BST 2018"},
	}

	val, err := buildLabels(request)

	if err == nil {
		t.Errorf("Expected an error, got %d values", len(val))
	}

}

// Test generated using Keploy
func TestValidateSecrets_MissingSecret_ReturnsError(t *testing.T) {
	secretMountPath := "/tmp/secrets"
	secrets := []string{"missing-secret"}

	err := validateSecrets(secretMountPath, secrets)
	if err == nil {
		t.Fatalf("Expected an error, but got nil")
	}

	expectedError := "unable to find secret: missing-secret"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Unexpected error message: got %v want %v", err.Error(), expectedError)
	}
}

// Test generated using Keploy
func TestPrepareEnv_EmptyEnvProcess_NoEnvVars(t *testing.T) {
	envProcess := ""
	envVars := map[string]string{}

	result := prepareEnv(envProcess, envVars)

	expected := []string{}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result: got %v want %v", result, expected)
	}
}

// Test generated using Keploy
func TestPrepareEnv_EnvProcessSet_NoEnvVars(t *testing.T) {
	envProcess := "python index.py"
	envVars := map[string]string{}

	result := prepareEnv(envProcess, envVars)

	expected := []string{"fprocess=python index.py"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result: got %v want %v", result, expected)
	}
}

// Test generated using Keploy
func TestPrepareEnv_EnvProcessEmpty_EnvVarsSet(t *testing.T) {
	envProcess := ""
	envVars := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	result := prepareEnv(envProcess, envVars)

	expected := []string{"VAR1=value1", "VAR2=value2"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result: got %v want %v", result, expected)
	}
}

// Test generated using Keploy
func TestValidateSecrets_AllSecretsExist_NoError(t *testing.T) {
	secretMountPath := t.TempDir()
	secrets := []string{"secret1", "secret2"}

	// Create dummy secret files
	for _, secret := range secrets {
		secretPath := filepath.Join(secretMountPath, secret)
		err := os.WriteFile(secretPath, []byte("dummy data"), 0600)
		if err != nil {
			t.Fatalf("Failed to create secret file: %v", err)
		}
	}

	err := validateSecrets(secretMountPath, secrets)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}
