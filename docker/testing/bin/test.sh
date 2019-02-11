#!/usr/bin/env bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/.." && pwd)"
cd "${DIR}"

check_command_success() {
  echo "Checking that '${*}' is successful"
  if ! "${@}"; then
    echo "Expected '${*}' to be successful but had error" >&2
    exit 1
  fi
}

check_command_output() {
  echo "Checking that '${*:2}' results in '${1}'"
  command_output="$("${@:2}")"
  if [ "${command_output}" != "${1}" ]; then
    echo "Expected: '${1}' Got: '${command_output}'" >&2
    exit 1
  fi
}

check_which() {
  check_command_output "${1}" command -v "$(basename "${1}")"
}

check_dir_not_exists() {
  echo "Checking that '${1}' does not exist"
  if [ -d "${1}" ]; then
    echo "Expected '${1}' to not exist" >&2
    exit 1
  fi
}

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
check_command_output "libprotoc 3.6.1" protoc --version
check_command_success protoc -I /usr/local/include -I proto -o /dev/null $(find proto -name '*.proto')
check_command_success rm -rf gen
check_command_success prototool compile proto
check_command_success prototool lint proto
check_command_success prototool format -l proto
check_command_success prototool generate proto
check_command_success rm -rf gen
check_dir_not_exists /root/.cache
