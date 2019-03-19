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
