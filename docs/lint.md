# Lint

`prototool lint` lints your Protobuf files.

Lint rules can be set using the configuration file. See the configuration at [etc/config/example/prototool.yaml](../etc/config/example/prototool.yaml) for all available options. There are two pre-configured groups of rules:

- `google`: This lint group follows the Style Guide at https://developers.google.com/protocol-buffers/docs/style. This is a small group of rules meant to enforce basic naming, and is widely followed. The style guide is copied to [etc/style/google/google.proto](../etc/style/google/google.proto).
- `uber1`: This lint group follows the V1 Style Guide at [etc/style/uber1/uber1.proto](../etc/style/uber1/uber1.proto). This is a very strict rule group and is meant to enforce consistent development patterns.
- `uber2`: This lint group is the V2 Style Guide, and makes some modifcations to more closely follow the Google Cloud APIs file
  structure, as well as adding even more rules to enforce more consistent development patterns. This lint group is under development.

To see the differences between lint groups, use the `--diff-lint-groups` flag:

```
prototool lint --diff-lint-groups google,uber2
```

Configuration of your group can be done by setting the `lint.group` option in your `prototool.yaml` file:

```yaml
lint:
  group: google
```

See the `prototool.yaml` files at [etc/style/google/prototool.yaml](../etc/style/google/prototool.yaml) and
[etc/style/uber1/prototool.yaml](../etc/style/uber1/prototool.yaml) for examples.

The `uber` lint group represents the default lint group, and will be used if no lint group is configured.

Linting also understands the concept of file headers, typically license headers. To specify a license header, add the following to your
`prototool.yaml`:

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

The `path` option specifies the path to the file that contains the header data. The `content` option specifies the content directly.
Only one of these can be specified. The `is_commented` option specifies whether the header data is already commented, and if not,
`// ` will be added before all non-empty lines, and `//` will be added before all empty lines. `is_commented` is optional and
generally will not be set if the file is not commented, for example if `path` points to a text LICENSE file.

If `lint.file_header.path` or `lint.file_header.content` is set, `prototool create`, `prototool format --fix`, and `prototool lint` will all take the file header into account.

See [internal/cmd/testdata/lint](../internal/cmd/testdata/lint) for additional examples of configurations, and run `prototool lint internal/cmd/testdata/lint/DIR` from a checkout of this repository to see example failures.

Files must be valid Protobuf that can be compiled with `protoc`, so prior to linting, `prototool lint` will compile your using `protoc`.
Note, however, this is very fast - for the two files in [etc/style/uber1](../etc/style/uber1), compiling and linting only takes approximately
3/100ths of a second:

```bash
$ time prototool lint etc/style/uber1

real	0m0.037s
user	0m0.026s
sys	0m0.017s
```

For all 694 Protobuf files currently in [googleapis](https://github.com/googleapis/googleapis), this takes approximately 3/4ths of a second:

```bash
$ cat prototool.yaml
protoc:
  allow_unused_imports: true
lint:
  group: google

$ time prototool lint .

real	0m0.734s
user	0m3.835s
sys	0m0.924s
```

