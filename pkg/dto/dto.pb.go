//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/dto/dto.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.15.6
// source: pkg/dto/dto.proto

package dto

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SuccessStatus int32

const (
	SuccessStatus_PARTIAL_SUCCESS SuccessStatus = 0
	SuccessStatus_FULLY_SUCCESS   SuccessStatus = 1
)

// Enum value maps for SuccessStatus.
var (
	SuccessStatus_name = map[int32]string{
		0: "PARTIAL_SUCCESS",
		1: "FULLY_SUCCESS",
	}
	SuccessStatus_value = map[string]int32{
		"PARTIAL_SUCCESS": 0,
		"FULLY_SUCCESS":   1,
	}
)

func (x SuccessStatus) Enum() *SuccessStatus {
	p := new(SuccessStatus)
	*p = x
	return p
}

func (x SuccessStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SuccessStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_dto_dto_proto_enumTypes[0].Descriptor()
}

func (SuccessStatus) Type() protoreflect.EnumType {
	return &file_pkg_dto_dto_proto_enumTypes[0]
}

func (x SuccessStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SuccessStatus.Descriptor instead.
func (SuccessStatus) EnumDescriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{0}
}

type FoundKey int32

const (
	FoundKey_KEY_NOT_FOUND FoundKey = 0
	FoundKey_KEY_FOUND     FoundKey = 1
)

// Enum value maps for FoundKey.
var (
	FoundKey_name = map[int32]string{
		0: "KEY_NOT_FOUND",
		1: "KEY_FOUND",
	}
	FoundKey_value = map[string]int32{
		"KEY_NOT_FOUND": 0,
		"KEY_FOUND":     1,
	}
)

func (x FoundKey) Enum() *FoundKey {
	p := new(FoundKey)
	*p = x
	return p
}

func (x FoundKey) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FoundKey) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_dto_dto_proto_enumTypes[1].Descriptor()
}

func (FoundKey) Type() protoreflect.EnumType {
	return &file_pkg_dto_dto_proto_enumTypes[1]
}

func (x FoundKey) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FoundKey.Descriptor instead.
func (FoundKey) EnumDescriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{1}
}

type PutRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key    []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Object []byte `protobuf:"bytes,2,opt,name=object,proto3" json:"object,omitempty"`
}

func (x *PutRequest) Reset() {
	*x = PutRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutRequest) ProtoMessage() {}

func (x *PutRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutRequest.ProtoReflect.Descriptor instead.
func (*PutRequest) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{0}
}

func (x *PutRequest) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *PutRequest) GetObject() []byte {
	if x != nil {
		return x.Object
	}
	return nil
}

type PutResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SuccessStatus SuccessStatus `protobuf:"varint,1,opt,name=successStatus,proto3,enum=dto.SuccessStatus" json:"successStatus,omitempty"`
}

func (x *PutResponse) Reset() {
	*x = PutResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutResponse) ProtoMessage() {}

func (x *PutResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutResponse.ProtoReflect.Descriptor instead.
func (*PutResponse) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{1}
}

func (x *PutResponse) GetSuccessStatus() SuccessStatus {
	if x != nil {
		return x.SuccessStatus
	}
	return SuccessStatus_PARTIAL_SUCCESS
}

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *GetRequest) Reset() {
	*x = GetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRequest) ProtoMessage() {}

func (x *GetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRequest.ProtoReflect.Descriptor instead.
func (*GetRequest) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{2}
}

func (x *GetRequest) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Object        []byte        `protobuf:"bytes,1,opt,name=object,proto3" json:"object,omitempty"`
	SuccessStatus SuccessStatus `protobuf:"varint,2,opt,name=successStatus,proto3,enum=dto.SuccessStatus" json:"successStatus,omitempty"`
	FoundKey      FoundKey      `protobuf:"varint,3,opt,name=foundKey,proto3,enum=dto.FoundKey" json:"foundKey,omitempty"`
}

func (x *GetResponse) Reset() {
	*x = GetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetResponse) ProtoMessage() {}

func (x *GetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetResponse.ProtoReflect.Descriptor instead.
func (*GetResponse) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{3}
}

func (x *GetResponse) GetObject() []byte {
	if x != nil {
		return x.Object
	}
	return nil
}

func (x *GetResponse) GetSuccessStatus() SuccessStatus {
	if x != nil {
		return x.SuccessStatus
	}
	return SuccessStatus_PARTIAL_SUCCESS
}

