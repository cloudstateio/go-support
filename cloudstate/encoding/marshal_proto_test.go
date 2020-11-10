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

package encoding

// func TestMarshalAnyProto(t *testing.T) {
//	event := eventsourced.IncrementByEvent{Value: 29}
//	any, err := MarshalAny(&event)
//	if err != nil {
//		t.Fatalf("failed to MarshalAny: %v", err)
//	}
//	expected := fmt.Sprintf("%s/%s", cloudstate.protoAnyBase, "IncrementByEvent")
//	if expected != any.GetTypeUrl() {
//		t.Fatalf("any.GetTypeUrl: %s is not: %s", any.GetTypeUrl(), expected)
//	}
//	event2 := &eventsourced.IncrementByEvent{}
//	if err := proto.Unmarshal(any.Value, event2); err != nil {
//		t.Fatalf("%v", err)
//	}
//	if event2.Value != event.Value {
//		t.Fatalf("event2.Value: %d != event.Value: %d", event2.Value, event.Value)
//	}
// }
