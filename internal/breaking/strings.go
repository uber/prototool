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

	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
)

var (
	messageFieldLabelToString = map[int32]string{
		1: "optional",
		2: "required",
		3: "repeated",
	}
	messageFieldTypeToString = map[int32]string{
		1:  "double",
		2:  "float",
		3:  "int64",
		4:  "uint64",
		5:  "int32",
		6:  "fixed64",
		7:  "fixed32",
		8:  "bool",
		9:  "string",
		10: "group",
		11: "message",
		12: "bytes",
		13: "uint32",
		14: "enum",
		15: "sfixed32",
		16: "sfixed64",
		17: "sint32",
		18: "sint64",
	}
)

func getMessageFieldLabelString(messageFieldLabel reflectv1.MessageField_Label) (string, error) {
	s, ok := messageFieldLabelToString[int32(messageFieldLabel)]
	if !ok {
		return "", fmt.Errorf("unknown MessageField.Label: %d", messageFieldLabel)
	}
	return s, nil
}

func getMessageFieldTypeString(messageFieldType reflectv1.MessageField_Type) (string, error) {
	s, ok := messageFieldTypeToString[int32(messageFieldType)]
	if !ok {
		return "", fmt.Errorf("unknown MessageField.Type: %d", messageFieldType)
	}
	return s, nil
}
