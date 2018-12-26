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

// Package extract is used to extract elements from FileDescriptorSets created
// from internal/protoc, for use in json-to-binary and back conversion, and for
// use for gRPC.
package extract

import (
	"sort"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.uber.org/zap"
)

// Packages is a set of extracted packages.
type Packages struct {
	// Map from fully-qualified name to package.
	// Fully-qualified name includes prefix '.'.
	NameToPackage map[string]*Package
}

// SortedPackages returns the list of packages sorted by name.
func (p *Packages) SortedPackages() []*Package {
	packages := make([]*Package, 0, len(p.NameToPackage))
	for _, pkg := range p.NameToPackage {
		packages = append(packages, pkg)
	}
	sort.Stable(sortPackages(packages))
	return packages
}

// Package is an extracted package.
type Package struct {
	// Fully-qualified name includes prefix '.'.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// The fully-qualified names of the direct dependencies.
	// For recursive dependencies, look these names up in the Packages struct.
	Deps []string `json:"deps,omitempty" yaml:"deps,omitempty"`
	// The fully-qualified names of the importers.
	// For recursive importers, look these names up in the Packages struct.
	Importers []string `json:"importers,omitempty" yaml:"importers,omitempty"`
}

// Field is an extracted field.
type Field struct {
	*descriptor.FieldDescriptorProto

	// Fully-qualified path includes prefix '.'.
	FullyQualifiedPath  string
	DescriptorProto     *descriptor.DescriptorProto
	FileDescriptorProto *descriptor.FileDescriptorProto
	FileDescriptorSet   *descriptor.FileDescriptorSet
}

// Message is an extracted message.
type Message struct {
	*descriptor.DescriptorProto

	// Fully-qualified path includes prefix '.'.
	FullyQualifiedPath  string
	FileDescriptorProto *descriptor.FileDescriptorProto
	FileDescriptorSet   *descriptor.FileDescriptorSet
}

// Service is an extracted service.
type Service struct {
	*descriptor.ServiceDescriptorProto

	// Fully-qualified path includes prefix '.'.
	FullyQualifiedPath  string
	FileDescriptorProto *descriptor.FileDescriptorProto
	FileDescriptorSet   *descriptor.FileDescriptorSet
}

// Getter extracts elements.
//
// Paths can begin with ".".
// The first FileDescriptorSet with a match will be returned.
type Getter interface {
	// Get the packages.
	GetPackages(fileDescriptorSets []*descriptor.FileDescriptorSet) (*Packages, error)
	// Get the field that matches the path.
	// Return non-nil value, or error otherwise including if nothing found.
	GetField(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Field, error)
	// Get the message that matches the path.
	// Return non-nil value, or error otherwise including if nothing found.
	GetMessage(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Message, error)
	// Get the service that matches the path.
	// Return non-nil value, or error otherwise including if nothing found.
	GetService(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Service, error)
}

// GetterOption is an option for a new Getter.
type GetterOption func(*getter)

// GetterWithLogger returns a GetterOption that uses the given logger.
//
// The default is to use zap.NewNop().
func GetterWithLogger(logger *zap.Logger) GetterOption {
	return func(getter *getter) {
		getter.logger = logger
	}
}

// NewGetter returns a new Getter.
func NewGetter(options ...GetterOption) Getter {
	return newGetter(options...)
}

type sortPackages []*Package

func (s sortPackages) Len() int          { return len(s) }
func (s sortPackages) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s sortPackages) Less(i int, j int) bool {
	if s[i] == nil && s[j] == nil {
	}
	if s[i] == nil && s[j] != nil {
		return true
	}
	if s[i] != nil && s[j] == nil {
		return false
	}
	return s[i].Name < s[j].Name
}
