package cmd

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/alexellis/faasd/pkg"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start faasd",
	RunE:  runUp,
}

func runUp(_ *cobra.Command, _ []string) error {

	wd, _ := os.Getwd()
	svcs := []pkg.Service{
		// pkg.Service{
		// 	Name:  "faas-containerd",
		// 	Env:   []string{"snapshotter=overlayfs"},
		// 	Image: "docker.io/alexellis2/faas-containerd:0.3.2",
		// 	Mounts: []pkg.Mount{
		// 		pkg.Mount{
		// 			Src:  "/run/containerd/containerd.sock",
		// 			Dest: "/run/containerd/containerd.sock",
		// 		},
		// 	},
		// 	Caps: []string{"CAP_SYS_ADMIN", "CAP_NET_RAW"},
		// },
		pkg.Service{
			Name:  "prometheus",
			Env:   []string{},
			Image: "docker.io/prom/prometheus:v2.14.0",
			Mounts: []pkg.Mount{
				pkg.Mount{
					Src:  path.Join(wd, "prometheus.yml"),
					Dest: "/etc/prometheus/prometheus.yml",
				},
			},
			Caps: []string{"CAP_NET_RAW"},
		},
		pkg.Service{
			Name: "gateway",
			Env: []string{
				"basic_auth=false",
				"functions_provider_url=http://faas-containerd:8081/",
				"direct_functions=false",
				"read_timeout=60s",
				"write_timeout=60s",
				"upstream_timeout=65s",
			},
			Image:  "docker.io/openfaas/gateway:0.17.4",
			Mounts: []pkg.Mount{},
			Caps:   []string{"CAP_NET_RAW"},
		},
	}

	start := time.Now()
	supervisor, err := pkg.NewSupervisor("/run/containerd/containerd.sock")
	if err != nil {
		return err
	}

	log.Printf("Supervisor created in: %s\n", time.Since(start).String())

	start = time.Now()

	err = supervisor.Start(svcs)

	if err != nil {
		return err
	}

	defer supervisor.Close()

	log.Printf("Supervisor init done in: %s\n", time.Since(start).String())

	time.Sleep(time.Minute * 5)

	return nil
}
