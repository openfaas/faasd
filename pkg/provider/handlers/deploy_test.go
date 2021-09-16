package handlers

import (
	"fmt"
	"reflect"
	"testing"

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
