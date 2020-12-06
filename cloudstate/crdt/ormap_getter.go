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

	"github.com/golang/protobuf/ptypes/any"
)

func (m *ORMap) Flag(key *any.Any) (*Flag, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if flag, ok := v.Value.(*Flag); ok {
			return flag, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type Flag but: %+v", key, v)
	}
	return nil, nil
}

func (m *ORMap) GCounter(key *any.Any) (*GCounter, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if counter, ok := v.Value.(*GCounter); ok {
			return counter, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type GCounter but: %+v", key, v)
	}
	return nil, nil
}

func (m *ORMap) GSet(key *any.Any) (*GSet, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if set, ok := v.Value.(*GSet); ok {
			return set, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type GSet but: %+v", key, v)
	}
	return nil, nil
}

func (m *ORMap) LWWRegister(key *any.Any) (*LWWRegister, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if r, ok := v.Value.(*LWWRegister); ok {
			return r, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type LWWRegister but: %+v", key, v)
	}
	return nil, nil
}

func (m *ORMap) ORMap(key *any.Any) (*ORMap, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if m, ok := v.Value.(*ORMap); ok {
			return m, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type ORMap but: %+v", key, v)
	}
	return nil, nil
}

func (m *ORMap) ORSet(key *any.Any) (*ORSet, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if set, ok := v.Value.(*ORSet); ok {
			return set, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type ORSet but: %+v", key, v)
	}
	return nil, nil
}

func (m *ORMap) PNCounter(key *any.Any) (*PNCounter, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if c, ok := v.Value.(*PNCounter); ok {
			return c, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type PNCounter but: %+v", key, v)
	}
	return nil, nil
}

func (m *ORMap) Vote(key *any.Any) (*Vote, error) {
	if v, has := m.value[m.hashAny(key)]; has {
		if c, ok := v.Value.(*Vote); ok {
			return c, nil
		}
		return nil, fmt.Errorf("value at key: %v is not of type Vote but: %+v", key, v)
	}
	return nil, nil
}
