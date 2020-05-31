/*
   Copyright 2020 The Compose Specification Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package types

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ConfigDetails are the details about a group of ConfigFiles
type ConfigDetails struct {
	Version     string
	WorkingDir  string
	ConfigFiles []ConfigFile
	Environment map[string]string
}

// LookupEnv provides a lookup function for environment variables
func (cd ConfigDetails) LookupEnv(key string) (string, bool) {
	v, ok := cd.Environment[key]
	return v, ok
}

// ConfigFile is a filename and the contents of the file as a Dict
type ConfigFile struct {
	Filename string
	Config   map[string]interface{}
}

// Config is a full compose file configuration
type Config struct {
	Filename   string                     `yaml:"-" json:"-"`
	Version    string                     `json:"version"`
	Services   Services                   `json:"services"`
	Networks   map[string]NetworkConfig   `yaml:",omitempty" json:"networks,omitempty"`
	Volumes    map[string]VolumeConfig    `yaml:",omitempty" json:"volumes,omitempty"`
	Secrets    map[string]SecretConfig    `yaml:",omitempty" json:"secrets,omitempty"`
	Configs    map[string]ConfigObjConfig `yaml:",omitempty" json:"configs,omitempty"`
	Extensions map[string]interface{}     `yaml:",inline" json:"-"`
}

// ServiceNames return names for all services in this Compose config
func (c Config) ServiceNames() []string {
	names := []string{}
	for _, s := range c.Services {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return names
}

// VolumeNames return names for all volumes in this Compose config
func (c Config) VolumeNames() []string {
	names := []string{}
	for k := range c.Volumes {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// NetworkNames return names for all volumes in this Compose config
func (c Config) NetworkNames() []string {
	names := []string{}
	for k := range c.Networks {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// SecretNames return names for all secrets in this Compose config
func (c Config) SecretNames() []string {
	names := []string{}
	for k := range c.Secrets {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// ConfigNames return names for all configs in this Compose config
func (c Config) ConfigNames() []string {
	names := []string{}
	for k := range c.Configs {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// GetServices retrieve services by names, or return all services if no name specified
func (c Config) GetServices(names []string) (Services, error) {
	if len(names) == 0 {
		return c.Services, nil
	}
	services := Services{}
	for _, name := range names {
		service, err := c.GetService(name)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}

// GetService retrieve a specific service by name
func (c Config) GetService(name string) (ServiceConfig, error) {
	for _, s := range c.Services {
		if s.Name == name {
			return s, nil
		}
	}
	return ServiceConfig{}, fmt.Errorf("no such service: %s", name)
}

type ServiceFunc func(service ServiceConfig) error

// WithServices run ServiceFunc on each service and dependencies in dependency order
func (c Config) WithServices(names []string, fn ServiceFunc) error {
	return c.withServices(names, fn, map[string]bool{})
}

func (c Config) withServices(names []string, fn ServiceFunc, done map[string]bool) error {
	services, err := c.GetServices(names)
	if err != nil {
		return err
	}
	for _, service := range services {
		if done[service.Name] {
			continue
		}
		dependencies := service.GetDependencies()
		if len(dependencies) > 0 {
			err := c.withServices(dependencies, fn, done)
			if err != nil {
				return err
			}
		}
		if err := fn(service); err != nil {
			return err
		}
		done[service.Name] = true
	}
	return nil
}

// MarshalJSON makes Config implement json.Marshaler
func (c Config) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"version":  c.Version,
		"services": c.Services,
	}

	if len(c.Networks) > 0 {
		m["networks"] = c.Networks
	}
	if len(c.Volumes) > 0 {
		m["volumes"] = c.Volumes
	}
	if len(c.Secrets) > 0 {
		m["secrets"] = c.Secrets
	}
	if len(c.Configs) > 0 {
		m["configs"] = c.Configs
	}
	for k, v := range c.Extensions {
		m[k] = v
	}
	return json.Marshal(m)
}
