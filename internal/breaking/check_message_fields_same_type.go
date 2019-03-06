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
	"fmt"

	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"

	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
)

func checkMessageFieldsSameType(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachMessageFieldPair(addFailure, from, to, checkMessageFieldsSameTypeMessageField)
}

func checkMessageFieldsSameTypeMessageField(addFailure func(*text.Failure), from *extract.MessageField, to *extract.MessageField) error {
	fromType := from.ProtoMessage().Type
	toType := to.ProtoMessage().Type
	// TODO: message type name
	if fromType != toType {
		fromTypeString, err := getMessageFieldTypeString(fromType)
		if err != nil {
			return err
		}
		toTypeString, err := getMessageFieldTypeString(toType)
		if err != nil {
			return err
		}
		addFailure(newMessageFieldsSameTypeFailure(from.Message().FullyQualifiedName(), from.ProtoMessage().Number, fromTypeString, toTypeString))
		return nil
	}
	switch fromType {
	case reflectv1.MessageField_TYPE_ENUM, reflectv1.MessageField_TYPE_GROUP, reflectv1.MessageField_TYPE_MESSAGE:
		fromTypeName := from.ProtoMessage().TypeName
		toTypeName := to.ProtoMessage().TypeName
		if fromTypeName == "" {
			return fmt.Errorf("fromTypeName empty")
		}
		if toTypeName == "" {
			return fmt.Errorf("toTypeName empty")
		}
		if fromTypeName != toTypeName {
			addFailure(newMessageFieldsSameTypeFailure(from.Message().FullyQualifiedName(), from.ProtoMessage().Number, fromTypeName, toTypeName))
		}
	}
	return nil
}

func newMessageFieldsSameTypeFailure(messageName string, fieldNumber int32, fromTypeString string, toTypeString string) *text.Failure {
	return newTextFailuref(`Message field "%d" on message %q changed type from %q to %q.`, fieldNumber, messageName, fromTypeString, toTypeString)
}
