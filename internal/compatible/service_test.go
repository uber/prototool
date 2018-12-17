package compatible

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	tests := []struct {
		desc     string
		original services
		updated  services
		err      string
	}{
		{
			desc:     "Valid update",
			original: services{"foo": &service{name: "foo"}},
			updated:  services{"foo": &service{name: "foo"}, "bar": &service{name: "bar"}},
		},
		{
			desc:     "Removed service",
			original: services{"foo": &service{name: "foo"}},
			updated:  services{"bar": &service{name: "bar"}},
			err:      `test.proto:1:1:wire:Service "foo" was removed.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			c := newTestChecker(t)
			fn := func(o, u descriptorProtoGroup) {
				c.checkServices(o.(services), u.(services))
			}
			check(t, c, fn, tt.original, tt.updated, tt.err)
		})
	}
}

func TestNewService(t *testing.T) {
	t.Run("Empty service", func(t *testing.T) {
		s := newService(&descriptor.ServiceDescriptorProto{Name: proto.String("service")}, nil /* location.Path */)
		assert.Equal(t, "service", s.name)
	})
	t.Run("Non-empty service", func(t *testing.T) {
		s := newService(&descriptor.ServiceDescriptorProto{
			Name: proto.String("service"),
			Method: []*descriptor.MethodDescriptorProto{
				{
					Name: proto.String("method"),
				},
			},
		}, nil /* location.Path */)

		require.Len(t, s.methods, 1)

		assert.Equal(t, "service", s.name)
		assert.Equal(t, "method", s.methods["method"].name)
	})
}
