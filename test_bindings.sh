#!/bin/bash
set -euxo pipefail

SCRIPT_DIR="${SCRIPT_DIR:-$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )}"
ROOT_DIR="$SCRIPT_DIR"

BINDINGS_DIR="$ROOT_DIR/binding_tests/generated"
BINARIES_DIR="$ROOT_DIR/target/debug"

pushd $BINDINGS_DIR/..
LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-}:$BINARIES_DIR" \
	CGO_LDFLAGS="-luniffi_fixtures -L$BINARIES_DIR -lm -ldl" \
	CGO_ENABLED=1 \
	go test -v
