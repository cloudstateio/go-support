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

func TestCRDTGCounter(t *testing.T) {
	s := newServer(t)
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("CrdtInit", func(t *testing.T) {
		p := newProxy(ctx, s)
		s.t = t
		defer p.teardown()
		t.Run("sending CrdtInit fo an unknown service should fail", func(t *testing.T) {
			tr := tester{t}
			p.init(&entity.CrdtInit{ServiceName: "unknown", EntityId: "unknown"})
			resp, err := p.Recv()
			tr.expectedNil(err)
			tr.expectedNotNil(resp)
			tr.expectedBool(len(resp.GetFailure().GetDescription()) > 0, true)
		})
	})

	command := "ProcessGCounter"
	t.Run("GCounter", func(t *testing.T) {
		entityID := "gcounter-0"
		p := newProxy(ctx, s)
		defer p.teardown()
		t.Run("sending CrdtInit should not fail", func(t *testing.T) {
			tr := tester{t}
			p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
			resp, err := p.Recv()
			tr.expectedNil(err)
			tr.expectedNil(resp)
		})
		t.Run("incrementing a GCounter should emit a client action and create state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(
				entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 8}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				r := crdt.GCounterResponse{}
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedUInt64(r.GetValue().GetValue(), 8)
				tr.expectedUInt64(m.Reply.GetStateAction().GetCreate().GetGcounter().GetValue(), 8)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("a second increment should emit a client action and an update state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(
				entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 8}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				r := crdt.GCounterResponse{}
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedUInt64(r.GetValue().GetValue(), 16)
				tr.expectedUInt64(m.Reply.GetStateAction().GetUpdate().GetGcounter().GetIncrement(), 8)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("get should return the counters value", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(entityID, command, gcounterRequest(&crdt.Get{Key: entityID})).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				r := crdt.GCounterResponse{}
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedUInt64(r.GetValue().GetValue(), 16)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("the counter should apply new state and return its value", func(t *testing.T) {
			tr := tester{t}
			p.state(&entity.GCounterState{Value: 24})
			switch m := p.command(
				entityID, command, gcounterRequest(&crdt.Get{Key: entityID}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				r := crdt.GCounterResponse{}
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedUInt64(r.GetValue().GetValue(), 24)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("the counter should apply a delta and return its value", func(t *testing.T) {
			tr := tester{t}
			p.delta(&entity.GCounterDelta{Increment: 8})
			switch m := p.command(
				entityID, command, gcounterRequest(&crdt.Get{Key: entityID}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				r := crdt.GCounterResponse{}
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedUInt64(r.GetValue().GetValue(), 32)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("deleting an entity should emit a delete state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.command(
				entityID, command, gcounterRequest(&crdt.Delete{Key: entityID}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				tr.expectedNotNil(m.Reply.GetClientAction().GetReply())
				tr.expectedNotNil(m.Reply.GetStateAction().GetDelete())
			default:
				tr.unexpected(m)
			}
		})
		t.Run("after an entity was deleted, we could initialise an another entity", func(t *testing.T) {
			// this is not explicit specified by the spec, but it says, that the user function should
			// clear its state and the proxy could close the stream anytime, but also does not say
			// the user function can close the stream. So our implementation would be prepared for a
			// new entity re-using the same stream (why not).
			p.init(&entity.CrdtInit{
				ServiceName: serviceName,
				EntityId:    "gcounter-xyz",
			})
			// nothing should be returned here
			resp, err := p.Recv()
			if err != nil {
				t.Fatal(err)
			}
			if resp != nil {
				t.Fatal("no response expected")
			}
		})
	})

	t.Run("GCounter Streamed", func(t *testing.T) {
		command := "ProcessGCounterStreamed"
		entityID := "gcounter-x-0"
		p := newProxy(ctx, s)
		defer p.teardown()
		t.Run("sending CrdtInit should not fail", func(t *testing.T) {
			tr := tester{t}
			p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
			resp, err := p.Recv()
			tr.expectedNil(err)
			tr.expectedNil(resp)
		})
		t.Run("incrementing a GCounter should emit a client action and create state action", func(t *testing.T) {
			tr := tester{t}
			switch m := p.commandStreamed(
				entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 8}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				r := crdt.GCounterResponse{}
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedUInt64(r.GetValue().GetValue(), 8)
				tr.expectedUInt64(m.Reply.GetStateAction().GetCreate().GetGcounter().GetValue(), 8)
			default:
				tr.unexpected(m)
			}
		})
		t.Run("incrementing a GCounter should emit a client action and create state action", func(t *testing.T) {
			tr := tester{t}
			prevCmdID := p.seq - 1
			switch m := p.command(
				entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 8}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				r := crdt.GCounterResponse{}
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				tr.expectedUInt64(r.GetValue().GetValue(), 16)
				tr.expectedUInt64(m.Reply.GetStateAction().GetUpdate().GetGcounter().GetIncrement(), 8)
			default:
				tr.unexpected(m)
			}
			// validate streaming
			recv, err := p.Recv()
			if err != nil {
				t.Fatal(err)
			}
			// there must be a response with the command id from the first command sent.
			tr.expectedNotNil(recv)
			tr.expectedInt64(recv.GetStreamedMessage().GetCommandId(), prevCmdID)
			r := crdt.GCounterResponse{}
			tr.toProto(recv.GetStreamedMessage().GetClientAction().GetReply().GetPayload(), &r)
			tr.expectedUInt64(r.GetValue().GetValue(), 16)
		})
	})

	t.Run("GCounter – CrdtDelete", func(t *testing.T) {
		entityID := "gcounter-0"
		tr := tester{t}
		p := newProxy(ctx, s)
		defer p.teardown()

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		switch m := p.command(
			entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 8}),
		).Message.(type) {
		case *entity.CrdtStreamOut_Failure:
			tr.unexpected(m)
		}
		p.sendDelete(delete{&entity.CrdtDelete{}})
		// nothing should be returned here
		resp, err := p.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Fatal("no response expected")
		}
		entityID = "gcounter-0.1"
		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		switch m := p.command(
			entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 16}),
		).Message.(type) {
		case *entity.CrdtStreamOut_Reply:
			tr.expectedUInt64(m.Reply.GetStateAction().GetCreate().GetGcounter().GetValue(), 16)
		default:
			tr.unexpected(m)
		}
	})

	t.Run("GCounter – unknown entity id used", func(t *testing.T) {
		entityID := "gcounter-1"
		tr := tester{t}
		p := newProxy(ctx, s)
		defer p.teardown()
		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		switch m := p.command(
			entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 8}),
		).Message.(type) {
		case *entity.CrdtStreamOut_Failure:
			tr.unexpected(m)
		}
		t.Run("calling GetGCounter for a non existing entity id should fail", func(t *testing.T) {
			tr := tester{t}
			entityID := "gcounter-1-xxx"
			switch m := p.command(
				entityID, command, gcounterRequest(&crdt.Get{Key: entityID}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Failure:
			default:
				tr.unexpected(m)
			}
		})
	})

	t.Run("GCounter – incompatible CRDT delta sequence", func(t *testing.T) {
		entityID := "gcounter-2"
		p := newProxy(ctx, s)
		defer p.teardown()
		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		t.Run("setting a delta without ever sending state should fail", func(t *testing.T) {
			t.Skip("we can't test this one for now")
			p.delta(&entity.GCounterDelta{Increment: 7})
			resp, err := p.Recv()
			if err != nil {
				t.Fatal(err)
			}
			if resp == nil {
				t.Fatal("response expected")
			}
			switch m := resp.Message.(type) {
			case *entity.CrdtStreamOut_Failure:
				// the expected failure
			default:
				t.Fatalf("got unexpected message: %+v", m)
			}
		})
	})

	t.Run("GCounter – incompatible CRDT delta used", func(t *testing.T) {
		entityID := "gcounter-3"
		tr := tester{t}
		p := newProxy(ctx, s)
		defer p.teardown()

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		switch m := p.command(
			entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 8}),
		).Message.(type) {
		case *entity.CrdtStreamOut_Failure:
			tr.unexpected(m)
		}
		t.Run("setting a delta for a different CRDT type should fail", func(t *testing.T) {
			tr := tester{t}
			p.delta(&entity.PNCounterDelta{Change: 7})
			// nothing should be returned here
			resp, err := p.Recv()
			if err != nil {
				t.Fatal(err)
			}
			if resp == nil {
				t.Fatal("response expected")
			}
			switch m := resp.Message.(type) {
			case *entity.CrdtStreamOut_Failure:
			default:
				tr.unexpected(m)
			}
		})
	})

	// TODO: check if that is enough
	t.Run("GCounter – inconsistent local state", func(t *testing.T) {
		entityID := "gcounter-4"
		tr := tester{t}
		p := newProxy(ctx, s)
		defer p.teardown()

		p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})
		p.command(entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 7}))
		switch m := p.command(
			entityID, command, gcounterRequest(&crdt.GCounterIncrement{Key: entityID, Value: 7, FailWith: "error"}),
		).Message.(type) {
		case *entity.CrdtStreamOut_Failure:
			tr.expectedString(m.Failure.Description, "error")
		default:
			tr.unexpected(m)
		}

		// TODO: revise this tests as we should have a stream still up after a client failure
		// _, err :=
		// if err == nil {
		//	t.Fatal("expected err")
		// }
		// if err != io.EOF {
		//	t.Fatal("expected io.EOF")
		// }
		//
		// switch m := p.sendCmdRecvReply(command{
		//	&entity.Command{EntityID: entityID, Name: "IncrementGCounter"},
		//	&crdt.GCounterIncrement{Key: entityID, Value: 9},
		// }).Message.(type) {
		// case *entity.CrdtStreamOut_Reply:
		//	value := crdt.GCounterValue{}
		//	err := encoding.UnmarshalAny(m.Reply.GetClientAction().GetReply().GetPayload(), &value)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	if got, want := value.GetValue(), uint64(8+9); got != want {
		//		t.Fatalf("got = %v; wanted: %d, for value:%+v", got, want, value)
		//	}
		//	if got, want := m.Reply.GetStateAction().GetUpdate().GetGcounter().GetIncrement(), uint64(9); got != want {
		//		t.Fatalf("got = %v; wanted: %d, for value:%+v", got, want, m.Reply)
		//	}
		// default:
		//	t.Fatalf("got unexpected message: %+v", m)
		// }
	})
}
