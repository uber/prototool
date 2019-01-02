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

package reflect

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	reflectpb "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect"
	"github.com/uber/prototool/internal/strs"
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

	packageSet *PackageSet
}

// ProtoMessage returns the underlying Protobuf messge.
func (p *Package) ProtoMessage() *reflectpb.Package {
	return p.protoMessage
}

// PackageSet returns the parent PackageSet.
func (p *Package) PackageSet() *PackageSet {
	return p.packageSet
}

// DependencyNames returns the sorted list of dependencies.
func (p *Package) DependencyNames() []string {
	return nil
}

// NewPackageSet returns a new valid PackageSet for the given
// FileDescriptorSets.
//
// The FileDescriptorSets can have FileDescriptorProtos with the same name, but
// they must be equal.
func NewPackageSet(fileDescriptorSets ...*descriptor.FileDescriptorSet) (*PackageSet, error) {
	protoMessage, err := newPackageSetProtoMessage(fileDescriptorSets...)
	if err != nil {
		return nil, err
	}
	return newPackageSetFromProtoMessage(protoMessage), nil
}

func newPackageSetProtoMessage(fileDescriptorSets ...*descriptor.FileDescriptorSet) (*reflectpb.PackageSet, error) {
	packageNameToFileNameToFileDescriptorProto, err := getPackageNameToFileNameToFileDescriptorProto(fileDescriptorSets)
	if err != nil {
		return nil, err
	}
	fileNameToPackageName, err := getFileNameToPackageName(packageNameToFileNameToFileDescriptorProto)
	if err != nil {
		return nil, err
	}
	packageNameToPackage := make(map[string]*reflectpb.Package)
	for packageName := range packageNameToFileNameToFileDescriptorProto {
		packageNameToPackage[packageName] = &reflectpb.Package{
			Name: packageName,
		}
	}
	for packageName, fileNameTofileDescriptorProto := range packageNameToFileNameToFileDescriptorProto {
		for _, fileDescriptorProto := range fileNameTofileDescriptorProto {
			for _, depFileName := range fileDescriptorProto.GetDependency() {
				depPackageName, ok := fileNameToPackageName[depFileName]
				if !ok {
					return nil, fmt.Errorf("no package for dep %s", depFileName)
				}
				if depPackageName != packageName {
					packageNameToPackage[packageName].DependencyNames = append(packageNameToPackage[packageName].DependencyNames, depPackageName)
				}
			}
		}
	}
	for _, pkg := range packageNameToPackage {
		pkg.DependencyNames = strs.DedupeSort(pkg.DependencyNames, nil)
	}
	packageSet := &reflectpb.PackageSet{
		Packages: make([]*reflectpb.Package, 0, len(packageNameToPackage)),
	}
	for _, pkg := range packageNameToPackage {
		packageSet.Packages = append(packageSet.Packages, pkg)
	}
	sort.Sort(sortPackages(packageSet.Packages))
	return packageSet, nil
}

func newPackageSetFromProtoMessage(protoMessage *reflectpb.PackageSet) *PackageSet {
	packageSet := &PackageSet{
		protoMessage:         protoMessage,
		packageNameToPackage: make(map[string]*Package),
	}
	for _, pkg := range packageSet.protoMessage.Packages {
		packageSet.packageNameToPackage[pkg.Name] = newPackageFromProtoMessage(pkg, packageSet)
	}
	return packageSet
}

func newPackageFromProtoMessage(protoMessage *reflectpb.Package, packageSet *PackageSet) *Package {
	return &Package{
		protoMessage: protoMessage,
		packageSet:   packageSet,
	}
}

func getPackageNameToFileNameToFileDescriptorProto(fileDescriptorSets []*descriptor.FileDescriptorSet) (map[string]map[string]*descriptor.FileDescriptorProto, error) {
	packageNameToFileNameToFileDescriptorProto := make(map[string]map[string]*descriptor.FileDescriptorProto)
	for _, fileDescriptorSet := range fileDescriptorSets {
		for _, fileDescriptorProto := range fileDescriptorSet.GetFile() {
			pkg := fileDescriptorProto.GetPackage()
			if pkg == "" {
				return nil, fmt.Errorf("no package on FileDescriptorProto")
			}
			if pkg[0] == '.' {
				return nil, fmt.Errorf("malformed package fully-qualified name %s on FileDescriptorProto %+v", pkg, fileDescriptorProto)
			}
			existingMap, ok := packageNameToFileNameToFileDescriptorProto[pkg]
			if !ok {
				existingMap = make(map[string]*descriptor.FileDescriptorProto)
				packageNameToFileNameToFileDescriptorProto[pkg] = existingMap
			}
			name := fileDescriptorProto.GetName()
			existing, ok := existingMap[name]
			if ok {
				// we don't technically need to do this verification but
				// we do for safety, if this ends up never erroring we
				// can remove this if it becomes a performance issue
				data, err := proto.Marshal(fileDescriptorProto)
				if err != nil {
					return nil, err
				}
				existingData, err := proto.Marshal(existing)
				if err != nil {
					return nil, err
				}
				if !bytes.Equal(data, existingData) {
					return nil, fmt.Errorf("unequal FileDescriptorProtos for %s", name)
				}
			} else {
				existingMap[name] = fileDescriptorProto
			}
		}
	}
	return packageNameToFileNameToFileDescriptorProto, nil
}

func getFileNameToPackageName(packageNameToFileNameToFileDescriptorProto map[string]map[string]*descriptor.FileDescriptorProto) (map[string]string, error) {
	fileNameToPackageName := make(map[string]string)
	for packageName, fileNameToFileDescriptorProto := range packageNameToFileNameToFileDescriptorProto {
		for fileName := range fileNameToFileDescriptorProto {
			existing, ok := fileNameToPackageName[fileName]
			if ok {
				if existing != packageName {
					return nil, fmt.Errorf("mismatched packages names %s and %s for file %s", existing, packageName, fileName)
				}
			} else {
				fileNameToPackageName[fileName] = packageName
			}
		}
	}
	return fileNameToPackageName, nil
}

type sortPackages []*reflectpb.Package

func (s sortPackages) Len() int          { return len(s) }
func (s sortPackages) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s sortPackages) Less(i int, j int) bool {
	if s[i] == nil && s[j] == nil {
		return false
	}
	if s[i] == nil && s[j] != nil {
		return true
	}
	if s[i] != nil && s[j] == nil {
		return false
	}
	return s[i].Name < s[j].Name
}
