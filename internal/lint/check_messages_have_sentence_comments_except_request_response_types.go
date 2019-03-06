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

var messagesHaveSentenceCommentsExceptRequestResponseTypesLinter = NewLinter(
	"MESSAGES_HAVE_SENTENCE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES",
	`Verifies that all non-extended messages except for request and response types have a comment that contains at least one complete sentence.`,
	checkMessagesHaveSentenceCommentsExceptRequestResponseTypes,
)

func checkMessagesHaveSentenceCommentsExceptRequestResponseTypes(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&messagesHaveSentenceCommentsExceptRequestResponseTypesVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type messagesHaveSentenceCommentsExceptRequestResponseTypesVisitor struct {
	baseAddVisitor
	messageNameToMessage map[string]*proto.Message
	requestResponseTypes map[string]struct{}
	nestedMessageNames   []string
}

func (v *messagesHaveSentenceCommentsExceptRequestResponseTypesVisitor) OnStart(*FileDescriptor) error {
	v.messageNameToMessage = nil
	v.requestResponseTypes = nil
	v.nestedMessageNames = nil
	return nil
}

func (v *messagesHaveSentenceCommentsExceptRequestResponseTypesVisitor) VisitMessage(message *proto.Message) {
	v.nestedMessageNames = append(v.nestedMessageNames, message.Name)
	for _, child := range message.Elements {
		child.Accept(v)
	}
	v.nestedMessageNames = v.nestedMessageNames[:len(v.nestedMessageNames)-1]

	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	if len(v.nestedMessageNames) > 0 {
		v.messageNameToMessage[strings.Join(v.nestedMessageNames, ".")+"."+message.Name] = message
	} else {
		v.messageNameToMessage[message.Name] = message
	}
}

func (v *messagesHaveSentenceCommentsExceptRequestResponseTypesVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v *messagesHaveSentenceCommentsExceptRequestResponseTypesVisitor) VisitRPC(rpc *proto.RPC) {
	if v.requestResponseTypes == nil {
		v.requestResponseTypes = make(map[string]struct{})
	}
	v.requestResponseTypes[rpc.RequestType] = struct{}{}
	v.requestResponseTypes[rpc.ReturnsType] = struct{}{}
}

func (v *messagesHaveSentenceCommentsExceptRequestResponseTypesVisitor) Finally() error {
	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	for messageName, message := range v.messageNameToMessage {
		if !message.IsExtend {
			if _, ok := v.requestResponseTypes[messageName]; !ok {
				if !hasCompleteSentenceComment(message.Comment) {
					v.AddFailuref(message.Position, `Message %q needs a comment with a complete sentence that starts on the first line of the comment.`, message.Name)
				}
			}
		}
	}
	return nil
}
