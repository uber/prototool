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

func TestMethod(t *testing.T) {
	tests := []struct {
		desc     string
		original methods
		updated  methods
		err      string
	}{
		{
			desc:     "Valid update",
			original: methods{"foo": &method{name: "foo"}},
			updated:  methods{"foo": &method{name: "foo"}, "bar": &method{name: "bar"}},
		},
		{
			desc:     "Removed method",
			original: methods{"foo": &method{name: "foo"}},
			updated:  methods{"bar": &method{name: "bar"}},
			err:      `test.proto:1:1:wire:Method "foo" was removed.`,
		},
		{
			desc:     "Updated input",
			original: methods{"foo": &method{name: "foo", input: "foo.FooRequest"}},
			updated:  methods{"foo": &method{name: "foo", input: "foo.AnotherRequest"}},
			err:      `test.proto:1:1:wire:Method "foo" had its request type updated from "foo.FooRequest" to "foo.AnotherRequest".`,
		},
		{
			desc:     "Updated output",
			original: methods{"foo": &method{name: "foo", output: "foo.FooResponse"}},
			updated:  methods{"foo": &method{name: "foo", output: "foo.AnotherResponse"}},
			err:      `test.proto:1:1:wire:Method "foo" had its response type updated from "foo.FooResponse" to "foo.AnotherResponse".`,
		},
		{
			desc:     "Updated client-streaming (wire-compatible)",
			original: methods{"foo": &method{name: "foo", clientStreaming: false}},
			updated:  methods{"foo": &method{name: "foo", clientStreaming: true}},
			err:      `test.proto:1:1:source:Method "foo" had its client streaming updated from "false" to "true".`,
		},
		{
			desc:     "Updated client-streaming (wire-incompatible)",
			original: methods{"foo": &method{name: "foo", clientStreaming: true}},
			updated:  methods{"foo": &method{name: "foo", clientStreaming: false}},
			err:      `test.proto:1:1:wire:Method "foo" had its client streaming updated from "true" to "false".`,
		},
		{
			desc:     "Updated client-streaming (wire-compatible)",
			original: methods{"foo": &method{name: "foo", serverStreaming: false}},
			updated:  methods{"foo": &method{name: "foo", serverStreaming: true}},
			err:      `test.proto:1:1:source:Method "foo" had its server streaming updated from "false" to "true".`,
		},
		{
			desc:     "Updated client-streaming (wire-incompatible)",
			original: methods{"foo": &method{name: "foo", serverStreaming: true}},
			updated:  methods{"foo": &method{name: "foo", serverStreaming: false}},
			err:      `test.proto:1:1:wire:Method "foo" had its server streaming updated from "true" to "false".`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			c := newTestChecker(t)
			fn := func(o, u descriptorProtoGroup) {
				c.checkMethods(o.(methods), u.(methods))
			}
			check(t, c, fn, tt.original, tt.updated, tt.err)
		})
	}
}

func TestNewMethod(t *testing.T) {
	t.Run("Empty method", func(t *testing.T) {
		m := newMethod(&descriptor.MethodDescriptorProto{Name: proto.String("method")}, nil /* location.Path */)
		assert.Equal(t, "method", m.name)
	})
	t.Run("Non-empty method", func(t *testing.T) {
		m := newMethod(&descriptor.MethodDescriptorProto{
			Name:            proto.String("method"),
			InputType:       proto.String(".foo.BarRequest"),
			OutputType:      proto.String(".foo.BarResponse"),
			ClientStreaming: proto.Bool(true),
		}, nil /* location.Path */)

		assert.Equal(t, "method", m.name)
		assert.Equal(t, "foo.BarRequest", m.input)
		assert.Equal(t, "foo.BarResponse", m.output)
		assert.True(t, m.clientStreaming)
		assert.False(t, m.serverStreaming)
	})
}
