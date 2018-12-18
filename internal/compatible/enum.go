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

type enums map[string]*enum

var _ descriptorProtoGroup = (enums)(nil)

func (es enums) Items() map[string]descriptorProto {
	items := make(map[string]descriptorProto)
	for s, e := range es {
		items[s] = e
	}
	return items
}

// enum represents a *descriptor.EnumDescriptorProto.
type enum struct {
	path   location.Path
	name   string
	values enumValues
}

var _ descriptorProto = (*enum)(nil)

func (e *enum) Name() string        { return e.name }
func (e *enum) Path() location.Path { return e.path }
func (e *enum) Type() string        { return fmt.Sprintf("Enum %q", e.name) }

type enumValues map[string]*enumValue

var _ descriptorProtoGroup = (enumValues)(nil)

func (es enumValues) Items() map[string]descriptorProto {
	m := make(map[string]descriptorProto)
	for i, e := range es {
		m[i] = e
	}
	return m
}

// enumValue represents a *descriptor.EnumValueDescriptorProto.
type enumValue struct {
	path   location.Path
	name   string
	number int32
}

var _ descriptorProto = (*enumValue)(nil)

func (e *enumValue) Name() string        { return e.name }
func (e *enumValue) Path() location.Path { return e.path }
func (e *enumValue) Type() string        { return fmt.Sprintf("Enum value %q (%d)", e.name, e.number) }

// hasEnums is implemented by both the *descriptor.FileDescriptorProto
// and *descriptor.DescriptorProto types.
type hasEnums interface {
	GetEnumType() []*descriptor.EnumDescriptorProto
}

// getEnums is used to construct a collection of enums from a
// type that has enums. Note that the location identifier
// differs based on the parent type. For file descriptors, the
// identifier is location.Enum, whereas for message descriptors
// the identifier is location.MessageEnum.
func getEnums(d hasEnums, p location.Path, id location.ID) enums {
	enums := make(enums, len(d.GetEnumType()))
	for i, e := range d.GetEnumType() {
		enums[e.GetName()] = newEnum(e, p.Scope(id, i))
	}
	return enums
}

func newEnum(ed *descriptor.EnumDescriptorProto, p location.Path) *enum {
	return &enum{
		path:   p,
		name:   ed.GetName(),
		values: getEnumValues(ed, p),
	}
}

func getEnumValues(ed *descriptor.EnumDescriptorProto, p location.Path) enumValues {
	values := make(enumValues, len(ed.GetValue()))
	for i, vd := range ed.GetValue() {
		values[strconv.Itoa(int(vd.GetNumber()))] = newEnumValue(vd, p.Scope(location.EnumValue, i))
	}
	return values
}

func newEnumValue(vd *descriptor.EnumValueDescriptorProto, p location.Path) *enumValue {
	return &enumValue{
		path:   p,
		name:   vd.GetName(),
		number: vd.GetNumber(),
	}
}

// checkEnums verifies that,
//  - None of the enum types were removed.
//  - None of an enum's values/numbers were removed.
//  - None of an enum's value names were updated.
func (c *fileChecker) checkEnums(original, updated enums) {
	c.checkRemovedItems(original, updated, location.Name)
	for i, ue := range updated {
		oe, ok := original[i]
		if ok {
			c.checkRemovedItems(oe.values, ue.values, location.EnumValueNumber)
		}
	}
}
