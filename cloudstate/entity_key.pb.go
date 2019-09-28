// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cloudstate/entity_key.proto

package cloudstate

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

var E_EntityKey = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         50002,
	Name:          "cloudstate.entity_key",
	Tag:           "varint,50002,opt,name=entity_key",
	Filename:      "cloudstate/entity_key.proto",
}

func init() {
	proto.RegisterExtension(E_EntityKey)
}

func init() { proto.RegisterFile("cloudstate/entity_key.proto", fileDescriptor_7bcabc3af9eb79b9) }

var fileDescriptor_7bcabc3af9eb79b9 = []byte{
	// 178 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4e, 0xce, 0xc9, 0x2f,
	0x4d, 0x29, 0x2e, 0x49, 0x2c, 0x49, 0xd5, 0x4f, 0xcd, 0x2b, 0xc9, 0x2c, 0xa9, 0x8c, 0xcf, 0x4e,
	0xad, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x42, 0x48, 0x4a, 0x29, 0xa4, 0xe7, 0xe7,
	0xa7, 0xe7, 0xa4, 0xea, 0x83, 0x65, 0x92, 0x4a, 0xd3, 0xf4, 0x53, 0x52, 0x8b, 0x93, 0x8b, 0x32,
	0x0b, 0x4a, 0xf2, 0x8b, 0x20, 0xaa, 0xad, 0xec, 0xb8, 0xb8, 0x10, 0x26, 0x08, 0xc9, 0xea, 0x41,
	0x34, 0xe8, 0xc1, 0x34, 0xe8, 0xb9, 0x65, 0xa6, 0xe6, 0xa4, 0xf8, 0x17, 0x94, 0x64, 0xe6, 0xe7,
	0x15, 0x4b, 0x5c, 0x6a, 0x63, 0x56, 0x60, 0xd4, 0xe0, 0x08, 0xe2, 0x84, 0x68, 0xf1, 0x4e, 0xad,
	0x74, 0x0a, 0xe2, 0xe2, 0xcd, 0xcc, 0xd7, 0x43, 0x58, 0x19, 0xe5, 0x98, 0x9e, 0x59, 0x92, 0x51,
	0x9a, 0xa4, 0x97, 0x9c, 0x9f, 0xab, 0x8f, 0x10, 0xce, 0xcc, 0xd7, 0x4f, 0xcf, 0xd7, 0x2d, 0x2e,
	0x2d, 0x28, 0xc8, 0x2f, 0x2a, 0x41, 0x12, 0x87, 0xb8, 0x2c, 0x39, 0x3f, 0xc7, 0x1a, 0x21, 0x96,
	0xc4, 0x06, 0x16, 0x34, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xb6, 0x70, 0x0b, 0x55, 0xe7, 0x00,
	0x00, 0x00,
}
