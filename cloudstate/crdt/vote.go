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
	"fmt"

	"github.com/cloudstateio/go-support/cloudstate/entity"
)

// A Vote is a CRDT which allows nodes to vote on a condition. It’s similar
// to a GCounter, each node has its own counter, and an odd value is considered
// a vote for the condition, while an even value is considered a vote against.
// The result of the vote is decided by taking the votes of all nodes that are
// currently members of the cluster (when a node leave, its vote is discarded).
// Multiple decision strategies can be used to decide the result of the vote,
// such as at least one, majority and all.
type Vote struct {
	selfVote        bool
	selfVoteChanged bool // delta seen
	voters          uint32
	votesFor        uint32
}

var _ CRDT = (*Vote)(nil)

func NewVote() *Vote {
	return &Vote{
		selfVote:        false,
		selfVoteChanged: false,
		voters:          1,
		votesFor:        0,
	}
}

// SelfVote is the vote of the current node,
// which is included in Voters and VotesFor.
func (v *Vote) SelfVote() bool {
	return v.selfVote
}

// Voters is the total number of voters.
func (v *Vote) Voters() uint32 {
	return v.voters
}

// VotesFor is the number of votes for.
func (v *Vote) VotesFor() uint32 {
	return v.votesFor
}

// AtLeastOne returns true if there is at least one voter for the condition.
func (v *Vote) AtLeastOne() bool {
	return v.votesFor > 0
}

// Majority returns true if the number of votes for is more than half the number of voters.
func (v *Vote) Majority() bool {
	return v.votesFor > v.voters/2
}

// All returns true if the number of votes for equals the number of voters.
func (v *Vote) All() bool {
	return v.votesFor == v.voters
}

// Vote votes with the given boolean for a condition.
func (v *Vote) Vote(vote bool) {
	if v.selfVote == vote {
		return
	}
	v.selfVoteChanged = !v.selfVoteChanged
	v.selfVote = vote
	if v.selfVote {
		v.votesFor += 1
	} else {
		v.votesFor -= 1
	}
}

func (v *Vote) HasDelta() bool {
	return v.selfVoteChanged
}

func (v *Vote) Delta() *entity.CrdtDelta {
	if !v.selfVoteChanged {
		return nil
	}
	return &entity.CrdtDelta{
		Delta: &entity.CrdtDelta_Vote{Vote: &entity.VoteDelta{
			SelfVote:    v.selfVote,
			VotesFor:    int32(v.votesFor), // TODO, we never overflow, yes?
			TotalVoters: int32(v.voters),
		}},
	}
}

func (v *Vote) resetDelta() {
	v.selfVoteChanged = false
}

func (v *Vote) applyDelta(delta *entity.CrdtDelta) error {
	d := delta.GetVote()
	if d == nil {
		return fmt.Errorf("unable to apply delta %+v to the Vote", delta)
	}
	v.selfVote = d.SelfVote
	v.voters = uint32(d.TotalVoters)
	v.votesFor = uint32(d.VotesFor)
	return nil
}
