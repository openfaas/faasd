package pkg

import (
	"path"
	"reflect"
	"testing"
)

func Test_ParseCompose(t *testing.T) {
	wd := "testdata"
	expected := map[string]Service{
		"basic-auth-plugin": {
			Name:  "basic-auth-plugin",
			Image: "docker.io/openfaas/basic-auth-plugin:0.18.17",
			Env: []string{
				"pass_filename=basic-auth-password",
				"port=8080",
				"secret_mount_path=/run/secrets",
				"user_filename=basic-auth-user",
			},
			Mounts: []Mount{
				{
					Src:  path.Join(wd, "secrets", "basic-auth-password"),
					Dest: path.Join("/run/secrets", "basic-auth-password"),
				},
				{
					Src:  path.Join(wd, "secrets", "basic-auth-user"),
					Dest: path.Join("/run/secrets", "basic-auth-user"),
				},
			},
			Caps: []string{"CAP_NET_RAW"},
		},
		"nats": {
			Name:  "nats",
			Image: "docker.io/library/nats-streaming:0.11.2",
			Args:  []string{"/nats-streaming-server", "-m", "8222", "--store=memory", "--cluster_id=faas-cluster"},
		},
		"prometheus": {
			Name:  "prometheus",
			Image: "docker.io/prom/prometheus:v2.14.0",
			Mounts: []Mount{
				{
					Src:  path.Join(wd, "prometheus.yml"),
					Dest: "/etc/prometheus/prometheus.yml",
				},
			},
			Caps: []string{"CAP_NET_RAW"},
		},
		"gateway": {
			Name: "gateway",
			Env: []string{
				"auth_proxy_pass_body=false",
				"auth_proxy_url=http://basic-auth-plugin:8080/validate",
				"basic_auth=true",
				"direct_functions=false",
				"faas_nats_address=nats",
				"faas_nats_port=4222",
				"functions_provider_url=http://faasd-provider:8081/",
				"read_timeout=60s",
				"scale_from_zero=true",
				"secret_mount_path=/run/secrets",
				"upstream_timeout=65s",
				"write_timeout=60s",
			},
			Image: "docker.io/openfaas/gateway:0.18.17",
			Mounts: []Mount{
				{
					Src:  path.Join(wd, "secrets", "basic-auth-password"),
					Dest: path.Join("/run/secrets", "basic-auth-password"),
				},
				{
					Src:  path.Join(wd, "secrets", "basic-auth-user"),
					Dest: path.Join("/run/secrets", "basic-auth-user"),
				},
			},
			Caps: []string{"CAP_NET_RAW"},
		},
		"queue-worker": {
			Name: "queue-worker",
			Env: []string{
				"ack_wait=5m5s",
				"basic_auth=true",
				"faas_gateway_address=gateway",
				"faas_nats_address=nats",
				"faas_nats_port=4222",
				"gateway_invoke=true",
				"max_inflight=1",
				"secret_mount_path=/run/secrets",
				"write_debug=false",
			},
			Image: "docker.io/openfaas/queue-worker:0.11.2",
			Mounts: []Mount{
				{
					Src:  path.Join(wd, "secrets", "basic-auth-password"),
					Dest: path.Join("/run/secrets", "basic-auth-password"),
				},
				{
					Src:  path.Join(wd, "secrets", "basic-auth-user"),
					Dest: path.Join("/run/secrets", "basic-auth-user"),
				},
			},
			Caps: []string{"CAP_NET_RAW"},
		},
	}

	compose, err := LoadComposeFile(wd, "docker-compose.yaml")
	if err != nil {
		t.Fatalf("can not read docker-compose fixture: %s", err)
	}

	services, err := ParseCompose(compose)
	if err != nil {
		t.Fatalf("can not parse compose services: %s", err)
	}

	if len(services) != len(expected) {
		t.Fatalf("expected: %d services, got: %d", len(expected), len(services))
	}

	for _, service := range services {
		exp, ok := expected[service.Name]
		if !ok {
			t.Fatalf("unexpected service: %s", service.Name)
		}

		if service.Name != exp.Name {
			t.Fatalf("incorrect service Name:\n\texpected: %s,\n\tgot: %s", exp.Name, service.Name)
		}

		if service.Image != exp.Image {
			t.Fatalf("incorrect service Image:\n\texpected: %s,\n\tgot: %s", exp.Image, service.Image)
		}

		equalStringSlice(t, exp.Env, service.Env)
		equalStringSlice(t, exp.Caps, service.Caps)
		equalStringSlice(t, exp.Args, service.Args)

		if !reflect.DeepEqual(exp.Mounts, service.Mounts) {
			t.Fatalf("incorrect service Mounts:\n\texpected: %+v,\n\tgot: %+v", exp.Mounts, service.Mounts)
		}
	}
}

