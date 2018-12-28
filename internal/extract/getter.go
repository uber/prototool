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

package extract

import (
	"bytes"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/strs"
	"go.uber.org/zap"
)

type getter struct {
	logger *zap.Logger
}

func newGetter(options ...GetterOption) *getter {
	getter := &getter{
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(getter)
	}
	return getter
}

func (g *getter) GetPackageSet(fileDescriptorSets []*descriptor.FileDescriptorSet) (*PackageSet, error) {
	packageNameToFileNameToFileDescriptorProto, err := getPackageNameToFileNameToFileDescriptorProto(fileDescriptorSets)
	if err != nil {
		return nil, err
	}
	fileNameToPackageName, err := getFileNameToPackageName(packageNameToFileNameToFileDescriptorProto)
	if err != nil {
		return nil, err
	}
	packageSet := &PackageSet{
		nameToPackage: make(map[string]*Package),
	}
	for packageName := range packageNameToFileNameToFileDescriptorProto {
		packageSet.nameToPackage[packageName] = &Package{
			name: packageName,
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
					packageSet.nameToPackage[packageName].deps = append(packageSet.nameToPackage[packageName].deps, depPackageName)
					packageSet.nameToPackage[depPackageName].importers = append(packageSet.nameToPackage[depPackageName].importers, packageName)
				}
			}
		}
	}
	for _, pkg := range packageSet.nameToPackage {
		pkg.deps = strs.DedupeSort(pkg.deps, nil)
		pkg.importers = strs.DedupeSort(pkg.importers, nil)
	}
	return packageSet, nil
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
