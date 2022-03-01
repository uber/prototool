# protoc

[Back to README.md](README.md)

Prototool wraps [protoc](https://github.com/protocolbuffers/protobuf) for much of it's
functionality, and manages the downloading and caching of `protoc` and the Well-Known
Types definitions.

Prototool will automatically download and cache `protoc` if it has not already been downloaded,
however the command `prototool cache update` can manually manage this process See [faq.md](faq.md)
for more details.

`protoc` is downloaded to the following directories based on flags and environment variables:

- If `--cache-path` is set, then this directory will be used. The user is expected to manually
  manage this directory, and `prototool cache delete` will have no effect on it.
- Otherwise, if `$PROTOTOOL_CACHE_PATH` is set, then this directory will be used. The user is
  expected to manually manage this directory, and `prototool cache delete` will have no effect on
  it.
- Otherwise, if `$XDG_CACHE_HOME` is set, then `$XDG_CACHE_HOME/prototool` will be used.
- Otherwise, if on Linux, `$HOME/.cache/prototool` will be used, or on Darwin,
  `$HOME/Library/Caches/prototool` will be used.

By default, `protoc` version `3.11.0` is downloaded, however this is configurable in your
`prototool.yaml` file.

```yaml
protoc:
  version: 3.11.0
```

Downloads are safe to run concurrently across processes, for example if using from Bazel, as
Prototool implements file locking to make sure there is no contention on writing to the cache.

If one prefers to download and manage `protoc` and the Well-Known Types outside of Prototool,
this can be done in one of three ways.

- By setting the `--protoc-url` flag to provide an alternate URL to download the `protoc` ZIP file
  from instead of GitHub Releases. This can be prefixed with `file://`, `http://`, or `https://`,
  so one can either download the relevant `protoc` ZIP file from GitHub Releases and store it
  locally, or upload the relevant `protoc` ZIP file to i.e. s3 and download from therel.
- By setting the `--protoc-bin-path` and `--protoc-wkt-path` flags at runtime for relevant
  commands. The Well-Known Type path should be the directory that includes the `google/protobuf`
  directory containing the Well-Known Types.
- By setting the `PROTOTOOL_PROTOC_BIN_PATH` and `PROTOTOOL_PROTOC_WKT_PATH` environment variables,
  as we do in the [provided Docker image](docker.md). These variables are analogous to the
  `--protoc-bin-path` and `--protoc-wkt-path` flags, however the flags take precedence.

If any of these options are set, the `protoc.version` option in the `prototool.yaml` file is
ignored.
