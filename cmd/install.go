package cmd

import (
	"fmt"
	"os"
	"path"

	systemd "github.com/alexellis/faasd/pkg/systemd"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install faasd",
	RunE:  runInstall,
}

func runInstall(_ *cobra.Command, _ []string) error {

	err := binExists("/usr/local/bin/", "faas-containerd")
	if err != nil {
		return err
	}

	err = binExists("/usr/local/bin/", "faasd")
	if err != nil {
		return err
	}

	err = binExists("/usr/local/bin/", "netns")
	if err != nil {
		return err
	}

	err = systemd.InstallUnit("faas-containerd")
	if err != nil {
		return err
	}

	err = systemd.InstallUnit("faasd")
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

func binExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
	return nil
}
