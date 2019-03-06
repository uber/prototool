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

	"github.com/uber/prototool/internal/protostrs"
	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
)

// PackageSet is the Golang wrapper for the Protobuf PackageSet object.
type PackageSet struct {
	protoMessage *reflectv1.PackageSet

	packageNameToPackage map[string]*Package
}

// ProtoMessage returns the underlying Protobuf message.
func (p *PackageSet) ProtoMessage() *reflectv1.PackageSet {
	return p.protoMessage
}

// PackageNameToPackage returns a map from package name to Package.
func (p *PackageSet) PackageNameToPackage() map[string]*Package {
	return p.packageNameToPackage
}

// WithoutBeta makes a copy of the PackageSet without any beta packages.
//
// Note that field type names may still refer to beta packages.
func (p *PackageSet) WithoutBeta() (*PackageSet, error) {
	return newPackageSet(p.protoMessage, true)
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

// ProtoMessage returns the underlying Protobuf message.
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
	valueNameToValue   map[string]*EnumValue
	valueNumberToValue map[int32]*EnumValue
}

// ProtoMessage returns the underlying Protobuf message.
func (e *Enum) ProtoMessage() *reflectv1.Enum {
	return e.protoMessage
}

// FullyQualifiedName returns the fully-qualified name.
func (e *Enum) FullyQualifiedName() string {
	return e.fullyQualifiedName
}

// ValueNameToValue returns the values of the given Enum.
func (e *Enum) ValueNameToValue() map[string]*EnumValue {
	return e.valueNameToValue
}

// ValueNumberToValue returns the values of the given Enum.
func (e *Enum) ValueNumberToValue() map[int32]*EnumValue {
	return e.valueNumberToValue
}

// EnumValue is the Golang wrapper for the Protobuf EnumValue object.
type EnumValue struct {
	protoMessage *reflectv1.EnumValue

	enum *Enum
}

// ProtoMessage returns the underlying Protobuf enum.
func (m *EnumValue) ProtoMessage() *reflectv1.EnumValue {
	return m.protoMessage
}

// Enum returns the parent Enum.
func (m *EnumValue) Enum() *Enum {
	return m.enum
}

// Message is the Golang wrapper for the Protobuf Message object.
type Message struct {
	protoMessage *reflectv1.Message

	fullyQualifiedName         string
	nestedEnumNameToEnum       map[string]*Enum
	nestedMessageNameToMessage map[string]*Message
	fieldNameToField           map[string]*MessageField
	fieldNumberToField         map[int32]*MessageField
	oneofNameToOneof           map[string]*MessageOneof
}

// ProtoMessage returns the underlying Protobuf message.
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

// FieldNameToField returns the fields of the given Message.
func (m *Message) FieldNameToField() map[string]*MessageField {
	return m.fieldNameToField
}

// FieldNumberToField returns the fields of the given Message.
func (m *Message) FieldNumberToField() map[int32]*MessageField {
	return m.fieldNumberToField
}

// OneofNameToOneof returns the oneofs of the given Message.
func (m *Message) OneofNameToOneof() map[string]*MessageOneof {
	return m.oneofNameToOneof
}

// MessageField is the Golang wrapper for the Protobuf MessageField object.
type MessageField struct {
	protoMessage *reflectv1.MessageField

	message      *Message
	messageOneof *MessageOneof
}

// ProtoMessage returns the underlying Protobuf message.
func (m *MessageField) ProtoMessage() *reflectv1.MessageField {
	return m.protoMessage
}

// Message returns the parent Message.
func (m *MessageField) Message() *Message {
	return m.message
}

// MessageOneof returns the parent MessageOneof.
//
// This will be nil if this field is not part of a oneof.
func (m *MessageField) MessageOneof() *MessageOneof {
	return m.messageOneof
}

// MessageOneof is the Golang wrapper for the Protobuf MessageOneof object.
type MessageOneof struct {
	protoMessage *reflectv1.MessageOneof

	message            *Message
	fieldNameToField   map[string]*MessageField
	fieldNumberToField map[int32]*MessageField
}

