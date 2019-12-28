package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"golang.org/x/sys/unix"
)

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

// From Stellar
func killTask(ctx context.Context, task containerd.Task) error {
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
			case <-time.After(5 * time.Second):
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

func PrepareImage(ctx context.Context, client *containerd.Client, imageName, snapshotter string) (containerd.Image, error) {

	var empty containerd.Image
	image, err := client.GetImage(ctx, imageName)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return empty, err
		}

		img, err := client.Pull(ctx, imageName, containerd.WithPullUnpack)
		if err != nil {
			return empty, fmt.Errorf("cannot pull: %s", err)
		}
		image = img
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
