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

func checkMessagesNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachPackagePair(addFailure, from, to, checkMessagesNotDeletedPackage)
}

func checkMessagesNotDeletedPackage(addFailure func(*text.Failure), from *extract.Package, to *extract.Package) error {
	return checkMessagesNotDeletedMap(addFailure, from.MessageNameToMessage(), to.MessageNameToMessage())
}

func checkMessagesNotDeletedMap(addFailure func(*text.Failure), from map[string]*extract.Message, to map[string]*extract.Message) error {
	for fromMessageName, fromMessage := range from {
		toMessage, ok := to[fromMessageName]
		if !ok {
			addFailure(newMessagesNotDeletedFailure(fromMessage.FullyQualifiedName()))
		} else if err := checkMessagesNotDeletedMap(addFailure, fromMessage.NestedMessageNameToMessage(), toMessage.NestedMessageNameToMessage()); err != nil {
			return err
		}
	}
	return nil
}

func newMessagesNotDeletedFailure(messageName string) *text.Failure {
	return newTextFailuref(`Message %q was deleted.`, messageName)
}
