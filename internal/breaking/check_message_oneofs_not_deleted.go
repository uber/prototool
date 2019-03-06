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

func checkMessageOneofsNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachMessagePair(addFailure, from, to, checkMessageOneofsNotDeletedMessage)
}

func checkMessageOneofsNotDeletedMessage(addFailure func(*text.Failure), from *extract.Message, to *extract.Message) error {
	fromOneofNameToOneof := from.OneofNameToOneof()
	toOneofNameToOneof := to.OneofNameToOneof()
	for oneofName := range fromOneofNameToOneof {
		if _, ok := toOneofNameToOneof[oneofName]; !ok {
			addFailure(newMessageOneofsNotDeletedFailure(from.FullyQualifiedName(), oneofName))
		}
	}
	return nil
}

func newMessageOneofsNotDeletedFailure(messageName string, oneofName string) *text.Failure {
	return newTextFailuref(`Message oneof %q on message %q was deleted.`, oneofName, messageName)
}
