package handlers

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"reflect"
	"testing"
)

func Test_BuildLabelsAndAnnotationsFromServiceSpec_Annotations(t *testing.T) {
	container := map[string]string{
		"qwer": "ty",
		"dvor": "ak",
		fmt.Sprintf("%scurrent-time", annotationLabelPrefix): "5 Nov 20:10:20 PST 1955",
		fmt.Sprintf("%sfuture-time", annotationLabelPrefix):  "21 Oct 20:10:20 PST 2015",
	}

	labels, annotation := buildLabelsAndAnnotations(container)

	if len(labels) != 2 {
		t.Errorf("want: %d labels got: %d", 2, len(labels))
	}

	if len(annotation) != 2 {
		t.Errorf("want: %d annotation got: %d", 1, len(annotation))
	}

	if _, ok := annotation["current-time"]; !ok {
		t.Errorf("want: '%s' entry in annotation map got: key not found", "current-time")
	}
}

func Test_SplitMountToSecrets(t *testing.T) {
	type test struct {
		Name     string
		Input    []specs.Mount
		Expected []string
	}
	tests := []test{
		{Name: "No matching openfaas secrets", Input: []specs.Mount{{Destination: "/foo/"}}, Expected: []string{}},
		{Name: "No Mounts", Input: []specs.Mount{{Destination: "/foo/"}}, Expected: []string{}},
		{Name: "One Mounts IS secret", Input: []specs.Mount{{Destination: "/var/openfaas/secrets/secret1"}}, Expected: []string{"secret1"}},
		{Name: "Multiple Mounts 1 secret", Input: []specs.Mount{{Destination: "/var/openfaas/secrets/secret1"}, {Destination: "/some/other/path"}}, Expected: []string{"secret1"}},
		{Name: "Multiple Mounts all secrets", Input: []specs.Mount{{Destination: "/var/openfaas/secrets/secret1"}, {Destination: "/var/openfaas/secrets/secret2"}}, Expected: []string{"secret1", "secret2"}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			got := readSecretsFromMounts(tc.Input)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Fatalf("expected %s, got %s", tc.Expected, got)
			}
		})
	}
}
