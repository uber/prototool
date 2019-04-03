// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package settings

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const (
	// DefaultConfigFilename is the default config filename.
	DefaultConfigFilename = "prototool.yaml"

	// GenPluginTypeNone says there is no specific plugin type.
	GenPluginTypeNone GenPluginType = iota
	// GenPluginTypeGo says the plugin is a Golang plugin that
	// is or uses github.com/golang/protobuf.
	// This will use GenGoPluginOptions.
	GenPluginTypeGo
	// GenPluginTypeGogo says the plugin is a Golang plugin that
	// is or uses github.com/gogo/protobuf.
	// This will use GenGoPluginOptions.
	GenPluginTypeGogo
)

var (
	// ConfigFilenames are all possible config filenames.
	ConfigFilenames = []string{
		DefaultConfigFilename,
		"prototool.json",
	}

	_genPluginTypeToString = map[GenPluginType]string{
		GenPluginTypeNone: "",
		GenPluginTypeGo:   "go",
		GenPluginTypeGogo: "gogo",
	}
	_stringToGenPluginType = map[string]GenPluginType{
		"":     GenPluginTypeNone,
		"go":   GenPluginTypeGo,
		"gogo": GenPluginTypeGogo,
	}

	_genPluginTypeToIsGo = map[GenPluginType]bool{
		GenPluginTypeNone: false,
		GenPluginTypeGo:   true,
		GenPluginTypeGogo: false,
	}
	_genPluginTypeToIsGogo = map[GenPluginType]bool{
		GenPluginTypeNone: false,
		GenPluginTypeGo:   false,
		GenPluginTypeGogo: true,
	}
)

// GenPluginType is a type of protoc plugin.
type GenPluginType int

// String implements fmt.Stringer.
func (g GenPluginType) String() string {
	if s, ok := _genPluginTypeToString[g]; ok {
		return s
	}
	return strconv.Itoa(int(g))
}

// The Is functions do not validate if the plugin type is known
// as this is supposed to be done in ConfigProvider.
// It's a lot easier if they just return a bool.

// IsGo returns true if the plugin type is associated with
// github.com/golang/protobuf.
func (g GenPluginType) IsGo() bool {
	return _genPluginTypeToIsGo[g]
}

// IsGogo returns true if the plugin type is associated with
// github.com/gogo/protobuf.
func (g GenPluginType) IsGogo() bool {
	return _genPluginTypeToIsGogo[g]
}

// ParseGenPluginType parses the GenPluginType from the given string.
//
// Input is case-insensitive.
func ParseGenPluginType(s string) (GenPluginType, error) {
	genPluginType, ok := _stringToGenPluginType[strings.ToLower(s)]
	if !ok {
		return GenPluginTypeNone, fmt.Errorf("could not parse %s to a GenPluginType", s)
	}
	return genPluginType, nil
}

// Config is the main config.
//
// Configs are derived from ExternalConfigs, which represent the Config
// in a more palpable format for configuration via a config file
// or flags.
//
// String slices will be deduped and sorted if returned from this package.
// Configs will be validated if returned from this package.
//
// All paths returned should be absolute paths. Outside of this package,
// all other internal packages should verify that all given paths are
// absolute, except for the internal/text package.
type Config struct {
	// The directory path of the config file, or the working directory path.
	// if no config file exists.
	// Expected to be absolute path.
	DirPath string
	// The prefixes to exclude.
	// Expected to be absolute paths.
	// Expected to be unique.
	ExcludePrefixes []string
	// The compile config.
	Compile CompileConfig
	// The create config.
	Create CreateConfig
	// Lint is a special case. If nothing is set, the defaults are used. Either IDs,
	// or Group/IncludeIDs/ExcludeIDs can be set, but not both. There can be no overlap
	// between IncludeIDs and ExcludeIDs.
	Lint LintConfig
	// The break config.
	Break BreakConfig
	// The gen config.
	Gen GenConfig
}

