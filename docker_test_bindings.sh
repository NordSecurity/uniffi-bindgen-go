#!/bin/bash
set -euxo pipefail

docker run \
    -ti --rm \
    --volume $HOME/go/pkg:/go/pkg \
    --volume $HOME/.cache/go-build:/root/.cache/go-build \
    --volume $PWD:/mounted_workdir \
    --workdir /mounted_workdir \
    golang:1.20 ./test_bindings.sh
