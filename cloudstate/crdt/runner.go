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

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
)

// runner runs a stream with the help of a context.
type runner struct {
	stream  entity.Crdt_HandleServer
	context *Context
}

// handleDelta handles an incoming delta message to be applied to the current state.
// A delta to be applied to the current value. It may be sent at any time as long
// as the user function already has value.
func (r *runner) handleDelta(delta *entity.CrdtDelta) error {
	if r.context.crdt == nil {
		s, err := newFor(delta)
		if err != nil {
			return err
		}
		r.context.crdt = s
	}
	return r.context.crdt.applyDelta(delta)
}

// handleCancellation handles an incoming cancellation message to be applied to
// the current state. A stream has been cancelled. A command handler may also
// register an onCancel callback to be notified when the stream is cancelled.
// The cancellation callback handler may update the crdt. This is useful if
// the crdt is being used to track connections, for example, when using Vote
// CRDTs to track a users online status.
func (r *runner) handleCancellation(cancelled *protocol.StreamCancelled) error {
	id := CommandID(cancelled.GetId())
	ctx := r.context.streamedCtx[id]
	// The cancelled stream is not allowed to handle changes, so we remove it.
	delete(r.context.streamedCtx, id)
	if ctx.cancel == nil {
		return r.sendCancelledMessage(&entity.CrdtStreamCancelledResponse{
			CommandId: id.Value(),
		})
	}
	// Notify the user about the cancellation.
	if err := ctx.cancelled(); err != nil {
		return err
	}
	stateAction := ctx.stateAction()
	err := r.sendCancelledMessage(&entity.CrdtStreamCancelledResponse{
		CommandId:   id.Value(),
		StateAction: stateAction,
		SideEffects: ctx.sideEffects,
	})
	if err != nil {
		return err
	}
	// The state has been changed therefore streamed change handlers get
	// informed.
	if stateAction != nil {
		return r.handleChange()
	}
	return nil
}

// handleCommand handles the received command.
// Cloudstate CRDTs support handling server streamed calls, that is, when the
// gRPC service call for a CRDT marks the return type as streamed. When a user
// function receives a streamed message, it is allowed to update the CRDT, on
// two occasions - when the call is first received, and when the client cancels
// the stream. If it wishes to make updates at other times, it can do so by
// emitting effects with the streamed messages that it sends down the stream.
// A user function can send a message down a stream in response to anything,
// however the Cloudstate supplied support libraries only allow sending messages
// in response to the CRDT changing. In this way, use cases that require monitoring
// the state of a CRDT can be implemented.
func (r *runner) handleCommand(cmd *protocol.Command) (streamError error) {
	if r.context.EntityID != EntityID(cmd.EntityId) {
		return fmt.Errorf("the command entity id: %s does not match the initialized entity id: %s", cmd.EntityId, r.context.EntityID)
	}
	ctx := r.context.commandContextFor(cmd)
	reply, err := ctx.runCommand(cmd)
	if err != nil && !errors.Is(err, protocol.ClientError{}) {
		return err
	}
	// TODO: error handling has to be clarified.
	// 	It seems, CRDT streams stopped for any error, even client failures.
	//	see: https://github.com/cloudstateio/cloudstate/pull/392
	if err != nil {
		ctx.fail(err)
	}
	// If the user function has failed, a client action failure will be sent.
	if ctx.failed != nil {
		reply = nil
	} else if err != nil {
		// On any other error, we return and close the stream.
		return err
	}
	clientAction, err := ctx.clientActionFor(reply)
	if err != nil {
		return err
	}
	if clientAction.GetFailure() != nil {
		ctx.failed = nil
		return r.sendCrdtReply(&entity.CrdtReply{
			CommandId:    ctx.CommandID.Value(),
			ClientAction: clientAction, // this is a ClientAction_Failure
		})
	}
	stateAction := ctx.stateAction()
	err = r.sendCrdtReply(&entity.CrdtReply{
		CommandId:    ctx.CommandID.Value(),
		ClientAction: clientAction,
		SideEffects:  ctx.sideEffects,
		StateAction:  stateAction,
		// TODO: Does a user choose to have a command streamed?
		//  Marking the reply streamed while having a stream flag on the command seems superfluous.
		//  According to the JVM implementation this means "stream accepted" and therefore shows the proxy if
		//  there are commands handling a streamed method.
		// TODO(spec): feedback on spec about this.
		Streamed: ctx.Streamed(),
	})
	if err != nil {
		return err
	}
	ctx.clearSideEffect()
	if stateAction != nil {
		if err := r.handleChange(); err != nil {
			return err
		}
	}
	if ctx.Streamed() {
		ctx.trackChanges()
	}
	return nil
}

func (r *runner) handleChange() error {
	for _, ctx := range r.context.streamedCtx {
		if ctx.change == nil {
			continue
		}
		reply, err := ctx.changed()

		// TODO: we have to clarify error path from here on.
		if errors.Is(err, ErrCtxFailCalled) {
			// ctx.clientActionFor will report a failure for that.
			reply = nil
		} else if err != nil {
			return err
		}
		clientAction, err := ctx.clientActionFor(reply)
		if err != nil {
			return err
		}
		if ctx.failed != nil {
			delete(ctx.streamedCtx, ctx.CommandID)
			if err := r.sendStreamedMessage(&entity.CrdtStreamedMessage{
				CommandId:    ctx.CommandID.Value(),
				ClientAction: clientAction,
			}); err != nil {
				return err
			}
			continue
		}
		if clientAction != nil || ctx.ended || len(ctx.sideEffects) > 0 {
			if ctx.ended {
				delete(ctx.streamedCtx, ctx.CommandID)
			}
			msg := &entity.CrdtStreamedMessage{
				CommandId:    ctx.CommandID.Value(),
				ClientAction: clientAction,
				SideEffects:  ctx.sideEffects,
				EndStream:    ctx.ended,
			}
			if err := r.sendStreamedMessage(msg); err != nil {
				return err
			}
			ctx.clearSideEffect()
			continue
		}
	}
	return nil
}

func (r *runner) sendStreamedMessage(msg *entity.CrdtStreamedMessage) error {
	return r.stream.Send(&entity.CrdtStreamOut{
		Message: &entity.CrdtStreamOut_StreamedMessage{
			StreamedMessage: msg,
		},
	})
}

func (r *runner) sendCancelledMessage(msg *entity.CrdtStreamCancelledResponse) error {
	return r.stream.Send(&entity.CrdtStreamOut{
		Message: &entity.CrdtStreamOut_StreamCancelledResponse{
			StreamCancelledResponse: msg,
		},
	})
}

func (r *runner) sendCrdtReply(reply *entity.CrdtReply) error {
	return r.stream.Send(&entity.CrdtStreamOut{
		Message: &entity.CrdtStreamOut_Reply{
			Reply: reply,
		},
	})
}
