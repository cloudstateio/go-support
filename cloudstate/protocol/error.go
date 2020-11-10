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

type ClientError struct {
	Err error
}

func (e ClientError) Is(err error) bool {
	_, ok := err.(ClientError)
	return ok
}

func (e ClientError) Error() string {
	return e.Err.Error()
}

func (e ClientError) Unwrap() error {
	return e.Err
}

type ServerError struct {
	Failure *Failure
	Err     error
}

func (e ServerError) Is(err error) bool {
	_, ok := err.(ServerError)
	return ok
}

func (e ServerError) Error() string {
	return e.Err.Error()
}

func (e ServerError) Unwrap() error {
	return e.Err
}