// CompileConfig is the compile config.
type CompileConfig struct {
	// The Protobuf version to use from https://github.com/protocolbuffers/protobuf/releases.
	// Must have a valid protoc zip file asset, so for example 3.5.0 is a valid version
	// but 3.5.0.1 is not.
	ProtobufVersion string
	// IncludePaths are the additional paths to include with -I to protoc.
	// Expected to be absolute paths.
	// Expected to be unique.
	IncludePaths []string
	// IncludeWellKnownTypes says to add the Google well-known types with -I to protoc.
	IncludeWellKnownTypes bool
	// AllowUnusedImports says to not error when an import is not used.
	AllowUnusedImports bool
}

// CreateConfig is the create config.
type CreateConfig struct {
	// The map from directory to the package to use as the base.
	// Directories expected to be absolute paths.
	DirPathToBasePackage map[string]string
}

// LintConfig is the lint config.
type LintConfig struct {
	// Group is the specific group of linters to use.
	// The default group is the "default" lint group, which is equal
	// to the "uber1" lint group.
	// Setting this value will result in NoDefault being ignored.
	Group string
	// NoDefault is set to exclude the default set of linters.
	// This value is ignored if Group is set.
	// Deprecated: Use group "empty" instead.
	NoDefault bool
	// IncludeIDs are the list of linter IDs to use in addition to the defaults.
	// Expected to be all uppercase.
	// Expected to be unique.
	// Expected to have no overlap with ExcludeIDs.
	IncludeIDs []string
	// ExcludeIDs are the list of linter IDs to exclude from the defaults.
	// Expected to be all uppercase.
	// Expected to be unique.
	// Expected to have no overlap with IncludeIDs.
	ExcludeIDs []string
	// IgnoreIDToFilePaths is the map of ID to absolute file path to ignore.
	// IDs expected to be all upper-case.
	// File paths expected to be absolute paths.
	IgnoreIDToFilePaths map[string][]string
	// FileHeader is contents of the file that contains the header for all
	// Protobuf files, typically a license header. If this is set and the
	// FILE_HEADER linter is turned on, files will be checked to begin
	// with the contents of this file, and format --fix will place this
	// header before the syntax declaration. Note that format --fix will delete
	// anything before the syntax declaration if this is set.
	FileHeader string
	// JavaPackagePrefix is the prefix for java packages. This only has an
	// effect if the linter FILE_OPTIONS_EQUAL_JAVA_PACKAGE_PREFIX is turned on.
	// This also affects create and format --fix.
	// The default behavior is to use "com".
	JavaPackagePrefix string
	// AllowSuppression says to honor @suppresswarnings annotations.
	AllowSuppression bool
}

// BreakConfig is the break config.
type BreakConfig struct {
	// Include beta packages in breaking change detection.
	IncludeBeta bool
	// Allow stable packages to depend on beta packages.
	// This is implicitly set if IncludeBeta is set.
	AllowBetaDeps bool
}

// GenConfig is the gen config.
type GenConfig struct {
	// The go plugin options.
	GoPluginOptions GenGoPluginOptions
	// The plugins.
	// These will be sorted by name if returned from this package.
	Plugins []GenPlugin
}

// GenGoPluginOptions are options for go plugins.
//
// This will be used for plugin types go, gogo, gogrpc, gogogrpc.
type GenGoPluginOptions struct {
	// The base import path. This should be the go path of the config file.
	// This is required for go plugins.
	ImportPath string
	// ExtraModifiers to include with Mfile=package.
	ExtraModifiers map[string]string
}

