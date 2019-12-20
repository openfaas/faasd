package pkg

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
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
	for _, svc := range svcs {
		fmt.Printf("Preparing: %s", svc.Name)

		fmt.Printf("Pulling: %s\n", svc.Image)
		bytesOut, err := pullImage(s.client, svc.Image)
		if err != nil {
			return err
		}
		fmt.Printf("Pull done for: %s, %d bytes\n", svc.Image, bytesOut)
	}

	return nil
}

func pullImage(client *containerd.Client, image string) (int64, error) {
	ctx := namespaces.WithNamespace(context.Background(), "default")

	pulled, err := client.Pull(ctx, image, containerd.WithPullUnpack)

	if err != nil {
		return 0, err
	}
	bytesOut, _ := pulled.Size(ctx)
	return bytesOut, nil
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
