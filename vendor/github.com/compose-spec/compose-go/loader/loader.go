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

package loader

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/envfile"
	interp "github.com/compose-spec/compose-go/interpolation"
	"github.com/compose-spec/compose-go/schema"
	"github.com/compose-spec/compose-go/template"
	"github.com/compose-spec/compose-go/types"
	units "github.com/docker/go-units"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Options supported by Load
type Options struct {
	// Skip schema validation
	SkipValidation bool
	// Skip interpolation
	SkipInterpolation bool
	// Interpolation options
	Interpolate *interp.Options
	// Discard 'env_file' entries after resolving to 'environment' section
	discardEnvFiles bool
}

// WithDiscardEnvFiles sets the Options to discard the `env_file` section after resolving to
// the `environment` section
func WithDiscardEnvFiles(opts *Options) {
	opts.discardEnvFiles = true
}

// ParseYAML reads the bytes from a file, parses the bytes into a mapping
// structure, and returns it.
func ParseYAML(source []byte) (map[string]interface{}, error) {
	var cfg interface{}
	if err := yaml.Unmarshal(source, &cfg); err != nil {
		return nil, err
	}
	cfgMap, ok := cfg.(map[interface{}]interface{})
	if !ok {
		return nil, errors.Errorf("Top-level object must be a mapping")
	}
	converted, err := convertToStringKeysRecursive(cfgMap, "")
	if err != nil {
		return nil, err
	}
	return converted.(map[string]interface{}), nil
}

// Load reads a ConfigDetails and returns a fully loaded configuration
func Load(configDetails types.ConfigDetails, options ...func(*Options)) (*types.Config, error) {
	if len(configDetails.ConfigFiles) < 1 {
		return nil, errors.Errorf("No files specified")
	}

	opts := &Options{
		Interpolate: &interp.Options{
			Substitute:      template.Substitute,
			LookupValue:     configDetails.LookupEnv,
			TypeCastMapping: interpolateTypeCastMapping,
		},
	}

	for _, op := range options {
		op(opts)
	}

	configs := []*types.Config{}
	var err error

	for _, file := range configDetails.ConfigFiles {
		configDict := file.Config
		version := schema.Version(configDict)
		if configDetails.Version == "" {
			configDetails.Version = version
		}
		if configDetails.Version != version {
			return nil, errors.Errorf("version mismatched between two composefiles : %v and %v", configDetails.Version, version)
		}

		if !opts.SkipInterpolation {
			configDict, err = interpolateConfig(configDict, *opts.Interpolate)
			if err != nil {
				return nil, err
			}
		}

		if !opts.SkipValidation {
			if err := schema.Validate(configDict, configDetails.Version); err != nil {
				return nil, err
			}
		}

		configDict = groupXFieldsIntoExtensions(configDict)

		cfg, err := loadSections(configDict, configDetails)
		if err != nil {
			return nil, err
		}
		cfg.Filename = file.Filename
		if opts.discardEnvFiles {
			for i := range cfg.Services {
				cfg.Services[i].EnvFile = nil
			}
		}

		configs = append(configs, cfg)
	}

	return merge(configs)
}

func groupXFieldsIntoExtensions(dict map[string]interface{}) map[string]interface{} {
	extras := map[string]interface{}{}
	for key, value := range dict {
		if strings.HasPrefix(key, "x-") {
			extras[key] = value
			delete(dict, key)
		}
		if d, ok := value.(map[string]interface{}); ok {
			dict[key] = groupXFieldsIntoExtensions(d)
		}
	}
	if len(extras) > 0 {
		dict["extensions"] = extras
	}
	return dict
}

