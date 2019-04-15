# Development

[Back to README.md](README.md)

Prototool is under active development. If you want to help, here's some places to start:

- Try out `prototool` and file feature requests or bug reports.
- Submit PRs with any changes you'd like to see made.

We appreciate any input you have!

Before filing an issue or submitting a PR, make sure to review the
[Issue Guidelines](../.github/ISSUE_TEMPLATE.md), and before submitting a PR, make sure to also
review the [PR Guidelines](../.github/PULL_REQUEST_TEMPLATE.md). The Issue Guidelines will show up
in the description field when filing a new issue, and the PR guidelines will show up in the
description field when submitting a PR, but clear the description field of this pre-populated text
once you've read it.

Note that development of Prototool will only work with Golang 1.12 or newer.

Before submitting a PR, make sure to:

- Run `make generate` to make sure there is no diff.
- Run `make` to make sure all tests pass. This is functionally equivalent to the tests run on CI.

The entire implementation is purposefully under the `internal` package to not expose any API for
the time being.

To use the locally-installed tools on your command line:

```
. $(make env)
```

## Maintainers

See [release.md](release.md) for the release process.

See https://github.com/uber/prototool/pull/417 for an example of how to update the default version
of `protoc`. Note that `etc/config/example/prototool.yaml` is automatically updated once you
update `DefaultProtocVersion` in `internal/vars/vars.go` and you run `make generate`.

See https://github.com/uber/prototool/pull/418 for an example of updating versions of dependencies.
