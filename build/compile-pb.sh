#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

# Cloudstate
protoc --go_out=plugins=grpc,paths=source_relative:cloudstate/protocol --proto_path=protobuf/protocol/cloudstate entity.proto
protoc --go_out=plugins=grpc,paths=source_relative:. --proto_path=protobuf/frontend/ cloudstate/entity_key.proto
protoc --go_out=plugins=grpc,paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol \
  --proto_path=protobuf/protocol/cloudstate crdt.proto
protoc --go_out=plugins=grpc,paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ \
  --proto_path=protobuf/protocol/cloudstate function.proto
protoc --go_out=plugins=grpc,paths=source_relative:cloudstate/entity --proto_path=protobuf/protocol/ \
  --proto_path=protobuf/protocol/cloudstate event_sourced.proto

# TCK CRDT
protoc --go_out=plugins=grpc,paths=source_relative:./tck/crdt \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck tck_crdt.proto

# TCK Eventsourced
protoc --go_out=plugins=grpc,paths=source_relative:./tck/eventsourced \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=protobuf/tck/cloudstate/tck/model \
  --proto_path=protobuf/tck eventsourced.proto

# CRDT shopping cart example
protoc --go_out=plugins=grpc,paths=source_relative:./example/crdt_shoppingcart/shoppingcart --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/crdt_shoppingcart/shoppingcart shoppingcart.proto hotitems.proto

protoc --go_out=plugins=grpc,paths=source_relative:./example/crdt_shoppingcart/domain --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/crdt_shoppingcart/domain domain.proto

# event sourced shopping cart example
protoc --go_out=plugins=grpc,paths=source_relative:./example/shoppingcart/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/shoppingcart shoppingcart.proto
protoc --go_out=plugins=grpc,paths=source_relative:./example/shoppingcart/persistence \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/shoppingcart/persistence domain.proto

# chat example
protoc --go_out=plugins=grpc,paths=source_relative:./example/chat/presence/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/chat/presence/ example/chat/presence/presence.proto
protoc --go_out=plugins=grpc,paths=source_relative:./example/chat/friends/ \
  --proto_path=protobuf/protocol \
  --proto_path=protobuf/frontend \
  --proto_path=protobuf/frontend/cloudstate \
  --proto_path=protobuf/proxy \
  --proto_path=example/chat/friends/ example/chat/friends/friends.proto
