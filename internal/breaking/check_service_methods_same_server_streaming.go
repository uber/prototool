// Copyright (c) 2021 Uber Technologies, Inc.
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

package breaking

import (
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"
)

func checkServiceMethodsSameServerStreaming(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachServiceMethodPair(addFailure, from, to, checkServiceMethodsSameServerStreamingServiceMethod)
}

func checkServiceMethodsSameServerStreamingServiceMethod(addFailure func(*text.Failure), from *extract.ServiceMethod, to *extract.ServiceMethod) error {
	fromStreaming := from.ProtoMessage().ServerStreaming
	toStreaming := to.ProtoMessage().ServerStreaming
	if fromStreaming != toStreaming {
		addFailure(newServiceMethodsSameServerStreamingFailure(from.Service().FullyQualifiedName(), from.ProtoMessage().Name, fromStreaming))
		return nil
	}
	return nil
}

func newServiceMethodsSameServerStreamingFailure(serviceName string, methodName string, fromStreaming bool) *text.Failure {
	fromStreamingString := "not server streaming"
	toStreamingString := "server streaming"
	if fromStreaming {
		fromStreamingString, toStreamingString = toStreamingString, fromStreamingString
	}
	return newTextFailuref(`Service method %q on service %q changed from %s to %s.`, methodName, serviceName, fromStreamingString, toStreamingString)
}