// GenPlugin is a plugin to use.
type GenPlugin struct {
	// The name of the plugin. For example, if you want to use
	// protoc-gen-gogoslick, the name is "gogoslick".
	Name string
	// The path to the executable. For example, if the name is "grpc-cpp"
	// but the path to the executable "protoc-gen-grpc-cpp" is "/usr/local/bin/grpc_cpp_plugin",
	// then this will be "/usr/local/bin/grpc_cpp_plugin".
	// This is a function so that we defer path lookups, there is functionality where we could
	// e.g. have this point to an absolute path on the system but we don't want it for all
	// calls, for example if you have a plugin installed in a Docker image but want to run
	// prototool grpc locally.
	// We could have made this a private field and attached a function to it but this keeps
	// the style of all config structs only having public fields.
	// https://github.com/uber/prototool/issues/325
	GetPath func() (string, error) `json:"-"`
	// The type, if any. This will be GenPluginTypeNone if
	// there is no specific type.
	Type GenPluginType
	// Extra flags to pass.
	// If there is an associated type, some flags may be generated,
	// for example plugins=grpc or Mfile=package modifiers.
	Flags string
	// The path to output to.
	// Must be relative in a config file.
	OutputPath OutputPath
	// If set, the output path will be set to "$OUTPUT_PATH/$(basename $OUTPUT_PATH).$FILE_SUFFIX"
	// Used for e.g. JAR generation for java or descriptor_set file name.
	FileSuffix string
	// Add the --include_imports flags to protoc.
	// Only valid if Name is descriptor_set.
	IncludeImports bool
	// Add the --include_source_info flags to protoc.
	// Only valid if Name is descriptor_set.
	IncludeSourceInfo bool
}

// OutputPath is an output path.
//
// We need the relative path for go package references for generation.
// TODO: we might want all paths to have the given path and absolute path,
// see if we need this.
type OutputPath struct {
	// Must be relative.
	RelPath string
	AbsPath string
}

// ExternalConfig is the external representation of Config.
//
// It is meant to be set by a YAML or JSON config file, or flags.
type ExternalConfig struct {
	Excludes []string `json:"excludes,omitempty" yaml:"excludes,omitempty"`
	Protoc   struct {
		AllowUnusedImports bool     `json:"allow_unused_imports,omitempty" yaml:"allow_unused_imports,omitempty"`
		Version            string   `json:"version,omitempty" yaml:"version,omitempty"`
		Includes           []string `json:"includes,omitempty" yaml:"includes,omitempty"`
	} `json:"protoc,omitempty" yaml:"protoc,omitempty"`
	Create struct {
		Packages []struct {
			Directory string `json:"directory,omitempty" yaml:"directory,omitempty"`
			Name      string `json:"name,omitempty" yaml:"name,omitempty"`
		} `json:"packages,omitempty" yaml:"packages,omitempty"`
	} `json:"create,omitempty" yaml:"create,omitempty"`
	Lint struct {
		Group   string `json:"group,omitempty" yaml:"group,omitempty"`
		Ignores []struct {
			ID    string   `json:"id,omitempty" yaml:"id,omitempty"`
			Files []string `json:"files,omitempty" yaml:"files,omitempty"`
		} `json:"ignores,omitempty" yaml:"ignores,omitempty"`
		Rules struct {
			NoDefault bool     `json:"no_default,omitempty" yaml:"no_default,omitempty"`
			Add       []string `json:"add,omitempty" yaml:"add,omitempty"`
			Remove    []string `json:"remove,omitempty" yaml:"remove,omitempty"`
		} `json:"rules,omitempty" yaml:"rules,omitempty"`
		FileHeader struct {
			Path        string `json:"path,omitempty" yaml:"path,omitempty"`
			Content     string `json:"content,omitempty" yaml:"content,omitempty"`
			IsCommented bool   `json:"is_commented,omitempty" yaml:"is_commented,omitempty"`
		} `json:"file_header,omitempty" yaml:"file_header,omitempty"`
		JavaPackagePrefix string `json:"java_package_prefix,omitempty" yaml:"java_package_prefix,omitempty"`
		// devel-mode only
		AllowSuppression bool `json:"allow_suppression,omitempty" yaml:"allow_suppression,omitempty"`
	} `json:"lint,omitempty" yaml:"lint,omitempty"`
	Break struct {
		IncludeBeta   bool `json:"include_beta,omitempty" yaml:"include_beta,omitempty"`
		AllowBetaDeps bool `json:"allow_beta_deps,omitempty" yaml:"allow_beta_deps,omitempty"`
	} `json:"break,omitempty" yaml:"break,omitempty"`
	Generate struct {
		GoOptions struct {
			ImportPath     string            `json:"import_path,omitempty" yaml:"import_path,omitempty"`
			ExtraModifiers map[string]string `json:"extra_modifiers,omitempty" yaml:"extra_modifiers,omitempty"`
		} `json:"go_options,omitempty" yaml:"go_options,omitempty"`
		Plugins []struct {
			Name              string `json:"name,omitempty" yaml:"name,omitempty"`
			Type              string `json:"type,omitempty" yaml:"type,omitempty"`
			Flags             string `json:"flags,omitempty" yaml:"flags,omitempty"`
			Output            string `json:"output,omitempty" yaml:"output,omitempty"`
			Path              string `json:"path,omitempty" yaml:"path,omitempty"`
			FileSuffix        string `json:"file_suffix,omitempty" yaml:"file_suffix,omitempty"`
			IncludeImports    bool   `json:"include_imports,omitempty" yaml:"include_imports,omitempty"`
			IncludeSourceInfo bool   `json:"include_source_info,omitempty" yaml:"include_source_info,omitempty"`
		} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	} `json:"generate,omitempty" yaml:"generate,omitempty"`
}

