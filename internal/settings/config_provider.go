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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/uber/prototool/internal/strs"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

type configProvider struct {
	logger    *zap.Logger
	develMode bool
}

func newConfigProvider(options ...ConfigProviderOption) *configProvider {
	configProvider := &configProvider{
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(configProvider)
	}
	return configProvider
}

func (c *configProvider) GetForDir(dirPath string) (Config, error) {
	filePath, err := c.GetFilePathForDir(dirPath)
	if err != nil {
		return Config{}, err
	}
	if filePath == "" {
		return getDefaultConfig(c.develMode, dirPath)
	}
	return c.Get(filePath)
}

func (c *configProvider) GetFilePathForDir(dirPath string) (string, error) {
	if !filepath.IsAbs(dirPath) {
		return "", fmt.Errorf("%s is not an absolute path", dirPath)
	}
	return getFilePathForDir(filepath.Clean(dirPath))
}

func (c *configProvider) Get(filePath string) (Config, error) {
	if !filepath.IsAbs(filePath) {
		return Config{}, fmt.Errorf("%s is not an absolute path", filePath)
	}
	filePath = filepath.Clean(filePath)
	return get(c.develMode, filePath)
}

func (c *configProvider) GetForData(dirPath string, externalConfigData string) (Config, error) {
	if !filepath.IsAbs(dirPath) {
		return Config{}, fmt.Errorf("%s is not an absolute path", dirPath)
	}
	dirPath = filepath.Clean(dirPath)
	var externalConfig ExternalConfig
	if err := jsonUnmarshalStrict([]byte(externalConfigData), &externalConfig); err != nil {
		return Config{}, err
	}
	return externalConfigToConfig(c.develMode, externalConfig, dirPath)
}

func (c *configProvider) GetExcludePrefixesForDir(dirPath string) ([]string, error) {
	if !filepath.IsAbs(dirPath) {
		return nil, fmt.Errorf("%s is not an absolute path", dirPath)
	}
	dirPath = filepath.Clean(dirPath)
	return getExcludePrefixesForDir(dirPath)
}

func (c *configProvider) GetExcludePrefixesForData(dirPath string, externalConfigData string) ([]string, error) {
	if !filepath.IsAbs(dirPath) {
		return nil, fmt.Errorf("%s is not an absolute path", dirPath)
	}
	dirPath = filepath.Clean(dirPath)
	var externalConfig ExternalConfig
	if err := jsonUnmarshalStrict([]byte(externalConfigData), &externalConfig); err != nil {
		return nil, err
	}
	return getExcludePrefixes(externalConfig.Excludes, dirPath)
}

// getFilePathForDir tries to find a file named by one of the ConfigFilenames starting in the
// given directory, and going up a directory until hitting root.
//
// The directory must be an absolute path.
//
// If no such file is found, "" is returned.
// If multiple files named by one of the ConfigFilenames are found in the same
// directory, error is returned.
func getFilePathForDir(dirPath string) (string, error) {
	for {
		filePath, err := getSingleFilePathForDir(dirPath)
		if err != nil {
			return "", err
		}
		if filePath != "" {
			return filePath, nil
		}
		if dirPath == "/" {
			return "", nil
		}
		dirPath = filepath.Dir(dirPath)
	}
}

// getSingleFilePathForDir gets the file named by one of the ConfigFilenames in the
// given directory. Having multiple such files results in an error being returned. If no file is
// found, this returns "".
func getSingleFilePathForDir(dirPath string) (string, error) {
	var filePaths []string
	for _, configFilename := range ConfigFilenames {
		filePath := filepath.Join(dirPath, configFilename)
		if _, err := os.Stat(filePath); err == nil {
			filePaths = append(filePaths, filePath)
		}
	}
	switch len(filePaths) {
	case 0:
		return "", nil
	case 1:
		return filePaths[0], nil
	default:
		return "", fmt.Errorf("multiple configuration files in the same directory: %v", filePaths)
	}
}

