package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/openfaas/faasd/pkg/cninetwork"
)

// GetFunction returns a function that matches name
func GetFunction(client *containerd.Client, name string, namespace string) (Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	fn := Function{}

	c, err := client.LoadContainer(ctx, name)
	if err != nil {
		return Function{}, fmt.Errorf("unable to find function: %s, error %w", name, err)
	}

	image, err := c.Image(ctx)
	if err != nil {
		return fn, err
	}

	containerName := c.ID()
	allLabels, labelErr := c.Labels(ctx)

	if labelErr != nil {
		log.Printf("cannot list container %s labels: %s", containerName, labelErr)
	}

	labels, annotations := buildLabelsAndAnnotations(allLabels)

	spec, err := c.Spec(ctx)
	if err != nil {
		return Function{}, fmt.Errorf("unable to load function %s error: %w", name, err)
	}

	info, err := c.Info(ctx)
	if err != nil {
		return Function{}, fmt.Errorf("can't load info for: %s, error %w", name, err)
	}

	envVars, envProcess := readEnvFromProcessEnv(spec.Process.Env)
	secrets := readSecretsFromMounts(spec.Mounts)

	fn.name = containerName
	fn.namespace = namespace
	fn.image = image.Name()
	fn.labels = labels
	fn.annotations = annotations
	fn.secrets = secrets
	fn.envVars = envVars
	fn.envProcess = envProcess
	fn.createdAt = info.CreatedAt
	fn.memoryLimit = readMemoryLimitFromSpec(spec)

	replicas := 0
	task, err := c.Task(ctx, nil)
	if err == nil {
		// Task for container exists
		svc, err := task.Status(ctx)
		if err != nil {
			return Function{}, fmt.Errorf("unable to get task status for container: %s %w", name, err)
		}

		if svc.Status == "running" {
			replicas = 1
			fn.pid = task.Pid()

			// Get container IP address
			ip, err := cninetwork.GetIPAddress(name, task.Pid())
			if err != nil {
				return Function{}, err
			}
			fn.IP = ip
		}
	} else {
		replicas = 0
	}

	fn.replicas = replicas
	return fn, nil
}

func readEnvFromProcessEnv(env []string) (map[string]string, string) {
	foundEnv := make(map[string]string)
	fprocess := ""
	for _, e := range env {
		kv := strings.Split(e, "=")
		if len(kv) == 1 {
			continue
		}

		if kv[0] == "PATH" {
			continue
		}

		if kv[0] == "fprocess" {
			fprocess = kv[1]
			continue
		}

		foundEnv[kv[0]] = kv[1]
	}

	return foundEnv, fprocess
}

func readSecretsFromMounts(mounts []specs.Mount) []string {
	secrets := []string{}
	for _, mnt := range mounts {
		x := strings.Split(mnt.Destination, "/var/openfaas/secrets/")
		if len(x) > 1 {
			secrets = append(secrets, x[1])
		}

	}
	return secrets
}

// buildLabelsAndAnnotations returns a separated list with labels first,
// followed by annotations by checking each key of ctrLabels for a prefix.
func buildLabelsAndAnnotations(ctrLabels map[string]string) (map[string]string, map[string]string) {
	labels := make(map[string]string)
	annotations := make(map[string]string)

	for k, v := range ctrLabels {
		if strings.HasPrefix(k, annotationLabelPrefix) {
			annotations[strings.TrimPrefix(k, annotationLabelPrefix)] = v
		} else {
			labels[k] = v
		}
	}

	return labels, annotations
}

func readMemoryLimitFromSpec(spec *specs.Spec) int64 {
	if spec.Linux == nil || spec.Linux.Resources == nil || spec.Linux.Resources.Memory == nil || spec.Linux.Resources.Memory.Limit == nil {
		return 0
	}
	return *spec.Linux.Resources.Memory.Limit
}
