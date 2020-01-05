package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/alexellis/faasd/pkg"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start faasd",
	RunE:  runUp,
}

// defaultCNIConf is a CNI configuration that enables network access to containers (docker-bridge style)
var defaultCNIConf = fmt.Sprintf(`
{
    "cniVersion": "0.4.0",
    "name": "%s",
    "plugins": [
      {
        "type": "bridge",
        "bridge": "%s",
        "isGateway": true,
        "ipMasq": true,
        "ipam": {
            "type": "host-local",
            "subnet": "%s",
            "routes": [
                { "dst": "0.0.0.0/0" }
            ]
        }
      },
      {
        "type": "firewall"
      }
    ]
}
`, pkg.DefaultNetworkName, pkg.DefaultBridgeName, pkg.DefaultSubnet)

const secretMountDir = "/run/secrets"

func runUp(_ *cobra.Command, _ []string) error {

	wd, _ := os.Getwd()

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
	case "armv7l":
		clientSuffix = "-armhf"
		break
	case "arm64":
	case "aarch64":
		clientSuffix = "-arm64"
	}

	if basicAuthErr := makeBasicAuthFiles(path.Join(path.Join(faasdwd, "secrets"))); basicAuthErr != nil {
		return errors.Wrap(basicAuthErr, "cannot create basic-auth-* files")
	}

	if makeNetworkErr := makeNetworkConfig(); makeNetworkErr != nil {
		return errors.Wrap(makeNetworkErr, "error creating network config")
	}

	services := makeServiceDefinitions(wd, clientSuffix)

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
	timeout := time.Second * 60
	proxyDoneCh := make(chan bool)

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

		// Close proxy
		proxyDoneCh <- true
		time.AfterFunc(shutdownTimeout, func() {
			wg.Done()
		})
	}()

	gatewayURLChan := make(chan string, 1)
	proxyPort := 8080
	proxy := pkg.NewProxy(proxyPort, timeout)
	go proxy.Start(gatewayURLChan, proxyDoneCh)

	go func() {
		wd, _ := os.Getwd()

		time.Sleep(3 * time.Second)

		fileData, fileErr := ioutil.ReadFile(path.Join(wd, "hosts"))
		if fileErr != nil {
			log.Println(fileErr)
			return
		}
		host := ""
		lines := strings.Split(string(fileData), "\n")
		for _, line := range lines {
			if strings.Index(line, "gateway") > -1 {
				host = line[:strings.Index(line, "\t")]
			}
		}
		log.Printf("[up] Sending %s to proxy\n", host)
		gatewayURLChan <- host + ":8080"
		close(gatewayURLChan)
	}()

	wg.Wait()
	return nil
}

func makeBasicAuthFiles(wd string) error {

	pwdFile := wd + "/basic-auth-password"
	authPassword, err := password.Generate(63, 10, 0, false, true)

	if err != nil {
		return err
	}

	err = makeFile(pwdFile, authPassword)
	if err != nil {
		return err
	}

	userFile := wd + "/basic-auth-user"
	err = makeFile(userFile, "admin")
	if err != nil {
		return err
	}

	return nil
}

func makeFile(filePath, fileContents string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		log.Printf("File exists: %q\n", filePath)
		return nil
	} else if os.IsNotExist(err) {
		log.Printf("Writing to: %q\n", filePath)
		return ioutil.WriteFile(filePath, []byte(fileContents), 0644)
	} else {
		return err
	}
}

func makeServiceDefinitions(wd, archSuffix string) []pkg.Service {

	return []pkg.Service{
		pkg.Service{
			Name:  "basic-auth-plugin",
			Image: "docker.io/openfaas/basic-auth-plugin:0.18.10" + archSuffix,
			Env: []string{
				"port=8080",
				"secret_mount_path=" + secretMountDir,
				"user_filename=basic-auth-user",
				"pass_filename=basic-auth-password",
			},
			Mounts: []pkg.Mount{
				pkg.Mount{
					Src:  path.Join(path.Join(wd, "secrets"), "basic-auth-password"),
					Dest: path.Join(secretMountDir, "basic-auth-password"),
				},
				pkg.Mount{
					Src:  path.Join(path.Join(wd, "secrets"), "basic-auth-user"),
					Dest: path.Join(secretMountDir, "basic-auth-user"),
				},
			},
			Caps: []string{"CAP_NET_RAW"},
			Args: nil,
		},
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
				"basic_auth=true",
				"functions_provider_url=http://faas-containerd:8081/",
				"direct_functions=false",
				"read_timeout=60s",
				"write_timeout=60s",
				"upstream_timeout=65s",
				"faas_nats_address=nats",
				"faas_nats_port=4222",
				"auth_proxy_url=http://basic-auth-plugin:8080/validate",
				"auth_proxy_pass_body=false",
				"secret_mount_path=" + secretMountDir,
			},
			Image: "docker.io/openfaas/gateway:0.18.8" + archSuffix,
			Mounts: []pkg.Mount{
				pkg.Mount{
					Src:  path.Join(path.Join(wd, "secrets"), "basic-auth-password"),
					Dest: path.Join(secretMountDir, "basic-auth-password"),
				},
				pkg.Mount{
					Src:  path.Join(path.Join(wd, "secrets"), "basic-auth-user"),
					Dest: path.Join(secretMountDir, "basic-auth-user"),
				},
			},
			Caps: []string{"CAP_NET_RAW"},
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
				"write_debug=false",
				"basic_auth=true",
				"secret_mount_path=" + secretMountDir,
			},
			Image: "docker.io/openfaas/queue-worker:0.9.0",
			Mounts: []pkg.Mount{
				pkg.Mount{
					Src:  path.Join(path.Join(wd, "secrets"), "basic-auth-password"),
					Dest: path.Join(secretMountDir, "basic-auth-password"),
				},
				pkg.Mount{
					Src:  path.Join(path.Join(wd, "secrets"), "basic-auth-user"),
					Dest: path.Join(secretMountDir, "basic-auth-user"),
				},
			},
			Caps: []string{"CAP_NET_RAW"},
		},
	}
}

func makeNetworkConfig() error {
	netConfig := path.Join(pkg.CNIConfDir, pkg.DefaultCNIConfFilename)
	log.Printf("Writing network config...\n")

	if !dirExists(pkg.CNIConfDir) {
		if err := os.MkdirAll(pkg.CNIConfDir, 0755); err != nil {
			return fmt.Errorf("cannot create directory: %s", pkg.CNIConfDir)
		}
	}

	if err := ioutil.WriteFile(netConfig, []byte(defaultCNIConf), 644); err != nil {
		return fmt.Errorf("cannot write network config: %s", pkg.DefaultCNIConfFilename)

	}
	return nil
}

func dirEmpty(dirname string) (b bool) {
	if !dirExists(dirname) {
		return
	}

	f, err := os.Open(dirname)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	// If the first file is EOF, the directory is empty
	if _, err = f.Readdir(1); err == io.EOF {
		b = true
	}
	return
}

func dirExists(dirname string) bool {
	exists, info := pathExists(dirname)
	if !exists {
		return false
	}

	return info.IsDir()
}

func pathExists(path string) (bool, os.FileInfo) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, info
}
