package handlers

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
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
	type testCase struct {
		Name  string
		Input []specs.Mount
		Want  []string
	}
	tests := []testCase{
		{Name: "No matching openfaas secrets", Input: []specs.Mount{{Destination: "/foo/"}}, Want: []string{}},
		{Name: "Nil mounts", Input: nil, Want: []string{}},
		{Name: "No Mounts", Input: []specs.Mount{{Destination: "/foo/"}}, Want: []string{}},
		{Name: "One Mounts IS secret", Input: []specs.Mount{{Destination: "/var/openfaas/secrets/secret1"}}, Want: []string{"secret1"}},
		{Name: "Multiple Mounts 1 secret", Input: []specs.Mount{{Destination: "/var/openfaas/secrets/secret1"}, {Destination: "/some/other/path"}}, Want: []string{"secret1"}},
		{Name: "Multiple Mounts all secrets", Input: []specs.Mount{{Destination: "/var/openfaas/secrets/secret1"}, {Destination: "/var/openfaas/secrets/secret2"}}, Want: []string{"secret1", "secret2"}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			got := readSecretsFromMounts(tc.Input)
			if !reflect.DeepEqual(got, tc.Want) {
				t.Fatalf("Want %s, got %s", tc.Want, got)
			}
		})
	}
}

func Test_ProcessEnvToEnvVars(t *testing.T) {
	type testCase struct {
		Name     string
		Input    []string
		Want     map[string]string
		fprocess string
	}
	tests := []testCase{
		{Name: "No matching EnvVars", Input: []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "fprocess=python index.py"}, Want: make(map[string]string), fprocess: "python index.py"},
		{Name: "No EnvVars", Input: []string{}, Want: make(map[string]string), fprocess: ""},
		{Name: "One EnvVar", Input: []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "fprocess=python index.py", "env=this"}, Want: map[string]string{"env": "this"}, fprocess: "python index.py"},
		{Name: "Multiple EnvVars", Input: []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "this=that", "env=var", "fprocess=python index.py"}, Want: map[string]string{"this": "that", "env": "var"}, fprocess: "python index.py"},
		{Name: "Nil EnvVars", Input: nil, Want: make(map[string]string)},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			got, fprocess := readEnvFromProcessEnv(tc.Input)
			if !reflect.DeepEqual(got, tc.Want) {
				t.Fatalf("Want: %s, got: %s", tc.Want, got)
			}

			if fprocess != tc.fprocess {
				t.Fatalf("Want fprocess: %s, got: %s", tc.fprocess, got)

			}
		})
	}
}

func Test_findNamespace(t *testing.T) {
	type testCase struct {
		Name            string
		foundNamespaces []string
		namespace       string
		Want            bool
	}
	tests := []testCase{
		{Name: "Namespace Found", namespace: "fn", foundNamespaces: []string{"fn", "openfaas-fn"}, Want: true},
		{Name: "namespace Not Found", namespace: "fn", foundNamespaces: []string{"openfaas-fn"}, Want: false},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			got := findNamespace(tc.namespace, tc.foundNamespaces)
			if got != tc.Want {
				t.Fatalf("Want %t, got %t", tc.Want, got)
			}
		})
	}
}

func Test_readMemoryLimitFromSpec(t *testing.T) {
	type testCase struct {
		Name string
		Spec *specs.Spec
		Want int64
	}
	testLimit := int64(64)
	tests := []testCase{
		{Name: "specs.Linux not found", Spec: &specs.Spec{Linux: nil}, Want: int64(0)},
		{Name: "specs.LinuxResource not found", Spec: &specs.Spec{Linux: &specs.Linux{Resources: nil}}, Want: int64(0)},
		{Name: "specs.LinuxMemory not found", Spec: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{Memory: nil}}}, Want: int64(0)},
		{Name: "specs.LinuxMemory.Limit not found", Spec: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{Memory: &specs.LinuxMemory{Limit: nil}}}}, Want: int64(0)},
		{Name: "Memory limit set as Want", Spec: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{Memory: &specs.LinuxMemory{Limit: &testLimit}}}}, Want: int64(64)},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			got := readMemoryLimitFromSpec(tc.Spec)
			if got != tc.Want {
				t.Fatalf("Want %d, got %d", tc.Want, got)
			}
		})
	}
}
