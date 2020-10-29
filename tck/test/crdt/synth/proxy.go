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
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
)

const recvTimeout = 1

type proxy struct {
	h       entity.Crdt_HandleClient
	t       *testing.T
	seq     int64
	recvC   chan resp
	cancelC chan bool
}

func newProxy(ctx context.Context, s *server) *proxy {
	s.t.Helper()
	h, err := entity.NewCrdtClient(s.conn).Handle(ctx)
	if err != nil {
		s.t.Fatal(err)
	}
	p := &proxy{
		t:       s.t,
		h:       h,
		seq:     1,
		recvC:   make(chan resp),
		cancelC: make(chan bool),
	}
	go func() {
		for {
			recv, err := p.h.Recv()
			select {
			case p.recvC <- resp{recv, err}:
			case <-p.cancelC:
				return
			}
		}
	}()
	return p
}

func (p *proxy) Send(in *entity.CrdtStreamIn) error {
	return p.h.Send(in)
}

func (p *proxy) teardown() {
	close(p.cancelC)
}

type command struct {
	c *protocol.Command
	m proto.Message
}

type state struct {
	s *entity.CrdtState
}

type delta struct {
	d *entity.CrdtDelta
}

type delete struct {
	d *entity.CrdtDelete
}

type resp struct {
	out *entity.CrdtStreamOut
	err error
}

func (p *proxy) Recv() (*entity.CrdtStreamOut, error) {
	select {
	case m := <-p.recvC:
		return m.out, m.err
	case <-time.After(recvTimeout * time.Second):
		p.t.Log("no response")
		return nil, nil
	}
}

func (p *proxy) init(i *entity.CrdtInit) {
	err := p.h.Send(&entity.CrdtStreamIn{
		Message: &entity.CrdtStreamIn_Init{Init: i},
	})
	if err != nil {
		p.t.Fatal(err)
	}
}

func (p *proxy) state(m proto.Message) {
	switch s := m.(type) {
	case *entity.PNCounterState:
		p.sendState(state{
			&entity.CrdtState{State: &entity.CrdtState_Pncounter{
				Pncounter: s,
			}},
		})
	case *entity.GCounterState:
		p.sendState(state{
			&entity.CrdtState{State: &entity.CrdtState_Gcounter{
				Gcounter: s,
			}},
		})
	case *entity.FlagState:
		p.sendState(state{
			&entity.CrdtState{State: &entity.CrdtState_Flag{
				Flag: s,
			}},
		})
	case *entity.VoteState:
		p.sendState(state{
			&entity.CrdtState{State: &entity.CrdtState_Vote{
				Vote: s,
			}},
		})
	case *entity.ORMapState:
		p.sendState(state{
			&entity.CrdtState{State: &entity.CrdtState_Ormap{
				Ormap: s,
			}},
		})
	default:
		p.t.Fatal("state type not found")
	}
}

func (p *proxy) sendState(st state) {
	if err := p.h.Send(stateMsg(st.s)); err != nil {
		p.t.Fatal(err)
	}
}

func (p *proxy) sendRecvState(st state) (*entity.CrdtStreamOut, error) {
	if err := p.h.Send(stateMsg(st.s)); err != nil {
		return nil, err
	}
	return p.Recv()
}

func (p *proxy) delta(m proto.Message) {
	switch d := m.(type) {
	case *entity.PNCounterDelta:
		p.sendDelta(delta{
			&entity.CrdtDelta{Delta: &entity.CrdtDelta_Pncounter{
				Pncounter: d,
			}}},
		)
	case *entity.GCounterDelta:
		p.sendDelta(delta{
			d: &entity.CrdtDelta{
				Delta: &entity.CrdtDelta_Gcounter{
					Gcounter: d,
				}}},
		)
	case *entity.FlagDelta:
		p.sendDelta(delta{
			d: &entity.CrdtDelta{
				Delta: &entity.CrdtDelta_Flag{
					Flag: d,
				}}},
		)
	case *entity.VoteDelta:
		p.sendDelta(delta{
			d: &entity.CrdtDelta{
				Delta: &entity.CrdtDelta_Vote{
					Vote: d,
				}}},
		)
	default:
		p.t.Fatal("state type not found")
	}
}

func (p *proxy) sendDelta(d delta) {
	p.t.Helper()
	if err := p.h.Send(deltaMsg(d.d)); err != nil {
		p.t.Fatal(err)
	}
}

func (p *proxy) sendDelete(d delete) {
	p.t.Helper()
	if err := p.h.Send(deleteMsg(d.d)); err != nil {
		p.t.Fatal(err)
	}
}

