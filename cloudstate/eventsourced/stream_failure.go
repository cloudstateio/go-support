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

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
)

// sendProtocolFailure sends a given error to the proxy. If the error is a protocol.ServerError a corresponding
// commandId is unwrapped and added to the failure. Any other failure is sent as a protocol failure.
//
// we send protocol.ClientError as a protocol.ClientAction_Failure for everything the user function would like to inform the client about.
// we send protocol.ServerError as a protocol.Failure for everything else that is not originated by the user function.
// whenever possible, the command id is set.
//
// failure semantics are defined here:
// - https://github.com/cloudstateio/cloudstate/issues/375#issuecomment-672336020
// - https://github.com/cloudstateio/cloudstate/pull/119#discussion_r375619440
// - https://github.com/cloudstateio/cloudstate/pull/392
// - https://github.com/cloudstateio/cloudstate/issues/375#issuecomment-671108797
//
// Any error coming not from context.fail, closes the stream, independently if it's a protocol error or an entity error.
func sendProtocolFailure(e error, s entity.EventSourced_HandleServer) error {
	var commandID int64 = 0
	var desc = e.Error()
	var se protocol.ServerError
	if errors.As(e, &se) {
		commandID = se.Failure.CommandId
		desc = se.Failure.Description
		if desc == "" {
			if u := errors.Unwrap(e); u != nil {
				desc = u.Error()
			}
		}
	}
	err := s.Send(&entity.EventSourcedStreamOut{
		Message: &entity.EventSourcedStreamOut_Failure{
			Failure: &protocol.Failure{
				CommandId:   commandID,
				Description: desc,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("send of EventSourcedStreamOut.Failure failed with: %w", err)
	}
	return nil
}
