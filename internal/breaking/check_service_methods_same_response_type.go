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

package breaking

import (
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"
)

func checkServiceMethodsSameResponseType(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachServiceMethodPair(addFailure, from, to, checkServiceMethodsSameResponseTypeServiceMethod)
}

func checkServiceMethodsSameResponseTypeServiceMethod(addFailure func(*text.Failure), from *extract.ServiceMethod, to *extract.ServiceMethod) error {
	fromTypeName := from.ProtoMessage().ResponseTypeName
	toTypeName := to.ProtoMessage().ResponseTypeName
	if fromTypeName != toTypeName {
		addFailure(newServiceMethodsSameResponseTypeFailure(from.Service().FullyQualifiedName(), from.ProtoMessage().Name, fromTypeName, toTypeName))
		return nil
	}
	return nil
}

func newServiceMethodsSameResponseTypeFailure(serviceName string, methodName string, fromTypeName string, toTypeName string) *text.Failure {
	return newTextFailuref(`Service method %q on service %q changed response type from %q to %q.`, methodName, serviceName, fromTypeName, toTypeName)
}
