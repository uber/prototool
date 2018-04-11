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
rm -rf "${BASE_DIR}"

for os in Darwin Linux; do
  for arch in x86_64; do
    dir="${BASE_DIR}/${os}/${arch}/prototool"
    tar_context_dir="$(dirname "${dir}")"
    tar_dir="prototool"
    mkdir -p "${dir}/bin"
    cp -R "etc/release/etc" "${dir}/etc"
    cp -R "etc/release/share" "${dir}/share"
    CGO_ENABLED=0 GOOS=$(goos "${os}") GOARCH=$(goarch "${arch}") \
      go build \
      -a \
      -installsuffix cgo \
      -ldflags "-X 'github.com/uber/prototool/internal/x/vars.GitCommit=$(git rev-list -1 HEAD)' -X 'github.com/uber/prototool/internal/x/vars.BuiltTimestamp=$(date -u)'" \
      -o "${dir}/bin/prototool" \
      cmd/prototool/main.go
    tar -C "${tar_context_dir}" -cvzf "${BASE_DIR}/prototool-${os}-${arch}.tar.gz" "${tar_dir}"
    cp "${dir}/bin/prototool" "${BASE_DIR}/prototool-${os}-${arch}"
  done
  rm -rf "${BASE_DIR:?/tmp}/${os}"
done
