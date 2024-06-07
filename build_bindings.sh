#!/bin/bash
set -euxo pipefail

SCRIPT_DIR="${SCRIPT_DIR:-$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )}"
ROOT_DIR="$SCRIPT_DIR"

if [[ "$OSTYPE" == "darwin"* ]]; then
ARCH=$(arch)

if [[ "$ARCH" == "arm64" ]]; then
TARGET="aarch64-apple-darwin"
else
TARGET="x86_64-apple-darwin"
fi

BINARIES_DIR="$ROOT_DIR/target/$TARGET/debug"
else 
BINARIES_DIR="$ROOT_DIR/target/debug"
fi

BINDINGS_DIR="$ROOT_DIR/binding_tests/generated"

rm -rf $BINDINGS_DIR
mkdir $BINDINGS_DIR

# FIXME: It would be better to generate and build fixtures one by one, instead of combining
# them all into the same library

if [[ "$OSTYPE" == "darwin"* ]]; then
LIB_FILE="$BINARIES_DIR/libuniffi_fixtures.dylib"
else 
LIB_FILE="$BINARIES_DIR/libuniffi_fixtures.so"
fi
$BINARIES_DIR/uniffi-bindgen-go $LIB_FILE --out-dir "$BINDINGS_DIR" --library --config "$ROOT_DIR/fixtures/uniffi.toml"
