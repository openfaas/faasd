package handlers

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Function struct {
	name      string
	namespace string
	image     string
	pid       uint32
	replicas  int
	IP        string
}

const (
	// FunctionNamespace is the containerd namespace functions are created
	FunctionNamespace = "openfaas-fn"
	FaasdNamespace    = "openfaas"
)

// NewContainer uses the containerd client to create a new container
func NewContainer(ctx context.Context, client containerd.Client, name string, image containerd.Image, snapshotter string, caps []string, mounts []specs.Mount, args []string, envs []string) (containerd.Container, error) {

	container, err := client.NewContainer(
		ctx,
		name,
		containerd.WithImage(image),
		containerd.WithSnapshotter(snapshotter),
		containerd.WithNewSnapshot(name+"-snapshot", image),
		containerd.WithNewSpec(oci.WithImageConfig(image),
			oci.WithCapabilities(caps),
			oci.WithMounts(mounts),
			withOCIArgs(args),
			oci.WithEnv(envs)),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to create container: %s, error: %s", name, err)
	}
	return container, nil
}

func withOCIArgs(args []string) oci.SpecOpts {
	if len(args) > 0 {
		return oci.WithProcessArgs(args...)
	}

	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {

		return nil
	}
}

// ListFunctions returns a map of all functions with running tasks on namespace
func ListFunctions(client *containerd.Client) (map[string]Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), FunctionNamespace)
	functions := make(map[string]Function)

	containers, _ := client.Containers(ctx)
	for _, k := range containers {
		name := k.ID()
		functions[name], _ = GetFunction(client, name)
	}
	return functions, nil
}

// GetFunction returns a function that matches name
func GetFunction(client *containerd.Client, name string) (Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), FunctionNamespace)
	c, err := client.LoadContainer(ctx, name)

	if err == nil {

		image, _ := c.Image(ctx)
		f := Function{
			name:      c.ID(),
			namespace: FunctionNamespace,
			image:     image.Name(),
		}

		replicas := 0
		task, err := c.Task(ctx, nil)
		if err == nil {
			// Task for container exists
			svc, err := task.Status(ctx)
			if err != nil {
				return Function{}, fmt.Errorf("unable to get task status for container: %s %s", name, err)
			}
			if svc.Status == "running" {
				replicas = 1
				f.pid = task.Pid()
				// Get container IP address
				ip, _ := GetIPfromPID(int(task.Pid()))
				f.IP = ip.String()
			}
		} else {
			replicas = 0
		}

		f.replicas = replicas
		return f, nil

	}
	return Function{}, fmt.Errorf("unable to find function: %s, error %s", name, err)
}
