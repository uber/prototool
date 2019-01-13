// Code generated by protoc-gen-go. DO NOT EDIT.
// source: uber/foo/v1/foo.proto

package foov1

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import duration "github.com/golang/protobuf/ptypes/duration"
import timestamp "github.com/golang/protobuf/ptypes/timestamp"
import v1 "github.com/uber/prototool/example/gen/go/uber/bar/v1"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Hello is a hello.
type Hello int32

const (
	Hello_HELLO_INVALID Hello = 0
	Hello_HELLO_UNSET   Hello = 1
	Hello_HELLO_TREE    Hello = 2
	Hello_HELLO_BALLOON Hello = 3
)

var Hello_name = map[int32]string{
	0: "HELLO_INVALID",
	1: "HELLO_UNSET",
	2: "HELLO_TREE",
	3: "HELLO_BALLOON",
}
var Hello_value = map[string]int32{
	"HELLO_INVALID": 0,
	"HELLO_UNSET":   1,
	"HELLO_TREE":    2,
	"HELLO_BALLOON": 3,
}

func (x Hello) String() string {
	return proto.EnumName(Hello_name, int32(x))
}
func (Hello) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{0}
}

// Bar is a bar.
type Bar int32

const (
	Bar_BAR_INVALID Bar = 0
)

var Bar_name = map[int32]string{
	0: "BAR_INVALID",
}
var Bar_value = map[string]int32{
	"BAR_INVALID": 0,
}

func (x Bar) String() string {
	return proto.EnumName(Bar_name, int32(x))
}
func (Bar) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{1}
}

// Baz is a baz.
type Foo_Bar_Baz int32

const (
	Foo_Bar_BAZ_INVALID Foo_Bar_Baz = 0
)

var Foo_Bar_Baz_name = map[int32]string{
	0: "BAZ_INVALID",
}
var Foo_Bar_Baz_value = map[string]int32{
	"BAZ_INVALID": 0,
}

func (x Foo_Bar_Baz) String() string {
	return proto.EnumName(Foo_Bar_Baz_name, int32(x))
}
func (Foo_Bar_Baz) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{0, 0, 0}
}

// Bat is a bat.
type Foo_Bar_Bat int32

const (
	Foo_Bar_BAT_INVALID Foo_Bar_Bat = 0
)

var Foo_Bar_Bat_name = map[int32]string{
	0: "BAT_INVALID",
}
var Foo_Bar_Bat_value = map[string]int32{
	"BAT_INVALID": 0,
}

func (x Foo_Bar_Bat) String() string {
	return proto.EnumName(Foo_Bar_Bat_name, int32(x))
}
func (Foo_Bar_Bat) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{0, 0, 1}
}

// Foo is a foo.
type Foo struct {
	Bar                  *Foo_Bar `protobuf:"bytes,1,opt,name=bar,proto3" json:"bar,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Foo) Reset()         { *m = Foo{} }
func (m *Foo) String() string { return proto.CompactTextString(m) }
func (*Foo) ProtoMessage()    {}
func (*Foo) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{0}
}
func (m *Foo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Foo.Unmarshal(m, b)
}
func (m *Foo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Foo.Marshal(b, m, deterministic)
}
func (dst *Foo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Foo.Merge(dst, src)
}
func (m *Foo) XXX_Size() int {
	return xxx_messageInfo_Foo.Size(m)
}
func (m *Foo) XXX_DiscardUnknown() {
	xxx_messageInfo_Foo.DiscardUnknown(m)
}

var xxx_messageInfo_Foo proto.InternalMessageInfo

func (m *Foo) GetBar() *Foo_Bar {
	if m != nil {
		return m.Bar
	}
	return nil
}

// Bar is a bar.
type Foo_Bar struct {
	Baz                  Foo_Bar_Baz `protobuf:"varint,1,opt,name=baz,proto3,enum=uber.foo.v1.Foo_Bar_Baz" json:"baz,omitempty"`
	Bat                  Foo_Bar_Bat `protobuf:"varint,2,opt,name=bat,proto3,enum=uber.foo.v1.Foo_Bar_Bat" json:"bat,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Foo_Bar) Reset()         { *m = Foo_Bar{} }
func (m *Foo_Bar) String() string { return proto.CompactTextString(m) }
func (*Foo_Bar) ProtoMessage()    {}
func (*Foo_Bar) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{0, 0}
}
func (m *Foo_Bar) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Foo_Bar.Unmarshal(m, b)
}
func (m *Foo_Bar) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Foo_Bar.Marshal(b, m, deterministic)
}
func (dst *Foo_Bar) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Foo_Bar.Merge(dst, src)
}
func (m *Foo_Bar) XXX_Size() int {
	return xxx_messageInfo_Foo_Bar.Size(m)
}
func (m *Foo_Bar) XXX_DiscardUnknown() {
	xxx_messageInfo_Foo_Bar.DiscardUnknown(m)
}