// get reads the config at the given path.
//
// This is expected to be in YAML or JSON format, which is denoted by the file extension.
func get(develMode bool, filePath string) (Config, error) {
	externalConfig, err := getExternalConfig(filePath)
	if err != nil {
		return Config{}, err
	}
	return externalConfigToConfig(develMode, externalConfig, filepath.Dir(filePath))
}

func getDefaultConfig(develMode bool, dirPath string) (Config, error) {
	return externalConfigToConfig(develMode, ExternalConfig{}, dirPath)
}

func getExternalConfig(filePath string) (ExternalConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ExternalConfig{}, err
	}
	if len(data) == 0 {
		return ExternalConfig{}, nil
	}
	externalConfig := ExternalConfig{}
	switch filepath.Ext(filePath) {
	case ".json":
		if err := jsonUnmarshalStrict(data, &externalConfig); err != nil {
			return ExternalConfig{}, err
		}
		return externalConfig, nil
	case ".yaml":
		if err := yaml.UnmarshalStrict(data, &externalConfig); err != nil {
			return ExternalConfig{}, err
		}
		return externalConfig, nil
	default:
		return ExternalConfig{}, fmt.Errorf("unknown config file extension, must be .json or .yaml: %s", filePath)
	}
}

// externalConfigToConfig converts an ExternalConfig to a Config.
//
// This will return a valid Config, or an error.
func externalConfigToConfig(develMode bool, e ExternalConfig, dirPath string) (Config, error) {
	excludePrefixes, err := getExcludePrefixes(e.Excludes, dirPath)
	if err != nil {
		return Config{}, err
	}
	includePaths := make([]string, 0, len(e.Protoc.Includes))
	for _, includePath := range strs.SortUniq(e.Protoc.Includes) {
		if !filepath.IsAbs(includePath) {
			includePath = filepath.Join(dirPath, includePath)
		}
		includePath = filepath.Clean(includePath)
		includePaths = append(includePaths, includePath)
	}
	ignoreIDToFilePaths := make(map[string][]string)
	for _, ignore := range e.Lint.Ignores {
		id := strings.ToUpper(ignore.ID)
		for _, protoFilePath := range ignore.Files {
			if !filepath.IsAbs(protoFilePath) {
				protoFilePath = filepath.Join(dirPath, protoFilePath)
			}
			protoFilePath = filepath.Clean(protoFilePath)
			if _, ok := ignoreIDToFilePaths[id]; !ok {
				ignoreIDToFilePaths[id] = make([]string, 0)
			}
			ignoreIDToFilePaths[id] = append(ignoreIDToFilePaths[id], protoFilePath)
		}
	}

	genPlugins := make([]GenPlugin, len(e.Generate.Plugins))
	for i, plugin := range e.Generate.Plugins {
		genPluginType, err := ParseGenPluginType(plugin.Type)
		if err != nil {
			return Config{}, err
		}
		if plugin.Output == "" {
			return Config{}, fmt.Errorf("output path required for plugin %s", plugin.Name)
		}
		var relPath, absPath string
		if filepath.IsAbs(plugin.Output) {
			absPath = filepath.Clean(plugin.Output)
			relPath, err = filepath.Rel(dirPath, absPath)
			if err != nil {
				return Config{}, fmt.Errorf("failed to resolve plugin %q output absolute path %q to a relative path with base %q: %v", plugin.Name, absPath, dirPath, err)
			}
		} else {
			relPath = plugin.Output
			absPath = filepath.Clean(filepath.Join(dirPath, relPath))
		}
		if plugin.FileSuffix != "" && plugin.FileSuffix[0] == '.' {
			return Config{}, fmt.Errorf("file_suffix begins with '.' but should not include the '.': %s", plugin.FileSuffix)
		}
		if plugin.Name != "descriptor_set" {
			if plugin.IncludeImports {
				return Config{}, fmt.Errorf("include_imports is only valid for the descriptor_set plugin but set on %q", plugin.Name)
			}
			if plugin.IncludeSourceInfo {
				return Config{}, fmt.Errorf("include_source_info is only valid for the descriptor_set plugin but set on %q", plugin.Name)
			}
		}
		genPlugins[i] = GenPlugin{
			Name:              plugin.Name,
			GetPath:           getPluginPathFunc(plugin.Path),
			Type:              genPluginType,
			Flags:             plugin.Flags,
			FileSuffix:        plugin.FileSuffix,
			IncludeImports:    plugin.IncludeImports,
			IncludeSourceInfo: plugin.IncludeSourceInfo,
			OutputPath: OutputPath{
				RelPath: relPath,
				AbsPath: absPath,
			},
		}
	}
	sort.Slice(genPlugins, func(i int, j int) bool { return genPlugins[i].Name < genPlugins[j].Name })

	createDirPathToBasePackage := make(map[string]string)
	for _, pkg := range e.Create.Packages {
		relDirPath := pkg.Directory
		basePackage := pkg.Name
		if relDirPath == "" {
			return Config{}, fmt.Errorf("directory for create package is empty")
		}
		if basePackage == "" {
			return Config{}, fmt.Errorf("name for create package is empty")
		}
		if filepath.IsAbs(relDirPath) {
			return Config{}, fmt.Errorf("directory for create package must be relative: %s", relDirPath)
		}
		createDirPathToBasePackage[filepath.Clean(filepath.Join(dirPath, relDirPath))] = basePackage
	}
	// to make testing easier
	if len(createDirPathToBasePackage) == 0 {
		createDirPathToBasePackage = nil
	}

	var fileHeader string
	if e.Lint.FileHeader.Path != "" || e.Lint.FileHeader.Content != "" {
		if e.Lint.FileHeader.Path != "" && e.Lint.FileHeader.Content != "" {
			return Config{}, fmt.Errorf("must only specify either file header path or content")
		}
		var fileHeaderContent string
		if e.Lint.FileHeader.Path != "" {
			if filepath.IsAbs(e.Lint.FileHeader.Path) {
				return Config{}, fmt.Errorf("path for file header must be relative: %s", e.Lint.FileHeader.Path)
			}
			fileHeaderData, err := ioutil.ReadFile(filepath.Join(dirPath, e.Lint.FileHeader.Path))
			if err != nil {
				return Config{}, err
			}
			fileHeaderContent = string(fileHeaderData)
		} else { // if e.Lint.FileHeader.Content != ""
			fileHeaderContent = e.Lint.FileHeader.Content
		}
		fileHeaderLines := getFileHeaderLines(fileHeaderContent)
		if !e.Lint.FileHeader.IsCommented {
			for i, fileHeaderLine := range fileHeaderLines {
				if fileHeaderLine == "" {
					fileHeaderLines[i] = "//"
				} else {
					fileHeaderLines[i] = "// " + fileHeaderLine
				}
			}
		}
		fileHeader = strings.Join(fileHeaderLines, "\n")
		if fileHeader == "" {
			return Config{}, fmt.Errorf("file header path or content specified but result was empty file header")
		}
	}

	if !develMode {
		if e.Lint.AllowSuppression {
			return Config{}, fmt.Errorf("allow_suppression is not allowed outside of internal prototool tests")
		}
	}

	config := Config{
		DirPath:         dirPath,
		ExcludePrefixes: excludePrefixes,
		Compile: CompileConfig{
			ProtobufVersion:       e.Protoc.Version,
			IncludePaths:          includePaths,
			IncludeWellKnownTypes: true, // Always include the well-known types.
			AllowUnusedImports:    e.Protoc.AllowUnusedImports,
		},
		Create: CreateConfig{
			DirPathToBasePackage: createDirPathToBasePackage,
		},
		Lint: LintConfig{
			IncludeIDs:          strs.SortUniqModify(e.Lint.Rules.Add, strings.ToUpper),
			ExcludeIDs:          strs.SortUniqModify(e.Lint.Rules.Remove, strings.ToUpper),
			Group:               strings.ToLower(e.Lint.Group),
			NoDefault:           e.Lint.Rules.NoDefault,
			IgnoreIDToFilePaths: ignoreIDToFilePaths,
			FileHeader:          fileHeader,
			JavaPackagePrefix:   e.Lint.JavaPackagePrefix,
			AllowSuppression:    e.Lint.AllowSuppression,
		},
		Break: BreakConfig{
			IncludeBeta:   e.Break.IncludeBeta,
			AllowBetaDeps: e.Break.AllowBetaDeps,
		},
		Gen: GenConfig{
			GoPluginOptions: GenGoPluginOptions{
				ImportPath:     e.Generate.GoOptions.ImportPath,
				ExtraModifiers: e.Generate.GoOptions.ExtraModifiers,
			},
			Plugins: genPlugins,
		},
	}

	for _, genPlugin := range config.Gen.Plugins {
		// TODO: technically protoc-gen-protoc-gen-foo is a valid
		// plugin binary with name protoc-gen-foo, but do we want
		// to error if protoc-gen- is a prefix of a name?
		// I think this will be a common enough mistake that we
		// can remove this later. Or, do we want names to include
		// the protoc-gen- part?
		if strings.HasPrefix(genPlugin.Name, "protoc-gen-") {
			return Config{}, fmt.Errorf("plugin name provided was %s, do not include the protoc-gen- prefix", genPlugin.Name)
		}
		if _, ok := _genPluginTypeToString[genPlugin.Type]; !ok {
			return Config{}, fmt.Errorf("unknown GenPluginType: %v", genPlugin.Type)
		}
		if (genPlugin.Type.IsGo() || genPlugin.Type.IsGogo()) && config.Gen.GoPluginOptions.ImportPath == "" {
			return Config{}, fmt.Errorf("go plugin %s specified but no import path provided", genPlugin.Name)
		}
	}

	if intersection := strs.Intersection(config.Lint.IncludeIDs, config.Lint.ExcludeIDs); len(intersection) > 0 {
		return Config{}, fmt.Errorf("config had intersection of %v between lint_include and lint_exclude", intersection)
	}
	return config, nil
}

