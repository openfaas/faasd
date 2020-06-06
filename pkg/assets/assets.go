//go:generate vfsgendev -source="github.com/openfaas/faasd/pkg/assets".Config
package assets

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WriteConfigFiles(dst string) error {
	err := WriteCompose(dst)
	if err != nil {
		return err
	}

	err = WritePrometheus(dst)
	if err != nil {
		return err
	}

	err = WriteResolv(dst)
	if err != nil {
		return err
	}

	return nil
}

// WriteCompose will write the default docker-compose asset to the specified location
func WriteCompose(dst string) error {
	file, err := Config.Open("docker-compose.yaml")
	if err != nil {
		return fmt.Errorf("can not open docker-compose asset: %w", err)
	}
	dst = filepath.Join(dst, "docker-compose.yaml")
	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("can not open destination location: %w", err)
	}
	defer to.Close()

	_, err = io.Copy(to, file)
	return err
}

// WritePrometheus will write the default prometheus configuration to the destination
func WritePrometheus(dst string) error {
	file, err := Config.Open("prometheus.yml")
	if err != nil {
		return fmt.Errorf("can not open prometheus config asset: %w", err)
	}
	dst = filepath.Join(dst, "prometheus.yml")
	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("can not open destination location: %w", err)
	}
	defer to.Close()

	_, err = io.Copy(to, file)
	return err
}

// WriteResolv will write the default resolv configuration to the destination
func WriteResolv(dst string) error {
	file, err := Config.Open("resolv.conf")
	if err != nil {
		return fmt.Errorf("can not open prometheus config asset: %w", err)
	}
	dst = filepath.Join(dst, "resolv.conf")
	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("can not open destination location: %w", err)
	}
	defer to.Close()

	_, err = io.Copy(to, file)
	return err
}
