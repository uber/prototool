#!/usr/bin/env bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/.." && pwd)"
cd "${DIR}"

check_env() {
  echo "Checking that ${1}=${2}"
  if [ "${!1}" != "${2}" ]; then
    echo "Expected '${1}' to be '${2}' but was '${!1}'" >&2
    exit 1
  fi
}

check_which() {
  check_command_output "${1}" command -v "$(basename "${1}")"
}

check_command_output() {
  echo "Checking that '${*:2}' results in '${1}'"
  command_output="$("${@:2}")"
  if [ "${command_output}" != "${1}" ]; then
    echo "Expected: '${1}' Got: '${command_output}'" >&2
    exit 1
  fi
}

check_command_output_file() {
  tmp_file1="$(mktemp)"
  tmp_file2="$(mktemp)"
  trap 'rm -rf "${tmp_file1}"' EXIT
  trap 'rm -rf "${tmp_file2}"' EXIT
  echo "Checking that '${*:2}' results in the contents of '${1}'"
  cat "${1}" | sort > "${tmp_file1}"
  "${@:2}" | sort > "${tmp_file2}"
  if ! diff "${tmp_file1}" "${tmp_file2}"; then
    echo "Diff detected" >&2
    exit 1
  fi
}


check_command_success() {
  echo "Checking that '${*}' is successful"
  if ! "${@}"; then
    echo "Expected '${*}' to be successful but had error" >&2
    exit 1
  fi
}

check_dir_not_exists() {
  echo "Checking that '${1}' does not exist"
  if [ -d "${1}" ]; then
    echo "Expected '${1}' to not exist" >&2
    exit 1
  fi
}

check_env GOGO_PROTOBUF_VERSION 1.3.1
check_env GOLANG_PROTOBUF_VERSION 1.4.2
check_env GRPC_VERSION 1.28.1
check_env GRPC_GATEWAY_VERSION 1.14.6
check_env GRPC_WEB_VERSION 1.1.0
check_env PROTOBUF_VERSION 3.12.2
check_env TWIRP_VERSION 5.11.0
check_env YARPC_VERSION 1.46.0
check_env PROTOTOOL_PROTOC_BIN_PATH /usr/bin/protoc
check_env PROTOTOOL_PROTOC_WKT_PATH /usr/include
check_command_output "libprotoc 3.12.2" protoc --version
check_command_output_file etc/wkt.txt find /usr/include -type f
check_which /usr/bin/protoc
check_which /usr/bin/grpc_cpp_plugin
check_which /usr/bin/grpc_csharp_plugin
check_which /usr/bin/grpc_node_plugin
check_which /usr/bin/grpc_objective_c_plugin
check_which /usr/bin/grpc_php_plugin
check_which /usr/bin/grpc_python_plugin
check_which /usr/bin/grpc_ruby_plugin
check_which /usr/local/bin/protoc-gen-go
check_which /usr/local/bin/protoc-gen-gofast
check_which /usr/local/bin/protoc-gen-gogo
check_which /usr/local/bin/protoc-gen-gogofast
check_which /usr/local/bin/protoc-gen-gogofaster
check_which /usr/local/bin/protoc-gen-gogoslick
check_which /usr/local/bin/protoc-gen-grpc-gateway
check_which /usr/local/bin/protoc-gen-grpc-web
check_which /usr/local/bin/protoc-gen-swagger
check_which /usr/local/bin/protoc-gen-twirp
check_which /usr/local/bin/protoc-gen-twirp_python
check_which /usr/local/bin/protoc-gen-yarpc-go
check_which /usr/local/bin/prototool
check_command_success protoc -o /dev/null $(find proto -name '*.proto')
check_command_success rm -rf gen
check_command_success prototool compile proto
check_command_success prototool lint proto
check_command_success prototool format -l proto
check_command_success prototool generate proto
check_command_success rm -rf gen
check_dir_not_exists /root/.cache
