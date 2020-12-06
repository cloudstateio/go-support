package value

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type Context struct {
	EntityID EntityID
	// Entity describes the instance that is used as an entity.
	Entity *Entity
	// Instance is the instance of the entity this context is for.
	Instance EntityHandler
	// ctx is the context.Context from the stream this context is assigned to.
	ctx context.Context

	update      bool
	delete      bool
	forward     *protocol.Forward
	failure     error
	sideEffects []*protocol.SideEffect
	state       *any.Any
}

func (c *Context) Forward(forward *protocol.Forward) {
	c.forward = forward
	c.failure = nil
}

func (c *Context) SideEffect(effect *protocol.SideEffect) {
	c.sideEffects = append(c.sideEffects, effect)
}

func (c *Context) entityReply(command *protocol.Command, reply *any.Any) *entity.ValueEntityReply {
	if c.failure != nil {
		return &entity.ValueEntityReply{
			CommandId: command.Id,
			// SideEffects: c.sideEffects, // TODO: as per TCK we don't return sideEffect, why?
			ClientAction: &protocol.ClientAction{
				Action: &protocol.ClientAction_Failure{
					Failure: &protocol.Failure{
						CommandId:   command.Id,
						Description: c.failure.Error(),
						// Restart:     len(c.sideEffects) > 0, TODO: how do we support restarts?
						Restart: false,
					},
				},
			},
		}
	}
	if c.forward != nil && c.update {
		return &entity.ValueEntityReply{
			CommandId: command.Id,
			ClientAction: &protocol.ClientAction{
				Action: &protocol.ClientAction_Forward{
					Forward: c.forward,
				},
			},
			SideEffects: c.sideEffects,
			StateAction: &entity.ValueEntityAction{
				Action: &entity.ValueEntityAction_Update{
					Update: &entity.ValueEntityUpdate{
						Value: c.state,
					},
				},
			},
		}
	}
	if c.forward != nil {
		return &entity.ValueEntityReply{
			CommandId: command.Id,
			ClientAction: &protocol.ClientAction{
				Action: &protocol.ClientAction_Forward{
					Forward: c.forward,
				},
			},
			SideEffects: c.sideEffects,
		}
	}
	if c.delete {
		return &entity.ValueEntityReply{
			CommandId: command.Id,
			ClientAction: &protocol.ClientAction{
				Action: &protocol.ClientAction_Reply{Reply: &protocol.Reply{
					Payload:  reply,
					Metadata: command.Metadata,
				}},
			},
			SideEffects: c.sideEffects,
			StateAction: &entity.ValueEntityAction{
				Action: &entity.ValueEntityAction_Delete{Delete: &entity.ValueEntityDelete{}},
			},
		}
	}
	if c.update {
		return &entity.ValueEntityReply{
			CommandId: command.Id,
			ClientAction: &protocol.ClientAction{
				Action: &protocol.ClientAction_Reply{Reply: &protocol.Reply{
					Payload:  reply,
					Metadata: command.Metadata,
				}},
			},
			SideEffects: c.sideEffects,
			StateAction: &entity.ValueEntityAction{
				Action: &entity.ValueEntityAction_Update{
					Update: &entity.ValueEntityUpdate{
						Value: c.state,
					},
				},
			},
		}
	}
	if reply != nil {
		return &entity.ValueEntityReply{
			CommandId: command.Id,
			ClientAction: &protocol.ClientAction{
				Action: &protocol.ClientAction_Reply{Reply: &protocol.Reply{
					Payload:  reply,
					Metadata: command.Metadata,
				}},
			},
			SideEffects: c.sideEffects,
		}
	}
	return nil
}

func (c *Context) runCommand(cmd *protocol.Command) (*any.Any, error) {
	// unmarshal the commands message
	msgName := strings.TrimPrefix(cmd.GetPayload().GetTypeUrl(), "type.googleapis.com/")
	messageType := proto.MessageType(msgName)
	message, ok := reflect.New(messageType.Elem()).Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("messageType is no proto.Message: %v", messageType)
	}
	if err := proto.Unmarshal(cmd.Payload.Value, message); err != nil {
		return nil, err
	}
	return c.Instance.HandleCommand(c, cmd.Name, message)
}

func (c *Context) Delete() {
	c.update = false
	c.delete = true
	c.state = nil
}

func (c *Context) Update(state *any.Any, err error) error {
	if err != nil {
		return err
	}
	c.update = true
	c.delete = false
	c.state = state
	return nil
}
