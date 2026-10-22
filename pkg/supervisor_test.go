package pkg

import (
	"path"
	"path/filepath"
	"reflect"
	"testing"

	compose "github.com/compose-spec/compose-go/types"
)

func Test_ParseCompose(t *testing.T) {

	wd := "testdata"
	mountBase, err := filepath.Abs(wd)
	if err != nil {
		t.Fatal(err)
	}

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
					Src:  path.Join(mountBase, "secrets", "basic-auth-password"),
					Dest: path.Join("/run/secrets", "basic-auth-password"),
				},
				{
					Src:  path.Join(mountBase, "secrets", "basic-auth-user"),
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
					Src:  path.Join(mountBase, "prometheus.yml"),
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
					Src:  path.Join(mountBase, "secrets", "basic-auth-password"),
					Dest: path.Join("/run/secrets", "basic-auth-password"),
				},
				{
					Src:  path.Join(mountBase, "secrets", "basic-auth-user"),
					Dest: path.Join("/run/secrets", "basic-auth-user"),
				},
			},
			Caps:      []string{"CAP_NET_RAW"},
			DependsOn: []string{"basic-auth-plugin", "nats", "prometheus"},
			Ports:     []ServicePort{{Port: 8080, TargetPort: 8080}},
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
					Src:  path.Join(mountBase, "secrets", "basic-auth-password"),
					Dest: path.Join("/run/secrets", "basic-auth-password"),
				},
				{
					Src:  path.Join(mountBase, "secrets", "basic-auth-user"),
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
			equalStringSlice(t, exp.DependsOn, service.DependsOn)
			if !reflect.DeepEqual(exp.Ports, service.Ports) {
				t.Fatalf("want ports %+v, got %+v", exp.Ports, service.Ports)
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
			t.Fatalf("unexpected value at postition %d: want %v, got %v", i, want[i], found[i])
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

func Test_ConvertPorts(t *testing.T) {
	tests := []struct {
		name      string
		published string
		want      uint32
		wantErr   bool
	}{
		{name: "published", published: "8080", want: 8080},
		{name: "unpublished"},
		{name: "invalid", published: "invalid", wantErr: true},
		{name: "out of range", published: "65536", wantErr: true},
		{name: "reversed range", published: "8082-8080", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := convertPorts([]compose.ServicePortConfig{{Published: tt.published, Target: 80, HostIP: "127.0.0.1"}})
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr {
				return
			}
			want := []ServicePort{{Port: tt.want, TargetPort: 80, HostIP: "127.0.0.1"}}
			if !reflect.DeepEqual(ports, want) {
				t.Fatalf("want %+v, got %+v", want, ports)
			}
		})
	}
}

func Test_LoadComposeFile_Arm64Interpolation(t *testing.T) {
	config, err := LoadComposeFileWithArch("testdata", "docker-compose.yaml", testArchGetter("aarch64", "Linux"))
	if err != nil {
		t.Fatal(err)
	}
	for _, service := range config.Services {
		if service.Name == "gateway" {
			if service.Image != "docker.io/openfaas/gateway:0.18.17-arm64" {
				t.Fatalf("unexpected image: %s", service.Image)
			}
			return
		}
	}
	t.Fatal("gateway missing")
}

func Test_ParseCompose_PortRanges(t *testing.T) {
	for _, tc := range []struct {
		ports string
		want  []ServicePort
	}{
		{"8080-8082:80-82", []ServicePort{{Port: 8080, TargetPort: 80}, {Port: 8081, TargetPort: 81}, {Port: 8082, TargetPort: 82}}},
		{"8080-8082:80", []ServicePort{{Port: 8080, TargetPort: 80}, {Port: 8081, TargetPort: 80}, {Port: 8082, TargetPort: 80}}},
	} {
		t.Run(tc.ports, func(t *testing.T) {
			ports, err := compose.ParsePortConfig(tc.ports)
			if err != nil {
				t.Fatal(err)
			}
			services, err := ParseCompose(&compose.Project{Services: compose.Services{{Name: "test", Ports: ports}}})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(services[0].Ports, tc.want) {
				t.Fatalf("want %+v, got %+v", tc.want, services[0].Ports)
			}
		})
	}
}