var xxx_messageInfo_Foo_Bar proto.InternalMessageInfo

func (m *Foo_Bar) GetBaz() Foo_Bar_Baz {
	if m != nil {
		return m.Baz
	}
	return Foo_Bar_BAZ_INVALID
}

func (m *Foo_Bar) GetBat() Foo_Bar_Bat {
	if m != nil {
		return m.Bat
	}
	return Foo_Bar_BAT_INVALID
}

// Barr is a barr.
type Barr struct {
	Hello                int64                `protobuf:"varint,1,opt,name=hello,proto3" json:"hello,omitempty"`
	BarrTime             *timestamp.Timestamp `protobuf:"bytes,2,opt,name=barr_time,json=barrTime,proto3" json:"barr_time,omitempty"`
	BarHello             *v1.Hello            `protobuf:"bytes,3,opt,name=bar_hello,json=barHello,proto3" json:"bar_hello,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Barr) Reset()         { *m = Barr{} }
func (m *Barr) String() string { return proto.CompactTextString(m) }
func (*Barr) ProtoMessage()    {}
func (*Barr) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{1}
}
func (m *Barr) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Barr.Unmarshal(m, b)
}
func (m *Barr) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Barr.Marshal(b, m, deterministic)
}
func (dst *Barr) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Barr.Merge(dst, src)
}
func (m *Barr) XXX_Size() int {
	return xxx_messageInfo_Barr.Size(m)
}
func (m *Barr) XXX_DiscardUnknown() {
	xxx_messageInfo_Barr.DiscardUnknown(m)
}

var xxx_messageInfo_Barr proto.InternalMessageInfo

func (m *Barr) GetHello() int64 {
	if m != nil {
		return m.Hello
	}
	return 0
}

func (m *Barr) GetBarrTime() *timestamp.Timestamp {
	if m != nil {
		return m.BarrTime
	}
	return nil
}

func (m *Barr) GetBarHello() *v1.Hello {
	if m != nil {
		return m.BarHello
	}
	return nil
}

// Another is another message.
type Another struct {
	One     int64            `protobuf:"varint,1,opt,name=one,proto3" json:"one,omitempty"`
	Two     string           `protobuf:"bytes,2,opt,name=two,proto3" json:"two,omitempty"`
	Another *Another         `protobuf:"bytes,3,opt,name=another,proto3" json:"another,omitempty"`
	Four    []string         `protobuf:"bytes,4,rep,name=four,proto3" json:"four,omitempty"`
	Hello   Hello            `protobuf:"varint,5,opt,name=hello,proto3,enum=uber.foo.v1.Hello" json:"hello,omitempty"`
	M       map[string]int64 `protobuf:"bytes,6,rep,name=m,proto3" json:"m,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// Types that are valid to be assigned to OneofOneof:
	//	*Another_Seven
	//	*Another_Eight
	OneofOneof           isAnother_OneofOneof `protobuf_oneof:"oneof_oneof"`
	Nine                 []*Another           `protobuf:"bytes,9,rep,name=nine,proto3" json:"nine,omitempty"`
	AnotherDuration      *duration.Duration   `protobuf:"bytes,10,opt,name=another_duration,json=anotherDuration,proto3" json:"another_duration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Another) Reset()         { *m = Another{} }
func (m *Another) String() string { return proto.CompactTextString(m) }
func (*Another) ProtoMessage()    {}
func (*Another) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{2}
}
func (m *Another) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Another.Unmarshal(m, b)
}
func (m *Another) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Another.Marshal(b, m, deterministic)
}
func (dst *Another) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Another.Merge(dst, src)
}
func (m *Another) XXX_Size() int {
	return xxx_messageInfo_Another.Size(m)
}
func (m *Another) XXX_DiscardUnknown() {
	xxx_messageInfo_Another.DiscardUnknown(m)
}

var xxx_messageInfo_Another proto.InternalMessageInfo

func (m *Another) GetOne() int64 {
	if m != nil {
		return m.One
	}
	return 0
}

func (m *Another) GetTwo() string {
	if m != nil {
		return m.Two
	}
	return ""
}

func (m *Another) GetAnother() *Another {
	if m != nil {
		return m.Another
	}
	return nil
}

func (m *Another) GetFour() []string {
	if m != nil {
		return m.Four
	}
	return nil
}

func (m *Another) GetHello() Hello {
	if m != nil {
		return m.Hello
	}
	return Hello_HELLO_INVALID
}

func (m *Another) GetM() map[string]int64 {
	if m != nil {
		return m.M
	}
	return nil
}

type isAnother_OneofOneof interface {
	isAnother_OneofOneof()
}

type Another_Seven struct {
	Seven int64 `protobuf:"varint,7,opt,name=seven,proto3,oneof"`
}

type Another_Eight struct {
	Eight string `protobuf:"bytes,8,opt,name=eight,proto3,oneof"`
}

func (*Another_Seven) isAnother_OneofOneof() {}

func (*Another_Eight) isAnother_OneofOneof() {}

func (m *Another) GetOneofOneof() isAnother_OneofOneof {
	if m != nil {
		return m.OneofOneof
	}
	return nil
}

func (m *Another) GetSeven() int64 {
	if x, ok := m.GetOneofOneof().(*Another_Seven); ok {
		return x.Seven
	}
	return 0
}

func (m *Another) GetEight() string {
	if x, ok := m.GetOneofOneof().(*Another_Eight); ok {
		return x.Eight
	}
	return ""
}

func (m *Another) GetNine() []*Another {
	if m != nil {
		return m.Nine
	}
	return nil
}

func (m *Another) GetAnotherDuration() *duration.Duration {
	if m != nil {
		return m.AnotherDuration
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Another) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Another_OneofMarshaler, _Another_OneofUnmarshaler, _Another_OneofSizer, []interface{}{
		(*Another_Seven)(nil),
		(*Another_Eight)(nil),
	}
}

func _Another_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Another)
	// oneof_oneof
	switch x := m.OneofOneof.(type) {
	case *Another_Seven:
		b.EncodeVarint(7<<3 | proto.WireVarint)
		b.EncodeVarint(uint64(x.Seven))
	case *Another_Eight:
		b.EncodeVarint(8<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.Eight)
	case nil:
	default:
		return fmt.Errorf("Another.OneofOneof has unexpected type %T", x)
	}
	return nil
}

func _Another_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Another)
	switch tag {
	case 7: // oneof_oneof.seven
		if wire != proto.WireVarint {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeVarint()
		m.OneofOneof = &Another_Seven{int64(x)}
		return true, err
	case 8: // oneof_oneof.eight
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.OneofOneof = &Another_Eight{x}
		return true, err
	default:
		return false, nil
	}
}

func _Another_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Another)
	// oneof_oneof
	switch x := m.OneofOneof.(type) {
	case *Another_Seven:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(x.Seven))
	case *Another_Eight:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.Eight)))
		n += len(x.Eight)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// HasWKT has Well-Known Types.
type HasWKT struct {
	WktDuration          *duration.Duration `protobuf:"bytes,1,opt,name=wkt_duration,json=wktDuration,proto3" json:"wkt_duration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *HasWKT) Reset()         { *m = HasWKT{} }
func (m *HasWKT) String() string { return proto.CompactTextString(m) }
func (*HasWKT) ProtoMessage()    {}
func (*HasWKT) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{3}
}
func (m *HasWKT) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HasWKT.Unmarshal(m, b)
}
func (m *HasWKT) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HasWKT.Marshal(b, m, deterministic)
}
func (dst *HasWKT) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HasWKT.Merge(dst, src)
}
func (m *HasWKT) XXX_Size() int {
	return xxx_messageInfo_HasWKT.Size(m)
}
func (m *HasWKT) XXX_DiscardUnknown() {
	xxx_messageInfo_HasWKT.DiscardUnknown(m)
}

