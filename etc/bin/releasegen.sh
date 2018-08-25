#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

goos() {
  case "${1}" in
    Darwin) echo darwin ;;
    Linux) echo linux ;;
    *) return 1 ;;
  esac
}

goarch() {
  case "${1}" in
    x86_64) echo amd64 ;;
    *) return 1 ;;
  esac
}

BASE_DIR="release"

go get github.com/Masterminds/glide
rm -rf vendor
glide install
rm -rf "${BASE_DIR}"
for os in Darwin Linux; do
  for arch in x86_64; do
    dir="${BASE_DIR}/${os}/${arch}/prototool"
    tar_context_dir="$(dirname "${dir}")"
    tar_dir="prototool"
    mkdir -p "${dir}/bin"
    mkdir -p "${dir}/etc/bash_completion.d"
    mkdir -p "${dir}/etc/zsh/site-functions"
    mkdir -p "${dir}/share/man/man1"
    go run internal/cmd/gen-prototool-bash-completion/main.go > "${dir}/etc/bash_completion.d/prototool"
    go run internal/cmd/gen-prototool-zsh-completion/main.go > "${dir}/etc/zsh/site-functions/_prototool"
    go run internal/cmd/gen-prototool-manpages/main.go "${dir}/share/man/man1"
    CGO_ENABLED=0 GOOS=$(goos "${os}") GOARCH=$(goarch "${arch}") \
      go build \
      -a \
      -installsuffix cgo \
      -ldflags "-X 'github.com/uber/prototool/internal/vars.GitCommit=$(git rev-list -1 HEAD)' -X 'github.com/uber/prototool/internal/vars.BuiltTimestamp=$(date -u)'" \
      -o "${dir}/bin/prototool" \
      internal/cmd/prototool/main.go
    tar -C "${tar_context_dir}" -cvzf "${BASE_DIR}/prototool-${os}-${arch}.tar.gz" "${tar_dir}"
    cp "${dir}/bin/prototool" "${BASE_DIR}/prototool-${os}-${arch}"
  done
  rm -rf "${BASE_DIR:?/tmp}/${os}"
done
