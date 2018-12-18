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
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/location"
)

const (
	_labelPrefix   = "LABEL_"
	_typePrefix    = "TYPE_"
	_repeatedLabel = "repeated"
	_singularLabel = "singular"
	_none          = "none"
)

type fields map[string]*field

var _ descriptorProtoGroup = (fields)(nil)

func (fs fields) Items() map[string]descriptorProto {
	items := make(map[string]descriptorProto)
	for i, f := range fs {
		items[i] = f
	}
	return items
}

// field represents a *descriptor.FieldDescriptorProto.
type field struct {
	path     location.Path
	typeID   location.ID
	name     string
	typeName string
	jsonName string
	label    string
	oneof    string
	number   int32
}

var _ descriptorProto = (*field)(nil)

func (f *field) Name() string        { return f.name }
func (f *field) Path() location.Path { return f.path }
func (f *field) Type() string        { return fmt.Sprintf("Field %q (%d)", f.name, f.number) }

func newField(fd *descriptor.FieldDescriptorProto, os []*descriptor.OneofDescriptorProto, p location.Path) *field {
	typeName, id := getFieldType(fd)
	return &field{
		path:     p,
		name:     fd.GetName(),
		jsonName: fd.GetJsonName(),
		typeName: typeName,
		typeID:   id,
		label:    getFieldLabel(fd),
		number:   fd.GetNumber(),
		oneof:    getOneof(fd, os),
	}
}

// checkFields verifies that,
//  - None of the field values/numbers were removed.
//  - None of the field names were updated.
//  - None of the field json names were updated.
//  - None of the field types were updated.
//  - None of the field labels were updated.
//  - None of the field oneof declarations were updated.
func (c *fileChecker) checkFields(original, updated fields) {
	c.checkRemovedItems(original, updated, location.FieldNumber)
	for i, uf := range updated {
		if of, ok := original[i]; ok {
			c.checkField(of, uf)
		}
	}
}

func (c *fileChecker) checkField(original, updated *field) {
	c.checkUpdatedAttribute(
		original,
		Wire,
		"json name",
		original.jsonName,
		updated.jsonName,
		location.Name,
	)
	c.checkUpdatedAttribute(
		original,
		getFieldTypeSeverity(original.typeName, updated.typeName),
		"type",
		original.typeName,
		updated.typeName,
		original.typeID,
	)
	c.checkUpdatedAttribute(
		original,
		Wire,
		"oneof declaration",
		original.oneof,
		updated.oneof,
		location.Name,
	)
	// The label can be safely evolved from "singular" to
	// "repeated" with respect to wire-compatibility.
	//
	// In the case of a "singular" label, the label
	// doesn't actually exist, so we set our target
	// to the field's name.
	severity, target := Source, location.Name
	if original.label == _repeatedLabel {
		// If the original label was "repeated", then
		// any update is wire-incompatible, i.e.
		// an update from "repeated" to "singular".
		severity, target = Wire, location.FieldLabel
	}
	c.checkUpdatedAttribute(
		original,
		severity,
		"label",
		original.label,
		updated.label,
		target,
	)
}

// getFieldType maps the given field's type to a more aesthetically appropriate
// representation. We prioritize the TypeName representation if it is set.
// Note that the location.ID changes based on whether the type was derived from
// the field's type name or its plain type.
//
//  For example,
//   FieldDescriptorProto.TypeName == ".foo.Bar"
//   FieldDescriptorProto.Type     == "TYPE_DOUBLE"
func getFieldType(fd *descriptor.FieldDescriptorProto) (string, location.ID) {
	if typeName := strings.TrimPrefix(fd.GetTypeName(), "."); typeName != "" {
		return typeName, location.FieldTypeName
	}
	return lowerTrimPrefix(fd.GetType().String(), _typePrefix), location.FieldType
}

// getFieldLabel maps the given field's label to a more aesthetically
// appropriate representation. Note that this also handles
// the mapping of the default label values to "singular".
//
// In proto3, the only valid label is "repeated"; "optional"
// and "required" are no longer valid.
func getFieldLabel(fd *descriptor.FieldDescriptorProto) string {
	if l := lowerTrimPrefix(fd.GetLabel().String(), _labelPrefix); l == _repeatedLabel {
		return l
	}
	return _singularLabel
}

// getOneof maps the given field's oneof index to a more aesthetically
// appropriate representation. The field contains a reference to the
// index of the containing message's oneof declarations, so we map
// the index to its corresponding name.
//
// Given that the default value of the oneof index is 0, yet this also
// represents a valid index (the first oneof, zero-indexed), we instead
// transform a missing oneof declaration to "none".
func getOneof(fd *descriptor.FieldDescriptorProto, os []*descriptor.OneofDescriptorProto) string {
	if fd.OneofIndex != nil {
		return os[fd.GetOneofIndex()].GetName()
	}
	return _none
}

// lowerTrimPrefix trims the prefix from the given string and
// transforms it to its lowercase equivalent.
// This is useful for strings received from fields, such as
// FieldDescriptorProto.Type and FielDescriptor.Label.
//
//  For example,
//   FieldDescriptorProto.Type:  "TYPE_STRING"     -> "string"
//   FieldDescriptorProto.Label: "LABEL_REPEATED"  -> "repeated"
func lowerTrimPrefix(s, prefix string) string {
	return strings.ToLower(
		strings.TrimPrefix(
			s,
			prefix,
		),
	)
}

// getFieldTypeSeverity determines the severity of a field type update.
// An int32 can change into a sint32 and still be wire-compatible,
// for example. In this case, this function would return a "source"
// severity, since it would only affect the generated source code.
//
// For more details, see:
// https://developers.google.com/protocol-buffers/docs/proto3#updating
func getFieldTypeSeverity(original, updated string) Severity {
	if compatibleWireFormat(original, updated) {
		return Source
	}
	return Wire
}

var (
	// Each []string group contains interchangeable types.
	_interchangeableGroups = [][]string{
		{"int32", "uint32", "int64", "uint64", "bool"},
		{"sint32", "sint64"},
		{"fixed32", "sfixed32"},
		{"fixed64", "sfixed64"},
	}
	_interchangeableGroupIndexes []map[string]struct{}
)

func init() {
	_interchangeableGroupIndexes = make([]map[string]struct{}, len(_interchangeableGroups))
	for i, group := range _interchangeableGroups {
		m := make(map[string]struct{}, len(group))
		for _, g := range group {
			m[g] = struct{}{}
		}
		_interchangeableGroupIndexes[i] = m
	}
}

func compatibleWireFormat(original, updated string) bool {
	for _, interchangeable := range _interchangeableGroupIndexes {
		if _, ok := interchangeable[updated]; ok {
			_, ok = interchangeable[original]
			return ok
		}
	}
	return false
}