// ProtoMessage returns the underlying Protobuf message.
func (m *MessageOneof) ProtoMessage() *reflectv1.MessageOneof {
	return m.protoMessage
}

// Message returns the parent Message.
func (m *MessageOneof) Message() *Message {
	return m.message
}

// FieldNameToField returns the fields of the given MessageOneof.
func (m *MessageOneof) FieldNameToField() map[string]*MessageField {
	return m.fieldNameToField
}

// FieldNumberToField returns the fields of the given MessageOneof.
func (m *MessageOneof) FieldNumberToField() map[int32]*MessageField {
	return m.fieldNumberToField
}

// Service is the Golang wrapper for the Protobuf Service object.
type Service struct {
	protoMessage *reflectv1.Service

	fullyQualifiedName string
	methodNameToMethod map[string]*ServiceMethod
}

// ProtoMessage returns the underlying Protobuf message.
func (s *Service) ProtoMessage() *reflectv1.Service {
	return s.protoMessage
}

// FullyQualifiedName returns the fully-qualified name.
func (s *Service) FullyQualifiedName() string {
	return s.fullyQualifiedName
}

// MethodNameToMethod returns the methods of the given Service.
func (s *Service) MethodNameToMethod() map[string]*ServiceMethod {
	return s.methodNameToMethod
}

// ServiceMethod is the Golang wrapper for the Protobuf ServiceMethod object.
type ServiceMethod struct {
	protoMessage *reflectv1.ServiceMethod

	service *Service
}

// ProtoMessage returns the underlying Protobuf service.
func (m *ServiceMethod) ProtoMessage() *reflectv1.ServiceMethod {
	return m.protoMessage
}

// Service returns the parent Service.
func (m *ServiceMethod) Service() *Service {
	return m.service
}

// NewPackageSet returns a new PackageSet for the given reflect PackageSet.
func NewPackageSet(protoMessage *reflectv1.PackageSet) (*PackageSet, error) {
	return newPackageSet(protoMessage, false)
}

