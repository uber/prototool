// Code generated by protoc-gen-go. DO NOT EDIT.
// source: uber/bar/v1/bar.proto

package barv1

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Hello struct {
	Hello                int64    `protobuf:"varint,1,opt,name=hello,proto3" json:"hello,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Hello) Reset()         { *m = Hello{} }
func (m *Hello) String() string { return proto.CompactTextString(m) }
func (*Hello) ProtoMessage()    {}
func (*Hello) Descriptor() ([]byte, []int) {
	return fileDescriptor_bar_ab45b574cbbc9a08, []int{0}
}
func (m *Hello) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Hello.Unmarshal(m, b)
}
func (m *Hello) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Hello.Marshal(b, m, deterministic)
}
func (dst *Hello) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Hello.Merge(dst, src)
}
func (m *Hello) XXX_Size() int {
	return xxx_messageInfo_Hello.Size(m)
}
func (m *Hello) XXX_DiscardUnknown() {
	xxx_messageInfo_Hello.DiscardUnknown(m)
}

var xxx_messageInfo_Hello proto.InternalMessageInfo

func (m *Hello) GetHello() int64 {
	if m != nil {
		return m.Hello
	}
	return 0
}

func init() {
	proto.RegisterType((*Hello)(nil), "uber.bar.v1.Hello")
}

func init() { proto.RegisterFile("uber/bar/v1/bar.proto", fileDescriptor_bar_ab45b574cbbc9a08) }

var fileDescriptor_bar_ab45b574cbbc9a08 = []byte{
	// 138 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2d, 0x4d, 0x4a, 0x2d,
	0xd2, 0x4f, 0x4a, 0x2c, 0xd2, 0x2f, 0x33, 0x04, 0x51, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42,
	0xdc, 0x20, 0x61, 0x3d, 0x10, 0xbf, 0xcc, 0x50, 0x49, 0x96, 0x8b, 0xd5, 0x23, 0x35, 0x27, 0x27,
	0x5f, 0x48, 0x84, 0x8b, 0x35, 0x03, 0xc4, 0x90, 0x60, 0x54, 0x60, 0xd4, 0x60, 0x0e, 0x82, 0x70,
	0x9c, 0xdc, 0xb8, 0xf8, 0x93, 0xf3, 0x73, 0xf5, 0x90, 0x74, 0x38, 0x71, 0x38, 0x25, 0x16, 0x05,
	0x80, 0x0c, 0x0a, 0x60, 0x8c, 0x62, 0x4d, 0x4a, 0x2c, 0x2a, 0x33, 0x5c, 0xc4, 0xc4, 0x1c, 0xea,
	0x14, 0xb1, 0x8a, 0x89, 0x3b, 0x14, 0xa4, 0xcc, 0x29, 0xb1, 0x48, 0x2f, 0xcc, 0xf0, 0x14, 0x84,
	0x17, 0xe3, 0x94, 0x58, 0x14, 0x13, 0x66, 0x98, 0xc4, 0x06, 0xb6, 0xda, 0x18, 0x10, 0x00, 0x00,
	0xff, 0xff, 0xad, 0x26, 0xa1, 0xb3, 0x93, 0x00, 0x00, 0x00,
}
