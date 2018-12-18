// Copyright (c) 2018 Uber Technologies, Inc.
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

package compatible

import (
	"fmt"
	"strconv"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/location"
)

type messages map[string]*message

var _ descriptorProtoGroup = (messages)(nil)

func (ms messages) Items() map[string]descriptorProto {
	items := make(map[string]descriptorProto)
	for i, m := range ms {
		items[i] = m
	}
	return items
}

// message represents a *descriptor.DescriptorProto.
type message struct {
	path     location.Path
	name     string
	fields   fields
	enums    enums
	messages messages
	oneofs   oneofs
}

var _ descriptorProto = (*message)(nil)

func (m *message) Name() string        { return m.name }
func (m *message) Path() location.Path { return m.path }
func (m *message) Type() string        { return fmt.Sprintf("Message %q", m.name) }

func newMessage(md *descriptor.DescriptorProto, p location.Path) *message {
	oneofs := make(oneofs, len(md.GetOneofDecl()))
	for i, o := range md.GetOneofDecl() {
		oneofs[o.GetName()] = newOneof(o, p.Scope(location.Oneof, i))
	}
	fields := make(fields, len(md.GetField()))
	for i, f := range md.GetField() {
		fields[strconv.Itoa(int(f.GetNumber()))] = newField(f, md.GetOneofDecl(), p.Scope(location.Field, i))
	}
	messages := make(messages, len(md.GetNestedType()))
	for i, m := range md.GetNestedType() {
		messages[m.GetName()] = newMessage(m, p.Scope(location.NestedType, i))
	}
	return &message{
		path:     p,
		name:     md.GetName(),
		fields:   fields,
		enums:    getEnums(md, p, location.MessageEnum),
		messages: messages,
		oneofs:   oneofs,
	}
}

// checkMessages verifies that,
//  - None of the messages were removed.
//  - None of the messages' fields were inappropriately updated.
//  - None of the messages' enums were inappropriately updated.
//  - None of the messages' nested messages were inappropriately updated.
//  - None of the messages' oneofs were inappropriately updated.
func (c *fileChecker) checkMessages(original, updated messages) {
	c.checkRemovedItems(original, updated, location.Name)
	for i, um := range updated {
		if om, ok := original[i]; ok {
			c.checkMessage(om, um)
		}
	}
}

func (c *fileChecker) checkMessage(original, updated *message) {
	c.checkFields(original.fields, updated.fields)
	c.checkEnums(original.enums, updated.enums)
	c.checkMessages(original.messages, updated.messages)
	c.checkOneofs(original.oneofs, updated.oneofs)
}
