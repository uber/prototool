#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

if echo "${GOPATH}" | grep : >/dev/null; then
	echo "error: GOPATH must be one directory, but has multiple directories separated by colons: ${GOPATH}" >&2
	exit 1
fi

COVER=cover
ROOT_PKG=github.com/uber/prototool

if [[ -d "$COVER" ]]; then
	rm -rf "$COVER"
fi
mkdir -p "$COVER"

ignorePkgs=""

filterIgnorePkgs() {
  if [[ -z "${ignorePkgs}" ]]; then
    cat
  else
    grep -v "${ignorePkgs}"
  fi
}

# If a package directory has a .nocover file, don't count it when calculating
# coverage.
filter=""
for pkg in "$@"; do
	if [[ -f "$GOPATH/src/$pkg/.nocover" ]]; then
		if [[ -n "$filter" ]]; then
			ignorePkgs="$ignorePkgs\|"
			filter="$filter, "
		fi
		ignorePkgs="$ignorePkgs$pkg/"
		filter="$filter\"$pkg\": true"
	fi
done

i=0
commands_file="$(mktemp)"
echo 'commands:' >> "${commands_file}"
trap 'rm -rf "${commands_file}"' EXIT
for pkg in "$@"; do
	if ! ls "${GOPATH}/src/${pkg}" | grep _test\.go$ >/dev/null; then
		continue
	fi
	i=$((i + 1))

	extracoverpkg=""
	if [[ -f "$GOPATH/src/$pkg/.extra-coverpkg" ]]; then
		extracoverpkg=$( \
			sed -e "s|^|$pkg/|g" < "$GOPATH/src/$pkg/.extra-coverpkg" \
			| tr '\n' ',')
	fi

	coverpkg=$(go list -json "$pkg" | jq -r '
		.Deps + .TestImports + .XTestImports
		| . + ["'"$pkg"'"]
		| unique
		| map
			( select(startswith("'"$ROOT_PKG"'"))
			| select(contains("/vendor/") | not)
			| select({'"$filter"'}[.] | not)
			)
		| join(",")
	')
	if [[ -n "$extracoverpkg" ]]; then
		coverpkg="$extracoverpkg$coverpkg"
	fi

	args=""
	if [[ -n "$coverpkg" ]]; then
		args="-coverprofile $COVER/cover.${i}.out -covermode=count -coverpkg $coverpkg"
	fi

  echo go test ${args} "${pkg}"
  go test ${args} "${pkg}" 2>&1 | grep -v 'warning: no packages being tested depend on'
done

rm -f coverage.txt

# Merge cross-package coverage and then split the result into main and
# experimental coverages.
#
# We ignore packages in the form "footest" and any mock files.
gocovmerge "$COVER"/*.out \
	| filterIgnorePkgs \
	> coverage.txt
