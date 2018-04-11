// Copyright (c) 2018 Uber Technologies, Inc.
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
	// DefaultExcludePrefixes are the default prefixes to exclude.
	DefaultExcludePrefixes = []string{
		"vendor",
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
// absolute, except for the internal/x/text package.
type Config struct {
	// The working directory path.
	// Expected to be absolute path.
	DirPath string
	// The prefixes to exclude.
	// Expected to be absolute paths.
	// Expected to be unique.
	ExcludePrefixes []string
	// The compile config.
	Compile CompileConfig
	// Lint is a special case. If nothing is set, the defaults are used. Either IDs,
	// or Group/IncludeIDs/ExcludeIDs can be set, but not both. There can be no overlap
	// between IncludeIDs and ExcludeIDs.
	Lint LintConfig
	// The format config.
	Format FormatConfig
	// The gen config.
	Gen GenConfig
}

// CompileConfig is the compile config.
type CompileConfig struct {
	// The Protobuf version to use from https://github.com/google/protobuf/releases.
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

// LintConfig is the lint config.
//
// Either IDs, or Group/IncludeIDs/ExcludeIDs can be set, but not both.
type LintConfig struct {
	// IDs are the list of linter IDs to use.
	// Expected to not be set if Group/IncludeIDs/ExcludeIDs are set.
	// Expected to be all uppercase.
	// Expected to be unique.
	IDs []string
	// Group is the name of the lint group to use.
	// Expected to not be set if IDs is set.
	// Expected to be all lowercase.
	Group string
	// IncludeIDs are the list of linter IDs to use in addition to the defaults.
	// Expected to not be set if IDs is set.
	// Expected to be all uppercase.
	// Expected to be unique.
	// Expected to have no overlap with ExcludeIDs.
	IncludeIDs []string
	// ExcludeIDs are the list of linter IDs to exclude from the defaults.
	// Expected to not be set if IDs is set.
	// Expected to be all uppercase.
	// Expected to be unique.
	// Expected to have no overlap with IncludeIDs.
	ExcludeIDs []string
	// IgnoreIDToFilePaths is the map of ID to absolute file path to ignore.
	// IDs expected to be all upper-case.
	// File paths expected to be absolute paths.
	IgnoreIDToFilePaths map[string][]string
}

// FormatConfig is the format config.
type FormatConfig struct {
	// The indent to use. This is the actual Golang string to use as the indent,
	// where the external repesentation will be Xt or Xs, where X >= 1 and "t"
	// represents tabs, "s" represents spaces.
	// If empty, use two spaces.
	Indent string
	// Use semicolons to finish RPC definitions when possible, ie when the associated
	// RPC hs no options. Otherwise always use {}.
	RPCUseSemicolons bool
	// Trim the newline from the end of the file. Otherwise ends the file with a newline.
	TrimNewline bool
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
	// The base import path. This should be the go path of the prototool.yaml file.
	// This is required for go plugins.
	ImportPath string
	// Do not include default modifiers with Mfile=package.
	// By default, modifiers are included for the Well-Known Types, and for
	// all files in the compilation relative to the import path.
	// Generally do not set this unless you know what you are doing.
	NoDefaultModifiers bool
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
	Path string
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
	Excludes           []string `json:"excludes,omitempty" yaml:"excludes,omitempty"`
	NoDefaultExcludes  bool     `json:"no_default_excludes,omitempty" yaml:"no_default_excludes,omitempty"`
	ProtocVersion      string   `json:"protoc_version,omitempty" yaml:"protoc_version,omitempty"`
	ProtocIncludes     []string `json:"protoc_includes,omitempty" yaml:"protoc_includes,omitempty"`
	ProtocIncludeWKT   bool     `json:"protoc_include_wkt,omitempty" yaml:"protoc_include_wkt,omitempty"`
	AllowUnusedImports bool     `json:"allow_unused_imports,omitempty" yaml:"allow_unused_imports,omitempty"`
	Lint               struct {
		IDs             []string            `json:"ids,omitempty" yaml:"ids,omitempty"`
		Group           string              `json:"group,omitempty" yaml:"group,omitempty"`
		IncludeIDs      []string            `json:"include_ids,omitempty" yaml:"include_ids,omitempty"`
		ExcludeIDs      []string            `json:"exclude_ids,omitempty" yaml:"exclude_ids,omitempty"`
		IgnoreIDToFiles map[string][]string `json:"ignore_id_to_files,omitempty" yaml:"ignore_id_to_files,omitempty"`
	} `json:"lint,omitempty" yaml:"lint,omitempty"`
	Format struct {
		Indent           string `json:"indent,omitempty" yaml:"indent,omitempty"`
		RPCUseSemicolons bool   `json:"rpc_use_semicolons,omitempty" yaml:"rpc_use_semicolons,omitempty"`
		TrimNewline      bool   `json:"trim_newline,omitempty" yaml:"trim_newline,omitempty"`
	} ` json:"format,omitempty" yaml:"format,omitempty"`
	Gen struct {
		GoOptions struct {
			ImportPath         string            `json:"import_path,omitempty" yaml:"import_path,omitempty"`
			NoDefaultModifiers bool              `json:"no_default_modifiers,omitempty" yaml:"no_default_modifiers,omitempty"`
			ExtraModifiers     map[string]string `json:"extra_modifiers,omitempty" yaml:"extra_modifiers,omitempty"`
		} `json:"go_options,omitempty" yaml:"go_options,omitempty"`
		PluginOverrides map[string]string `json:"plugin_overrides,omitempty" yaml:"plugin_overrides,omitempty"`
		Plugins         []struct {
			Name   string `json:"name,omitempty" yaml:"name,omitempty"`
			Type   string `json:"type,omitempty" yaml:"type,omitempty"`
			Flags  string `json:"flags,omitempty" yaml:"flags,omitempty"`
			Output string `json:"output,omitempty" yaml:"output,omitempty"`
		} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	} `json:"gen,omitempty" yaml:"gen,omitempty"`
}

// ConfigProvider provides Configs.
type ConfigProvider interface {
	// GetForDir tries to find a file named DefaultConfigFilename starting in the
	// given directory, and going up a directory until hitting root.
	//
	// The directory must be an absolute path.
	//
	// If such a file is found, it is read as an ExternalConfig and converted to a Config.
	// If no such file is found, Config{} is returned.
	GetForDir(dirPath string) (Config, error)
	// Get tries to find a file named filePath with a config.
	//
	// The path must be an absolute path.
	//
	// If such a file is found, it is read as an ExternalConfig and converted to a Config.
	// If no such file is found, Config{} is returned.
	Get(filePath string) (Config, error)
	// GetFilePathForDir tries to find a file named DefaultConfigFilename starting in the
	// given directory, and going up a directory until hitting root.
	//
	// The directory must be an absolute path.
	//
	// If such a file is found, it is returned.
	// If no such file is found, "" is returned.
	GetFilePathForDir(dirPath string) (string, error)

	// GetForDir tries to find a file named DefaultConfigFilename in the given
	// directory and returns the cleaned absolute exclude prefixes. Unlike other functions
	// on ConfigProvider, this has no recursive functionality - if there is no
	// config file, nothing is returned.
	GetExcludePrefixesForDir(dirPath string) ([]string, error)
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

// NewConfigProvider returns a new ConfigProvider.
func NewConfigProvider(options ...ConfigProviderOption) ConfigProvider {
	return newConfigProvider(options...)
}
