#!/bin/bash
set -euxo pipefail

if [[ "$OSTYPE" == "darwin"* ]]; then
ARCH=$(arch)

if [[ "$ARCH" == "arm64" ]]; then
cargo build --package uniffi-bindgen-go --package uniffi-bindgen-go-fixtures --target aarch64-apple-darwin
else
cargo build --package uniffi-bindgen-go --package uniffi-bindgen-go-fixtures --target x86_64-apple-darwin
fi

else
cargo build --package uniffi-bindgen-go --package uniffi-bindgen-go-fixtures
fi
