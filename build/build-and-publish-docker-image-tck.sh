#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

docker build -t gcr.io/mrcllnz/cloudstate-go-tck:latest -f ./build/TCK.Dockerfile .
docker push gcr.io/mrcllnz/cloudstate-go-tck:latest