// Copyright (c) 2021 Uber Technologies, Inc.
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

package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/cmd/testdata/grpc/gen/grpcpb"
	"github.com/uber/prototool/internal/lint"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/vars"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func TestDownload(t *testing.T) {
	assertExact(t, true, false, 0, ``, "cache", "update", "testdata/foo")
	fileInfo, err := os.Stat(filepath.Join("testcache", "protobuf", vars.DefaultProtocVersion))
	assert.NoError(t, err)
	assert.True(t, fileInfo.IsDir())
	fileInfo, err = os.Stat(filepath.Join("testcache", "protobuf", vars.DefaultProtocVersion+".lock"))
	assert.NoError(t, err)
	assert.False(t, fileInfo.IsDir())
}

func TestCompile(t *testing.T) {
	t.Parallel()
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/errors_on_import/dep_errors.proto:6:1:Expected ";".`,
		"testdata/compile/errors_on_import/dep_errors.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/errors_on_import/dep_errors.proto:6:1:Expected ";".`,
		"testdata/compile/errors_on_import",
	)
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/extra_import/extra_import.proto:6:1:Import "dep.proto" was not used.`,
		"testdata/compile/extra_import/extra_import.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/json/json_camel_case_conflict.proto:7:9:The JSON camel-case name of field "helloworld" conflicts with field "helloWorld". This is not allowed in proto3.`,
		"testdata/compile/json/json_camel_case_conflict.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/semicolon/missing_package_semicolon.proto:5:1:Expected ";".`,
		"testdata/compile/semicolon/missing_package_semicolon.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/syntax/missing_syntax.proto:1:1:No syntax specified. Please use 'syntax = "proto2";' or 'syntax = "proto3";' to specify a syntax version.
		testdata/compile/syntax/missing_syntax.proto:4:3:Expected "required", "optional", or "repeated".`,
		"testdata/compile/syntax/missing_syntax.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/recursive/one.proto:5:1:File recursively imports itself one.proto -> two.proto -> one.proto.
		testdata/compile/recursive/one.proto:5:1:Import "two.proto" was not found or had errors.`,
		"testdata/compile/recursive/one.proto",
	)
	assertDoCompileFiles(
		t,
		true,
		false,
		``,
		"testdata/compile/proto2/syntax_proto2.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		false,
		`testdata/compile/notimported/not_imported.proto:11:3:"foo.Dep" seems to be defined in "dep.proto", which is not imported by "not_imported.proto".  To use it here, please add the necessary import.`,
		"testdata/compile/notimported/not_imported.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		true,
		`{"filename":"testdata/compile/errors_on_import/dep_errors.proto","line":6,"column":1,"message":"Expected \";\"."}`,
		"testdata/compile/errors_on_import/dep_errors.proto",
	)
}

func TestInit(t *testing.T) {
	t.Parallel()
	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	require.NotEmpty(t, tmpDir)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	assertDo(t, false, false, 0, "", "config", "init", tmpDir)
	assertDo(t, false, false, 1, fmt.Sprintf("%s already exists", filepath.Join(tmpDir, settings.DefaultConfigFilename)), "config", "init", tmpDir)
}

