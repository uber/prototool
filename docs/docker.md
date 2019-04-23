# Prototool Docker Image

[Back to README.md](README.md)

We provide a Docker image with `prototool`, `protoc`, and common Protobuf plugins pre-installed.
Most plugins are compressed with [UPX](https://github.com/upx/upx). As of this writing, the resulting
image is around 79MB. This provides a consistent environment to generate your Protobuf stubs.

This is in early development.

## Docker Hub

This image is hosted at [hub.docker.com/r/uber/prototool](https://hub.docker.com/r/uber/prototool).

## Usage

Bind your input directory as a volume to `/work`, and call your command, for example `prototool generate`:

```
docker run -v "$(pwd):/work" uber/prototool:latest prototool generate
```

You can build on top of this image as well if you have custom requirements.

## Included

The following libraries are included. This is not meant to be exhaustive - these represent our view of the most
commonly-used, stable, maintained libraries.

| Name | Version | Binaries |
| --- | --- | --- |
| prototool | (varies) | prototool |
| [protoc] | 3.6.1 | protoc |
| [grpc] | 1.19.1 | grpc_cpp_plugin<br>grpc_csharp_plugin<br>grpc_node_plugin<br>grpc_objective_c_plugin<br>grpc_php_plugin<br>grpc_python_plugin<br>grpc_ruby_plugin |
| [golang/protobuf] | 1.3.1 | protoc-gen-go |
| [gogo/protobuf] | 1.2.1 | protoc-gen-gofast<br>protoc-gen-gogo<br>protoc-gen-gogofast<br>protoc-gen-gogofaster<br>protoc-gen-gogoslick |
| [grpc-gateway] | 1.8.5 | protoc-gen-grpc-gateway<br>protoc-gen-swagger |
| [grpc-web] | 1.0.4 | protoc-gen-grpc-web |
| [twirp] | 5.7.0 | protoc-gen-twirp<br>protoc-gen-twirp_python |
| [yarpc] | 1.37.2 | protoc-gen-yarpc-go |

The Well-Known Types are copied to `/usr/include`. The packages `bash`, `curl`, and `git` are also installed.

If you think another library should be included, propose it in a GitHub issue and we will evaluate it.

## Versioning

Images are pushed for every commit to the dev branch as the tags `uber/prototool:dev, uber:prototool:latest`, and
every minor release starting with `v1.4.0` will have a tag e.g. `uber/prototool:1.4.0`. Note that as opposed
to the rest of Prototool, there is no breaking change guarantee between minor releases - we do not account
for breaking changes in libraries we provide within this image, and will update them regularly on `dev`.
We recommend pinning to one of the minor release Docker image tags once they are available.

## Development

To update the Docker image, edit the [Dockerfile](../Dockerfile).

Note that for version changes, the versions are copied in four places: once for each layer in the
Dockerfile (sharing these is harder than you think), once in [etc/docker/testing/bin/test.sh](../etc/docker/testing/bin/test.sh),
and once in this documentation.

Updates of `protobuf` and `grpc` must match the current versions for `alpine:edge` for now. See [here](https://pkgs.alpinelinux.org/packages?name=protobuf&branch=edge&repo=main&arch=x86_64) and [here](https://pkgs.alpinelinux.org/packages?name=grpc&branch=edge&repo=testing&arch=x86_64) for the current versions.

Local development commands:

```
# build the docker image
make dockerbuild
# test a built docker image
make dockertest
# build and then test
make dockerall
```

The test files are in [etc/docker/testing](../etc/docker/testing).

[protoc]: https://github.com/protocolbuffers/protobuf
[grpc]: https://github.com/grpc/grpc
[golang/protobuf]: https://github.com/golang/protobuf
[gogo/protobuf]: https://github.com/gogo/protobuf
[grpc-gateway]: https://github.com/grpc-ecosystem/grpc-gateway
[grpc-web]: https://github.com/grpc/grpc-web
[twirp]: https://github.com/twitchtv/twirp
[yarpc]: https://github.com/yarpc/yarpc-go
