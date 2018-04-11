# prototool [![Mit License][mit-img]][mit] [![Build Status][ci-img]][ci]

[Protobuf](https://developers.google.com/protocol-buffers) is one of the best interface description languages out there - it's widely adopted, and after over 15 years of use, it's practically bulletproof. However, working with Protobuf and maintaining consistency across your Protobuf files can be a pain - protoc, while being a tool that has stood the test of time, is non-trivial to use, and the Protobuf community has not developed common standards with regards to stub generation. Prototool aims to solve this by making working with Protobuf much simpler.

Prototool lets you:

- Handle installation of protoc and the import of all of the Well-Known Types behind the scenes in a platform-independent manner without any work on the part of the user.
- Standardize building of your Protobuf files with a common configuration, abstracting away all of the pain of protoc for you.
- Lint your Protobuf files with common linting rules according to a Style Guide.
- Format your Protobuf files in a consistent manner.
- Generate stubs using any plugin based on a simple configuration file, including handling imports of all the Well-Known Types.
- Call gRPC endpoints with ease, taking care of the JSON to binary conversion for you.
- Output errors and lint failures in a common file:line:column:message format, making integration with editors possible, Vim integration is provided out of the box.

Prototool accomplishes this by downloading and calling protoc on the fly for you, handing error messages from protoc and your plugins, and using the generated FileDescriptorSets for internal functionality, as well as wrapping a few great external libraries already in the Protobuf ecosystem.

  * [Current Status](#current-status)
  * [Installation](#installation)
  * [Quick Start](#quick-start)
  * [Full Example](#full-example)
  * [Configuration](#configuration)
  * [File Discovery](#file-discovery)
  * [Command Overview](#command-overview)
    * [prototool init](#prototool-init)
    * [prototool compile](#prototool-compile)
    * [prototool gen](#prototool-gen)
    * [prototool lint](#prototool-lint)
    * [prototool format](#prototool-format)
    * [prototool files](#prototool-files)
    * [prototool protoc-commands](#prototool-protoc-commands)
    * [prototool grpc](#prototool-grpc)
  * [gRPC Example](#grpc-example)
  * [Tips and Tricks](#tips-and-tricks)
  * [Vim Integration](#vim-integration)
  * [Development](#development)
  * [Special Thanks](#special-thanks)

## Current Status

Prototool is stil in the early alpha stages, and should not be used in production yet. Expect significant breaking changes before the v1.0 release. To help with development, head to the [Development](#development) section and follow along!

## Installation

Install Prototool from GitHub Releases.

```
curl -sSL https://github.com/uber/prototool/releases/download/v0.1.0/prototool-$(uname -s)-$(uname -m) \
  -o /usr/local/bin/prototool && \
  chmod +x /usr/local/bin/prototool && \
  prototool -h
```

## Quick Start

We'll start with a general overview of the commands. There are more commands, and we will get into usage below, but this shows the basic functionality.

```
prototool help
prototool lint path/to/foo.proto path/to/bar.proto # file mode, specify multiple specific files
prototool lint idl/uber # directory mode, search for all .proto files recursively, obeying exclude_paths in prototool.yaml files
prototool lint # same as "prototool lint .", by default the current directory is used in directory mode
prototool files idl/uber # list the files that will be used after applying exclude_paths from corresponding prototool.yaml files
prototool list-linters # list all current lint rules being used
prototool compile idl/uber # make sure all .proto files in idl/uber compile, but do not generate stubs
prototool gen idl/uber # generate stubs, see the generation directives in the config file example
prototool protoc-commands idl/uber # print out the protoc commands that would be invoked with prototool compile idl/uber
prototool protoc-commands --gen idl/uber # print out the protoc commands that would be invoked with prototool gen idl/uber
prototool grpc idl/uber 0.0.0.0:8080 foo.ExcitedService/Exclamation '{"value":"hello"}' # call the foo.ExcitedService method Exclamation with the given data on 0.0.0.0:8080
cd $(prototool download) # download prints out the cached protoc dir, so this changes to the cache directory
```

## Full Example

See the [example](example) directory.

The make command `make example` runs prototool while installing the necessary plugins.

## Configuration

Prototool operates using a config file named `prototool.yaml`. For non-trivial use, you should have a config file checked in to at least the root of your repository. It is important because the directory of an associated config file is passed to `protoc` as an include directory with `-I`, so this is the logical location your Protobuf file imports should start from.

Recommended base config file:

```yaml
protoc_version: 3.5.1
```

The command `prototool init` will generate a config file in the current directory with all available configuration options commented out except `protoc_version`. See [etc/config/example/prototool.yaml](etc/config/example/prototool.yaml) for the config file that `prototool init --uncomment` generates.

When specifying a directory or set of files for Prototool to operate on, Prototool will search for config files for each directory starting at the given path, and going up a directory until hitting root. If no config file is found, Prototool will use default values and operate as if there was a config file in the current directory, including the current directory with `-I` to `protoc`.

While almost all projects should not have multiple `prototool.yaml` files (and this [may be enforced before v1.0](https://github.com/uber/prototool/issues/10)), as of now, multiple `prototool.yaml` files corresponding to multiple found directories with Protobuf files may be used. For example, if you have the following layout:

```
.
├── a
│   ├── d
│   │   ├── file.proto
│   │   ├── file2.proto
│   │   ├── file3.proto
│   │   └── prototool.yaml
│   ├── e
│   │   └── file.proto
│   ├── f
│   │   └── file.proto
│   └── file.proto
├── b
│   ├── file.proto
│   ├── g
│   │   └── h
│   │       └── file.proto
│   └── prototool.yaml
├── c
│   ├── file.proto
│   └── i
│       └── file.proto
└── prototool.yaml
```

Everything under `a/d` will use `a/d/prototool.yaml`, everything under `b`, `b/g/h` will use `b/prototool.yaml`, and everything under `a`, `a/e`, `a/f`, `c`, `c/i` will use `prototool.yaml`. See [internal/x/file/testdata](internal/x/file/testdata) for the most current example.

## File Discovery

In most Prototool commands, you will see help along the following lines:

```
$ prototool help lint
Lint proto files and compile with protoc to check for failures.

Usage:
  prototool lint dirOrProtoFiles... [flags]

Flags:
      --dir-mode   Run as if the directory the file was given, but only print the errors from the file. Useful for integration with editors.
```

`dirOrProtoFiles...` can take multiple forms:

- You can specify multiple files. If this is done, these files will be explicitly used for `protoc` calls.
- You can specify exactly one directory. If this is done, Prototool goes up until it finds a `prototool.yaml` file (or uses the current directory if none is found), and then walks starting at this location for all `.proto` files, and these are used, except for files in the `excludes` lists in `prototool.yaml` files.
- You can specify exactly one file, along with `--dir-mode`. This has the effect as if you specified the directory of this file (using the logic above), but errors are only printed for that file. This is useful for e.g. Vim integration.

The idea with "directory builds" is that you often need more than just one file to do a `protoc` call, for example if you have types in other files in the same package that are not referenced by their fully-qualified name, and/or if you need to know what directories to specify with `-I` to `protoc` (by default, the directory of the `prototool.yaml` file is used).

In general practice, directory builds are what you always want to do. File builds were just added for convenience, and [may be removed](https://github.com/uber/prototool/issues/16).

## Command Overview

Let's go over some of the basic commands. There are more commands than listed here, and [some may be removed before v1.0](https://github.com/uber/prototool/issues/11), but the following commands are what you mostly need to know.

##### `prototool init`

Create a `prototool.yaml` file in the current directory, with all options except `protoc_version` commented out.

##### `prototool compile`

Compile your Protobuf files, but do not generate stubs. This has the effect of calling `protoc` with `-o /dev/null`.

##### `prototool gen`

Compile your Protobuf files and generate stubs according to the rules in your `prototool.yaml` file. See [example/idl/uber/prototool.yaml](example/idl/uber/prototool.yaml) for an example.

##### `prototool lint`

Lint your Protobuf files. The default rule set follows the Style Guide at [etc/style/uber/uber.proto](etc/style/uber/uber.proto). You can add or exclude lint rules in your `prototool.yaml` file. The default rule set is "strict", and we are working on having two main sets of rules, as well as refining the Style Guide, in [this issue](https://github.com/uber/prototool/issues/3).

##### `prototool format`

Format a Protobuf file and print the formatted file to stdout. There are flags to perform different actions:

- `-d` Write a diff instead.
- `-l` Write a lint error in the form file:line:column:message if a file is unformatted.
- `-w` Overwrite the existing file instead.

##### `prototool files`

Print the list of all files that will be used given the input `dirOrProtoFiles...`. Useful for debugging.

##### `prototool protoc-commands`

Print all `protoc` commands that would be run on `prototool compile`. Add the `--gen` flag to print all commands that would be run on `prototool gen`.

##### `prototool grpc`

Call a gRPC endpoint using a JSON input. What this does behind the scenes:

- Compiles your Protobuf files with `protoc`, generating a `FileDescriptorSet`.
- Uses the `FileDescriptorSet` to figure out the request and response type for the endpoint, and to convert the JSON input to binary.
- Calls the gRPC endpoint.
- Uses the `FileDescriptorSet` to convert the resulting binary back to JSON, and prints it out for you.

All these steps take on the order of milliseconds, for example the overhead for a file with four dependencies is about 30ms, so there is little overhead for CLI calls to gRPC.

## gRPC Example

There is a full example for gRPC in the [example](example) directory. Run `make init example` to make sure everything is installed and generated.

Start the example server in a separate terminal by doing `go run example/cmd/excited/main.go`.

`prototool grpc dirOrProtoFiles... serverAddress package.service/Method requestData`

`requestData` can either be the JSON data to input, or `-` which will result in the input being read from stdin.

```
$ make init example # make sure everything is built just in case

$ cat input.json
{"value":"hello"}

$ cat input.json | prototool grpc example 0.0.0.0:8080 foo.ExcitedService/Exclamation -
{
  "value": "hello!"
}

$ cat input.json | prototool grpc example 0.0.0.0:8080 foo.ExcitedService/ExclamationServerStream -
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

$ cat input.json
{"value":"hello"}
{"value":"salutations"}

$ cat input.json | prototool grpc example 0.0.0.0:8080 foo.ExcitedService/ExclamationClientStream -
{
  "value": "hellosalutations!"
}

$ cat input.json | prototool grpc example 0.0.0.0:8080 foo.ExcitedService/ExclamationBidiStream -
{
  "value": "hello!"
}
{
  "value": "salutations!"
}
```

## Tips and Tricks

Prototool is meant to help enforce a consistent development style for Protobuf, and as such you should follow some basic rules:

- Have all your imports start from the directory your `prototool.yaml` is in. While there is a configuration option `protoc_includes` to denote extra include directories, this is not recommended.
- Have all Protobuf files in the same directory use the same `package`, and use the same values for `go_package` and `java_package`.
- Do not use long-form `go_package` values, ie use `foopb`, not `github.com/bar/baz/foo;foopb`. This helps `prototool gen` do the best job.

## Vim Integration

This repository is a self-contained plugin for use with the [ALE Lint Engine](https://github.com/w0rp/ale). It should be similarly easy to add support for Syntastic, Neomake, etc later.

The Vim integration will currently provide lint errors, optionally regenerate all the stubs, and optionally format your files on save.

The plugin is under [vim/prototool](vim/prototool), so your plugin manager needs to point there instead of the base of this repository. Assuming you are using Vundle, copy/paste [etc/vim/example/vimrc](etc/vim/example/vimrc) into your vimrc and you should be good to go.

Editor integration is a key goal of Prototool. We've demonstrated support internally for Intellij, and hope that we have integration for more editors in the future.

## Development

Prototool is under active development, if you want to help, here's some places to start:

- Try out `prototool` and file issues, including points that are unclear in the documentation.
- Put up PRs with any changes you'd like to see made. We can't guarantee that many PRs will get merged for now, but we appreciate any input!
- Follow along on the [big ticket items](https://github.com/uber/prototool/issues?q=is%3Aissue+is%3Aopen+label%3A%22big+ticket+item%22) to see the major development points.

Over the coming months, we hope to push to a v1.0.

A note on package layout: all Golang code except for `cmd/prototool/main.go` is purposefully under the `internal` package to not expose any API for the time being. Within the internal package, anything under `internal/x` has not been reviewed, and is especially unstable. Any package in `internal` not in `internal/x` has been fully reviewed and is more stable.

## Special Thanks

Prototool uses some external libraries that deserve special mention and thanks for their contribution to Prototool's functionality:

- [github.com/emicklei/proto](https://github.com/emicklei/proto) - The Golang Protobuf parsing library that started it all, and is still used for the linting and formatting functionality. We can't thank Ernest Micklei enough for his help and putting up with all the [filed issues](https://github.com/emicklei/proto/issues?q=is%3Aissue+is%3Aclosed).
- [github.com/jhump/protoreflect](https://github.com/jhump/protoreflect) - Used for the JSON to binary and back conversion. Josh Humphries is an amazing developer, thank you so much.
- [github.com/fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl) - Still used for the gRPC functionality. Again a thank you to Josh Humphries and the team over at FullStory for their work.

[mit-img]: http://img.shields.io/badge/License-MIT-blue.svg
[mit]: https://github.com/uber/prototool/blob/master/LICENSE

[ci-img]: https://img.shields.io/travis/uber/prototool/dev.svg
[ci]: https://travis-ci.org/uber/prototool/builds
