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
