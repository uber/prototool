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
