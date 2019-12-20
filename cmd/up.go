package cmd

import (
	"log"
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

	svcs := []pkg.Service{
		pkg.Service{
			Name:  "faas-containerd",
			Env:   []string{},
			Image: "docker.io/alexellis2/faas-containerd:0.2.0",
			Mounts: []pkg.Mount{
				pkg.Mount{
					Src:  "/run/containerd/containerd.sock",
					Dest: "/run/containerd/containerd.sock",
				},
			},
		},
		pkg.Service{
			Name:   "gateway",
			Env:    []string{},
			Image:  "docker.io/openfaas/gateway:0.18.7",
			Mounts: []pkg.Mount{},
		},
		pkg.Service{
			Name:   "prometheus",
			Env:    []string{},
			Image:  "docker.io/prom/prometheus:v2.14.0",
			Mounts: []pkg.Mount{},
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
