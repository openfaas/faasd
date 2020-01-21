package config

import (
	"strconv"
	"testing"
)

type EnvBucket struct {
	Items map[string]string
}

func NewEnvBucket() EnvBucket {
	return EnvBucket{
		Items: make(map[string]string),
	}
}

func (e EnvBucket) Getenv(key string) string {
	return e.Items[key]
}

func (e EnvBucket) Setenv(key string, value string) {
	e.Items[key] = value
}

func Test_SetSockByEnv(t *testing.T) {
	defaultSock := "/run/containerd/containerd.sock"
	expectedSock := "/non/default/value.sock"
	env := NewEnvBucket()
	_, config, err := ReadFromEnv(env)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if config.Sock != defaultSock {
		t.Fatalf("expected %q, got %q", defaultSock, config.Sock)
	}

	env.Setenv("sock", expectedSock)
	_, config, err = ReadFromEnv(env)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if config.Sock != expectedSock {
		t.Fatalf("expected %q, got %q", expectedSock, config.Sock)
	}
}

func Test_SetServiceTimeout(t *testing.T) {
	defaultTimeout := "1m0s"

	env := NewEnvBucket()
	config, _, err := ReadFromEnv(env)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if config.ReadTimeout.String() != defaultTimeout {
		t.Fatalf("expected %q, got %q", defaultTimeout, config.ReadTimeout)
	}

	if config.WriteTimeout.String() != defaultTimeout {
		t.Fatalf("expected %q, got %q", defaultTimeout, config.WriteTimeout)
	}

	newTimeout := "30s"
	env.Setenv("service_timeout", newTimeout)
	config, _, err = ReadFromEnv(env)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if config.ReadTimeout.String() != newTimeout {
		t.Fatalf("expected %q, got %q", newTimeout, config.ReadTimeout)
	}

	if config.WriteTimeout.String() != newTimeout {
		t.Fatalf("expected %q, got %q", newTimeout, config.WriteTimeout)
	}
}

func Test_SetPort(t *testing.T) {
	defaultPort := 8081

	env := NewEnvBucket()
	config, _, err := ReadFromEnv(env)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if config.TCPPort == nil {
		t.Fatal("expected non-nil TCPPort")
	}
	if *config.TCPPort != defaultPort {
		t.Fatalf("expected %d, got %d", defaultPort, config.TCPPort)
	}

	newPort := 9091
	newPortStr := strconv.Itoa(newPort)
	env.Setenv("port", newPortStr)
	config, _, err = ReadFromEnv(env)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if config.TCPPort == nil {
		t.Fatal("expected non-nil TCPPort")
	}
	if *config.TCPPort != newPort {
		t.Fatalf("expected %d, got %d", newPort, config.TCPPort)
	}
}
