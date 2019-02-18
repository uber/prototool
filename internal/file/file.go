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

package file

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/uber/prototool/internal/settings"
	"go.uber.org/zap"
)

// DefaultWalkTimeout is the default walk timeout.
const DefaultWalkTimeout time.Duration = 3 * time.Second

var rootDirPath = filepath.Dir(string(filepath.Separator))

// ProtoSet represents a set of .proto files and an associated config.
//
// ProtoSets will be validated if returned from this package.
type ProtoSet struct {
	// The working directory path.
	// Must be absolute.
	// Must be cleaned.
	WorkDirPath string
	// The given directory path.
	// Must be absolute.
	// Must be cleaned.
	DirPath string
	// The directory path to slice of .proto files.
	// All paths must be absolute.
	// All paths must reside within DirPath.
	// Must be cleaned.
	// The directory paths will always reside within the config DirPath,
	// that is filepath.Rel(Config.DirPath, DirPath) will never return
	// error and always return a non-empty string. Note the string could be ".".
	// The ProtoFiles will always be in the directory specified by the key.
	DirPathToFiles map[string][]*ProtoFile
	// The associated Config.
	// Must be valid.
	// The DirPath on the config may differ from the DirPath on the ProtoSet.
	Config settings.Config
}

// ProtoFile represents a .proto file.
type ProtoFile struct {
	// The path to the .proto file.
	// Must be absolute.
	// Must be cleaned.
	Path string
	// The path to display in output.
	// This will be relative to the working directory, or the absolute path
	// if the file was outside the working directory.
	DisplayPath string
}

// ProtoSetProvider provides ProtoSets.
type ProtoSetProvider interface {
	// GetForDir gets the ProtoSet for the given dirPath.
	// The ProtoSet will have the config assocated with all files associated with
	// the ProtoSet.
	//
	// This will return all .proto files in the directory of the associated config file
	// and all it's subdirectories, or the given directory and its subdirectories
	// if there is no config file.
	//
	// Config will be searched for starting at the directory of each .proto file
	// and going up a directory until hitting root.
	// Returns an error if there is not exactly one ProtoSet.
	GetForDir(workDirPath string, dirPath string) (*ProtoSet, error)
}

// ProtoSetProviderOption is an option for a new ProtoSetProvider.
type ProtoSetProviderOption func(*protoSetProvider)

// ProtoSetProviderWithLogger returns a ProtoSetProviderOption that uses the given logger.
//
// The default is to use zap.NewNop().
func ProtoSetProviderWithLogger(logger *zap.Logger) ProtoSetProviderOption {
	return func(protoSetProvider *protoSetProvider) {
		protoSetProvider.logger = logger
	}
}

// ProtoSetProviderWithDevelMode returns a ProtoSetProviderOption that allows devel-mode.
func ProtoSetProviderWithDevelMode() ProtoSetProviderOption {
	return func(protoSetProvider *protoSetProvider) {
		protoSetProvider.develMode = true
	}
}

// ProtoSetProviderWithConfigData returns a ProtoSetProviderOption that uses the given configuration
// data instead of using configuration files that are found. This acts as if there is only one
// configuration file at the current working directory. All found configuration files are ignored.
func ProtoSetProviderWithConfigData(configData string) ProtoSetProviderOption {
	return func(protoSetProvider *protoSetProvider) {
		protoSetProvider.configData = configData
	}
}

// ProtoSetProviderWithWalkTimeout returns a ProtoSetProviderOption will timeout after walking
// a directory structure when searching for Protobuf files after the given amount of time.
//
// The default is to timeout after DefaultTimeoutDuration.
// Set to 0 for no timeout.
func ProtoSetProviderWithWalkTimeout(walkTimeout time.Duration) ProtoSetProviderOption {
	return func(protoSetProvider *protoSetProvider) {
		protoSetProvider.walkTimeout = walkTimeout
	}
}

// NewProtoSetProvider returns a new ProtoSetProvider.
func NewProtoSetProvider(options ...ProtoSetProviderOption) ProtoSetProvider {
	return newProtoSetProvider(options...)
}

// AbsClean returns the cleaned absolute path of the given path.
func AbsClean(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if !filepath.IsAbs(path) {
		return filepath.Abs(path)
	}
	return filepath.Clean(path), nil
}

// CheckAbs is a convenience functions for determining
// whether a path is an absolute path.
func CheckAbs(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("expected absolute path but was %s", path)
	}
	return nil
}

// IsExcluded determines whether the given filePath should be excluded.
//
// absConfigDirPath represents the absolute path to the configuration file.
// This is used to determine when we should stop checking for excludes.
func IsExcluded(absFilePath string, absConfigDirPath string, absExcludePaths ...string) bool {
	for _, absExcludePath := range absExcludePaths {
		for curFilePath := absFilePath; curFilePath != absConfigDirPath && curFilePath != rootDirPath; curFilePath = filepath.Dir(curFilePath) {
			if curFilePath == absExcludePath {
				return true
			}
		}
	}
	return false

}