func loadSections(config map[string]interface{}, configDetails types.ConfigDetails) (*types.Config, error) {
	var err error
	cfg := types.Config{
		Version: schema.Version(config),
	}

	var loaders = []struct {
		key string
		fnc func(config map[string]interface{}) error
	}{
		{
			key: "services",
			fnc: func(config map[string]interface{}) error {
				cfg.Services, err = LoadServices(config, configDetails.WorkingDir, configDetails.LookupEnv)
				return err
			},
		},
		{
			key: "networks",
			fnc: func(config map[string]interface{}) error {
				cfg.Networks, err = LoadNetworks(config, configDetails.Version)
				return err
			},
		},
		{
			key: "volumes",
			fnc: func(config map[string]interface{}) error {
				cfg.Volumes, err = LoadVolumes(config)
				return err
			},
		},
		{
			key: "secrets",
			fnc: func(config map[string]interface{}) error {
				cfg.Secrets, err = LoadSecrets(config, configDetails)
				return err
			},
		},
		{
			key: "configs",
			fnc: func(config map[string]interface{}) error {
				cfg.Configs, err = LoadConfigObjs(config, configDetails)
				return err
			},
		},
		{
			key: "extensions",
			fnc: func(config map[string]interface{}) error {
				if len(config) > 0 {
					cfg.Extensions = config
				}
				return err
			},
		},
	}
	for _, loader := range loaders {
		if err := loader.fnc(getSection(config, loader.key)); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func getSection(config map[string]interface{}, key string) map[string]interface{} {
	section, ok := config[key]
	if !ok {
		return make(map[string]interface{})
	}
	return section.(map[string]interface{})
}

func sortedKeys(set map[string]bool) []string {
	var keys []string
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func getProperties(services map[string]interface{}, propertyMap map[string]string) map[string]string {
	output := map[string]string{}

	for _, service := range services {
		if serviceDict, ok := service.(map[string]interface{}); ok {
			for property, description := range propertyMap {
				if _, isSet := serviceDict[property]; isSet {
					output[property] = description
				}
			}
		}
	}

	return output
}

// ForbiddenPropertiesError is returned when there are properties in the Compose
// file that are forbidden.
type ForbiddenPropertiesError struct {
	Properties map[string]string
}

func (e *ForbiddenPropertiesError) Error() string {
	return "Configuration contains forbidden properties"
}

func getServices(configDict map[string]interface{}) map[string]interface{} {
	if services, ok := configDict["services"]; ok {
		if servicesDict, ok := services.(map[string]interface{}); ok {
			return servicesDict
		}
	}

	return map[string]interface{}{}
}

// Transform converts the source into the target struct with compose types transformer
// and the specified transformers if any.
func Transform(source interface{}, target interface{}, additionalTransformers ...Transformer) error {
	data := mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			createTransformHook(additionalTransformers...),
			mapstructure.StringToTimeDurationHookFunc()),
		Result:   target,
		Metadata: &data,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(source)
}

// TransformerFunc defines a function to perform the actual transformation
type TransformerFunc func(interface{}) (interface{}, error)

// Transformer defines a map to type transformer
type Transformer struct {
	TypeOf reflect.Type
	Func   TransformerFunc
}

func createTransformHook(additionalTransformers ...Transformer) mapstructure.DecodeHookFuncType {
	transforms := map[reflect.Type]func(interface{}) (interface{}, error){
		reflect.TypeOf(types.External{}):                         transformExternal,
		reflect.TypeOf(types.HealthCheckTest{}):                  transformHealthCheckTest,
		reflect.TypeOf(types.ShellCommand{}):                     transformShellCommand,
		reflect.TypeOf(types.StringList{}):                       transformStringList,
		reflect.TypeOf(map[string]string{}):                      transformMapStringString,
		reflect.TypeOf(types.UlimitsConfig{}):                    transformUlimits,
		reflect.TypeOf(types.UnitBytes(0)):                       transformSize,
		reflect.TypeOf([]types.ServicePortConfig{}):              transformServicePort,
		reflect.TypeOf(types.ServiceSecretConfig{}):              transformStringSourceMap,
		reflect.TypeOf(types.ServiceConfigObjConfig{}):           transformStringSourceMap,
		reflect.TypeOf(types.StringOrNumberList{}):               transformStringOrNumberList,
		reflect.TypeOf(map[string]*types.ServiceNetworkConfig{}): transformServiceNetworkMap,
		reflect.TypeOf(types.Mapping{}):                          transformMappingOrListFunc("=", false),
		reflect.TypeOf(types.MappingWithEquals{}):                transformMappingOrListFunc("=", true),
		reflect.TypeOf(types.Labels{}):                           transformMappingOrListFunc("=", false),
		reflect.TypeOf(types.MappingWithColon{}):                 transformMappingOrListFunc(":", false),
		reflect.TypeOf(types.HostsList{}):                        transformListOrMappingFunc(":", false),
		reflect.TypeOf(types.ServiceVolumeConfig{}):              transformServiceVolumeConfig,
		reflect.TypeOf(types.BuildConfig{}):                      transformBuildConfig,
		reflect.TypeOf(types.Duration(0)):                        transformStringToDuration,
	}

	for _, transformer := range additionalTransformers {
		transforms[transformer.TypeOf] = transformer.Func
	}

	return func(_ reflect.Type, target reflect.Type, data interface{}) (interface{}, error) {
		transform, ok := transforms[target]
		if !ok {
			return data, nil
		}
		return transform(data)
	}
}

// keys needs to be converted to strings for jsonschema
func convertToStringKeysRecursive(value interface{}, keyPrefix string) (interface{}, error) {
	if mapping, ok := value.(map[interface{}]interface{}); ok {
		dict := make(map[string]interface{})
		for key, entry := range mapping {
			str, ok := key.(string)
			if !ok {
				return nil, formatInvalidKeyError(keyPrefix, key)
			}
			var newKeyPrefix string
			if keyPrefix == "" {
				newKeyPrefix = str
			} else {
				newKeyPrefix = fmt.Sprintf("%s.%s", keyPrefix, str)
			}
			convertedEntry, err := convertToStringKeysRecursive(entry, newKeyPrefix)
			if err != nil {
				return nil, err
			}
			dict[str] = convertedEntry
		}
		return dict, nil
	}
	if list, ok := value.([]interface{}); ok {
		var convertedList []interface{}
		for index, entry := range list {
			newKeyPrefix := fmt.Sprintf("%s[%d]", keyPrefix, index)
			convertedEntry, err := convertToStringKeysRecursive(entry, newKeyPrefix)
			if err != nil {
				return nil, err
			}
			convertedList = append(convertedList, convertedEntry)
		}
		return convertedList, nil
	}
	return value, nil
}

func formatInvalidKeyError(keyPrefix string, key interface{}) error {
	var location string
	if keyPrefix == "" {
		location = "at top level"
	} else {
		location = fmt.Sprintf("in %s", keyPrefix)
	}
	return errors.Errorf("Non-string key %s: %#v", location, key)
}

// LoadServices produces a ServiceConfig map from a compose file Dict
// the servicesDict is not validated if directly used. Use Load() to enable validation
func LoadServices(servicesDict map[string]interface{}, workingDir string, lookupEnv template.Mapping) ([]types.ServiceConfig, error) {
	var services []types.ServiceConfig

	for name, serviceDef := range servicesDict {
		serviceConfig, err := LoadService(name, serviceDef.(map[string]interface{}), workingDir, lookupEnv)
		if err != nil {
			return nil, err
		}
		services = append(services, *serviceConfig)
	}

	return services, nil
}

// LoadService produces a single ServiceConfig from a compose file Dict
// the serviceDict is not validated if directly used. Use Load() to enable validation
func LoadService(name string, serviceDict map[string]interface{}, workingDir string, lookupEnv template.Mapping) (*types.ServiceConfig, error) {
	serviceConfig := &types.ServiceConfig{}
	if err := Transform(serviceDict, serviceConfig); err != nil {
		return nil, err
	}
	serviceConfig.Name = name

	if err := resolveEnvironment(serviceConfig, workingDir, lookupEnv); err != nil {
		return nil, err
	}

	if err := resolveVolumePaths(serviceConfig.Volumes, workingDir, lookupEnv); err != nil {
		return nil, err
	}

	return serviceConfig, nil
}

func resolveEnvironment(serviceConfig *types.ServiceConfig, workingDir string, lookupEnv template.Mapping) error {
	environment := types.MappingWithEquals{}

	if len(serviceConfig.EnvFile) > 0 {
		for _, file := range serviceConfig.EnvFile {
			filePath := absPath(workingDir, file)
			fileVars, err := envfile.Parse(filePath)
			if err != nil {
				return err
			}
			environment.OverrideBy(fileVars.Resolve(lookupEnv).RemoveEmpty())
		}
	}

	environment.OverrideBy(serviceConfig.Environment.Resolve(lookupEnv))
	serviceConfig.Environment = environment
	return nil
}

func resolveVolumePaths(volumes []types.ServiceVolumeConfig, workingDir string, lookupEnv template.Mapping) error {
	for i, volume := range volumes {
		if volume.Type != "bind" {
			continue
		}

		if volume.Source == "" {
			return errors.New(`invalid mount config for type "bind": field Source must not be empty`)
		}

		filePath := expandUser(volume.Source, lookupEnv)
		// Check if source is an absolute path (either Unix or Windows), to
		// handle a Windows client with a Unix daemon or vice-versa.
		//
		// Note that this is not required for Docker for Windows when specifying
		// a local Windows path, because Docker for Windows translates the Windows
		// path into a valid path within the VM.
		if !path.IsAbs(filePath) && !isAbs(filePath) {
			filePath = absPath(workingDir, filePath)
		}
		volume.Source = filePath
		volumes[i] = volume
	}
	return nil
}

// TODO: make this more robust
func expandUser(path string, lookupEnv template.Mapping) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			logrus.Warn("cannot expand '~', because the environment lacks HOME")
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func transformUlimits(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case int:
		return types.UlimitsConfig{Single: value}, nil
	case map[string]interface{}:
		ulimit := types.UlimitsConfig{}
		if v, ok := value["soft"]; ok {
			ulimit.Soft = v.(int)
		}
		if v, ok := value["hard"]; ok {
			ulimit.Hard = v.(int)
		}
		return ulimit, nil
	default:
		return data, errors.Errorf("invalid type %T for ulimits", value)
	}
}

// LoadNetworks produces a NetworkConfig map from a compose file Dict
// the source Dict is not validated if directly used. Use Load() to enable validation
func LoadNetworks(source map[string]interface{}, version string) (map[string]types.NetworkConfig, error) {
	networks := make(map[string]types.NetworkConfig)
	err := Transform(source, &networks)
	if err != nil {
		return networks, err
	}
	for name, network := range networks {
		if !network.External.External {
			continue
		}
		switch {
		case network.External.Name != "":
			if network.Name != "" {
				return nil, errors.Errorf("network %s: network.external.name and network.name conflict; only use network.name", name)
			}
			logrus.Warnf("network %s: network.external.name is deprecated in favor of network.name", name)
			network.Name = network.External.Name
			network.External.Name = ""
		case network.Name == "":
			network.Name = name
		}
		networks[name] = network
	}
	return networks, nil
}

func externalVolumeError(volume, key string) error {
	return errors.Errorf(
		"conflicting parameters \"external\" and %q specified for volume %q",
		key, volume)
}

// LoadVolumes produces a VolumeConfig map from a compose file Dict
// the source Dict is not validated if directly used. Use Load() to enable validation
func LoadVolumes(source map[string]interface{}) (map[string]types.VolumeConfig, error) {
	volumes := make(map[string]types.VolumeConfig)
	if err := Transform(source, &volumes); err != nil {
		return volumes, err
	}

	for name, volume := range volumes {
		if !volume.External.External {
			continue
		}
		switch {
		case volume.Driver != "":
			return nil, externalVolumeError(name, "driver")
		case len(volume.DriverOpts) > 0:
			return nil, externalVolumeError(name, "driver_opts")
		case len(volume.Labels) > 0:
			return nil, externalVolumeError(name, "labels")
		case volume.External.Name != "":
			if volume.Name != "" {
				return nil, errors.Errorf("volume %s: volume.external.name and volume.name conflict; only use volume.name", name)
			}
			logrus.Warnf("volume %s: volume.external.name is deprecated in favor of volume.name", name)
			volume.Name = volume.External.Name
			volume.External.Name = ""
		case volume.Name == "":
			volume.Name = name
		}
		volumes[name] = volume
	}
	return volumes, nil
}

// LoadSecrets produces a SecretConfig map from a compose file Dict
// the source Dict is not validated if directly used. Use Load() to enable validation
func LoadSecrets(source map[string]interface{}, details types.ConfigDetails) (map[string]types.SecretConfig, error) {
	secrets := make(map[string]types.SecretConfig)
	if err := Transform(source, &secrets); err != nil {
		return secrets, err
	}
	for name, secret := range secrets {
		obj, err := loadFileObjectConfig(name, "secret", types.FileObjectConfig(secret), details)
		if err != nil {
			return nil, err
		}
		secretConfig := types.SecretConfig(obj)
		secrets[name] = secretConfig
	}
	return secrets, nil
}

// LoadConfigObjs produces a ConfigObjConfig map from a compose file Dict
// the source Dict is not validated if directly used. Use Load() to enable validation
func LoadConfigObjs(source map[string]interface{}, details types.ConfigDetails) (map[string]types.ConfigObjConfig, error) {
	configs := make(map[string]types.ConfigObjConfig)
	if err := Transform(source, &configs); err != nil {
		return configs, err
	}
	for name, config := range configs {
		obj, err := loadFileObjectConfig(name, "config", types.FileObjectConfig(config), details)
		if err != nil {
			return nil, err
		}
		configConfig := types.ConfigObjConfig(obj)
		configs[name] = configConfig
	}
	return configs, nil
}

func loadFileObjectConfig(name string, objType string, obj types.FileObjectConfig, details types.ConfigDetails) (types.FileObjectConfig, error) {
	// if "external: true"
	switch {
	case obj.External.External:
		// handle deprecated external.name
		if obj.External.Name != "" {
			if obj.Name != "" {
				return obj, errors.Errorf("%[1]s %[2]s: %[1]s.external.name and %[1]s.name conflict; only use %[1]s.name", objType, name)
			}
			logrus.Warnf("%[1]s %[2]s: %[1]s.external.name is deprecated in favor of %[1]s.name", objType, name)
			obj.Name = obj.External.Name
			obj.External.Name = ""
		} else {
			if obj.Name == "" {
				obj.Name = name
			}
		}
		// if not "external: true"
	case obj.Driver != "":
		if obj.File != "" {
			return obj, errors.Errorf("%[1]s %[2]s: %[1]s.driver and %[1]s.file conflict; only use %[1]s.driver", objType, name)
		}
	default:
		obj.File = absPath(details.WorkingDir, obj.File)
	}

	return obj, nil
}

func absPath(workingDir string, filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}
	return filepath.Join(workingDir, filePath)
}

