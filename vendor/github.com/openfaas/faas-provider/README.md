faas-provider
==============

This faas-provider can be used to write your own back-end for OpenFaaS. The Golang SDK can be vendored into your project so that you can provide a provider which is compliant and compatible with the OpenFaaS gateway.

![Conceptual diagram](docs/conceptual.png)

The faas-provider provides CRUD for functions and an invoke capability. If you complete the required endpoints then you will be able to use your container orchestrator or back-end system with the existing OpenFaaS ecosystem and tooling.

Read more: [The power of interfaces in OpenFaaS](https://blog.alexellis.io/the-power-of-interfaces-openfaas/)

### Recommendations

The following is used in OpenFaaS and recommended for those seeking to build their own back-ends:

* License: MIT
* Language: Golang

### How to use this project

All the required HTTP routes are configured automatically including a HTTP server on port 8080. Your task is to implement the supplied HTTP handler functions.

Examples:

**OpenFaaS for Kubernetes**

See the [main.go](https://github.com/openfaas/faas-netes/blob/master/main.go) file in the [faas-netes](https://github.com/openfaas/faas-netes) Kubernetes backend.

**OpenFaaS for containerd (faasd)**

See [provider.go](https://github.com/openfaas/faasd/blob/master/cmd/provider.go#L100) for the [faasd backend](https://github.com/openfaas/faasd/)

I.e.:

```go
	timeout := 8 * time.Second
	bootstrapHandlers := bootTypes.FaaSHandlers{
		ListNamespaces: handlers.MakeNamespaceLister(),
		FunctionProxy:  handlers.MakeProxyHandler(),
		FunctionLister: handlers.MakeFunctionLister(),
		DeployFunction: handlers.MakeDeployFunctionHandler(),
		DeleteFunction: handlers.MakeDeleteFunctionHandler(),
		UpdateFunction: handlers.MakeUpdateFunctionHandler(),
		FunctionStatus: handlers.MakeFunctionStatusHandler(),
		ScaleFunction: 	handlers.MakeScaleFunctionHandler(),
		Secrets: 	  	handlers.MakeSecretHandler(),
		Logs: 			handlers.MakeLogsHandler(),
		Info: 			handlers.MakeInfoHandler(),
		Health: 		handlers.MakeHealthHandler(),
	}

	var port int
	port = 8080
	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		TCPPort:      &port,
	}

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
```

