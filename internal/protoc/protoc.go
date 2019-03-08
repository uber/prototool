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

package protoc

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"go.uber.org/zap"
)

// Downloader downloads and caches protobuf.
type Downloader interface {
	// Download protobuf.
	//
	// If already downloaded, this has no effect. This is thread-safe.
	// This will download to ${XDG_CACHE_HOME}/prototool/$(uname -s)/$(uname -m)
	// unless overridden by a DownloaderOption.
	// If ${XDG_CACHE_HOME} is not set, it defaults to ${HOME}/Library/Caches on
	// Darwin, and ${HOME}/.cache on Linux.
	// If ${HOME} is not set, an error will be returned.
	//
	// Returns the path to the downloaded protobuf artifacts.
	//
	// ProtocPath and WellKnownTypesIncludePath implicitly call this.
	Download() (string, error)

	// Get the path to protoc.
	//
	// If not downloaded, this downloads and caches protobuf. This is thread-safe.
	ProtocPath() (string, error)

	// Get the path to include for the well-known types.
	//
	// Inside this directory will be the subdirectories google/protobuf.
	//
	// If not downloaded, this downloads and caches protobuf. This is thread-safe.
	WellKnownTypesIncludePath() (string, error)

	// Delete any downloaded artifacts.
	//
	// This is not thread-safe and no calls to other functions can be reliably
	// made simultaneously.
	Delete() error
}

// DownloaderOption is an option for a new Downloader.
type DownloaderOption func(*downloader)

// DownloaderWithLogger returns a DownloaderOption that uses the given logger.
//
// The default is to use zap.NewNop().
func DownloaderWithLogger(logger *zap.Logger) DownloaderOption {
	return func(downloader *downloader) {
		downloader.logger = logger
	}
}

// DownloaderWithCachePath returns a DownloaderOption that uses the given cachePath.
//
// The default is ${XDG_CACHE_HOME}/prototool/$(uname -s)/$(uname -m).
func DownloaderWithCachePath(cachePath string) DownloaderOption {
	return func(downloader *downloader) {
		downloader.cachePath = cachePath
	}
}

// DownloaderWithProtocBinPath returns a DownloaderOption that uses the given protoc binary path.
func DownloaderWithProtocBinPath(protocBinPath string) DownloaderOption {
	return func(downloader *downloader) {
		downloader.protocBinPath = protocBinPath
	}
}

// DownloaderWithProtocWKTPath returns a DownloaderOption that uses the given path to include
// the well-known types.
func DownloaderWithProtocWKTPath(protocWKTPath string) DownloaderOption {
	return func(downloader *downloader) {
		downloader.protocWKTPath = protocWKTPath
	}
}

// DownloaderWithProtocURL returns a DownloaderOption that uses the given protoc zip file URL.
//
// The default is https://github.com/protocolbuffers/protobuf/releases/download/vVERSION/protoc-VERSION-OS-ARCH.zip.
func DownloaderWithProtocURL(protocURL string) DownloaderOption {
	return func(downloader *downloader) {
		downloader.protocURL = protocURL
	}
}

// NewDownloader returns a new Downloader for the given config and DownloaderOptions.
func NewDownloader(config settings.Config, options ...DownloaderOption) (Downloader, error) {
	return newDownloader(config, options...)
}

// FileDescriptorSet is a wrapper for descriptor.FileDescriptorSet.
//
// This will contain both the files specified by ProtoFiles and all imports.
type FileDescriptorSet struct {
	*descriptor.FileDescriptorSet
	// The containing ProtoSet.
	ProtoSet *file.ProtoSet
	// The absolute directory path for the built files in this FileDescriptorSet.
	// This directory path will always reside within the ProtoSetDirPath,
	// that is filepath.Rel(ProtoSetDirPath, DirPath) will never return
	// error and always return a non-empty string. Note the string could be ".".
	DirPath string
	// The ProtoFiles for the built files in this FileDescriptorSet.
	// The directory of Path will always be equal to DirPath.
	ProtoFiles []*file.ProtoFile
}

// FileDescriptorSets are a slice of FileDescriptorSet objects.
type FileDescriptorSets []*FileDescriptorSet