func getExcludePrefixesForDir(dirPath string) ([]string, error) {
	filePath, err := getSingleFilePathForDir(dirPath)
	if err != nil {
		return nil, err
	}
	if filePath == "" {
		return []string{}, nil
	}
	externalConfig, err := getExternalConfig(filePath)
	if err != nil {
		return nil, err
	}
	return getExcludePrefixes(externalConfig.Excludes, dirPath)
}

func getExcludePrefixes(excludes []string, dirPath string) ([]string, error) {
	excludePrefixes := make([]string, 0, len(excludes))
	for _, excludePrefix := range strs.SortUniq(excludes) {
		if !filepath.IsAbs(excludePrefix) {
			excludePrefix = filepath.Join(dirPath, excludePrefix)
		}
		excludePrefix = filepath.Clean(excludePrefix)
		if excludePrefix == dirPath {
			return nil, fmt.Errorf("cannot exclude directory of config file: %s", dirPath)
		}
		if !strings.HasPrefix(excludePrefix, dirPath) {
			return nil, fmt.Errorf("cannot exclude directory outside of config file directory %s: %s", dirPath, excludePrefix)
		}
		excludePrefixes = append(excludePrefixes, excludePrefix)
	}
	return excludePrefixes, nil
}

// jsonUnmarshalStrict makes sure there are no unknown fields when unmarshalling.
// This matches what yaml.UnmarshalStrict does basically.
// json.Unmarshal allows unknown fields.
func jsonUnmarshalStrict(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func getFileHeaderLines(content string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		lines = append(lines, strings.TrimSpace(line))
	}
	return lines
}

func getPluginPathFunc(path string) func() (string, error) {
	return func() (string, error) {
		if path == "" {
			return "", nil
		}
		return exec.LookPath(path)
	}
}
