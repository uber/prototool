# Prototool

[![MIT License][mit-img]][mit] [![GitHub Release][release-img]][release] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Docker Image][docker-img]][docker] [![Homebrew Package][homebrew-img]][homebrew] [![AUR Package][aur-img]][aur]

**Update:  We recommend checking out [Buf](https://github.com/bufbuild/buf), which is under active development.  There are a ton of docs for getting started, including for [migration from Prototool](https://buf.build/docs/migration-prototool).**


[Protobuf](https://developers.google.com/protocol-buffers) is one of the best interface description
languages out there - it's widely adopted, and after over 15 years of use, it's practically
bulletproof. However, working with Protobuf and maintaining consistency across your Protobuf files
can be a pain - `protoc`, while being a tool that has stood the test of time, is non-trivial to
use, and the Protobuf community has not developed common standards with regards to stub generation.
Prototool aims to solve this by making working with Protobuf much simpler.

Prototool lets you:

- Handle installation of `protoc` and the import of all of the Well-Known Types behind the scenes
  in a platform-independent manner.
- Standardize building of your Protobuf files with a common [configuration](#configuration).
- [Lint](#prototool-lint) your Protobuf files with common linting rules according to
  [Google' Style Guide](https://developers.google.com/protocol-buffers/docs/style),
  [Uber's V1 Style Guide](../etc/style/uber1/uber1.proto),
  [Uber's V2 Style Guide](../style/README.md), or your own set of configured lint rules.
- [Format](#prototool-format) your Protobuf files in a consistent manner.
- [Create](#prototool-create) Protobuf files from a template that passes lint, taking care of
  package naming for you.
- [Generate](#prototool-generate) stubs using any plugin based on a simple configuration file,
  including handling imports of all the Well-Known Types.
- Call [gRPC](#prototool-grpc) endpoints with ease, taking care of the JSON to binary
  conversion for you.
- Check for [breaking changes](#prototool-break-check) on a per-package basis, verifying that your
  API never breaks.
- Output errors and lint failures in a common `file:line:column:message` format, making integration
  with editors possible, [Vim integration](#vim-integration) is provided out of the box.

Prototool accomplishes this by downloading and calling `protoc` on the fly for you, handing error
messages from `protoc` and your plugins, and using the generated `FileDescriptorSets` for internal
functionality, as well as wrapping a few great external libraries already in the Protobuf
ecosystem. Compiling, linting and formatting commands run in around 3/100ths of second for a single
Protobuf file, or under a second for a larger number (500+) of Protobuf files.

## Table Of Contents

  * [Installation](#installation)
  * [Quick Start](#quick-start)
  * [Full Example](#full-example)
  * [Configuration](#configuration)
  * [File Discovery](#file-discovery)
  * [Command Overview](#command-overview)
    * [prototool config init](#prototool-config-init)
    * [prototool compile](#prototool-compile)
    * [prototool generate](#prototool-generate)
    * [prototool lint](#prototool-lint)
    * [prototool format](#prototool-format)
    * [prototool create](#prototool-create)
    * [prototool files](#prototool-files)
    * [prototool break check](#prototool-break-check)
    * [prototool descriptor-set](#prototool-descriptor-set)
    * [prototool grpc](#prototool-grpc)
  * [Tips and Tricks](#tips-and-tricks)
  * [Vim Integration](#vim-integration)
  * [Stability](#stability)
  * [Development](#development)
  * [FAQ](#faq)
  * [Special Thanks](#special-thanks)

## Installation

Prototool can be installed on Mac OS X or Linux through a variety of methods.

*See [install.md](install.md) for full instructions.*

## Quick Start

We'll start with a general overview of the commands. There are more commands, and we will get into]
usage below, but this shows the basic functionality.

```bash
prototool help
prototool lint idl/uber # search for all .proto files recursively, obeying exclude_paths in prototool.yaml or prototool.json files
prototool lint # same as "prototool lint .", by default the current directory is used in directory mode
prototool create foo.proto # create the file foo.proto from a template that passes lint
prototool files idl/uber # list the files that will be used after applying exclude_paths from corresponding prototool.yaml or prototool.json files
prototool lint --list-linters # list all current lint rules being used
prototool lint --list-all-lint-groups # list all available lint groups, currently "google" and "uber"
prototool compile idl/uber # make sure all .proto files in idl/uber compile, but do not generate stubs
prototool generate idl/uber # generate stubs, see the generation directives in the config file example
prototool grpc idl/uber --address 0.0.0.0:8080 --method foo.ExcitedService/Exclamation --data '{"value":"hello"}' # call the foo.ExcitedService method Exclamation with the given data on 0.0.0.0:8080
prototool descriptor-set --include-imports idl/uber # generate a FileDescriptorSet for all files under idl/uber, outputting to stdout, a given file, or a temporary file
prototool break check idl/uber --git-branch master # check for breaking changes as compared to the Protobuf definitions in idl/uber on the master branch
```

## Full Example

See the [example](../example) directory.

The make command `make example` runs prototool while installing the necessary plugins.

## Configuration

Prototool operates using a config file named either `prototool.yaml` or `prototool.json`. Only one
of `prototool.yaml` or `prototool.json` can exist in a given directory. For non-trivial use, you
should have a config file checked in to at least the root of your repository. It is important
because the directory of an associated config file is passed to `protoc` as an include directory
with `-I`, so this is the logical location your Protobuf file imports should start from.

Recommended base config file:

```yaml
protoc:
  version: 3.8.0
lint:
  group: uber2
```

*See [protoc.md](protoc.md) for how Prototool handles working with `protoc`.*

The command `prototool config init` will generate a config file in the current directory with the
currently recommended options set.

When specifying a directory or set of files for Prototool to operate on, Prototool will search for
config files for each directory starting at the given path, and going up a directory until hitting
root. If no config file is found, Prototool will use default values and operate as if there was a
config file in the current directory, including the current directory with `-I` to `protoc`.

If multiple `prototool.yaml` or `prototool.json` files are found that match the input directory or
files, an error will be returned.

See [etc/config/example/prototool.yaml](../etc/config/example/prototool.yaml) all available
options.

## File Discovery

In most Prototool commands, you will see help along the following lines:

```bash
$ prototool help lint
Lint proto files and compile with protoc to check for failures.

Usage:
  prototool lint [dirOrFile] [flags]
```

`dirOrFile` can take two forms:

- You can specify exactly one directory. If this is done, Prototool goes up until it finds a
  `prototool.yaml` or `prototool.json` file (or uses the current directory if none is found), and
  then uses this config for all `.proto` files under the given directory recursively, except for
  files in the `excludes` lists in `prototool.yaml` or `prototool.json` files.
- You can specify exactly one file. This has the effect as if you specified the directory of this
  file (using the logic above), but errors are only printed for that file. This is useful for
  e.g. Vim integration.
- You can specify nothing. This has the effect as if you specified the current directory as the
  directory.

The idea with "directory builds" is that you often need more than just one file to do a `protoc`
call, for example if you have types in other files in the same package that are not referenced by
their fully-qualified name, and/or if you need to know what directories to specify with `-I` to
`protoc` (by default, the directory of the `prototool.yaml` or `prototool.json` file is used).

## Command Overview

Let's go over some of the basic commands.

##### `prototool config init`

Create a `prototool.yaml` file in the current directory with the currently recommended options set.

Pass the `--document` flag to generate a `prototool.yaml` file with all other options documented
and commented out.

Pass the `--uncomment` flag to generate `prototool.yaml` file with all options documented but
uncommented.

See [etc/config/example/prototool.yaml](../etc/config/example/prototool.yaml) for the config file
that `prototool config init --uncomment` generates.

##### `prototool compile`

Compile your Protobuf files, but do not generate stubs. This has the effect of calling `protoc`
with `-o /dev/null`.

Pass the `--dry-run` flag to see the `protoc` commands that Prototool runs behind the scenes.

##### `prototool generate`

Compile your Protobuf files and generate stubs according to the rules in your `prototool.yaml` or
`prototool.json` file.

See [etc/config/example/prototool.yaml](../etc/config/example/prototool.yaml) for all available
options. There are special options available for Golang plugins, and plugins that output a single
file instead of a set of files. Specifically, you can output a single JAR for the built-in `protoc`
`java` plugin, and you can output a file with the serialized `FileDescriptorSet` using the built-in
`protoc` `descriptor_set` plugin, optionally also calling `--include_imports` and/or
`--include_source_info`.

Pass the `--dry-run` flag to see the `protoc` commands that Prototool runs behind the scenes.

See [example/proto/prototool.yaml](../example/proto/prototool.yaml) for a full example.

##### `prototool lint`

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

The flag `--generate-ignores` will help with migrating to a given lint group by generating
the configuration to ignore existing lint failures on a per-file basis.

*See [lint.md](lint.md) for full instructions.*

##### `prototool format`

Format a Protobuf file and print the formatted file to stdout. There are flags to perform different
actions:

- `-d` Write a diff instead.
- `-f` Fix the file according to the Style Guide. This will have different behavior if the `uber2`
  lint group is set.
- `-l` Write a lint error in the form file:line:column:message if a file is unformatted.
- `-w` Overwrite the existing file instead.

##### `prototool create`

Create Protobuf files from a template. With the provided Vim integration, this will automatically
create new files that pass lint when a new file is opened.

*See [create.md](create.md) for full instructions.*

##### `prototool files`

Print the list of all files that will be used given the input `dirOrFile`. Useful for debugging.

##### `prototool break check`

Protobuf is a great way to represent your APIs and generate stubs in each language you develop
with. As such, Protobuf APIs should be stable so as not to break consumers across repositories.
Even in a monorepo context, making sure that your Protobuf APIs do not introduce breaking
changes is important so that different deployed versions of your services do not have
wire incompatibilities.

Prototool exposes a breaking change detector through the `prototool break check` command. This will
check your current Protobuf definitions against a past version of your Protobuf definitions to see
if there are any source or wire incompatible changes. Some notes on this command:

- The breaking change detection operates on a **per-package** basis, not per-file - definitions
  can be moved between files within the same Protobuf package without being considered breaking.
- The breaking change detector can either check against a given git branch or tag, or it can check
  against a previous state saved with the `prototool break descriptor-set` command.
- The breaking change detector understands the concept of **beta vs. stable packages**, discussed
  in the [V2 Style Guide](../style/README.md#package-versioning). By default, the breaking change
  detector will not check beta packages for breaking changes, and will not allow stable packages to
  depend on beta packages, however both of these options are configurable in your `prototool.yaml`
  file.

*See [breaking.md](breaking.md) for full instructions.*

##### `prototool descriptor-set`

Produce a serialized `FileDescriptorSet` for all Protobuf definitions. By default, the serialized
`FileDescriptorSet` is printed to stdout. There are a few options:

- `--include-imports, --include-source-info` are analagous to `protoc`'s `--include_imports,
  --include_source_info` flags.
- `--json` outputs the FileDescriptorSet as JSON instead of binary.
- `-o` writes the `FileDescriptorSet` to the given output file path.
- `--tmp` writes the `FileDescriptorset` to a temporary file and prints the file path.

The outputted `FileDescriptorSet` is a merge of all produced `FileDescriptorSets` for each
Protobuf package compiled.

This command is useful in a few situations.

One such situation is with external gRPC tools such as [grpcurl](https://github.com/fullstorydev/grpcurl)
or [ghz](https://ghz.sh). Both tools take a path to a serialized `FileDescriptorSet` for use to
figure out the request/response structure of RPCs when the gRPC reflection service is not available.
`prototool descriptor-set` can be used to generate these `FileDescriptorSet`s on the fly.

```bash
grpcurl -protoset $(prototool descriptor-set --include-imports --tmp) ...
ghz -protoset $(prototool descriptor-set --include-imports --tmp) ...
```

You can also just save the file once and not re-compile each time.

```bash
prototool descriptor-set --include-imports -o descriptor_set.bin
grpcurl -protoset descriptor_set.bin ...
ghz -protoset descriptor_set.bin ...
```

Another situation is to use `jq` to make arbitrary queries on your Protobuf definitions.

For example, if your Protobuf definitions are in `path/to/proto`, the following will print
all message names.

```bash
prototool descriptor-set path/to/proto --json | \
  jq '.file[] | select(.messageType != null) | .messageType[] | .name' | \
  sort | uniq
```

##### `prototool grpc`

Call a gRPC endpoint using a JSON input. What this does behind the scenes:

- Compiles your Protobuf files with `protoc`, generating a `FileDescriptorSet`.
- Uses the `FileDescriptorSet` to figure out the request and response type for the endpoint, and to
  convert the JSON input to binary.
- Calls the gRPC endpoint.
- Uses the `FileDescriptorSet` to convert the resulting binary back to JSON, and prints it out for
  you.

*See [grpc.md](grpc.md) for full instructions.*

## Tips and Tricks

Prototool is meant to help enforce a consistent development style for Protobuf, and as such you
should follow some basic rules:

- Have all your imports start from the directory your `prototool.yaml` or `prototool.json` file is
  in. While there is a configuration option `protoc.includes` to denote extra include directories,
  this is not recommended.
- Have all Protobuf files in the same directory use the same `package`.
- Do not use long-form `go_package` values, ie use `foopb`, not `github.com/bar/baz/foo;foopb`.
  This helps `prototool generate` do the best job.

## Vim Integration

This repository is a self-contained plugin for use with the
[ALE Lint Engine](https://github.com/w0rp/ale). The Vim integration will currently compile, provide
lint errors, do generation of your stubs, and format your files on save. It will also optionally
create new files from a template when opened.

*See [vim.md](vim.md) for full instructions.*

## Stability

Prototool is generally available, and conforms to [SemVer](https://semver.org), so Prototool will
not have any breaking changes on a given major version, with some exceptions:

- Commands under the `x` top-level command are experimental, and may change or be deleted between
  minor versions of Prototool. We expect such commands to be promoted to stable within a few minor
  releases, however development is still in-progress.
- The output of the formatter may change between minor versions. This has not happened yet, but we
  may change the format in the future to reflect things such as max line lengths.
- The breaking change detector's output format currently does not output filename, line, or column.
  This is an expected upgrade in the future, so the output will likely change. This is viewed as
  purely an upgrade, so until this is done, do not parse `prototool break check` output in scripts.
- The breaking change detector may have additional checks added between minor versions, and
  therefore a change that might not have been breaking previously might become a breaking change.
  This may become stable in the near future, and at this time we'll denote that no more checks
  will be added.

## Development

See [development.md](development.md) for concerns related to Prototool development.

See [maintenance.md](maintenance.md) for maintenance-related tasks.

## FAQ

See [faq.md](faq.md) for answers to frequently asked questions.

## Special Thanks

Prototool uses some external libraries that deserve special mention and thanks for their
contribution to Prototool's functionality:

- [github.com/emicklei/proto](https://github.com/emicklei/proto) - The Golang Protobuf parsing
  library that started it all, and is still used for the linting and formatting functionality. We
  can't thank Ernest Micklei enough for his help and putting up with all the
  [filed issues](https://github.com/emicklei/proto/issues?q=is%3Aissue+is%3Aclosed).
- [github.com/jhump/protoreflect](https://github.com/jhump/protoreflect) - Used for the JSON to
  binary and back conversion. Josh Humphries is an amazing developer, thank you so much.
- [github.com/fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl) - Still used for the
  gRPC functionality. Again a thank you to Josh Humphries and the team over at FullStory for their
  work.

[mit-img]: http://img.shields.io/badge/License-MIT-blue.svg
[mit]: https://github.com/uber/prototool/blob/master/LICENSE

[release-img]: https://img.shields.io/github/release/uber/prototool/all.svg
[release]: https://github.com/uber/prototool/releases

[ci-img]: https://img.shields.io/buildkite/5faf32c23003786e641b9140ee98175b81c8bae973ae188415/dev.svg
[ci]: https://buildkite.com/uberopensource/prototool

[cov-img]: https://codecov.io/gh/uber/prototool/branch/dev/graph/badge.svg
[cov]: https://codecov.io/gh/uber/prototool/branch/dev

[docker-img]: https://img.shields.io/docker/pulls/uber/prototool.svg
[docker]: https://hub.docker.com/r/uber/prototool

[homebrew-img]: https://img.shields.io/homebrew/v/prototool.svg
[homebrew]: https://formulae.brew.sh/formula/prototool

[aur-img]: https://img.shields.io/aur/version/prototool-bin.svg
[aur]: https://aur.archlinux.org/packages/prototool-bin
