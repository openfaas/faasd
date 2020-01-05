package cmd

import (
	"github.com/alexellis/faasd/pkg"
	"log"
	"path"

	systemd "github.com/alexellis/faasd/pkg/systemd"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install faasd",
	RunE:  runInstall,
}

const faasdwd = "/run/faasd"
const faasContainerdwd = "/run/faas-containerd"

func runInstall(cmd *cobra.Command, _ []string) error {

	if err := pkg.EnsureWorkingDir(path.Join(faasdwd, "secrets")); err != nil {
		return err
	}

	if err := pkg.EnsureWorkingDir(faasContainerdwd); err != nil {
		return err
	}

	if basicAuthErr := pkg.MakeBasicAuthFiles(path.Join(faasdwd, "secrets")); basicAuthErr != nil {
		return errors.Wrap(basicAuthErr, "cannot create basic-auth-* files")
	}

	if err := pkg.CopyFile("prometheus.yml", faasdwd); err != nil {
		return err
	}

	if err := pkg.CopyFile("resolv.conf", faasdwd); err != nil {
		return err
	}

	// If we are only installing the files into the /run/ directories we skip installing systemd files
	prepareOnly, _ := cmd.Flags().GetBool("prepare")
	if !prepareOnly {
		log.Println("Installing systemd services")
		if err := checkBinaries(); err != nil {
			return err
		}

		if err := installSystemdServices(); err != nil {
			return err
		}

	}

	return nil
}

func checkBinaries() error {
	binaries := []string{"faas-containerd", "faasd", "netns"}

	for _, binary := range binaries {
		if err := pkg.BinaryExists("/usr/local/bin/", binary); err != nil {
			return err
		}
	}

	return nil
}

func installSystemdServices() error {

	err := systemd.InstallUnit("faas-containerd", map[string]string{
		"Cwd":             faasContainerdwd,
		"SecretMountPath": path.Join(faasdwd, "secrets")})

	if err != nil {
		return err
	}

	err = systemd.InstallUnit("faasd", map[string]string{"Cwd": faasdwd})
	if err != nil {
		return err
	}

	err = systemd.DaemonReload()
	if err != nil {
		return err
	}

	err = systemd.Enable("faas-containerd")
	if err != nil {
		return err
	}

	err = systemd.Enable("faasd")
	if err != nil {
		return err
	}

	err = systemd.Start("faas-containerd")
	if err != nil {
		return err
	}

	err = systemd.Start("faasd")
	if err != nil {
		return err
	}
	return nil
}
