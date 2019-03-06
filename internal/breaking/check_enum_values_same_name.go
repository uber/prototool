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

func checkEnumValuesSameName(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachEnumValuePair(addFailure, from, to, checkEnumValuesSameNameEnumValue)
}

func checkEnumValuesSameNameEnumValue(addFailure func(*text.Failure), from *extract.EnumValue, to *extract.EnumValue) error {
	fromName := from.ProtoMessage().Name
	toName := to.ProtoMessage().Name
	if fromName != toName {
		addFailure(newEnumValuesSameNameFailure(from.Enum().FullyQualifiedName(), from.ProtoMessage().Number, fromName, toName))
		return nil
	}
	return nil
}

func newEnumValuesSameNameFailure(enumName string, valueNumber int32, fromName string, toName string) *text.Failure {
	return newTextFailuref(`Enum value "%d" on enum %q changed from %q to %q.`, valueNumber, enumName, fromName, toName)
}
