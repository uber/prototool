# Maintenance

[Back to README.md](README.md)

## Releasing a new version

See [release.md](release.md) for the release process.

## Updating the default version of protoc

Check https://github.com/protocolbuffers/protobuf/releases regularly for new releases. A release
that can be used for Prototool must include `protoc-VERSION-linux-x86_64.zip` and
`protoc-VERSION-osx-x86_64.zip`, which small releases sometimes do not.

The following files need to be updated:

- All files in `docs` that contain the current `protoc` version. As of this writing, this is
  `docs/README.md`, `docs/faq.md`, and `docs/protoc.md`.
- `internal/vars/vars.go`
- `example/proto/prototool.yaml`

Once these files are updated, run `make generate`. Do not manually update
`etc/config/example/prototool.yaml` as this will be automatically updated with `make generate`.

See https://github.com/uber/prototool/pull/417 for an example.

## Updating dependencies

There are some issues running `go get -u ./...` multiple times with Golang Modules as of writing
this documentation, so be careful when updating dependencies. Right now, the easiest way to make
sure dependencies are up to date is to run the corresponding Makefile target.

```
make updatedeps
```

This will do a complete update of the `go.mod`, `go.sum`, and `bazel/deps.bzl` files. This should
be revisited in the future, however.

See https://github.com/uber/prototool/pull/443 for an example.

## Updating Docker image dependencies

To update the Docker image, edit the [Dockerfile](../Dockerfile).

Note that for version changes, the versions are copied in four places: once for each layer in the
Dockerfile (sharing these is harder than you think), once in
[etc/docker/testing/bin/test.sh](../etc/docker/testing/bin/test.sh), and once in
[docker.md](docker.md).

Updates of `protobuf` and `grpc` must match the current versions for `alpine:edge` for now. See
[here](https://pkgs.alpinelinux.org/packages?name=protobuf&branch=edge&repo=main&arch=x86_64) and
[here](https://pkgs.alpinelinux.org/packages?name=grpc&branch=edge&repo=testing&arch=x86_64) for
the current versions.

See https://github.com/uber/prototool/pull/437 for an example.