var transformMapStringString TransformerFunc = func(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case map[string]interface{}:
		return toMapStringString(value, false), nil
	case map[string]string:
		return value, nil
	default:
		return data, errors.Errorf("invalid type %T for map[string]string", value)
	}
}

var transformExternal TransformerFunc = func(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case bool:
		return map[string]interface{}{"external": value}, nil
	case map[string]interface{}:
		return map[string]interface{}{"external": true, "name": value["name"]}, nil
	default:
		return data, errors.Errorf("invalid type %T for external", value)
	}
}

var transformServicePort TransformerFunc = func(data interface{}) (interface{}, error) {
	switch entries := data.(type) {
	case []interface{}:
		// We process the list instead of individual items here.
		// The reason is that one entry might be mapped to multiple ServicePortConfig.
		// Therefore we take an input of a list and return an output of a list.
		ports := []interface{}{}
		for _, entry := range entries {
			switch value := entry.(type) {
			case int:
				parsed, err := types.ParsePortConfig(fmt.Sprint(value))
				if err != nil {
					return data, err
				}
				for _, v := range parsed {
					ports = append(ports, v)
				}
			case string:
				parsed, err := types.ParsePortConfig(value)
				if err != nil {
					return data, err
				}
				for _, v := range parsed {
					ports = append(ports, v)
				}
			case map[string]interface{}:
				ports = append(ports, value)
			default:
				return data, errors.Errorf("invalid type %T for port", value)
			}
		}
		return ports, nil
	default:
		return data, errors.Errorf("invalid type %T for port", entries)
	}
}

