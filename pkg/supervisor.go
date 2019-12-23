package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/alexellis/faasd/pkg/weave"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	"golang.org/x/sys/unix"

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
		fmt.Printf("Preparing: %s\n", svc.Name)

		img, err := prepareImage(ctx, s.client, svc.Image)
		if err != nil {
			return err
		}
		images[svc.Name] = img
		size, _ := img.Size(ctx)
		fmt.Printf("Prepare done for: %s, %d bytes\n", svc.Image, size)

	}

	for _, svc := range svcs {
		fmt.Printf("Reconciling: %s\n", svc.Name)

		image := images[svc.Name]

		container, containerErr := s.client.LoadContainer(ctx, svc.Name)

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
				fmt.Println("Status:", status.Status)

				// if status.Status == containerd.Running {
				log.Println("need to kill", svc.Name)
				err := killTask(ctx, t)
				if err != nil {
					return fmt.Errorf("error killing task %s, %s, %s", container.ID(), svc.Name, err)
				}
				// }
			}

			err = container.Delete(ctx, containerd.WithSnapshotCleanup)
			if err != nil {
				return fmt.Errorf("error deleting container %s, %s, %s", container.ID(), svc.Name, err)
			}

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
			log.Println(containerCreateErr)
			return containerCreateErr
		}

		log.Printf("Created container %s\n", newContainer.ID())

		task, err := newContainer.NewTask(ctx, cio.NewCreator(cio.WithStdio))
		if err != nil {
			log.Println(err)
			return err
		}

		ip := getIP(container.ID(), task.Pid())

		hosts, _ := ioutil.ReadFile("hosts")

		hosts = []byte(string(hosts) + fmt.Sprintf(`
%s	%s
`, ip, svc.Name))
		writeErr := ioutil.WriteFile("hosts", hosts, 0644)

		if writeErr != nil {
			log.Println("Error writing hosts file")
		}
		// os.Chown("hosts", 101, 101)

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

func prepareImage(ctx context.Context, client *containerd.Client, imageName string) (containerd.Image, error) {
	snapshotter := defaultSnapshotter

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
