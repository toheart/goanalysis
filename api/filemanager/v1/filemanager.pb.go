// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.31.1
// source: filemanager/v1/filemanager.proto

package v1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 文件类型枚举
type FileType int32

const (
	FileType_FILE_TYPE_UNSPECIFIED FileType = 0
	FileType_FILE_TYPE_RUNTIME     FileType = 1 // 运行时文件
	FileType_FILE_TYPE_STATIC      FileType = 2 // 静态分析文件
)

// Enum value maps for FileType.
var (
	FileType_name = map[int32]string{
		0: "FILE_TYPE_UNSPECIFIED",
		1: "FILE_TYPE_RUNTIME",
		2: "FILE_TYPE_STATIC",
	}
	FileType_value = map[string]int32{
		"FILE_TYPE_UNSPECIFIED": 0,
		"FILE_TYPE_RUNTIME":     1,
		"FILE_TYPE_STATIC":      2,
	}
)

func (x FileType) Enum() *FileType {
	p := new(FileType)
	*p = x
	return p
}

func (x FileType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FileType) Descriptor() protoreflect.EnumDescriptor {
	return file_filemanager_v1_filemanager_proto_enumTypes[0].Descriptor()
}

func (FileType) Type() protoreflect.EnumType {
	return &file_filemanager_v1_filemanager_proto_enumTypes[0]
}

func (x FileType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FileType.Descriptor instead.
func (FileType) EnumDescriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{0}
}

