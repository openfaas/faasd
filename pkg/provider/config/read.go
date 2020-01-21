package config

import (
	"time"

	types "github.com/openfaas/faas-provider/types"
)

type ProviderConfig struct {
	// Sock is the address of the containerd socket
	Sock string
}

// ReadFromEnv loads the FaaSConfig and the Containerd specific config form the env variables
func ReadFromEnv(hasEnv types.HasEnv) (*types.FaaSConfig, *ProviderConfig, error) {
	config, err := types.ReadConfig{}.Read(hasEnv)
	if err != nil {
		return nil, nil, err
	}

	serviceTimeout := types.ParseIntOrDurationValue(hasEnv.Getenv("service_timeout"), time.Second*60)

	config.EnableHealth = true
	config.ReadTimeout = serviceTimeout
	config.WriteTimeout = serviceTimeout

	port := types.ParseIntValue(hasEnv.Getenv("port"), 8081)
	config.TCPPort = &port

	providerConfig := &ProviderConfig{
		Sock: types.ParseString(hasEnv.Getenv("sock"), "/run/containerd/containerd.sock"),
	}

	return config, providerConfig, nil
}
