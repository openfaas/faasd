package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/alexellis/faasd/pkg/service"
	"github.com/alexellis/faasd/pkg/weave"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"

	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const defaultSnapshotter = "overlayfs"

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

func (s *Supervisor) Remove(svcs []Service) error {
	ctx := namespaces.WithNamespace(context.Background(), "default")

	for _, svc := range svcs {
		err := service.Remove(ctx, s.client, svc.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Supervisor) Start(svcs []Service) error {
	ctx := namespaces.WithNamespace(context.Background(), "default")

	wd, _ := os.Getwd()

	writeHostsErr := ioutil.WriteFile(path.Join(wd, "hosts"),
		[]byte(`127.0.0.1	localhost
172.19.0.1	faas-containerd`), 0644)

	if writeHostsErr != nil {
		return fmt.Errorf("cannot write hosts file: %s", writeHostsErr)
	}
	// os.Chown("hosts", 101, 101)

	images := map[string]containerd.Image{}

	for _, svc := range svcs {
		fmt.Printf("Preparing: %s with image: %s\n", svc.Name, svc.Image)

		img, err := service.PrepareImage(ctx, s.client, svc.Image, defaultSnapshotter)
		if err != nil {
			return err
		}
		images[svc.Name] = img
		size, _ := img.Size(ctx)
		fmt.Printf("Prepare done for: %s, %d bytes\n", svc.Image, size)
	}

	for _, svc := range svcs {
		fmt.Printf("Reconciling: %s\n", svc.Name)

		containerErr := service.Remove(ctx, s.client, svc.Name)
		if containerErr != nil {
			return containerErr
		}

		image := images[svc.Name]

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

		mounts = append(mounts, specs.Mount{
			Destination: "/etc/resolv.conf",
			Type:        "bind",
			Source:      path.Join(wd, "resolv.conf"),
			Options:     []string{"rbind", "ro"},
		})

		mounts = append(mounts, specs.Mount{
			Destination: "/etc/hosts",
			Type:        "bind",
			Source:      path.Join(wd, "hosts"),
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
			containerd.WithNewSpec(oci.WithImageConfig(image),
				oci.WithCapabilities(svc.Caps),
				oci.WithMounts(mounts),
				withOCIArgs(svc.Args),
				hook,
				oci.WithEnv(svc.Env)),
		)

		if containerCreateErr != nil {
			log.Printf("Error creating container %s\n", containerCreateErr)
			return containerCreateErr
		}

		log.Printf("Created container %s\n", newContainer.ID())

		task, err := newContainer.NewTask(ctx, cio.NewCreator(cio.WithStdio))
		if err != nil {
			log.Printf("Error creating task: %s\n", err)
			return err
		}

		ip := getIP(newContainer.ID(), task.Pid())

		hosts, _ := ioutil.ReadFile("hosts")

		hosts = []byte(string(hosts) + fmt.Sprintf(`
%s	%s
`, ip, svc.Name))
		writeErr := ioutil.WriteFile("hosts", hosts, 0644)

		if writeErr != nil {
			log.Printf("Error writing file %s %s\n", "hosts", writeErr)
		}
		// os.Chown("hosts", 101, 101)

		_, err = task.Wait(ctx)
		if err != nil {
			log.Printf("Wait err: %s\n", err)
			return err
		}

		log.Printf("Task: %s\tContainer: %s\n", task.ID(), newContainer.ID())
		// log.Println("Exited: ", exitStatusC)

		if err = task.Start(ctx); err != nil {
			log.Printf("Task err: %s\n", err)
			return err
		}
	}

	return nil
}

func getIP(containerID string, taskPID uint32) string {
	// https://github.com/weaveworks/weave/blob/master/net/netdev.go

	peerIDs, err := weave.ConnectedToBridgeVethPeerIds("netns0")
	if err != nil {
		log.Fatal(err)
	}

	addrs, addrsErr := weave.GetNetDevsByVethPeerIds(int(taskPID), peerIDs)
	if addrsErr != nil {
		log.Fatal(addrsErr)
	}
	if len(addrs) > 0 {
		return addrs[0].CIDRs[0].IP.String()
	}

	return ""
}

type Service struct {
	Image  string
	Env    []string
	Name   string
	Mounts []Mount
	Caps   []string
	Args   []string
}

type Mount struct {
	Src  string
	Dest string
}

func withOCIArgs(args []string) oci.SpecOpts {
	if len(args) > 0 {
		return oci.WithProcessArgs(args...)
	}

	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {

		return nil
	}

}