var xxx_messageInfo_HasWKT proto.InternalMessageInfo

func (m *HasWKT) GetWktDuration() *duration.Duration {
	if m != nil {
		return m.WktDuration
	}
	return nil
}

// Bazz is a bazzzz.
type Bazz struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Bazz) Reset()         { *m = Bazz{} }
func (m *Bazz) String() string { return proto.CompactTextString(m) }
func (*Bazz) ProtoMessage()    {}
func (*Bazz) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_18696cee75944a2c, []int{4}
}
func (m *Bazz) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Bazz.Unmarshal(m, b)
}
func (m *Bazz) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Bazz.Marshal(b, m, deterministic)
}
func (dst *Bazz) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Bazz.Merge(dst, src)
}
func (m *Bazz) XXX_Size() int {
	return xxx_messageInfo_Bazz.Size(m)
}
func (m *Bazz) XXX_DiscardUnknown() {
	xxx_messageInfo_Bazz.DiscardUnknown(m)
}

var xxx_messageInfo_Bazz proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Foo)(nil), "uber.foo.v1.Foo")
	proto.RegisterType((*Foo_Bar)(nil), "uber.foo.v1.Foo.Bar")
	proto.RegisterType((*Barr)(nil), "uber.foo.v1.Barr")
	proto.RegisterType((*Another)(nil), "uber.foo.v1.Another")
	proto.RegisterMapType((map[string]int64)(nil), "uber.foo.v1.Another.MEntry")
	proto.RegisterType((*HasWKT)(nil), "uber.foo.v1.HasWKT")
	proto.RegisterType((*Bazz)(nil), "uber.foo.v1.Bazz")
	proto.RegisterEnum("uber.foo.v1.Hello", Hello_name, Hello_value)
	proto.RegisterEnum("uber.foo.v1.Bar", Bar_name, Bar_value)
	proto.RegisterEnum("uber.foo.v1.Foo_Bar_Baz", Foo_Bar_Baz_name, Foo_Bar_Baz_value)
	proto.RegisterEnum("uber.foo.v1.Foo_Bar_Bat", Foo_Bar_Bat_name, Foo_Bar_Bat_value)
}

