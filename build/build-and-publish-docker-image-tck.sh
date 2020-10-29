#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

DOCKER_BUILDKIT=1 docker build -t cloudstateio/cloudstate-go-tck:latest -f ./build/TCK.Dockerfile . || exit $?
docker push cloudstateio/cloudstate-go-tck:latest
