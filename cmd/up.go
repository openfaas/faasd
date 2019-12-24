package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/alexellis/faasd/pkg"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start faasd",
	RunE:  runUp,
}

func runUp(_ *cobra.Command, _ []string) error {

	clientArch, clientOS := env.GetClientArch()

	if clientOS != "Linux" {
		return fmt.Errorf("You can only use faasd on Linux")
	}
	clientSuffix := ""
	switch clientArch {
	case "x86_64":
		clientSuffix = ""
		break
	case "armhf":
	case "arm64":
		clientSuffix = clientArch
		break
	case "aarch64":
		clientSuffix = "arm64"
	}

	services := makeServiceDefinitions(clientSuffix)

	start := time.Now()
	supervisor, err := pkg.NewSupervisor("/run/containerd/containerd.sock")
	if err != nil {
		return err
	}

	log.Printf("Supervisor created in: %s\n", time.Since(start).String())

	start = time.Now()

	err = supervisor.Start(services)

	if err != nil {
		return err
	}

	defer supervisor.Close()

	log.Printf("Supervisor init done in: %s\n", time.Since(start).String())

	shutdownTimeout := time.Second * 1

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

		log.Printf("faasd: waiting for SIGTERM or SIGINT\n")
		<-sig

		log.Printf("Signal received.. shutting down server in %s\n", shutdownTimeout.String())
		err := supervisor.Remove(services)
		if err != nil {
			fmt.Println(err)
		}
		time.AfterFunc(shutdownTimeout, func() {
			wg.Done()
		})
	}()

	wg.Wait()
	return nil
}

func makeServiceDefinitions(archSuffix string) []pkg.Service {
	wd, _ := os.Getwd()

	return []pkg.Service{
		pkg.Service{
			Name:  "nats",
			Env:   []string{""},
			Image: "docker.io/library/nats-streaming:0.11.2",
			Caps:  []string{},
			Args:  []string{"/nats-streaming-server", "-m", "8222", "--store=memory", "--cluster_id=faas-cluster"},
		},
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
				"faas_nats_address=nats",
				"faas_nats_port=4222",
			},
			Image:  "docker.io/openfaas/gateway:0.18.7" + archSuffix,
			Mounts: []pkg.Mount{},
			Caps:   []string{"CAP_NET_RAW"},
		},
		pkg.Service{
			Name: "queue-worker",
			Env: []string{
				"faas_nats_address=nats",
				"faas_nats_port=4222",
				"gateway_invoke=true",
				"faas_gateway_address=gateway",
				"ack_wait=5m5s",
				"max_inflight=1",
				"faas_print_body=true",
			},
			Image:  "docker.io/openfaas/queue-worker:0.9.0",
			Mounts: []pkg.Mount{},
			Caps:   []string{"CAP_NET_RAW"},
		},
	}
}
