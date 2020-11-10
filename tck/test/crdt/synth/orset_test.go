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

func TestCRDTORSet(t *testing.T) {
	s := newServer(t)
	s.newClientConn()
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	type pair struct {
		Left  string
		Right int64
	}
	t.Run("ORSet", func(t *testing.T) {
		entityID := "orset-1"
		command := "ProcessORSet"
		p := newProxy(ctx, s)
		defer p.teardown()

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		t.Run("ORSetAdd emits client action and create state action", func(t *testing.T) {
			tr := tester{t}
			one, err := encoding.Struct(pair{"one", 1})
			if err != nil {
				t.Fatal(err)
			}
			switch m := p.command(entityID, command, orsetRequest(&crdt.ORSetAdd{Value: &crdt.AnySupportType{
				Value: &crdt.AnySupportType_AnyValue{AnyValue: one}},
			}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				var r crdt.ORSetResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				one, err := encoding.Struct(pair{"one", 1})
				if err != nil {
					t.Fatal(err)
				}
				tr.expectedOneIn(r.GetValue().GetValues(), one)
				// state
				tr.expectedNotNil(m.Reply.GetStateAction().GetCreate())
				tr.expectedOneIn(m.Reply.GetStateAction().GetCreate().GetOrset().GetItems(), one)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("ORSetRemove emits client action and update state action", func(t *testing.T) {
			tr := tester{t}
			two, err := encoding.Struct(pair{"two", 2})
			if err != nil {
				t.Fatal(err)
			}
			p.command(entityID, command, orsetRequest(&crdt.ORSetAdd{Value: &crdt.AnySupportType{
				Value: &crdt.AnySupportType_AnyValue{AnyValue: two}},
			}))
			one, err := encoding.Struct(pair{"one", 1})
			if err != nil {
				t.Fatal(err)
			}
			switch m := p.command(entityID, command, orsetRequest(&crdt.ORSetRemove{Value: &crdt.AnySupportType{
				Value: &crdt.AnySupportType_AnyValue{AnyValue: one}},
			}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				tr.expectedNil(m.Reply.GetClientAction().GetForward())
				var r crdt.ORSetResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				two, err := encoding.Struct(pair{"two", 2})
				if err != nil {
					t.Fatal(err)
				}
				tr.expectedOneIn(r.GetValue().GetValues(), two)
				tr.expectedInt(len(r.GetValue().GetValues()), 1)
				// state
				tr.expectedNotNil(m.Reply.GetStateAction().GetUpdate())
				one, err := encoding.Struct(pair{"one", 1})
				if err != nil {
					t.Fatal(err)
				}
				tr.expectedOneIn(m.Reply.GetStateAction().GetUpdate().GetOrset().GetRemoved(), one)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("Delete emits client action and delete state action", func(t *testing.T) {
			tr := tester{t}
			two, err := encoding.Struct(pair{"two", 2})
			if err != nil {
				t.Fatal(err)
			}
			p.command(entityID, command, orsetRequest(&crdt.ORSetAdd{Value: &crdt.AnySupportType{
				Value: &crdt.AnySupportType_AnyValue{AnyValue: two}},
			}))
			switch m := p.command(
				entityID, command, orsetRequest(&crdt.Delete{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				tr.expectedNil(m.Reply.GetSideEffects())
				var r crdt.ORSetResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedNotNil(m.Reply.GetStateAction().GetDelete())
			default:
				tr.unexpected(m)
			}
		})
	})
}
