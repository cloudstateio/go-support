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

package crdt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
)

type TestModel struct {
	id          crdt.EntityID
	gCounter    *crdt.GCounter
	pnCounter   *crdt.PNCounter
	gSet        *crdt.GSet
	orSet       *crdt.ORSet
	flag        *crdt.Flag
	lwwRegister *crdt.LWWRegister
	vote        *crdt.Vote
	orMap       *crdt.ORMap
}

func NewEntity(id crdt.EntityID) *TestModel {
	return &TestModel{id: id}
}

func (s *TestModel) Set(_ *crdt.Context, c crdt.CRDT) error {
	switch v := c.(type) {
	case *crdt.GCounter:
		s.gCounter = v
	case *crdt.PNCounter:
		s.pnCounter = v
	case *crdt.GSet:
		s.gSet = v
	case *crdt.ORSet:
		s.orSet = v
	case *crdt.Flag:
		s.flag = v
	case *crdt.LWWRegister:
		s.lwwRegister = v
	case *crdt.Vote:
		s.vote = v
	case *crdt.ORMap:
		s.orMap = v
	}
	return nil
}

func (s *TestModel) Default(c *crdt.Context) (crdt.CRDT, error) {
	switch strings.Split(c.EntityID.String(), "-")[0] {
	case "gcounter":
		return crdt.NewGCounter(), nil
	case "pncounter":
		return crdt.NewPNCounter(), nil
	case "gset":
		return crdt.NewGSet(), nil
	case "orset":
		return crdt.NewORSet(), nil
	case "flag":
		return crdt.NewFlag(), nil
	case "lwwregister":
		return crdt.NewLWWRegister(nil), nil
	case "vote":
		return crdt.NewVote(), nil
	case "ormap":
		return crdt.NewORMap(), nil
	default:
		return nil, errors.New("unknown entity type")
	}
}

