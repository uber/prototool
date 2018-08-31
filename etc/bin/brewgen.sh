#!/bin/sh

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

BUILD_DIR="brew"

rm -rf vendor
glide install
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}/bin"
mkdir -p "${BUILD_DIR}/etc/bash_completion.d"
mkdir -p "${BUILD_DIR}/etc/zsh/site-functions"
mkdir -p "${BUILD_DIR}/share/man/man1"
go run cmd/gen-prototool-bash-completion/main.go > "${BUILD_DIR}/etc/bash_completion.d/prototool"
go run cmd/gen-prototool-zsh-completion/main.go > "${BUILD_DIR}/etc/zsh/site-functions/_prototool"
go run cmd/gen-prototool-manpages/main.go "${BUILD_DIR}/share/man/man1"
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
  go build \
  -a \
  -installsuffix cgo \
  -ldflags "-X 'github.com/uber/prototool/internal/vars.GitCommit=$(git rev-list -1 HEAD)' -X 'github.com/uber/prototool/internal/vars.BuiltTimestamp=$(date -u)'" \
  -o "${BUILD_DIR}/bin/prototool" \
  cmd/prototool/main.go
