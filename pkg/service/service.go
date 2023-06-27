package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"golang.org/x/sys/unix"
)

// dockerConfigDir contains "config.json"
const dockerConfigDir = "/var/lib/faasd/.docker/"

// Remove removes a container
func Remove(ctx context.Context, client *containerd.Client, name string, killTimeout time.Duration) error {

	container, err := client.LoadContainer(ctx, name)
	if err != nil {
		// Perhaps the container was already removed, but the snapshot is still there
		service := client.SnapshotService("")
		key := name + "snapshot"

		// Don't return an error if the snapshot doesn't exist
		if _, err := client.SnapshotService("").Stat(ctx, key); err == nil {
			service.Remove(ctx, key)
		}
		return nil
	}

	taskFound := true
	t, err := container.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			taskFound = false
		} else {
			return fmt.Errorf("unable to get task %w: ", err)
		}
	}

	if taskFound {
		status, err := t.Status(ctx)
		if err != nil {
			log.Printf("Unable to get status for: %s, error: %s", name, err.Error())
		} else {
			log.Printf("Status of %s is: %s\n", name, status.Status)
		}

		if err = killTask(ctx, t, killTimeout); err != nil {
			return fmt.Errorf("error killing task %s, %s, %w", container.ID(), name, err)
		}
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("error deleting container %s, %s, %w", container.ID(), name, err)
	}

	return nil
}

// Adapted from Stellar - https://github.com/stellarproject
func killTask(ctx context.Context, task containerd.Task, killTimeout time.Duration) error {

	wg := &sync.WaitGroup{}
	wg.Add(1)
	var err error

	go func() {
		id := task.ID()

		defer wg.Done()
		if task != nil {
			wait, err := task.Wait(ctx)
			if err != nil {
				log.Printf("error waiting on task: %s: %s", id, err)
				return
			}

			if err := task.Kill(ctx, unix.SIGTERM, containerd.WithKillAll); err != nil {
				log.Printf("error killing task: %s with SIGTERM: %s", id, err)
			}

			select {
			case <-wait:
				_, err := task.Delete(ctx)
				if err != nil {
					log.Printf("error deleting task: %s: %s", id, err)
				}

				return
			case <-time.After(killTimeout):
				if err := task.Kill(ctx, unix.SIGKILL, containerd.WithKillAll); err != nil {
					log.Printf("error killing task: %s with SIGTERM: %s", id, err)
				}
				return
			}
		}
	}()
	wg.Wait()

	return err
}

func getResolver(ctx context.Context, configFile *configfile.ConfigFile) (remotes.Resolver, error) {
	// credsFunc is based on https://github.com/moby/buildkit/blob/0b130cca040246d2ddf55117eeff34f546417e40/session/auth/authprovider/authprovider.go#L35
	credFunc := func(host string) (string, string, error) {
		if host == "registry-1.docker.io" {
			host = "https://index.docker.io/v1/"
		}
		ac, err := configFile.GetAuthConfig(host)
		if err != nil {
			return "", "", err
		}
		if ac.IdentityToken != "" {
			return "", ac.IdentityToken, nil
		}
		return ac.Username, ac.Password, nil
	}

	authOpts := []docker.AuthorizerOpt{docker.WithAuthCreds(credFunc)}
	authorizer := docker.NewDockerAuthorizer(authOpts...)
	opts := docker.ResolverOptions{
		Hosts: docker.ConfigureDefaultRegistries(docker.WithAuthorizer(authorizer)),
	}
	return docker.NewResolver(opts), nil
}

func PrepareImage(ctx context.Context, client *containerd.Client, imageName, snapshotter string, pullAlways bool) (containerd.Image, error) {
	var (
		empty    containerd.Image
		resolver remotes.Resolver
	)

	if _, statErr := os.Stat(filepath.Join(dockerConfigDir, config.ConfigFileName)); statErr == nil {
		configFile, err := config.Load(dockerConfigDir)
		if err != nil {
			return nil, err
		}
		resolver, err = getResolver(ctx, configFile)
		if err != nil {
			return empty, err
		}
	} else if !os.IsNotExist(statErr) {
		return empty, statErr
	}

	var image containerd.Image
	if pullAlways {
		img, err := pullImage(ctx, client, resolver, imageName)
		if err != nil {
			return empty, err
		}

		image = img
	} else {
		img, err := client.GetImage(ctx, imageName)
		if err != nil {
			if !errdefs.IsNotFound(err) {
				return empty, err
			}
			img, err := pullImage(ctx, client, resolver, imageName)
			if err != nil {
				return empty, err
			}
			image = img
		} else {
			image = img
		}
	}

	unpacked, err := image.IsUnpacked(ctx, snapshotter)
	if err != nil {
		return empty, fmt.Errorf("cannot check if unpacked: %s", err)
	}

	if !unpacked {
		if err := image.Unpack(ctx, snapshotter); err != nil {
			return empty, fmt.Errorf("cannot unpack: %s", err)
		}
	}

	return image, nil
}

func pullImage(ctx context.Context, client *containerd.Client, resolver remotes.Resolver, imageName string) (containerd.Image, error) {

	var empty containerd.Image

	rOpts := []containerd.RemoteOpt{
		containerd.WithPullUnpack,
	}

	if resolver != nil {
		rOpts = append(rOpts, containerd.WithResolver(resolver))
	}

	img, err := client.Pull(ctx, imageName, rOpts...)
	if err != nil {
		return empty, fmt.Errorf("cannot pull: %s", err)
	}

	return img, nil
}
