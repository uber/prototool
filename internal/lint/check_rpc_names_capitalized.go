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
	"github.com/uber/prototool/internal/strs"
	"github.com/uber/prototool/internal/text"
)

var rpcNamesCapitalizedLinter = NewLinter(
	"RPC_NAMES_CAPITALIZED",
	"Verifies that all RPC names are Capitalized.",
	checkRPCNamesCapitalized,
)

func checkRPCNamesCapitalized(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(rpcNamesCapitalizedVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type rpcNamesCapitalizedVisitor struct {
	baseAddVisitor
}

func (v rpcNamesCapitalizedVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v rpcNamesCapitalizedVisitor) VisitRPC(rpc *proto.RPC) {
	if !strs.IsCapitalized(rpc.Name) {
		v.AddFailuref(rpc.Position, "RPC name %q must be capitalized.", rpc.Name)
	}
}
