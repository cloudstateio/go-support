#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

protoc --go-grpc_out=paths=source_relative:cloudstate/protocol --proto_path=protobuf/protocol/cloudstate entity.proto
protoc --go_out=paths=source_relative:cloudstate/protocol --proto_path=protobuf/protocol/cloudstate entity.proto
protoc --go-grpc_out=paths=source_relative:. --proto_path=protobuf/frontend/ cloudstate/entity_key.proto
protoc --go_out=paths=source_relative:. --proto_path=protobuf/frontend/ cloudstate/entity_key.proto
protoc --go-grpc_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol --proto_path=protobuf/protocol/cloudstate crdt.proto
protoc --go_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol --proto_path=protobuf/protocol/cloudstate crdt.proto
protoc --go-grpc_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ --proto_path=protobuf/protocol/cloudstate event_sourced.proto
protoc --go_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ --proto_path=protobuf/protocol/cloudstate event_sourced.proto
protoc --go-grpc_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ --proto_path=protobuf/protocol/cloudstate action.proto
protoc --go_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ --proto_path=protobuf/protocol/cloudstate action.proto
protoc --go-grpc_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ --proto_path=protobuf/protocol/cloudstate value_entity.proto
protoc --go_out=paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ --proto_path=protobuf/protocol/cloudstate value_entity.proto

# TCK CRDT
protoc --go-grpc_out=paths=source_relative:./tck/crdt \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck tck_crdt.proto
protoc --go_out=paths=source_relative:./tck/crdt \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck tck_crdt.proto

# TCK Eventsourced
protoc --go-grpc_out=paths=source_relative:./tck/eventsourced \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck/cloudstate/tck/model \
  --proto_path=protobuf/tck eventsourced.proto
protoc --go_out=paths=source_relative:./tck/eventsourced \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck/cloudstate/tck/model \
  --proto_path=protobuf/tck eventsourced.proto

# TCK Action
protoc --go-grpc_out=paths=source_relative:./tck/action \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck/cloudstate/tck/model \
  --proto_path=protobuf/tck tck_action.proto
protoc --go_out=paths=source_relative:./tck/action \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck/cloudstate/tck/model \
  --proto_path=protobuf/tck tck_action.proto

# TCK Value Entity
protoc --go-grpc_out=paths=source_relative:./tck/value_entity \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck/cloudstate/tck/model \
  --proto_path=protobuf/tck tck_valueentity.proto
protoc --go_out=paths=source_relative:./tck/value_entity \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck/cloudstate/tck/model \
  --proto_path=protobuf/tck tck_valueentity.proto

# CRDT shopping cart example
protoc --go-grpc_out=paths=source_relative:./example/crdt_shoppingcart/shoppingcart --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/crdt_shoppingcart/shoppingcart shoppingcart.proto hotitems.proto
protoc --go_out=paths=source_relative:./example/crdt_shoppingcart/shoppingcart --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/crdt_shoppingcart/shoppingcart shoppingcart.proto hotitems.proto

protoc --go-grpc_out=paths=source_relative:./example/crdt_shoppingcart/domain --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/crdt_shoppingcart/domain domain.proto
protoc --go_out=paths=source_relative:./example/crdt_shoppingcart/domain --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/crdt_shoppingcart/domain domain.proto

# shopping cart example
protoc --go-grpc_out=paths=source_relative:./example/shoppingcart/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/shoppingcart shoppingcart.proto
protoc --go_out=paths=source_relative:./example/shoppingcart/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/shoppingcart shoppingcart.proto
protoc --go-grpc_out=paths=source_relative:./example/shoppingcart/persistence \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/shoppingcart/persistence domain.proto
protoc --go_out=paths=source_relative:./example/shoppingcart/persistence \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/shoppingcart/persistence domain.proto

# chat example
protoc --go-grpc_out=paths=source_relative:./example/chat/presence/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/chat/presence/ example/chat/presence/presence.proto
protoc --go_out=paths=source_relative:./example/chat/presence/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/chat/presence/ example/chat/presence/presence.proto
protoc --go-grpc_out=paths=source_relative:./example/chat/friends/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/chat/friends/ example/chat/friends/friends.proto
protoc --go_out=paths=source_relative:./example/chat/friends/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/chat/friends/ example/chat/friends/friends.proto
