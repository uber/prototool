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

UNAME_OS="$(uname -s)"
UNAME_ARCH="$(uname -m)"

TMP_BASE=".tmp"
TMP="${TMP_BASE}/${UNAME_OS}/${UNAME_ARCH}"
TMP_LIB="${TMP}/lib"
TMP_BIN="${TMP}/bin"

DEP_VERSION="0.5.0"
DEP="${TMP_BIN}/dep-${DEP_VERSION}"

DEP_LIB="${TMP_LIB}/dep-${DEP_VERSION}"
if [ "${UNAME_OS}" = "Darwin" ]; then
  DEP_OS="darwin"
else
  DEP_OS="linux"
fi
if [ "${UNAME_ARCH}" = "x86_64" ]; then
  DEP_ARCH="amd64"
fi

rm -rf "${DEP_LIB}"
mkdir -p "${TMP_BIN}" "${DEP_LIB}"
curl -sSL "https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-${DEP_OS}-${DEP_ARCH}" -o "${DEP}"
chmod +x "${DEP}"

rm -rf vendor
"${DEP}" ensure -v

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
