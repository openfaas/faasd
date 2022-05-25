// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package bootstrap

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
)

// NameExpression for a function / service
const NameExpression = "-a-zA-Z_0-9."

var r *mux.Router

// Mark this as a Golang "package"
func init() {
	r = mux.NewRouter()
}

// Router gives access to the underlying router for when new routes need to be added.
func Router() *mux.Router {
	return r
}

// Serve load your handlers into the correct OpenFaaS route spec. This function is blocking.
func Serve(handlers *types.FaaSHandlers, config *types.FaaSConfig) {

	if config.EnableBasicAuth {
		reader := auth.ReadBasicAuthFromDisk{
			SecretMountPath: config.SecretMountPath,
		}

		credentials, err := reader.Read()
		if err != nil {
			log.Fatal(err)
		}

		handlers.FunctionReader = auth.DecorateWithBasicAuth(handlers.FunctionReader, credentials)
		handlers.DeployHandler = auth.DecorateWithBasicAuth(handlers.DeployHandler, credentials)
		handlers.DeleteHandler = auth.DecorateWithBasicAuth(handlers.DeleteHandler, credentials)
		handlers.UpdateHandler = auth.DecorateWithBasicAuth(handlers.UpdateHandler, credentials)
		handlers.ReplicaReader = auth.DecorateWithBasicAuth(handlers.ReplicaReader, credentials)
		handlers.ReplicaUpdater = auth.DecorateWithBasicAuth(handlers.ReplicaUpdater, credentials)
		handlers.InfoHandler = auth.DecorateWithBasicAuth(handlers.InfoHandler, credentials)
		handlers.SecretHandler = auth.DecorateWithBasicAuth(handlers.SecretHandler, credentials)
		handlers.LogHandler = auth.DecorateWithBasicAuth(handlers.LogHandler, credentials)
	}

	// System (auth) endpoints
	r.HandleFunc("/system/functions", handlers.FunctionReader).Methods(http.MethodGet)
	r.HandleFunc("/system/functions", handlers.DeployHandler).Methods(http.MethodPost)
	r.HandleFunc("/system/functions", handlers.DeleteHandler).Methods(http.MethodDelete)
	r.HandleFunc("/system/functions", handlers.UpdateHandler).Methods(http.MethodPut)

	r.HandleFunc("/system/function/{name:["+NameExpression+"]+}", handlers.ReplicaReader).Methods(http.MethodGet)
	r.HandleFunc("/system/scale-function/{name:["+NameExpression+"]+}", handlers.ReplicaUpdater).Methods(http.MethodPost)
	r.HandleFunc("/system/info", handlers.InfoHandler).Methods(http.MethodGet)

	r.HandleFunc("/system/secrets", handlers.SecretHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
	r.HandleFunc("/system/logs", handlers.LogHandler).Methods(http.MethodGet)

	r.HandleFunc("/system/namespaces", handlers.ListNamespaceHandler).Methods(http.MethodGet)

	// Open endpoints
	r.HandleFunc("/function/{name:["+NameExpression+"]+}", handlers.FunctionProxy)
	r.HandleFunc("/function/{name:["+NameExpression+"]+}/", handlers.FunctionProxy)
	r.HandleFunc("/function/{name:["+NameExpression+"]+}/{params:.*}", handlers.FunctionProxy)

	if handlers.HealthHandler != nil {
		r.HandleFunc("/healthz", handlers.HealthHandler).Methods(http.MethodGet)
	}

	readTimeout := config.ReadTimeout
	writeTimeout := config.WriteTimeout

	port := 8080
	if config.TCPPort != nil {
		port = *config.TCPPort
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1MB - can be overridden by setting Server.MaxHeaderBytes.
		Handler:        r,
	}

	log.Fatal(s.ListenAndServe())
}
