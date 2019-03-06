// Copyright (c) 2019 Uber Technologies, Inc.
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

package reflect

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/require"
	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
	ptesting "github.com/uber/prototool/internal/testing"
)

func TestOne(t *testing.T) {
	testNewPackageSet(
		t,
		"one",
		`
{
  "packages": [
    {
      "name": "uber.proto.bar.v1",
      "messages": [
        {
          "name": "OneBar",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            }
          ]
        },
        {
          "name": "OneBaz",
          "messageFields": [
            {
              "name": "one_foo",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.OneFoo"
            },
            {
              "name": "one_bar",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.OneBar"
            },
            {
              "name": "two_foo",
              "number": 3,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.TwoFoo"
            },
            {
              "name": "two_bar",
              "number": 4,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.TwoBar"
            }
          ]
        },
        {
          "name": "OneFoo",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            },
            {
              "name": "one_bar",
              "number": 3,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.OneBar"
            }
          ]
        },
        {
          "name": "TwoBar",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            }
          ]
        },
        {
          "name": "TwoFoo",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            },
            {
              "name": "two_bar",
              "number": 3,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.TwoBar"
            }
          ]
        }
      ]
    },
    {
      "name": "uber.proto.foo.v1",
      "dependencyNames": [
        "uber.proto.bar.v1"
      ],
      "enums": [
        {
          "name": "Enum",
          "enumValues": [
            {
              "name": "ENUM_INVALID"
            },
            {
              "name": "ENUM_FOO",
              "number": 1
            },
            {
              "name": "ENUM_BAR",
              "number": 2
            }
          ]
        }
      ],
      "messages": [
        {
          "name": "BarRequest",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            }
          ]
        },
        {
          "name": "BarResponse",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            }
          ]
        },
        {
          "name": "FooRequest",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            }
          ]
        },
        {
          "name": "FooResponse",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            }
          ]
        },
        {
          "name": "OneBar",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            }
          ]
        },
        {
          "name": "OneBat",
          "messageFields": [
            {
              "name": "one_foo",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.OneFoo"
            },
            {
              "name": "one_bar",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.OneBar"
            },
            {
              "name": "two_foo",
              "number": 3,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.TwoFoo"
            },
            {
              "name": "two_bar",
              "number": 4,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.bar.v1.TwoBar"
            }
          ]
        },
        {
          "name": "OneBaz",
          "messageFields": [
            {
              "name": "one_foo",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.OneFoo"
            },
            {
              "name": "one_bar",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.OneBar"
            },
            {
              "name": "two_foo",
              "number": 3,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.TwoFoo"
            },
            {
              "name": "two_bar",
              "number": 4,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.TwoBar"
            }
          ]
        },
        {
          "name": "OneFoo",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            },
            {
              "name": "one_bar",
              "number": 3,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.OneBar"
            }
          ]
        },
        {
          "name": "Simple",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            },
            {
              "name": "three",
              "number": 3,
              "label": "LABEL_REPEATED",
              "type": "TYPE_INT64"
            },
            {
              "name": "four",
              "number": 4,
              "label": "LABEL_REPEATED",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.Simple.FourEntry"
            },
            {
              "name": "five",
              "number": 5,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "one_bat",
              "number": 6,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.OneBat"
            }
          ],
          "messageOneofs": [
            {
              "name": "test_oneof",
              "fieldNumbers": [
                5,
                6
              ]
            }
          ],
          "nestedMessages": [
            {
              "name": "FourEntry",
              "messageFields": [
                {
                  "name": "key",
                  "number": 1,
                  "label": "LABEL_OPTIONAL",
                  "type": "TYPE_INT64"
                },
                {
                  "name": "value",
                  "number": 2,
                  "label": "LABEL_OPTIONAL",
                  "type": "TYPE_STRING"
                }
              ]
            },
            {
              "name": "Nested",
              "messageFields": [
                {
                  "name": "one",
                  "number": 1,
                  "label": "LABEL_OPTIONAL",
                  "type": "TYPE_INT64"
                },
                {
                  "name": "two",
                  "number": 2,
                  "label": "LABEL_OPTIONAL",
                  "type": "TYPE_STRING"
                }
              ]
            }
          ],
          "nestedEnums": [
            {
              "name": "NestedEnum",
              "enumValues": [
                {
                  "name": "NESTED_ENUM_INVALID"
                },
                {
                  "name": "NESTED_ENUM_FOO",
                  "number": 1
                },
                {
                  "name": "NESTED_ENUM_BAR",
                  "number": 2
                }
              ]
            }
          ]
        },
        {
          "name": "TwoBar",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            }
          ]
        },
        {
          "name": "TwoFoo",
          "messageFields": [
            {
              "name": "one",
              "number": 1,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_INT64"
            },
            {
              "name": "two",
              "number": 2,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_STRING"
            },
            {
              "name": "two_bar",
              "number": 3,
              "label": "LABEL_OPTIONAL",
              "type": "TYPE_MESSAGE",
              "typeName": "uber.proto.foo.v1.TwoBar"
            }
          ]
        }
      ],
      "services": [
        {
          "name": "SomeAPI",
          "serviceMethods": [
            {
              "name": "Bar",
              "requestTypeName": "uber.proto.foo.v1.BarRequest",
              "responseTypeName": "uber.proto.foo.v1.BarResponse"
            },
            {
              "name": "Foo",
              "requestTypeName": "uber.proto.foo.v1.FooRequest",
              "responseTypeName": "uber.proto.foo.v1.FooResponse"
            }
          ]
        }
      ]
    }
  ]
}
`,
	)
}

func testNewPackageSet(t *testing.T, subDirPath string, packageSetJSON string) {
	fileDescriptorSets := ptesting.RequireGetFileDescriptorSets(t, ".", "testdata/"+subDirPath)
	packageSet, err := NewPackageSet(fileDescriptorSets.Unwrap()...)
	require.NoError(t, err)
	// It's much easier to edit JSON than the actual Golang structs all the type
	expectedPackageSet := requireUnmarshalPackageSet(t, packageSetJSON)
	require.Equal(t, expectedPackageSet, packageSet)
}

func requireUnmarshalPackageSet(t *testing.T, s string) *reflectv1.PackageSet {
	packageSet, err := unmarshalPackageSet(s)
	require.NoError(t, err)
	return packageSet
}

func unmarshalPackageSet(s string) (*reflectv1.PackageSet, error) {
	packageSet := &reflectv1.PackageSet{}
	if err := jsonpb.UnmarshalString(s, packageSet); err != nil {
		return nil, err
	}
	return packageSet, nil
}
