# Create

[Back to README.md](README.md)

`prototool create` creates Protobuf files from a template. With the provided Vim integration, this
will automatically create new files that pass lint when a new file is opened.

Assuming the filename `example_create_file.proto`, the file will look like the following:

```proto
syntax = "proto3";

package some.pkg;

option go_package = "pkgpb";
option java_multiple_files = true;
option java_outer_classname = "ExampleCreateFileProto";
option java_package = "com.some.pkg.pb";
```

If using the [uber2 lint group](lint.md), the file will look like the following:

```proto
syntax = "proto3";

package some.pkg.v1;

option csharp_namespace = "Some.Pkg.V1"
option go_package = "pkgv1";
option java_multiple_files = true;
option java_outer_classname = "ExampleCreateFileProto";
option java_package = "com.some.pkg.v1";
option objc_class_prefix = "SPX";
option php_namespace = "Some\\Pkg\\V1";
```

This matches what the linter expects, and closely matches the
[Google Cloud APIs File Structure](https://cloud.google.com/apis/design/file_structure).

The package `some.pkg` will be computed as follows:

- If `--package` is specified, `some.pkg` will be the value passed to `--package`.
- Otherwise, if there is no `prototool.yaml` or `prototool.json` that would apply to the new file,
  use `uber.prototool.generated`, or `uber.prototool.generated.v1` if using the `uber2` lint group.
- Otherwise, if there is a `prototool.yaml` or `prototool.json` file, check if it has a `packages`
  setting under the `create` section (see
  [etc/config/example/prototool.yaml](../etc/config/example/prototool.yaml) for an example).
  If it does, this package, concatenated with the relative path from the directory with the
  `prototool.yaml` or `prototool.json` file will be used.
- Otherwise, if there is no `packages` directive, just use the relative path from the directory
  with the `prototool.yaml` or `prototool.json` file. If the file is in the same directory as the
  `prototool.yaml` file, use `uber.prototool.generated` or `uber.prototool.generated.v1`.

For example, assume you have the following file at `repo/prototool.yaml`:

```yaml
create:
  packages:
    - directory: idl
      name: uber
    - directory: idl/baz
      name: special
```

- `prototool create repo/idl/foo/bar/bar.proto` will have the package `uber.foo.bar`.
- `prototool create repo/idl/bar.proto` will have the package `uber`.
- `prototool create repo/idl/baz/baz.proto` will have the package `special`.
- `prototool create repo/idl/baz/bat/bat.proto` will have the package `special.bat`.
- `prototool create repo/another/dir/bar.proto` will have the package `another.dir`.
- `prototool create repo/bar.proto` will have the package `uber.prototool.generated`.

This is meant to mimic what you generally want - a base package for your idl directory, followed
by packages matching the directory structure.

Note you can override the directory that the `prototool.yaml` or `prototool.json` file is in as
well. If we update our file at `repo/prototool.yaml` to this:

```yaml
create:
  packages:
    - directory: .
      name: foo.bar
```

Then `prototool create repo/bar.proto` will have the package `foo.bar`, and
`prototool create repo/another/dir/bar.proto` will have the package `foo.bar.another.dir`.