func init() { proto.RegisterFile("uber/foo/v1/foo.proto", fileDescriptor_foo_18696cee75944a2c) }

var fileDescriptor_foo_18696cee75944a2c = []byte{
	// 625 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x53, 0x51, 0x6f, 0x12, 0x41,
	0x10, 0xee, 0x71, 0x40, 0xcb, 0x9c, 0x6d, 0x71, 0x53, 0x9b, 0x13, 0x13, 0x6d, 0x78, 0x30, 0xd8,
	0x87, 0x23, 0xa0, 0x89, 0xc6, 0xf8, 0xc2, 0xa5, 0x5c, 0x68, 0x44, 0xda, 0xac, 0xb4, 0x9a, 0xa6,
	0x09, 0xd9, 0xd3, 0xa5, 0x25, 0x85, 0x9b, 0x66, 0xbb, 0xd0, 0x94, 0x1f, 0xe0, 0x9b, 0x7f, 0xc2,
	0x47, 0xfd, 0x0d, 0xfe, 0x01, 0x7f, 0x95, 0x99, 0xdd, 0xbb, 0xd0, 0x2a, 0xd1, 0x17, 0x98, 0xf9,
	0xf6, 0x9b, 0xd9, 0x6f, 0xbe, 0x9b, 0x85, 0x07, 0xd3, 0x58, 0xaa, 0xfa, 0x10, 0xb1, 0x3e, 0x6b,
	0xd0, 0x5f, 0x70, 0xa9, 0x50, 0x23, 0xf3, 0x08, 0x0e, 0x28, 0x9f, 0x35, 0x2a, 0x8f, 0xcf, 0x10,
	0xcf, 0xc6, 0xb2, 0x6e, 0x8e, 0xe2, 0xe9, 0xb0, 0xfe, 0x79, 0xaa, 0x84, 0x1e, 0x61, 0x62, 0xc9,
	0x95, 0x27, 0x7f, 0x9e, 0xeb, 0xd1, 0x44, 0x5e, 0x69, 0x31, 0xb9, 0x4c, 0x09, 0xf6, 0x92, 0x58,
	0x28, 0xba, 0x24, 0x16, 0xca, 0xc2, 0xd5, 0x9f, 0x0e, 0xb8, 0x11, 0x22, 0x7b, 0x0a, 0x6e, 0x2c,
	0x94, 0xef, 0xec, 0x38, 0x35, 0xaf, 0xb9, 0x15, 0xdc, 0xba, 0x3a, 0x88, 0x10, 0x83, 0x50, 0x28,
	0x4e, 0x84, 0xca, 0x57, 0x07, 0xdc, 0x50, 0x28, 0xb6, 0x4b, 0xfc, 0xb9, 0xe1, 0x6f, 0x34, 0xfd,
	0x65, 0xfc, 0x20, 0x14, 0x73, 0xaa, 0x99, 0x5b, 0xae, 0xf6, 0x73, 0xff, 0xe4, 0x6a, 0xe2, 0xea,
	0xea, 0x36, 0xb5, 0x9f, 0xb3, 0x4d, 0xf0, 0xc2, 0xd6, 0xc9, 0x60, 0xbf, 0x77, 0xdc, 0xea, 0xee,
	0xef, 0x95, 0x57, 0x2c, 0xae, 0x2d, 0xde, 0xbf, 0x85, 0x7f, 0x71, 0x20, 0x1f, 0x0a, 0xa5, 0xd8,
	0x16, 0x14, 0xce, 0xe5, 0x78, 0x8c, 0x46, 0x92, 0xcb, 0x6d, 0xc2, 0x5e, 0x42, 0x29, 0x16, 0x4a,
	0x0d, 0xc8, 0x0d, 0x23, 0xc0, 0x6b, 0x56, 0x02, 0x6b, 0x55, 0x90, 0x59, 0x15, 0xf4, 0x33, 0xab,
	0xf8, 0x1a, 0x91, 0x29, 0x65, 0x75, 0x53, 0x38, 0xb0, 0x2d, 0x5d, 0x53, 0xc8, 0xac, 0x72, 0xf2,
	0x6e, 0xd6, 0x08, 0x3a, 0x74, 0x62, 0x0a, 0x4c, 0x54, 0xfd, 0xe1, 0xc2, 0x6a, 0x2b, 0x41, 0x7d,
	0x2e, 0x15, 0x2b, 0x83, 0x8b, 0x89, 0x4c, 0x95, 0x50, 0x48, 0x88, 0xbe, 0x46, 0xa3, 0xa0, 0xc4,
	0x29, 0x64, 0x01, 0xac, 0x0a, 0x4b, 0x4f, 0xdb, 0xdf, 0x35, 0x3d, 0x6d, 0xc5, 0x33, 0x12, 0x63,
	0x90, 0x1f, 0xe2, 0x54, 0xf9, 0xf9, 0x1d, 0xb7, 0x56, 0xe2, 0x26, 0x66, 0xb5, 0x6c, 0xe6, 0x82,
	0xb1, 0x96, 0xdd, 0xe9, 0x60, 0x05, 0xa6, 0x3e, 0x3c, 0x03, 0x67, 0xe2, 0x17, 0x77, 0xdc, 0x9a,
	0xd7, 0x7c, 0xb4, 0xec, 0x9e, 0xe0, 0x5d, 0x3b, 0xd1, 0xea, 0x86, 0x3b, 0x13, 0xb6, 0x0d, 0x85,
	0x2b, 0x39, 0x93, 0x89, 0xbf, 0x4a, 0xf2, 0x3b, 0x2b, 0xdc, 0xa6, 0x84, 0xcb, 0xd1, 0xd9, 0xb9,
	0xf6, 0xd7, 0x68, 0x08, 0xc2, 0x4d, 0xca, 0x6a, 0x90, 0x4f, 0x46, 0x89, 0xf4, 0x4b, 0xa6, 0xfb,
	0xf2, 0x29, 0x0c, 0x83, 0xed, 0x41, 0x39, 0x9d, 0x66, 0x90, 0x6d, 0xaf, 0x0f, 0x66, 0xf6, 0x87,
	0x7f, 0x7d, 0x93, 0xbd, 0x94, 0xc0, 0x37, 0xd3, 0x92, 0x0c, 0xa8, 0xbc, 0x80, 0xa2, 0x15, 0x4b,
	0xa6, 0x5e, 0xc8, 0x1b, 0x63, 0x73, 0x89, 0x53, 0x48, 0x4b, 0x30, 0x13, 0xe3, 0xa9, 0xfd, 0xd4,
	0x2e, 0xb7, 0xc9, 0xeb, 0xdc, 0x2b, 0x27, 0x5c, 0x07, 0x0f, 0x13, 0x89, 0xc3, 0x81, 0xf9, 0xad,
	0x46, 0x50, 0xec, 0x88, 0xab, 0x0f, 0x6f, 0xfb, 0xec, 0x0d, 0xdc, 0xbb, 0xbe, 0xd0, 0x0b, 0x41,
	0xce, 0xff, 0x04, 0x79, 0xd7, 0x17, 0x3a, 0x4b, 0xaa, 0x45, 0xda, 0xbe, 0xf9, 0x7c, 0xb7, 0x07,
	0x05, 0xe3, 0x37, 0xbb, 0x0f, 0xeb, 0x9d, 0x76, 0xb7, 0x7b, 0xb0, 0x58, 0x51, 0xda, 0x59, 0x0b,
	0x1d, 0xf5, 0xde, 0xb7, 0xfb, 0x65, 0x87, 0x6d, 0x00, 0x58, 0xa0, 0xcf, 0xdb, 0xed, 0x72, 0x6e,
	0x51, 0x13, 0xb6, 0xba, 0xdd, 0x83, 0x83, 0x5e, 0xd9, 0xdd, 0xdd, 0xb6, 0xaf, 0xcc, 0xac, 0x3b,
	0x5f, 0xf4, 0x0a, 0x23, 0xd8, 0xfc, 0x84, 0x93, 0xdb, 0x1e, 0x87, 0x6b, 0x11, 0xe2, 0x21, 0xa9,
	0x3c, 0x74, 0x4e, 0x0a, 0x43, 0xc4, 0x59, 0xe3, 0x5b, 0xce, 0x3d, 0x8a, 0x3e, 0x7e, 0xcf, 0x79,
	0x47, 0x44, 0xa3, 0x27, 0x76, 0xdc, 0xf8, 0x65, 0xb3, 0xd3, 0x08, 0xf1, 0xf4, 0xb8, 0x11, 0x17,
	0xcd, 0x5c, 0xcf, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0x4c, 0xde, 0xd3, 0x9d, 0x7b, 0x04, 0x00,
	0x00,
}
