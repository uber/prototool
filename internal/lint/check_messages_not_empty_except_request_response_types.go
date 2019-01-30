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
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var messagesNotEmptyExceptRequestResponseTypesLinter = NewLinter(
	"MESSAGES_NOT_EMPTY_EXCEPT_REQUEST_RESPONSE_TYPES",
	`Verifies that all messages except for request and response types are not empty.`,
	checkMessagesNotEmptyExceptRequestResponseTypes,
)

func checkMessagesNotEmptyExceptRequestResponseTypes(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&messagesNotEmptyExceptRequestResponseTypesVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type messagesNotEmptyExceptRequestResponseTypesVisitor struct {
	baseAddVisitor
	messageNameToWrapper map[string]*messagesNotEmptyExceptRequestResponseTypesWrapper
	requestResponseTypes map[string]struct{}
	nestedMessageNames   []string
	curFieldCount        int
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) OnStart(*FileDescriptor) error {
	v.messageNameToWrapper = nil
	v.requestResponseTypes = nil
	v.nestedMessageNames = nil
	v.curFieldCount = 0
	return nil
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) VisitMessage(message *proto.Message) {
	v.nestedMessageNames = append(v.nestedMessageNames, message.Name)
	oldFieldCount := v.curFieldCount
	v.curFieldCount = 0
	for _, child := range message.Elements {
		child.Accept(v)
	}
	wrapper := &messagesNotEmptyExceptRequestResponseTypesWrapper{
		message:    message,
		fieldCount: v.curFieldCount,
	}
	v.curFieldCount = oldFieldCount
	v.nestedMessageNames = v.nestedMessageNames[:len(v.nestedMessageNames)-1]

	if v.messageNameToWrapper == nil {
		v.messageNameToWrapper = make(map[string]*messagesNotEmptyExceptRequestResponseTypesWrapper)
	}
	if len(v.nestedMessageNames) > 0 {
		v.messageNameToWrapper[strings.Join(v.nestedMessageNames, ".")+"."+message.Name] = wrapper
	} else {
		v.messageNameToWrapper[message.Name] = wrapper
	}
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) VisitOneof(oneof *proto.Oneof) {
	for _, element := range oneof.Elements {
		element.Accept(v)
	}
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) VisitNormalField(field *proto.NormalField) {
	v.curFieldCount++
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) VisitOneofField(field *proto.OneOfField) {
	v.curFieldCount++
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) VisitMapField(field *proto.MapField) {
	v.curFieldCount++
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) VisitRPC(rpc *proto.RPC) {
	if v.requestResponseTypes == nil {
		v.requestResponseTypes = make(map[string]struct{})
	}
	v.requestResponseTypes[rpc.RequestType] = struct{}{}
	v.requestResponseTypes[rpc.ReturnsType] = struct{}{}
}

func (v *messagesNotEmptyExceptRequestResponseTypesVisitor) Finally() error {
	if v.messageNameToWrapper == nil {
		v.messageNameToWrapper = make(map[string]*messagesNotEmptyExceptRequestResponseTypesWrapper)
	}
	for messageName, wrapper := range v.messageNameToWrapper {
		if _, ok := v.requestResponseTypes[messageName]; !ok {
			if wrapper.fieldCount == 0 {
				v.AddFailuref(wrapper.message.Position, `Message %q should not be empty. Only request and response types are allowed to be empty.`, messageName)
			}
		}
	}
	return nil
}

type messagesNotEmptyExceptRequestResponseTypesWrapper struct {
	message    *proto.Message
	fieldCount int
}
