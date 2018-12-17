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
