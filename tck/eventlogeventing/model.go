package eventlogeventing

import (
	"errors"
	"fmt"

	"github.com/cloudstateio/go-support/cloudstate/action"
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
)

type EventLogSubscriberModel struct {
}

func (e *EventLogSubscriberModel) HandleCommand(ctx *action.Context, name string, msg proto.Message) error {
	var id = ""
	for _, entry := range ctx.Metadata().GetEntries() {
		if entry.GetKey() == "ce-subject" {
			id = entry.GetStringValue()
			break
		}
	}
	switch name {
	case "Effect":
		r, err := encoding.MarshalAny(&Response{
			Id:      id,
			Message: msg.(*EffectRequest).GetMessage(),
		})
		if err != nil {
			return err
		}
		ctx.RespondWith(r)
	case "ProcessAnyEvent":
		var event JsonEvent
		if err := encoding.UnmarshalJSON(msg.(*any.Any), &event); err != nil {
			return err
		}
		r, err := encoding.MarshalAny(&Response{
			Id:      id,
			Message: event.Message,
		})
		if err != nil {
			return err
		}
		ctx.RespondWith(r)
	case "ProcessEventOne":
		one := msg.(*EventOne)
		if err := convert(ctx, ctx.Metadata(), one.GetStep()); err != nil {
			return err
		}
	case "ProcessEventTwo":
		two := msg.(*EventTwo)
		for _, step := range two.GetStep() {
			err := convert(ctx, ctx.Metadata(), step)
			if err != nil {
				return err
			}
			if err = ctx.Respond(nil); err != nil {
				return err
			}
		}
		ctx.Cancel()
	}
	return nil
}

func convert(ctx *action.Context, metadata *protocol.Metadata, step *ProcessStep) error {
	var id = ""
	for _, entry := range metadata.GetEntries() {
		if entry.GetKey() == "ce-subject" {
			id = entry.GetStringValue()
		}
	}
	if step.GetReply() != nil {
		x, err := encoding.MarshalAny(&Response{
			Id:      id,
			Message: step.GetReply().GetMessage(),
		})
		if err != nil {
			return err
		}
		ctx.RespondWith(x)
		return nil
	}
	if step.GetForward() != nil {
		payload, err := encoding.MarshalAny(&EffectRequest{
			Id:      id,
			Message: step.GetForward().GetMessage(),
		})
		if err != nil {
			return err
		}
		ctx.Forward(&protocol.Forward{
			ServiceName: "cloudstate.tck.model.eventlogeventing.EventLogSubscriberModel",
			CommandName: "Effect",
			Payload:     payload,
			Metadata:    ctx.Metadata(),
		})
		return nil
	}
	return errors.New("No reply or forward")
}

type EventSourcedEntityOne struct{}

func (EventSourcedEntityOne) HandleCommand(ctx *eventsourced.Context, name string, cmd proto.Message) (reply proto.Message, err error) {
	switch c := cmd.(type) {
	case *EmitEventRequest:
		switch e := c.GetEvent().(type) {
		case *EmitEventRequest_EventOne:
			ctx.Emit(e.EventOne)
		case *EmitEventRequest_EventTwo:
			ctx.Emit(e.EventTwo)
		}
	default:
		return nil, fmt.Errorf("unkown command: %q", c)
	}
	return &empty.Empty{}, nil
}

func (EventSourcedEntityOne) HandleEvent(ctx *eventsourced.Context, event interface{}) error {
	return nil
}

type EventSourcedEntityTwo struct{}

type JsonMessage struct {
	Message string `json:"message"`
}

func (EventSourcedEntityTwo) HandleCommand(ctx *eventsourced.Context, name string, cmd proto.Message) (reply proto.Message, err error) {
	switch name {
	case "EmitJsonEvent":
		json, err := encoding.JSON(JsonMessage{
			Message: cmd.(*JsonEvent).GetMessage()},
		)
		if err != nil {
			return nil, err
		}
		ctx.Emit(json)
	default:
		return nil, fmt.Errorf("unkown command: %q", name)
	}
	return &empty.Empty{}, nil
}

func (EventSourcedEntityTwo) HandleEvent(ctx *eventsourced.Context, event interface{}) error {
	return nil
}
