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
	"strings"
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var requestResponseTypesInSameFileLinter = NewLinter(
	"REQUEST_RESPONSE_TYPES_IN_SAME_FILE",
	"Verifies that all request and response types are in the same file as their corresponding service and are not nested messages.",
	checkRequestResponseTypesInSameFile,
)

func checkRequestResponseTypesInSameFile(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&requestResponseTypesInSameFileVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type requestResponseTypesInSameFileVisitor struct {
	baseAddVisitor
	messageTypes       map[string]struct{}
	rpcs               []*requestResponseTypesInSameFileVisitorRPCMapKey
	nestedMessageNames []string
}

func (v *requestResponseTypesInSameFileVisitor) OnStart(*FileDescriptor) error {
	v.messageTypes = nil
	v.rpcs = nil
	v.nestedMessageNames = nil
	return nil
}

func (v *requestResponseTypesInSameFileVisitor) VisitMessage(message *proto.Message) {
	v.nestedMessageNames = append(v.nestedMessageNames, message.Name)
	for _, child := range message.Elements {
		child.Accept(v)
	}
	v.nestedMessageNames = v.nestedMessageNames[:len(v.nestedMessageNames)-1]

	if v.messageTypes == nil {
		v.messageTypes = make(map[string]struct{})
	}
	if len(v.nestedMessageNames) > 0 {
		v.messageTypes[strings.Join(v.nestedMessageNames, ".")+"."+message.Name] = struct{}{}
	} else {
		v.messageTypes[message.Name] = struct{}{}
	}
}

func (v *requestResponseTypesInSameFileVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v *requestResponseTypesInSameFileVisitor) VisitRPC(rpc *proto.RPC) {
	v.rpcs = append(v.rpcs, &requestResponseTypesInSameFileVisitorRPCMapKey{
		Position:     rpc.Position,
		RequestType:  rpc.RequestType,
		ResponseType: rpc.ReturnsType,
	})
}

func (v *requestResponseTypesInSameFileVisitor) Finally() error {
	if v.messageTypes == nil {
		v.messageTypes = make(map[string]struct{})
	}
	for _, rpc := range v.rpcs {
		if _, ok := v.messageTypes[rpc.RequestType]; !ok {
			v.AddFailuref(rpc.Position, "Request type %q should be defined in the same file as the corresponding service.", rpc.RequestType)
		} else if strings.ContainsRune(rpc.RequestType, '.') {
			v.AddFailuref(rpc.Position, "Request type %q is a nested message and only top-level messages should be request types.", rpc.RequestType)
		}
		if _, ok := v.messageTypes[rpc.ResponseType]; !ok {
			v.AddFailuref(rpc.Position, "Response type %q should be defined in the same file as the corresponding service.", rpc.ResponseType)
		} else if strings.ContainsRune(rpc.ResponseType, '.') {
			v.AddFailuref(rpc.Position, "Response type %q is a nested message and only top-level messages should be response types.", rpc.ResponseType)
		}
	}
	return nil
}

type requestResponseTypesInSameFileVisitorRPCMapKey struct {
	Position     scanner.Position
	RequestType  string
	ResponseType string
}
