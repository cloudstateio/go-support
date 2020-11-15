#!/usr/bin/env bash
set -o nounset

function rnd() {
  cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w ${1:-32} | head -n 1
}

PROXY_IMAGE=${2:-cloudstateio/cloudstate-proxy-core:latest}
PROXY="cloudstate-proxy-$(rnd)"

set -x
# run the proxy
docker run --rm --name "$PROXY" -p 9000:9000 -e USER_FUNCTION_HOST=host.docker.internal -e USER_FUNCTION_PORT=8090 \
  "${PROXY_IMAGE}" -Dcloudstate.proxy.passivation-timeout=30s -Dconfig.resource=dev-mode.conf || exit $?
tck_status=$?

exit $tck_status
