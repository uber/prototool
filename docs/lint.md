# Lint

[Back to README.md](README.md)

`prototool lint` lints your Protobuf files.

Lint rules can be set using the configuration file. See the configuration at
[etc/config/example/prototool.yaml](../etc/config/example/prototool.yaml) for all available
options. There are three pre-configured groups of rules, the setting of which is integral to the
`prototool lint`, `prototool create`, and `prototool format` commands:

- `uber2`: This lint group follows the [V2 Uber Style Guide](../style/README.md), and makes some
  modifications to more closely follow the Google Cloud APIs file structure, as well as adding even
  more rules to enforce more consistent development patterns. This is the lint group we recommend
  using.
- `uber1`: This lint group follows the [V1 Uber Style Guide](../etc/style/uber1/uber1.proto). For
  backwards compatibility reasons, this is the default lint group, however we recommend using the
  `uber2` lint group.
- `google`: This lint group follows the
  [Google Style Guide](https://developers.google.com/protocol-buffers/docs/style). This is a small
  group of rules meant to enforce basic naming. The style guide is copied to
  [etc/style/google/google.proto](../etc/style/google/google.proto).

To see the differences between lint groups, use the `--diff-lint-groups` flag:

```
prototool lint --diff-lint-groups uber1,uber2
```

Configuration of your group can be done by setting the `lint.group` option in your `prototool.yaml`
file:

```yaml
lint:
  group: uber2
```

See the `prototool.yaml` files at
[etc/style/google/prototool.yaml](../etc/style/google/prototool.yaml) and
[etc/style/uber1/prototool.yaml](../etc/style/uber1/prototool.yaml) for examples.

There is also the special lint group `empty`, which has no lint rules. This allows one to specify
only the linters they want in `lint.rules.add`:

```yaml
lint:
  group: empty
  rules:
    add:
      - MESSAGE_NAMES_CAMEL_CASE
      - MESSAGE_NAMES_CAPITALIZED
```

You can configure ignoring of lint rules on a per-file basis:

```yaml
lint:
  ignores:
    - id: MESSAGE_NAMES_CAMEL_CASE
      files:
        - foo.proto
        - bar/baz.proto
```

To generate the a YAML configuration for currently-failing lint rules that can be copied into your
configuration file, use `--generate-ignores`. This will lint your files, ignoring the existing
setting for `lint.ignores`, and print a new value for it. Note that you should make sure not to
touch other settings for `lint` in your configuration file as this flag only generates the
`lint.ignores` option.

```
prototool lint path/to/dir --generate-ignores
```

Linting also understands the concept of file headers, typically license headers. To specify a file
header, add the following to your `prototool.yaml`:

```yaml
lint:
  file_header:
    path: path/to/header.txt
    is_commented: true
```

Alternatively, directly specify the content:

```yaml
lint:
  file_header:
    content: |
      //
      // Acme, Inc. (c) 2019
      //
    is_commented: true
```

The `path` option specifies the path to the file that contains the header data. The `content`
option specifies the content directly. Only one of these can be specified. The `is_commented`
option specifies whether the header data is already commented, and if not, `// ` will be added
before all non-empty lines, and `//` will be added before all empty lines. `is_commented` is
optional and generally will not be set if the file is not commented, for example if `path` points
to a text LICENSE file.

If `lint.file_header.path` or `lint.file_header.content` is set, `prototool lint`,
`prototool create`, and `prototool format --fix` will all take the file header into account.

See [internal/cmd/testdata/lint](../internal/cmd/testdata/lint) for additional examples of
configurations, and run `prototool lint internal/cmd/testdata/lint/DIR` from a checkout of this
repository to see example failures.

Files must be valid Protobuf that can be compiled with `protoc`, so prior to linting,
`prototool lint` will compile your using `protoc`. Note, however, this is very fast - for the two
files in [etc/style/uber1](../etc/style/uber1), compiling and linting only takes approximately
3/100ths of a second:

```bash
$ time prototool lint etc/style/uber1

real	0m0.037s
user	0m0.026s
sys	0m0.017s
```

For all 694 Protobuf files currently in [googleapis](https://github.com/googleapis/googleapis),
this takes approximately 3/4ths of a second:

```bash
$ git remote -v
origin	https://github.com/googleapis/googleapis (fetch)
origin	https://github.com/googleapis/googleapis (push)

$ cat prototool.yaml
protoc:
  allow_unused_imports: true
lint:
  group: google

$ time prototool lint

real	0m0.734s
user	0m3.835s
sys	0m0.924s
```

## Available Lint Rules

Following is a list of all lint rules provided by Prototool as well as what lint group they belong to (if applicable).

| Rule | Description | Lint Group |
| --- | --- | --- |
| COMMENTS_NO_C_STYLE | Verifies that there are no `/* C-style */` comments. | uber1, uber2|
| COMMENTS_NO_INLINE | Verifies that there are no inline comments. | uber2 |
| ENUM_FIELD_NAMES_UPPER_SNAKE_CASE | Verifies that all enum field names are `UPPER_SNAKE_CASE`. | google, uber1, uber2 |
| ENUM_FIELD_NAMES_UPPERCASE | Verifies that all enum field names are `UPPERCASE`. | none |
| ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE | Verifies that all enum fields are prefixed with `ENUM_NAME_`. | uber2 |
| ENUM_FIELD_PREFIXES | Verifies that all enum fields are prefixed with `[NESTED_MESSAGE_NAME_]ENUM_NAME_`. | uber1 |
| ENUM_FIELDS_HAVE_COMMENTS | Verifies that all enum fields have a comment of the form `// FIELD_NAME ...`. | none |
| ENUM_FIELDS_HAVE_SENTENCE_COMMENTS | Verifies that all enum fields have a comment that contains at least one complete sentence. | none |
| ENUM_NAMES_CAMEL_CASE | Verifies that all enum names are CamelCase. | google, uber1, uber2 |
| ENUM_NAMES_CAPITALIZED | Verifies that all enum names are Capitalized. | google, uber1, uber2 |
| ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE | Verifies that all enum zero value names are `ENUM_NAME_INVALID`. | uber2 |
| ENUM_ZERO_VALUES_INVALID | Verifies that all enum zero value names are `[NESTED_MESSAGE_NAME_]ENUM_NAME_INVALID`. | uber1 |
| ENUMS_HAVE_COMMENTS | Verifies that all enums have a comment of the form `// EnumName ...`. | none |
| ENUMS_HAVE_SENTENCE_COMMENTS | Verifies that all enums have a comment that contains at least one complete sentence. | uber2 |
| ENUMS_NO_ALLOW_ALIAS | Verifies that no enums use the option `allow_alias`. | uber1, uber2 |
| FIELDS_NOT_RESERVED | Verifies that no message or enum has a reserved field. | uber2 |
| FILE_HEADER | Verifies that the file header matches the expected file header if the file_header option is set in the configuration file. | google, uber1, uber2 |
| FILE_NAMES_LOWER_SNAKE_CASE | Verifies that the file name is `lower_snake_case.proto`. | uber2 |
| FILE_OPTIONS_EQUAL_CSHARP_NAMESPACE_CAPITALIZED | Verifies that the file option `csharp_namespace` is the capitalized version of the package. | uber2 |
| FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX | Verifies that the file option `go_package` is equal to `$(basename PACKAGE)pb`. | uber1 |
| FILE_OPTIONS_EQUAL_GO_PACKAGE_V2_SUFFIX | Verifies that the file option `go_package` is equal to the last two values of the package separated by "."s, or just the package name if there are no "."s. | uber2 |
| FILE_OPTIONS_EQUAL_JAVA_MULTIPLE_FILES_TRUE | Verifies that the file option `java_multiple_files` is equal to true. | uber1, uber2 |
FILE_OPTIONS_EQUAL_JAVA_OUTER_CLASSNAME_PROTO_SUFFIX | Verifies that the file option `java_outer_classname` is equal to `$(upperCamelCase $(basename FILE))Proto.` | uber1, uber2 |
FILE_OPTIONS_EQUAL_JAVA_PACKAGE_COM_PREFIX | Verifies that the file option `java_package` is equal to `com.PACKAGE`. | uber1 |
FILE_OPTIONS_EQUAL_JAVA_PACKAGE_PREFIX | Verifies that the file option `java_package` is equal to `PREFIX.PACKAGE`, with `PREFIX` defaulting to `com` and configurable in your configuration file. | uber2 |
FILE_OPTIONS_EQUAL_OBJC_CLASS_PREFIX_ABBR | Verifies that the file option `objc_class_prefix` is the abbreviated version of the package. | uber2 |
FILE_OPTIONS_EQUAL_PHP_NAMESPACE_CAPITALIZED | Verifies that the file option "php_namespace" is the capitalized version of the package. | uber2 |
FILE_OPTIONS_REQUIRE_CSHARP_NAMESPACE | Verifies that the file option `csharp_namespace` is set. | uber2 |
FILE_OPTIONS_REQUIRE_GO_PACKAGE | Verifies that the file option `go_package` is set. | uber1, uber2 |
FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES | Verifies that the file option `java_multiple_files` is set. | uber1, uber2 |
FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME | Verifies that the file option `java_outer_classname` is set. | uber1, uber2 |
FILE_OPTIONS_REQUIRE_JAVA_PACKAGE | Verifies that the file option `java_package` is set. | uber1, uber2 |
FILE_OPTIONS_REQUIRE_OBJC_CLASS_PREFIX | Verifies that the file option `objc_class_prefix` is set. | uber2 |
FILE_OPTIONS_REQUIRE_PHP_NAMESPACE | Verifies that the file option `php_namespace` is set. | uber2 |
FILE_OPTIONS_REQUIRE_RUBY_PACKAGE | Verifies that the file option `ruby_package` is set. | none |
FILE_OPTIONS_CSHARP_NAMESPACE_SAME_IN_DIR | Verifies that the file option `csharp_namespace` of all files in a directory are the same. | uber2 |
FILE_OPTIONS_GO_PACKAGE_SAME_IN_DIR | Verifies that the file option `go_package` of all files in a directory are the same. | uber1, uber2 |
FILE_OPTIONS_JAVA_MULTIPLE_FILES_SAME_IN_DIR | Verifies that the file option `java_multiple_files` of all files in a directory are the same. | uber1, uber2 |
FILE_OPTIONS_JAVA_PACKAGE_SAME_IN_DIR | Verifies that the file option `java_package` of all files in a directory are the same. | uber1, uber2 |
FILE_OPTIONS_OBJC_CLASS_PREFIX_SAME_IN_DIR | Verifies that the file option `objc_class_prefix` of all files in a directory are the same. | uber2 |
FILE_OPTIONS_PHP_NAMESPACE_SAME_IN_DIR | Verifies that the file option `php_namespace` of all files in a directory are the same. | uber2 |
FILE_OPTIONS_UNSET_JAVA_MULTIPLE_FILES | Verifies that the file option `java_multiple_files` is unset. | none |
FILE_OPTIONS_UNSET_JAVA_OUTER_CLASSNAME | Verifies that the file option `java_outer_classname` is unset. | none |
FILE_OPTIONS_GO_PACKAGE_NOT_LONG_FORM | Verifies that the file option `go_package` is not of the form `go/import/path;package`. | uber1, uber2 |
GOGO_NOT_IMPORTED | Verifies that the `gogo.proto` file from `gogo/protobuf` is not imported. | none |
IMPORTS_NOT_PUBLIC | Verifies that there are no public imports. | uber2 |
IMPORTS_NOT_WEAK | Verifies that there are no weak imports. | uber2 |
MESSAGE_FIELD_NAMES_FILENAME | Verifies that all message field names do not contain   `file_name` as `filename` should be used instead. | uber2 |
MESSAGE_FIELD_NAMES_FILEPATH | Verifies that all message field names do not contain   `file_path` as `filepath` should be used instead. | uber2 |
MESSAGE_FIELD_NAMES_LOWER_SNAKE_CASE | Verifies that all message field names are `lower_snake_case`. | google, uber1, uber2 |
MESSAGE_FIELD_NAMES_LOWERCASE | Verifies that all message field names are `lowercase`. | none |
MESSAGE_FIELD_NAMES_NO_DESCRIPTOR | Verifies that all message field names are not `descriptor`, which results in a collision in Java-generated code. | uber2 |
MESSAGE_FIELDS_DURATION | Verifies that all non-map fields that contain "duration" in their name are of type `google.protobuf.Duration`. | none |
MESSAGE_FIELDS_HAVE_COMMENTS | Verifies that all message fields have a comment of the form `// field_name ...`. | none |
MESSAGE_FIELDS_HAVE_SENTENCE_COMMENTS | Verifies that all message fields have a comment that contains at least one complete sentence. | none |
MESSAGE_FIELDS_NO_JSON_NAME | Verifies that no message field has the `json_name` option set. | uber2 |
MESSAGE_FIELDS_NOT_FLOATS | Verifies that all message fields are not floats. | none |
MESSAGE_FIELDS_TIME | Verifies that all non-map fields that contain `time` in their name are of type `google.protobuf.Timestamp`. | none |
MESSAGE_NAMES_CAMEL_CASE | Verifies that all non-extended message names are CamelCase. | google, uber1, uber2 |
MESSAGE_NAMES_CAPITALIZED | Verifies that all non-extended message names are Capitalized. | google, uber1, uber2 |
MESSAGES_HAVE_COMMENTS_EXCEPT_REQUEST_RESPONSE_TYPES | Verifies that all non-extended messages except for request and response types have a comment of the form `// MessageName ...`. | none |
MESSAGES_HAVE_COMMENTS | Verifies that all non-extended messages have a comment of the form `// MessageName ...`. | none |
MESSAGES_HAVE_SENTENCE_COMMENTS
_EXCEPT_REQUEST_RESPONSE_TYPES | Verifies that all non-extended messages except for request and response types have a comment that contains at least one complete sentence. | uber2 |
MESSAGES_NOT_EMPTY_EXCEPT_REQUEST_RESPONSE_TYPES | Verifies that all messages except for request and response types are not empty. | none |
NAMES_NO_COMMON | Verifies that no type name contains the word `common` because `common` has no semantic meaning, consider using a name that reflects the type instead. | uber2 |
NAMES_NO_DATA | Verifies that no type name contains the word `data` because `data` is a decorator and all types on Protobuf are data.  Consider merging this information into a higher-level type, or if you must have such a type, Use `info` instead. | uber2 |
NAMES_NO_UUID | Verifies that no type name contains the word `uuid` because UUIDs in Protobuf are named ID instead of UUID. | uber2 |
ONEOF_NAMES_LOWER_SNAKE_CASE | Verifies that all `oneof` names are `lower_snake_case`. | uber1, uber2 |
PACKAGE_IS_DECLARED | Verifies that there is one and only one `package` declaration. | uber1, uber2 |
PACKAGE_LOWER_CASE | Verifies that there is one and only one `package` declaration and that the package name only contains characters in the range `a-z0-9` and periods. | uber2 |
PACKAGE_LOWER_SNAKE_CASE | Verifies that there is one and only one `package` declaration and the package is lower_snake.case. | uber1 |
PACKAGE_MAJOR_BETA_VERSIONED | Verifies that there is one and only one `package` declaration and the package is of the form `package.vMAJORVERSION` or `package.vMAJORVERSIONbetaBETAVERSION` with versions >=1. | uber2 |
PACKAGE_NO_KEYWORDS | Verifies that no packages contain one of the following keywords as part of the name when split on "." :  `internal`, `public`, `private`, `protected`, `std`. | uber2 |
PACKAGES_SAME_IN_DIR | Verifies that the packages of all files in a directory are the same. | uber1, uber2 |
REQUEST_RESPONSE_NAMES_MATCH_RPC | Verifies that all request names are of the pattern `RpcNameRequest` and all response names are of the pattern `RpcNameResponse`. | uber2 |
REQUEST_RESPONSE_TYPES_AFTER_SERVICE | Verifies that request and response types are defined after any services and the response type is defined after the request type. | uber2 |
REQUEST_RESPONSE_TYPES_IN_SAME_FILE | Verifies that all request and response types are in the same file as their corresponding service and are not nested messages. | uber1, uber2 |
REQUEST_RESPONSE_TYPES_ONLY_IN_FILE | Verifies that only request and response types are the only types in the same file as their corresponding service. | uber2 |
REQUEST_RESPONSE_TYPES_UNIQUE | Verifies that all request and response types are unique to each RPC. | uber1, uber2 |
RPC_NAMES_CAMEL_CASE | Verifies that all RPC names are CamelCase. | google, uber1, uber2 |
RPC_NAMES_CAPITALIZED | Verifies that all RPC names are Capitalized. | google, uber1, uber2 |
RPC_OPTIONS_NO_GOOGLE_API_HTTP | Verifies that the RPC option `google.api.http` is not used. | none |
RPCS_HAVE_COMMENTS | Verifies that all rpcs have a comment of the form `// RPCName ...`. | none |
RPCS_HAVE_SENTENCE_COMMENTS | Verifies that all rpcs have a comment that contains at least one complete sentence. | uber2 |
RPCS_NO_STREAMING | Verifies that all rpcs are unary. | none |
SERVICE_NAMES_API_SUFFIX | Verifies that all service names end with `API`. | uber2 |
SERVICE_NAMES_CAMEL_CASE | Verifies that all service names are CamelCase. | google, uber1, uber2 |
SERVICE_NAMES_CAPITALIZED | Verifies that all service names are Capitalized. | google, uber1, uber2 |
SERVICE_NAMES_MATCH_FILE_NAME | Verifies that there is one service per file and the file name is of the pattern `service_name_lower_snake_case.proto`. | uber2 |
SERVICE_NAMES_NO_PLURALS | Verifies that all CamelCase service names do not contain plural components. | none |
SERVICES_HAVE_COMMENTS | Verifies that all services have a comment of the form `// ServiceName ...`. | none |
SERVICES_HAVE_SENTENCE_COMMENTS | Verifies that all services have a comment that contains at least one complete sentence. | uber2 |
SYNTAX_PROTO3 | Verifies that the syntax is `proto3`. | uber1, uber2 |
WKT_DIRECTLY_IMPORTED | Verifies that the Well-Known Types are directly imported using `google/protobuf/` as the base of the import. | uber1, uber2 |
WKT_DURATION_SUFFIX | Verifies that all field names of type `google.protobuf.Duration` are named `duration` or end in `_duration`. | uber2 |
WKT_TIMESTAMP_SUFFIX | Verifies that all field names of type `google.protobuf.Timestamp` are named `time` or end in `_time_`. | uber2 |
