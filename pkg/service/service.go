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
func Remove(ctx context.Context, client *containerd.Client, name string) error {

	container, containerErr := client.LoadContainer(ctx, name)

	if containerErr == nil {
		found := true
		t, err := container.Task(ctx, nil)
		if err != nil {
			if errdefs.IsNotFound(err) {
				found = false
			} else {
				return fmt.Errorf("unable to get task %s: ", err)
			}
		}

		if found {
			status, _ := t.Status(ctx)
			fmt.Printf("Status of %s is: %s\n", name, status.Status)

			log.Printf("Need to kill %s\n", name)
			err := killTask(ctx, t)
			if err != nil {
				return fmt.Errorf("error killing task %s, %s, %s", container.ID(), name, err)
			}
		}

		err = container.Delete(ctx, containerd.WithSnapshotCleanup)
		if err != nil {
			return fmt.Errorf("error deleting container %s, %s, %s", container.ID(), name, err)
		}
	} else {
		service := client.SnapshotService("")
		key := name + "snapshot"
		if _, err := client.SnapshotService("").Stat(ctx, key); err == nil {
			service.Remove(ctx, key)
		}
	}
	return nil
}

// Adapted from Stellar - https://github.com/stellar
func killTask(ctx context.Context, task containerd.Task) error {

	killTimeout := 30 * time.Second

	wg := &sync.WaitGroup{}
	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		if task != nil {
			wait, err := task.Wait(ctx)
			if err != nil {
				err = fmt.Errorf("error waiting on task: %s", err)
				return
			}
			if err := task.Kill(ctx, unix.SIGTERM, containerd.WithKillAll); err != nil {
				log.Printf("error killing container task: %s", err)
			}

			select {
			case <-wait:
				task.Delete(ctx)
				return
			case <-time.After(killTimeout):
				if err := task.Kill(ctx, unix.SIGKILL, containerd.WithKillAll); err != nil {
					log.Printf("error force killing container task: %s", err)
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

	if _, stErr := os.Stat(filepath.Join(dockerConfigDir, config.ConfigFileName)); stErr == nil {
		configFile, err := config.Load(dockerConfigDir)
		if err != nil {
			return nil, err
		}
		resolver, err = getResolver(ctx, configFile)
		if err != nil {
			return empty, err
		}
	} else if !os.IsNotExist(stErr) {
		return empty, stErr
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
