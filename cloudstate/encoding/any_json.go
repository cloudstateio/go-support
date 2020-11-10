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

package encoding

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	JSONTypeURLPrefix = "json.cloudstate.io"
)

func JSON(value interface{}) (*any.Any, error) {
	return MarshalJSON(value)
}

// MarshalJSON encodes a struct type into its Cloudstate Any JSON value.
func MarshalJSON(value interface{}) (*any.Any, error) {
	typeOf := reflect.TypeOf(value)
	if typeOf.Kind() == reflect.Ptr {
		typeOf = reflect.ValueOf(value).Elem().Type()
	}
	buffer := proto.NewBuffer(make([]byte, 0))
	buffer.SetDeterministic(true)
	typeURL := fmt.Sprintf("%s/%s.%s", JSONTypeURLPrefix, typeOf.PkgPath(), typeOf.Name())
	_ = buffer.EncodeVarint(fieldKey | proto.WireBytes)
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	_ = buffer.EncodeRawBytes(bytes)
	return &any.Any{
		TypeUrl: typeURL,
		Value:   buffer.Bytes(),
	}, nil
}

// UnmarshalPrimitive decodes a Cloudstate Any protobuf message into its JSON value.
func UnmarshalJSON(any *any.Any, target interface{}) error {
	if !strings.HasPrefix(any.GetTypeUrl(), JSONTypeURLPrefix) {
		return ErrNotMarshalled
	}
	buffer := proto.NewBuffer(any.GetValue())
	_, err := buffer.DecodeVarint()
	if err != nil {
		return ErrNotUnmarshalled
	}
	bytes, err := buffer.DecodeRawBytes(true)
	if err != nil {
		return ErrNotUnmarshalled
	}
	return json.Unmarshal(bytes, target)
}
