#!/usr/bin/env bash
set -o nounset

function rnd() {
  cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w ${1:-32} | head -n 1
}

PROXY_IMAGE=${2:-cloudstateio/cloudstate-proxy-core:latest}
PROXY="cloudstate-proxy-$(rnd)"
TCK_IMAGE=${3:-cloudstateio/cloudstate-tck:latest}
TCK="cloudstate-tck-$(rnd)"

finally() {
  docker rm -f "$PROXY"
}
trap finally EXIT
set -x

docker run -d --name "$PROXY" -p 9000:9000 -e USER_FUNCTION_HOST=host.docker.internal -e USER_FUNCTION_PORT=8090 "${PROXY_IMAGE}" -Dconfig.resource=dev-mode.conf || exit $?
docker run --rm --name cloudstate-tck -p 8090:8090 -e TCK_HOST=0.0.0.0 -e TCK_PROXY_HOST=host.docker.internal -e TCK_SERVICE_HOST=host.docker.internal "${TCK_IMAGE}"
tck_status=$?

exit $tck_status
