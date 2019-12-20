package pkg

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Supervisor struct {
	client *containerd.Client
}

func NewSupervisor(sock string) (*Supervisor, error) {
	client, err := containerd.New(sock)
	if err != nil {
		panic(err)
	}

	return &Supervisor{
		client: client,
	}, nil
}

func (s *Supervisor) Close() {
	defer s.client.Close()
}

func (s *Supervisor) Start(svcs []Service) error {
	ctx := namespaces.WithNamespace(context.Background(), "default")

	images := map[string]containerd.Image{}

	for _, svc := range svcs {
		fmt.Printf("Preparing: %s", svc.Name)

		fmt.Printf("Pulling: %s\n", svc.Image)
		img, err := pullImage(ctx, s.client, svc.Image)
		if err != nil {
			return err
		}
		images[svc.Name] = img
		size, _ := img.Size(ctx)
		fmt.Printf("Pull done for: %s, %d bytes\n", svc.Image, size)

	}

	for _, svc := range svcs {
		fmt.Printf("Starting: %s\n", svc.Name)

		image := images[svc.Name]

		container, containerErr := s.client.ContainerService().Get(ctx, svc.Name)

		if containerErr == nil {
			taskReq := &tasks.GetRequest{
				ContainerID: container.ID,
			}

			task, err := s.client.TaskService().Get(ctx, taskReq)

			if task != nil && task.Process != nil {
				zero := time.Time{}
				fmt.Println(task.Process.ExitedAt, "=", zero)

				if task.Process.ExitedAt == zero {
					log.Println("need to kill", svc.Name)
					killReq := tasks.KillRequest{
						ContainerID: container.ID,
						Signal:      uint32(syscall.SIGTERM),
					}
					em, err := s.client.TaskService().Kill(ctx, &killReq)
					if err != nil {
						return fmt.Errorf("error killing task %s, %s, %s", container.ID, svc.Name, err)
					}
					time.Sleep(1 * time.Second)
					log.Println("em", em)
				}

				deleteReq := tasks.DeleteTaskRequest{
					ContainerID: container.ID,
				}
				_, err = s.client.TaskService().Delete(ctx, &deleteReq)
				if err != nil {
					return fmt.Errorf("error deleting task %s, %s, %s", container.ID, svc.Name, err)
				}

			}

			err = s.client.ContainerService().Delete(ctx, svc.Name)
			fmt.Println(err)

			err = s.client.SnapshotService("").Remove(ctx, svc.Name+"-snapshot")
			fmt.Println(err)
		}

		mounts := []specs.Mount{}
		if len(svc.Mounts) > 0 {
			for _, mnt := range svc.Mounts {
				mounts = append(mounts, specs.Mount{
					Source:      mnt.Src,
					Destination: mnt.Dest,
					Type:        "bind",
					Options:     []string{"rbind", "rw"},
				})
			}

		}

		wd, _ := os.Getwd()
		resolv := path.Join(wd, "resolv.conf")
		log.Println("Using ", resolv)
		mounts = append(mounts, specs.Mount{
			Destination: "/etc/resolv.conf",
			Type:        "bind",
			Source:      resolv,
			Options:     []string{"rbind", "ro"},
		})

		hook := func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
			if s.Hooks == nil {
				s.Hooks = &specs.Hooks{}
			}
			netnsPath, err := exec.LookPath("netns")
			if err != nil {
				return err
			}

			s.Hooks.Prestart = []specs.Hook{
				{
					Path: netnsPath,
					Args: []string{
						"netns",
					},
					Env: os.Environ(),
				},
			}
			return nil
		}

		newContainer, containerCreateErr := s.client.NewContainer(
			ctx,
			svc.Name,
			containerd.WithImage(image),
			containerd.WithNewSnapshot(svc.Name+"-snapshot", image),
			containerd.WithNewSpec(oci.WithImageConfig(image), oci.WithCapabilities([]string{"CAP_NET_RAW"}), oci.WithMounts(mounts), hook),
		)

		if containerCreateErr != nil {
			log.Println(containerCreateErr)
			return containerCreateErr
		}
		fmt.Println("created", newContainer.ID())

		task, err := newContainer.NewTask(ctx, cio.NewCreator(cio.WithStdio))
		if err != nil {
			log.Println(err)
			return err
		}

		exitStatusC, err := task.Wait(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("Exited: ", exitStatusC)

		// call start on the task to execute the redis server
		if err := task.Start(ctx); err != nil {
			log.Println("Task err: ", err)
			return err
		}
	}

	return nil
}

func pullImage(ctx context.Context, client *containerd.Client, image string) (containerd.Image, error) {

	pulled, err := client.Pull(ctx, image, containerd.WithPullUnpack)

	if err != nil {
		return nil, err
	}

	return pulled, nil
}

type Service struct {
	Image  string
	Env    []string
	Name   string
	Mounts []Mount
}

type Mount struct {
	Src  string
	Dest string
}
