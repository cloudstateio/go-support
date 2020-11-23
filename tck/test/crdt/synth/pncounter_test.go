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

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/tck/crdt"
)

func TestCRDTPNCounter(t *testing.T) {
	s := newServer(t)
	s.newClientConn()
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("PNCounter", func(t *testing.T) {
		entityID := "pncounter-1"
		command := "ProcessPNCounter"
		p := newProxy(ctx, s)
		defer p.teardown()
		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})

		t.Run("incrementing a PNCounter should emit client action and create-state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(
				entityID, command, pncounterRequest(&crdt.PNCounterIncrement{Key: entityID, Value: 7}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				var r crdt.PNCounterResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedInt64(r.GetValue().GetValue(), 7)
				// tr.expectedInt64(m.Reply.GetStateAction().GetCreate().GetPncounter().GetValue(), 7)
				tr.expectedInt64(m.Reply.GetStateAction().GetUpdate().GetPncounter().GetChange(), 7)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("a second increment should emit a client action and an update-state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(
				entityID, command, pncounterRequest(&crdt.PNCounterIncrement{Key: entityID, Value: 7}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				var r crdt.PNCounterResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedInt64(r.GetValue().GetValue(), 14)
				tr.expectedInt64(m.Reply.GetStateAction().GetUpdate().GetPncounter().GetChange(), 7)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("a decrement should emit a client action and an update state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(
				entityID, command, pncounterRequest(&crdt.PNCounterDecrement{Key: entityID, Value: 28}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				var r crdt.PNCounterResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedInt64(r.GetValue().GetValue(), -14)
				tr.expectedInt64(m.Reply.GetStateAction().GetUpdate().GetPncounter().GetChange(), -28)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("the counter should apply a delta and return its value", func(t *testing.T) {
			tr := tester{t}
			p.delta(&entity.PNCounterDelta{Change: -56})
			switch m := p.command(
				entityID, command, pncounterRequest(&crdt.Get{Key: entityID}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				var r crdt.PNCounterResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedInt64(r.GetValue().GetValue(), -70)
			default:
				tr.unexpected(m)
			}
		})
	})
}
