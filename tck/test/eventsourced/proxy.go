package eventsourced

import (
	"context"
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
)

type command struct {
	c *protocol.Command
	m proto.Message
}

type proxy struct {
	h   entity.EventSourced_HandleClient
	t   *testing.T
	seq int64
}

func newProxy(ctx context.Context, s *server) *proxy {
	s.t.Helper()
	h, err := entity.NewEventSourcedClient(s.conn).Handle(ctx)
	if err != nil {
		s.t.Fatal(err)
	}
	return &proxy{t: s.t, h: h, seq: 1}
}

func (p *proxy) checkCommandID(m interface{}) {
	p.t.Helper()
	switch m := m.(type) {
	case *entity.EventSourcedStreamOut_Reply:
		if got, want := m.Reply.CommandId, p.seq; got != want {
			p.t.Fatalf("got command id: %d; want: %d", got, want)
		}
	case *entity.EventSourcedStreamOut_Failure:
		if got, want := m.Failure.CommandId, p.seq; got != want {
			p.t.Fatalf("got command id: %d; want: %d", got, want)
		}
	default:
		p.t.Fatalf("unexpected message: %+v", m)
	}
}

func (p *proxy) sendRecvCmd(cmd command) *entity.EventSourcedStreamOut {
	p.t.Helper()
	if cmd.c.Id == 0 {
		p.seq++
		cmd.c.Id = p.seq
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
	recv, err := p.h.Recv()
	if err != nil {
		p.t.Fatal(err)
	}
	return recv
}

func (p *proxy) sendRecvCmdErr(cmd command) (*entity.EventSourcedStreamOut, error) {
	p.t.Helper()
	if cmd.c.Id == 0 {
		p.seq++
		cmd.c.Id = p.seq
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
	return p.h.Recv()
}

func (p *proxy) sendInit(init *entity.EventSourcedInit) {
	p.t.Helper()
	if err := p.h.Send(initMsg(init)); err != nil {
		p.t.Fatal(err)
	}
}

func (p *proxy) sendEvent(e *entity.EventSourcedEvent) {
	err := p.h.Send(eventMsg(e))
	if err != nil {
		p.t.Fatal(err)
	}
}

func eventMsg(e *entity.EventSourcedEvent) *entity.EventSourcedStreamIn {
	return &entity.EventSourcedStreamIn{
		Message: &entity.EventSourcedStreamIn_Event{
			Event: e,
		},
	}
}

func initMsg(i *entity.EventSourcedInit) *entity.EventSourcedStreamIn {
	return &entity.EventSourcedStreamIn{
		Message: &entity.EventSourcedStreamIn_Init{
			Init: i,
		},
	}
}

func commandMsg(c *protocol.Command) *entity.EventSourcedStreamIn {
	return &entity.EventSourcedStreamIn{
		Message: &entity.EventSourcedStreamIn_Command{
			Command: c,
		},
	}
}
