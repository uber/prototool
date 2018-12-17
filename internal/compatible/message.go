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
