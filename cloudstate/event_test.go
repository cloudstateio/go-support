//
//  Copyright 2019 Lightbend Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package cloudstate

import (
	"fmt"
	"sync/atomic"
	"testing"
)

type AnEntity struct {
	EventEmitter
}

func TestMultipleSubscribers(t *testing.T) {
	e := AnEntity{EventEmitter: NewEmitter()}
	n := int64(0)
	e.Subscribe(&Subscription{
		OnNext: func(event interface{}) error {
			atomic.AddInt64(&n, 1)
			return nil
		},
	})
	e.Emit(1)
	if n != 1 {
		t.Fail()
	}
	e.Emit(1)
	if n != 2 {
		t.Fail()
	}
	e.Subscribe(&Subscription{
		OnNext: func(event interface{}) error {
			atomic.AddInt64(&n, 1)
			return nil
		},
	})
	e.Emit(1)
	if n != 4 {
		t.Fail()
	}
}

func TestEventEmitter(t *testing.T) {
	e := AnEntity{EventEmitter: NewEmitter()}
	s := make([]string, 0)
	ee := fmt.Errorf("int types are no supported")
	e.Subscribe(&Subscription{
		OnNext: func(event interface{}) error {
			switch v := event.(type) {
			case int:
				return ee
			case string:
				s = append(s, v)
			}
			return nil
		},
		OnErr: func(err error) {
			if err != ee {
				t.Errorf("received unexpected error: %v", err)
			}
		},
	})
	// emit something that triggers an error we'd expect to happen
	e.Emit(1)
	// emit a string that gets to the list
	e.Emit("john")
	if len(s) != 1 || s[0] != "john" {
		t.Errorf("john was not in the list: %+v", s)
	}
}
