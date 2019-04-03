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

// Package exec brings all the functionality of Prototool together in a format
// easily consumable by CLI libraries. It is effectively the glue between
// internal/cmd and all other packages.
package exec

import (
	"io"

	"go.uber.org/zap"
)

// ExitError is an error that signals to exit with a certain code.
type ExitError struct {
	Code    int
	Message string
}

// Error implements error.
func (e *ExitError) Error() string {
	return e.Message
}

// Runner runs commands.
//
// The args given are the args from the command line.
// Each additional parameter generally refers to a command-specific flag.
type Runner interface {
	Init(args []string, uncomment bool, document bool) error
	Create(args []string, pkg string) error
	Version() error
	CacheUpdate(args []string) error
	CacheDelete() error
	Files(args []string) error
	Compile(args []string, dryRun bool) error
	Gen(args []string, dryRun bool) error
	Lint(args []string, listAllLinters bool, listLinters bool, listAllLintGroups bool, listLintGroup string, diffLintGroups string, generateIgnores bool) error
	Format(args []string, overwrite, diffMode, lintMode, fix bool) error
	All(args []string, disableFormat, disableLint, fix bool) error
	GRPC(args, headers []string, address, method, data, callTimeout, connectTimeout, keepaliveTime string, stdin bool, details bool, tls bool, insecure bool, cacert string, cert string, key string, serverName string) error
	InspectPackages(args []string) error
	InspectPackageDeps(args []string, name string) error
	InspectPackageImporters(args []string, name string) error
	BreakCheck(args []string, gitBranch string, descriptorSetPath string) error
	BreakDescriptorSet(args []string, outputPath string) error
	DescriptorSet(args []string, includeImports bool, includeSourceInfo bool, outputPath string, tmp bool) error
}

// RunnerOption is an option for a new Runner.
type RunnerOption func(*runner)

// RunnerWithLogger returns a RunnerOption that uses the given logger.
//
// The default is to use zap.NewNop().
func RunnerWithLogger(logger *zap.Logger) RunnerOption {
	return func(runner *runner) {
		runner.logger = logger
	}
}

// RunnerWithDevelMode returns a RunnerOption that allows devel-mode.
func RunnerWithDevelMode() RunnerOption {
	return func(runner *runner) {
		runner.develMode = true
	}
}

// RunnerWithCachePath returns a RunnerOption that uses the given cache path.
func RunnerWithCachePath(cachePath string) RunnerOption {
	return func(runner *runner) {
		runner.cachePath = cachePath
	}
}

// RunnerWithConfigData returns a RunnerOption that uses the given config path.
func RunnerWithConfigData(configData string) RunnerOption {
	return func(runner *runner) {
		runner.configData = configData
	}
}

// RunnerWithJSON returns a RunnerOption that will print failures as JSON.
func RunnerWithJSON() RunnerOption {
	return func(runner *runner) {
		runner.json = true
	}
}

// RunnerWithErrorFormat returns a RunnerOption that uses the given colon-separated
// error format. The default is filename:line:column:message.
func RunnerWithErrorFormat(errorFormat string) RunnerOption {
	return func(runner *runner) {
		runner.errorFormat = errorFormat
	}
}

// RunnerWithProtocBinPath returns a RunnerOption that uses the given protoc binary path.
func RunnerWithProtocBinPath(protocBinPath string) RunnerOption {
	return func(runner *runner) {
		runner.protocBinPath = protocBinPath
	}
}

// RunnerWithProtocWKTPath returns a RunnerOption that uses the given path to include the well-known types.
func RunnerWithProtocWKTPath(protocWKTPath string) RunnerOption {
	return func(runner *runner) {
		runner.protocWKTPath = protocWKTPath
	}
}

// RunnerWithProtocURL returns a RunnerOption that uses the given protoc zip file URL.
func RunnerWithProtocURL(protocURL string) RunnerOption {
	return func(runner *runner) {
		runner.protocURL = protocURL
	}
}

// NewRunner returns a new Runner.
//
// workDirPath should generally be the current directory.
// input and output generally refer to stdin and stdout.
func NewRunner(workDirPath string, input io.Reader, output io.Writer, options ...RunnerOption) Runner {
	return newRunner(workDirPath, input, output, options...)
}
