package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"

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

	err = binExists("/usr/local/bin/", "netns")
	if err != nil {
		return err
	}

	err = installUnit("faas-containerd")
	if err != nil {
		return err
	}

	err = installUnit("faasd")
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

func installUnit(name string) error {

	tmpl, err := template.ParseFiles("./hack/" + name + ".service")

	wd, _ := os.Getwd()
	var tpl bytes.Buffer
	userData := struct {
		Cwd string
	}{
		Cwd: wd,
	}

	err = tmpl.Execute(&tpl, userData)
	if err != nil {
		return err
	}

	err = writeUnit(name+".service", tpl.Bytes())

	if err != nil {
		return err
	}
	return nil
}

func writeUnit(name string, data []byte) error {
	f, err := os.Create(filepath.Join("/lib/systemd/system", name))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