// 获取文件信息请求
type GetFileInfoRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件ID
	Id            int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFileInfoRequest) Reset() {
	*x = GetFileInfoRequest{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFileInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFileInfoRequest) ProtoMessage() {}

func (x *GetFileInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFileInfoRequest.ProtoReflect.Descriptor instead.
func (*GetFileInfoRequest) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{0}
}

func (x *GetFileInfoRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

// 获取文件信息响应
type GetFileInfoReply struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件信息
	FileInfo      *FileInfo `protobuf:"bytes,1,opt,name=fileInfo,proto3" json:"fileInfo,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFileInfoReply) Reset() {
	*x = GetFileInfoReply{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFileInfoReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFileInfoReply) ProtoMessage() {}

func (x *GetFileInfoReply) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFileInfoReply.ProtoReflect.Descriptor instead.
func (*GetFileInfoReply) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{1}
}

func (x *GetFileInfoReply) GetFileInfo() *FileInfo {
	if x != nil {
		return x.FileInfo
	}
	return nil
}

// 获取文件列表请求
type ListFilesRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件类型
	FileType FileType `protobuf:"varint,1,opt,name=fileType,proto3,enum=filemanager.v1.FileType" json:"fileType,omitempty"`
	// 分页限制
	Limit int32 `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	// 分页偏移
	Offset        int32 `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListFilesRequest) Reset() {
	*x = ListFilesRequest{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListFilesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFilesRequest) ProtoMessage() {}

func (x *ListFilesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFilesRequest.ProtoReflect.Descriptor instead.
func (*ListFilesRequest) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{2}
}

func (x *ListFilesRequest) GetFileType() FileType {
	if x != nil {
		return x.FileType
	}
	return FileType_FILE_TYPE_UNSPECIFIED
}

func (x *ListFilesRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *ListFilesRequest) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

// 获取文件列表响应
type ListFilesReply struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件信息列表
	Files []*FileInfo `protobuf:"bytes,1,rep,name=files,proto3" json:"files,omitempty"`
	// 总数
	Total         int64 `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListFilesReply) Reset() {
	*x = ListFilesReply{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListFilesReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFilesReply) ProtoMessage() {}

func (x *ListFilesReply) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFilesReply.ProtoReflect.Descriptor instead.
func (*ListFilesReply) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{3}
}

func (x *ListFilesReply) GetFiles() []*FileInfo {
	if x != nil {
		return x.Files
	}
	return nil
}

func (x *ListFilesReply) GetTotal() int64 {
	if x != nil {
		return x.Total
	}
	return 0
}

// 删除文件请求
type DeleteFileRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件ID
	Id            int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteFileRequest) Reset() {
	*x = DeleteFileRequest{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteFileRequest) ProtoMessage() {}

func (x *DeleteFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteFileRequest.ProtoReflect.Descriptor instead.
func (*DeleteFileRequest) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteFileRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

// 删除文件响应
type DeleteFileReply struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 是否成功
	Success       bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteFileReply) Reset() {
	*x = DeleteFileReply{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteFileReply) ProtoMessage() {}

func (x *DeleteFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteFileReply.ProtoReflect.Descriptor instead.
func (*DeleteFileReply) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{5}
}

func (x *DeleteFileReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

// 下载文件请求
type DownloadFileRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件ID
	Id            int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DownloadFileRequest) Reset() {
	*x = DownloadFileRequest{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DownloadFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadFileRequest) ProtoMessage() {}

func (x *DownloadFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadFileRequest.ProtoReflect.Descriptor instead.
func (*DownloadFileRequest) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{6}
}

func (x *DownloadFileRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

// 下载文件响应
type DownloadFileReply struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件内容（Base64编码）
	FileContent string `protobuf:"bytes,1,opt,name=fileContent,proto3" json:"fileContent,omitempty"`
	// 文件名
	FileName string `protobuf:"bytes,2,opt,name=fileName,proto3" json:"fileName,omitempty"`
	// 内容类型
	ContentType   string `protobuf:"bytes,3,opt,name=contentType,proto3" json:"contentType,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DownloadFileReply) Reset() {
	*x = DownloadFileReply{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DownloadFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadFileReply) ProtoMessage() {}

func (x *DownloadFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadFileReply.ProtoReflect.Descriptor instead.
func (*DownloadFileReply) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{7}
}

func (x *DownloadFileReply) GetFileContent() string {
	if x != nil {
		return x.FileContent
	}
	return ""
}

func (x *DownloadFileReply) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *DownloadFileReply) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

// 文件信息
type FileInfo struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 文件ID
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// 文件名
	FileName string `protobuf:"bytes,2,opt,name=fileName,proto3" json:"fileName,omitempty"`
	// 文件类型
	FileType FileType `protobuf:"varint,3,opt,name=fileType,proto3,enum=filemanager.v1.FileType" json:"fileType,omitempty"`
	// 文件大小（字节）
	FileSize int64 `protobuf:"varint,4,opt,name=fileSize,proto3" json:"fileSize,omitempty"`
	// 内容类型
	ContentType string `protobuf:"bytes,5,opt,name=contentType,proto3" json:"contentType,omitempty"`
	// 上传时间（ISO 8601格式）
	UploadTime string `protobuf:"bytes,6,opt,name=uploadTime,proto3" json:"uploadTime,omitempty"`
	// 文件描述
	Description string `protobuf:"bytes,7,opt,name=description,proto3" json:"description,omitempty"`
	// 文件路径
	FilePath      string `protobuf:"bytes,8,opt,name=filePath,proto3" json:"filePath,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FileInfo) Reset() {
	*x = FileInfo{}
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FileInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileInfo) ProtoMessage() {}

func (x *FileInfo) ProtoReflect() protoreflect.Message {
	mi := &file_filemanager_v1_filemanager_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileInfo.ProtoReflect.Descriptor instead.
func (*FileInfo) Descriptor() ([]byte, []int) {
	return file_filemanager_v1_filemanager_proto_rawDescGZIP(), []int{8}
}

func (x *FileInfo) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *FileInfo) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *FileInfo) GetFileType() FileType {
	if x != nil {
		return x.FileType
	}
	return FileType_FILE_TYPE_UNSPECIFIED
}

func (x *FileInfo) GetFileSize() int64 {
	if x != nil {
		return x.FileSize
	}
	return 0
}

func (x *FileInfo) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *FileInfo) GetUploadTime() string {
	if x != nil {
		return x.UploadTime
	}
	return ""
}

func (x *FileInfo) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FileInfo) GetFilePath() string {
	if x != nil {
		return x.FilePath
	}
	return ""
}

