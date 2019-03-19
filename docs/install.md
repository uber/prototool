# Installation

[Back to README.md](README.md)

## Brew

Prototool can be installed on Mac OS X via [Homebrew](https://brew.sh/) or Linux via
[Linuxbrew](http://linuxbrew.sh/).

```bash
brew install prototool
```

This installs the `prototool` binary, along with bash completion, zsh completion, and man pages.

## GitHub Releases

You can also install all of the assets on Linux or without Homebrew from GitHub Releases.

```bash
curl -sSL \
  https://github.com/uber/prototool/releases/download/v1.4.0/prototool-$(uname -s)-$(uname -m).tar.gz | \
  tar -C /usr/local --strip-components 1 -xz
```

If you do not want to install bash completion, zsh completion, or man pages, you can install just
the `prototool` binary from GitHub Releases as well.

```bash
curl -sSL \
  https://github.com/uber/prototool/releases/download/v1.4.0/prototool-$(uname -s)-$(uname -m) \
  -o /usr/local/bin/prototool && \
  chmod +x /usr/local/bin/prototool
```

## Golang Modules

You can also install the `prototool` binary using `go get` if using Go 1.11+ with Modules enabled.
You can specify a branch such as `dev`, or a specific commit.

```bash
GO111MODULE=on go get github.com/uber/prototool/cmd/prototool@dev
```

To install to a specific location, use the `GOBIN` environment variable.

```bash
GO111MODULE=on GOBIN=/path/to/bin go get github.com/uber/prototool/cmd/prototool@dev
```

If using Go 1.12+, you can install without affecting any `go.mod` file by using a temporary
directory. The below is a shell script that would accomplish this.

```bash
#!/bin/bash

set -euo pipefail

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT
cd "${TMP}"

GO111MODULE=on GOBIN=/path/to/bin go get github.com/uber/prototool/cmd/prototool@dev
```

The below is a `Makefile` snippet that would accomplish installing Prototool for use with any
`make` targets.

```make
SHELL := /bin/bash -o pipefail

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_BIN = $(TMP)/bin
TMP_VERSIONS := $(TMP)/versions

export GO111MODULE := on
export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)

# This is the only variable that ever should change.
# This can be a branch, tag, or commit.
# When changed, the given version of Prototool will be installed to
# .tmp/$(uname -s)/(uname -m)/bin/prototool
PROTOTOOL_VERSION := v1.4.0

PROTOTOOL := $(TMP_VERSIONS)/prototool/$(PROTOTOOL_VERSION)
$(PROTOTOOL):
	$(eval PROTOTOOL_TMP := $(shell mktemp -d))
	cd $(PROTOTOOL_TMP); go get github.com/uber/prototool/cmd/prototool@$(PROTOTOOL_VERSION)
	@rm -rf $(PROTOTOOL_TMP)
	@rm -rf $(dir $(PROTOTOOL))
	@mkdir -p $(dir $(PROTOTOOL))
	@touch $(PROTOTOOL)

# proto is a target that uses prototool.
# By depending on $(PROTOTOOL), prototool will be installed on the Makefile's path.
# Since the path above has the temporary GOBIN at the front, this will use the
# locally installed prototool.
.PHONY: proto
proto: $(PROTOTOOL)
  prototool generate
```

## Docker

Prototool can also be used within Docker. See [docker.md](docker.md) for more details.
