// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// MakeHandler configures handlers for the OpenFaaS route specification using
// the supplied Prometheus registerer and gatherer.
func MakeHandler(handlers *types.FaaSHandlers, config *types.FaaSConfig,
	registerer prometheus.Registerer, gatherer prometheus.Gatherer) http.Handler {
	return makeHandler(mux.NewRouter(), handlers, config, registerer,
		promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
}

func makeHandler(router *mux.Router, handlers *types.FaaSHandlers, config *types.FaaSConfig,
	registerer prometheus.Registerer, metricsHandler http.Handler) http.Handler {
	configuredHandlers := *handlers
	handlers = &configuredHandlers

	if handlers.ListNamespaces == nil {
		handlers.ListNamespaces = func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Feature not implemented in this version of OpenFaaS", http.StatusNotImplemented)
		}
	}

	if config.EnableBasicAuth {
		reader := auth.ReadBasicAuthFromDisk{
			SecretMountPath: config.SecretMountPath,
		}

		credentials, err := reader.Read()
		if err != nil {
			log.Fatalf("failed to read basic auth credentials: %s", err)
		}

		handlers.FunctionLister = auth.DecorateWithBasicAuth(handlers.FunctionLister, credentials)
		handlers.DeployFunction = auth.DecorateWithBasicAuth(handlers.DeployFunction, credentials)
		handlers.DeleteFunction = auth.DecorateWithBasicAuth(handlers.DeleteFunction, credentials)
		handlers.UpdateFunction = auth.DecorateWithBasicAuth(handlers.UpdateFunction, credentials)
		handlers.FunctionStatus = auth.DecorateWithBasicAuth(handlers.FunctionStatus, credentials)
		handlers.ScaleFunction = auth.DecorateWithBasicAuth(handlers.ScaleFunction, credentials)
		handlers.Info = auth.DecorateWithBasicAuth(handlers.Info, credentials)
		handlers.Secrets = auth.DecorateWithBasicAuth(handlers.Secrets, credentials)
		handlers.Logs = auth.DecorateWithBasicAuth(handlers.Logs, credentials)
		decorateNamespaceHandlersWithBasicAuth(handlers, credentials)

		if handlers.Telemetry != nil {
			handlers.Telemetry = auth.DecorateWithBasicAuth(handlers.Telemetry, credentials)
		}
	}

	hm := newHttpMetrics(registerer)

	// System (auth) endpoints
	router.HandleFunc("/system/functions", hm.InstrumentHandler(handlers.FunctionLister, "")).Methods(http.MethodGet)
	router.HandleFunc("/system/functions", hm.InstrumentHandler(handlers.DeployFunction, "")).Methods(http.MethodPost)
	router.HandleFunc("/system/functions", hm.InstrumentHandler(handlers.DeleteFunction, "")).Methods(http.MethodDelete)
	router.HandleFunc("/system/functions", hm.InstrumentHandler(handlers.UpdateFunction, "")).Methods(http.MethodPut)

	router.HandleFunc("/system/function/{name:["+NameExpression+"]+}",
		hm.InstrumentHandler(handlers.FunctionStatus, "/system/function")).Methods(http.MethodGet)
	router.HandleFunc("/system/scale-function/{name:["+NameExpression+"]+}",
		hm.InstrumentHandler(handlers.ScaleFunction, "/system/scale-function")).Methods(http.MethodPost)

	router.HandleFunc("/system/info",
		hm.InstrumentHandler(handlers.Info, "")).Methods(http.MethodGet)

	router.HandleFunc("/system/secrets",
		hm.InstrumentHandler(handlers.Secrets, "")).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)

	router.HandleFunc("/system/logs",
		hm.InstrumentHandler(handlers.Logs, "")).Methods(http.MethodGet)

	router.HandleFunc("/system/namespaces", hm.InstrumentHandler(handlers.ListNamespaces, "")).Methods(http.MethodGet)

	// Only register the mutate namespace handler if it is defined
	if handlers.MutateNamespace != nil {
		router.HandleFunc("/system/namespace/{name:["+NameExpression+"]*}",
			hm.InstrumentHandler(handlers.MutateNamespace, "")).Methods(http.MethodPost, http.MethodDelete, http.MethodPut, http.MethodGet)
	} else {
		router.HandleFunc("/system/namespace/{name:["+NameExpression+"]*}",
			hm.InstrumentHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Feature not implemented in this version of OpenFaaS", http.StatusNotImplemented)
			}), "")).Methods(http.MethodGet)
	}

	proxyHandler := handlers.FunctionProxy

	// Open endpoints
	router.HandleFunc("/function/{name:["+NameExpression+"]+}", proxyHandler)
	router.HandleFunc("/function/{name:["+NameExpression+"]+}/", proxyHandler)
	router.HandleFunc("/function/{name:["+NameExpression+"]+}/{params:.*}", proxyHandler)

	if handlers.Health != nil {
		router.HandleFunc("/healthz", handlers.Health).
			Methods(http.MethodGet, http.MethodHead)
	}

	if handlers.Telemetry != nil {
		router.HandleFunc("/system/telemetry", hm.InstrumentHandler(handlers.Telemetry, "")).Methods(http.MethodGet)
	}

	router.HandleFunc("/metrics", metricsHandler.ServeHTTP)

	return router
}

// Serve loads handlers into the OpenFaaS route specification. This function is blocking.
func Serve(ctx context.Context, handlers *types.FaaSHandlers, config *types.FaaSConfig) {
	handler := makeHandler(r, handlers, config, prometheus.DefaultRegisterer, promhttp.Handler())

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
		Handler:        handler,
	}

	// Start server in a goroutine
	go func() {
		if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Shutdown server when context is done.
	<-ctx.Done()
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("Failed to shut down provider gracefully: %s", err)
	}
}

func decorateNamespaceHandlersWithBasicAuth(handlers *types.FaaSHandlers, credentials *auth.BasicAuthCredentials) {
	if handlers.ListNamespaces != nil {
		handlers.ListNamespaces = auth.DecorateWithBasicAuth(handlers.ListNamespaces, credentials)
	}

	if handlers.MutateNamespace != nil {
		handlers.MutateNamespace = auth.DecorateWithBasicAuth(handlers.MutateNamespace, credentials)
	}
}
