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

import "github.com/uber/prototool/internal/location"

// descriptorProtoGroup represents a collection of descriptorProto
// types.
//
// We define a separate interface here so that we can act upon
// both the concrete type for type-specific validation (e.g.
// a field's label), and a generic type for general validation
// (e.g. whether a type was removed).
//
// In general, the key is the descriptorProto's name.
// Numbered descriptorProto types, such as fields and enum
// values, use the string-equivalent representation of their
// number.
type descriptorProtoGroup interface {
	Items() map[string]descriptorProto
}

// descriptorProto is an interface implemented by all of the
// Protobuf types we have defined internally.
//
//  For example,
//   m := &message{name: "foo", path: [4, 0]}
//   m.Name() == "foo"
//   m.Path() == [4, 0]
//   m.Type() == `Message "foo"`
//
//   f := &field{name: "bar", number: 0, path: [4, 0, 2, 0]}
//   f.Name() == "bar"
//   f.Path() == [4, 0, 2, 0]
//   f.Type() == `Field "bar" (0)`
type descriptorProto interface {
	Name() string
	Path() location.Path
	Type() string
}

// checkRemovedItems determines if any of the descriptorProto types in the
// original collection were removed from updated collection.
func (c *fileChecker) checkRemovedItems(original, updated descriptorProtoGroup, id location.ID) {
	originalItems, updatedItems := original.Items(), updated.Items()
	for name, u := range updatedItems {
		if o, ok := originalItems[name]; ok {
			c.checkRenamedItem(o, u)
		}
	}
	for name, o := range originalItems {
		if _, ok := updatedItems[name]; !ok {
			c.AddErrorf(
				o.Path().Target(id),
				Wire,
				"%s was removed.",
				o.Type(),
			)
		}
	}
}

// checkRenamedItem is a short-hand function for determining if the original
// decsriptorProto had its name updated.
func (c *fileChecker) checkRenamedItem(original, updated descriptorProto) {
	c.checkUpdatedAttribute(
		original,
		Source,
		"name",
		original.Name(),
		updated.Name(),
		location.Name,
	)
}

// checkUpdatedAttribute determines if the original descriptorProto's attribute
// was updated.
func (c *fileChecker) checkUpdatedAttribute(typ descriptorProto, severity Severity, attribute, original, updated string, id location.ID) {
	if original != updated {
		c.AddErrorf(
			typ.Path().Target(id),
			severity,
			"%s had its %s updated from %q to %q.",
			typ.Type(),
			attribute,
			original,
			updated,
		)
	}
}
