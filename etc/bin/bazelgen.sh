#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

init_workspace() {
cat << EOF > WORKSPACE
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.18.0/rules_go-0.18.0.tar.gz"],
    sha256 = "301c8b39b0808c49f98895faa6aa8c92cbd605ab5ad4b6a3a652da33a1a2ba2e",
)
http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.17.0/bazel-gazelle-0.17.0.tar.gz"],
    sha256 = "3c681998538231a2d24d0c07ed5a7658cb72bfb5fd4bf9911157c0e9ac6a2687",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()
EOF
}

rm -f WORKSPACE
rm -rf bazel

init_workspace

mkdir -p bazel
touch bazel/BUILD.bazel
cat << EOF > bazel/deps.bzl
load("@bazel_gazelle//:deps.bzl", "go_repository")

def prototool_deps(**kwargs):
EOF

go mod tidy -v
bazel run //:gazelle
rm -f internal/cmd/testdata/grpc/BUILD.bazel

go mod tidy -v
bazel run //:gazelle -- update-repos -from_file=go.mod

FIRST_GO_REPOSITORY_LINE_NUMBER="$(grep -n ^go_repository WORKSPACE | head -1 | cut -f 1 -d :)"
tail -n "+${FIRST_GO_REPOSITORY_LINE_NUMBER}" WORKSPACE | grep -v ^$ | sed 's/^\(.*\)$/    \1/' >> bazel/deps.bzl

rm -f WORKSPACE
init_workspace
cat << EOF >> WORKSPACE

load("//bazel:deps.bzl", "prototool_deps")
prototool_deps()
EOF
