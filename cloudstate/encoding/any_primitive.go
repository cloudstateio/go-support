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
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"math"
	"reflect"
)

const (
	PrimitiveTypeURLPrefix = "p.cloudstate.io"

	primitiveTypeURLPrefixInt32  = PrimitiveTypeURLPrefix + "/int32"
	primitiveTypeURLPrefixInt64  = PrimitiveTypeURLPrefix + "/int64"
	primitiveTypeURLPrefixString = PrimitiveTypeURLPrefix + "/string"
	primitiveTypeURLPrefixFloat  = PrimitiveTypeURLPrefix + "/float"
	primitiveTypeURLPrefixDouble = PrimitiveTypeURLPrefix + "/double"
	primitiveTypeURLPrefixBool   = PrimitiveTypeURLPrefix + "/bool"
	primitiveTypeURLPrefixBytes  = PrimitiveTypeURLPrefix + "/bytes"
)

const fieldKey = 1 << 3

func MarshalPrimitive(i interface{}) (*any.Any, error) {
	buf := make([]byte, 0)
	buffer := proto.NewBuffer(buf)
	buffer.SetDeterministic(true)
	// see https://developers.google.com/protocol-buffers/docs/encoding#structure
	var typeUrl string
	switch val := i.(type) {
	case int32:
		typeUrl = primitiveTypeURLPrefixInt32
		_ = buffer.EncodeVarint(fieldKey | proto.WireVarint)
		_ = buffer.EncodeVarint(uint64(val))
	case int64:
		typeUrl = primitiveTypeURLPrefixInt64
		_ = buffer.EncodeVarint(fieldKey | proto.WireVarint)
		_ = buffer.EncodeVarint(uint64(val))
	case string:
		typeUrl = primitiveTypeURLPrefixString
		_ = buffer.EncodeVarint(fieldKey | proto.WireBytes)
		if err := buffer.EncodeStringBytes(val); err != nil {
			return nil, err
		}
	case float32:
		typeUrl = primitiveTypeURLPrefixFloat
		_ = buffer.EncodeVarint(fieldKey | proto.WireFixed32)
		_ = buffer.EncodeFixed32(uint64(math.Float32bits(val)))
	case float64:
		typeUrl = primitiveTypeURLPrefixDouble
		_ = buffer.EncodeVarint(fieldKey | proto.WireFixed64)
		_ = buffer.EncodeFixed64(math.Float64bits(val))
	case bool:
		typeUrl = primitiveTypeURLPrefixBool
		_ = buffer.EncodeVarint(fieldKey | proto.WireVarint)
		switch val {
		case true:
			_ = buffer.EncodeVarint(1)
		case false:
			_ = buffer.EncodeVarint(0)
		}
	case []byte:
		typeUrl = primitiveTypeURLPrefixBytes
		_ = buffer.EncodeVarint(fieldKey | proto.WireBytes)
		if err := buffer.EncodeRawBytes(val); err != nil {
			return nil, err
		}
	case interface{}:
		typeOf := reflect.TypeOf(val)
		if typeOf.Kind() == reflect.Struct {
			typeUrl = jsonTypeURLPrefix + "/" + typeOf.PkgPath() + "." + typeOf.Name()
			_ = buffer.EncodeVarint(fieldKey | proto.WireBytes)
			bytes, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}
			_ = buffer.EncodeRawBytes(bytes)
		} else {
			return nil, ErrNotMarshalled
		}
	default:
		return nil, ErrNotMarshalled
	}
	return &any.Any{
		TypeUrl: typeUrl,
		Value:   buffer.Bytes(),
	}, nil
}

// UnmarshalPrimitive decodes a CloudState Any proto message
// into its primitive value.
func UnmarshalPrimitive(any *any.Any) (interface{}, error) {
	buffer := proto.NewBuffer(any.GetValue())
	if any.GetTypeUrl() == primitiveTypeURLPrefixInt32 {
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
	if any.GetTypeUrl() == primitiveTypeURLPrefixInt64 {
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
	if any.GetTypeUrl() == primitiveTypeURLPrefixString {
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
	if any.GetTypeUrl() == primitiveTypeURLPrefixFloat {
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
	if any.GetTypeUrl() == primitiveTypeURLPrefixDouble {
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
	if any.GetTypeUrl() == primitiveTypeURLPrefixBool {
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
	if any.GetTypeUrl() == primitiveTypeURLPrefixBytes {
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
