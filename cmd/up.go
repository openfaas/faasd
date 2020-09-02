package cmd

import (
	"fmt"
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
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/openfaas/faasd/pkg"
)

// upConfig are the CLI flags used by the `faasd up` command to deploy the faasd service
type upConfig struct {
	// composeFilePath is the path to the compose file specifying the faasd service configuration
	// See https://compose-spec.io/ for more information about the spec,
	//
	// currently, this must be the name of a file in workingDir, which is set to the value of
	// `faasdwd = /var/lib/faasd`
	composeFilePath string

	// working directory to assume the compose file is in, should be faasdwd.
	// this is not configurable but may be in the future.
	workingDir string
}

func init() {
	configureUpFlags(upCmd.Flags())
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start faasd",
	RunE:  runUp,
}

func runUp(cmd *cobra.Command, _ []string) error {

	printVersion()

	cfg, err := parseUpFlags(cmd)
	if err != nil {
		return err
	}

	services, err := loadServiceDefinition(cfg)
	if err != nil {
		return err
	}

	basicAuthErr := makeBasicAuthFiles(path.Join(cfg.workingDir, "secrets"))
	if basicAuthErr != nil {
		return errors.Wrap(basicAuthErr, "cannot create basic-auth-* files")
	}

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
	go proxy.Start(gatewayURLChan, proxyDoneCh, "Gateway")

	prometheusURLChan := make(chan string, 1)
	prometheusProxyPort := 9090
	prometheusProxy := pkg.NewProxy(prometheusProxyPort, timeout)
	go prometheusProxy.Start(prometheusURLChan, proxyDoneCh, "Prometheus")

	go func() {
		time.Sleep(3 * time.Second)

		fileData, fileErr := ioutil.ReadFile(path.Join(cfg.workingDir, "hosts"))
		if fileErr != nil {
			log.Println(fileErr)
			return
		}

		host := ""
		prometheusHost := ""
		lines := strings.Split(string(fileData), "\n")
		for _, line := range lines {
			if strings.Index(line, "gateway") > -1 {
				host = line[:strings.Index(line, "\t")]
			}

			if strings.Index(line, "prometheus") > -1 {
				prometheusHost = line[:strings.Index(line, "\t")]
			}
		}
		log.Printf("[up] Sending %s to Gateway proxy\n", host)
		gatewayURLChan <- host + ":8080"
		log.Printf("[up] Sending %s to Prometheus proxy\n", prometheusHost)
		prometheusURLChan <- prometheusHost + ":9090"
		close(gatewayURLChan)
		close(prometheusURLChan)
	}()

	wg.Wait()
	return nil
}

func makeBasicAuthFiles(wd string) error {

	pwdFile := path.Join(wd, "basic-auth-password")
	authPassword, err := password.Generate(63, 10, 0, false, true)

	if err != nil {
		return err
	}

	err = makeFile(pwdFile, authPassword)
	if err != nil {
		return err
	}

	userFile := path.Join(wd, "basic-auth-user")
	err = makeFile(userFile, "admin")
	if err != nil {
		return err
	}

	return nil
}

// makeFile will create a file with the specified content if it does not exist yet.
// if the file already exists, the method is a noop.
func makeFile(filePath, fileContents string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		log.Printf("File exists: %q\n", filePath)
		return nil
	} else if os.IsNotExist(err) {
		log.Printf("Writing to: %q\n", filePath)
		return ioutil.WriteFile(filePath, []byte(fileContents), workingDirectoryPermission)
	} else {
		return err
	}
}

// load the docker compose file and then parse it as supervisor Services
// the logic for loading the compose file comes from the compose reference implementation
// https://github.com/compose-spec/compose-ref/blob/master/compose-ref.go#L353
func loadServiceDefinition(cfg upConfig) ([]pkg.Service, error) {

	serviceConfig, err := pkg.LoadComposeFile(cfg.workingDir, cfg.composeFilePath)
	if err != nil {
		return nil, err
	}

	return pkg.ParseCompose(serviceConfig)
}

// ConfigureUpFlags will define the flags for the `faasd up` command. The flag struct, configure, and
// parse are split like this to simplify testability.
func configureUpFlags(flags *flag.FlagSet) {
	flags.StringP("file", "f", "docker-compose.yaml", "compose file specifying the faasd service configuration")
}

// ParseUpFlags will load the flag values into an upFlags object. Errors will be underlying
// Get errors from the pflag library.
func parseUpFlags(cmd *cobra.Command) (upConfig, error) {
	parsed := upConfig{}
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return parsed, errors.Wrap(err, "can not parse compose file path flag")
	}

	parsed.composeFilePath = path
	parsed.workingDir = faasdwd
	return parsed, err
}
