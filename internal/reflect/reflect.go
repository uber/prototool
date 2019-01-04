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
	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
	"github.com/uber/prototool/internal/strs"
)

// NewPackageSet returns a new valid PackageSet for the given
// FileDescriptorSets.
//
// The FileDescriptorSets can have FileDescriptorProtos with the same name, but
// they must be equal.
func NewPackageSet(fileDescriptorSets ...*descriptor.FileDescriptorSet) (*reflectv1.PackageSet, error) {
	packageNameToFileNameToFileDescriptorProto, err := getPackageNameToFileNameToFileDescriptorProto(fileDescriptorSets)
	if err != nil {
		return nil, err
	}
	packageNameToPackage, err := getBasePackageNameToPackage(packageNameToFileNameToFileDescriptorProto)
	if err != nil {
		return nil, err
	}
	if err := populateDependencies(packageNameToPackage, packageNameToFileNameToFileDescriptorProto); err != nil {
		return nil, err
	}
	return getPackageSet(packageNameToPackage)
}

// helper for NewPackageSet
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

// helper for NewPackageSet
func getBasePackageNameToPackage(
	packageNameToFileNameToFileDescriptorProto map[string]map[string]*descriptor.FileDescriptorProto,
) (map[string]*reflectv1.Package, error) {
	packageNameToPackage := make(map[string]*reflectv1.Package)
	for packageName := range packageNameToFileNameToFileDescriptorProto {
		packageNameToPackage[packageName] = &reflectv1.Package{
			Name: packageName,
		}
	}
	return packageNameToPackage, nil
}

// helper for NewPackageSet
func populateDependencies(
	packageNameToPackage map[string]*reflectv1.Package,
	packageNameToFileNameToFileDescriptorProto map[string]map[string]*descriptor.FileDescriptorProto,
) error {
	fileNameToPackageName, err := getFileNameToPackageName(packageNameToFileNameToFileDescriptorProto)
	if err != nil {
		return err
	}
	for packageName, fileNameTofileDescriptorProto := range packageNameToFileNameToFileDescriptorProto {
		for _, fileDescriptorProto := range fileNameTofileDescriptorProto {
			for _, depFileName := range fileDescriptorProto.GetDependency() {
				depPackageName, ok := fileNameToPackageName[depFileName]
				if !ok {
					return fmt.Errorf("no package for dep %s", depFileName)
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
	return nil
}

// helper for NewPackageSet
func populateEnums(
	packageNameToPackage map[string]*reflectv1.Package,
	packageNameToFileNameToFileDescriptorProto map[string]map[string]*descriptor.FileDescriptorProto,
) error {
	for packageName, fileNameTofileDescriptorProto := range packageNameToFileNameToFileDescriptorProto {
		pkg := packageNameToPackage[packageName]
		for _, fileDescriptorProto := range fileNameTofileDescriptorProto {
			for _, enumDescriptorProto := range fileDescriptorProto.GetEnumType() {
				enum, err := newEnum(enumDescriptorProto)
				if err != nil {
					return err
				}
				pkg.Enums = append(pkg.Enums, enum)
			}
		}
	}
	// TODO(sort)
	return nil
}

// helper for NewPackageSet
func getPackageSet(packageNameToPackage map[string]*reflectv1.Package) (*reflectv1.PackageSet, error) {
	packageSet := &reflectv1.PackageSet{
		Packages: make([]*reflectv1.Package, 0, len(packageNameToPackage)),
	}
	for _, pkg := range packageNameToPackage {
		packageSet.Packages = append(packageSet.Packages, pkg)
	}
	sort.Sort(sortPackages(packageSet.Packages))
	return packageSet, nil
}

// helper for populateDependencies
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

// helper for populateEnums
func newEnum(enumDescriptorProto *descriptor.EnumDescriptorProto) (*reflectv1.Enum, error) {
	// TODO
	return nil, nil
}

// helper for getPackageSet
type sortPackages []*reflectv1.Package

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
