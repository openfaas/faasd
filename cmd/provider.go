package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/containerd/containerd"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faasd/pkg/cninetwork"
	faasdlogs "github.com/openfaas/faasd/pkg/logs"
	"github.com/openfaas/faasd/pkg/provider/config"
	"github.com/openfaas/faasd/pkg/provider/handlers"
	"github.com/spf13/cobra"
)

func makeProviderCmd() *cobra.Command {
	var command = &cobra.Command{
		Use:   "provider",
		Short: "Run the faasd-provider",
	}

	command.Flags().String("pull-policy", "Always", `Set to "Always" to force a pull of images upon deployment, or "IfNotPresent" to try to use a cached image.`)

	command.RunE = func(_ *cobra.Command, _ []string) error {

		pullPolicy, flagErr := command.Flags().GetString("pull-policy")
		if flagErr != nil {
			return flagErr
		}

		alwaysPull := false
		if pullPolicy == "Always" {
			alwaysPull = true
		}

		config, providerConfig, err := config.ReadFromEnv(types.OsEnv{})
		if err != nil {
			return err
		}

		log.Printf("faasd-provider starting..\tService Timeout: %s\n", config.WriteTimeout.String())

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		writeHostsErr := ioutil.WriteFile(path.Join(wd, "hosts"),
			[]byte(`127.0.0.1	localhost`), workingDirectoryPermission)

		if writeHostsErr != nil {
			return fmt.Errorf("cannot write hosts file: %s", writeHostsErr)
		}

		writeResolvErr := ioutil.WriteFile(path.Join(wd, "resolv.conf"),
			[]byte(`nameserver 8.8.8.8`), workingDirectoryPermission)

		if writeResolvErr != nil {
			return fmt.Errorf("cannot write resolv.conf file: %s", writeResolvErr)
		}

		cni, err := cninetwork.InitNetwork()
		if err != nil {
			return err
		}

		client, err := containerd.New(providerConfig.Sock)
		if err != nil {
			return err
		}

		defer client.Close()
		wg := sync.WaitGroup{}
		wg.Add(1)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		go func() {

			log.Printf("faasd-provider: waiting for SIGTERM or SIGINT to stop.\n")
			<-sig

			log.Printf("Signal received.. shutting down provider and functions.\n")

			functions, err := handlers.ListFunctions(client)
			if err != nil {
				log.Printf("[Read] error listing functions. Error: %s\n", err)
			}

			for _, function := range functions {
				log.Printf("Stopping function %s.\n", function.Name)
				handlers.ScaleFunction(client, cni, function.Name, 0)
			}
			fmt.Println("Finished scaling functions to 0.")
			wg.Done()
			// Shouldn't do this since bootstrap.Serve won't exit gracefully
			os.Exit(0)
		}()

		invokeResolver := handlers.NewInvokeResolver(client)

		userSecretPath := path.Join(wd, "secrets")

		bootstrapHandlers := types.FaaSHandlers{
			FunctionProxy:        proxy.NewHandlerFunc(*config, invokeResolver),
			DeleteHandler:        handlers.MakeDeleteHandler(client, cni),
			DeployHandler:        handlers.MakeDeployHandler(client, cni, userSecretPath, alwaysPull),
			FunctionReader:       handlers.MakeReadHandler(client),
			ReplicaReader:        handlers.MakeReplicaReaderHandler(client),
			ReplicaUpdater:       handlers.MakeReplicaUpdateHandler(client, cni),
			UpdateHandler:        handlers.MakeUpdateHandler(client, cni, userSecretPath, alwaysPull),
			HealthHandler:        func(w http.ResponseWriter, r *http.Request) {},
			InfoHandler:          handlers.MakeInfoHandler(Version, GitCommit),
			ListNamespaceHandler: listNamespaces(),
			SecretHandler:        handlers.MakeSecretHandler(client, userSecretPath),
			LogHandler:           logs.NewLogHandlerFunc(faasdlogs.New(), config.ReadTimeout),
		}

		log.Printf("Listening on TCP port: %d\n", *config.TCPPort)
		bootstrap.Serve(&bootstrapHandlers, config)

		wg.Wait()
		return nil
	}

	return command
}

func listNamespaces() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		list := []string{""}
		out, _ := json.Marshal(list)
		w.Write(out)
	}
}
