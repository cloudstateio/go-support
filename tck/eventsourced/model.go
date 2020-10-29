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

package eventsourced

import (
	"errors"
	"fmt"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
)

type TestModel struct {
	state string
}

func NewTestModel(id eventsourced.EntityID) eventsourced.EntityHandler {
	return &TestModel{}
}

func (m *TestModel) HandleCommand(ctx *eventsourced.Context, name string, cmd proto.Message) (reply proto.Message, err error) {
	switch c := cmd.(type) {
	case *Request:
		for _, action := range c.GetActions() {
			switch a := action.GetAction().(type) {
			case *RequestAction_Emit:
				ctx.Emit(&Persisted{Value: a.Emit.GetValue()})
			case *RequestAction_Forward:
				req, err := encoding.MarshalAny(&Request{Id: a.Forward.Id})
				if err != nil {
					return nil, err
				}
				ctx.Forward(&protocol.Forward{
					ServiceName: "cloudstate.tck.model.EventSourcedTwo",
					CommandName: "Call",
					Payload:     req,
				})
			case *RequestAction_Effect:
				req, err := encoding.MarshalAny(&Request{Id: a.Effect.Id})
				if err != nil {
					return nil, err
				}
				ctx.Effect(&protocol.SideEffect{
					ServiceName: "cloudstate.tck.model.EventSourcedTwo",
					CommandName: "Call",
					Payload:     req,
					Synchronous: a.Effect.Synchronous,
				})
			case *RequestAction_Fail:
				return nil, errors.New(a.Fail.GetMessage())
			}
		}
	}
	return &Response{Message: m.state}, nil
}

func (m *TestModel) HandleEvent(ctx *eventsourced.Context, event interface{}) error {
	switch c := event.(type) {
	case *Persisted:
		m.state += c.GetValue()
		return nil
	}
	return errors.New("event not handled")
}

func (m *TestModel) Snapshot(ctx *eventsourced.Context) (snapshot interface{}, err error) {
	return &Persisted{Value: m.state}, nil
}

func (m *TestModel) HandleSnapshot(ctx *eventsourced.Context, snapshot interface{}) error {
	switch s := snapshot.(type) {
	case *Persisted:
		m.state = s.GetValue()
		return nil
	}
	return errors.New("snapshot not handled")
}

type TestModelTwo struct {
}

func (m *TestModelTwo) HandleCommand(ctx *eventsourced.Context, name string, cmd proto.Message) (reply proto.Message, err error) {
	switch cmd.(type) {
	case *Request:
		return &Response{}, nil
	}
	return nil, fmt.Errorf("unhandled command: %q", name)
}

func (m *TestModelTwo) HandleEvent(ctx *eventsourced.Context, event interface{}) error {
	return nil
}

func NewTestModelTwo(id eventsourced.EntityID) eventsourced.EntityHandler {
	return &TestModelTwo{}
}