func (s *TestModel) HandleCommand(cc *crdt.CommandContext, name string, cmd proto.Message) (*any.Any, error) {
	switch c := cmd.(type) {
	case *GCounterRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *GCounterRequestAction_Increment:
				if cc.Streamed() {
					cc.ChangeFunc(func(c *crdt.CommandContext) (*any.Any, error) {
						return encoding.MarshalAny(&GCounterResponse{
							Value: &GCounterValue{Value: s.gCounter.Value()}},
						)
					})
				}
				s.gCounter.Increment(a.Increment.GetValue())
				if with := a.Increment.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&GCounterResponse{
					Value: &GCounterValue{Value: s.gCounter.Value()}},
				)
			case *GCounterRequestAction_Get:
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&GCounterResponse{Value: &GCounterValue{Value: s.gCounter.Value()}})
			case *GCounterRequestAction_Delete:
				cc.Delete()
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&GCounterResponse{})
			}
		}
	case *PNCounterRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *PNCounterRequestAction_Increment:
				s.pnCounter.Increment(a.Increment.Value)
				if with := a.Increment.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&PNCounterResponse{Value: &PNCounterValue{Value: s.pnCounter.Value()}})
			case *PNCounterRequestAction_Decrement:
				s.pnCounter.Decrement(a.Decrement.Value)
				if with := a.Decrement.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&PNCounterResponse{Value: &PNCounterValue{Value: s.pnCounter.Value()}})
			case *PNCounterRequestAction_Get:
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&PNCounterResponse{Value: &PNCounterValue{Value: s.pnCounter.Value()}})
			case *PNCounterRequestAction_Delete:
				cc.Delete()
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&PNCounterResponse{})
			}
		}
	case *GSetRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *GSetRequestAction_Get:
				v := make([]*AnySupportType, 0, len(s.gSet.Value()))
				for _, a := range s.gSet.Value() {
					v = append(v, asAnySupportType(a))
				}
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&GSetResponse{Value: &GSetValue{Values: v}})
			case *GSetRequestAction_Delete:
				cc.Delete()
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&GSetResponse{})
			case *GSetRequestAction_Add:
				anySupportAdd(s.gSet, a.Add.Value)
				v := make([]*AnySupportType, 0, len(s.gSet.Value()))
				for _, a := range s.gSet.Value() {
					if strings.HasPrefix(a.TypeUrl, encoding.JSONTypeURLPrefix) {
						v = append(v, &AnySupportType{
							Value: &AnySupportType_AnyValue{AnyValue: a},
						})
						continue
					}
					v = append(v, asAnySupportType(a))
				}
				if with := a.Add.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&GSetResponse{Value: &GSetValue{Values: v}})
			}
		}
	case *ORSetRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *ORSetRequestAction_Get:
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&ORSetResponse{Value: &ORSetValue{Values: s.orSet.Value()}})
			case *ORSetRequestAction_Delete:
				cc.Delete()
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&ORSetResponse{})
			case *ORSetRequestAction_Add:
				anySupportAdd(s.orSet, a.Add.Value)
				if with := a.Add.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&ORSetResponse{Value: &ORSetValue{Values: s.orSet.Value()}})
			case *ORSetRequestAction_Remove:
				anySupportRemove(s.orSet, a.Remove.Value)
				if with := a.Remove.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&ORSetResponse{Value: &ORSetValue{Values: s.orSet.Value()}})
			}
		}
	case *FlagRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *FlagRequestAction_Get:
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&FlagResponse{Value: &FlagValue{Value: s.flag.Value()}})
			case *FlagRequestAction_Delete:
				cc.Delete()
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&FlagResponse{})
			case *FlagRequestAction_Enable:
				if with := a.Enable.FailWith; with != "" {
					return nil, errors.New(with)
				}
				s.flag.Enable()
				return encoding.MarshalAny(&FlagResponse{Value: &FlagValue{Value: s.flag.Value()}})
			}
		}
	case *LWWRegisterRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *LWWRegisterRequestAction_Get:
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&LWWRegisterResponse{Value: &LWWRegisterValue{Value: s.lwwRegister.Value()}})
			case *LWWRegisterRequestAction_Delete:
				cc.Delete()
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&FlagResponse{})
			case *LWWRegisterRequestAction_Set:
				if with := a.Set.FailWith; with != "" {
					return nil, errors.New(with)
				}
				anySupportAdd(&anySupportAdderSetter{s.lwwRegister}, a.Set.GetValue())
				return encoding.MarshalAny(&LWWRegisterResponse{Value: &LWWRegisterValue{Value: s.lwwRegister.Value()}})
			case *LWWRegisterRequestAction_SetWithClock:
				if with := a.SetWithClock.FailWith; with != "" {
					return nil, errors.New(with)
				}
				anySupportSetClock(
					s.lwwRegister,
					a.SetWithClock.GetValue(),
					crdt.Clock(uint64(a.SetWithClock.GetClock().Number())),
					a.SetWithClock.CustomClockValue,
				)
				return encoding.MarshalAny(&LWWRegisterResponse{Value: &LWWRegisterValue{Value: s.lwwRegister.Value()}})
			}
		}
	case *VoteRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *VoteRequestAction_Get:
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&VoteResponse{
					SelfVote: s.vote.SelfVote(),
					Voters:   s.vote.Voters(),
					VotesFor: s.vote.VotesFor(),
				})
			case *VoteRequestAction_Delete:
				cc.Delete()
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&empty.Empty{})
			case *VoteRequestAction_Vote:
				s.vote.Vote(a.Vote.GetValue())
				if with := a.Vote.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&VoteResponse{
					SelfVote: s.vote.SelfVote(),
					Voters:   s.vote.Voters(),
					VotesFor: s.vote.VotesFor(),
				})
			}
		}
	case *ORMapRequest:
		for _, as := range c.GetActions() {
			switch a := as.Action.(type) {
			case *ORMapRequestAction_Get:
				if with := a.Get.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(orMapResponse(s.orMap))
			case *ORMapRequestAction_Delete:
				if with := a.Delete.FailWith; with != "" {
					return nil, errors.New(with)
				}
				cc.Delete()
				return encoding.MarshalAny(orMapResponse(s.orMap))
			case *ORMapRequestAction_SetKey:
				// s.orMap.Set(a.SetKey.EntryKey, a.SetKey.Value)
				// TODO: how to set it?
				return encoding.MarshalAny(orMapResponse(s.orMap))
			case *ORMapRequestAction_Request:
				// we reuse this entities implementation to handle
				// requests for CRDT values of this ORMap.
				//
				var entityID string
				var req proto.Message
				switch t := a.Request.GetRequest().(type) {
				case *ORMapActionRequest_GCounterRequest:
					entityID = "gcounter"
					req = t.GCounterRequest
				case *ORMapActionRequest_FlagRequest:
					entityID = "flag"
					req = t.FlagRequest
				case *ORMapActionRequest_GsetRequest:
					entityID = "gset"
					req = t.GsetRequest
				case *ORMapActionRequest_LwwRegisterRequest:
					entityID = "lwwregister"
					req = t.LwwRegisterRequest
				case *ORMapActionRequest_OrMapRequest: // yeah, really!
					entityID = "ormap"
					req = t.OrMapRequest
				case *ORMapActionRequest_OrSetRequest:
					entityID = "orset"
					req = t.OrSetRequest
				case *ORMapActionRequest_VoteRequest:
					entityID = "pncounter"
					req = t.VoteRequest
				case *ORMapActionRequest_PnCounterRequest:
					entityID = "pncounter"
					req = t.PnCounterRequest
				}
				if err := s.runRequest(
					&crdt.CommandContext{Context: &crdt.Context{EntityID: crdt.EntityID(entityID)}},
					a.Request.GetEntryKey(),
					req,
				); err != nil {
					return nil, err
				}
				if with := a.Request.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(orMapResponse(s.orMap))
			case *ORMapRequestAction_DeleteKey:
				s.orMap.Delete(a.DeleteKey.EntryKey)
				if with := a.DeleteKey.FailWith; with != "" {
					return nil, errors.New(with)
				}
				return encoding.MarshalAny(&ORMapResponse{})
			}
		}
	}
	return nil, errors.New("unhandled command")
}