// Unwrap converts f to []*descriptor.FileDescriptorSet.
//
// Used for backwards compatibility with existing code that is based on
// descriptor.FileDescriptorSets.
func (f FileDescriptorSets) Unwrap() []*descriptor.FileDescriptorSet {
	if f == nil {
		return nil
	}
	d := make([]*descriptor.FileDescriptorSet, len(f))
	for i, e := range f {
		d[i] = e.FileDescriptorSet
	}
	return d
}

// CompileResult is the result of a compile
type CompileResult struct {
	// The failures from all calls.
	Failures []*text.Failure
	// Will not be set if there are any failures.
	//
	// Will only be set if the CompilerWithFileDescriptorSet
	// option is used.
	FileDescriptorSets FileDescriptorSets
}

// Compiler compiles protobuf files.
type Compiler interface {
	// Compile the protobuf files with protoc.
	//
	// If there are compile failures, they will be returned in the slice
	// and there will be no error. The caller can determine if this is
	// an error case. If there is any other type of error, or some output
	// from protoc cannot be interpreted, an error will be returned.
	Compile(*file.ProtoSet) (*CompileResult, error)

	// Return the protoc commands that would be run on Compile.
	//
	// This will ignore the CompilerWithFileDescriptorSet option.
	ProtocCommands(*file.ProtoSet) ([]string, error)
}

// CompilerOption is an option for a new Compiler.
type CompilerOption func(*compiler)

// CompilerWithLogger returns a CompilerOption that uses the given logger.
//
// The default is to use zap.NewNop().
func CompilerWithLogger(logger *zap.Logger) CompilerOption {
	return func(compiler *compiler) {
		compiler.logger = logger
	}
}

// CompilerWithCachePath returns a CompilerOption that uses the given cachePath.
//
// The default is ${XDG_CACHE_HOME}/prototool/$(uname -s)/$(uname -m).
func CompilerWithCachePath(cachePath string) CompilerOption {
	return func(compiler *compiler) {
		compiler.cachePath = cachePath
	}
}

// CompilerWithProtocBinPath returns a CompilerOption that uses the given protoc binary path.
//
func CompilerWithProtocBinPath(protocBinPath string) CompilerOption {
	return func(compiler *compiler) {
		compiler.protocBinPath = protocBinPath
	}
}

// CompilerWithProtocWKTPath returns a CompilerOption that uses the given path to include the
// well-known types.
func CompilerWithProtocWKTPath(protocWKTPath string) CompilerOption {
	return func(compiler *compiler) {
		compiler.protocWKTPath = protocWKTPath
	}
}

// CompilerWithProtocURL returns a CompilerOption that uses the given protoc zip file URL.
//
// The default is https://github.com/protocolbuffers/protobuf/releases/download/vVERSION/protoc-VERSION-OS-ARCH.zip.
func CompilerWithProtocURL(protocURL string) CompilerOption {
	return func(compiler *compiler) {
		compiler.protocURL = protocURL
	}
}

// CompilerWithGen says to also generate the code.
func CompilerWithGen() CompilerOption {
	return func(compiler *compiler) {
		compiler.doGen = true
	}
}

// CompilerWithFileDescriptorSet says to also return the FileDescriptorSet.
func CompilerWithFileDescriptorSet() CompilerOption {
	return func(compiler *compiler) {
		compiler.doFileDescriptorSet = true
	}
}

// CompilerWithFileDescriptorSetFullControl says to also return the FileDescriptorSet but
// with extra controls.
//
// This is added for backwards compatibility within the codebase.
func CompilerWithFileDescriptorSetFullControl(includeImports bool, includeSourceInfo bool) CompilerOption {
	return func(compiler *compiler) {
		compiler.doFileDescriptorSet = true
		compiler.fileDescriptorSetFullControl = true
		compiler.fileDescriptorSetIncludeImports = includeImports
		compiler.fileDescriptorSetIncludeSourceInfo = includeSourceInfo
	}
}

// NewCompiler returns a new Compiler.
func NewCompiler(options ...CompilerOption) Compiler {
	return newCompiler(options...)
}