// ConfigProvider provides Configs.
type ConfigProvider interface {
	// GetForDir tries to find a file named by one of the ConfigFilenames starting in the
	// given directory, and going up a directory until hitting root.
	//
	// The directory must be an absolute path.
	//
	// If such a file is found, it is read as an ExternalConfig and converted to a Config.
	// If no such file is found, the default config is returned.
	// If multiple files named by one of the ConfigFilenames are found in the same
	// directory, error is returned.
	GetForDir(dirPath string) (Config, error)
	// Get tries to find a file named filePath with a config.
	//
	// The path must be an absolute path.
	// The file must have either the extension .yaml or .json.
	//
	// If such a file is found, it is read as an ExternalConfig and converted to a Config.
	// If no such file is found, error is returned.
	Get(filePath string) (Config, error)
	// GetFilePathForDir tries to find a file named by one of the ConfigFilenames starting in the
	// given directory, and going up a directory until hitting root.
	//
	// The directory must be an absolute path.
	//
	// If such a file is found, it is returned.
	// If no such file is found, "" is returned.
	// If multiple files named by one of the ConfigFilenames are found in the same
	// directory, error is returned.
	GetFilePathForDir(dirPath string) (string, error)
	// GetForData returns a Config for the given ExternalConfigData in JSON format.
	// The Config will be as if there was a configuration file at the given dirPath.
	GetForData(dirPath string, externalConfigData string) (Config, error)

	// GetExcludePrefixesForDir tries to find a file named by one of the ConfigFilenames in the given
	// directory and returns the cleaned absolute exclude prefixes. Unlike other functions
	// on ConfigProvider, this has no recursive functionality - if there is no
	// config file, nothing is returned.
	// If multiple files named by one of the ConfigFilenames are found in the same
	// directory, error is returned.
	GetExcludePrefixesForDir(dirPath string) ([]string, error)
	// GetExcludePrefixesForData gets the exclude prefixes for the given ExternalConfigData in JSON format.
	// The logic will act is if there was a configuration file at the given dirPath.
	GetExcludePrefixesForData(dirPath string, externalConfigData string) ([]string, error)
}

// ConfigProviderOption is an option for a new ConfigProvider.
type ConfigProviderOption func(*configProvider)

// ConfigProviderWithLogger returns a ConfigProviderOption that uses the given logger.
//
// The default is to use zap.NewNop().
func ConfigProviderWithLogger(logger *zap.Logger) ConfigProviderOption {
	return func(configProvider *configProvider) {
		configProvider.logger = logger
	}
}

// ConfigProviderWithDevelMode returns a ConfigProviderOption that allows devel-mode.
func ConfigProviderWithDevelMode() ConfigProviderOption {
	return func(configProvider *configProvider) {
		configProvider.develMode = true
	}
}

// NewConfigProvider returns a new ConfigProvider.
func NewConfigProvider(options ...ConfigProviderOption) ConfigProvider {
	return newConfigProvider(options...)
}