func orMapResponse(orMap *crdt.ORMap) *ORMapResponse {
	r := &ORMapResponse{
		Keys: &ORMapKeys{Values: orMap.Keys()},
		Entries: &ORMapEntries{
			Values: make([]*ORMapEntry, 0),
		},
	}
	for _, k := range orMap.Keys() {
		var value *any.Any
		switch s := orMap.Get(k).(type) {
		case *crdt.GCounter:
			val, err := encoding.Struct(s)
			if err != nil {
				panic(err)
			}
			value = val
		}
		r.Entries.Values = append(r.Entries.Values,
			&ORMapEntry{
				EntryKey: k,
				Value:    value,
			},
		)
	}
	return r
}

// runRequest runs the given action request using a temporary TestModel
// with the requests corresponding CRDT as its state.
// runRequest adds the CRDT if not already present in the map.
func (s *TestModel) runRequest(ctx *crdt.CommandContext, key *any.Any, req proto.Message) error {
	m := &TestModel{}
	if !s.orMap.HasKey(key) {
		c, err := m.Default(ctx.Context)
		if err != nil {
			return err
		}
		s.orMap.Set(key, c) // triggers a set delta
	}
	if err := m.Set(ctx.Context, s.orMap.Get(key)); err != nil {
		return err
	}
	if _, err := m.HandleCommand(ctx, "", req); err != nil {
		return err
	}
	return nil
}

func asAnySupportType(x *any.Any) *AnySupportType {
	switch x.TypeUrl {
	case encoding.PrimitiveTypeURLPrefixBool:
		return &AnySupportType{
			Value: &AnySupportType_BoolValue{BoolValue: encoding.DecodeBool(x)},
		}
	case encoding.PrimitiveTypeURLPrefixBytes:
		return &AnySupportType{
			Value: &AnySupportType_BytesValue{BytesValue: encoding.DecodeBytes(x)},
		}
	case encoding.PrimitiveTypeURLPrefixFloat:
		return &AnySupportType{
			Value: &AnySupportType_FloatValue{FloatValue: encoding.DecodeFloat32(x)},
		}
	case encoding.PrimitiveTypeURLPrefixDouble:
		return &AnySupportType{
			Value: &AnySupportType_DoubleValue{DoubleValue: encoding.DecodeFloat64(x)},
		}
	case encoding.PrimitiveTypeURLPrefixInt32:
		return &AnySupportType{
			Value: &AnySupportType_Int32Value{Int32Value: encoding.DecodeInt32(x)},
		}
	case encoding.PrimitiveTypeURLPrefixInt64:
		return &AnySupportType{
			Value: &AnySupportType_Int64Value{Int64Value: encoding.DecodeInt64(x)},
		}
	case encoding.PrimitiveTypeURLPrefixString:
		return &AnySupportType{
			Value: &AnySupportType_StringValue{StringValue: encoding.DecodeString(x)},
		}
	}
	panic(fmt.Sprintf("no mapping found for TypeUrl: %v", x.TypeUrl)) // we're allowed to panic here :)
}

