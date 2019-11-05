//
// Copyright 2019 Lightbend Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudstate

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

// marshalAny marshals a proto.Message to a any.Any value.
func marshalAny(pb interface{}) (*any.Any, error) {
	// TODO: protobufs are expected here, but Cloudstate supports other formats
	message, ok := pb.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("got a non-proto message as protobuf: %v", pb)
	}
	bytes, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", err, ErrMarshal)
	}
	return &any.Any{
		TypeUrl: fmt.Sprintf("%s/%s", protoAnyBase, proto.MessageName(message)),
		Value:   bytes,
	}, nil
}

// marshalEventsAny receives the events emitted through the handling of a command
// and marshals them to the event serialized form.
func marshalEventsAny(entityContext *EntityInstanceContext) ([]*any.Any, error) {
	events := make([]*any.Any, 0)
	if emitter, ok := entityContext.EntityInstance.Instance.(EventEmitter); ok {
		for _, evt := range emitter.Events() {
			event, err := marshalAny(evt)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
		emitter.Clear()
	}
	return events, nil
}
