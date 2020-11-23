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

package crdt2

import (
	"errors"
	"reflect"
	"sort"
	"strings"

	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type CrdtTckModelEntity struct {
	gCounter    *crdt.GCounter
	pnCounter   *crdt.PNCounter
	gSet        *crdt.GSet
	orSet       *crdt.ORSet
	flag        *crdt.Flag
	lwwRegister *crdt.LWWRegister
	vote        *crdt.Vote
	orMap       *crdt.ORMap
}

func NewCrdtTckModelEntity(id crdt.EntityID) crdt.EntityHandler {
	return &CrdtTckModelEntity{}
}

func NewCrdtTwoEntity(id crdt.EntityID) crdt.EntityHandler {
	return &CrdtTwoEntity{}
}

func (e *CrdtTckModelEntity) HandleCommand(ctx *crdt.CommandContext, name string, msg proto.Message) (*any.Any, error) {
	if ctx.Streamed() {
		r, ok := msg.(*StreamedRequest)
		if !ok {
			return nil, nil
		}
		ctx.ChangeFunc(func(c *crdt.CommandContext) (*any.Any, error) {
			for _, effect := range r.GetEffects() {
				req, err := encoding.MarshalAny(&Request{
					Id:      effect.GetId(),
					Actions: nil,
				})
				if err != nil {
					return nil, err
				}
				c.SideEffect(&protocol.SideEffect{
					ServiceName: "cloudstate.tck.model.crdt.CrdtTwo",
					CommandName: "Call",
					Payload:     req,
					Synchronous: effect.GetSynchronous(),
				})

				if r.GetEndState() != nil {
					state, err := crdtState(c.CRDT())
					if err != nil {
						return nil, err
					}
					if reflect.DeepEqual(state, c.CRDT()) {
						ctx.EndStream()
					}
				}
				if r.GetCancelUpdate() != nil {
					c.CancelFunc(func(c *crdt.CommandContext) error {
						r.GetCancelUpdate()
						return nil
					})
				}
			}
			state, err := crdtState(c.CRDT())
			if err != nil {
				return nil, err
			}
			return encoding.MarshalAny(&Response{
				State: state,
			})
		})
	}

	r, ok := msg.(*Request)
	if !ok {
		return nil, nil
	}
	forwarding := false
	for _, action := range r.GetActions() {
		switch a := action.Action.(type) {
		case *RequestAction_Update:
			if err := applyUpdate(ctx.CRDT(), a.Update); err != nil {
				return nil, err
			}
			switch a.Update.GetWriteConsistency() {
			case UpdateWriteConsistency_LOCAL:
				ctx.WriteConsistency(entity.CrdtWriteConsistency_LOCAL)
			case UpdateWriteConsistency_MAJORITY:
				ctx.WriteConsistency(entity.CrdtWriteConsistency_MAJORITY)
			case UpdateWriteConsistency_ALL:
				ctx.WriteConsistency(entity.CrdtWriteConsistency_ALL)
			}
		case *RequestAction_Delete:
			ctx.Delete()
		case *RequestAction_Forward:
			forwarding = true
			r, err := encoding.MarshalAny(&Request{
				Id:      a.Forward.GetId(),
				Actions: nil,
			})
			if err != nil {
				return nil, err
			}
			ctx.Forward(&protocol.Forward{
				ServiceName: "cloudstate.tck.model.crdt.CrdtTwo",
				CommandName: "Call",
				Payload:     r,
			})
		case *RequestAction_Effect:
			r, err := encoding.MarshalAny(&Request{
				Id:      a.Effect.GetId(),
				Actions: nil,
			})
			if err != nil {
				return nil, err
			}
			ctx.SideEffect(&protocol.SideEffect{
				ServiceName: "cloudstate.tck.model.crdt.CrdtTwo",
				CommandName: "Call",
				Payload:     r,
				Synchronous: a.Effect.GetSynchronous(),
			})
		case *RequestAction_Fail:
			return nil, protocol.ClientError{
				Err: errors.New(a.Fail.GetMessage()),
			}
		}
	}

	if forwarding {
		return nil, nil
	}
	state, err := crdtState(ctx.CRDT())
	if err != nil {
		return nil, err
	}
	return encoding.MarshalAny(&Response{
		State: state,
	})
}

func crdtState(c crdt.CRDT) (*State, error) {
	var state = &State{}
	switch t := c.(type) {
	case *crdt.GCounter:
		state.Value = &State_Gcounter{&GCounterValue{Value: t.Value()}}
	case *crdt.PNCounter:
		state.Value = &State_Pncounter{Pncounter: &PNCounterValue{Value: t.Value()}}
	case *crdt.GSet:
		set := make([]string, 0, len(t.Value()))
		for i, a := range t.Value() {
			set[i] = encoding.DecodeString(a)
		}
		sort.Strings(set)
		state.Value = &State_Gset{Gset: &GSetValue{
			Elements: set,
		}}
	case *crdt.ORSet:
		set := make([]string, 0, len(t.Value()))
		for i, a := range t.Value() {
			set[i] = encoding.DecodeString(a)
		}
		sort.Strings(set)
		state.Value = &State_Orset{Orset: &ORSetValue{
			Elements: set,
		}}
	case *crdt.LWWRegister:
		state.Value = &State_Lwwregister{Lwwregister: &LWWRegisterValue{
			Value: encoding.DecodeString(t.Value()),
		}}
	case *crdt.ORMap:
		values := make([]*ORMapEntryValue, len(t.Entries()))
		for i, entry := range t.Entries() {
			s, err := crdtState(entry.Value)
			if err != nil {
				return nil, err
			}
			values[i] = &ORMapEntryValue{
				Key:   encoding.DecodeString(entry.Key),
				Value: s,
			}
		}
		state.Value = &State_Ormap{Ormap: &ORMapValue{
			Entries: values,
		}}
	case *crdt.Vote:
		state.Value = &State_Vote{Vote: &VoteValue{
			SelfVote:    t.SelfVote(),
			VotesFor:    int32(t.VotesFor()),
			TotalVoters: int32(t.Voters()),
		}}
	}
	return state, nil
}

func applyUpdate(c crdt.CRDT, update *Update) error {
	switch u := update.GetUpdate().(type) {
	case *Update_Gcounter:
		c.(*crdt.GCounter).Increment(u.Gcounter.GetIncrement())
	case *Update_Pncounter:
		c.(*crdt.PNCounter).Increment(u.Pncounter.GetChange())
	case *Update_Gset:
		c.(*crdt.GSet).Add(encoding.String(u.Gset.GetAdd()))
	case *Update_Orset:
		switch ac := u.Orset.Action.(type) {
		case *ORSetUpdate_Add:
			c.(*crdt.ORSet).Add(encoding.String(ac.Add))
		case *ORSetUpdate_Remove:
			c.(*crdt.ORSet).Remove(encoding.String(ac.Remove))
		case *ORSetUpdate_Clear:
			if ac.Clear {
				c.(*crdt.ORSet).Clear()
			}
		}
	case *Update_Lwwregister:
		register := c.(*crdt.LWWRegister)
		if u.Lwwregister.GetClock() == nil {
			register.Set(encoding.String(u.Lwwregister.GetValue()))
			break
		}
		switch u.Lwwregister.GetClock().GetClockType() {
		case LWWRegisterClockType_DEFAULT:
			register.Set(encoding.String(u.Lwwregister.GetValue()))
		case LWWRegisterClockType_REVERSE:
			register.SetWithClock(
				encoding.String(u.Lwwregister.GetValue()), crdt.Reverse, 0,
			)
		case LWWRegisterClockType_CUSTOM:
			register.SetWithClock(
				encoding.String(u.Lwwregister.GetValue()),
				crdt.Custom,
				u.Lwwregister.GetClock().GetCustomClockValue(),
			)
		}
	case *Update_Flag:
		c.(*crdt.Flag).Enable()
	case *Update_Ormap:
		switch a := u.Ormap.Action.(type) {
		case *ORMapUpdate_Add:
			if c.(*crdt.ORMap).HasKey(encoding.String(a.Add)) {
				break
			}
			newCRDT, err := createCRDT(crdt.EntityID(a.Add))
			if err != nil {
				return err
			}
			c.(*crdt.ORMap).Set(encoding.String(a.Add), newCRDT)
			if err = applyUpdate(newCRDT, u.Ormap.GetUpdate().GetUpdate()); err != nil {
				return err
			}
		case *ORMapUpdate_Update:
			id := crdt.EntityID(u.Ormap.GetUpdate().GetKey())
			var crdtValue crdt.CRDT
			if !c.(*crdt.ORMap).HasKey(encoding.String(u.Ormap.GetUpdate().GetKey())) {
				created, err := createCRDT(id)
				if err != nil {
					return err
				}
				crdtValue = created
			}
			if err := applyUpdate(crdtValue, u.Ormap.GetUpdate().GetUpdate()); err != nil {
				return err
			}
		case *ORMapUpdate_Remove:
			c.(*crdt.ORMap).Delete(encoding.String(a.Remove))
		case *ORMapUpdate_Clear:
			c.(*crdt.ORMap).Clear()
		}
	case *Update_Vote:
		c.(*crdt.Vote).Vote(u.Vote.GetSelfVote())
	}
	return nil
}

func (CrdtTckModelEntity) Default(ctx *crdt.Context) (crdt.CRDT, error) {
	return createCRDT(ctx.EntityID)
}

func createCRDT(id crdt.EntityID) (crdt.CRDT, error) {
	switch strings.Split(id.String(), "-")[0] {
	case "GCounter":
		return crdt.NewGCounter(), nil
	case "PNCounter":
		return crdt.NewPNCounter(), nil
	case "GSet":
		return crdt.NewGSet(), nil
	case "ORSet":
		return crdt.NewORSet(), nil
	case "Flag":
		return crdt.NewFlag(), nil
	case "LWWregister":
		return crdt.NewLWWRegister(nil), nil
	case "Vote":
		return crdt.NewVote(), nil
	case "ORMap":
		return crdt.NewORMap(), nil
	default:
		return nil, errors.New("unknown entity type")
	}
}

func (e *CrdtTckModelEntity) Set(ctx *crdt.Context, state crdt.CRDT) error {
	switch s := state.(type) {
	case *crdt.GCounter:
		e.gCounter = s
	case *crdt.PNCounter:
		e.pnCounter = s
	case *crdt.GSet:
		e.gSet = s
	case *crdt.ORSet:
		e.orSet = s
	case *crdt.Flag:
		e.flag = s
	case *crdt.LWWRegister:
		e.lwwRegister = s
	case *crdt.Vote:
		e.vote = s
	case *crdt.ORMap:
		e.orMap = s
	}
	return nil
}

type CrdtTwoEntity struct{}

func (CrdtTwoEntity) HandleCommand(ctx *crdt.CommandContext, name string, msg proto.Message) (*any.Any, error) {
	return encoding.MarshalAny(&Response{})
}

func (CrdtTwoEntity) Default(ctx *crdt.Context) (crdt.CRDT, error) {
	return nil, nil
}

func (CrdtTwoEntity) Set(ctx *crdt.Context, state crdt.CRDT) error {
	return nil
}
