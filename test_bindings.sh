#!/bin/bash
set -euxo pipefail

SCRIPT_DIR="${SCRIPT_DIR:-$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )}"
ROOT_DIR="$SCRIPT_DIR"

BINDINGS_DIR="$ROOT_DIR/binding_tests/generated"
BINARIES_DIR="$ROOT_DIR/target/debug"

rm -rf $BINDINGS_DIR
mkdir $BINDINGS_DIR

# FIXME: It would be better to generate and build fixtures one by one, instead of combining
# them all into the same library

target/debug/uniffi-bindgen-go "$BINARIES_DIR/libuniffi_fixtures.dylib" --out-dir "$BINDINGS_DIR" --library --config "$ROOT_DIR/fixtures/uniffi.toml"

pushd $BINDINGS_DIR/..
LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-}:$BINARIES_DIR" \
	CGO_LDFLAGS="-luniffi_fixtures -L$BINARIES_DIR -lm -ldl" \
	CGO_ENABLED=1 \
	go test -v
