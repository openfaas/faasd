package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/containerd/containerd"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	"github.com/openfaas/faas-provider/types"
	faasd "github.com/openfaas/faasd/pkg"
	"github.com/openfaas/faasd/pkg/cninetwork"
	faasdlogs "github.com/openfaas/faasd/pkg/logs"
	"github.com/openfaas/faasd/pkg/provider/config"
	"github.com/openfaas/faasd/pkg/provider/handlers"
	"github.com/spf13/cobra"
)

const secretDirPermission = 0755

type ResponseCapture struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		status:         200,
	}
}

func (rc *ResponseCapture) WriteHeader(statusCode int) {
	rc.status = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *ResponseCapture) GetBody() []byte {
	return rc.body.Bytes()
}

func makeProviderCmd() *cobra.Command {
	var command = &cobra.Command{
		Use:   "provider",
		Short: "Run the faasd-provider",
	}

	command.RunE = runProviderE
	command.PreRunE = preRunE

	return command
}

func runProviderE(cmd *cobra.Command, _ []string) error {

	config, providerConfig, err := config.ReadFromEnv(types.OsEnv{})
	if err != nil {
		return err
	}

	log.Printf("faasd-provider starting..\tService Timeout: %s\n", config.WriteTimeout.String())
	printVersion()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(wd, "hosts"),
		[]byte(`127.0.0.1	localhost`), workingDirectoryPermission); err != nil {
		return fmt.Errorf("cannot write hosts file: %s", err)
	}

	if err := os.WriteFile(path.Join(wd, "resolv.conf"),
		[]byte(`nameserver 8.8.8.8
nameserver 8.8.4.4`), workingDirectoryPermission); err != nil {
		return fmt.Errorf("cannot write resolv.conf file: %s", err)
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

	invokeResolver := handlers.NewInvokeResolver(client)

	baseUserSecretsPath := path.Join(wd, "secrets")
	if err := moveSecretsToDefaultNamespaceSecrets(
		baseUserSecretsPath,
		faasd.DefaultFunctionNamespace); err != nil {
		return err
	}

	alwaysPull := true
	bootstrapHandlers := types.FaaSHandlers{
		FunctionProxy:   httpHeaderMiddlewareWithCapture(proxy.NewHandlerFunc(*config, invokeResolver, false)),
		DeleteFunction:  httpHeaderMiddleware(handlers.MakeDeleteHandler(client, cni)),
		DeployFunction:  httpHeaderMiddleware(handlers.MakeDeployHandler(client, cni, baseUserSecretsPath, alwaysPull)),
		FunctionLister:  httpHeaderMiddleware(handlers.MakeReadHandler(client)),
		FunctionStatus:  httpHeaderMiddleware(handlers.MakeReplicaReaderHandler(client)),
		ScaleFunction:   httpHeaderMiddleware(handlers.MakeReplicaUpdateHandler(client, cni)),
		UpdateFunction:  httpHeaderMiddleware(handlers.MakeUpdateHandler(client, cni, baseUserSecretsPath, alwaysPull)),
		Health:          httpHeaderMiddleware(func(w http.ResponseWriter, r *http.Request) {}),
		Info:            httpHeaderMiddleware(handlers.MakeInfoHandler(faasd.Version, faasd.GitCommit)),
		ListNamespaces:  httpHeaderMiddleware(handlers.MakeNamespacesLister(client)),
		Secrets:         httpHeaderMiddleware(handlers.MakeSecretHandler(client.NamespaceService(), baseUserSecretsPath)),
		Logs:            httpHeaderMiddleware(logs.NewLogHandlerFunc(faasdlogs.New(), config.ReadTimeout)),
		MutateNamespace: httpHeaderMiddleware(handlers.MakeMutateNamespace(client)),
	}

	log.Printf("Listening on: 0.0.0.0:%d", *config.TCPPort)
	bootstrap.Serve(cmd.Context(), &bootstrapHandlers, config)
	return nil
}

/*
* Mutiple namespace support was added after release 0.13.0
* Function will help users to migrate on multiple namespace support of faasd
 */
func moveSecretsToDefaultNamespaceSecrets(baseSecretPath string, defaultNamespace string) error {
	newSecretPath := path.Join(baseSecretPath, defaultNamespace)

	err := ensureSecretsDir(newSecretPath)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(baseSecretPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {

			newPath := path.Join(newSecretPath, f.Name())

			// A non-nil error means the file wasn't found in the
			// destination path
			if _, err := os.Stat(newPath); err != nil {
				oldPath := path.Join(baseSecretPath, f.Name())

				if err := copyFile(oldPath, newPath); err != nil {
					return err
				}

				log.Printf("[Migration] Copied %s to %s", oldPath, newPath)
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s failed %w", src, err)
	}
	defer inputFile.Close()

	outputFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_APPEND, secretDirPermission)
	if err != nil {
		return fmt.Errorf("opening %s failed %w", dst, err)
	}
	defer outputFile.Close()

	// Changed from os.Rename due to issue in #201
	if _, err := io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("writing into %s failed %w", outputFile.Name(), err)
	}

	return nil
}

func httpHeaderMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-OpenFaaS-EULA", "openfaas-ce")
		next.ServeHTTP(w, r)
	}
}

func httpHeaderMiddlewareWithCapture(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the EULA header
		w.Header().Set("X-OpenFaaS-EULA", "openfaas-ce")
		w.Write([]byte("\nThis is the message appended by me to test the reading on CMD line....\n"))

		capture := NewResponseCapture(w)
		next.ServeHTTP(capture, r)
		responseBody := capture.GetBody()
		log.Printf("Response body: %s", responseBody)
	}
}

func testMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Harsh", "true")
		w.Write([]byte("This is a test\n"))
		next.ServeHTTP(w, r)
	}
}
