#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

function fetch() {
  local path=$1
  mkdir -p protobuf/$(dirname $path)
  curl -o protobuf/${path} https://raw.githubusercontent.com/cloudstateio/cloudstate/master/protocols/${path}
  #sed 's/^option java_package.*/option go_package = "${go_package}";/' protobuf/${path}
}

# CloudState protocol
fetch "protocol/cloudstate/entity.proto"
fetch "protocol/cloudstate/event_sourced.proto"
fetch "protocol/cloudstate/function.proto"
fetch "protocol/cloudstate/crdt.proto"

# TCK shopping cart example
fetch "example/shoppingcart/shoppingcart.proto"
fetch "example/shoppingcart/persistence/domain.proto"

# CloudState frontend
fetch "frontend/cloudstate/entity_key.proto"

# dependencies
fetch "proxy/grpc/reflection/v1alpha/reflection.proto"
fetch "frontend/google/api/annotations.proto"
fetch "frontend/google/api/http.proto"
