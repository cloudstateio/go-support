#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

# CloudState protocol
protoc --go_out=plugins=grpc,paths=source_relative:. --proto_path=protobuf/frontend/ protobuf/frontend/cloudstate/entity_key.proto
protoc --go_out=plugins=grpc:. --proto_path=protobuf/protocol/ protobuf/protocol/cloudstate/entity.proto
protoc --go_out=plugins=grpc:. --proto_path=protobuf/protocol/ protobuf/protocol/cloudstate/event_sourced.proto
protoc --go_out=plugins=grpc:. --proto_path=protobuf/protocol/ protobuf/protocol/cloudstate/function.proto

# TCK shopping cart sample
protoc --go_out=plugins=grpc:. --proto_path=protobuf/protocol/ --proto_path=protobuf/frontend/ --proto_path=protobuf/proxy/ --proto_path=protobuf/example/ protobuf/example/shoppingcart/shoppingcart.proto
protoc --go_out=plugins=grpc,paths=source_relative:tck/shoppingcart/persistence --proto_path=protobuf/protocol/ --proto_path=protobuf/frontend/ --proto_path=protobuf/proxy/ --proto_path=protobuf/example/shoppingcart/persistence/ protobuf/example/shoppingcart/persistence/domain.proto
