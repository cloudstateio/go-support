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

package synth

import (
	"context"
	"testing"
	"time"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/tck/crdt"
)

func TestCRDTORMap(t *testing.T) {
	s := newServer(t)
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("ORMap", func(t *testing.T) {
		entityID := "ormap-1"
		command := "ProcessORMap"
		p := newProxy(ctx, s)
		defer p.teardown()

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		t.Run("Get", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				ormapRequest(&crdt.Get{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				tr.expectedNil(m.Reply.GetStateAction())
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNotNil(m.Reply.GetClientAction())
			default:
				tr.unexpected(m)
			}
		})
		t.Run("Set – GCounter", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				ormapRequest(&crdt.ORMapActionRequest{
					EntryKey: encoding.String("niner"),
					Request: &crdt.ORMapActionRequest_GCounterRequest{
						GCounterRequest: gcounterRequest(&crdt.GCounterIncrement{Value: 9}),
					},
				}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// tr.expectedNotNil(m.Reply.GetStateAction().GetCreate())
				tr.expectedNotNil(m.Reply.GetStateAction().GetUpdate())
				tr.expectedNotNil(m.Reply.GetClientAction())
			default:
				tr.unexpected(m)
			}
			switch m := p.command(entityID, command,
				ormapRequest(&crdt.ORMapActionRequest{
					EntryKey: encoding.String("niner"),
					Request: &crdt.ORMapActionRequest_GCounterRequest{
						GCounterRequest: gcounterRequest(&crdt.GCounterIncrement{Value: 18}),
					},
				}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				tr.expectedNotNil(m.Reply.GetClientAction())

				var r crdt.ORMapResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedInt(len(r.GetKeys().GetValues()), 1)
				tr.expectedOneIn(r.GetKeys().GetValues(), encoding.String("niner"))
				tr.expectedInt(len(r.GetEntries().GetValues()), 1)

				tr.expectedNotNil(m.Reply.GetStateAction().GetUpdate())
				tr.expectedInt(len(m.Reply.GetStateAction().GetUpdate().GetOrmap().GetUpdated()), 1)
				tr.expectedUInt64(
					m.Reply.GetStateAction().GetUpdate().GetOrmap().GetUpdated()[0].Delta.GetGcounter().Increment,
					18,
				)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("Delete", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				ormapRequest(&crdt.Delete{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				tr.expectedNotNil(m.Reply.GetStateAction().GetDelete())
			default:
				tr.unexpected(m)
			}
		})
	})
}