func newPackageSet(protoMessage *reflectv1.PackageSet, withoutBeta bool) (*PackageSet, error) {
	packageSet := &PackageSet{
		protoMessage:         protoMessage,
		packageNameToPackage: make(map[string]*Package),
	}
	for _, pkg := range packageSet.protoMessage.Packages {
		if !ignorePackage(withoutBeta, pkg.Name) {
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
	}
	for packageName, pkg := range packageSet.packageNameToPackage {
		for _, dependencyName := range pkg.protoMessage.DependencyNames {
			dependency, ok := packageSet.packageNameToPackage[dependencyName]
			if !ok {
				if ignorePackage(withoutBeta, dependencyName) {
					continue
				}
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
			extractMessage, err := newMessage(message, packageName)
			if err != nil {
				return nil, err
			}
			pkg.messageNameToMessage[message.Name] = extractMessage
		}
		for _, service := range pkg.protoMessage.Services {
			pkg.serviceNameToService[service.Name] = newService(service, packageName)
		}
	}
	return packageSet, nil
}

func newEnum(protoMessage *reflectv1.Enum, encapsulatingFullyQualifiedName string) *Enum {
	enum := &Enum{
		protoMessage:       protoMessage,
		fullyQualifiedName: getFullyQualifiedName(encapsulatingFullyQualifiedName, protoMessage.Name),
		valueNameToValue:   make(map[string]*EnumValue),
		valueNumberToValue: make(map[int32]*EnumValue),
	}
	for _, value := range protoMessage.EnumValues {
		enumValue := newEnumValue(value, enum)
		enum.valueNameToValue[value.Name] = enumValue
		enum.valueNumberToValue[value.Number] = enumValue
	}
	return enum
}

func newEnumValue(protoMessage *reflectv1.EnumValue, enum *Enum) *EnumValue {
	return &EnumValue{
		protoMessage: protoMessage,
		enum:         enum,
	}
}

func newMessage(protoMessage *reflectv1.Message, encapsulatingFullyQualifiedName string) (*Message, error) {
	message := &Message{
		protoMessage:               protoMessage,
		fullyQualifiedName:         getFullyQualifiedName(encapsulatingFullyQualifiedName, protoMessage.Name),
		nestedEnumNameToEnum:       make(map[string]*Enum),
		nestedMessageNameToMessage: make(map[string]*Message),
		fieldNameToField:           make(map[string]*MessageField),
		fieldNumberToField:         make(map[int32]*MessageField),
		oneofNameToOneof:           make(map[string]*MessageOneof),
	}
	for _, nestedEnum := range protoMessage.NestedEnums {
		message.nestedEnumNameToEnum[nestedEnum.Name] = newEnum(nestedEnum, message.fullyQualifiedName)
	}
	for _, nestedMessage := range protoMessage.NestedMessages {
		extractMessage, err := newMessage(nestedMessage, message.fullyQualifiedName)
		if err != nil {
			return nil, err
		}
		message.nestedMessageNameToMessage[nestedMessage.Name] = extractMessage
	}
	for _, field := range protoMessage.MessageFields {
		messageField := newMessageField(field, message)
		message.fieldNameToField[field.Name] = messageField
		message.fieldNumberToField[field.Number] = messageField
	}
	for _, oneof := range protoMessage.MessageOneofs {
		messageOneof := newMessageOneof(oneof, message)
		message.oneofNameToOneof[oneof.Name] = messageOneof
	}
	if err := linkMessageFieldsOneofs(message.fieldNumberToField, message.oneofNameToOneof); err != nil {
		return nil, err
	}
	return message, nil
}

func newMessageField(protoMessage *reflectv1.MessageField, message *Message) *MessageField {
	return &MessageField{
		protoMessage: protoMessage,
		message:      message,
	}
}

func newMessageOneof(protoMessage *reflectv1.MessageOneof, message *Message) *MessageOneof {
	return &MessageOneof{
		protoMessage:       protoMessage,
		message:            message,
		fieldNameToField:   make(map[string]*MessageField),
		fieldNumberToField: make(map[int32]*MessageField),
	}
}

func newService(protoMessage *reflectv1.Service, encapsulatingFullyQualifiedName string) *Service {
	service := &Service{
		protoMessage:       protoMessage,
		fullyQualifiedName: getFullyQualifiedName(encapsulatingFullyQualifiedName, protoMessage.Name),
		methodNameToMethod: make(map[string]*ServiceMethod),
	}
	for _, method := range protoMessage.ServiceMethods {
		serviceMethod := newServiceMethod(method, service)
		service.methodNameToMethod[method.Name] = serviceMethod
	}
	return service
}

func newServiceMethod(protoMessage *reflectv1.ServiceMethod, service *Service) *ServiceMethod {
	return &ServiceMethod{
		protoMessage: protoMessage,
		service:      service,
	}
}

func linkMessageFieldsOneofs(fieldNumberToField map[int32]*MessageField, oneofNameToOneof map[string]*MessageOneof) error {
	for oneofName, oneof := range oneofNameToOneof {
		for _, fieldNumber := range oneof.protoMessage.FieldNumbers {
			field, ok := fieldNumberToField[fieldNumber]
			if !ok {
				return fmt.Errorf("oneof %s has field number %d which is not in the encapsulating message", oneofName, fieldNumber)
			}
			field.messageOneof = oneof
			oneof.fieldNameToField[field.protoMessage.Name] = field
			oneof.fieldNumberToField[field.protoMessage.Number] = field
		}
	}
	return nil
}

func getFullyQualifiedName(encapsulatingFullyQualifiedName string, name string) string {
	if encapsulatingFullyQualifiedName == "" {
		return name
	}
	return encapsulatingFullyQualifiedName + "." + name
}

func ignorePackage(withoutBeta bool, packageName string) bool {
	// if we are not ignoring beta packages, do not ignore
	if !withoutBeta {
		return false
	}
	// betaVersion is 0 if we can't parse this into a beta package
	_, betaVersion, _ := protostrs.MajorBetaVersion(packageName)
	return betaVersion > 0
}
