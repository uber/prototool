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
	"fmt"
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

const rpcOptionsNoGoogleAPIHTTPSuppressableAnnotation = "google-api-http"

var rpcOptionsNoGoogleAPIHTTPLinter = NewLinter(
	"RPC_OPTIONS_NO_GOOGLE_API_HTTP",
	fmt.Sprintf(`Suppressable with "@suppresswarnings %s". Verifies that the RPC option "google.api.http" is not used.`, rpcOptionsNoGoogleAPIHTTPSuppressableAnnotation),
	checkRPCOptionsNoGoogleAPIHTTP,
)

func checkRPCOptionsNoGoogleAPIHTTP(add func(*text.Failure), dirPath string, descriptors []*proto.Proto) error {
	return runVisitor(&rpcOptionsNoGoogleAPIHTTPVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type rpcOptionsNoGoogleAPIHTTPVisitor struct {
	baseAddVisitor
	rpc *proto.RPC
}

func (v *rpcOptionsNoGoogleAPIHTTPVisitor) OnStart(*proto.Proto) error {
	v.rpc = nil
	return nil
}

func (v *rpcOptionsNoGoogleAPIHTTPVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v *rpcOptionsNoGoogleAPIHTTPVisitor) VisitRPC(rpc *proto.RPC) {
	v.rpc = rpc
	for _, child := range rpc.Elements {
		child.Accept(v)
	}
	v.rpc = nil
}

func (v *rpcOptionsNoGoogleAPIHTTPVisitor) VisitOption(option *proto.Option) {
	if strings.HasPrefix(option.Name, "(google.api.http)") || strings.HasPrefix(option.Name, "(.google.api.http)") {
		if v.rpc != nil && isSuppressed(v.rpc.Comment, rpcOptionsNoGoogleAPIHTTPSuppressableAnnotation) {
			return
		}
		v.AddFailuref(option.Position, `Option "google.api.http" is not allowed. This option signifies HTTP/REST usage, use the RPC framework instead. This can be suppressed by adding "@suppresswarnings %s" to the RPC comment.`, rpcOptionsNoGoogleAPIHTTPSuppressableAnnotation)
	}
}