func TestLint(t *testing.T) {
	t.Parallel()
	assertDoLintFile(
		t,
		true,
		"",
		"testdata/foo/success.proto",
	)
	assertDoLintFile(
		t,
		true,
		"",
		"testdata/lint/version2",
	)
	//assertDoLintFile(
	//t,
	//true,
	//"",
	//"../../etc/style/google",
	//)
	//assertDoLintFile(
	//t,
	//true,
	//"",
	//"../../etc/style/uber1",
	//)
	//assertDoLintFile(
	//t,
	//true,
	//"",
	//"../../style",
	//)
	assertDoLintFile(
		t,
		false,
		"1:1:SYNTAX_PROTO3",
		"testdata/lint/syntaxproto2/syntax_proto2.proto",
	)
	assertDoLintFile(
		t,
		false,
		"11:1:MESSAGE_NAMES_CAPITALIZED",
		"testdata/lint/capitalized/message_name_not_capitalized.proto",
	)
	assertDoLintFile(
		t,
		false,
		`1:1:FILE_OPTIONS_REQUIRE_GO_PACKAGE
		1:1:FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES
		1:1:FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME
		1:1:FILE_OPTIONS_REQUIRE_JAVA_PACKAGE`,
		"testdata/lint/required/file_options_required.proto",
	)
	assertDoLintFile(
		t,
		false,
		`1:1:FILE_OPTIONS_REQUIRE_GO_PACKAGE
		1:1:FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES
		1:1:FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME
		1:1:FILE_OPTIONS_REQUIRE_JAVA_PACKAGE
		1:1:PACKAGE_IS_DECLARED`,
		"testdata/lint/base/base_file.proto",
	)
	assertDoLintFile(
		t,
		false,
		`5:1:FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX
		6:1:FILE_OPTIONS_EQUAL_JAVA_MULTIPLE_FILES_TRUE
		7:1:FILE_OPTIONS_EQUAL_JAVA_OUTER_CLASSNAME_PROTO_SUFFIX
		8:1:FILE_OPTIONS_EQUAL_JAVA_PACKAGE_COM_PREFIX`,
		"testdata/lint/fileoptions/file_options_incorrect.proto",
	)
	assertDoLintFile(
		t,
		false,
		`9:1:FILE_OPTIONS_EQUAL_JAVA_PACKAGE_PREFIX`,
		"testdata/lint/fileoptionsjava/file_options_incorrect.proto",
	)
	assertDoLintFiles(
		t,
		false,
		`testdata/lint/samedir/bar1.proto:1:1:PACKAGES_SAME_IN_DIR
		testdata/lint/samedir/foo1.proto:1:1:PACKAGES_SAME_IN_DIR
		testdata/lint/samedir/foo2.proto:1:1:PACKAGES_SAME_IN_DIR`,
		"testdata/lint/samedir",
	)
	assertDoLintFiles(
		t,
		false,
		`testdata/lint/samedirgopkg/bar1.proto:1:1:FILE_OPTIONS_GO_PACKAGE_SAME_IN_DIR
		testdata/lint/samedirgopkg/foo1.proto:1:1:FILE_OPTIONS_GO_PACKAGE_SAME_IN_DIR
		testdata/lint/samedirgopkg/foo2.proto:1:1:FILE_OPTIONS_GO_PACKAGE_SAME_IN_DIR`,
		"testdata/lint/samedirgopkg",
	)
	assertDoLintFiles(
		t,
		false,
		`testdata/lint/samedirjavapkg/bar1.proto:1:1:FILE_OPTIONS_JAVA_PACKAGE_SAME_IN_DIR
		testdata/lint/samedirjavapkg/foo1.proto:1:1:FILE_OPTIONS_JAVA_PACKAGE_SAME_IN_DIR
		testdata/lint/samedirjavapkg/foo2.proto:1:1:FILE_OPTIONS_JAVA_PACKAGE_SAME_IN_DIR`,
		"testdata/lint/samedirjavapkg",
	)
	assertDoLintFile(
		t,
		false,
		`1:1:FILE_OPTIONS_REQUIRE_GO_PACKAGE
		1:1:FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES
		1:1:FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME
		1:1:FILE_OPTIONS_REQUIRE_JAVA_PACKAGE
		3:1:PACKAGE_LOWER_SNAKE_CASE
		7:1:MESSAGE_NAMES_CAPITALIZED
		9:1:MESSAGE_NAMES_CAMEL_CASE
		12:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		13:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		14:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		15:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		22:3:COMMENTS_NO_C_STYLE
		23:3:COMMENTS_NO_C_STYLE
		26:1:SERVICE_NAMES_CAPITALIZED
		28:1:SERVICE_NAMES_CAMEL_CASE
		46:3:REQUEST_RESPONSE_TYPES_UNIQUE
		46:3:REQUEST_RESPONSE_TYPES_UNIQUE
		47:3:REQUEST_RESPONSE_TYPES_UNIQUE
		47:3:REQUEST_RESPONSE_TYPES_UNIQUE
		48:3:RPC_NAMES_CAPITALIZED
		49:3:REQUEST_RESPONSE_TYPES_IN_SAME_FILE
		49:3:REQUEST_RESPONSE_TYPES_UNIQUE
		50:3:REQUEST_RESPONSE_TYPES_IN_SAME_FILE
		50:3:REQUEST_RESPONSE_TYPES_IN_SAME_FILE
		50:3:REQUEST_RESPONSE_TYPES_UNIQUE
		58:3:ENUM_FIELD_PREFIXES
		64:7:ENUM_FIELD_PREFIXES
		64:7:ENUM_ZERO_VALUES_INVALID
		67:7:ENUM_ZERO_VALUES_INVALID
		73:3:ENUM_ZERO_VALUES_INVALID
		76:1:COMMENTS_NO_C_STYLE
		80:3:COMMENTS_NO_C_STYLE
		82:3:COMMENTS_NO_C_STYLE
		84:5:COMMENTS_NO_C_STYLE
		90:3:ENUM_FIELD_NAMES_UPPER_SNAKE_CASE
		93:1:ENUM_NAMES_CAMEL_CASE
		98:3:ENUMS_NO_ALLOW_ALIAS
		108:5:ENUMS_NO_ALLOW_ALIAS
		`,
		"testdata/lint/lots/lots.proto",
	)
	assertDoLintFile(
		t,
		false,
		`1:1:FILE_OPTIONS_REQUIRE_GO_PACKAGE
		1:1:FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES
		1:1:FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME
		1:1:FILE_OPTIONS_REQUIRE_JAVA_PACKAGE
		3:1:PACKAGE_LOWER_SNAKE_CASE
		7:1:MESSAGES_HAVE_COMMENTS
		7:1:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		7:1:MESSAGE_NAMES_CAPITALIZED
		9:1:MESSAGES_HAVE_COMMENTS
		9:1:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		9:1:MESSAGE_NAMES_CAMEL_CASE
		11:1:MESSAGES_HAVE_COMMENTS
		11:1:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		12:3:MESSAGE_FIELD_NAMES_LOWERCASE
		12:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		13:3:MESSAGE_FIELD_NAMES_LOWERCASE
		13:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		14:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		15:3:MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
		22:3:COMMENTS_NO_C_STYLE
		23:3:COMMENTS_NO_C_STYLE
		26:1:SERVICES_HAVE_COMMENTS
		26:1:SERVICE_NAMES_CAPITALIZED
		28:1:SERVICES_HAVE_COMMENTS
		28:1:SERVICE_NAMES_CAMEL_CASE
		30:1:MESSAGES_HAVE_COMMENTS
		31:1:MESSAGES_HAVE_COMMENTS
		34:1:MESSAGES_HAVE_COMMENTS
		36:5:MESSAGES_HAVE_COMMENTS
		36:5:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		38:1:MESSAGES_HAVE_COMMENTS
		40:1:MESSAGES_HAVE_COMMENTS
		40:1:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		41:3:MESSAGES_HAVE_COMMENTS
		44:1:SERVICES_HAVE_COMMENTS
		45:3:RPCS_HAVE_COMMENTS
		46:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		46:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		46:3:REQUEST_RESPONSE_TYPES_UNIQUE
		46:3:REQUEST_RESPONSE_TYPES_UNIQUE
		46:3:RPCS_HAVE_COMMENTS
		47:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		47:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		47:3:REQUEST_RESPONSE_TYPES_UNIQUE
		47:3:REQUEST_RESPONSE_TYPES_UNIQUE
		47:3:RPCS_HAVE_COMMENTS
		48:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		48:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		48:3:RPCS_HAVE_COMMENTS
		48:3:RPC_NAMES_CAPITALIZED
		49:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		49:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		49:3:REQUEST_RESPONSE_TYPES_IN_SAME_FILE
		49:3:REQUEST_RESPONSE_TYPES_UNIQUE
		49:3:RPCS_HAVE_COMMENTS
		50:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		50:3:REQUEST_RESPONSE_NAMES_MATCH_RPC
		50:3:REQUEST_RESPONSE_TYPES_IN_SAME_FILE
		50:3:REQUEST_RESPONSE_TYPES_IN_SAME_FILE
		50:3:REQUEST_RESPONSE_TYPES_UNIQUE
		50:3:RPCS_HAVE_COMMENTS
		53:1:ENUMS_HAVE_COMMENTS
		58:3:ENUM_FIELD_PREFIXES
		61:1:MESSAGES_HAVE_COMMENTS
		61:1:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		62:3:MESSAGES_HAVE_COMMENTS
		62:3:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		63:5:ENUMS_HAVE_COMMENTS
		64:7:ENUM_FIELD_PREFIXES
		64:7:ENUM_ZERO_VALUES_INVALID
		66:5:ENUMS_HAVE_COMMENTS
		67:7:ENUM_ZERO_VALUES_INVALID
		72:1:ENUMS_HAVE_COMMENTS
		73:3:ENUM_ZERO_VALUES_INVALID
		76:1:COMMENTS_NO_C_STYLE
		78:1:MESSAGES_HAVE_COMMENTS
		78:1:MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES
		80:3:COMMENTS_NO_C_STYLE
		82:3:COMMENTS_NO_C_STYLE
		84:5:COMMENTS_NO_C_STYLE
		88:1:ENUMS_HAVE_COMMENTS
		90:3:ENUM_FIELD_NAMES_UPPERCASE
		90:3:ENUM_FIELD_NAMES_UPPER_SNAKE_CASE
		93:1:ENUMS_HAVE_COMMENTS
		93:1:ENUM_NAMES_CAMEL_CASE`,
		"testdata/lint/allgroup/lots.proto",
	)
	assertDoLintFile(
		t,
		false,
		`1:1:FILE_OPTIONS_REQUIRE_GO_PACKAGE
		1:1:FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES
		1:1:FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME
		1:1:FILE_OPTIONS_REQUIRE_JAVA_PACKAGE`,
		"testdata/lint/keyword/package_starts_with_keyword.proto",
	)
	assertDoLintFile(
		t,
		false,
		`5:1:FILE_OPTIONS_GO_PACKAGE_NOT_LONG_FORM`,
		"testdata/lint/gopackagelongform/gopackagelongform.proto",
	)
	assertDoLintFile(
		t,
		false,
		`11:3:MESSAGE_FIELDS_NO_JSON_NAME
		12:12:MESSAGE_FIELDS_NO_JSON_NAME
		13:3:MESSAGE_FIELDS_NO_JSON_NAME
		15:5:MESSAGE_FIELDS_NO_JSON_NAME
		16:5:MESSAGE_FIELDS_NO_JSON_NAME`,
		"testdata/lint/nojsonname/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`11:3:FIELDS_NOT_RESERVED
		12:3:FIELDS_NOT_RESERVED
		14:5:FIELDS_NOT_RESERVED
		15:5:FIELDS_NOT_RESERVED
		20:5:FIELDS_NOT_RESERVED`,
		"testdata/lint/noreserved/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`10:1:SERVICE_NAMES_API_SUFFIX
		12:1:SERVICE_NAMES_API_SUFFIX`,
		"testdata/lint/apisuffix/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`13:1:REQUEST_RESPONSE_TYPES_AFTER_SERVICE`,
		"testdata/lint/afterservice/foo/v1/hello_api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`13:1:REQUEST_RESPONSE_TYPES_AFTER_SERVICE
		15:1:REQUEST_RESPONSE_TYPES_AFTER_SERVICE`,
		"testdata/lint/afterservice/foo/v1/hello2api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`19:1:REQUEST_RESPONSE_TYPES_AFTER_SERVICE`,
		"testdata/lint/afterservice/foo/v1/hello3api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`1:20:COMMENTS_NO_INLINE
		3:17:COMMENTS_NO_INLINE
		6:30:COMMENTS_NO_INLINE
		15:23:COMMENTS_NO_INLINE
		21:21:COMMENTS_NO_INLINE
		30:25:COMMENTS_NO_INLINE
		36:20:COMMENTS_NO_INLINE
		37:37:COMMENTS_NO_INLINE
		38:37:COMMENTS_NO_INLINE
		41:23:COMMENTS_NO_INLINE
		47:18:COMMENTS_NO_INLINE
		48:35:COMMENTS_NO_INLINE
		49:35:COMMENTS_NO_INLINE`,
		"testdata/lint/inlinecomments/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`13:28:COMMENTS_NO_INLINE
		16:58:COMMENTS_NO_INLINE`,
		"testdata/lint/inlinecomments/foo/v1/hello_api.proto",
	)
	assertDoLintFile(
		t,
		false,
		`13:1:NAMES_NO_COMMON
		14:3:NAMES_NO_COMMON
		17:1:NAMES_NO_DATA
		18:3:NAMES_NO_DATA
		21:1:NAMES_NO_UUID
		22:3:NAMES_NO_UUID
		25:1:NAMES_NO_COMMON
		26:3:NAMES_NO_UUID
		27:5:NAMES_NO_UUID
		28:7:NAMES_NO_UUID
		35:1:NAMES_NO_COMMON
		35:1:NAMES_NO_DATA
		39:1:NAMES_NO_UUID
		40:3:NAMES_NO_DATA
		42:3:NAMES_NO_DATA
		43:5:NAMES_NO_COMMON
		47:1:NAMES_NO_COMMON
		48:3:NAMES_NO_COMMON
		51:1:NAMES_NO_DATA
		52:3:NAMES_NO_COMMON
		55:1:NAMES_NO_UUID
		56:3:NAMES_NO_COMMON`,
		"testdata/lint/naming/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`13:1:MESSAGES_NOT_EMPTY_EXCEPT_REQUEST_RESPONSE_TYPES
		14:3:MESSAGES_NOT_EMPTY_EXCEPT_REQUEST_RESPONSE_TYPES
		15:5:MESSAGES_NOT_EMPTY_EXCEPT_REQUEST_RESPONSE_TYPES`,
		"testdata/lint/notempty/foo/v1/hello.proto",
	)
	assertDoLintFile(
		t,
		false,
		``,
		"testdata/lint/notempty/foo/v1/hello_api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`14:3:MESSAGE_FIELD_NAMES_NO_DESCRIPTOR
		18:3:MESSAGE_FIELD_NAMES_NO_DESCRIPTOR
		22:3:MESSAGE_FIELD_NAMES_NO_DESCRIPTOR
		26:3:MESSAGE_FIELD_NAMES_NO_DESCRIPTOR
		30:3:MESSAGE_FIELD_NAMES_NO_DESCRIPTOR
		34:3:MESSAGE_FIELD_NAMES_NO_DESCRIPTOR`,
		"testdata/lint/nodescriptor/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`16:7:MESSAGE_FIELD_NAMES_FILENAME
		18:9:MESSAGE_FIELD_NAMES_FILENAME
		21:5:MESSAGE_FIELD_NAMES_FILENAME
		23:3:MESSAGE_FIELD_NAMES_FILENAME
		24:3:MESSAGE_FIELD_NAMES_FILENAME
		30:7:MESSAGE_FIELD_NAMES_FILEPATH
		32:9:MESSAGE_FIELD_NAMES_FILEPATH
		35:5:MESSAGE_FIELD_NAMES_FILEPATH
		37:3:MESSAGE_FIELD_NAMES_FILEPATH
		38:3:MESSAGE_FIELD_NAMES_FILEPATH`,
		"testdata/lint/fieldnamesfilename/hello.proto",
	)

	assertDoLintFile(
		t,
		false,
		`16:7:MESSAGE_FIELDS_TIME
		18:9:MESSAGE_FIELDS_TIME
		21:5:MESSAGE_FIELDS_TIME
		23:3:MESSAGE_FIELDS_TIME
		30:7:MESSAGE_FIELDS_DURATION
		32:9:MESSAGE_FIELDS_DURATION
		35:5:MESSAGE_FIELDS_DURATION
		37:3:MESSAGE_FIELDS_DURATION`,
		"testdata/lint/fieldstimeduration/hello.proto",
	)

	assertDoLintFile(
		t,
		false,
		`17:5:RPC_OPTIONS_NO_GOOGLE_API_HTTP
		22:5:RPC_OPTIONS_NO_GOOGLE_API_HTTP
		25:5:RPC_OPTIONS_NO_GOOGLE_API_HTTP
		30:5:RPC_OPTIONS_NO_GOOGLE_API_HTTP`,
		"testdata/lint/nogoogleapihttp/foo/v1/hello_api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`15:3:RPCS_NO_STREAMING
		16:3:RPCS_NO_STREAMING
		17:3:RPCS_NO_STREAMING`,
		"testdata/lint/nostreaming/foo/v1/hello_api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`10:1:GOGO_NOT_IMPORTED`,
		"testdata/lint/gogonotimported/gogonotimported.proto",
	)

	assertDoLintFile(
		t,
		false,
		`10:1:IMPORTS_NOT_PUBLIC`,
		"testdata/lint/importsnotpublic/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`10:1:IMPORTS_NOT_WEAK`,
		"testdata/lint/importsnotweak/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`3:1:PACKAGE_MAJOR_BETA_VERSIONED`,
		"testdata/lint/majorbetaversioned/foo/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`20:3:MESSAGE_FIELDS_NOT_FLOATS
		21:3:MESSAGE_FIELDS_NOT_FLOATS`,
		"testdata/lint/floats/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`17:3:MESSAGE_FIELDS_NOT_FLOATS
		19:3:MESSAGE_FIELDS_NOT_FLOATS
		20:3:MESSAGE_FIELDS_NOT_FLOATS
		21:3:MESSAGE_FIELDS_NOT_FLOATS`,
		"testdata/lint/floatsnosuppress/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`14:3:MESSAGE_FIELDS_NOT_FLOATS`,
		"testdata/lint/ignoredir/foo/v1/foo.proto",
	)
	assertDoLintFile(
		t,
		true,
		``,
		"testdata/lint/ignoredir/bar/v1/bar.proto",
	)

	assertDoLintFile(
		t,
		false,
		`23:1:REQUEST_RESPONSE_TYPES_ONLY_IN_FILE
		30:1:REQUEST_RESPONSE_TYPES_ONLY_IN_FILE`,
		"testdata/lint/onlyinfile/foo/v1/hello_api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`13:1:SERVICE_NAMES_MATCH_FILE_NAME
		15:1:SERVICE_NAMES_MATCH_FILE_NAME`,
		"testdata/lint/servicenamefilename/foo/v1/another_api.proto",
	)

	assertDoLintFile(
		t,
		false,
		`13:1:SERVICE_NAMES_MATCH_FILE_NAME`,
		"testdata/lint/servicenamefilename/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`15:1:SERVICE_NAMES_NO_PLURALS
		19:1:SERVICE_NAMES_NO_PLURALS
		23:1:SERVICE_NAMES_NO_PLURALS
		25:1:SERVICE_NAMES_NO_PLURALS
		25:1:SERVICE_NAMES_NO_PLURALS
		45:1:SERVICE_NAMES_NO_PLURALS`,
		"testdata/lint/noplurals/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`3:1:PACKAGE_NO_KEYWORDS`,
		"testdata/lint/nokeywords/foo/public/public.proto",
	)
	assertDoLintFile(
		t,
		false,
		``,
		"testdata/lint/nokeywords/foo/public/public_suppressed.proto",
	)

	assertDoLintFile(
		t,
		false,
		`3:1:PACKAGE_LOWER_CASE`,
		"testdata/lint/packagelowercase/foo_bar/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`1:1:FILE_NAMES_LOWER_SNAKE_CASE`,
		"testdata/lint/filename/fileNameInvalid.proto",
	)

	assertDoLintFile(
		t,
		false,
		`14:3:WKT_TIMESTAMP_SUFFIX
		15:3:WKT_DURATION_SUFFIX
		17:5:WKT_TIMESTAMP_SUFFIX
		18:5:WKT_DURATION_SUFFIX`,
		"testdata/lint/wktsuffix/foo/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`16:3:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		16:3:ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE
		17:3:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		21:3:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		21:3:ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE
		31:5:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		31:5:ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE
		32:5:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		36:5:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		36:5:ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE
		46:7:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		46:7:ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE
		47:7:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		51:7:ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE
		51:7:ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE`,
		"testdata/lint/enumexceptmessages/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`15:3:ENUM_FIELDS_HAVE_COMMENTS
		24:7:ENUM_FIELDS_HAVE_COMMENTS
		26:5:MESSAGE_FIELDS_HAVE_COMMENTS
		27:5:MESSAGE_FIELDS_HAVE_COMMENTS
		29:7:MESSAGE_FIELDS_HAVE_COMMENTS
		34:5:ENUM_FIELDS_HAVE_COMMENTS
		36:3:MESSAGE_FIELDS_HAVE_COMMENTS
		37:3:MESSAGE_FIELDS_HAVE_COMMENTS
		39:5:MESSAGE_FIELDS_HAVE_COMMENTS`,
		"testdata/lint/fieldscomments/foo/v1/foo.proto",
	)

	assertDoLintFile(
		t,
		false,
		`15:3:ENUM_FIELDS_HAVE_SENTENCE_COMMENTS
		24:7:ENUM_FIELDS_HAVE_SENTENCE_COMMENTS
		26:5:MESSAGE_FIELDS_HAVE_SENTENCE_COMMENTS
		27:5:MESSAGE_FIELDS_HAVE_SENTENCE_COMMENTS
		29:7:MESSAGE_FIELDS_HAVE_SENTENCE_COMMENTS
		34:5:ENUM_FIELDS_HAVE_SENTENCE_COMMENTS
		36:3:MESSAGE_FIELDS_HAVE_SENTENCE_COMMENTS
		37:3:MESSAGE_FIELDS_HAVE_SENTENCE_COMMENTS
		39:5:MESSAGE_FIELDS_HAVE_SENTENCE_COMMENTS`,
		"testdata/lint/fieldssentencecomments/foo/v1/foo.proto",
	)
}

func TestLintConfigDataOverride(t *testing.T) {
	assertDoLintFile(
		t,
		false,
		`5:1:FILE_OPTIONS_GO_PACKAGE_NOT_LONG_FORM`,
		"testdata/lint/gopackagelongform/gopackagelongform.proto",
		"--config-data",
		`{"lint":{"rules":{"remove":["FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX"]}}}`,
	)
	assertDoLintFile(
		t,
		false,
		`5:1:FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX`,
		"testdata/lint/gopackagelongform/gopackagelongform.proto",
		"--config-data",
		`{"lint":{"rules":{"remove":["FILE_OPTIONS_GO_PACKAGE_NOT_LONG_FORM"]}}}`,
	)
	assertDoLintFile(
		t,
		false,
		`5:1:FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX
		5:1:FILE_OPTIONS_GO_PACKAGE_NOT_LONG_FORM`,
		"testdata/lint/gopackagelongform/gopackagelongform.proto",
		"--config-data",
		`{}`,
	)
	assertExact(
		t,
		true,
		true,
		1,
		`json: unknown field "unknown_key"`,
		"lint",
		"testdata/lint/gopackagelongform/gopackagelongform.proto",
		"--config-data",
		`{"unknown_key":"foo"}`,
	)
}

func TestGoldenFormat(t *testing.T) {
	t.Parallel()
	assertGoldenFormat(t, false, false, "testdata/format/proto3/foo/bar/bar.proto")
	assertGoldenFormat(t, false, false, "testdata/format/proto2/foo/bar/bar_proto2.proto")
	assertGoldenFormat(t, false, false, "testdata/format/proto3/foo/foo.proto")
	assertGoldenFormat(t, false, false, "testdata/format/proto2/foo/foo_proto2.proto")
	assertGoldenFormat(t, false, true, "testdata/format-fix/foo.proto")
	assertGoldenFormat(t, false, true, "testdata/format-fix-v2/foo.proto")
}

func TestCreate(t *testing.T) {
	t.Parallel()
	// package override with also matching shorter override "a"
	// make sure uses "a/b"
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/one/a/b/bar/baz.proto",
		"",
		`syntax = "proto3";

package foo.bar;

option go_package = "barpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foo.bar";`,
	)
	// create same file again but do not remove, should fail
	assertDoCreateFile(
		t,
		false, // do not expect success
		false, // do not remove
		"testdata/create/one/a/b/bar/baz.proto",
		"",
		``,
	)
	// use the --package flag
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/one/a/b/bar/baz.proto",
		"bat", // --package value
		`syntax = "proto3";

package bat;

option go_package = "batpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.bat";`,
	)
	// package override but a shorter one "a"
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/one/a/c/bar/baz.proto",
		"",
		`syntax = "proto3";

package foobar.c.bar;

option go_package = "barpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foobar.c.bar";`,
	)
	// no package override, do default b.c.bar
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/one/b/c/bar/baz.proto",
		"",
		`syntax = "proto3";

package b.c.bar;

option go_package = "barpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.b.c.bar";`,
	)
	// in dir with prototool.yaml, use default package
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/one/baz.proto",
		"",
		`syntax = "proto3";

package uber.prototool.generated;

option go_package = "generatedpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.uber.prototool.generated";`,
	)
	// in dir with prototool.yaml with override
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/two/baz.proto",
		"",
		`// this
// is a
// header

syntax = "proto3";

package foo;

option go_package = "foopb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foo";`,
	)
}

func TestCreateV2(t *testing.T) {
	t.Parallel()
	// package override with also matching shorter override "a"
	// make sure uses "a/b"
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2one/a/b/bar/baz.proto",
		"",
		`syntax = "proto3";

package foo.bar;

option csharp_namespace = "Foo.Bar";
option go_package = "barpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foo.bar";
option objc_class_prefix = "FBX";
option php_namespace = "Foo\\Bar";`,
	)
	// create same file again but do not remove, should fail
	assertDoCreateFile(
		t,
		false, // do not expect success
		false, // do not remove
		"testdata/create/version2one/a/b/bar/baz.proto",
		"",
		``,
	)
	// use the --package flag
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2one/a/b/bar/baz.proto",
		"bat", // --package value
		`syntax = "proto3";

package bat;

option csharp_namespace = "Bat";
option go_package = "batpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.bat";
option objc_class_prefix = "BXX";
option php_namespace = "Bat";`,
	)
	// package override but a shorter one "a"
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2one/a/c/bar/baz.proto",
		"",
		`syntax = "proto3";

package foobar.c.bar;

option csharp_namespace = "Foobar.C.Bar";
option go_package = "barpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foobar.c.bar";
option objc_class_prefix = "FCB";
option php_namespace = "Foobar\\C\\Bar";`,
	)
	// no package override, do default b.c.bar
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2one/b/c/bar/baz.proto",
		"",
		`syntax = "proto3";

package b.c.bar;

option csharp_namespace = "B.C.Bar";
option go_package = "barpb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.b.c.bar";
option objc_class_prefix = "BCB";
option php_namespace = "B\\C\\Bar";`,
	)
	// in dir with prototool.yaml with override
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2two/baz.proto",
		"",
		`// this
// is a
// header

syntax = "proto3";

package foo;

option csharp_namespace = "Foo";
option go_package = "foopb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foo";
option objc_class_prefix = "FXX";
option php_namespace = "Foo";`,
	)
}

func TestCreateV2MajorBetaVersion(t *testing.T) {
	t.Parallel()
	// package override with also matching shorter override "a"
	// make sure uses "a/b"
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2three/a/b/bar/v1/baz.proto",
		"",
		`syntax = "proto3";

package foo.bar.v1;

option csharp_namespace = "Foo.Bar.V1";
option go_package = "barv1";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foo.bar.v1";
option objc_class_prefix = "FBX";
option php_namespace = "Foo\\Bar\\V1";`,
	)
	// create same file again but do not remove, should fail
	assertDoCreateFile(
		t,
		false, // do not expect success
		false, // do not remove
		"testdata/create/version2three/a/b/bar/v1/baz.proto",
		"",
		``,
	)
	// package override but a shorter one "a"
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2three/a/c/bar/v1/baz.proto",
		"",
		`syntax = "proto3";

package foobar.c.bar.v1;

option csharp_namespace = "Foobar.C.Bar.V1";
option go_package = "barv1";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foobar.c.bar.v1";
option objc_class_prefix = "FCB";
option php_namespace = "Foobar\\C\\Bar\\V1";`,
	)
	// no package override, do default b.c.bar
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2three/b/c/bar/v1beta1/baz.proto",
		"",
		`syntax = "proto3";

package b.c.bar.v1beta1;

option csharp_namespace = "B.C.Bar.V1Beta1";
option go_package = "barv1beta1";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.b.c.bar.v1beta1";
option objc_class_prefix = "BCB";
option php_namespace = "B\\C\\Bar\\V1Beta1";`,
	)
	// in dir with prototool.yaml, use default package
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2three/baz.proto",
		"",
		`syntax = "proto3";

package uber.prototool.generated.v1;

option csharp_namespace = "Uber.Prototool.Generated.V1";
option go_package = "generatedv1";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.uber.prototool.generated.v1";
option objc_class_prefix = "UPG";
option php_namespace = "Uber\\Prototool\\Generated\\V1";`,
	)
	// in dir with prototool.yaml with override
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2four/baz.proto",
		"",
		`syntax = "proto3";

package foo.v1;

option csharp_namespace = "Foo.V1";
option go_package = "foov1";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foo.v1";
option objc_class_prefix = "FXX";
option php_namespace = "Foo\\V1";`,
	)
	// in dir with prototool.yaml with override with java_package_prefix
	assertDoCreateFile(
		t,
		true,
		true,
		"testdata/create/version2five/baz.proto",
		"",
		`syntax = "proto3";

package foo.v1;

option csharp_namespace = "Foo.V1";
option go_package = "foov1";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "au.com.foo.v1";
option objc_class_prefix = "FXX";
option php_namespace = "Foo\\V1";`,
	)
}

func TestGRPC(t *testing.T) {
	t.Parallel()
	const (
		serverCrt         = "testdata/grpc/tls/server.crt"
		serverKey         = "testdata/grpc/tls/server.key"
		clientCrt         = "testdata/grpc/tls/client.crt"
		clientKey         = "testdata/grpc/tls/client.key"
		caCrt             = "testdata/grpc/tls/cacert.crt"
		ssServerCrt       = "testdata/grpc/tls/self-signed-server.crt"
		ssServerKey       = "testdata/grpc/tls/self-signed-server.key"
		ssClientCrt       = "testdata/grpc/tls/self-signed-client.crt"
		ssClientKey       = "testdata/grpc/tls/self-signed-client.key"
		helloJSONValue    = `{"value":"hello"}`
		helloExclaimValue = `{"value":"hello!"}`
		grpcFilePath      = "testdata/grpc/grpc.proto"
		exclamationMethod = "grpc.ExcitedService/Exclamation"
	)
	assertGRPC(t,
		0,
		helloExclaimValue,
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
	)
	assertGRPC(t,
		0,
		`{"value":"hellosalutations!"}`,
		grpcFilePath,
		"grpc.ExcitedService/ExclamationClientStream",
		`{"value":"hello"}
		{"value":"salutations"}`,
	)
	assertGRPC(t,
		0,
		`{"value":"h"}
		{"value":"e"}
		{"value":"l"}
		{"value":"l"}
		{"value":"o"}
		{"value":"!"}`,
		grpcFilePath,
		"grpc.ExcitedService/ExclamationServerStream",
		helloJSONValue,
	)
	assertGRPC(t,
		0,
		`
		{"value":"hello!"}
		{"value":"salutations!"}
		`,
		grpcFilePath,
		"grpc.ExcitedService/ExclamationBidiStream",
		`{"value":"hello"}
		{"value":"salutations"}`,
	)
	assertGRPC(t,
		0,
		`{"headers":{"content-type":["application/grpc"]}}
		{"response":{"value":"h"}}
		{"response":{"value":"e"}}
		{"response":{"value":"l"}}
		{"response":{"value":"l"}}
		{"response":{"value":"o"}}
		{"response":{"value":"!"}}`,
		grpcFilePath,
		"grpc.ExcitedService/ExclamationServerStream",
		helloJSONValue,
		`--details`,
	)
	assertGRPC(t,
		255,
		"tls must be specified if insecure, cacert, cert, key or server-name are specified",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		"--cert",
		clientCrt,
	)
	assertGRPC(t,
		255,
		"if cert is specified, key must be specified",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		"--tls",
		"--cert",
		clientCrt,
		"--cacert",
		caCrt,
	)
	assertGRPC(t,
		255,
		"if insecure then cacert, cert, key, and server-name must not be specified",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		"--tls",
		"--cert",
		clientCrt,
		"--insecure",
	)
	// CA issued server certificate valid when using ca cert as server cert
	assertGRPCTLS(t,
		0,
		helloExclaimValue,
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		serverCrt,
		serverKey,
		caCrt,
	)
	// Self signed server certificate valid when using self-signed server cert
	assertGRPCTLS(t,
		0,
		helloExclaimValue,
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		ssServerCrt,
		ssServerKey,
		ssServerCrt,
	)
	// Self signed server certificate valid when using system CA and insecure flag
	assertGRPCTLS(t,
		0,
		helloExclaimValue,
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		ssServerCrt,
		ssServerKey,
		"",
		"--tls",
		"--insecure",
	)
	// server uses plaintext but client attempts a TLS connection
	assertGRPC(t,
		1,
		"tls: first record does not look like a TLS handshake",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		"--tls",
	)
	// Self signed server certificate invalid when using ca cert as server cert
	assertGRPCTLS(t,
		1,
		"x509: certificate signed by unknown authority",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		ssServerCrt,
		ssServerKey,
		caCrt,
	)
	// Server uses TLS but client does not
	assertGRPCTLS(t,
		1,
		"context deadline exceeded",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		serverCrt,
		serverKey,
		"",
	)
	// Mututal TLS with CA issued server certificate and self-signed client certificate valid when using ca cert as server cert
	assertGRPCmTLS(t,
		0,
		helloExclaimValue,
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		serverCrt,
		serverKey,
		caCrt,
		clientCrt,
		clientKey,
		caCrt,
	)
	// Mututal TLS with self-signed server and client certificates properly exchanged
	assertGRPCmTLS(t,
		0,
		helloExclaimValue,
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		ssServerCrt,
		ssServerKey,
		ssServerCrt,
		ssClientCrt,
		ssClientKey,
		ssClientCrt,
	)

	// Mututal TLS with CA issued server certificate but using self-signed client certificate NOT valid when using ca cert as client cert
	// assertGRPCmTLS(t,
	// 	1,
	// 	"remote error: tls: bad certificate",
	// 	grpcFilePath,
	// 	exclamationMethod,
	// 	helloJSONValue,
	// 	serverCrt,
	// 	serverKey,
	// 	caCrt,
	// 	ssClientCrt,
	// 	ssClientKey,
	// 	caCrt,
	// )
	// Mututal TLS with CA issued client certificate and self-signed server certificate NOT valid when using server cert CA as client CA
	// assertGRPCmTLS(t,
	// 	1,
	// 	"remote error: tls: bad certificate",
	// 	grpcFilePath,
	// 	exclamationMethod,
	// 	helloJSONValue,
	// 	ssServerCrt,
	// 	ssServerKey,
	// 	ssServerCrt,
	// 	clientCrt,
	// 	clientKey,
	// 	ssServerCrt,
	// )
	// Mututal TLS with self-signed certificate and self-signed client certificate NOT valid when using ca cert as server cert
	assertGRPCmTLS(t,
		1,
		"x509: certificate signed by unknown authority",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		ssServerCrt,
		ssServerKey,
		caCrt,
		ssClientCrt,
		ssClientKey,
		ssClientCrt,
	)
	// server uses mutual TLS but client does not use TLS at all
	assertGRPCmTLS(t,
		1,
		"context deadline exceeded",
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		ssServerCrt,
		ssServerKey,
		"",
		"",
		"",
		ssClientCrt,
	)
	// server uses mutual TLS but client does not use mutual TLS
	// assertGRPCmTLS(t,
	// 	1,
	// 	"remote error: tls: bad certificate",
	// 	grpcFilePath,
	// 	exclamationMethod,
	// 	helloJSONValue,
	// 	ssServerCrt,
	// 	ssServerKey,
	// 	ssServerCrt,
	// 	"",
	// 	"",
	// 	ssClientCrt,
	// )

	assertGRPCmTLS(t,
		0,
		helloExclaimValue,
		grpcFilePath,
		exclamationMethod,
		helloJSONValue,
		ssClientCrt,
		ssClientKey,
		ssClientCrt,
		ssClientCrt,
		ssClientKey,
		ssClientCrt,
		"--server-name", "client.local",
	)

	assertGRPCExclamationError(t,
		errors.New("test"),
		1,
		`{"status":{"code":2,"message":"test"}}
		{"trailers":{"content-type":["application/grpc"]}}
		rpc error: code = Unknown desc = test`,
		"testdata/grpc/grpc.proto",
		"grpc.ExcitedService/Exclamation",
		`{"value":"hello"}`,
		`--details`,
	)

	st, err := status.New(codes.InvalidArgument, "test").WithDetails(&grpcpb.Foo{Bar: "baz"})
	assert.NoError(t, err)
	assertGRPCExclamationError(t,
		status.ErrorProto(st.Proto()),
		1,
		`{"status":{"code":3,"message":"test","details":[{"@type":"type.googleapis.com/grpc.Foo","bar":"baz"}]}}
		{"trailers":{"content-type":["application/grpc"]}}
		rpc error: code = InvalidArgument desc = test`,
		"testdata/grpc/grpc.proto",
		"grpc.ExcitedService/Exclamation",
		`{"value":"hello"}`,
		`--details`,
	)
}

func TestVersion(t *testing.T) {
	t.Parallel()
	assertRegexp(t, false, false, 0, fmt.Sprintf("Version:.*%s\nDefault protoc version:.*%s\n", vars.Version, vars.DefaultProtocVersion), "version")
}

func TestVersionJSON(t *testing.T) {
	t.Parallel()
	assertRegexp(t, false, false, 0, fmt.Sprintf(`(?s){.*"version":.*"%s",.*"default_protoc_version":.*"%s".*}`, vars.Version, vars.DefaultProtocVersion), "version", "--json")
}

func TestDescriptorSet(t *testing.T) {
	t.Parallel()
	for _, includeSourceInfo := range []bool{false, true} {
		assertDescriptorSet(
			t,
			true,
			"testdata/foo",
			false,
			includeSourceInfo,
			"success.proto",
			"bar/dep.proto",
		)
		assertDescriptorSet(
			t,
			true,
			"testdata/foo/bar",
			false,
			includeSourceInfo,
			"bar/dep.proto",
		)
		assertDescriptorSet(
			t,
			true,
			"testdata/foo",
			true,
			includeSourceInfo,
			"success.proto",
			"bar/dep.proto",
			"google/protobuf/timestamp.proto",
		)
		assertDescriptorSet(
			t,
			true,
			"testdata/foo/bar",
			true,
			includeSourceInfo,
			"bar/dep.proto",
		)
	}
}

func TestInspectPackages(t *testing.T) {
	t.Parallel()
	assertExact(
		t,
		true,
		true,
		0,
		`bar
foo
google.protobuf`,
		"x", "inspect", "packages", "testdata/foo",
	)
}

func TestInspectPackageDeps(t *testing.T) {
	t.Parallel()
	assertExact(
		t,
		true,
		true,
		0,
		`bar
google.protobuf`,
		"x", "inspect", "package-deps", "testdata/foo", "--name", "foo",
	)
	assertExact(
		t,
		true,
		true,
		0,
		``,
		"x", "inspect", "package-deps", "testdata/foo", "--name", "bar",
	)
	assertExact(
		t,
		true,
		true,
		0,
		``,
		"x", "inspect", "package-deps", "testdata/foo", "--name", "google.protobuf",
	)
}

func TestInspectPackageImporters(t *testing.T) {
	t.Parallel()
	assertExact(
		t,
		true,
		true,
		0,
		``,
		"x", "inspect", "package-importers", "testdata/foo", "--name", "foo",
	)
	assertExact(
		t,
		true,
		true,
		0,
		`foo`,
		"x", "inspect", "package-importers", "testdata/foo", "--name", "bar",
	)
	assertExact(
		t,
		true,
		true,
		0,
		`foo`,
		"x", "inspect", "package-importers", "testdata/foo", "--name", "google.protobuf",
	)
}

func TestGenerateIgnores(t *testing.T) {
	t.Parallel()
	assertExact(
		t,
		true,
		true,
		0,
		`lint:
  ignores:
  - id: COMMENTS_NO_C_STYLE
    files:
    - lots.proto
  - id: ENUMS_NO_ALLOW_ALIAS
    files:
    - lots.proto
  - id: ENUM_FIELD_NAMES_UPPER_SNAKE_CASE
    files:
    - lots.proto
  - id: ENUM_FIELD_PREFIXES
    files:
    - lots.proto
  - id: ENUM_NAMES_CAMEL_CASE
    files:
    - lots.proto
  - id: ENUM_ZERO_VALUES_INVALID
    files:
    - lots.proto
  - id: FILE_OPTIONS_EQUAL_JAVA_OUTER_CLASSNAME_PROTO_SUFFIX
    files:
    - bar/dep.proto
  - id: FILE_OPTIONS_REQUIRE_GO_PACKAGE
    files:
    - lots.proto
  - id: FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES
    files:
    - lots.proto
  - id: FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME
    files:
    - lots.proto
  - id: FILE_OPTIONS_REQUIRE_JAVA_PACKAGE
    files:
    - lots.proto
  - id: MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE
    files:
    - lots.proto
  - id: MESSAGE_NAMES_CAMEL_CASE
    files:
    - lots.proto
  - id: MESSAGE_NAMES_CAPITALIZED
    files:
    - lots.proto
  - id: PACKAGE_LOWER_SNAKE_CASE
    files:
    - lots.proto
  - id: REQUEST_RESPONSE_TYPES_IN_SAME_FILE
    files:
    - lots.proto
  - id: REQUEST_RESPONSE_TYPES_UNIQUE
    files:
    - lots.proto
  - id: RPC_NAMES_CAPITALIZED
    files:
    - lots.proto
  - id: SERVICE_NAMES_CAMEL_CASE
    files:
    - lots.proto
  - id: SERVICE_NAMES_CAPITALIZED
    files:
    - lots.proto`,
		"lint", "--generate-ignores", "testdata/lint/lots",
	)
	assertExact(
		t,
		true,
		true,
		0,
		``,
		"lint", "--generate-ignores", "testdata/lint/version2",
	)
}

func TestListLinters(t *testing.T) {
	assertLinters(t, lint.DefaultLinters, "lint", "--list-linters", "testdata/lint/base")
	assertLinters(t, lint.Uber1Linters, "lint", "--list-linters", "testdata/lint/base")
	assertLinters(t, lint.EmptyLinters, "lint", "--list-linters", "testdata/lint/empty")
	assertLinterIDs(t, []string{"RPC_NAMES_CAMEL_CASE"}, "lint", "--list-linters", "testdata/lint/emptycustom")
}

func TestListAllLinters(t *testing.T) {
	assertLinters(t, lint.AllLinters, "lint", "--list-all-linters", "testdata/lint/base")
}

func TestListAllLintGroups(t *testing.T) {
	assertExact(t, true, true, 0, "empty\ngoogle\nuber1\nuber2", "lint", "--list-all-lint-groups")
}

func TestListLintGroup(t *testing.T) {
	assertLinters(t, lint.EmptyLinters, "lint", "--list-lint-group", "empty", "testdata/lint/base")
	assertLinters(t, lint.GoogleLinters, "lint", "--list-lint-group", "google", "testdata/lint/base")
	assertLinters(t, lint.Uber1Linters, "lint", "--list-lint-group", "uber1", "testdata/lint/base")
	assertLinters(t, lint.Uber2Linters, "lint", "--list-lint-group", "uber2", "testdata/lint/base")
}

func TestDiffLintGroups(t *testing.T) {
	assertExact(
		t,
		true,
		true,
		0,
		`> COMMENTS_NO_C_STYLE
> ENUMS_NO_ALLOW_ALIAS
> ENUM_FIELD_PREFIXES
> ENUM_ZERO_VALUES_INVALID
> FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX
> FILE_OPTIONS_EQUAL_JAVA_MULTIPLE_FILES_TRUE
> FILE_OPTIONS_EQUAL_JAVA_OUTER_CLASSNAME_PROTO_SUFFIX
> FILE_OPTIONS_EQUAL_JAVA_PACKAGE_COM_PREFIX
> FILE_OPTIONS_GO_PACKAGE_NOT_LONG_FORM
> FILE_OPTIONS_GO_PACKAGE_SAME_IN_DIR
> FILE_OPTIONS_JAVA_MULTIPLE_FILES_SAME_IN_DIR
> FILE_OPTIONS_JAVA_PACKAGE_SAME_IN_DIR
> FILE_OPTIONS_REQUIRE_GO_PACKAGE
> FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES
> FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME
> FILE_OPTIONS_REQUIRE_JAVA_PACKAGE
> ONEOF_NAMES_LOWER_SNAKE_CASE
> PACKAGES_SAME_IN_DIR
> PACKAGE_IS_DECLARED
> PACKAGE_LOWER_SNAKE_CASE
> REQUEST_RESPONSE_TYPES_IN_SAME_FILE
> REQUEST_RESPONSE_TYPES_UNIQUE
> SYNTAX_PROTO3
> WKT_DIRECTLY_IMPORTED`,
		"lint", "--diff-lint-groups", "google,uber1",
	)
}

func TestFiles(t *testing.T) {
	assertExact(t, false, false, 0, `testdata/foo/bar/dep.proto
testdata/foo/success.proto`, "files", "testdata/foo")
}

func TestGenerateDescriptorSetSameDirAsConfigFile(t *testing.T) {
	t.Parallel()
	// https://github.com/uber/prototool/issues/389
	generatedFilePath := "testdata/generate/descriptorset/descriptorset.bin"
	if _, err := os.Stat(generatedFilePath); err == nil {
		assert.NoError(t, os.Remove(generatedFilePath))
	}
	_, exitCode := testDo(t, true, false, "generate", filepath.Dir(generatedFilePath))
	assert.Equal(t, 0, exitCode)
	_, err := os.Stat(generatedFilePath)
	assert.NoError(t, err)
}

func assertLinters(t *testing.T, linters []lint.Linter, args ...string) {
	linterIDs := make([]string, 0, len(linters))
	for _, linter := range linters {
		linterIDs = append(linterIDs, linter.ID())
	}
	sort.Strings(linterIDs)
	assertDo(t, true, true, 0, strings.Join(linterIDs, "\n"), args...)
}

func assertLinterIDs(t *testing.T, linterIDs []string, args ...string) {
	sort.Strings(linterIDs)
	assertDo(t, true, true, 0, strings.Join(linterIDs, "\n"), args...)
}

func assertDoCompileFiles(t *testing.T, expectSuccess bool, asJSON bool, expectedLinePrefixes string, filePaths ...string) {
	lines := getCleanLines(expectedLinePrefixes)
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	cmd := []string{"compile"}
	if asJSON {
		cmd = append(cmd, "--json")
	}
	assertDo(t, true, true, expectedExitCode, strings.Join(lines, "\n"), append(cmd, filePaths...)...)
}

func assertDoCreateFile(t *testing.T, expectSuccess bool, remove bool, filePath string, pkgOverride string, expectedFileData string) {
	assert.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0755))
	if remove {
		_ = os.Remove(filePath)
	}
	args := []string{"create", filePath}
	if pkgOverride != "" {
		args = append(args, "--package", pkgOverride)
	}
	_, exitCode := testDo(t, false, false, args...)
	if expectSuccess {
		assert.Equal(t, 0, exitCode)
		fileData, err := ioutil.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, expectedFileData, string(fileData))
	} else {
		assert.NotEqual(t, 0, exitCode)
	}
}

func assertDoLintFile(t *testing.T, expectSuccess bool, expectedLinePrefixesWithoutFile string, filePath string, args ...string) {
	lines := getCleanLines(expectedLinePrefixesWithoutFile)
	for i, line := range lines {
		lines[i] = filePath + ":" + line
	}
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	assertDo(t, true, true, expectedExitCode, strings.Join(lines, "\n"), append([]string{"lint", filePath}, args...)...)
}

func assertDoLintFiles(t *testing.T, expectSuccess bool, expectedLinePrefixes string, filePaths ...string) {
	lines := getCleanLines(expectedLinePrefixes)
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	assertDo(t, true, true, expectedExitCode, strings.Join(lines, "\n"), append([]string{"lint"}, filePaths...)...)
}

func assertGoldenFormat(t *testing.T, expectSuccess bool, fix bool, filePath string) {
	args := []string{"format"}
	if fix {
		args = append(args, "--fix")
	}
	args = append(args, filePath)
	output, exitCode := testDo(t, true, true, args...)
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	assert.Equal(t, expectedExitCode, exitCode)
	golden, err := ioutil.ReadFile(filePath + ".golden")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(string(golden)), output)
}

func assertDescriptorSet(t *testing.T, expectSuccess bool, dirOrFile string, includeImports bool, includeSourceInfo bool, expectedNames ...string) {
	args := []string{"descriptor-set", "--cache-path", "testcache"}
	if includeImports {
		args = append(args, "--include-imports")
	}
	if includeSourceInfo {
		args = append(args, "--include-source-info")
	}
	args = append(args, dirOrFile)
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	buffer := bytes.NewBuffer(nil)
	exitCode := do(true, args, os.Stdin, buffer, buffer)
	assert.Equal(t, expectedExitCode, exitCode)

	fileDescriptorSet := &descriptor.FileDescriptorSet{}
	assert.NoError(t, proto.Unmarshal(buffer.Bytes(), fileDescriptorSet), buffer.String())
	names := make([]string, 0, len(fileDescriptorSet.File))
	for _, fileDescriptorProto := range fileDescriptorSet.File {
		names = append(names, fileDescriptorProto.GetName())
	}
	sort.Strings(expectedNames)
	sort.Strings(names)
	assert.Equal(t, expectedNames, names)
}

func assertGRPC(t *testing.T, expectedExitCode int, expectedLinePrefixes string, filePath string, method string, jsonData string, extraFlags ...string) {
	assertGRPCExclamationError(t, nil, expectedExitCode, expectedLinePrefixes, filePath, method, jsonData, extraFlags...)
}

func assertGRPCExclamationError(t *testing.T, exclamationError error, expectedExitCode int, expectedLinePrefixes string, filePath string, method string, jsonData string, extraFlags ...string) {
	excitedTestCase := startExcitedTestCase(t, exclamationError)
	defer excitedTestCase.Close()
	assertDoStdin(t, strings.NewReader(jsonData), true, true, expectedExitCode, expectedLinePrefixes, append([]string{"grpc", filePath, "--address", excitedTestCase.Address(), "--method", method, "--stdin", "--connect-timeout", "500ms"}, extraFlags...)...)
}

// GRPC Server TLS assert
func assertGRPCTLS(t *testing.T, expectedExitCode int, expectedLinePrefixes string, filePath string, method string, jsonData string, serverCrt string, serverKey string, caCrt string, extraFlags ...string) {
	assertGRPCmTLS(t, expectedExitCode, expectedLinePrefixes, filePath, method, jsonData, serverCrt, serverKey, caCrt, "", "", "", extraFlags...)
}

// GRPC Mutual TLS assert
func assertGRPCmTLS(t *testing.T, expectedExitCode int, expectedLinePrefixes string, filePath string, method string, jsonData string, serverCrt string, serverKey string, serverCaCert string, clientCert string, clientKey string, clientCaCert string, extraFlags ...string) {
	var excitedTestCase *excitedTestCase
	if clientCaCert != "" {
		excitedTestCase = startmTLSExcitedTestCase(t, serverCrt, serverKey, clientCaCert)
	} else {
		excitedTestCase = startTLSExcitedTestCase(t, serverCrt, serverKey)
	}
	defer excitedTestCase.Close()
	args := []string{"grpc", filePath, "--address", excitedTestCase.Address(), "--method", method, "--stdin", "--connect-timeout", "500ms"}
	if serverCaCert != "" {
		args = append(args, "--cacert", serverCaCert, "--tls")
	}
	if clientCert != "" {
		args = append(args, "--cert", clientCert, "--key", clientKey)
	}
	assertDoStdin(t, strings.NewReader(jsonData), true, true, expectedExitCode, expectedLinePrefixes, append(args, extraFlags...)...)
}

func assertRegexp(t *testing.T, withCachePath bool, extraErrorFormat bool, expectedExitCode int, expectedRegexp string, args ...string) {
	stdout, exitCode := testDo(t, withCachePath, extraErrorFormat, args...)
	assert.Equal(t, expectedExitCode, exitCode)
	matched, err := regexp.MatchString(expectedRegexp, stdout)
	assert.NoError(t, err)
	assert.True(t, matched, "Expected regex %s but got %s", expectedRegexp, stdout)
}

func assertExact(t *testing.T, withCachePath bool, extraErrorFormat bool, expectedExitCode int, expectedStdout string, args ...string) {
	stdout, exitCode := testDo(t, withCachePath, extraErrorFormat, args...)
	assert.Equal(t, expectedExitCode, exitCode)
	assert.Equal(t, expectedStdout, stdout)
}

func assertDoStdin(t *testing.T, stdin io.Reader, withCachePath bool, extraErrorFormat bool, expectedExitCode int, expectedLinePrefixes string, args ...string) {
	assertDoInternal(t, stdin, withCachePath, extraErrorFormat, expectedExitCode, expectedLinePrefixes, args...)
}

func assertDo(t *testing.T, withCachePath bool, extraErrorFormat bool, expectedExitCode int, expectedLinePrefixes string, args ...string) {
	assertDoInternal(t, nil, withCachePath, extraErrorFormat, expectedExitCode, expectedLinePrefixes, args...)
}

func testDoStdin(t *testing.T, stdin io.Reader, withCachePath bool, extraErrorFormat bool, args ...string) (string, int) {
	return testDoInternal(stdin, withCachePath, extraErrorFormat, args...)
}

func testDo(t *testing.T, withCachePath bool, extraErrorFormat bool, args ...string) (string, int) {
	return testDoInternal(nil, withCachePath, extraErrorFormat, args...)
}

func getCleanLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

type excitedTestCase struct {
	listener      net.Listener
	grpcServer    *grpc.Server
	excitedServer *excitedServer
}

func startTLSExcitedTestCase(t *testing.T, serverCert string, serverKey string) *excitedTestCase {
	creds, err := credentials.NewServerTLSFromFile(serverCert, serverKey)
	require.NoError(t, err)
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	return startExcitedTestCaseWithServer(t, nil, grpcServer)
}

func startmTLSExcitedTestCase(t *testing.T, serverCert string, serverKey string, clientCaCerts string) *excitedTestCase {
	certificate, err := tls.LoadX509KeyPair(serverCert, serverKey)
	require.NoError(t, err)

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(clientCaCerts)
	require.NoError(t, err)

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		require.NoError(t, err)
	}
	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	return startExcitedTestCaseWithServer(t, nil, grpcServer)
}

func startExcitedTestCase(t *testing.T, exclamationError error) *excitedTestCase {
	return startExcitedTestCaseWithServer(t, exclamationError, grpc.NewServer())
}

func startExcitedTestCaseWithServer(t *testing.T, exclamationError error, grpcServer *grpc.Server) *excitedTestCase {
	listener, err := getFreeListener()
	require.NoError(t, err)
	excitedServer := newExcitedServer(exclamationError)
	grpcpb.RegisterExcitedServiceServer(grpcServer, excitedServer)
	go func() { _ = grpcServer.Serve(listener) }()
	return &excitedTestCase{
		listener:      listener,
		grpcServer:    grpcServer,
		excitedServer: excitedServer,
	}
}

func (c *excitedTestCase) Address() string {
	if c.listener == nil {
		return ""
	}
	return c.listener.Addr().String()
}

func (c *excitedTestCase) Close() {
	if c.grpcServer != nil {
		c.grpcServer.Stop()
	}
}

type excitedServer struct {
	exclamationError error
}

func newExcitedServer(exclamationError error) *excitedServer {
	return &excitedServer{
		exclamationError: exclamationError,
	}
}

func (s *excitedServer) Exclamation(ctx context.Context, request *grpcpb.ExclamationRequest) (*grpcpb.ExclamationResponse, error) {
	if s.exclamationError != nil {
		return nil, s.exclamationError
	}
	return &grpcpb.ExclamationResponse{
		Value: request.Value + "!",
	}, nil
}

func (s *excitedServer) ExclamationClientStream(streamServer grpcpb.ExcitedService_ExclamationClientStreamServer) error {
	value := ""
	for request, err := streamServer.Recv(); err != io.EOF; request, err = streamServer.Recv() {
		if err != nil {
			return err
		}
		value += request.Value
	}
	return streamServer.SendAndClose(&grpcpb.ExclamationResponse{
		Value: value + "!",
	})
}

func (s *excitedServer) ExclamationServerStream(request *grpcpb.ExclamationRequest, streamServer grpcpb.ExcitedService_ExclamationServerStreamServer) error {
	for _, c := range request.Value {
		if err := streamServer.Send(&grpcpb.ExclamationResponse{
			Value: string(c),
		}); err != nil {
			return err
		}
	}
	return streamServer.Send(&grpcpb.ExclamationResponse{
		Value: "!",
	})
}

func (s *excitedServer) ExclamationBidiStream(streamServer grpcpb.ExcitedService_ExclamationBidiStreamServer) error {
	for request, err := streamServer.Recv(); err != io.EOF; request, err = streamServer.Recv() {
		if err != nil {
			return err
		}
		if err := streamServer.Send(&grpcpb.ExclamationResponse{
			Value: request.Value + "!",
		}); err != nil {
			return err
		}
	}
	return nil
}

// do not use these in tests

func assertDoInternal(t *testing.T, stdin io.Reader, withCachePath bool, extraErrorFormat bool, expectedExitCode int, expectedLinePrefixes string, args ...string) {
	stdout, exitCode := testDoStdin(t, stdin, withCachePath, extraErrorFormat, args...)
	outputSplit := getCleanLines(stdout)
	assert.Equal(t, expectedExitCode, exitCode, strings.Join(outputSplit, "\n"))
	expectedLinePrefixesSplit := getCleanLines(expectedLinePrefixes)
	require.Equal(t, len(expectedLinePrefixesSplit), len(outputSplit), strings.Join(outputSplit, "\n"))
	for i, expectedLinePrefix := range expectedLinePrefixesSplit {
		assert.True(t, strings.HasPrefix(outputSplit[i], expectedLinePrefix), "%s %d %s", expectedLinePrefix, i, strings.Join(outputSplit, "\n"))
	}
}

func testDoInternal(stdin io.Reader, withCachePath bool, extraErrorFormat bool, args ...string) (string, int) {
	if stdin == nil {
		stdin = os.Stdin
	}
	if withCachePath {
		args = append(
			args,
			"--cache-path", "testcache",
		)
	}
	if extraErrorFormat {
		args = append(
			args,
			"--error-format", "filename:line:column:id:message",
		)
	}
	buffer := bytes.NewBuffer(nil)
	// develMode is on, so we have access to all commands
	exitCode := do(true, args, stdin, buffer, buffer)
	return strings.TrimSpace(buffer.String()), exitCode
}

func getFreeListener() (net.Listener, error) {
	address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return net.ListenTCP("tcp", address)
}
