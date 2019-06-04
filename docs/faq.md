# FAQ

[Back to README.md](README.md)

## Pre-Cache Protoc

*Question:* How do I download `protoc` ahead of time as part of a Docker build/CI pipeline?

*Answer*: `prototool cache update`.

You can pass both `--cache-path` and `--config-data` flags to this command to customize the
invocation.

```bash
# Basic invocation which will cache using the default behavior. See prototool help cache update for more details.
prototool cache update
# Cache to a specific directory path/to/cache
prototool cache update --cache-path path/to/cache
# Cache using custom configuration data instead of finding a prototool.yaml file using the file discovery mechanism
prototool cache update --config-data '{"protoc":{"version":"3.8.0"}}'
```

There is also a command `prototool cache delete` which will delete all cached assets of
`prototool`, however this command does not accept the `--cache-path` flag - if you specify a custom
directory, you should clean it up on your own, we don't want to effectively call `rm -rf DIR` via a
`prototool` command on a location we don't know about.

## Alpine Linux Issues

*Question:* Help! Prototool is failing when I use it within a Docker image based on Alpine Linux!

*Answer:* https://github.com/sgerrand/alpine-pkg-glibc

```
apk --no-cache add ca-certificates wget
wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub
wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.29-r0/glibc-2.29-r0.apk
apk add glibc-2.29-r0.apk
```

`protoc` is not statically compiled, and adding this package fixes the problem.

## Managing External Plugins/Docker

*Question:* Can Prototool manage my external plugins such as protoc-gen-go?

*Answer:* Unfortunately, no. This was an explicit design decision - Prototool is not meant to
"know the world", instead Prototool just takes care of what it is good at (managing your Protobuf
build) to keep Prototool simple, leaving you to do external plugin management. Prototool does
provide the ability to use the "built-in" output directives
`cpp, csharp, java, js, objc, php, python, ruby` provided by `protoc` out of the box, however.

If you want to have a consistent build environment for external plugins, we recommend creating a
Docker image. We provide a basic Docker image at
[hub.docker.com/r/uber/prototool](https://hub.docker.com/r/uber/prototool), defined in the
[Dockerfile](../Dockerfile) within this repository.

*See [docker.md](docker.md) for more details.*

## Lint/Format Choices

*Question:* I don't like some of the choices made in the Style Guide and that are enforced by
default by the linter and/or I don't like the choices made in the formatter. Can we change some
things?

*Answer:* Sorry, but we can't - The goal of Prototool is to provide a straightforward Style Guide
and consistent formatting that minimizes various issues that arise from Protobuf usage across large
organizations. There are pros and cons to many of the choices in the Style Guide, but it's our
belief that the best answer is a **single** answer, sometimes regardless of what that single answer
is.

We do have multiple lint groups available, see the help section on `prototool lint` above.

It is possible to ignore lint rules via configuration. However, especially if starting from a clean
slate, we highly recommend using all default lint rules for consistency.

Many of the lint rules exist to mitigate backwards compatibility problems as schemas evolves. For
example: requiring a unique request-response pair per RPC - while this potentially resuls in
duplicated messages, this makes it impossible to affect an adjacent RPC by adding or modifying an
existing field.
