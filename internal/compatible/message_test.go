package compatible

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		desc     string
		original messages
		updated  messages
		err      string
	}{
		{
			desc:     "Valid update",
			original: messages{"foo": &message{name: "foo"}},
			updated:  messages{"foo": &message{name: "foo"}, "bar": &message{name: "bar"}},
		},
		{
			desc:     "Removed message",
			original: messages{"foo": &message{name: "foo"}},
			updated:  messages{"bar": &message{name: "bar"}},
			err:      `test.proto:1:1:wire:Message "foo" was removed.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			c := newTestChecker(t)
			fn := func(o, u descriptorProtoGroup) {
				c.checkMessages(o.(messages), u.(messages))
			}
			check(t, c, fn, tt.original, tt.updated, tt.err)
		})
	}
}

func TestNewMessage(t *testing.T) {
	t.Run("Empty message", func(t *testing.T) {
		m := newMessage(&descriptor.DescriptorProto{Name: proto.String("msg")}, nil /* location.Path */)
		assert.Equal(t, "msg", m.name)
	})
	t.Run("Non-empty message", func(t *testing.T) {
		m := newMessage(&descriptor.DescriptorProto{
			Name: proto.String("msg"),
			Field: []*descriptor.FieldDescriptorProto{
				{
					Name:   proto.String("field"),
					Number: proto.Int32(1),
				},
			},
			NestedType: []*descriptor.DescriptorProto{
				{
					Name: proto.String("nested"),
				},
			},
			OneofDecl: []*descriptor.OneofDescriptorProto{
				{
					Name: proto.String("oneof"),
				},
			},
		}, nil /* location.Path */)

		require.Len(t, m.fields, 1)
		require.Len(t, m.messages, 1)
		require.Len(t, m.oneofs, 1)

		assert.Equal(t, "msg", m.name)
		assert.Equal(t, "field", m.fields["1"].name)
		assert.Equal(t, "nested", m.messages["nested"].name)
		assert.Equal(t, "oneof", m.oneofs["oneof"].name)
	})
}
