#!/bin/bash
set -euxo pipefail

cargo build --package uniffi-bindgen-go --package uniffi-bindgen-go-fixtures
