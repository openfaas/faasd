package systemd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/alexellis/faasd/pkg"
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

func InstallUnit(name string, tokens map[string]string) error {
	if len(tokens["Cwd"]) == 0 {
		return fmt.Errorf("key Cwd expected in tokens parameter")
	}
	tmplName := "./" + name + ".service"
	file, err := pkg.Assets.Open(tmplName)
	if err != nil {
		return fmt.Errorf("error opening template %s, error %s", tmplName, err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading template %s, error %s", tmplName, err)
	}

	tmpl := template.New(name)
	tmpl, err = tmpl.Parse(string(b))

	if err != nil {
		return fmt.Errorf("error generating template %s, error %s", tmplName, err)
	}

	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, tokens)
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