func (p *proxy) sendRecvDelta(d delta) (*entity.CrdtStreamOut, error) {
	if err := p.h.Send(deltaMsg(d.d)); err != nil {
		return nil, err
	}
	return p.Recv()
}

func (p *proxy) command(entityID string, name string, m proto.Message) *entity.CrdtStreamOut {
	return p.sendCmdRecvReply(command{
		&protocol.Command{EntityId: entityID, Name: name},
		m,
	})
}

func (p *proxy) commandStreamed(entityID string, name string, m proto.Message) *entity.CrdtStreamOut {
	return p.sendCmdRecvReply(command{
		&protocol.Command{EntityId: entityID, Name: name, Streamed: true},
		m,
	})
}

func (p *proxy) sendCmdRecvReply(cmd command) *entity.CrdtStreamOut {
	p.t.Helper()
	if cmd.c.Id == 0 {
		cmd.c.Id = p.seq
		defer func() { p.seq++ }()
	}
	any, err := encoding.MarshalAny(cmd.m)
	if err != nil {
		p.t.Fatal(err)
	}
	cmd.c.Payload = any
	err = p.h.Send(commandMsg(cmd.c))
	if err != nil {
		p.t.Fatal(err)
	}
	msg, err := p.Recv()
	if err != nil {
		p.t.Fatal(err)
	}
	if msg == nil {
		p.t.Fatal("nothing received")
	}
	switch msg.Message.(type) {
	case *entity.CrdtStreamOut_Failure:
	default:
		p.checkCommandID(msg)
	}
	return msg
}

func (p *proxy) sendCmdRecvErr(cmd command) (*entity.CrdtStreamOut, error) {
	p.t.Helper()
	if cmd.c.Id == 0 {
		cmd.c.Id = p.seq
		defer func() { p.seq++ }()
	}
	any, err := encoding.MarshalAny(cmd.m)
	if err != nil {
		return nil, err
	}
	cmd.c.Payload = any
	err = p.h.Send(commandMsg(cmd.c))
	if err != nil {
		return nil, err
	}
	recv, err := p.Recv()
	if err != nil {
		return nil, err
	}
	if recv == nil {
		p.t.Fatal("nothing received")
	}
	switch recv.Message.(type) {
	case *entity.CrdtStreamOut_Failure:
	default:
		p.checkCommandID(recv)
	}
	return recv, nil
}

func commandMsg(c *protocol.Command) *entity.CrdtStreamIn {
	return &entity.CrdtStreamIn{
		Message: &entity.CrdtStreamIn_Command{
			Command: c,
		},
	}
}

func stateMsg(s *entity.CrdtState) *entity.CrdtStreamIn {
	return &entity.CrdtStreamIn{
		Message: &entity.CrdtStreamIn_State{
			State: s,
		},
	}
}

func deltaMsg(d *entity.CrdtDelta) *entity.CrdtStreamIn {
	return &entity.CrdtStreamIn{
		Message: &entity.CrdtStreamIn_Changed{
			Changed: d,
		},
	}
}

func deleteMsg(d *entity.CrdtDelete) *entity.CrdtStreamIn {
	return &entity.CrdtStreamIn{
		Message: &entity.CrdtStreamIn_Deleted{
			Deleted: d,
		},
	}
}

func (p *proxy) checkCommandID(msg interface{}) {
	p.t.Helper()
	switch m := msg.(type) {
	case *entity.CrdtStreamOut:
		switch out := m.Message.(type) {
		case *entity.CrdtStreamOut_Reply:
			if got, want := out.Reply.CommandId, p.seq; got != want {
				p.t.Fatalf("command = %v; wanted: %d, for message:%+v", got, want, out)
			}
		case *entity.CrdtStreamOut_Failure:
			if got, want := out.Failure.CommandId, p.seq; got != want {
				p.t.Fatalf("command = %v; wanted: %d, for message:%+v", got, want, out)
			}
		case *entity.CrdtStreamOut_StreamedMessage:
			if got, want := out.StreamedMessage.CommandId, p.seq; got != want {
				p.t.Fatalf("command = %v; wanted: %d, for message:%+v", got, want, out)
			}
		case *entity.CrdtStreamOut_StreamCancelledResponse:
			if got, want := out.StreamCancelledResponse.CommandId, p.seq; got != want {
				p.t.Fatalf("command = %v; wanted: %d, for message:%+v", got, want, out)
			}
		default:
			p.t.Fatalf("unexpected message: %+v", m)
		}
	default:
		p.t.Fatalf("unexpected message: %+v", m)
	}
}
