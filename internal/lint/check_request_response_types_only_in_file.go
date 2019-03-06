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

package lint

import (
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var requestResponseTypesOnlyInFileLinter = NewLinter(
	"REQUEST_RESPONSE_TYPES_ONLY_IN_FILE",
	"Verifies that only request and response types are the only types in the same file as their corresponding service.",
	checkRequestResponseTypesOnlyInFile,
)

func checkRequestResponseTypesOnlyInFile(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&requestResponseTypesOnlyInFileVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type requestResponseTypesOnlyInFileVisitor struct {
	baseAddVisitor
	messageNameToMessage map[string]*proto.Message
	enumNameToEnum       map[string]*proto.Enum
	rpcs                 []*requestResponseTypesOnlyInFileVisitorRPCMapKey
}

func (v *requestResponseTypesOnlyInFileVisitor) OnStart(*FileDescriptor) error {
	v.enumNameToEnum = nil
	v.messageNameToMessage = nil
	v.rpcs = nil
	return nil
}

func (v *requestResponseTypesOnlyInFileVisitor) VisitEnum(enum *proto.Enum) {
	if v.enumNameToEnum == nil {
		v.enumNameToEnum = make(map[string]*proto.Enum)
	}
	v.enumNameToEnum[enum.Name] = enum
}

func (v *requestResponseTypesOnlyInFileVisitor) VisitMessage(message *proto.Message) {
	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	// only does top-level messages
	// nested messages are verified in checkRequestResponseTypesInSameFile
	v.messageNameToMessage[message.Name] = message
}

func (v *requestResponseTypesOnlyInFileVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v *requestResponseTypesOnlyInFileVisitor) VisitRPC(rpc *proto.RPC) {
	v.rpcs = append(v.rpcs, &requestResponseTypesOnlyInFileVisitorRPCMapKey{
		Position:     rpc.Position,
		RequestType:  rpc.RequestType,
		ResponseType: rpc.ReturnsType,
	})
}

func (v *requestResponseTypesOnlyInFileVisitor) Finally() error {
	if len(v.rpcs) == 0 {
		return nil
	}
	if v.enumNameToEnum == nil {
		v.enumNameToEnum = make(map[string]*proto.Enum)
	}
	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	for _, rpc := range v.rpcs {
		delete(v.messageNameToMessage, rpc.RequestType)
		delete(v.messageNameToMessage, rpc.ResponseType)
	}
	for _, enum := range v.enumNameToEnum {
		v.AddFailuref(enum.Position, "Enum %q is in the same file as a service and should be in a separate file.", enum.Name)
	}
	for _, message := range v.messageNameToMessage {
		v.AddFailuref(message.Position, "Message %q is not a request or response of any service in this file and should be in a separate file.", message.Name)
	}
	return nil
}

type requestResponseTypesOnlyInFileVisitorRPCMapKey struct {
	Position     scanner.Position
	RequestType  string
	ResponseType string
}
