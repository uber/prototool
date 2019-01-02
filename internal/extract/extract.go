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

// Package extract is used to extract elements from reflect PackageSets.
package extract

import (
	"fmt"

	reflectpb "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect"
)

// PackageSet is the Golang wrapper for the Protobuf PackageSet object.
type PackageSet struct {
	protoMessage *reflectpb.PackageSet

	packageNameToPackage map[string]*Package
}

// ProtoMessage returns the underlying Protobuf messge.
func (p *PackageSet) ProtoMessage() *reflectpb.PackageSet {
	return p.protoMessage
}

// PackageNameToPackage returns a map from package name to Package.
func (p *PackageSet) PackageNameToPackage() map[string]*Package {
	return p.packageNameToPackage
}

// Package is the Golang wrapper for the Protobuf Package object.
type Package struct {
	protoMessage *reflectpb.Package

	packageSet                 *PackageSet
	dependencyNameToDependency map[string]*Package
	importerNameToImporter     map[string]*Package
}

// ProtoMessage returns the underlying Protobuf messge.
func (p *Package) ProtoMessage() *reflectpb.Package {
	return p.protoMessage
}

// PackageSet returns the parent PackageSet.
func (p *Package) PackageSet() *PackageSet {
	return p.packageSet
}

// DependencyNameToDependency returns the direct dependencies of the given Package.
func (p *Package) DependencyNameToDependency() map[string]*Package {
	return p.dependencyNameToDependency
}

// ImporterNameToImporter returns the direct importers of the given Package.
func (p *Package) ImporterNameToImporter() map[string]*Package {
	return p.importerNameToImporter
}

// NewPackageSet returns a new PackageSet for the given reflect PackageSet.
func NewPackageSet(protoMessage *reflectpb.PackageSet) (*PackageSet, error) {
	packageSet := &PackageSet{
		protoMessage:         protoMessage,
		packageNameToPackage: make(map[string]*Package),
	}
	for _, pkg := range packageSet.protoMessage.Packages {
		packageSet.packageNameToPackage[pkg.Name] = &Package{
			protoMessage:               pkg,
			packageSet:                 packageSet,
			dependencyNameToDependency: make(map[string]*Package),
			importerNameToImporter:     make(map[string]*Package),
		}
	}
	for packageName, pkg := range packageSet.packageNameToPackage {
		for _, dependencyName := range pkg.protoMessage.DependencyNames {
			dependency, ok := packageSet.packageNameToPackage[dependencyName]
			if !ok {
				return nil, fmt.Errorf("no package for name %s", dependencyName)
			}
			pkg.dependencyNameToDependency[dependencyName] = dependency
			dependency.importerNameToImporter[packageName] = pkg
		}
	}
	return packageSet, nil
}
