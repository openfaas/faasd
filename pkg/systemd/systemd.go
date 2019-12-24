package systemd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	execute "github.com/alexellis/go-execute/pkg/v1"
)

func Enable(unit string) error {
	task := execute.ExecTask{Command: "systemctl",
		Args:        []string{"enable", unit},
		StreamStdio: false,
	}

	res, err := task.Execute()
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("error executing task %s %v, stderr: %s", task.Command, task.Args, res.Stderr)
	}

	return nil
}

func Start(unit string) error {
	task := execute.ExecTask{Command: "systemctl",
		Args:        []string{"start", unit},
		StreamStdio: false,
	}

	res, err := task.Execute()
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("error executing task %s %v, stderr: %s", task.Command, task.Args, res.Stderr)
	}

	return nil
}

func DaemonReload() error {
	task := execute.ExecTask{Command: "systemctl",
		Args:        []string{"daemon-reload"},
		StreamStdio: false,
	}

	res, err := task.Execute()
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("error executing task %s %v, stderr: %s", task.Command, task.Args, res.Stderr)
	}

	return nil
}

func InstallUnit(name string) error {

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