type anySupportAdder interface {
	Add(x *any.Any)
}

type anySupportSetter interface {
	Set(x *any.Any)
}

type anySupportAdderSetter struct {
	anySupportSetter
}

func (s *anySupportAdderSetter) Add(x *any.Any) {
	s.Set(x)
}

type anySupportRemover interface {
	Remove(x *any.Any)
}

func anySupportRemove(r anySupportRemover, t *AnySupportType) {
	switch v := t.Value.(type) {
	case *AnySupportType_AnyValue:
		r.Remove(v.AnyValue)
	case *AnySupportType_StringValue:
		r.Remove(encoding.String(v.StringValue))
	case *AnySupportType_BytesValue:
		r.Remove(encoding.Bytes(v.BytesValue))
	case *AnySupportType_BoolValue:
		r.Remove(encoding.Bool(v.BoolValue))
	case *AnySupportType_DoubleValue:
		r.Remove(encoding.Float64(v.DoubleValue))
	case *AnySupportType_FloatValue:
		r.Remove(encoding.Float32(v.FloatValue))
	case *AnySupportType_Int32Value:
		r.Remove(encoding.Int32(v.Int32Value))
	case *AnySupportType_Int64Value:
		r.Remove(encoding.Int64(v.Int64Value))
	}
}

func anySupportAdd(a anySupportAdder, t *AnySupportType) {
	switch v := t.Value.(type) {
	case *AnySupportType_AnyValue:
		a.Add(v.AnyValue)
	case *AnySupportType_StringValue:
		a.Add(encoding.String(v.StringValue))
	case *AnySupportType_BytesValue:
		a.Add(encoding.Bytes(v.BytesValue))
	case *AnySupportType_BoolValue:
		a.Add(encoding.Bool(v.BoolValue))
	case *AnySupportType_DoubleValue:
		a.Add(encoding.Float64(v.DoubleValue))
	case *AnySupportType_FloatValue:
		a.Add(encoding.Float32(v.FloatValue))
	case *AnySupportType_Int32Value:
		a.Add(encoding.Int32(v.Int32Value))
	case *AnySupportType_Int64Value:
		a.Add(encoding.Int64(v.Int64Value))
	}
}

func anySupportSetClock(r *crdt.LWWRegister, t *AnySupportType, clock crdt.Clock, customValue int64) {
	switch v := t.Value.(type) {
	case *AnySupportType_AnyValue:
		r.SetWithClock(v.AnyValue, clock, customValue)
	case *AnySupportType_StringValue:
		r.SetWithClock(encoding.String(v.StringValue), clock, customValue)
	case *AnySupportType_BytesValue:
		r.SetWithClock(encoding.Bytes(v.BytesValue), clock, customValue)
	case *AnySupportType_BoolValue:
		r.SetWithClock(encoding.Bool(v.BoolValue), clock, customValue)
	case *AnySupportType_DoubleValue:
		r.SetWithClock(encoding.Float64(v.DoubleValue), clock, customValue)
	case *AnySupportType_FloatValue:
		r.SetWithClock(encoding.Float32(v.FloatValue), clock, customValue)
	case *AnySupportType_Int32Value:
		r.SetWithClock(encoding.Int32(v.Int32Value), clock, customValue)
	case *AnySupportType_Int64Value:
		r.SetWithClock(encoding.Int64(v.Int64Value), clock, customValue)
	}
}
