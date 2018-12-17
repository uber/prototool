package compatible

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
)

func TestOneof(t *testing.T) {
	tests := []struct {
		desc     string
		original oneofs
		updated  oneofs
		err      string
	}{
		{
			desc:     "Valid update",
			original: oneofs{"foo": &oneof{name: "foo"}},
			updated:  oneofs{"foo": &oneof{name: "foo"}, "bar": &oneof{name: "bar"}},
		},
		{
			desc:     "Removed oneof",
			original: oneofs{"foo": &oneof{name: "foo"}},
			updated:  oneofs{"bar": &oneof{name: "bar"}},
			err:      `test.proto:1:1:wire:Oneof "foo" was removed.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			c := newTestChecker(t)
			fn := func(o, u descriptorProtoGroup) {
				c.checkOneofs(o.(oneofs), u.(oneofs))
			}
			check(t, c, fn, tt.original, tt.updated, tt.err)
		})
	}
}

func TestNewOneof(t *testing.T) {
	o := newOneof(&descriptor.OneofDescriptorProto{Name: proto.String("oneof")}, nil /* location.Path */)
	assert.Equal(t, "oneof", o.name)
}
