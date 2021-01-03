package pkg

const (
	// FunctionNamespace is the default containerd namespace functions are created
	FunctionNamespace = "openfaas-fn"

	// faasdNamespace is the containerd namespace services are created
	faasdNamespace = "openfaas"

	faasServicesPullAlways = false

	defaultSnapshotter = "overlayfs"
)