func equalStringSlice(t *testing.T, expected, found []string) {
	t.Helper()
	if (expected == nil) != (found == nil) {
		t.Fatalf("unexpected nil slice: expected %+v, got %+v", expected, found)
	}

	if len(expected) != len(found) {
		t.Fatalf("unequal slice length: expected %+v, got %+v", expected, found)
	}

	for i := range expected {
		if expected[i] != found[i] {
			t.Fatalf("unexpected value at postition %d: expected %s, got %s", i, expected[i], found[i])
		}
	}
}

func equalMountSlice(t *testing.T, expected, found []Mount) {
	t.Helper()
	if (expected == nil) != (found == nil) {
		t.Fatalf("unexpected nil slice: expected %+v, got %+v", expected, found)
	}

	if len(expected) != len(found) {
		t.Fatalf("unequal slice length: expected %+v, got %+v", expected, found)
	}

	for i := range expected {
		if !reflect.DeepEqual(expected[i], found[i]) {
			t.Fatalf("unexpected value at postition %d: expected %s, got %s", i, expected[i], found[i])
		}
	}
}

func Test_GetArchSuffix(t *testing.T) {
	cases := []struct {
		name      string
		expected  string
		foundArch string
		foundOS   string
		err       string
	}{
		{
			name:    "error if os is not linux",
			foundOS: "mac",
			err:     "you can only use faasd with Linux",
		},
		{
			name:      "x86 has no suffix",
			foundOS:   "Linux",
			foundArch: "x86_64",
			expected:  "",
		},
		{
			name:      "unknown arch has no suffix",
			foundOS:   "Linux",
			foundArch: "anything_else",
			expected:  "",
		},
		{
			name:      "armhf has armhf suffix",
			foundOS:   "Linux",
			foundArch: "armhf",
			expected:  "-armhf",
		},
		{
			name:      "armv7l has armhf suffix",
			foundOS:   "Linux",
			foundArch: "armv7l",
			expected:  "-armhf",
		},
		{
			name:      "arm64 has arm64 suffix",
			foundOS:   "Linux",
			foundArch: "arm64",
			expected:  "-arm64",
		},
		{
			name:      "aarch64 has arm64 suffix",
			foundOS:   "Linux",
			foundArch: "aarch64",
			expected:  "-arm64",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			suffix, err := GetArchSuffix(testArchGetter(tc.foundArch, tc.foundOS))
			if tc.err != "" && err == nil {
				t.Fatalf("expected error %s but got nil", tc.err)
			} else if tc.err != "" && err.Error() != tc.err {
				t.Fatalf("expected error %s, got %s", tc.err, err.Error())
			} else if tc.err == "" && err != nil {
				t.Fatalf("unexpected error %s", err.Error())
			}

			if suffix != tc.expected {
				t.Fatalf("expected suffix %s, got %s", tc.expected, suffix)
			}
		})
	}
}

func testArchGetter(arch, os string) ArchGetter {
	return func() (string, string) {
		return arch, os
	}
}