var transformStringSourceMap TransformerFunc = func(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case string:
		return map[string]interface{}{"source": value}, nil
	case map[string]interface{}:
		return data, nil
	default:
		return data, errors.Errorf("invalid type %T for secret", value)
	}
}

var transformBuildConfig TransformerFunc = func(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case string:
		return map[string]interface{}{"context": value}, nil
	case map[string]interface{}:
		return data, nil
	default:
		return data, errors.Errorf("invalid type %T for service build", value)
	}
}

var transformServiceVolumeConfig TransformerFunc = func(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case string:
		return ParseVolume(value)
	case map[string]interface{}:
		return data, nil
	default:
		return data, errors.Errorf("invalid type %T for service volume", value)
	}
}

var transformServiceNetworkMap TransformerFunc = func(value interface{}) (interface{}, error) {
	if list, ok := value.([]interface{}); ok {
		mapValue := map[interface{}]interface{}{}
		for _, name := range list {
			mapValue[name] = nil
		}
		return mapValue, nil
	}
	return value, nil
}

var transformStringOrNumberList TransformerFunc = func(value interface{}) (interface{}, error) {
	list := value.([]interface{})
	result := make([]string, len(list))
	for i, item := range list {
		result[i] = fmt.Sprint(item)
	}
	return result, nil
}

