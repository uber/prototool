# Installation

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

```bash
go get github.com/uber/prototool/cmd/prototool@dev
```

You may want to use [gobin](https://github.com/myitcv/gobin) to install `prototool` outside of a module.

```bash
# Install to $GOBIN, or $GOPATH/bin if $GOBIN is not set, or $HOME/go/bin if neither are set
gobin github.com/uber/prototool/cmd/prototool@dev
# Install to /path/to/bin
GOBIN=/path/to/bin gobin github.com/uber/prototool/cmd/prototool@dev
```
