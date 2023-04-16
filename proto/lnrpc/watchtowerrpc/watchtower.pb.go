// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v4.22.3
// source: proto/lnrpc/watchtowerrpc/watchtower.proto

package watchtowerrpc

import (
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

type GetInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetInfoRequest) Reset() {
	*x = GetInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInfoRequest) ProtoMessage() {}

func (x *GetInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInfoRequest.ProtoReflect.Descriptor instead.
func (*GetInfoRequest) Descriptor() ([]byte, []int) {
	return file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescGZIP(), []int{0}
}

type GetInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The public key of the watchtower.
	Pubkey []byte `protobuf:"bytes,1,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
	// The listening addresses of the watchtower.
	Listeners []string `protobuf:"bytes,2,rep,name=listeners,proto3" json:"listeners,omitempty"`
	// The URIs of the watchtower.
	Uris []string `protobuf:"bytes,3,rep,name=uris,proto3" json:"uris,omitempty"`
}

func (x *GetInfoResponse) Reset() {
	*x = GetInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInfoResponse) ProtoMessage() {}

func (x *GetInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInfoResponse.ProtoReflect.Descriptor instead.
func (*GetInfoResponse) Descriptor() ([]byte, []int) {
	return file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescGZIP(), []int{1}
}

func (x *GetInfoResponse) GetPubkey() []byte {
	if x != nil {
		return x.Pubkey
	}
	return nil
}

func (x *GetInfoResponse) GetListeners() []string {
	if x != nil {
		return x.Listeners
	}
	return nil
}

func (x *GetInfoResponse) GetUris() []string {
	if x != nil {
		return x.Uris
	}
	return nil
}

var File_proto_lnrpc_watchtowerrpc_watchtower_proto protoreflect.FileDescriptor

var file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x77, 0x61,
	0x74, 0x63, 0x68, 0x74, 0x6f, 0x77, 0x65, 0x72, 0x72, 0x70, 0x63, 0x2f, 0x77, 0x61, 0x74, 0x63,
	0x68, 0x74, 0x6f, 0x77, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x77, 0x61,
	0x74, 0x63, 0x68, 0x74, 0x6f, 0x77, 0x65, 0x72, 0x72, 0x70, 0x63, 0x22, 0x10, 0x0a, 0x0e, 0x47,
	0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x5b, 0x0a,
	0x0f, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x6c, 0x69, 0x73, 0x74,
	0x65, 0x6e, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x6c, 0x69, 0x73,
	0x74, 0x65, 0x6e, 0x65, 0x72, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x72, 0x69, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x75, 0x72, 0x69, 0x73, 0x32, 0x56, 0x0a, 0x0a, 0x57, 0x61,
	0x74, 0x63, 0x68, 0x74, 0x6f, 0x77, 0x65, 0x72, 0x12, 0x48, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x1d, 0x2e, 0x77, 0x61, 0x74, 0x63, 0x68, 0x74, 0x6f, 0x77, 0x65, 0x72,
	0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x77, 0x61, 0x74, 0x63, 0x68, 0x74, 0x6f, 0x77, 0x65, 0x72, 0x72,
	0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6c, 0x69, 0x67, 0x68, 0x74, 0x6e, 0x69, 0x6e, 0x67, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72,
	0x6b, 0x2f, 0x6c, 0x6e, 0x64, 0x2f, 0x6c, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x77, 0x61, 0x74, 0x63,
	0x68, 0x74, 0x6f, 0x77, 0x65, 0x72, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescOnce sync.Once
	file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescData = file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDesc
)

func file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescGZIP() []byte {
	file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescOnce.Do(func() {
		file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescData)
	})
	return file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDescData
}

var file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_lnrpc_watchtowerrpc_watchtower_proto_goTypes = []interface{}{
	(*GetInfoRequest)(nil),  // 0: watchtowerrpc.GetInfoRequest
	(*GetInfoResponse)(nil), // 1: watchtowerrpc.GetInfoResponse
}
var file_proto_lnrpc_watchtowerrpc_watchtower_proto_depIdxs = []int32{
	0, // 0: watchtowerrpc.Watchtower.GetInfo:input_type -> watchtowerrpc.GetInfoRequest
	1, // 1: watchtowerrpc.Watchtower.GetInfo:output_type -> watchtowerrpc.GetInfoResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_lnrpc_watchtowerrpc_watchtower_proto_init() }
func file_proto_lnrpc_watchtowerrpc_watchtower_proto_init() {
	if File_proto_lnrpc_watchtowerrpc_watchtower_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetInfoRequest); i {
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
		file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetInfoResponse); i {
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
			RawDescriptor: file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_lnrpc_watchtowerrpc_watchtower_proto_goTypes,
		DependencyIndexes: file_proto_lnrpc_watchtowerrpc_watchtower_proto_depIdxs,
		MessageInfos:      file_proto_lnrpc_watchtowerrpc_watchtower_proto_msgTypes,
	}.Build()
	File_proto_lnrpc_watchtowerrpc_watchtower_proto = out.File
	file_proto_lnrpc_watchtowerrpc_watchtower_proto_rawDesc = nil
	file_proto_lnrpc_watchtowerrpc_watchtower_proto_goTypes = nil
	file_proto_lnrpc_watchtowerrpc_watchtower_proto_depIdxs = nil
}
