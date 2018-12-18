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
	"github.com/uber/prototool/internal/location"
)

type testHasEnums struct {
	enums []*descriptor.EnumDescriptorProto
}

func (c *testHasEnums) GetEnumType() []*descriptor.EnumDescriptorProto {
	if c != nil {
		return c.enums
	}
	return nil
}

func newTestEnumContainer(t *testing.T, name string, values map[int32]string) *testHasEnums {
	return &testHasEnums{
		enums: []*descriptor.EnumDescriptorProto{
			{
				Name:  proto.String(name),
				Value: newTestEnumValues(t, values),
			},
		},
	}
}

func newTestEnumValues(t *testing.T, values map[int32]string) []*descriptor.EnumValueDescriptorProto {
	ev := make([]*descriptor.EnumValueDescriptorProto, len(values))
	var i int
	for num, name := range values {
		ev[i] = &descriptor.EnumValueDescriptorProto{
			Name:   proto.String(name),
			Number: proto.Int32(num),
		}
		i++
	}
	return ev
}

func TestEnum(t *testing.T) {
	tests := []struct {
		desc           string
		originalType   string
		updatedType    string
		originalValues map[int32]string
		updatedValues  map[int32]string
		err            string
	}{
		{
			desc:          "Valid update",
			originalType:  "foo",
			updatedType:   "foo",
			updatedValues: map[int32]string{1: "new"},
		},
		{
			desc:         "Removed enum type",
			originalType: "old",
			updatedType:  "new",
			err:          `test.proto:1:1:wire:Enum "old" was removed.`,
		},
		{
			desc:           "Removed enum value number",
			originalType:   "foo",
			updatedType:    "foo",
			originalValues: map[int32]string{1: "foo", 2: "foo"},
			updatedValues:  map[int32]string{1: "foo"},
			err:            `test.proto:1:1:wire:Enum value "foo" (2) was removed.`,
		},
		{
			desc:           "Renamed enum value number",
			originalType:   "foo",
			updatedType:    "foo",
			originalValues: map[int32]string{1: "old"},
			updatedValues:  map[int32]string{1: "new", 2: "another"},
			err:            `test.proto:1:1:source:Enum value "old" (1) had its name updated from "old" to "new".`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			original := getEnums(
				newTestEnumContainer(t, tt.originalType, tt.originalValues),
				nil,           // location.Path
				location.Name, // location.ID
			)
			updated := getEnums(
				newTestEnumContainer(t, tt.updatedType, tt.updatedValues),
				nil,           // location.Path
				location.Name, // location.ID
			)

			c := newTestChecker(t)
			fn := func(o, u descriptorProtoGroup) {
				c.checkEnums(o.(enums), u.(enums))
			}
			check(t, c, fn, original, updated, tt.err)
		})
	}
}
