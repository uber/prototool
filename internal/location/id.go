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

package location

// ID is used to identify a variety of Proto types. These
// identifiers are not clearly documented, unfortunately.
// However, they can be found in the FileDescriptorSet found
// in the protoc-gen-go plugin.
//
// The following document should be treated as the source
// of truth for a variety of different Proto types:
// https://github.com/golang/protobuf/blob/master/protoc-gen-go/descriptor/descriptor.proto
//
// The syntax, for example, is identified by 12, and its
// corresponding path is [12]. This is determined by the
// following line:
// https://github.com/golang/protobuf/blob/master/protoc-gen-go/descriptor/descriptor.proto#L89
type ID int32

// The following identifiers represent Proto types, options, and/or
// qualitities of a specific type. All of the available Proto types are
// documented in in the language specification. For more, see:
// https://developers.google.com/protocol-buffers/docs/reference/proto3-spec
const (
	Syntax         ID = 12
	Package        ID = 2
	FileOption     ID = 8
	Message        ID = 4
	Field          ID = 2
	NestedType     ID = 3
	MessageEnum    ID = 4
	Oneof          ID = 8
	Enum           ID = 5
	EnumValue      ID = 2
	EnumOption     ID = 3
	Service        ID = 6
	Method         ID = 2
	MethodRequest  ID = 2
	MethodResponse ID = 3

	Name            ID = 1
	EnumValueNumber ID = 2
	FieldLabel      ID = 4
	FieldNumber     ID = 3
	FieldType       ID = 5
	FieldTypeName   ID = 6

	AllowAlias         ID = 2
	JavaPackage        ID = 1
	JavaOuterClassname ID = 8
	JavaMultipleFiles  ID = 10
	GoPackage          ID = 11
)
