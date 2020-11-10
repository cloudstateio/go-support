package encoding

import "github.com/golang/protobuf/ptypes/any"

func Int32(i int32) *any.Any {
	primitive, _ := MarshalPrimitive(i)
	return primitive
}

func DecodeInt32(a *any.Any) int32 {
	i, _ := UnmarshalPrimitive(a)
	return i.(int32)
}

func Int64(i int64) *any.Any {
	primitive, _ := MarshalPrimitive(i)
	return primitive
}

func DecodeInt64(a *any.Any) int64 {
	i, _ := UnmarshalPrimitive(a)
	return i.(int64)
}

func StringMust(s string) *any.Any {
	primitive, err := MarshalPrimitive(s)
	if err != nil {
		panic(err)
	}
	return primitive
}

func String(s string) *any.Any {
	primitive, _ := MarshalPrimitive(s)
	return primitive
}

func Struct(s interface{}) (*any.Any, error) {
	return MarshalJSON(s)
}

func DecodeStruct(a *any.Any, s interface{}) error {
	return UnmarshalJSON(a, s)
}

func DecodeString(a *any.Any) string {
	i, _ := UnmarshalPrimitive(a)
	return i.(string)
}

func Float32(f float32) *any.Any {
	primitive, _ := MarshalPrimitive(f)
	return primitive
}

func DecodeFloat32(a *any.Any) float32 {
	i, _ := UnmarshalPrimitive(a)
	return i.(float32)
}

func Float64(f float64) *any.Any {
	primitive, _ := MarshalPrimitive(f)
	return primitive
}

func DecodeFloat64(a *any.Any) float64 {
	i, _ := UnmarshalPrimitive(a)
	return i.(float64)
}

func Bool(b bool) *any.Any {
	primitive, _ := MarshalPrimitive(b)
	return primitive
}

func DecodeBool(a *any.Any) bool {
	i, _ := UnmarshalPrimitive(a)
	return i.(bool)
}

func Bytes(b []byte) *any.Any {
	primitive, _ := MarshalPrimitive(b)
	return primitive
}

func DecodeBytes(a *any.Any) []byte {
	i, _ := UnmarshalPrimitive(a)
	return i.([]byte)
}
