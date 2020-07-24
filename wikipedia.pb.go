// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.3
// source: wikipedia.proto

package wikipedia

import (
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type Revision struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   int32                `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Ts   *timestamp.Timestamp `protobuf:"bytes,2,opt,name=ts,proto3" json:"ts,omitempty"`
	Text string               `protobuf:"bytes,3,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *Revision) Reset() {
	*x = Revision{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wikipedia_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Revision) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Revision) ProtoMessage() {}

func (x *Revision) ProtoReflect() protoreflect.Message {
	mi := &file_wikipedia_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Revision.ProtoReflect.Descriptor instead.
func (*Revision) Descriptor() ([]byte, []int) {
	return file_wikipedia_proto_rawDescGZIP(), []int{0}
}

func (x *Revision) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Revision) GetTs() *timestamp.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *Revision) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

type Link struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TargetTitle string `protobuf:"bytes,1,opt,name=target_title,json=targetTitle,proto3" json:"target_title,omitempty"`
}

func (x *Link) Reset() {
	*x = Link{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wikipedia_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Link) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Link) ProtoMessage() {}

func (x *Link) ProtoReflect() protoreflect.Message {
	mi := &file_wikipedia_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Link.ProtoReflect.Descriptor instead.
func (*Link) Descriptor() ([]byte, []int) {
	return file_wikipedia_proto_rawDescGZIP(), []int{1}
}

func (x *Link) GetTargetTitle() string {
	if x != nil {
		return x.TargetTitle
	}
	return ""
}

type LinkedPage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PageTitle string  `protobuf:"bytes,1,opt,name=page_title,json=pageTitle,proto3" json:"page_title,omitempty"`
	PageId    int32   `protobuf:"varint,2,opt,name=page_id,json=pageId,proto3" json:"page_id,omitempty"`
	Links     []*Link `protobuf:"bytes,3,rep,name=links,proto3" json:"links,omitempty"`
}

func (x *LinkedPage) Reset() {
	*x = LinkedPage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wikipedia_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LinkedPage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LinkedPage) ProtoMessage() {}

