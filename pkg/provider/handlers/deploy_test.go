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
		label      map[string]string
		annotation map[string]string
		result     map[string]string
	}{
		{nil, nil, map[string]string{}},
		{map[string]string{"L1": "V1"}, nil, map[string]string{"L1": "V1"}},
		{nil, map[string]string{"A1": "V2"}, map[string]string{fmt.Sprintf("%sA1", annotationLabelPrefix): "V2"}},
		{
			map[string]string{"L1": "V1"}, map[string]string{"A1": "V2"},
			map[string]string{"L1": "V1", fmt.Sprintf("%sA1", annotationLabelPrefix): "V2"},
		},
	}

	for _, pair := range tables {

		request := &types.FunctionDeployment{
			Labels:      &pair.label,
			Annotations: &pair.annotation,
		}

		val, err := buildLabels(request)

		if err != nil {
			t.Fatalf("want: no error got: %v", err)
		}

		if !reflect.DeepEqual(val, pair.result) {
			t.Errorf("Got: %s, expected %s", val, pair.result)
		}
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

	fmt.Printf("%s, %s\n", val, err)
	if err == nil {
		t.Errorf("Expected an error, got %d values", len(val))
	}

}