var File_filemanager_v1_filemanager_proto protoreflect.FileDescriptor

const file_filemanager_v1_filemanager_proto_rawDesc = "" +
	"\n" +
	" filemanager/v1/filemanager.proto\x12\x0efilemanager.v1\x1a\x1cgoogle/api/annotations.proto\"$\n" +
	"\x12GetFileInfoRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\"H\n" +
	"\x10GetFileInfoReply\x124\n" +
	"\bfileInfo\x18\x01 \x01(\v2\x18.filemanager.v1.FileInfoR\bfileInfo\"v\n" +
	"\x10ListFilesRequest\x124\n" +
	"\bfileType\x18\x01 \x01(\x0e2\x18.filemanager.v1.FileTypeR\bfileType\x12\x14\n" +
	"\x05limit\x18\x02 \x01(\x05R\x05limit\x12\x16\n" +
	"\x06offset\x18\x03 \x01(\x05R\x06offset\"V\n" +
	"\x0eListFilesReply\x12.\n" +
	"\x05files\x18\x01 \x03(\v2\x18.filemanager.v1.FileInfoR\x05files\x12\x14\n" +
	"\x05total\x18\x02 \x01(\x03R\x05total\"#\n" +
	"\x11DeleteFileRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\"+\n" +
	"\x0fDeleteFileReply\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\"%\n" +
	"\x13DownloadFileRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\"s\n" +
	"\x11DownloadFileReply\x12 \n" +
	"\vfileContent\x18\x01 \x01(\tR\vfileContent\x12\x1a\n" +
	"\bfileName\x18\x02 \x01(\tR\bfileName\x12 \n" +
	"\vcontentType\x18\x03 \x01(\tR\vcontentType\"\x88\x02\n" +
	"\bFileInfo\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x1a\n" +
	"\bfileName\x18\x02 \x01(\tR\bfileName\x124\n" +
	"\bfileType\x18\x03 \x01(\x0e2\x18.filemanager.v1.FileTypeR\bfileType\x12\x1a\n" +
	"\bfileSize\x18\x04 \x01(\x03R\bfileSize\x12 \n" +
	"\vcontentType\x18\x05 \x01(\tR\vcontentType\x12\x1e\n" +
	"\n" +
	"uploadTime\x18\x06 \x01(\tR\n" +
	"uploadTime\x12 \n" +
	"\vdescription\x18\a \x01(\tR\vdescription\x12\x1a\n" +
	"\bfilePath\x18\b \x01(\tR\bfilePath*R\n" +
	"\bFileType\x12\x19\n" +
	"\x15FILE_TYPE_UNSPECIFIED\x10\x00\x12\x15\n" +
	"\x11FILE_TYPE_RUNTIME\x10\x01\x12\x14\n" +
	"\x10FILE_TYPE_STATIC\x10\x022\xc3\x03\n" +
	"\vFileManager\x12l\n" +
	"\vGetFileInfo\x12\".filemanager.v1.GetFileInfoRequest\x1a .filemanager.v1.GetFileInfoReply\"\x17\x82\xd3\xe4\x93\x02\x11\x12\x0f/api/files/{id}\x12a\n" +
	"\tListFiles\x12 .filemanager.v1.ListFilesRequest\x1a\x1e.filemanager.v1.ListFilesReply\"\x12\x82\xd3\xe4\x93\x02\f\x12\n" +
	"/api/files\x12i\n" +
	"\n" +
	"DeleteFile\x12!.filemanager.v1.DeleteFileRequest\x1a\x1f.filemanager.v1.DeleteFileReply\"\x17\x82\xd3\xe4\x93\x02\x11*\x0f/api/files/{id}\x12x\n" +
	"\fDownloadFile\x12#.filemanager.v1.DownloadFileRequest\x1a!.filemanager.v1.DownloadFileReply\" \x82\xd3\xe4\x93\x02\x1a\x12\x18/api/files/{id}/downloadB5Z3github.com/toheart/goanalysis/api/filemanager/v1;v1b\x06proto3"