func (x *GetResponse) GetFoundKey() FoundKey {
	if x != nil {
		return x.FoundKey
	}
	return FoundKey_KEY_NOT_FOUND
}

type PutRepRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key         []byte       `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Object      []byte       `protobuf:"bytes,2,opt,name=object,proto3" json:"object,omitempty"`
	Vectorclock *VectorClock `protobuf:"bytes,3,opt,name=vectorclock,proto3" json:"vectorclock,omitempty"`
}

func (x *PutRepRequest) Reset() {
	*x = PutRepRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutRepRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutRepRequest) ProtoMessage() {}

func (x *PutRepRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutRepRequest.ProtoReflect.Descriptor instead.
func (*PutRepRequest) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{4}
}

func (x *PutRepRequest) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *PutRepRequest) GetObject() []byte {
	if x != nil {
		return x.Object
	}
	return nil
}

func (x *PutRepRequest) GetVectorclock() *VectorClock {
	if x != nil {
		return x.Vectorclock
	}
	return nil
}

type PutRepResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vectorclock *VectorClock `protobuf:"bytes,1,opt,name=vectorclock,proto3" json:"vectorclock,omitempty"`
}

func (x *PutRepResponse) Reset() {
	*x = PutRepResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutRepResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutRepResponse) ProtoMessage() {}

func (x *PutRepResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutRepResponse.ProtoReflect.Descriptor instead.
func (*PutRepResponse) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{5}
}

func (x *PutRepResponse) GetVectorclock() *VectorClock {
	if x != nil {
		return x.Vectorclock
	}
	return nil
}

type GetRepRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key         []byte       `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Vectorclock *VectorClock `protobuf:"bytes,2,opt,name=vectorclock,proto3" json:"vectorclock,omitempty"`
}

func (x *GetRepRequest) Reset() {
	*x = GetRepRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRepRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRepRequest) ProtoMessage() {}

func (x *GetRepRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRepRequest.ProtoReflect.Descriptor instead.
func (*GetRepRequest) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{6}
}

func (x *GetRepRequest) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *GetRepRequest) GetVectorclock() *VectorClock {
	if x != nil {
		return x.Vectorclock
	}
	return nil
}

type GetRepResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Object      []byte       `protobuf:"bytes,1,opt,name=object,proto3" json:"object,omitempty"`
	Vectorclock *VectorClock `protobuf:"bytes,2,opt,name=vectorclock,proto3" json:"vectorclock,omitempty"`
}

func (x *GetRepResponse) Reset() {
	*x = GetRepResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRepResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRepResponse) ProtoMessage() {}

func (x *GetRepResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRepResponse.ProtoReflect.Descriptor instead.
func (*GetRepResponse) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{7}
}

func (x *GetRepResponse) GetObject() []byte {
	if x != nil {
		return x.Object
	}
	return nil
}

func (x *GetRepResponse) GetVectorclock() *VectorClock {
	if x != nil {
		return x.Vectorclock
	}
	return nil
}

