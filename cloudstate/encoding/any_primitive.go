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
	"math"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	PrimitiveTypeURLPrefix = "p.cloudstate.io"
	ProtoAnyBase           = "type.googleapis.com"

	PrimitiveTypeURLPrefixInt32  = PrimitiveTypeURLPrefix + "/int32"
	PrimitiveTypeURLPrefixInt64  = PrimitiveTypeURLPrefix + "/int64"
	PrimitiveTypeURLPrefixString = PrimitiveTypeURLPrefix + "/string"
	PrimitiveTypeURLPrefixFloat  = PrimitiveTypeURLPrefix + "/float"
	PrimitiveTypeURLPrefixDouble = PrimitiveTypeURLPrefix + "/double"
	PrimitiveTypeURLPrefixBool   = PrimitiveTypeURLPrefix + "/bool"
	PrimitiveTypeURLPrefixBytes  = PrimitiveTypeURLPrefix + "/bytes"
)

const fieldKey = 1 << 3

func MarshalPrimitive(i interface{}) (*any.Any, error) {
	buffer := proto.NewBuffer(make([]byte, 0))
	buffer.SetDeterministic(true)
	// see https://developers.google.com/protocol-buffers/docs/encoding#structure
	var typeURL string
	switch val := i.(type) {
	case int32:
		typeURL = PrimitiveTypeURLPrefixInt32
		_ = buffer.EncodeVarint(fieldKey | proto.WireVarint)
		_ = buffer.EncodeVarint(uint64(val))
	case int64:
		typeURL = PrimitiveTypeURLPrefixInt64
		_ = buffer.EncodeVarint(fieldKey | proto.WireVarint)
		_ = buffer.EncodeVarint(uint64(val))
	case string:
		typeURL = PrimitiveTypeURLPrefixString
		if val != "" {
			// see: https://cloudstate.io/docs/contribute/serialization.html#primitive-value-stability
			_ = buffer.EncodeVarint(fieldKey | proto.WireBytes)
			if err := buffer.EncodeStringBytes(val); err != nil {
				return nil, err
			}
		}
	case float32:
		typeURL = PrimitiveTypeURLPrefixFloat
		_ = buffer.EncodeVarint(fieldKey | proto.WireFixed32)
		_ = buffer.EncodeFixed32(uint64(math.Float32bits(val)))
	case float64:
		typeURL = PrimitiveTypeURLPrefixDouble
		_ = buffer.EncodeVarint(fieldKey | proto.WireFixed64)
		_ = buffer.EncodeFixed64(math.Float64bits(val))
	case bool:
		typeURL = PrimitiveTypeURLPrefixBool
		_ = buffer.EncodeVarint(fieldKey | proto.WireVarint)
		switch val {
		case true:
			_ = buffer.EncodeVarint(1)
		case false:
			_ = buffer.EncodeVarint(0)
		}
	case []byte:
		typeURL = PrimitiveTypeURLPrefixBytes
		_ = buffer.EncodeVarint(fieldKey | proto.WireBytes)
		if err := buffer.EncodeRawBytes(val); err != nil {
			return nil, err
		}
	default:
		return nil, ErrNotMarshalled
	}
	return &any.Any{
		TypeUrl: typeURL,
		Value:   buffer.Bytes(),
	}, nil
}

// UnmarshalPrimitive decodes a CloudState Any proto message
// into its primitive value.
func UnmarshalPrimitive(any *any.Any) (interface{}, error) {
	buffer := proto.NewBuffer(any.GetValue())
	if any.GetTypeUrl() == PrimitiveTypeURLPrefixInt32 {
		_, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		value, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		return int32(value), nil
	}
	if any.GetTypeUrl() == PrimitiveTypeURLPrefixInt64 {
		_, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		value, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		return int64(value), nil
	}
	if any.GetTypeUrl() == PrimitiveTypeURLPrefixString {
		_, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		value, err := buffer.DecodeStringBytes()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		return value, nil
	}
	if any.GetTypeUrl() == PrimitiveTypeURLPrefixFloat {
		_, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		value, err := buffer.DecodeFixed32()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		return math.Float32frombits(uint32(value)), nil
	}
	if any.GetTypeUrl() == PrimitiveTypeURLPrefixDouble {
		_, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		value, err := buffer.DecodeFixed64()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		return math.Float64frombits(value), nil
	}
	if any.GetTypeUrl() == PrimitiveTypeURLPrefixBool {
		_, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		value, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		return value == 1, nil
	}
	if any.GetTypeUrl() == PrimitiveTypeURLPrefixBytes {
		_, err := buffer.DecodeVarint()
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		value, err := buffer.DecodeRawBytes(true)
		if err != nil {
			return nil, ErrNotUnmarshalled
		}
		return value, nil
	}
	return nil, nil
}
