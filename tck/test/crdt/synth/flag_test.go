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

func TestCRDTFlag(t *testing.T) {
	s := newServer(t)
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("Flag", func(t *testing.T) {
		entityID := "flag-1"
		command := "ProcessFlag"
		p := newProxy(ctx, s)
		defer p.teardown()

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		t.Run("Get emits client action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				flagRequest(&crdt.Get{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				tr.expectedNotNil(m.Reply.GetClientAction())
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetStateAction())
				// action reply
				tr.expectedNotNil(m.Reply.GetClientAction().GetReply())
				var f crdt.FlagValue
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &f)
				tr.expectedFalse(f.GetValue())
			default:
				tr.unexpected(m)
			}
		})
		t.Run("FlagEnable emits client action and create state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				flagRequest(&crdt.FlagEnable{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				var f crdt.FlagResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &f)
				tr.expectedTrue(f.GetValue().GetValue())
				// state action
				tr.expectedNil(m.Reply.GetStateAction().GetUpdate())
				tr.expectedNil(m.Reply.GetStateAction().GetDelete())
				tr.expectedNotNil(m.Reply.GetStateAction().GetCreate())
				tr.expectedTrue(m.Reply.GetStateAction().GetCreate().GetFlag().GetValue())
			default:
				tr.unexpected(m)
			}
		})
		t.Run("Delete emits client action and delete state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				flagRequest(&crdt.Delete{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				var f crdt.FlagResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &f)
				// state action
				tr.expectedNil(m.Reply.GetStateAction().GetCreate())
				tr.expectedNil(m.Reply.GetStateAction().GetUpdate())
				tr.expectedNotNil(m.Reply.GetStateAction().GetDelete())
			default:
				tr.unexpected(m)
			}
		})
		t.Run("FlagState should reflect state", func(t *testing.T) {
			tr := tester{t}
			p := newProxy(ctx, s)
			defer p.teardown()

			entityID = "flag-2"
			p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
			p.state(&entity.FlagState{Value: true})
			switch m := p.command(entityID, command,
				flagRequest(&crdt.Get{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				tr.expectedNil(m.Reply.GetStateAction().GetCreate())
				var f crdt.FlagResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &f)
				tr.expectedTrue(f.GetValue().GetValue())
				// state action
				tr.expectedNil(m.Reply.GetStateAction().GetCreate())
				tr.expectedNil(m.Reply.GetStateAction().GetUpdate())
				tr.expectedNil(m.Reply.GetStateAction().GetDelete())
			default:
				tr.unexpected(m)
			}
		})
		t.Run("FlagDelta should reflect state", func(t *testing.T) {
			tr := tester{t}
			p := newProxy(ctx, s)
			defer p.teardown()

			entityID = "flag-3"
			p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
			p.delta(&entity.FlagDelta{Value: true})
			switch m := p.command(entityID, command,
				flagRequest(&crdt.Get{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				var f crdt.FlagResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &f)
				tr.expectedTrue(f.GetValue().GetValue())
				// state action
				tr.expectedNil(m.Reply.GetStateAction().GetCreate())
				tr.expectedNil(m.Reply.GetStateAction().GetUpdate())
				tr.expectedNil(m.Reply.GetStateAction().GetDelete())
			default:
				tr.unexpected(m)
			}
		})
	})
}
