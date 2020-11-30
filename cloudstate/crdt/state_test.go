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
	"reflect"
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/entity"
)

func Test_newFor(t *testing.T) {
	type args struct {
		state *entity.CrdtDelta
	}
	tests := []struct {
		name string
		args args
		want CRDT
	}{
		{"newForFlag", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Flag{}}}, NewFlag()},
		{"newForGCounter", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Gcounter{}}}, NewGCounter()},
		{"newForGset", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Gset{}}}, NewGSet()},
		{"newForLWWRegister", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Lwwregister{}}}, NewLWWRegister(nil)},
		{"newForORMap", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Ormap{}}}, NewORMap()},
		{"newForORSet", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Orset{}}}, NewORSet()},
		{"newForPNCounter", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Pncounter{}}}, NewPNCounter()},
		{"newForVote", args{state: &entity.CrdtDelta{Delta: &entity.CrdtDelta_Vote{}}}, NewVote()},
		{"newForNil", args{state: &entity.CrdtDelta{Delta: nil}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := newFor(tt.args.state)
			if err != nil {
				if tt.want == nil {
					return
				}
				t.Fatal(err)
			}
			if got := state; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFor() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
