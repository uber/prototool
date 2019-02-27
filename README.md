# prototool [![Mit License][mit-img]][mit] [![GitHub Release][release-img]][release] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Docker Image][docker-img]][docker]

[Protobuf](https://developers.google.com/protocol-buffers) is one of the best interface description languages out there - it's widely adopted, and after over 15 years of use, it's practically bulletproof. However, working with Protobuf and maintaining consistency across your Protobuf files can be a pain - protoc, while being a tool that has stood the test of time, is non-trivial to use, and the Protobuf community has not developed common standards with regards to stub generation. Prototool aims to solve this by making working with Protobuf much simpler.

Prototool lets you:

- Handle installation of `protoc` and the import of all of the Well-Known Types behind the scenes in a platform-independent manner without any work on the part of the user.
- Standardize building of your Protobuf files with a common [configuration](#configuration), abstracting away all of the pain of protoc for you.
- [Lint](#prototool-lint) your Protobuf files with common linting rules according to [Google' Style Guide](https://developers.google.com/protocol-buffers/docs/style), [Uber's V1 Style Guide](https://github.com/uber/prototool/blob/master/etc/style/uber1/uber1.proto), or your own set of configured lint rules.
- [Format](#prototool-format) your Protobuf files in a consistent manner.
- [Create](#prototool-create) Protobuf files from a template that passes lint, taking care of package naming for you.
- [Generate](#prototool-generate) stubs using any plugin based on a simple configuration file, including handling imports of all the Well-Known Types.
- Call [gRPC](#prototool-grpc) endpoints with ease, taking care of the JSON to binary conversion for you.
- Output errors and lint failures in a common `file:line:column:message` format, making integration with editors possible, [Vim integration](#vim-integration) is provided out of the box.

Prototool accomplishes this by downloading and calling `protoc` on the fly for you, handing error messages from `protoc` and your plugins, and using the generated `FileDescriptorSets` for internal functionality, as well as wrapping a few great external libraries already in the Protobuf ecosystem. Compiling, linting and formatting commands run in around 3/100ths of second for a single Protobuf file, or under a second for a larger number (500+) of Protobuf files.

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
    * [prototool grpc](#prototool-grpc)
  * [gRPC Example](#grpc-example)
  * [Tips and Tricks](#tips-and-tricks)
  * [Vim Integration](#vim-integration)
  * [Stability](#stability)
  * [Development](#development)
  * [FAQ](#faq)
    * [Pre-Cache Protoc](#pre-cache-protoc)
    * [Alpine Linux Issues](#alpine-linux-issues)
    * [Managing External Plugins/Docker](#managing-external-pluginsdocker)
    * [Lint/Format Choices](#lintformat-choices)
  * [Special Thanks](#special-thanks)

The [docs](docs) folder has additional documentation on specific topics and is referenced below when
discussing a topic that has further instructions.

## Installation

Prototool can be installed on Mac OS X or Linux through a variety of methods.

*See [docs/install.md](docs/install.md) for full instructions.*

## Quick Start

We'll start with a general overview of the commands. There are more commands, and we will get into usage below, but this shows the basic functionality.

```bash
prototool help
prototool lint path/to/foo.proto path/to/bar.proto # file mode, specify multiple specific files
prototool lint idl/uber # directory mode, search for all .proto files recursively, obeying exclude_paths in prototool.yaml or prototool.json files
prototool lint # same as "prototool lint .", by default the current directory is used in directory mode
prototool create foo.proto # create the file foo.proto from a template that passes lint
prototool files idl/uber # list the files that will be used after applying exclude_paths from corresponding prototool.yaml or prototool.json files
prototool lint --list-linters # list all current lint rules being used
prototool lint --list-all-lint-groups # list all available lint groups, currently "google" and "uber"
prototool compile idl/uber # make sure all .proto files in idl/uber compile, but do not generate stubs
prototool generate idl/uber # generate stubs, see the generation directives in the config file example
prototool grpc idl/uber --address 0.0.0.0:8080 --method foo.ExcitedService/Exclamation --data '{"value":"hello"}' # call the foo.ExcitedService method Exclamation with the given data on 0.0.0.0:8080
```

## Full Example

See the [example](example) directory.

The make command `make example` runs prototool while installing the necessary plugins.

## Configuration

Prototool operates using a config file named either `prototool.yaml` or `prototool.json`. Only one of `prototool.yaml` or `prototool.json` can exist in a given directory. For non-trivial use, you should have a config file checked in to at least the root of your repository. It is important because the directory of an associated config file is passed to `protoc` as an include directory with `-I`, so this is the logical location your Protobuf file imports should start from.

Recommended base config file:

```yaml
protoc:
  version: 3.6.1
```

The command `prototool config init` will generate a config file in the current directory with all available configuration options commented out except `protoc.version`. See [etc/config/example/prototool.yaml](etc/config/example/prototool.yaml) for the config file that `prototool config init --uncomment` generates.

When specifying a directory or set of files for Prototool to operate on, Prototool will search for config files for each directory starting at the given path, and going up a directory until hitting root. If no config file is found, Prototool will use default values and operate as if there was a config file in the current directory, including the current directory with `-I` to `protoc`.

If multiple `prototool.yaml` or `prototool.json` files are found that match the input directory or files, an error will be returned.

## File Discovery

In most Prototool commands, you will see help along the following lines:

```bash
$ prototool help lint
Lint proto files and compile with protoc to check for failures.

Usage:
  prototool lint [dirOrFile] [flags]
```

`dirOrFile` can take two forms:

- You can specify exactly one directory. If this is done, Prototool goes up until it finds a `prototool.yaml` or `prototool.json` file (or uses the current directory if none is found), and then uses this config for all `.proto` files under the given directory recursively, except for files in the `excludes` lists in `prototool.yaml` or `prototool.json` files.
- You can specify exactly one file. This has the effect as if you specified the directory of this file (using the logic above), but errors are only printed for that file. This is useful for e.g. Vim integration.
- You can specify nothing. This has the effect as if you specified the current directory as the directory.

The idea with "directory builds" is that you often need more than just one file to do a `protoc` call, for example if you have types in other files in the same package that are not referenced by their fully-qualified name, and/or if you need to know what directories to specify with `-I` to `protoc` (by default, the directory of the `prototool.yaml` or `prototool.json` file is used).

## Command Overview

Let's go over some of the basic commands.

##### `prototool config init`

Create a `prototool.yaml` file in the current directory, with all options except `protoc.version` commented out.

##### `prototool compile`

Compile your Protobuf files, but do not generate stubs. This has the effect of calling `protoc` with `-o /dev/null`.

##### `prototool generate`

Compile your Protobuf files and generate stubs according to the rules in your `prototool.yaml` or `prototool.json` file.

See [etc/config/example/prototool.yaml](etc/config/example/prototool.yaml) for all available options. There are special
options available for Golang plugins, and plugins that output a single file instead of a set of files. Specifically, you
can output a single JAR for the built-in `protoc` `java` plugin, and you can output a file with the serialized
`FileDescriptorSet` using the built-in `protoc` `descriptor_set` plugin, optionally also calling `--include_imports`
and/or `--include_source_info`.

See [example/idl/uber/prototool.yaml](example/idl/uber/prototool.yaml) for a full example.

##### `prototool lint`

Lint your Protobuf files. Lint rules can be set using the configuration file. See the configuration at [etc/config/example/prototool.yaml](../etc/config/example/prototool.yaml) for all available options.
There are three pre-configured groups of rules: `google`, `uber1`, and `uber2`.

*See [docs/lint.md](docs/lint.md) for full instructions.*

##### `prototool format`

Format a Protobuf file and print the formatted file to stdout. There are flags to perform different actions:

- `-d` Write a diff instead.
- `-f` Fix the file according to the Style Guide.
- `-l` Write a lint error in the form file:line:column:message if a file is unformatted.
- `-w` Overwrite the existing file instead.

Concretely, the `-f` flag can be used so that the values for `java_multiple_files`, `java_outer_classname`, and `java_package` are updated to reflect what is expected by the
[Google Cloud APIs file structure](https://cloud.google.com/apis/design/file_structure), and the value of `go_package` is updated to reflect what we expect for the
Uber Style Guide. By formatting, the linting for these values will pass by default. See the documentation below for `prototool create` for an example.

##### `prototool create`

Create Protobuf files from a template. With the provided Vim integration, this will automatically create new files
that pass lint when a new file is opened.

*See [docs/create.md](docs/create.md) for full instructions.*

##### `prototool files`

Print the list of all files that will be used given the input `dirOrFile`. Useful for debugging.

##### `prototool grpc`

Call a gRPC endpoint using a JSON input. What this does behind the scenes:

- Compiles your Protobuf files with `protoc`, generating a `FileDescriptorSet`.
- Uses the `FileDescriptorSet` to figure out the request and response type for the endpoint, and to convert the JSON input to binary.
- Calls the gRPC endpoint.
- Uses the `FileDescriptorSet` to convert the resulting binary back to JSON, and prints it out for you.

*See [docs/grpc.md](docs/grpc.md) for full instructions.*

## Tips and Tricks

Prototool is meant to help enforce a consistent development style for Protobuf, and as such you should follow some basic rules:

- Have all your imports start from the directory your `prototool.yaml` or `prototool.json` file is in. While there is a configuration option `protoc.includes` to denote extra include directories, this is not recommended.
- Have all Protobuf files in the same directory use the same `package`, and use the same values for `go_package`, `java_multiple_files`, `java_outer_classname`, and `java_package`.
- Do not use long-form `go_package` values, ie use `foopb`, not `github.com/bar/baz/foo;foopb`. This helps `prototool generate` do the best job.

## Vim Integration

This repository is a self-contained plugin for use with the [ALE Lint Engine](https://github.com/w0rp/ale). The Vim integration will currently compile, provide lint errors, do generation of your stubs, and format your files on save. It will also optionally create new files from a template when opened.

*See [docs/vim.md](docs/vim.md) for full instructions.*

## Stability

Prototool is generally available, and conforms to [SemVer](https://semver.org), so Prototool will not have any breaking changes on a given
major version, with some exceptions:

- The output of the formatter may change between minor versions. This has not happened yet, but we may change the format in the future to
  reflect things such as max line lengths.
- The breaking change detector may have additional checks added between minor versions, and therefore a change that might not have been
  breaking previously might become a breaking change.
- The `PACKAGE_NO_KEYWORDS` linter on the `uber2` lint group may have additional keywords added.
- The `SERVICE_NAMES_NO_PLURALS` linter on the `uber2` lint group ignores certain plurals such as "data". We may add additional ignored
  plurals in the future, so plurals that are not ignored now may be ignored later.

## Development

Prototool is under active development. If you want to help, here's some places to start:

- Try out `prototool` and file feature requests or bug reports.
- Submit PRs with any changes you'd like to see made.

We appreciate any input you have!

Before filing an issue or submitting a PR, make sure to review the [Issue Guidelines](https://github.com/uber/prototool/blob/dev/.github/ISSUE_TEMPLATE.md), and before submitting a PR, make sure to also review
the [PR Guidelines](https://github.com/uber/prototool/blob/dev/.github/PULL_REQUEST_TEMPLATE.md). The Issue Guidelines will show up in the description field when filing a new issue, and the PR guidelines will show up in the
description field when submitting a PR, but clear the description field of this pre-populated text once you've read it :-)

Note that development of Prototool will only work with Golang 1.12 or newer.

Before submitting a PR, make sure to:

- Run `make generate` to make sure there is no diff.
- Run `make` to make sure all tests pass. This is functionally equivalent to the tests run on CI.

The entire implementation is purposefully under the `internal` package to not expose any API for the time being.

## FAQ

##### Pre-Cache Protoc

*Question:* How do I download `protoc` ahead of time as part of a Docker build/CI pipeline?

*Answer*: `prototool cache update`.

You can pass both `--cache-path` and `--config-data` flags to this command to customize the invocation.

```bash
# Basic invocation which will cache using the default behavior. See prototool help cache update for more details.
prototool cache update
# Cache to a specific directory path/to/cache
prototool cache update --cache-path path/to/cache
# Cache using custom configuration data instead of finding a prototool.yaml file using the file discovery mechanism
prototool cache update --config-data '{"protoc":{"version":"3.6.1"}}'
```

There is also a command `prototool cache delete` which will delete all cached assets of `prototool`,
however this command does not accept the `--cache-path` flag - if you specify a custom directory, you
should clean it up on your own, we don't want to effectively call `rm -rf DIR` via a `prototool` command
on a location we don't know about.

##### Alpine Linux Issues

*Question:* Help! Prototool is failing when I use it within a Docker image based on Alpine Linux!

*Answer:* `apk add libc6-compat`

`protoc` is not statically compiled, and adding this package fixes the problem.

##### Managing External Plugins/Docker

*Question:* Can Prototool manage my external plugins such as protoc-gen-go?

*Answer:* Unfortunately, no. This was an explicit design decision - Prototool is not meant to "know the world", instead
Prototool just takes care of what it is good at (managing your Protobuf build) to keep Prototool simple, leaving you to do
external plugin management. Prototool does provide the ability to use the "built-in" output directives `cpp, csharp, java, js, objc, php, python, ruby`
provided by `protoc` out of the box, however.

If you want to have a consistent build environment for external plugins, we recommend creating a Docker image. We provide
a basic Docker image at [hub.docker.com/r/uber/prototool](https://hub.docker.com/r/uber/prototool), defined in the [Dockerfile](Dockerfile)
within this repository.

*See [docks/docker.md](docs/docker.md) for more details.*

##### Lint/Format Choices

*Question:* I don't like some of the choices made in the Style Guide and that are enforced by default by the linter and/or I don't like
the choices made in the formatter. Can we change some things?

*Answer:* Sorry, but we can't - The goal of Prototool is to provide a straightforward Style Guide and consistent formatting that minimizes various issues that arise from Protobuf usage across large organizations. There are pros and cons to many of the choices in the Style Guide, but it's our belief that the best answer is a **single** answer, sometimes regardless of what that single answer is.

We do have multiple lint groups available, see the help section on `prototool lint` above.

It is possible to ignore lint rules via configuration. However, especially if starting from a clean slate, we highly recommend using all default lint rules for consistency.

Many of the lint rules exist to mitigate backwards compatibility problems as schemas evolves. For example: requiring a unique request-response pair per RPC - while this potentially resuls in duplicated messages, this makes it impossible to affect an adjacent RPC by adding or modifying an existing field.

## Special Thanks

Prototool uses some external libraries that deserve special mention and thanks for their contribution to Prototool's functionality:

- [github.com/emicklei/proto](https://github.com/emicklei/proto) - The Golang Protobuf parsing library that started it all, and is still used for the linting and formatting functionality. We can't thank Ernest Micklei enough for his help and putting up with all the [filed issues](https://github.com/emicklei/proto/issues?q=is%3Aissue+is%3Aclosed).
- [github.com/jhump/protoreflect](https://github.com/jhump/protoreflect) - Used for the JSON to binary and back conversion. Josh Humphries is an amazing developer, thank you so much.
- [github.com/fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl) - Still used for the gRPC functionality. Again a thank you to Josh Humphries and the team over at FullStory for their work.

[mit-img]: http://img.shields.io/badge/License-MIT-blue.svg
[mit]: https://github.com/uber/prototool/blob/master/LICENSE

[release-img]: https://img.shields.io/github/release/uber/prototool/all.svg
[release]: https://github.com/uber/prototool/releases

[ci-img]: https://img.shields.io/travis/uber/prototool/dev.svg
[ci]: https://travis-ci.org/uber/prototool/builds

[cov-img]: https://codecov.io/gh/uber/prototool/branch/dev/graph/badge.svg
[cov]: https://codecov.io/gh/uber/prototool/branch/dev

[docker-img]: https://img.shields.io/docker/pulls/uber/prototool.svg
[docker]: https://hub.docker.com/r/uber/prototool
