package cmd

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/openfaas/faasd/pkg/assets"
	systemd "github.com/openfaas/faasd/pkg/systemd"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install faasd",
	RunE:  runInstall,
}

const workingDirectoryPermission = 0644

const faasdwd = "/var/lib/faasd"

const faasdProviderWd = "/var/lib/faasd-provider"

func runInstall(_ *cobra.Command, _ []string) error {

	if err := ensureWorkingDir(path.Join(faasdwd, "secrets")); err != nil {
		return err
	}

	if err := ensureWorkingDir(faasdProviderWd); err != nil {
		return err
	}

	if basicAuthErr := makeBasicAuthFiles(path.Join(faasdwd, "secrets")); basicAuthErr != nil {
		return errors.Wrap(basicAuthErr, "cannot create basic-auth-* files")
	}

	if err := copyConfig(faasdwd); err != nil {
		return err
	}

	err := binExists("/usr/local/bin/", "faasd")
	if err != nil {
		return err
	}

	err = systemd.InstallUnit("faasd-provider", map[string]string{
		"Cwd":             faasdProviderWd,
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

	err = systemd.Enable("faasd-provider")
	if err != nil {
		return err
	}

	err = systemd.Enable("faasd")
	if err != nil {
		return err
	}

	err = systemd.Start("faasd-provider")
	if err != nil {
		return err
	}

	err = systemd.Start("faasd")
	if err != nil {
		return err
	}

	fmt.Println(`Login with:
  sudo cat /var/lib/faasd/secrets/basic-auth-password | faas-cli login -s`)

	return nil
}

func binExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
	return nil
}

func ensureWorkingDir(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		err = os.MkdirAll(folder, workingDirectoryPermission)
		if err != nil {
			return err
		}
	}

	return nil
}

// copyConfig writes the required faasd configuration files to the destFolder
// if the files exist locally, those files will be used, otherwise the default embedded
// assets are used.  This allows the user to customize the installation by using `faasd generate`
// to create and then edit the configuration files.
func copyConfig(destFolder string) (err error) {
	if fileExists("docker-compose.yaml") {
		err = cp("docker-compose.yaml", destFolder)
	} else {
		err = assets.WriteCompose(destFolder)
	}

	if err != nil {
		return err
	}

	if fileExists("prometheus.yml") {
		err = cp("prometheus.yml", destFolder)
	} else {
		err = assets.WritePrometheus(destFolder)
	}

	if err != nil {
		return err
	}

	if fileExists("resolv.conf") {
		err = cp("resolv.conf", destFolder)
	} else {
		err = assets.WriteResolv(destFolder)
	}

	if err != nil {
		return err
	}

	return nil
}

func cp(source, destFolder string) error {
	file, err := os.Open(source)
	if err != nil {
		return err

	}
	defer file.Close()

	out, err := os.Create(path.Join(destFolder, source))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	return err
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
