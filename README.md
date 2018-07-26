# prototool [![Mit License][mit-img]][mit] [![GitHub Release][release-img]][release] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

[Protobuf](https://developers.google.com/protocol-buffers) is one of the best interface description languages out there - it's widely adopted, and after over 15 years of use, it's practically bulletproof. However, working with Protobuf and maintaining consistency across your Protobuf files can be a pain - protoc, while being a tool that has stood the test of time, is non-trivial to use, and the Protobuf community has not developed common standards with regards to stub generation. Prototool aims to solve this by making working with Protobuf much simpler.

Prototool lets you:

- Handle installation of protoc and the import of all of the Well-Known Types behind the scenes in a platform-independent manner without any work on the part of the user.
- Standardize building of your Protobuf files with a common configuration, abstracting away all of the pain of protoc for you.
- Lint your Protobuf files with common linting rules according to a Style Guide.
- Format your Protobuf files in a consistent manner.
- Generate Protobuf files from a template that passes lint, taking care of package naming for you.
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
    * [prototool create](#prototool-create)
    * [prototool files](#prototool-files)
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
curl -sSL https://github.com/uber/prototool/releases/download/v0.4.0/prototool-$(uname -s)-$(uname -m) \
  -o /usr/local/bin/prototool && \
  chmod +x /usr/local/bin/prototool
```

Prototool is purposefully all put under internal in an attempt to emphasize that you should not install `prototool` with `go get`.
Although you can technically `go get` from `internal`, we do not check in the vendor directory, so `go get` will not respect the
versions in `glide.yaml`. We have specific version requirements, so not using these can and probably will result in errors when building
`prototool` and/or result in unexpected runtime errors.

## Quick Start

We'll start with a general overview of the commands. There are more commands, and we will get into usage below, but this shows the basic functionality.

```
prototool help
prototool lint path/to/foo.proto path/to/bar.proto # file mode, specify multiple specific files
prototool lint idl/uber # directory mode, search for all .proto files recursively, obeying exclude_paths in prototool.yaml files
prototool lint # same as "prototool lint .", by default the current directory is used in directory mode
prototool create foo.proto # create the file foo.proto from a template that passes lint
prototool files idl/uber # list the files that will be used after applying exclude_paths from corresponding prototool.yaml files
prototool lint --list-linters # list all current lint rules being used
prototool compile idl/uber # make sure all .proto files in idl/uber compile, but do not generate stubs
prototool gen idl/uber # generate stubs, see the generation directives in the config file example
prototool grpc idl/uber --address 0.0.0.0:8080 --method foo.ExcitedService/Exclamation --data '{"value":"hello"}' # call the foo.ExcitedService method Exclamation with the given data on 0.0.0.0:8080
```

## Full Example

See the [example](example) directory.

The make command `make example` runs prototool while installing the necessary plugins.

## Configuration

Prototool operates using a config file named `prototool.yaml`. For non-trivial use, you should have a config file checked in to at least the root of your repository. It is important because the directory of an associated config file is passed to `protoc` as an include directory with `-I`, so this is the logical location your Protobuf file imports should start from.

Recommended base config file:

```yaml
protoc_version: 3.6.0
```

The command `prototool init` will generate a config file in the current directory with all available configuration options commented out except `protoc_version`. See [etc/config/example/prototool.yaml](etc/config/example/prototool.yaml) for the config file that `prototool init --uncomment` generates.

When specifying a directory or set of files for Prototool to operate on, Prototool will search for config files for each directory starting at the given path, and going up a directory until hitting root. If no config file is found, Prototool will use default values and operate as if there was a config file in the current directory, including the current directory with `-I` to `protoc`.

If multiple `prototool.yaml` files are found that match the input directory or files, an error will be returned. We have an ongoing discussion about whether to allow multiple `prototool.yaml` files, see [this issue](https://github.com/uber/prototool/issues/10) for more details.

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

By default, the values for `java_multiple_files`, `java_outer_classname`, and `java_package` are updated
to reflect what is expected by the [Google Cloud APIs file structure](https://cloud.google.com/apis/design/file_structure),
and the value of `go_package` is updated to reflect what we expect for the default Style Guide. By formatting, the linting for
these values will pass by default. See the documentation below for [prototool create](#prototool-create) for an example. This functionality
can be suppressed by passing the flag `--no-rewrite` to `prototool format`.

##### `prototool create`

Create a Protobuf file from a template that passes lint. Assuming the filename `example_create_file.proto`, the file will look like the following:

```proto
syntax = "proto3";

package SOME.PKG;

option go_package = "PKGpb";
option java_multiple_files = true;
option java_outer_classname = "ExampleCreateFileProto";
option java_package = "com.SOME.PKG.pb";
```

This matches what the linter expects. `SOME.PKG` will be computed as follows:

- If `--package` is specified, `SOME.PKG` will be the value passed to `--package`.
- Otherwise, if there is no `prototool.yaml` that would apply to the new file, use `uber.prototool.generated`.
- Otherwise, if there is a `prototool.yaml` file, check if it has a `dir_to_base_package` setting under the
  `create` section (see [etc/config/example/prototool.yaml](etc/config/example/prototool.yaml) for an example).
  If it does, this package, concatenated with the relative path from the directory with the `prototool.yaml`
  will be used.
- Otherwise, if there is no `dir_to_base_package` directive, just use the relative path from the directory
  with the `prototool.yaml` file. If the file is in the same directory as the `prototoo.yaml` file,
  use `uber.prototool.generated`

For example, assume you have the following file at `repo/prototool.yaml`:

```yaml
create:
  dir_to_base_package:
    idl: uber
    idl/baz: special
```

- `prototool create repo/idl/foo/bar/bar.proto` will have the package `uber.foo.bar`.
- `prototool create repo/idl/bar.proto` will have the package `uber`.
- `prototool create repo/idl/baz/baz.proto` will have the package `special`.
- `prototool create repo/idl/baz/bat/bat.proto` will have the package `special.bat`.
- `prototool create repo/another/dir/bar.proto` will have the package `another.dir`.
- `prototool create repo/bar.proto` will have the package `uber.prototool.generated`.

This is meant to mimic what you generally want - a base package for your idl directory, followed
by packages matching the directory structure.

Note you can override the directory that the `prototool.yaml` file is in as well. If we update our
file at `repo/prototool.yaml` to this:

```yaml
create:
  dir_to_base_package:
    .: foo.bar
```

Then `prototool create repo/bar.proto` will have the package `foo.bar`, and `prototool create repo/another/dir/bar.proto`
will have the package `foo.bar.another.dir`.

If [Vim integration](#vim-integration) is set up, files will be generated when you open a new Protobuf file.

##### `prototool files`

Print the list of all files that will be used given the input `dirOrProtoFiles...`. Useful for debugging.

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

`prototool grpc dirOrProtoFiles... --address serverAddress --method package.service/Method --data 'requestData'`

Either use `--data 'requestData'` as the the JSON data to input, or `--stdin` which will result in the input being read from stdin as JSON.

```
$ make init example # make sure everything is built just in case

$ cat input.json
{"value":"hello"}

$ cat input.json | prototool grpc example --address 0.0.0.0:8080 --method foo.ExcitedService/Exclamation --stdin
{
  "value": "hello!"
}

$ cat input.json | prototool grpc example --address 0.0.0.0:8080 --method foo.ExcitedService/ExclamationServerStream --stdin
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

$ cat input.json | prototool grpc example --address 0.0.0.0:8080 --method foo.ExcitedService/ExclamationClientStream --stdin
{
  "value": "hellosalutations!"
}

$ cat input.json | prototool grpc example --address 0.0.0.0:8080 --method foo.ExcitedService/ExclamationBidiStream --stdin
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
- Have all Protobuf files in the same directory use the same `package`, and use the same values for `go_package`, `java_multiple_files`, `java_outer_classname`, and `java_package`.
- Do not use long-form `go_package` values, ie use `foopb`, not `github.com/bar/baz/foo;foopb`. This helps `prototool gen` do the best job.

## Vim Integration

This repository is a self-contained plugin for use with the [ALE Lint Engine](https://github.com/w0rp/ale). It should be similarly easy to add support for Syntastic, Neomake, etc later.

The Vim integration will currently provide lint errors, optionally regenerate all the stubs, and optionally format your files on save. It
will also optionally create new files from a template when opened.

The plugin is under [vim/prototool](vim/prototool), so your plugin manager needs to point there instead of the base of this repository. Assuming you are using Vundle, copy/paste the following into your vimrc and you should be good to go.

```vim
" Prototool must be installed as a binary for the Vim integration to work.

" Add ale and prototool with your package manager.
" Note that Vundle does not allow setting of a branch, and downloads
" from dev by default. There may be minor changes to the Vim integration
" on dev between releases, but this won't be common. To make sure you are
" on the same branch as your Prototool install, go into your Vim bundle
" directory and checkout the branch of the release you are on.
Vundle 'w0rp/ale'
Vundle 'uber/prototool' { 'rtp':'vim/prototool' }

" I would recommend setting just this for Golang, as well as the necessary set for proto.
let g:ale_linters = {
\   'go': ['golint'],
\   'proto': ['prototool'],
\}
" If you don't set this, it will get annoying.
let g:ale_lint_on_text_changed = 'never'
" Set to 'lint' to not do code generation.
" Set to 'compile' to not do linting either and just compile without code generation.
"let g:ale_proto_prototool_command = 'compile'

" I have <leader> mapped to ",", uncomment this to set leader.
"let mapleader=","

" ,f will toggle formatting on and off.
" Change to PrototoolFormatNoRewriteToggle to toggle with --no-rewrite instead.
nnoremap <silent> <leader>f :call PrototoolFormatToggle()<CR>
" ,c will toggle create on and off.
nnoremap <silent> <leader>c :call PrototoolCreateToggle()<CR>

" Uncomment this to enable formatting by default.
"call PrototoolFormatEnable()
" Uncomment this to enable formatting with --no-rewrite by default.
"call PrototoolFormatNoRewriteEnable()
" Uncomment this to disable creating Protobuf files from a template by default.
"call PrototoolCreateDisable()
```

Editor integration is a key goal of Prototool. We've demonstrated support internally for Intellij, and hope that we have integration for more editors in the future.

## Development

Prototool is under active development, if you want to help, here's some places to start:

- Try out `prototool` and file issues, including points that are unclear in the documentation.
- Put up PRs with any changes you'd like to see made. We can't guarantee that many PRs will get merged for now, but we appreciate any input!
- Follow along on the [big ticket items](https://github.com/uber/prototool/issues?q=is%3Aissue+is%3Aopen+label%3A%22big+ticket+item%22) to see the major development points.

Over the coming months, we hope to push to a v1.0.

Note that development of Prototool will only work with Golang 1.10 or newer. On initially cloning the repository, run `make init` if you have not already to download dependencies to `vendor`.

Before submitting a PR, make sure to:

- Run `make generate` to make sure there is no diff.
- Run `make` to make sure all tests pass. This is functionally equivalent to the tests run on CI.

All Golang code is purposefully under the `internal` package to not expose any API for the time being.

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