type VectorClock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vclock    map[int64]int64 `protobuf:"bytes,1,rep,name=vclock,proto3" json:"vclock,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	MachineID int64           `protobuf:"varint,2,opt,name=machineID,proto3" json:"machineID,omitempty"`
}

func (x *VectorClock) Reset() {
	*x = VectorClock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VectorClock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VectorClock) ProtoMessage() {}

func (x *VectorClock) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VectorClock.ProtoReflect.Descriptor instead.
func (*VectorClock) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{8}
}

func (x *VectorClock) GetVclock() map[int64]int64 {
	if x != nil {
		return x.Vclock
	}
	return nil
}

func (x *VectorClock) GetMachineID() int64 {
	if x != nil {
		return x.MachineID
	}
	return 0
}

type Consistent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NumVNodes int64    `protobuf:"varint,1,opt,name=numVNodes,proto3" json:"numVNodes,omitempty"`
	Nodes     []uint64 `protobuf:"varint,2,rep,packed,name=nodes,proto3" json:"nodes,omitempty"`
}

func (x *Consistent) Reset() {
	*x = Consistent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_dto_dto_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Consistent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Consistent) ProtoMessage() {}

func (x *Consistent) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_dto_dto_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Consistent.ProtoReflect.Descriptor instead.
func (*Consistent) Descriptor() ([]byte, []int) {
	return file_pkg_dto_dto_proto_rawDescGZIP(), []int{9}
}

func (x *Consistent) GetNumVNodes() int64 {
	if x != nil {
		return x.NumVNodes
	}
	return 0
}

func (x *Consistent) GetNodes() []uint64 {
	if x != nil {
		return x.Nodes
	}
	return nil
}

var File_pkg_dto_dto_proto protoreflect.FileDescriptor

var file_pkg_dto_dto_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x6b, 0x67, 0x2f, 0x64, 0x74, 0x6f, 0x2f, 0x64, 0x74, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x03, 0x64, 0x74, 0x6f, 0x22, 0x36, 0x0a, 0x0a, 0x50, 0x75, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x22, 0x47, 0x0a, 0x0b, 0x50, 0x75, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x38, 0x0a, 0x0d, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x53, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0d, 0x73, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x1e, 0x0a, 0x0a, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x8a, 0x01, 0x0a, 0x0b, 0x47, 0x65,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x12, 0x38, 0x0a, 0x0d, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x53,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0d, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x29, 0x0a, 0x08, 0x66,
	0x6f, 0x75, 0x6e, 0x64, 0x4b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e,
	0x64, 0x74, 0x6f, 0x2e, 0x46, 0x6f, 0x75, 0x6e, 0x64, 0x4b, 0x65, 0x79, 0x52, 0x08, 0x66, 0x6f,
	0x75, 0x6e, 0x64, 0x4b, 0x65, 0x79, 0x22, 0x6d, 0x0a, 0x0d, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x12, 0x32, 0x0a, 0x0b, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x63, 0x6c, 0x6f, 0x63, 0x6b,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x56, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x43, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0b, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x44, 0x0a, 0x0e, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a, 0x0b, 0x76, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x64,
	0x74, 0x6f, 0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x43, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0b,
	0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x55, 0x0a, 0x0d, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x32,
	0x0a, 0x0b, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x43, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0b, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x63, 0x6c, 0x6f,
	0x63, 0x6b, 0x22, 0x5c, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x32, 0x0a, 0x0b,
	0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x10, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x43, 0x6c,
	0x6f, 0x63, 0x6b, 0x52, 0x0b, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x63, 0x6c, 0x6f, 0x63, 0x6b,
	0x22, 0x9c, 0x01, 0x0a, 0x0b, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x43, 0x6c, 0x6f, 0x63, 0x6b,
	0x12, 0x34, 0x0a, 0x06, 0x76, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x43, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x56, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06,
	0x76, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1c, 0x0a, 0x09, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e,
	0x65, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x6d, 0x61, 0x63, 0x68, 0x69,
	0x6e, 0x65, 0x49, 0x44, 0x1a, 0x39, 0x0a, 0x0b, 0x56, 0x63, 0x6c, 0x6f, 0x63, 0x6b, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x44, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x73, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x1c, 0x0a,
	0x09, 0x6e, 0x75, 0x6d, 0x56, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x6e, 0x75, 0x6d, 0x56, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x05, 0x6e,
	0x6f, 0x64, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x04, 0x42, 0x02, 0x10, 0x01, 0x52, 0x05,
	0x6e, 0x6f, 0x64, 0x65, 0x73, 0x2a, 0x37, 0x0a, 0x0d, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x13, 0x0a, 0x0f, 0x50, 0x41, 0x52, 0x54, 0x49, 0x41,
	0x4c, 0x5f, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00, 0x12, 0x11, 0x0a, 0x0d, 0x46,
	0x55, 0x4c, 0x4c, 0x59, 0x5f, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x01, 0x2a, 0x2c,
	0x0a, 0x08, 0x46, 0x6f, 0x75, 0x6e, 0x64, 0x4b, 0x65, 0x79, 0x12, 0x11, 0x0a, 0x0d, 0x4b, 0x45,
	0x59, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10, 0x00, 0x12, 0x0d, 0x0a,
	0x09, 0x4b, 0x45, 0x59, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10, 0x01, 0x32, 0x67, 0x0a, 0x0d,
	0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x2a, 0x0a,
	0x03, 0x50, 0x75, 0x74, 0x12, 0x0f, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x50, 0x75, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x50, 0x75, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x2a, 0x0a, 0x03, 0x47, 0x65, 0x74,
	0x12, 0x0f, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x10, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x32, 0x81, 0x01, 0x0a, 0x15, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x12,
	0x33, 0x0a, 0x06, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x12, 0x12, 0x2e, 0x64, 0x74, 0x6f, 0x2e,
	0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e,
	0x64, 0x74, 0x6f, 0x2e, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x33, 0x0a, 0x06, 0x47, 0x65, 0x74, 0x52, 0x65, 0x70, 0x12, 0x12,
	0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x13, 0x2e, 0x64, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x70, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x65, 0x79, 0x75, 0x68, 0x61, 0x6e, 0x67,
	0x30, 0x2f, 0x44, 0x53, 0x43, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x64, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_dto_dto_proto_rawDescOnce sync.Once
	file_pkg_dto_dto_proto_rawDescData = file_pkg_dto_dto_proto_rawDesc
)

func file_pkg_dto_dto_proto_rawDescGZIP() []byte {
	file_pkg_dto_dto_proto_rawDescOnce.Do(func() {
		file_pkg_dto_dto_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_dto_dto_proto_rawDescData)
	})
	return file_pkg_dto_dto_proto_rawDescData
}

var file_pkg_dto_dto_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_pkg_dto_dto_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_pkg_dto_dto_proto_goTypes = []interface{}{
	(SuccessStatus)(0),     // 0: dto.SuccessStatus
	(FoundKey)(0),          // 1: dto.FoundKey
	(*PutRequest)(nil),     // 2: dto.PutRequest
	(*PutResponse)(nil),    // 3: dto.PutResponse
	(*GetRequest)(nil),     // 4: dto.GetRequest
	(*GetResponse)(nil),    // 5: dto.GetResponse
	(*PutRepRequest)(nil),  // 6: dto.PutRepRequest
	(*PutRepResponse)(nil), // 7: dto.PutRepResponse
	(*GetRepRequest)(nil),  // 8: dto.GetRepRequest
	(*GetRepResponse)(nil), // 9: dto.GetRepResponse
	(*VectorClock)(nil),    // 10: dto.VectorClock
	(*Consistent)(nil),     // 11: dto.Consistent
	nil,                    // 12: dto.VectorClock.VclockEntry
}
var file_pkg_dto_dto_proto_depIdxs = []int32{
	0,  // 0: dto.PutResponse.successStatus:type_name -> dto.SuccessStatus
	0,  // 1: dto.GetResponse.successStatus:type_name -> dto.SuccessStatus
	1,  // 2: dto.GetResponse.foundKey:type_name -> dto.FoundKey
	10, // 3: dto.PutRepRequest.vectorclock:type_name -> dto.VectorClock
	10, // 4: dto.PutRepResponse.vectorclock:type_name -> dto.VectorClock
	10, // 5: dto.GetRepRequest.vectorclock:type_name -> dto.VectorClock
	10, // 6: dto.GetRepResponse.vectorclock:type_name -> dto.VectorClock
	12, // 7: dto.VectorClock.vclock:type_name -> dto.VectorClock.VclockEntry
	2,  // 8: dto.KeyValueStore.Put:input_type -> dto.PutRequest
	4,  // 9: dto.KeyValueStore.Get:input_type -> dto.GetRequest
	6,  // 10: dto.KeyValueStoreInternal.PutRep:input_type -> dto.PutRepRequest
	8,  // 11: dto.KeyValueStoreInternal.GetRep:input_type -> dto.GetRepRequest
	3,  // 12: dto.KeyValueStore.Put:output_type -> dto.PutResponse
	5,  // 13: dto.KeyValueStore.Get:output_type -> dto.GetResponse
	7,  // 14: dto.KeyValueStoreInternal.PutRep:output_type -> dto.PutRepResponse
	9,  // 15: dto.KeyValueStoreInternal.GetRep:output_type -> dto.GetRepResponse
	12, // [12:16] is the sub-list for method output_type
	8,  // [8:12] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_pkg_dto_dto_proto_init() }
func file_pkg_dto_dto_proto_init() {
	if File_pkg_dto_dto_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_dto_dto_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutRepRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutRepResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRepRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRepResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VectorClock); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_dto_dto_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Consistent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_dto_dto_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_pkg_dto_dto_proto_goTypes,
		DependencyIndexes: file_pkg_dto_dto_proto_depIdxs,
		EnumInfos:         file_pkg_dto_dto_proto_enumTypes,
		MessageInfos:      file_pkg_dto_dto_proto_msgTypes,
	}.Build()
	File_pkg_dto_dto_proto = out.File
	file_pkg_dto_dto_proto_rawDesc = nil
	file_pkg_dto_dto_proto_goTypes = nil
	file_pkg_dto_dto_proto_depIdxs = nil
}
