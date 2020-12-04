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
	"fmt"
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

func (e *CrdtTckModelEntity) processStreamed(ctx *crdt.CommandContext, name string, msg proto.Message) (*any.Any, error) {
	r, ok := msg.(*StreamedRequest)
	if !ok {
		return nil, nil
	}
	fmt.Printf("processStreamed: %+v\n", ctx.EntityID)
	if ctx.Streamed() {
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
			}
			if r.GetEndState() != nil {
				state, err := crdtState(c.CRDT())
				if err != nil {
					return nil, err
				}
				if proto.Equal(state, r.GetEndState()) {
					ctx.EndStream()
				}
			}
			if r.GetEmpty() {
				return nil, nil
			}
			state, err := crdtState(c.CRDT())
			if err != nil {
				return nil, err
			}
			return encoding.MarshalAny(&Response{
				State: state,
			})
		})
		if u := r.GetCancelUpdate(); u != nil {
			ctx.CancelFunc(func(c *crdt.CommandContext) error {
				return applyUpdate(c.CRDT(), u)
			})
		}
	}
	if u := r.GetInitialUpdate(); u != nil {
		if err := applyUpdate(ctx.CRDT(), u); err != nil {
			return nil, err
		}
	}
	if r.GetEmpty() {
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

func (e *CrdtTckModelEntity) HandleCommand(ctx *crdt.CommandContext, name string, msg proto.Message) (*any.Any, error) {
	if name == "ProcessStreamed" {
		return e.processStreamed(ctx, name, msg)
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
		set := make([]string, len(t.Value()))
		for i, a := range t.Value() {
			set[i] = encoding.DecodeString(a)
		}
		sort.Strings(set)
		state.Value = &State_Gset{Gset: &GSetValue{
			Elements: set,
		}}
	case *crdt.ORSet:
		set := make([]string, len(t.Value()))
		for i, a := range t.Value() {
			set[i] = encoding.DecodeString(a)
		}
		sort.Strings(set)
		state.Value = &State_Orset{Orset: &ORSetValue{
			Elements: set,
		}}
	case *crdt.LWWRegister:
		value := ""
		if t.Value() != nil {
			value = encoding.DecodeString(t.Value())
		}
		state.Value = &State_Lwwregister{Lwwregister: &LWWRegisterValue{
			Value: value,
		}}
	case *crdt.Flag:
		state.Value = &State_Flag{Flag: &FlagValue{
			Value: t.Value(),
		}}
	case *crdt.ORMap:
		values := make([]*ORMapEntryValue, len(t.Entries()))
		for i, entry := range t.Entries() {
			state, err := crdtState(entry.Value)
			if err != nil {
				return nil, err
			}
			values[i] = &ORMapEntryValue{
				Key:   encoding.DecodeString(entry.Key),
				Value: state,
			}
		}
		sort.Sort(sortedORMapEntryValues(values))
		state.Value = &State_Ormap{Ormap: &ORMapValue{
			Entries: values,
		}}
	case *crdt.Vote:
		state.Value = &State_Vote{Vote: &VoteValue{
			SelfVote:    t.SelfVote(),
			VotesFor:    int32(t.VotesFor()),
			TotalVoters: int32(t.Voters()),
		}}
	default:
		return nil, fmt.Errorf("state not created for: %v", c)
	}
	return state, nil
}

type sortedORMapEntryValues []*ORMapEntryValue

func (s sortedORMapEntryValues) Len() int           { return len(s) }
func (s sortedORMapEntryValues) Less(i, j int) bool { return s[i].Key < s[j].Key }
func (s sortedORMapEntryValues) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func applyUpdate(c crdt.CRDT, update *Update) error {
	switch u := update.GetUpdate().(type) {
	case *Update_Gcounter:
		c.(*crdt.GCounter).Increment(u.Gcounter.GetIncrement())
	case *Update_Pncounter:
		c.(*crdt.PNCounter).Increment(u.Pncounter.GetChange())
	case *Update_Gset:
		c.(*crdt.GSet).Add(encoding.String(u.Gset.GetAdd()))
	case *Update_Orset:
		switch a := u.Orset.Action.(type) {
		case *ORSetUpdate_Add:
			c.(*crdt.ORSet).Add(encoding.String(a.Add))
		case *ORSetUpdate_Remove:
			c.(*crdt.ORSet).Remove(encoding.String(a.Remove))
		case *ORSetUpdate_Clear:
			if a.Clear {
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
				encoding.String(u.Lwwregister.GetValue()),
				crdt.Reverse,
				0,
			)
		case LWWRegisterClockType_CUSTOM:
			register.SetWithClock(
				encoding.String(u.Lwwregister.GetValue()),
				crdt.Custom,
				u.Lwwregister.GetClock().GetCustomClockValue(),
			)
		case LWWRegisterClockType_CUSTOM_AUTO_INCREMENT:
			register.SetWithClock(
				encoding.String(u.Lwwregister.GetValue()),
				crdt.CustomAutoIncrement,
				u.Lwwregister.GetClock().GetCustomClockValue(),
			)
		}
	case *Update_Flag:
		c.(*crdt.Flag).Enable()
	case *Update_Ormap:
		orMap := c.(*crdt.ORMap)
		switch a := u.Ormap.Action.(type) {
		case *ORMapUpdate_Add:
			if orMap.HasKey(encoding.String(a.Add)) {
				break
			}
			created, err := createCRDT(crdt.EntityID(a.Add))
			if err != nil {
				return err
			}
			orMap.Set(encoding.String(a.Add), created)
			if err = applyUpdate(created, u.Ormap.GetUpdate().GetUpdate()); err != nil {
				return err
			}
		case *ORMapUpdate_Update:
			key := encoding.String(u.Ormap.GetUpdate().GetKey())
			value := orMap.Get(key)
			var err error
			if value == nil {
				value, err = createCRDT(crdt.EntityID(u.Ormap.GetUpdate().GetKey()))
				if err != nil {
					return err
				}
				orMap.Set(key, value)
			}
			if err := applyUpdate(value, u.Ormap.GetUpdate().GetUpdate()); err != nil {
				return err
			}
		case *ORMapUpdate_Remove:
			orMap.Delete(encoding.String(a.Remove))
		case *ORMapUpdate_Clear:
			orMap.Clear()
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
	case "LWWRegister":
		return crdt.NewLWWRegister(encoding.String("")), nil
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