var (
	file_filemanager_v1_filemanager_proto_rawDescOnce sync.Once
	file_filemanager_v1_filemanager_proto_rawDescData []byte
)

func file_filemanager_v1_filemanager_proto_rawDescGZIP() []byte {
	file_filemanager_v1_filemanager_proto_rawDescOnce.Do(func() {
		file_filemanager_v1_filemanager_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_filemanager_v1_filemanager_proto_rawDesc), len(file_filemanager_v1_filemanager_proto_rawDesc)))
	})
	return file_filemanager_v1_filemanager_proto_rawDescData
}

var file_filemanager_v1_filemanager_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_filemanager_v1_filemanager_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_filemanager_v1_filemanager_proto_goTypes = []any{
	(FileType)(0),               // 0: filemanager.v1.FileType
	(*GetFileInfoRequest)(nil),  // 1: filemanager.v1.GetFileInfoRequest
	(*GetFileInfoReply)(nil),    // 2: filemanager.v1.GetFileInfoReply
	(*ListFilesRequest)(nil),    // 3: filemanager.v1.ListFilesRequest
	(*ListFilesReply)(nil),      // 4: filemanager.v1.ListFilesReply
	(*DeleteFileRequest)(nil),   // 5: filemanager.v1.DeleteFileRequest
	(*DeleteFileReply)(nil),     // 6: filemanager.v1.DeleteFileReply
	(*DownloadFileRequest)(nil), // 7: filemanager.v1.DownloadFileRequest
	(*DownloadFileReply)(nil),   // 8: filemanager.v1.DownloadFileReply
	(*FileInfo)(nil),            // 9: filemanager.v1.FileInfo
}
var file_filemanager_v1_filemanager_proto_depIdxs = []int32{
	9, // 0: filemanager.v1.GetFileInfoReply.fileInfo:type_name -> filemanager.v1.FileInfo
	0, // 1: filemanager.v1.ListFilesRequest.fileType:type_name -> filemanager.v1.FileType
	9, // 2: filemanager.v1.ListFilesReply.files:type_name -> filemanager.v1.FileInfo
	0, // 3: filemanager.v1.FileInfo.fileType:type_name -> filemanager.v1.FileType
	1, // 4: filemanager.v1.FileManager.GetFileInfo:input_type -> filemanager.v1.GetFileInfoRequest
	3, // 5: filemanager.v1.FileManager.ListFiles:input_type -> filemanager.v1.ListFilesRequest
	5, // 6: filemanager.v1.FileManager.DeleteFile:input_type -> filemanager.v1.DeleteFileRequest
	7, // 7: filemanager.v1.FileManager.DownloadFile:input_type -> filemanager.v1.DownloadFileRequest
	2, // 8: filemanager.v1.FileManager.GetFileInfo:output_type -> filemanager.v1.GetFileInfoReply
	4, // 9: filemanager.v1.FileManager.ListFiles:output_type -> filemanager.v1.ListFilesReply
	6, // 10: filemanager.v1.FileManager.DeleteFile:output_type -> filemanager.v1.DeleteFileReply
	8, // 11: filemanager.v1.FileManager.DownloadFile:output_type -> filemanager.v1.DownloadFileReply
	8, // [8:12] is the sub-list for method output_type
	4, // [4:8] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_filemanager_v1_filemanager_proto_init() }
func file_filemanager_v1_filemanager_proto_init() {
	if File_filemanager_v1_filemanager_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_filemanager_v1_filemanager_proto_rawDesc), len(file_filemanager_v1_filemanager_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_filemanager_v1_filemanager_proto_goTypes,
		DependencyIndexes: file_filemanager_v1_filemanager_proto_depIdxs,
		EnumInfos:         file_filemanager_v1_filemanager_proto_enumTypes,
		MessageInfos:      file_filemanager_v1_filemanager_proto_msgTypes,
	}.Build()
	File_filemanager_v1_filemanager_proto = out.File
	file_filemanager_v1_filemanager_proto_goTypes = nil
	file_filemanager_v1_filemanager_proto_depIdxs = nil
}
