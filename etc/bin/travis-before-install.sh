#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

BAZEL_VERSION=0.21.0
BAZEL_OS=linux
BAZEL_ARCH=x86_64
BAZEL_URL="https://github.com/bazelbuild/bazel/releases/download/${BAZEL_VERSION}/bazel-${BAZEL_VERSION}-installer-${BAZEL_OS}-${BAZEL_ARCH}.sh"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

BAZEL_INSTALL_SH="${TMPDIR}/bazel-install.sh"

wget -O "${BAZEL_INSTALL_SH}" "${BAZEL_URL}"
chmod +x "${BAZEL_INSTALL_SH}"
"${BAZEL_INSTALL_SH}" --user