var transformStringList TransformerFunc = func(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case string:
		return []string{value}, nil
	case []interface{}:
		return value, nil
	default:
		return data, errors.Errorf("invalid type %T for string list", value)
	}
}

func transformMappingOrListFunc(sep string, allowNil bool) TransformerFunc {
	return func(data interface{}) (interface{}, error) {
		return transformMappingOrList(data, sep, allowNil), nil
	}
}

func transformListOrMappingFunc(sep string, allowNil bool) TransformerFunc {
	return func(data interface{}) (interface{}, error) {
		return transformListOrMapping(data, sep, allowNil), nil
	}
}

func transformListOrMapping(listOrMapping interface{}, sep string, allowNil bool) interface{} {
	switch value := listOrMapping.(type) {
	case map[string]interface{}:
		return toStringList(value, sep, allowNil)
	case []interface{}:
		return listOrMapping
	}
	panic(errors.Errorf("expected a map or a list, got %T: %#v", listOrMapping, listOrMapping))
}

func transformMappingOrList(mappingOrList interface{}, sep string, allowNil bool) interface{} {
	switch value := mappingOrList.(type) {
	case map[string]interface{}:
		return toMapStringString(value, allowNil)
	case ([]interface{}):
		result := make(map[string]interface{})
		for _, value := range value {
			parts := strings.SplitN(value.(string), sep, 2)
			key := parts[0]
			switch {
			case len(parts) == 1 && allowNil:
				result[key] = nil
			case len(parts) == 1 && !allowNil:
				result[key] = ""
			default:
				result[key] = parts[1]
			}
		}
		return result
	}
	panic(errors.Errorf("expected a map or a list, got %T: %#v", mappingOrList, mappingOrList))
}

