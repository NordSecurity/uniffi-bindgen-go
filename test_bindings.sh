#!/bin/bash
set -euxo pipefail

SCRIPT_DIR="${SCRIPT_DIR:-$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )}"
ROOT_DIR="$SCRIPT_DIR"

BINDINGS_DIR="$ROOT_DIR/binding_tests/generated"
BINARIES_DIR="$ROOT_DIR/target/debug"

rm -rf $BINDINGS_DIR
mkdir $BINDINGS_DIR
function bindings() {
    target/debug/uniffi-bindgen-go $1 --out-dir "$BINDINGS_DIR" --config="uniffi-test-fixtures.toml"
}

bindings 3rd-party/uniffi-rs/examples/arithmetic/src/arithmetic.udl
bindings 3rd-party/uniffi-rs/examples/callbacks/src/callbacks.udl
bindings 3rd-party/uniffi-rs/examples/type-limits/src/custom-types.udl
bindings 3rd-party/uniffi-rs/examples/geometry/src/geometry.udl
bindings 3rd-party/uniffi-rs/examples/rondpoint/src/rondpoint.udl
bindings 3rd-party/uniffi-rs/examples/sprites/src/sprites.udl
bindings 3rd-party/uniffi-rs/examples/todolist/src/todolist.udl
bindings 3rd-party/uniffi-rs/fixtures/callbacks/src/callbacks.udl
bindings 3rd-party/uniffi-rs/fixtures/coverall/src/coverall.udl
bindings 3rd-party/uniffi-rs/fixtures/uniffi-fixture-time/src/chronological.udl
bindings fixtures/destroy/src/destroy.udl
bindings fixtures/errors/src/errors.udl
bindings fixtures/objects/src/objects.udl

pushd $BINDINGS_DIR/..
LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-}:$BINARIES_DIR" \
	CGO_LDFLAGS="-luniffi_fixtures -L$BINARIES_DIR -lm -ldl" \
	CGO_ENABLED=1 \
	go test -v
