#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

if ! which jq > /dev/null; then
  echo "error: jq must be installed to run code coverage" >&2
  exit 1
fi

COVER=cover
ROOT_PKG=github.com/uber/prototool

if [[ -d "$COVER" ]]; then
	rm -rf "$COVER"
fi
mkdir -p "$COVER"

i=0
for pkg in "$@"; do
	i=$((i + 1))

	coverpkg=$(go list -json "$pkg" | jq -r '
		.Deps + .TestImports + .XTestImports
		| . + ["'"$pkg"'"]
		| unique
		| map
			( select(startswith("'"$ROOT_PKG"'"))
			| select(contains("/vendor/") | not)
			)
		| join(",")
	')

	args=""
	if [[ -n "$coverpkg" ]]; then
		args="-coverprofile $COVER/cover.${i}.out -covermode=count -coverpkg $coverpkg"
	fi

  echo go test ${args} "${pkg}"
  go test ${args} "${pkg}" 2>&1 | grep -v 'warning: no packages being tested depend on'
done

rm -f coverage.txt
gocovmerge "$COVER"/*.out > coverage.txt
