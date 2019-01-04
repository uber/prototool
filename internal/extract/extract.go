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

	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
)

// PackageSet is the Golang wrapper for the Protobuf PackageSet object.
type PackageSet struct {
	protoMessage *reflectv1.PackageSet

	packageNameToPackage map[string]*Package
}

// ProtoMessage returns the underlying Protobuf messge.
func (p *PackageSet) ProtoMessage() *reflectv1.PackageSet {
	return p.protoMessage
}

// PackageNameToPackage returns a map from package name to Package.
func (p *PackageSet) PackageNameToPackage() map[string]*Package {
	return p.packageNameToPackage
}

// Package is the Golang wrapper for the Protobuf Package object.
type Package struct {
	protoMessage *reflectv1.Package

	packageSet                 *PackageSet
	dependencyNameToDependency map[string]*Package
	importerNameToImporter     map[string]*Package
	enumNameToEnum             map[string]*Enum
	messageNameToMessage       map[string]*Message
	serviceNameToService       map[string]*Service
}

// ProtoMessage returns the underlying Protobuf messge.
func (p *Package) ProtoMessage() *reflectv1.Package {
	return p.protoMessage
}

// FullyQualifiedName returns the fully-qualified name.
func (p *Package) FullyQualifiedName() string {
	return p.protoMessage.Name
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

// EnumNameToEnum returns the nested enums of the given Package.
func (p *Package) EnumNameToEnum() map[string]*Enum {
	return p.enumNameToEnum
}

// MessageNameToMessage returns the nested messages of the given Package.
func (p *Package) MessageNameToMessage() map[string]*Message {
	return p.messageNameToMessage
}

// ServiceNameToService returns the nested services of the given Package.
func (p *Package) ServiceNameToService() map[string]*Service {
	return p.serviceNameToService
}

// Enum is the Golang wrapper for the Protobuf Enum object.
type Enum struct {
	protoMessage *reflectv1.Enum

	fullyQualifiedName string
}

// ProtoMessage returns the underlying Protobuf messge.
func (e *Enum) ProtoMessage() *reflectv1.Enum {
	return e.protoMessage
}

// FullyQualifiedName returns the fully-qualified name.
func (e *Enum) FullyQualifiedName() string {
	return e.fullyQualifiedName
}

// Message is the Golang wrapper for the Protobuf Message object.
type Message struct {
	protoMessage *reflectv1.Message

	fullyQualifiedName         string
	nestedEnumNameToEnum       map[string]*Enum
	nestedMessageNameToMessage map[string]*Message
}

// ProtoMessage returns the underlying Protobuf messge.
func (m *Message) ProtoMessage() *reflectv1.Message {
	return m.protoMessage
}

// FullyQualifiedName returns the fully-qualified name.
func (m *Message) FullyQualifiedName() string {
	return m.fullyQualifiedName
}

// NestedEnumNameToEnum returns the nested enums of the given Message.
func (m *Message) NestedEnumNameToEnum() map[string]*Enum {
	return m.nestedEnumNameToEnum
}

// NestedMessageNameToMessage returns the nested messages of the given Message.
func (m *Message) NestedMessageNameToMessage() map[string]*Message {
	return m.nestedMessageNameToMessage
}

// Service is the Golang wrapper for the Protobuf Service object.
type Service struct {
	protoMessage *reflectv1.Service

	fullyQualifiedName string
}

// ProtoMessage returns the underlying Protobuf messge.
func (s *Service) ProtoMessage() *reflectv1.Service {
	return s.protoMessage
}

// FullyQualifiedName returns the fully-qualified name.
func (s *Service) FullyQualifiedName() string {
	return s.fullyQualifiedName
}

// NewPackageSet returns a new PackageSet for the given reflect PackageSet.
func NewPackageSet(protoMessage *reflectv1.PackageSet) (*PackageSet, error) {
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
			enumNameToEnum:             make(map[string]*Enum),
			messageNameToMessage:       make(map[string]*Message),
			serviceNameToService:       make(map[string]*Service),
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
	for packageName, pkg := range packageSet.packageNameToPackage {
		for _, enum := range pkg.protoMessage.Enums {
			pkg.enumNameToEnum[enum.Name] = newEnum(enum, packageName)
		}
		for _, message := range pkg.protoMessage.Messages {
			pkg.messageNameToMessage[message.Name] = newMessage(message, packageName)
		}
		for _, service := range pkg.protoMessage.Services {
			pkg.serviceNameToService[service.Name] = newService(service, packageName)
		}
	}
	return packageSet, nil
}

func newEnum(protoMessage *reflectv1.Enum, encapsulatingFullyQualifiedName string) *Enum {
	return &Enum{
		protoMessage:       protoMessage,
		fullyQualifiedName: getFullyQualifiedName(encapsulatingFullyQualifiedName, protoMessage.Name),
	}
}

func newMessage(protoMessage *reflectv1.Message, encapsulatingFullyQualifiedName string) *Message {
	message := &Message{
		protoMessage:               protoMessage,
		fullyQualifiedName:         getFullyQualifiedName(encapsulatingFullyQualifiedName, protoMessage.Name),
		nestedEnumNameToEnum:       make(map[string]*Enum),
		nestedMessageNameToMessage: make(map[string]*Message),
	}
	for _, nestedEnum := range protoMessage.NestedEnums {
		message.nestedEnumNameToEnum[nestedEnum.Name] = newEnum(nestedEnum, message.fullyQualifiedName)
	}
	for _, nestedMessage := range protoMessage.NestedMessages {
		message.nestedMessageNameToMessage[nestedMessage.Name] = newMessage(nestedMessage, message.fullyQualifiedName)
	}
	return message
}

func newService(protoMessage *reflectv1.Service, encapsulatingFullyQualifiedName string) *Service {
	return &Service{
		protoMessage:       protoMessage,
		fullyQualifiedName: getFullyQualifiedName(encapsulatingFullyQualifiedName, protoMessage.Name),
	}
}

func getFullyQualifiedName(encapsulatingFullyQualifiedName string, name string) string {
	if encapsulatingFullyQualifiedName == "" {
		return name
	}
	return encapsulatingFullyQualifiedName + "." + name
}
