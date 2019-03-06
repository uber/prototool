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
	if len(fileDescriptorSets) == 0 {
		return &reflectv1.PackageSet{}, nil
	}
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
	for packageName, fileNameToFileDescriptorProto := range packageNameToFileNameToFileDescriptorProto {
		pkg, ok := packageNameToPackage[packageName]
		if !ok {
			return nil, fmt.Errorf("no package for name %s", packageName)
		}
		if err := populateEnums(pkg, fileNameToFileDescriptorProto); err != nil {
			return nil, err
		}
		if err := populateMessages(pkg, fileNameToFileDescriptorProto); err != nil {
			return nil, err
		}
		if err := populateServices(pkg, fileNameToFileDescriptorProto); err != nil {
			return nil, err
		}
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
		if len(pkg.DependencyNames) != 0 {
			pkg.DependencyNames = strs.SortUniq(pkg.DependencyNames)
		}
	}
	return nil
}

// helper for NewPackageSet
func populateEnums(
	pkg *reflectv1.Package,
	fileNameToFileDescriptorProto map[string]*descriptor.FileDescriptorProto,
) error {
	for _, fileDescriptorProto := range fileNameToFileDescriptorProto {
		enums, err := getEnums(fileDescriptorProto.GetEnumType())
		if err != nil {
			return err
		}
		if len(enums) > 0 {
			pkg.Enums = append(pkg.Enums, enums...)
		}
	}
	sort.Slice(pkg.Enums, func(i int, j int) bool { return pkg.Enums[i].Name < pkg.Enums[j].Name })
	return nil
}

// helper for NewPackageSet
func populateMessages(
	pkg *reflectv1.Package,
	fileNameToFileDescriptorProto map[string]*descriptor.FileDescriptorProto,
) error {
	for _, fileDescriptorProto := range fileNameToFileDescriptorProto {
		messages, err := getMessages(fileDescriptorProto.GetMessageType())
		if err != nil {
			return err
		}
		if len(messages) > 0 {
			pkg.Messages = append(pkg.Messages, messages...)
		}
	}
	sort.Slice(pkg.Messages, func(i int, j int) bool { return pkg.Messages[i].Name < pkg.Messages[j].Name })
	return nil
}

// helper for NewPackageSet
func populateServices(
	pkg *reflectv1.Package,
	fileNameToFileDescriptorProto map[string]*descriptor.FileDescriptorProto,
) error {
	for _, fileDescriptorProto := range fileNameToFileDescriptorProto {
		services, err := getServices(fileDescriptorProto.GetService())
		if err != nil {
			return err
		}
		if len(services) > 0 {
			pkg.Services = append(pkg.Services, services...)
		}
	}
	sort.Slice(pkg.Services, func(i int, j int) bool { return pkg.Services[i].Name < pkg.Services[j].Name })
	return nil
}

