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
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/entity"
)

func TestVote(t *testing.T) {
	t.Run("should have zero voters, votes and no self vote when instantiated", func(t *testing.T) {
		v := NewVote()
		if v.VotesFor() != 0 {
			t.Fatalf("v.VotesFor(): %v; want: %v", v.VotesFor(), 0)
		}
		if v.Voters() != 1 {
			t.Fatalf("v.Voters(): %v; want: %v", v.Voters(), 1)
		}
		if v.SelfVote() != false {
			t.Fatalf("v.SelfVote(): %v; want: %v", v.SelfVote(), false)
		}
	})
	// t.Run("should reflect a state update", func(t *testing.T) {
	// 	v := NewVote()
	// 	err := v.applyState(encDecState(&entity.CrdtState{
	// 		State: &entity.CrdtState_Vote{
	// 			Vote: &entity.VoteState{
	// 				TotalVoters: 5,
	// 				VotesFor:    3,
	// 				SelfVote:    true,
	// 			}},
	// 	}))
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	if v.VotesFor() != 3 {
	// 		t.Fatalf("v.VotesFor(): %v; want: %v", v.VotesFor(), 3)
	// 	}
	// 	if v.Voters() != 5 {
	// 		t.Fatalf("v.Voters(): %v; want: %v", v.Voters(), 5)
	// 	}
	// 	if v.SelfVote() != true {
	// 		t.Fatalf("v.SelfVote(): %v; want: %v", v.SelfVote(), true)
	// 	}
	// })
	t.Run("should reflect a delta update", func(t *testing.T) {
		v := NewVote()
		if err := v.applyDelta(encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Vote{
				Vote: &entity.VoteDelta{
					TotalVoters: 5,
					VotesFor:    3,
					SelfVote:    false,
				}},
		})); err != nil {
			t.Fatal(err)
		}
		if err := v.applyDelta(encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Vote{
				Vote: &entity.VoteDelta{
					TotalVoters: 4,
					VotesFor:    2,
				}},
		})); err != nil {
			t.Fatal(err)
		}
		if v.VotesFor() != 2 {
			t.Fatalf("v.VotesFor(): %v; want: %v", v.VotesFor(), 2)
		}
		if v.Voters() != 4 {
			t.Fatalf("v.Voters(): %v; want: %v", v.Voters(), 4)
		}
		if v.SelfVote() != false {
			t.Fatalf("v.SelfVote(): %v; want: %v", v.SelfVote(), false)
		}
	})
	t.Run("should generate deltas", func(t *testing.T) {
		v := NewVote()
		v.Vote(true)
		delta := encDecDelta(v.Delta())
		v.resetDelta()
		if dv := delta.GetVote().GetSelfVote(); dv != true {
			t.Fatalf("delta.SelfVote(): %v; want: %v", dv, true)
		}
		if v.VotesFor() != 1 {
			t.Fatalf("v.VotesFor(): %v; want: %v", v.VotesFor(), 1)
		}
		if v.SelfVote() != true {
			t.Fatalf("v.SelfVote(): %v; want: %v", v.SelfVote(), true)
		}

		v.Vote(false)
		delta = encDecDelta(v.Delta())
		v.resetDelta()
		if dv := delta.GetVote().GetSelfVote(); dv != false {
			t.Fatalf("delta.SelfVote(): %v; want: %v", dv, false)
		}
		if v.VotesFor() != 0 {
			t.Fatalf("v.VotesFor(): %v; want: %v", v.VotesFor(), 0)
		}
		if v.SelfVote() != false {
			t.Fatalf("v.SelfVote(): %v; want: %v", v.SelfVote(), false)
		}
	})
	t.Run("should return its state", func(t *testing.T) {
		v := NewVote()
		v.resetDelta()
		if sv := v.SelfVote(); sv != false {
			t.Fatalf("state.GetSelfVote(): %v; want: %v", sv, false)
		}
		if vf := v.VotesFor(); vf != 0 {
			t.Fatalf("state.GetVotesFor(): %v; want: %v", vf, 0)
		}
		if tv := v.Voters(); tv != 1 {
			t.Fatalf("state.GetTotalVoters(): %v; want: %v", tv, 1)
		}

		v.Vote(true)
		v.resetDelta()
		if sv := v.SelfVote(); sv != true {
			t.Fatalf("state.GetSelfVote(): %v; want: %v", sv, true)
		}
		if vf := v.VotesFor(); vf != 1 {
			t.Fatalf("state.GetVotesFor(): %v; want: %v", vf, 1)
		}
		if tv := v.Voters(); tv != 1 {
			t.Fatalf("state.GetTotalVoters(): %v; want: %v", tv, 1)
		}
	})

	voteDelta := func(vs *entity.VoteDelta) *entity.CrdtDelta {
		return encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Vote{Vote: vs},
		})
	}
	t.Run("should correctly calculate a majority vote", func(t *testing.T) {
		var tests = []struct {
			vs  *entity.VoteDelta
			maj bool
		}{
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 3, SelfVote: true}, true},
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 2, SelfVote: true}, false},
			{&entity.VoteDelta{TotalVoters: 6, VotesFor: 3, SelfVote: true}, false},
			{&entity.VoteDelta{TotalVoters: 6, VotesFor: 4, SelfVote: true}, true},
			{&entity.VoteDelta{TotalVoters: 1, VotesFor: 0, SelfVote: false}, false},
			{&entity.VoteDelta{TotalVoters: 1, VotesFor: 1, SelfVote: true}, true},
		}
		v := NewVote()
		for _, test := range tests {
			if err := v.applyDelta(voteDelta(test.vs)); err != nil {
				t.Fatal(err)
			}
			if v.Majority() != test.maj {
				t.Fatalf("test: %+v, v.Majority(): %v; want: %v", test.vs, v.Majority(), test.maj)
			}
		}
	})
	t.Run("should correctly calculate an at least one vote", func(t *testing.T) {
		var tests = []struct {
			vs      *entity.VoteDelta
			atLeast bool
		}{
			{&entity.VoteDelta{TotalVoters: 1, VotesFor: 0, SelfVote: false}, false},
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 0, SelfVote: false}, false},
			{&entity.VoteDelta{TotalVoters: 1, VotesFor: 1, SelfVote: true}, true},
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 1, SelfVote: true}, true},
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 3, SelfVote: true}, true},
		}
		v := NewVote()
		for _, test := range tests {
			if err := v.applyDelta(voteDelta(test.vs)); err != nil {
				t.Fatal(err)
			}
			if v.AtLeastOne() != test.atLeast {
				t.Fatalf("test: %+v, v.AtLeastOne(): %v; want: %v", test.vs, v.AtLeastOne(), test.atLeast)
			}
		}
	})
	t.Run("should correctly calculate an all votes", func(t *testing.T) {
		var tests = []struct {
			vs  *entity.VoteDelta
			all bool
		}{
			{&entity.VoteDelta{TotalVoters: 1, VotesFor: 0, SelfVote: false}, false},
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 0, SelfVote: false}, false},
			{&entity.VoteDelta{TotalVoters: 1, VotesFor: 1, SelfVote: true}, true},
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 3, SelfVote: true}, false},
			{&entity.VoteDelta{TotalVoters: 5, VotesFor: 5, SelfVote: true}, true},
		}
		v := NewVote()
		for _, test := range tests {
			if err := v.applyDelta(voteDelta(test.vs)); err != nil {
				t.Fatal(err)
			}
			if v.All() != test.all {
				t.Fatalf("test: %+v, v.All(): %v; want: %v", test.vs, v.All(), test.all)
			}
		}
	})
}
