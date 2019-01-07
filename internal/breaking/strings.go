package breaking

import (
	"fmt"

	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
)

var messageFieldTypeToString = map[int32]string{
	1:  "double",
	2:  "float",
	3:  "int64",
	4:  "uint64",
	5:  "int32",
	6:  "fixed64",
	7:  "fixed32",
	8:  "bool",
	9:  "string",
	10: "group",
	11: "message",
	12: "bytes",
	13: "uint32",
	14: "enum",
	15: "sfixed32",
	16: "sfixed64",
	17: "sint32",
	18: "sint64",
}

func getMessageFieldTypeString(messageFieldType reflectv1.MessageField_Type) (string, error) {
	s, ok := messageFieldTypeToString[int32(messageFieldType)]
	if !ok {
		return "", fmt.Errorf("unknown MessageField.Type: %d", messageFieldType)
	}
	return s, nil
}
