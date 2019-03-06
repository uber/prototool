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

func checkMessageFieldsSameName(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachMessageFieldPair(addFailure, from, to, checkMessageFieldsSameNameMessageField)
}

func checkMessageFieldsSameNameMessageField(addFailure func(*text.Failure), from *extract.MessageField, to *extract.MessageField) error {
	fromName := from.ProtoMessage().Name
	toName := to.ProtoMessage().Name
	if fromName != toName {
		addFailure(newMessageFieldsSameNameFailure(from.Message().FullyQualifiedName(), from.ProtoMessage().Number, fromName, toName))
		return nil
	}
	return nil
}

func newMessageFieldsSameNameFailure(messageName string, fieldNumber int32, fromName string, toName string) *text.Failure {
	return newTextFailuref(`Message field "%d" on message %q changed from %q to %q.`, fieldNumber, messageName, fromName, toName)
}
