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

func TestCRDTVote(t *testing.T) {
	s := newServer(t)
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("Vote", func(t *testing.T) {
		entityID := "vote-1"
		command := "ProcessVote"
		p := newProxy(ctx, s)
		defer p.teardown()

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		t.Run("Get emits client action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				voteRequest(&crdt.Get{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				tr.expectedNil(m.Reply.GetStateAction().GetCreate())
				var r crdt.VoteResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedFalse(r.SelfVote)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("Vote emits client and state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command,
				voteRequest(&crdt.VoteVote{Value: true}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetClientAction().GetFailure())
				var r crdt.VoteResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedTrue(r.SelfVote)
				tr.expectedUInt32(r.VotesFor, 1)
				tr.expectedUInt32(r.Voters, 1)
				// state
				tr.expectedNotNil(m.Reply.GetStateAction().GetCreate())
				tr.expectedTrue(m.Reply.GetStateAction().GetCreate().GetVote().GetSelfVote())
				tr.expectedUInt32(m.Reply.GetStateAction().GetCreate().GetVote().GetVotesFor(), 1)
				tr.expectedUInt32(m.Reply.GetStateAction().GetCreate().GetVote().GetTotalVoters(), 1)
			default:
				tr.unexpected(m)
			}

			switch m := p.command(entityID, command,
				voteRequest(&crdt.VoteVote{Value: false}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				var r crdt.VoteResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedFalse(r.SelfVote)
				tr.expectedUInt32(r.VotesFor, 0)
				tr.expectedUInt32(r.Voters, 1)
				// state
				tr.expectedNotNil(m.Reply.GetStateAction().GetUpdate())
				tr.expectedFalse(m.Reply.GetStateAction().GetUpdate().GetVote().GetSelfVote())
				tr.expectedUInt32(uint32(m.Reply.GetStateAction().GetUpdate().GetVote().GetVotesFor()), 0)
				tr.expectedUInt32(uint32(m.Reply.GetStateAction().GetUpdate().GetVote().GetTotalVoters()), 1)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("VoteState", func(t *testing.T) {
			tr := tester{t}
			p.state(&entity.VoteState{
				TotalVoters: 6,
				VotesFor:    3,
				SelfVote:    true,
			})
			switch m := p.command(entityID, command,
				voteRequest(&crdt.Get{}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				// action reply
				tr.expectedNil(m.Reply.GetSideEffects())
				tr.expectedNil(m.Reply.GetStateAction())
				tr.expectedNotNil(m.Reply.GetClientAction().GetReply())
				var r crdt.VoteResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedTrue(r.SelfVote)
				tr.expectedUInt32(r.Voters, 6)
				tr.expectedUInt32(r.VotesFor, 3)
			default:
				tr.unexpected(m)
			}
		})
	})
	t.Run("VoteState", func(t *testing.T) {
		entityID := "vote-1"
		command := "ProcessVote"
		p := newProxy(ctx, s)
		defer p.teardown()
		tr := tester{t}

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		p.command(entityID, command,
			voteRequest(&crdt.VoteVote{Value: true}),
		)
		p.delta(&entity.VoteDelta{
			TotalVoters: 7,
			VotesFor:    3,
			SelfVote:    false,
		})
		switch m := p.command(entityID, command,
			voteRequest(&crdt.Get{}),
		).Message.(type) {
		case *entity.CrdtStreamOut_Reply:
			// action reply
			tr.expectedNil(m.Reply.GetSideEffects())
			tr.expectedNil(m.Reply.GetStateAction())
			tr.expectedNotNil(m.Reply.GetClientAction().GetReply())
			var r crdt.VoteResponse
			tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
			tr.expectedTrue(r.SelfVote) // Delta does not affect self vote!
			tr.expectedUInt32(r.Voters, 7)
			tr.expectedUInt32(r.VotesFor, 3)
		default:
			tr.unexpected(m)
		}
	})
}
