package pkg

import (
	"path"
	"reflect"
	"testing"
)

func Test_ParseCompose(t *testing.T) {

	wd := "testdata"

	want := map[string]Service{
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
			Caps:      []string{"CAP_NET_RAW"},
			DependsOn: []string{"nats"},
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

	compose, err := LoadComposeFileWithArch(wd, "docker-compose.yaml", func() (string, string) { return "x86_64", "Linux" })
	if err != nil {
		t.Fatalf("can't read docker-compose file: %s", err)
	}

	services, err := ParseCompose(compose)
	if err != nil {
		t.Fatalf("can't parse compose services: %s", err)
	}

	if len(services) != len(want) {
		t.Fatalf("want: %d services, got: %d", len(want), len(services))
	}

	for _, service := range services {
		exp, ok := want[service.Name]

		if service.Name == "gateway" {
			if len(service.DependsOn) == 0 {
				t.Fatalf("gateway should have at least one depends_on entry")
			}
		}

		if !ok {
			t.Fatalf("incorrect service: %s", service.Name)
		}

		if service.Name != exp.Name {
			t.Fatalf("incorrect service Name:\n\twant: %s,\n\tgot: %s", exp.Name, service.Name)
		}

		if service.Image != exp.Image {
			t.Fatalf("incorrect service Image:\n\twant: %s,\n\tgot: %s", exp.Image, service.Image)
		}

		equalStringSlice(t, exp.Env, service.Env)
		equalStringSlice(t, exp.Caps, service.Caps)
		equalStringSlice(t, exp.Args, service.Args)

		if !reflect.DeepEqual(exp.Mounts, service.Mounts) {
			t.Fatalf("incorrect service Mounts:\n\twant: %+v,\n\tgot: %+v", exp.Mounts, service.Mounts)
		}
	}
}

func equalStringSlice(t *testing.T, want, found []string) {
	t.Helper()
	if (want == nil) != (found == nil) {
		t.Fatalf("unexpected nil slice: want %+v, got %+v", want, found)
	}

	if len(want) != len(found) {
		t.Fatalf("unequal slice length: want %+v, got %+v", want, found)
	}

	for i := range want {
		if want[i] != found[i] {
			t.Fatalf("unexpected value at postition %d: want %s, got %s", i, want[i], found[i])
		}
	}
}

func equalMountSlice(t *testing.T, want, found []Mount) {
	t.Helper()
	if (want == nil) != (found == nil) {
		t.Fatalf("unexpected nil slice: want %+v, got %+v", want, found)
	}

	if len(want) != len(found) {
		t.Fatalf("unequal slice length: want %+v, got %+v", want, found)
	}

	for i := range want {
		if !reflect.DeepEqual(want[i], found[i]) {
			t.Fatalf("unexpected value at postition %d: want %s, got %s", i, want[i], found[i])
		}
	}
}

func Test_GetArchSuffix(t *testing.T) {
	cases := []struct {
		name      string
		want      string
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
			want:      "",
		},
		{
			name:      "unknown arch has no suffix",
			foundOS:   "Linux",
			foundArch: "anything_else",
			want:      "",
		},
		{
			name:      "armhf has armhf suffix",
			foundOS:   "Linux",
			foundArch: "armhf",
			want:      "-armhf",
		},
		{
			name:      "armv7l has armhf suffix",
			foundOS:   "Linux",
			foundArch: "armv7l",
			want:      "-armhf",
		},
		{
			name:      "arm64 has arm64 suffix",
			foundOS:   "Linux",
			foundArch: "arm64",
			want:      "-arm64",
		},
		{
			name:      "aarch64 has arm64 suffix",
			foundOS:   "Linux",
			foundArch: "aarch64",
			want:      "-arm64",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			suffix, err := GetArchSuffix(testArchGetter(tc.foundArch, tc.foundOS))
			if tc.err != "" && err == nil {
				t.Fatalf("want error %s but got nil", tc.err)
			} else if tc.err != "" && err.Error() != tc.err {
				t.Fatalf("want error %s, got %s", tc.err, err.Error())
			} else if tc.err == "" && err != nil {
				t.Fatalf("unexpected error %s", err.Error())
			}

			if suffix != tc.want {
				t.Fatalf("want suffix %s, got %s", tc.want, suffix)
			}
		})
	}
}

func testArchGetter(arch, os string) ArchGetter {
	return func() (string, string) {
		return arch, os
	}
}
