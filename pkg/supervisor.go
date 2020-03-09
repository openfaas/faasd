package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/oci"
	gocni "github.com/containerd/go-cni"
	"github.com/openfaas/faasd/pkg/cninetwork"
	"github.com/openfaas/faasd/pkg/service"

	"github.com/containerd/containerd/namespaces"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	defaultSnapshotter         = "overlayfs"
	workingDirectoryPermission = 0644
	// faasdNamespace is the containerd namespace services are created
	faasdNamespace         = "default"
	faasServicesPullAlways = false
)

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

type Supervisor struct {
	client *containerd.Client
	cni    gocni.CNI
}

func NewSupervisor(sock string) (*Supervisor, error) {
	client, err := containerd.New(sock)
	if err != nil {
		return nil, err
	}

	cni, err := cninetwork.InitNetwork()
	if err != nil {
		return nil, err
	}

	return &Supervisor{
		client: client,
		cni:    cni,
	}, nil
}

func (s *Supervisor) Start(svcs []Service) error {
	ctx := namespaces.WithNamespace(context.Background(), faasdNamespace)

	wd, _ := os.Getwd()

	gw, err := cninetwork.CNIGateway()
	if err != nil {
		return err
	}
	hosts := fmt.Sprintf(`
127.0.0.1	localhost
%s	faasd-provider`, gw)

	writeHostsErr := ioutil.WriteFile(path.Join(wd, "hosts"),
		[]byte(hosts), workingDirectoryPermission)

	if writeHostsErr != nil {
		return fmt.Errorf("cannot write hosts file: %s", writeHostsErr)
	}

	images := map[string]containerd.Image{}

	for _, svc := range svcs {
		fmt.Printf("Preparing: %s with image: %s\n", svc.Name, svc.Image)

		img, err := service.PrepareImage(ctx, s.client, svc.Image, defaultSnapshotter, faasServicesPullAlways)
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

		newContainer, containerCreateErr := s.client.NewContainer(
			ctx,
			svc.Name,
			containerd.WithImage(image),
			containerd.WithNewSnapshot(svc.Name+"-snapshot", image),
			containerd.WithNewSpec(oci.WithImageConfig(image),
				oci.WithCapabilities(svc.Caps),
				oci.WithMounts(mounts),
				withOCIArgs(svc.Args),
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

		labels := map[string]string{}
		network, err := cninetwork.CreateCNINetwork(ctx, s.cni, task, labels)

		if err != nil {
			return err
		}

		ip, err := cninetwork.GetIPAddress(network, task)
		if err != nil {
			return err
		}
		log.Printf("%s has IP: %s\n", newContainer.ID(), ip.String())

		hosts, _ := ioutil.ReadFile("hosts")

		hosts = []byte(string(hosts) + fmt.Sprintf(`
%s	%s
`, ip, svc.Name))
		writeErr := ioutil.WriteFile("hosts", hosts, workingDirectoryPermission)

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

func (s *Supervisor) Close() {
	defer s.client.Close()
}

func (s *Supervisor) Remove(svcs []Service) error {
	ctx := namespaces.WithNamespace(context.Background(), faasdNamespace)

	for _, svc := range svcs {
		log.Printf("Removing service %s\n", svc.Name)
		err := cninetwork.DeleteCNINetwork(ctx, s.cni, s.client, svc.Name)
		if err != nil {
			log.Printf("[Delete] error removing CNI network for %s, %s\n", svc.Name, err)
			return err
		}

		err = service.Remove(ctx, s.client, svc.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func withOCIArgs(args []string) oci.SpecOpts {
	if len(args) > 0 {
		return oci.WithProcessArgs(args...)
	}

	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {
		return nil
	}
}
