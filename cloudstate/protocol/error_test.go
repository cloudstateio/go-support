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

package protocol

import (
	"errors"
	"fmt"
	"testing"
)

var ErrTest1 = errors.New("test1 error")

func TestClientFailure_Error(t *testing.T) {
	t.Run("test protocol error", func(t *testing.T) {
		failure0 := ServerError{
			Failure: &Failure{CommandId: 0},
			Err:     fmt.Errorf("its an unusual error: %w", ErrTest1),
		}

		if is := errors.Is(&failure0, ErrTest1); !is {
			t.Error("errors.Is(ErrTest1, failure0)")
		}

		wrapped0 := fmt.Errorf("wrapped0: %w", failure0)
		f := &ServerError{}
		if !errors.As(wrapped0, f) {
			t.Error("!errors.As(wrapped0, f)")
		}

		wrapped1 := errors.Unwrap(f)
		wrapped2 := errors.Unwrap(wrapped1)
		wrapped3 := errors.Unwrap(wrapped2)
		fmt.Printf("f: %v\n", wrapped3)
		if wrapped3 != nil {
			t.Error("wrapped3 != nil")
		}
	})
}
