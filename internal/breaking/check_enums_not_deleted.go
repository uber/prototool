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

func checkEnumsNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	if err := forEachPackagePair(addFailure, from, to, checkEnumsNotDeletedPackage); err != nil {
		return err
	}
	// if a message is deleted, the breaking change detector will fail, so we don't need to check for this
	// and produce two warnings, one for the message, the other for the nested enum
	return forEachMessagePair(addFailure, from, to, checkEnumsNotDeletedMessage)
}

func checkEnumsNotDeletedPackage(addFailure func(*text.Failure), from *extract.Package, to *extract.Package) error {
	return checkEnumsNotDeletedMap(addFailure, from.EnumNameToEnum(), to.EnumNameToEnum())
}

func checkEnumsNotDeletedMessage(addFailure func(*text.Failure), from *extract.Message, to *extract.Message) error {
	return checkEnumsNotDeletedMap(addFailure, from.NestedEnumNameToEnum(), to.NestedEnumNameToEnum())
}

func checkEnumsNotDeletedMap(addFailure func(*text.Failure), from map[string]*extract.Enum, to map[string]*extract.Enum) error {
	for fromEnumName, fromEnum := range from {
		if _, ok := to[fromEnumName]; !ok {
			addFailure(newEnumsNotDeletedFailure(fromEnum.FullyQualifiedName()))
		}
	}
	return nil
}

func newEnumsNotDeletedFailure(enumName string) *text.Failure {
	return newTextFailuref(`Enum %q was deleted.`, enumName)
}
