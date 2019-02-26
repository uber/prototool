#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

rm -f WORKSPACE
rm -f bazel/deps.bzl

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

mkdir -p bazel

touch bazel/BUILD.bazel

cat << EOF > bazel/deps.bzl
load("@bazel_gazelle//:deps.bzl", "go_repository")

def prototool_deps(**kwargs):
EOF

cat << EOF > WORKSPACE
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "io_bazel_rules_go",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.17.1/rules_go-0.17.1.tar.gz"],
    sha256 = "6776d68ebb897625dead17ae510eac3d5f6342367327875210df44dbe2aeeb19",
)
http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.16.0/bazel-gazelle-0.16.0.tar.gz"],
    sha256 = "7949fc6cc17b5b191103e97481cf8889217263acf52e00b560683413af204fcb",
)
load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()
EOF

bazel run //:gazelle
bazel run //:gazelle -- update-repos -from_file=go.mod
rm internal/cmd/testdata/grpc/BUILD.bazel

TMP_WORKSPACE="${TMPDIR}/WORKSPACE"
FIRST_GO_REPOSITORY_LINE_NUMBER="$(grep -n ^go_repository WORKSPACE | head -1 | cut -f 1 -d :)"
tail -n "+${FIRST_GO_REPOSITORY_LINE_NUMBER}" WORKSPACE | grep -v ^$ | sed 's/^\(.*\)$/    \1/' >> bazel/deps.bzl
head -n "$((${FIRST_GO_REPOSITORY_LINE_NUMBER} - 1))" WORKSPACE > "${TMP_WORKSPACE}"
cat << EOF >> "${TMP_WORKSPACE}"
load("//bazel:deps.bzl", "prototool_deps")

prototool_deps()
EOF
rm -f WORKSPACE
mv "${TMP_WORKSPACE}" WORKSPACE
