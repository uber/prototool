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

sha256() {
  if ! type sha256sum >/dev/null 2>/dev/null; then
    if ! type shasum >/dev/null 2>/dev/null; then
      echo "sha256sum and shasum are not installed" >&2
      return 1
    else
      shasum -a 256 "$@"
    fi
  else
    sha256sum "$@"
  fi
}

BASE_DIR="release"
rm -rf "${BASE_DIR}"
mkdir -p "${BASE_DIR}"
cd "${BASE_DIR}"

for os in Darwin Linux; do
  for arch in x86_64; do
    dir="${os}/${arch}/prototool"
    tar_context_dir="$(dirname "${dir}")"
    tar_dir="prototool"
    binary="prototool-${os}-${arch}"
    tarball="prototool-${os}-${arch}.tar.gz"
    mkdir -p "${dir}/bin"
    mkdir -p "${dir}/etc/bash_completion.d"
    mkdir -p "${dir}/etc/zsh/site-functions"
    mkdir -p "${dir}/share/man/man1"
    go run "${DIR}/internal/cmd/gen-prototool-bash-completion/main.go" > "${dir}/etc/bash_completion.d/prototool"
    go run "${DIR}/internal/cmd/gen-prototool-zsh-completion/main.go" > "${dir}/etc/zsh/site-functions/_prototool"
    go run "${DIR}/internal/cmd/gen-prototool-manpages/main.go" "${dir}/share/man/man1"
    CGO_ENABLED=0 GOOS=$(goos "${os}") GOARCH=$(goarch "${arch}") \
      go build \
      -a \
      -installsuffix cgo \
      -o "${dir}/bin/prototool" \
      "${DIR}/cmd/prototool/main.go"
    tar -C "${tar_context_dir}" -cvzf "${tarball}" "${tar_dir}"
    cp "${dir}/bin/prototool" "${binary}"
    sha256 "${binary}" > "${binary}.sha256sum"
    sha256 -c "${binary}.sha256sum"
    sha256 "${tarball}" > "${tarball}.sha256sum"
    sha256 -c "${tarball}.sha256sum"
  done
  rm -rf "${os}"
done
