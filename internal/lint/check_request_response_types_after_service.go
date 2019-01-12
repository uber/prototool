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

var requestResponseTypesAfterServiceLinter = NewLinter(
	"REQUEST_RESPONSE_TYPES_AFTER_SERVICE",
	"Verifies that request and response types are defined after any services and the response type is defined after the request type.",
	checkRequestResponseTypesAfterService,
)

func checkRequestResponseTypesAfterService(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&requestResponseTypesAfterServiceVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type requestResponseTypesAfterServiceVisitor struct {
	baseAddVisitor
	messageNameToMessage map[string]*proto.Message
	rpcs                 []*requestResponseTypesAfterServiceVisitorRPCMapKey
}

func (v *requestResponseTypesAfterServiceVisitor) OnStart(*FileDescriptor) error {
	v.messageNameToMessage = nil
	v.rpcs = nil
	return nil
}

func (v *requestResponseTypesAfterServiceVisitor) VisitMessage(message *proto.Message) {
	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	// only does top-level messages
	// nested messages are verified in checkRequestResponseTypesInSameFile
	v.messageNameToMessage[message.Name] = message
}

func (v *requestResponseTypesAfterServiceVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v *requestResponseTypesAfterServiceVisitor) VisitRPC(rpc *proto.RPC) {
	v.rpcs = append(v.rpcs, &requestResponseTypesAfterServiceVisitorRPCMapKey{
		Position:     rpc.Position,
		Name:         rpc.Name,
		RequestType:  rpc.RequestType,
		ResponseType: rpc.ReturnsType,
	})
}

func (v *requestResponseTypesAfterServiceVisitor) Finally() error {
	if len(v.rpcs) == 0 {
		return nil
	}
	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	for _, rpc := range v.rpcs {
		requestMessage, requestOK := v.messageNameToMessage[rpc.RequestType]
		if requestOK {
			if rpc.Position.Line > requestMessage.Position.Line {
				v.AddFailuref(requestMessage.Position, `Message %q is a request for RPC %q and should be defined in the file after the RPC.`, requestMessage.Name, rpc.Name)
			}
		}
		responseMessage, responseOK := v.messageNameToMessage[rpc.ResponseType]
		if responseOK {
			if rpc.Position.Line > responseMessage.Position.Line {
				v.AddFailuref(responseMessage.Position, `Message %q is a response for RPC %q and should be defined in the file after the RPC.`, responseMessage.Name, rpc.Name)
			}
		}
		if requestOK && responseOK {
			if requestMessage.Position.Line > responseMessage.Position.Line {
				v.AddFailuref(responseMessage.Position, `Message %q is a response for RPC %q and should be defined in the file after the request %q.`, responseMessage.Name, rpc.Name, requestMessage.Name)
			}
		}
	}
	return nil
}

type requestResponseTypesAfterServiceVisitorRPCMapKey struct {
	Position     scanner.Position
	Name         string
	RequestType  string
	ResponseType string
}
