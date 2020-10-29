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

type (
	ServiceName string
	EntityID    string
	CommandID   int64
)

func (i EntityID) String() string {
	return string(i)
}

func (sn ServiceName) String() string {
	return string(sn)
}

func (id CommandID) Value() int64 {
	return int64(id)
}
