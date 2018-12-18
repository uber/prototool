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

func TestField(t *testing.T) {
	tests := []struct {
		desc     string
		original fields
		updated  fields
		err      string
	}{
		{
			desc:     "Valid update",
			original: fields{"1": &field{number: 1, name: "unchanged"}},
			updated:  fields{"1": &field{number: 1, name: "unchanged"}, "2": &field{number: 2, name: "new"}},
		},
		{
			desc:     "Removed field number",
			original: fields{"1": &field{number: 1, name: "old"}},
			updated:  fields{"2": &field{number: 2, name: "new"}},
			err:      `test.proto:1:1:wire:Field "old" (1) was removed.`,
		},
		{
			desc:     "Updated json name",
			original: fields{"1": &field{number: 1, name: "old", jsonName: "old"}},
			updated:  fields{"1": &field{number: 1, name: "old", jsonName: "new"}},
			err:      `test.proto:1:1:wire:Field "old" (1) had its json name updated from "old" to "new".`,
		},
		{
			desc:     "Updated type name with wire-compatible types",
			original: fields{"1": &field{number: 1, name: "old", typeName: "int32"}},
			updated:  fields{"1": &field{number: 1, name: "old", typeName: "int64"}},
			err:      `test.proto:1:1:source:Field "old" (1) had its type updated from "int32" to "int64".`,
		},
		{
			desc:     "Updated type name with wire-incompatible types",
			original: fields{"1": &field{number: 1, name: "old", typeName: "fixed32"}},
			updated:  fields{"1": &field{number: 1, name: "old", typeName: "int64"}},
			err:      `test.proto:1:1:wire:Field "old" (1) had its type updated from "fixed32" to "int64".`,
		},
		{
			desc:     "Updated oneof declaration",
			original: fields{"1": &field{number: 1, name: "old", oneof: "old"}},
			updated:  fields{"1": &field{number: 1, name: "old", oneof: "new"}},
			err:      `test.proto:1:1:wire:Field "old" (1) had its oneof declaration updated from "old" to "new".`,
		},
		{
			desc:     `Updated label from "singular" to "repeated" (wire-compatible)`,
			original: fields{"1": &field{number: 1, name: "old", label: "singular"}},
			updated:  fields{"1": &field{number: 1, name: "old", label: "repeated"}},
			err:      `test.proto:1:1:source:Field "old" (1) had its label updated from "singular" to "repeated".`,
		},
		{
			desc:     `Updated label from "repeated" to "singular" (wire-incompatible)`,
			original: fields{"1": &field{number: 1, name: "old", label: "repeated"}},
			updated:  fields{"1": &field{number: 1, name: "old", label: "singular"}},
			err:      `test.proto:1:1:wire:Field "old" (1) had its label updated from "repeated" to "singular".`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			c := newTestChecker(t)
			fn := func(o, u descriptorProtoGroup) {
				c.checkFields(o.(fields), u.(fields))
			}
			check(t, c, fn, tt.original, tt.updated, tt.err)
		})
	}
}

func TestNewField(t *testing.T) {
	t.Run("Hydrated field", func(t *testing.T) {
		typ := descriptor.FieldDescriptorProto_TYPE_MESSAGE
		fd := &descriptor.FieldDescriptorProto{
			Name:     proto.String("name"),
			TypeName: proto.String(".foo.Bar"),
			Type:     &typ,
			JsonName: proto.String("json"),
			Number:   proto.Int32(1),
		}
		f := newField(fd, nil /* Oneofs */, nil /* location.Path */)
		assert.Equal(t, "name", f.name)
		assert.Equal(t, "foo.Bar", f.typeName)
		assert.Equal(t, "json", f.jsonName)
		assert.Equal(t, "singular", f.label)
		assert.Equal(t, int32(1), f.number)
	})
	t.Run("Type name is not set", func(t *testing.T) {
		typ := descriptor.FieldDescriptorProto_TYPE_MESSAGE
		fd := &descriptor.FieldDescriptorProto{
			Name:   proto.String("name"),
			Type:   &typ,
			Number: proto.Int32(1),
		}
		f := newField(fd, nil /* Oneofs */, nil /* location.Path */)
		assert.Equal(t, fd.GetName(), f.name)
		assert.Equal(t, "message", f.typeName)
		assert.Equal(t, "singular", f.label)
	})
	t.Run("Repeated field", func(t *testing.T) {
		label := descriptor.FieldDescriptorProto_LABEL_REPEATED
		fd := &descriptor.FieldDescriptorProto{
			Name:  proto.String("name"),
			Label: &label,
		}
		f := newField(fd, nil /* Oneofs */, nil /* location.Path */)
		assert.Equal(t, fd.GetName(), f.name)
		assert.Equal(t, "repeated", f.label)
	})
	t.Run("Oneof declaration", func(t *testing.T) {
		oneofs := []*descriptor.OneofDescriptorProto{
			{
				Name: proto.String("oneof"),
			},
		}
		fd := &descriptor.FieldDescriptorProto{
			Name:       proto.String("name"),
			OneofIndex: proto.Int32(0),
		}
		f := newField(fd, oneofs, nil /* location.Path */)
		assert.Equal(t, fd.GetName(), f.name)
		assert.Equal(t, "oneof", f.oneof)
	})
}

func TestWireCompatibleTypes(t *testing.T) {
	tests := []struct {
		desc           string
		original       string
		updated        string
		wantCompatible bool
	}{
		{
			desc:           "Identity",
			original:       "int32",
			updated:        "int32",
			wantCompatible: true,
		},
		{
			desc:           "Unsigned integer types",
			original:       "uint32",
			updated:        "uint64",
			wantCompatible: true,
		},
		{
			desc:           "Signed integer types",
			original:       "sint32",
			updated:        "sint64",
			wantCompatible: true,
		},
		{
			desc:           "Bool with compatible integer",
			original:       "bool",
			updated:        "int64",
			wantCompatible: true,
		},
		{
			desc:           "Bool with incompatible integer",
			original:       "bool",
			updated:        "sint32",
			wantCompatible: false,
		},
		{
			desc:           "Fixed 32-bit types",
			original:       "fixed32",
			updated:        "sfixed32",
			wantCompatible: true,
		},
		{
			desc:           "Fixed 64-bit types",
			original:       "fixed64",
			updated:        "sfixed64",
			wantCompatible: true,
		},
		{
			desc:           "Incompatible fixed types",
			original:       "fixed32",
			updated:        "fixed64",
			wantCompatible: false,
		},
		{
			desc:           "Primitive mixed with custom type",
			original:       "int32",
			updated:        "foo.Bar",
			wantCompatible: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			compatible := compatibleWireFormat(tt.original, tt.updated)
			assert.Equal(t, tt.wantCompatible, compatible)
		})
	}
}
