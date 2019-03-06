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
	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var requestResponseNamesMatchRPCLinter = NewLinter(
	"REQUEST_RESPONSE_NAMES_MATCH_RPC",
	"Verifies that all request names are RpcNameRequest and all response names are RpcNameResponse.",
	checkRequestResponseNamesMatchRPC,
)

func checkRequestResponseNamesMatchRPC(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(requestResponseNamesMatchRPCVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type requestResponseNamesMatchRPCVisitor struct {
	baseAddVisitor
}

func (v requestResponseNamesMatchRPCVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v requestResponseNamesMatchRPCVisitor) VisitRPC(rpc *proto.RPC) {
	// TODO: toCamelCase for rpc.Name
	if rpc.RequestType != rpc.Name+"Request" {
		v.AddFailuref(rpc.Position, "Name of request type %q should be %q.", rpc.RequestType, rpc.Name+"Request")
	}
	if rpc.ReturnsType != rpc.Name+"Response" {
		v.AddFailuref(rpc.Position, "Name of response type %q should be %q.", rpc.ReturnsType, rpc.Name+"Response")
	}
}
