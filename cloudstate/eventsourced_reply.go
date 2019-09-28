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

package cloudstate

import (
	"errors"
	"fmt"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
)

var ErrSendFailure = errors.New("unable to send a failure message")
var ErrSend = errors.New("unable to send a message")
var ErrMarshal = errors.New("unable to marshal a message")

var ErrFailure = errors.New("cloudstate failure")
var ErrClientActionFailure = errors.New("cloudstate client action failure")

func NewFailureError(format string, a ...interface{}) error {
	if len(a) != 0 {
		errorf := fmt.Errorf(fmt.Sprintf(format, a)+". %w", ErrFailure)
		return errorf
	} else {
		errorf := fmt.Errorf(format+". %w", ErrFailure)
		return errorf
	}
}

func NewClientActionFailureError(format string, a ...interface{}) error {
	errorf := fmt.Errorf(fmt.Sprintf(format, a)+". %w", ErrClientActionFailure)
	return errorf
}

type ProtocolFailure struct {
	protocol.Failure
	err error
}

func (f ProtocolFailure) Error() string {
	return f.err.Error()
}

func (f ProtocolFailure) Unwrap() error {
	return f.err
}

func NewProtocolFailure(failure protocol.Failure) error {
	return ProtocolFailure{
		Failure: failure,
		err:     ErrFailure,
	}
}

// handleFailure checks if a CloudState failure or client action failure should
// be sent to the proxy, otherwise handleFailure returns the original failure
func handleFailure(failure error, server protocol.EventSourced_HandleServer, cmdId int64) error {
	if errors.Is(failure, ErrFailure) {
		// TCK says: Failure was not received, or not well-formed: Failure(Failure(0,cloudstate failure)) was not reply (CloudStateTCK.scala:339)
		// FIXME: why not getting the failure from the ProtocolFailure
		//return sendFailure(&protocol.Failure{Description: failure.Error()}, server)
		return sendClientActionFailure(&protocol.Failure{
			CommandId:   cmdId,
			Description: failure.Error(),
		}, server)
	}
	if errors.Is(failure, ErrClientActionFailure) {
		return sendClientActionFailure(&protocol.Failure{
			CommandId:   cmdId,
			Description: failure.Error(),
		}, server)
	}
	return failure
}

func sendEventSourcedReply(reply *protocol.EventSourcedReply, server protocol.EventSourced_HandleServer) error {
	err := server.Send(&protocol.EventSourcedStreamOut{
		Message: &protocol.EventSourcedStreamOut_Reply{
			Reply: reply,
		},
	})
	if err != nil {
		return fmt.Errorf("%s, %w", err, ErrSend)
	}
	return err
}

func sendFailure(failure *protocol.Failure, server protocol.EventSourced_HandleServer) error {
	err := server.Send(&protocol.EventSourcedStreamOut{
		Message: &protocol.EventSourcedStreamOut_Failure{
			Failure: failure,
		},
	})
	if err != nil {
		err = fmt.Errorf("%s, %w", err, ErrSendFailure)
	}
	return err
}

func sendClientActionFailure(failure *protocol.Failure, server protocol.EventSourced_HandleServer) error {
	err := server.Send(&protocol.EventSourcedStreamOut{
		Message: &protocol.EventSourcedStreamOut_Reply{
			Reply: &protocol.EventSourcedReply{
				CommandId: failure.CommandId,
				ClientAction: &protocol.ClientAction{
					Action: &protocol.ClientAction_Failure{
						Failure: failure,
					},
				},
			},
		},
	})
	if err != nil {
		err = fmt.Errorf("%s, %w", err, ErrSendFailure)
	}
	return err
}
