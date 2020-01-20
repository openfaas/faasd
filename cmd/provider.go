package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/alexellis/faasd/pkg/config"
	"github.com/alexellis/faasd/pkg/handlers"
	"github.com/containerd/containerd"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/proxy"
	"github.com/openfaas/faas-provider/types"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start-provider",
	Short: "Start faas-containerd provider",
	RunE:  startProvider,
}

func startProvider(_ *cobra.Command, _ []string) error {

	config, providerConfig, err := config.ReadFromEnv(types.OsEnv{})
	if err != nil {
		panic(err)
	}
	log.Printf("faas-containerd provider tarting..\tVersion: %s\tCommit: %s\tService Timeout: %s\n", Version, GitCommit, config.WriteTimeout.String())

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	writeHostsErr := ioutil.WriteFile(path.Join(wd, "hosts"),
		[]byte(`127.0.0.1	localhost`), 0644)

	if writeHostsErr != nil {
		log.Fatalln(fmt.Errorf("cannot write hosts file: %s", writeHostsErr).Error())
	}

	writeResolvErr := ioutil.WriteFile(path.Join(wd, "resolv.conf"),
		[]byte(`nameserver 8.8.8.8`), 0644)

	if writeResolvErr != nil {
		log.Fatalln(fmt.Errorf("cannot write resolv.conf file: %s", writeResolvErr).Error())
	}

	cni, err := handlers.InitNetwork()
	if err != nil {
		panic(err)
	}

	client, err := containerd.New(providerConfig.Sock)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	invokeResolver := handlers.NewInvokeResolver(client)

	bootstrapHandlers := types.FaaSHandlers{
		FunctionProxy:        proxy.NewHandlerFunc(*config, invokeResolver),
		DeleteHandler:        handlers.MakeDeleteHandler(client, cni),
		DeployHandler:        handlers.MakeDeployHandler(client, cni),
		FunctionReader:       handlers.MakeReadHandler(client),
		ReplicaReader:        handlers.MakeReplicaReaderHandler(client),
		ReplicaUpdater:       handlers.MakeReplicaUpdateHandler(client, cni),
		UpdateHandler:        handlers.MakeUpdateHandler(client, cni),
		HealthHandler:        func(w http.ResponseWriter, r *http.Request) {},
		InfoHandler:          handlers.MakeInfoHandler(Version, GitCommit),
		ListNamespaceHandler: listNamespaces(),
	}

	log.Printf("Listening on TCP port: %d\n", *config.TCPPort)
	bootstrap.Serve(&bootstrapHandlers, config)

	return nil
}

func listNamespaces() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		list := []string{""}
		out, _ := json.Marshal(list)
		w.Write(out)
	}
}
