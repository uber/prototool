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
  * [Development](#development)
  * [FAQ](#faq)
    * [Pre-Cache Protoc](#pre-cache-protoc)
    * [Alpine Linux Issues](#alpine-linux-issues)
    * [Managing External Plugins/Docker](#managing-external-pluginsdocker)
    * [Lint/Format Choices](#lintformat-choices)
  * [Special Thanks](#special-thanks)

## Installation

Prototool can be installed on Mac OS X via [Homebrew](https://brew.sh/) or Linux via [Linuxbrew](http://linuxbrew.sh/).

```bash
brew install prototool
```

This installs the `prototool` binary, along with bash completion, zsh completion, and man pages.
You can also install all of the assets on Linux or without Homebrew from GitHub Releases.

```bash
curl -sSL https://github.com/uber/prototool/releases/download/v1.3.0/prototool-$(uname -s)-$(uname -m).tar.gz | \
  tar -C /usr/local --strip-components 1 -xz
```

If you do not want to install bash completion, zsh completion, or man mages, you can install just the
`prototool` binary from GitHub Releases as well.

```bash
curl -sSL https://github.com/uber/prototool/releases/download/v1.3.0/prototool-$(uname -s)-$(uname -m) \
  -o /usr/local/bin/prototool && \
  chmod +x /usr/local/bin/prototool
```

You can also install the `prototool` binary using `go get` if using go1.11+ with module support enabled.

```
go get github.com/uber/prototool/cmd/prototool@dev
```

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

- You can specify exactly one directory. If this is done, Prototool goes up until it finds a `prototool.yaml` or `prototool.json` file (or uses the current directory if none is found), and then walks starting at this location for all `.proto` files, and these are used, except for files in the `excludes` lists in `prototool.yaml` or `prototool.json` files.
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

Compile your Protobuf files and generate stubs according to the rules in your `prototool.yaml` or `prototool.json` file. See [example/idl/uber/prototool.yaml](example/idl/uber/prototool.yaml) for an example.

##### `prototool lint`

Lint your Protobuf files. The default rule set follows the Style Guide at [etc/style/uber/uber.proto](etc/style/uber/uber.proto). You can add or exclude lint rules in your `prototool.yaml` or `prototool.json` file. The default rule set is "strict", and we are working on having two main sets of rules.

##### `prototool format`

Format a Protobuf file and print the formatted file to stdout. There are flags to perform different actions:

- `-d` Write a diff instead.
- `-f` Fix the file according to the Style Guide.
- `-l` Write a lint error in the form file:line:column:message if a file is unformatted.
- `-w` Overwrite the existing file instead.

Concretely, the `-f` flag can be used so that the values for `java_multiple_files`, `java_outer_classname`, and `java_package` are updated to reflect what is expected by the
[Google Cloud APIs file structure](https://cloud.google.com/apis/design/file_structure), and the value of `go_package` is updated to reflect what we expect for the
default Style Guide. By formatting, the linting for these values will pass by default. See the documentation below for `prototool create` for an example.

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
- Otherwise, if there is no `prototool.yaml` or `prototool.json` that would apply to the new file, use `uber.prototool.generated`.
- Otherwise, if there is a `prototool.yaml` or `prototool.json` file, check if it has a `packages` setting under the
  `create` section (see [etc/config/example/prototool.yaml](etc/config/example/prototool.yaml) for an example).
  If it does, this package, concatenated with the relative path from the directory with the `prototool.yaml` or `prototool.json` file will be used.
- Otherwise, if there is no `packages` directive, just use the relative path from the directory
  with the `prototool.yaml` or `prototool.json` file. If the file is in the same directory as the `prototoo.yaml` file,
  use `uber.prototool.generated`

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

Note you can override the directory that the `prototool.yaml` or `prototool.json` file is in as well. If we update our
file at `repo/prototool.yaml` to this:

```yaml
create:
  packages:
    - directory: .
      name: foo.bar
```

Then `prototool create repo/bar.proto` will have the package `foo.bar`, and `prototool create repo/another/dir/bar.proto`
will have the package `foo.bar.another.dir`.

If [Vim integration](#vim-integration) is set up, files will be generated when you open a new Protobuf file.

##### `prototool files`

Print the list of all files that will be used given the input `dirOrFile`. Useful for debugging.

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

`prototool grpc [dirOrFile] --address serverAddress --method package.service/Method --data 'requestData'`

Either use `--data 'requestData'` as the the JSON data to input, or `--stdin` which will result in the input being read from stdin as JSON.

```bash
$ make init example # make sure everything is built just in case

$ prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/Exclamation \
  --data '{"value":"hello"}'
{
  "value": "hello!"
}

$ prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/ExclamationServerStream \
  --data '{"value":"hello"}'
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

$ cat input.json | prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/ExclamationClientStream \
  --stdin
{
  "value": "hellosalutations!"
}

$ cat input.json | prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/ExclamationBidiStream \
  --stdin
{
  "value": "hello!"
}
{
  "value": "salutations!"
}
```

## Tips and Tricks

Prototool is meant to help enforce a consistent development style for Protobuf, and as such you should follow some basic rules:

- Have all your imports start from the directory your `prototool.yaml` or `prototool.json` file is in. While there is a configuration option `protoc.includes` to denote extra include directories, this is not recommended.
- Have all Protobuf files in the same directory use the same `package`, and use the same values for `go_package`, `java_multiple_files`, `java_outer_classname`, and `java_package`.
- Do not use long-form `go_package` values, ie use `foopb`, not `github.com/bar/baz/foo;foopb`. This helps `prototool generate` do the best job.

## Vim Integration

This repository is a self-contained plugin for use with the [ALE Lint Engine](https://github.com/w0rp/ale). It should be similarly easy to add support for Syntastic, Neomake, etc.

The Vim integration will currently provide lint errors, optionally regenerate all the stubs, and optionally format your files on save. It
will also optionally create new files from a template when opened.

The plugin is under [vim/prototool](vim/prototool), so your plugin manager needs to point there instead of the base of this repository. Assuming you are using [vim-plug](https://github.com/junegunn/vim-plug), copy/paste the following into your vimrc and you should be good to go. If you are using [Vundle](https://github.com/VundleVim/Vundle.vim), just replace `Plug` with `Vundle` below.

```vim
" Prototool must be installed as a binary for the Vim integration to work.

" Add ale and prototool with your package manager.
" Note that Plug downloads from dev by default. There may be minor changes
" to the Vim integration on dev between releases, but this won't be common.
" To make sure you are on the same branch as your Prototool install, set
" the branch field in the options for uber/prototool per the vim-plug
" documentation. Vundle does not allow setting branches, so on Vundle,
" go into plug directory and checkout the branch of the release you are on.
Plug 'w0rp/ale'
Plug 'uber/prototool', { 'rtp':'vim/prototool' }

" We recommend setting just this for Golang, as well as the necessary set for proto.
let g:ale_linters = {
\   'go': ['golint'],
\   'proto': ['prototool'],
\}
" We recommend you set this.
let g:ale_lint_on_text_changed = 'never'
" Set to 'lint' to not do code generation.
" Set to 'compile' to not do linting either and just compile without code generation.
"let g:ale_proto_prototool_command = 'compile'

" We generally have <leader> mapped to ",", uncomment this to set leader.
"let mapleader=","

" ,f will toggle formatting on and off.
" Change to PrototoolFormatFixToggle to toggle with --fix instead.
nnoremap <silent> <leader>f :call PrototoolFormatToggle()<CR>
" ,c will toggle create on and off.
nnoremap <silent> <leader>c :call PrototoolCreateToggle()<CR>

" Uncomment this to enable formatting by default.
"call PrototoolFormatEnable()
" Uncomment this to enable formatting with --fix by default.
"call PrototoolFormatFixEnable()
" Uncomment this to disable creating Protobuf files from a template by default.
"call PrototoolCreateDisable()
```

Editor integration is a key goal of Prototool. We've demonstrated support internally for Intellij, and hope that we have integration for more editors in the future.

## Development

Prototool is under active development. If you want to help, here's some places to start:

- Try out `prototool` and file feature requests or bug reports.
- Submit PRs with any changes you'd like to see made.

We appreciate any input you have!

Before filing an issue or submitting a PR, make sure to review the [Issue Guidelines](https://github.com/uber/prototool/blob/dev/.github/ISSUE_TEMPLATE.md), and before submitting a PR, make sure to also review
the [PR Guidelines](https://github.com/uber/prototool/blob/dev/.github/PULL_REQUEST_TEMPLATE.md). The Issue Guidelines will show up in the description field when filing a new issue, and the PR guidelines will show up in the
description field when submitting a PR, but clear the description field of this pre-populated text once you've read it :-)

Note that development of Prototool will only work with Golang 1.10 or newer. On initially cloning the repository, run `make init` if you have not already to download dependencies to `vendor`.

Before submitting a PR, make sure to:

- Run `make generate` to make sure there is no diff.
- Run `make` to make sure all tests pass. This is functionally equivalent to the tests run on CI.

The entire implementation is purposefully under the `internal` package to not expose any API for the time being.

## FAQ

##### Pre-Cache Protoc

*Question:* How do I download `protoc` ahead of time as part of a Docker build/CI pipeline?

*Answer*: We used to have a command that did this, but removed it for simplicity and because the command as implemented did not properly
read the configuration file to figure out what version of `protoc` to download. We may re-add this command in the future, however
here is a technique to accomplish this, including as a `RUN` directive for Docker:

```bash
# the first rm -rf call is not needed if this is a RUN directive for Docker
rm -rf /tmp/prototool-bootstrap && \
  echo $'protoc\n  version: 3.6.1' > /tmp/prototool-bootstrap/prototool.yaml && \
  echo 'syntax = "proto3";' > /tmp/prototool-bootstrap/tmp.proto && \
  prototool compile /tmp/prototool-bootstrap && \
  rm -rf /tmp/prototool-bootstrap
```

The better way to do this as part of a bash script would be to use `mktemp -d` i.e.:

```bash
#!/bin/bash

set -euo pipefail

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

echo $'protoc\n  version: 3.6.1' > "${TMPDIR}/prototool.yaml"
echo 'syntax = "proto3";' > "${TMPDIR}/tmp.proto"
prototool compile "${TMPDIR}"
```

But for Darwin or Linux, the above should work.

##### Alpine Linux Issues

*Question:* Help! Prototool is failing when I use it within a Docker image based on Alpine Linux!

*Answer:* `apk add libc6-compat`

`protoc` is not statically compiled, and adding this packages fixes the problem.

##### Managing External Plugins/Docker

*Question:* Can Prototool manage my external plugins such as protoc-gen-go?

*Answer:* Unfortunately, no. This was an explicit design decision - Prototool is not meant to "know the world", instead
Prototool just takes care of what it is good at (managing your Protobuf build) to keep Prototool simple, leaving you to do
external plugin management. Prototool does provide the ability to use the "built-in" output directives `cpp, csharp, java, js, objc, php, python, ruby`
provided by `protoc` out of the box, however.

If you want to have a consistent build environment for external plugins, we recommend creating a Docker image. Here's an example `Dockerfile` that
results in a Docker image around 33MB that contains `prototool`, a cached `protoc`, and `protoc-gen-go`:

```dockerfile
FROM golang:1.11.4-alpine3.8 AS build

ARG PROTOTOOL_VERSION=1.3.0
ARG PROTOC_VERSION=3.6.1
ARG PROTOC_GEN_GO_VERSION=1.2.0

RUN \
  apk update && \
  apk add curl git libc6-compat && \
  rm -rf /var/cache/apk/*
RUN \
  curl -sSL https://github.com/uber/prototool/releases/download/v$PROTOTOOL_VERSION/prototool-Linux-x86_64 -o /bin/prototool && \
  chmod +x /bin/prototool
RUN \
  mkdir /tmp/prototool-bootstrap && \
  echo $'protoc:\n  version:' $PROTOC_VERSION > /tmp/prototool-bootstrap/prototool.yaml && \
  echo 'syntax = "proto3";' > /tmp/prototool-bootstrap/tmp.proto && \
  prototool compile /tmp/prototool-bootstrap && \
  rm -rf /tmp/prototool-bootstrap
RUN go get github.com/golang/protobuf/... && \
  cd /go/src/github.com/golang/protobuf && \
  git checkout v$PROTOC_GEN_GO_VERSION && \
  go install ./protoc-gen-go

FROM alpine:3.8

WORKDIR /in

RUN \
  apk update && \
  apk add libc6-compat && \
  rm -rf /var/cache/apk/*

COPY --from=build /bin/prototool /bin/prototool
COPY --from=build /root/.cache/prototool /root/.cache/prototool
COPY --from=build /go/bin/protoc-gen-go /bin/protoc-gen-go

ENTRYPOINT ["/bin/prototool"]
```

Assuming this is in a file named `Dockerfile` in your current directory, build the image with:

```bash
docker build -t me/prototool-env .
```

Then, assuming you are in the directory you want to pass to Prototool and you want to run `prototool compile`, run:

```bash
docker run -v $(pwd):/in me/prototool-env compile
```

##### Lint/Format Choices

*Question:* I don't like some of the choices made in the Style Guide and that are enforced by default by the linter and/or I don't like
the choices made in the formatter. Can we change some things?

*Answer:* Sorry, but we can't - The goal of Prototool is to provide a straightforward Style Guide and consistent formatting that minimizes various issues that arise from Protobuf usage across large organizations. There are pros and cons to many of the choices in the Style Guide, but it's our belief that the best answer is a **single** answer, sometimes regardless of what that single answer is.

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