var transformShellCommand TransformerFunc = func(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return shellwords.Parse(str)
	}
	return value, nil
}

var transformHealthCheckTest TransformerFunc = func(data interface{}) (interface{}, error) {
	switch value := data.(type) {
	case string:
		return append([]string{"CMD-SHELL"}, value), nil
	case []interface{}:
		return value, nil
	default:
		return value, errors.Errorf("invalid type %T for healthcheck.test", value)
	}
}

var transformSize TransformerFunc = func(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case int:
		return int64(value), nil
	case string:
		return units.RAMInBytes(value)
	}
	panic(errors.Errorf("invalid type for size %T", value))
}

var transformStringToDuration TransformerFunc = func(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case string:
		d, err := time.ParseDuration(value)
		if err != nil {
			return value, err
		}
		return types.Duration(d), nil
	default:
		return value, errors.Errorf("invalid type %T for duration", value)
	}
}

func toMapStringString(value map[string]interface{}, allowNil bool) map[string]interface{} {
	output := make(map[string]interface{})
	for key, value := range value {
		output[key] = toString(value, allowNil)
	}
	return output
}

func toString(value interface{}, allowNil bool) interface{} {
	switch {
	case value != nil:
		return fmt.Sprint(value)
	case allowNil:
		return nil
	default:
		return ""
	}
}

func toStringList(value map[string]interface{}, separator string, allowNil bool) []string {
	output := []string{}
	for key, value := range value {
		if value == nil && !allowNil {
			continue
		}
		output = append(output, fmt.Sprintf("%s%s%s", key, separator, value))
	}
	sort.Strings(output)
	return output
}