func (x *LinkedPage) ProtoReflect() protoreflect.Message {
	mi := &file_wikipedia_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LinkedPage.ProtoReflect.Descriptor instead.
func (*LinkedPage) Descriptor() ([]byte, []int) {
	return file_wikipedia_proto_rawDescGZIP(), []int{2}
}

func (x *LinkedPage) GetPageTitle() string {
	if x != nil {
		return x.PageTitle
	}
	return ""
}

func (x *LinkedPage) GetPageId() int32 {
	if x != nil {
		return x.PageId
	}
	return 0
}

func (x *LinkedPage) GetLinks() []*Link {
	if x != nil {
		return x.Links
	}
	return nil
}

type Page struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Title         string      `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Id            int32       `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
	Namespace     uint32      `protobuf:"varint,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
	RedirectTitle string      `protobuf:"bytes,4,opt,name=redirect_title,json=redirectTitle,proto3" json:"redirect_title,omitempty"`
	Revisions     []*Revision `protobuf:"bytes,5,rep,name=revisions,proto3" json:"revisions,omitempty"`
}

func (x *Page) Reset() {
	*x = Page{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wikipedia_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Page) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Page) ProtoMessage() {}

func (x *Page) ProtoReflect() protoreflect.Message {
	mi := &file_wikipedia_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Page.ProtoReflect.Descriptor instead.
func (*Page) Descriptor() ([]byte, []int) {
	return file_wikipedia_proto_rawDescGZIP(), []int{3}
}

func (x *Page) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Page) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Page) GetNamespace() uint32 {
	if x != nil {
		return x.Namespace
	}
	return 0
}

func (x *Page) GetRedirectTitle() string {
	if x != nil {
		return x.RedirectTitle
	}
	return ""
}

func (x *Page) GetRevisions() []*Revision {
	if x != nil {
		return x.Revisions
	}
	return nil
}

var File_wikipedia_proto protoreflect.FileDescriptor

var file_wikipedia_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x77, 0x69, 0x6b, 0x69, 0x72, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x1c, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x73, 0x65, 0x62, 0x6e,
	0x79, 0x62, 0x65, 0x72, 0x67, 0x2e, 0x77, 0x69, 0x6b, 0x69, 0x72, 0x65, 0x6c, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5a,
	0x0a, 0x08, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x12, 0x2a, 0x0a, 0x02, 0x74, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x02, 0x74, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x22, 0x29, 0x0a, 0x04, 0x4c, 0x69,
	0x6e, 0x6b, 0x12, 0x21, 0x0a, 0x0c, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74,
	0x54, 0x69, 0x74, 0x6c, 0x65, 0x22, 0x7e, 0x0a, 0x0a, 0x4c, 0x69, 0x6e, 0x6b, 0x65, 0x64, 0x50,
	0x61, 0x67, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x69, 0x74, 0x6c,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54, 0x69, 0x74,
	0x6c, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x06, 0x70, 0x61, 0x67, 0x65, 0x49, 0x64, 0x12, 0x38, 0x0a, 0x05, 0x6c,
	0x69, 0x6e, 0x6b, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x6f, 0x6d,
	0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x73, 0x65, 0x62, 0x6e, 0x79, 0x62, 0x65, 0x72,
	0x67, 0x2e, 0x77, 0x69, 0x6b, 0x69, 0x72, 0x65, 0x6c, 0x2e, 0x4c, 0x69, 0x6e, 0x6b, 0x52, 0x05,
	0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x22, 0xb7, 0x01, 0x0a, 0x04, 0x50, 0x61, 0x67, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74,
	0x69, 0x74, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61,
	0x63, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x5f, 0x74,
	0x69, 0x74, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x72, 0x65, 0x64, 0x69,
	0x72, 0x65, 0x63, 0x74, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x44, 0x0a, 0x09, 0x72, 0x65, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x63,
	0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x73, 0x65, 0x62, 0x6e, 0x79, 0x62,
	0x65, 0x72, 0x67, 0x2e, 0x77, 0x69, 0x6b, 0x69, 0x72, 0x65, 0x6c, 0x2e, 0x52, 0x65, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x72, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x42,
	0x1e, 0x5a, 0x1c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x65,
	0x62, 0x6e, 0x79, 0x62, 0x65, 0x72, 0x67, 0x2f, 0x77, 0x69, 0x6b, 0x69, 0x72, 0x65, 0x6c, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_wikipedia_proto_rawDescOnce sync.Once
	file_wikipedia_proto_rawDescData = file_wikipedia_proto_rawDesc
)

func file_wikipedia_proto_rawDescGZIP() []byte {
	file_wikipedia_proto_rawDescOnce.Do(func() {
		file_wikipedia_proto_rawDescData = protoimpl.X.CompressGZIP(file_wikipedia_proto_rawDescData)
	})
	return file_wikipedia_proto_rawDescData
}

var file_wikipedia_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_wikipedia_proto_goTypes = []interface{}{
	(*Revision)(nil),            // 0: com.github.sebnyberg.wikipedia.Revision
	(*Link)(nil),                // 1: com.github.sebnyberg.wikipedia.Link
	(*LinkedPage)(nil),          // 2: com.github.sebnyberg.wikipedia.LinkedPage
	(*Page)(nil),                // 3: com.github.sebnyberg.wikipedia.Page
	(*timestamp.Timestamp)(nil), // 4: google.protobuf.Timestamp
}
var file_wikipedia_proto_depIdxs = []int32{
	4, // 0: com.github.sebnyberg.wikipedia.Revision.ts:type_name -> google.protobuf.Timestamp
	1, // 1: com.github.sebnyberg.wikipedia.LinkedPage.links:type_name -> com.github.sebnyberg.wikipedia.Link
	0, // 2: com.github.sebnyberg.wikipedia.Page.revisions:type_name -> com.github.sebnyberg.wikipedia.Revision
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_wikipedia_proto_init() }
func file_wikipedia_proto_init() {
	if File_wikipedia_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_wikipedia_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Revision); i {
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
		file_wikipedia_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Link); i {
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
		file_wikipedia_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LinkedPage); i {
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
		file_wikipedia_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Page); i {
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
			RawDescriptor: file_wikipedia_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_wikipedia_proto_goTypes,
		DependencyIndexes: file_wikipedia_proto_depIdxs,
		MessageInfos:      file_wikipedia_proto_msgTypes,
	}.Build()
	File_wikipedia_proto = out.File
	file_wikipedia_proto_rawDesc = nil
	file_wikipedia_proto_goTypes = nil
	file_wikipedia_proto_depIdxs = nil
}