// helper for NewPackageSet
func getPackageSet(packageNameToPackage map[string]*reflectv1.Package) (*reflectv1.PackageSet, error) {
	if len(packageNameToPackage) == 0 {
		return &reflectv1.PackageSet{}, nil
	}
	packageSet := &reflectv1.PackageSet{
		Packages: make([]*reflectv1.Package, 0, len(packageNameToPackage)),
	}
	for _, pkg := range packageNameToPackage {
		packageSet.Packages = append(packageSet.Packages, pkg)
	}
	sort.Slice(packageSet.Packages, func(i int, j int) bool { return packageSet.Packages[i].Name < packageSet.Packages[j].Name })
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

func getEnums(enumDescriptorProtos []*descriptor.EnumDescriptorProto) ([]*reflectv1.Enum, error) {
	if len(enumDescriptorProtos) == 0 {
		return nil, nil
	}
	enums := make([]*reflectv1.Enum, 0, len(enumDescriptorProtos))
	for _, enumDescriptorProto := range enumDescriptorProtos {
		enum, err := newEnum(enumDescriptorProto)
		if err != nil {
			return nil, err
		}
		enums = append(enums, enum)
	}
	sort.Slice(enums, func(i int, j int) bool { return enums[i].Name < enums[j].Name })
	return enums, nil
}

func newEnum(enumDescriptorProto *descriptor.EnumDescriptorProto) (*reflectv1.Enum, error) {
	enum := &reflectv1.Enum{
		Name: enumDescriptorProto.GetName(),
	}
	for _, enumValueDescriptorProto := range enumDescriptorProto.GetValue() {
		enum.EnumValues = append(enum.EnumValues, &reflectv1.EnumValue{
			Name:   enumValueDescriptorProto.GetName(),
			Number: enumValueDescriptorProto.GetNumber(),
		})
	}
	sort.Slice(enum.EnumValues, func(i int, j int) bool { return enum.EnumValues[i].Number < enum.EnumValues[j].Number })
	return enum, nil
}

func getMessages(descriptorProtos []*descriptor.DescriptorProto) ([]*reflectv1.Message, error) {
	if len(descriptorProtos) == 0 {
		return nil, nil
	}
	messages := make([]*reflectv1.Message, 0, len(descriptorProtos))
	for _, descriptorProto := range descriptorProtos {
		message, err := newMessage(descriptorProto)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	sort.Slice(messages, func(i int, j int) bool { return messages[i].Name < messages[j].Name })
	return messages, nil
}

func newMessage(descriptorProto *descriptor.DescriptorProto) (*reflectv1.Message, error) {
	nestedMessages, err := getMessages(descriptorProto.GetNestedType())
	if err != nil {
		return nil, err
	}
	nestedEnums, err := getEnums(descriptorProto.GetEnumType())
	if err != nil {
		return nil, err
	}
	message := &reflectv1.Message{
		Name:           descriptorProto.GetName(),
		NestedMessages: nestedMessages,
		NestedEnums:    nestedEnums,
	}
	nameToMessageOneof := make(map[string]*reflectv1.MessageOneof, len(descriptorProto.GetOneofDecl()))
	for _, oneofDescriptorProto := range descriptorProto.GetOneofDecl() {
		nameToMessageOneof[oneofDescriptorProto.GetName()] = &reflectv1.MessageOneof{
			Name: oneofDescriptorProto.GetName(),
		}
	}
	for _, fieldDescriptorProto := range descriptorProto.GetField() {
		typeName := fieldDescriptorProto.GetTypeName()
		if typeName != "" {
			typeName, err = verifyFullyQualifiedNameAndStrip(typeName)
			if err != nil {
				return nil, err
			}
		}
		message.MessageFields = append(message.MessageFields, &reflectv1.MessageField{
			Name:   fieldDescriptorProto.GetName(),
			Number: fieldDescriptorProto.GetNumber(),
			// TODO: this is technically unsafe since we're just working on the assumption
			// that the numbers match up...which they do, but this isn't future proof
			// however, the values for descriptor.proto have not changed since proto1
			// which isn't even OSS, so we're probably fine for 10-20 years
			Type:     reflectv1.MessageField_Type(fieldDescriptorProto.GetType()),
			Label:    reflectv1.MessageField_Label(fieldDescriptorProto.GetLabel()),
			TypeName: typeName,
		})
		if fieldDescriptorProto.OneofIndex != nil {
			// TODO: super unsafe
			messageOneof := nameToMessageOneof[descriptorProto.GetOneofDecl()[fieldDescriptorProto.GetOneofIndex()].GetName()]
			messageOneof.FieldNumbers = append(messageOneof.FieldNumbers, fieldDescriptorProto.GetNumber())
		}
	}
	for _, messageOneof := range nameToMessageOneof {
		sort.Slice(messageOneof.FieldNumbers, func(i int, j int) bool { return messageOneof.FieldNumbers[i] < messageOneof.FieldNumbers[j] })
		message.MessageOneofs = append(message.MessageOneofs, messageOneof)
	}
	sort.Slice(message.MessageFields, func(i int, j int) bool { return message.MessageFields[i].Number < message.MessageFields[j].Number })
	sort.Slice(message.MessageOneofs, func(i int, j int) bool { return message.MessageOneofs[i].Name < message.MessageOneofs[j].Name })
	return message, nil
}

func getServices(serviceDescriptorProtos []*descriptor.ServiceDescriptorProto) ([]*reflectv1.Service, error) {
	if len(serviceDescriptorProtos) == 0 {
		return nil, nil
	}
	services := make([]*reflectv1.Service, 0, len(serviceDescriptorProtos))
	for _, serviceDescriptorProto := range serviceDescriptorProtos {
		service, err := newService(serviceDescriptorProto)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	sort.Slice(services, func(i int, j int) bool { return services[i].Name < services[j].Name })
	return services, nil
}

func newService(serviceDescriptorProto *descriptor.ServiceDescriptorProto) (*reflectv1.Service, error) {
	service := &reflectv1.Service{
		Name: serviceDescriptorProto.GetName(),
	}
	for _, methodDescriptorProto := range serviceDescriptorProto.GetMethod() {
		serviceMethod, err := newServiceMethod(methodDescriptorProto)
		if err != nil {
			return nil, err
		}
		service.ServiceMethods = append(service.ServiceMethods, serviceMethod)
	}
	sort.Slice(service.ServiceMethods, func(i int, j int) bool { return service.ServiceMethods[i].Name < service.ServiceMethods[j].Name })
	return service, nil
}

func newServiceMethod(methodDescriptorProto *descriptor.MethodDescriptorProto) (*reflectv1.ServiceMethod, error) {
	requestTypeName, err := verifyFullyQualifiedNameAndStrip(methodDescriptorProto.GetInputType())
	if err != nil {
		return nil, err
	}
	responseTypeName, err := verifyFullyQualifiedNameAndStrip(methodDescriptorProto.GetOutputType())
	if err != nil {
		return nil, err
	}
	return &reflectv1.ServiceMethod{
		Name:             methodDescriptorProto.GetName(),
		RequestTypeName:  requestTypeName,
		ResponseTypeName: responseTypeName,
		ClientStreaming:  methodDescriptorProto.GetClientStreaming(),
		ServerStreaming:  methodDescriptorProto.GetServerStreaming(),
	}, nil
}

func verifyFullyQualifiedNameAndStrip(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("name empty")
	}
	if s[0] != '.' {
		return "", fmt.Errorf("%s does not start with '.'", s)
	}
	return s[1:], nil
}
