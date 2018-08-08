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

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/cmd/testdata/grpc/gen/grpcpb"
	"github.com/uber/prototool/internal/lint"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/vars"
	"google.golang.org/grpc"
)

var (
	// testLock is to lock around prototool download in testDownload
	testLock sync.Mutex
)

func TestCompile(t *testing.T) {
	t.Parallel()
	assertDoCompileFiles(
		t,
		false,
		`testdata/compile/errors_on_import/dep_errors.proto:6:1:Expected ";".`,
		"testdata/compile/errors_on_import/dep_errors.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		`testdata/compile/errors_on_import/dep_errors.proto:6:1:Expected ";".`,
		"testdata/compile/errors_on_import",
	)
	assertDoCompileFiles(
		t,
		false,
		`testdata/compile/extra_import/extra_import.proto:1:1:Import "dep.proto" was not used.`,
		"testdata/compile/extra_import/extra_import.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		`testdata/compile/json/json_camel_case_conflict.proto:1:1:The JSON camel-case name of field "helloworld" conflicts with field "helloWorld". This is not allowed in proto3.`,
		"testdata/compile/json/json_camel_case_conflict.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		`testdata/compile/semicolon/missing_package_semicolon.proto:5:1:Expected ";".`,
		"testdata/compile/semicolon/missing_package_semicolon.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		`testdata/compile/syntax/missing_syntax.proto:1:1:No syntax specified. Please use 'syntax = "proto2";' or 'syntax = "proto3";' to specify a syntax version.
		testdata/compile/syntax/missing_syntax.proto:4:3:Expected "required", "optional", or "repeated".`,
		"testdata/compile/syntax/missing_syntax.proto",
	)
	assertDoCompileFiles(
		t,
		true,
		``,
		"testdata/compile/proto2/syntax_proto2.proto",
	)
	assertDoCompileFiles(
		t,
		false,
		`testdata/compile/notimported/not_imported.proto:11:3:"foo.Dep" seems to be defined in "dep.proto", which is not imported by "not_imported.proto".  To use it here, please add the necessary import.`,
		"testdata/compile/notimported/not_imported.proto",
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

	assertDo(t, 0, "", "config", "init", tmpDir)
	assertDo(t, 1, fmt.Sprintf("%s already exists", filepath.Join(tmpDir, settings.DefaultConfigFilename)), "config", "init", tmpDir)
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
}

func TestGoldenFormat(t *testing.T) {
	t.Parallel()
	assertGoldenFormat(t, false, false, "testdata/format/proto3/foo/bar/bar.proto")
	assertGoldenFormat(t, false, false, "testdata/format/proto2/foo/bar/bar_proto2.proto")
	assertGoldenFormat(t, false, false, "testdata/format/proto3/foo/foo.proto")
	assertGoldenFormat(t, false, false, "testdata/format/proto2/foo/foo_proto2.proto")
	assertGoldenFormat(t, false, true, "testdata/format-fix/foo.proto")
}

func TestJSONToBinaryToJSON(t *testing.T) {
	t.Parallel()
	assertJSONToBinaryToJSON(t, "testdata/foo/success.proto", "foo.Baz", `{"hello":100}`)
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
		`syntax = "proto3";

package foo;

option go_package = "foopb";
option java_multiple_files = true;
option java_outer_classname = "BazProto";
option java_package = "com.foo";`,
	)
}

func TestGRPC(t *testing.T) {
	t.Parallel()
	assertGRPC(t,
		0,
		`
		{
			"value": "hello!"
		}
		`,
		"testdata/grpc/grpc.proto",
		"grpc.ExcitedService/Exclamation",
		`{"value":"hello"}`,
	)
	assertGRPC(t,
		0,
		`
		{
			"value": "hellosalutations!"
		}
		`,
		"testdata/grpc/grpc.proto",
		"grpc.ExcitedService/ExclamationClientStream",
		`{"value":"hello"}
		{"value":"salutations"}`,
	)
	assertGRPC(t,
		0,
		`
		{
			"value": "h"
		}
		{
			"value": "e"
		}
		{
			"value": "l"
		}
		{
			"value": "l"
		}
		{
			"value": "o"
		}
		{
			"value": "!"
		}
		`,
		"testdata/grpc/grpc.proto",
		"grpc.ExcitedService/ExclamationServerStream",
		`{"value":"hello"}`,
	)
	assertGRPC(t,
		0,
		`
		{
			"value": "hello!"
		}
		{
			"value": "salutations!"
		}
		`,
		"testdata/grpc/grpc.proto",
		"grpc.ExcitedService/ExclamationBidiStream",
		`{"value":"hello"}
		{"value":"salutations"}`,
	)
}

func TestVersion(t *testing.T) {
	assertRegexp(t, 0, fmt.Sprintf("Version:.*%s\nDefault protoc version:.*%s\n", vars.Version, vars.DefaultProtocVersion), "version")
}

func TestListAllLintGroups(t *testing.T) {
	assertExact(t, 0, "all\ndefault", "list-all-lint-groups")
}

func TestDescriptorProto(t *testing.T) {
	assertExact(
		t,
		0,
		`{
  "name": "Baz",
  "field": [
    {
      "name": "hello",
      "number": 1,
      "label": "LABEL_OPTIONAL",
      "type": "TYPE_INT64",
      "jsonName": "hello"
    },
    {
      "name": "dep",
      "number": 2,
      "label": "LABEL_OPTIONAL",
      "type": "TYPE_MESSAGE",
      "typeName": ".bar.Dep",
      "jsonName": "dep"
    },
    {
      "name": "timestamp",
      "number": 3,
      "label": "LABEL_OPTIONAL",
      "type": "TYPE_MESSAGE",
      "typeName": ".google.protobuf.Timestamp",
      "jsonName": "timestamp"
    }
  ]
}`,
		"descriptor-proto", "testdata/foo/success.proto", "foo.Baz",
	)
}

func TestFieldDescriptorProto(t *testing.T) {
	assertExact(
		t,
		0,
		`{
  "name": "dep",
  "number": 2,
  "label": "LABEL_OPTIONAL",
  "type": "TYPE_MESSAGE",
  "typeName": ".bar.Dep",
  "jsonName": "dep"
}`,
		"field-descriptor-proto", "testdata/foo/success.proto", "foo.Baz.dep",
	)
}

func TestServiceDescriptorProto(t *testing.T) {
	assertExact(
		t,
		0,
		`{
  "name": "ExcitedService",
  "method": [
    {
      "name": "Exclamation",
      "inputType": ".grpc.ExclamationRequest",
      "outputType": ".grpc.ExclamationResponse",
      "options": {

      }
    },
    {
      "name": "ExclamationClientStream",
      "inputType": ".grpc.ExclamationRequest",
      "outputType": ".grpc.ExclamationResponse",
      "options": {

      },
      "clientStreaming": true
    },
    {
      "name": "ExclamationServerStream",
      "inputType": ".grpc.ExclamationRequest",
      "outputType": ".grpc.ExclamationResponse",
      "options": {

      },
      "serverStreaming": true
    },
    {
      "name": "ExclamationBidiStream",
      "inputType": ".grpc.ExclamationRequest",
      "outputType": ".grpc.ExclamationResponse",
      "options": {

      },
      "clientStreaming": true,
      "serverStreaming": true
    }
  ]
}`,
		"service-descriptor-proto", "testdata/grpc", "grpc.ExcitedService",
	)
}

func TestListLinters(t *testing.T) {
	assertLinters(t, lint.DefaultLinters, "lint", "--list-linters")
}

func TestListAllLinters(t *testing.T) {
	assertLinters(t, lint.AllLinters, "lint", "--list-all-linters")
}

func assertLinters(t *testing.T, linters []lint.Linter, args ...string) {
	linterIDs := make([]string, 0, len(linters))
	for _, linter := range linters {
		linterIDs = append(linterIDs, linter.ID())
	}
	sort.Strings(linterIDs)
	assertDo(t, 0, strings.Join(linterIDs, "\n"), args...)
}

func assertDoCompileFiles(t *testing.T, expectSuccess bool, expectedLinePrefixes string, filePaths ...string) {
	lines := getCleanLines(expectedLinePrefixes)
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	assertDo(t, expectedExitCode, strings.Join(lines, "\n"), append([]string{"compile"}, filePaths...)...)
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
	_, exitCode := testDo(t, args...)
	if expectSuccess {
		assert.Equal(t, 0, exitCode)
		fileData, err := ioutil.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, expectedFileData, string(fileData))
	} else {
		assert.NotEqual(t, 0, exitCode)
	}
}

func assertDoLintFile(t *testing.T, expectSuccess bool, expectedLinePrefixesWithoutFile string, filePath string) {
	lines := getCleanLines(expectedLinePrefixesWithoutFile)
	for i, line := range lines {
		lines[i] = filePath + ":" + line
	}
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	assertDo(t, expectedExitCode, strings.Join(lines, "\n"), "lint", filePath)
}

func assertDoLintFiles(t *testing.T, expectSuccess bool, expectedLinePrefixes string, filePaths ...string) {
	lines := getCleanLines(expectedLinePrefixes)
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	assertDo(t, expectedExitCode, strings.Join(lines, "\n"), append([]string{"lint"}, filePaths...)...)
}

func assertGoldenFormat(t *testing.T, expectSuccess bool, fix bool, filePath string) {
	args := []string{"format"}
	if fix {
		args = append(args, "--fix")
	}
	args = append(args, filePath)
	output, exitCode := testDo(t, args...)
	expectedExitCode := 0
	if !expectSuccess {
		expectedExitCode = 255
	}
	assert.Equal(t, expectedExitCode, exitCode)
	golden, err := ioutil.ReadFile(filePath + ".golden")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(string(golden)), output)
}

func assertJSONToBinaryToJSON(t *testing.T, filePath string, messagePath string, jsonData string) {
	stdout, exitCode := testDo(t, "json-to-binary", filePath, messagePath, jsonData)
	assert.Equal(t, 0, exitCode)
	stdout, exitCode = testDo(t, "binary-to-json", filePath, messagePath, stdout)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, jsonData, stdout)
}

func assertGRPC(t *testing.T, expectedExitCode int, expectedLinePrefixes string, filePath string, method string, jsonData string) {
	excitedTestCase := startExcitedTestCase(t)
	defer excitedTestCase.Close()
	assertDoStdin(t, strings.NewReader(jsonData), expectedExitCode, expectedLinePrefixes, "grpc", filePath, "--address", excitedTestCase.Address(), "--method", method, "--stdin")
}

func assertRegexp(t *testing.T, expectedExitCode int, expectedRegexp string, args ...string) {
	stdout, exitCode := testDo(t, args...)
	assert.Equal(t, expectedExitCode, exitCode)
	matched, err := regexp.MatchString(expectedRegexp, stdout)
	assert.NoError(t, err)
	assert.True(t, matched, "Expected regex %s but got %s", expectedRegexp, stdout)
}

func assertExact(t *testing.T, expectedExitCode int, expectedStdout string, args ...string) {
	stdout, exitCode := testDo(t, args...)
	assert.Equal(t, expectedExitCode, exitCode)
	assert.Equal(t, expectedStdout, stdout)
}

func assertDoStdin(t *testing.T, stdin io.Reader, expectedExitCode int, expectedLinePrefixes string, args ...string) {
	assertDoInternal(t, stdin, expectedExitCode, expectedLinePrefixes, args...)
}

func assertDo(t *testing.T, expectedExitCode int, expectedLinePrefixes string, args ...string) {
	assertDoInternal(t, nil, expectedExitCode, expectedLinePrefixes, args...)
}

func testDoStdin(t *testing.T, stdin io.Reader, args ...string) (string, int) {
	testDownload(t)
	return testDoInternal(stdin, args...)
}

func testDo(t *testing.T, args ...string) (string, int) {
	testDownload(t)
	return testDoInternal(nil, args...)
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

func startExcitedTestCase(t *testing.T) *excitedTestCase {
	listener, err := getFreeListener()
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	excitedServer := newExcitedServer()
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

type excitedServer struct{}

func newExcitedServer() *excitedServer {
	return &excitedServer{}
}

func (s *excitedServer) Exclamation(ctx context.Context, request *grpcpb.ExclamationRequest) (*grpcpb.ExclamationResponse, error) {
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

func assertDoInternal(t *testing.T, stdin io.Reader, expectedExitCode int, expectedLinePrefixes string, args ...string) {
	stdout, exitCode := testDoStdin(t, stdin, args...)
	outputSplit := getCleanLines(stdout)
	assert.Equal(t, expectedExitCode, exitCode, strings.Join(outputSplit, "\n"))
	expectedLinePrefixesSplit := getCleanLines(expectedLinePrefixes)
	require.Equal(t, len(expectedLinePrefixesSplit), len(outputSplit), strings.Join(outputSplit, "\n"))
	for i, expectedLinePrefix := range expectedLinePrefixesSplit {
		assert.True(t, strings.HasPrefix(outputSplit[i], expectedLinePrefix), "%s %d %s", expectedLinePrefix, i, strings.Join(outputSplit, "\n"))
	}
}

func testDownload(t *testing.T) {
	testLock.Lock()
	defer testLock.Unlock()
	// download checks if protoc is already downloaded to the cache location
	// if it is, then this is effectively a no-op
	// if it isn't, then this downloads to the cache
	stdout, exitCode := testDoInternal(nil, "download")
	require.Equal(t, 0, exitCode, "had non-zero exit code when downloading: %s", stdout)
}

func testDoInternal(stdin io.Reader, args ...string) (string, int) {
	args = append(args,
		"--print-fields", "filename:line:column:id:message",
	)
	if stdin == nil {
		stdin = os.Stdin
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
