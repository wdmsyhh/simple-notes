package v1

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	pbstore "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"github.com/wdmsyhh/simple-notes/store"
)

const (
	// MaxUploadBufferSizeBytes is the maximum upload buffer size (32 MiB)
	MaxUploadBufferSizeBytes = 32 << 20
	MebiByte                  = 1024 * 1024
)

var (
	// Valid MIME type pattern
	mimeTypeRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9!#$&\-\^_.]*/[a-zA-Z0-9][a-zA-Z0-9!#$&\-\^_.]*$`)
)

// validateFilename 验证文件名
// 此函数使用类似 memos 的更宽松的方法：
// - 防止路径遍历攻击
// - 防止文件名以空格或句点开头/结尾
// - 允许除路径分隔符外的大多数字符
// 参数：
//   filename - 文件名
// 返回：
//   bool - 文件名是否有效
func validateFilename(filename string) bool {
	// 检查长度
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}

	// 拒绝路径遍历尝试并确保不会创建额外的目录
	if !filepath.IsLocal(filename) || strings.ContainsAny(filename, "/\\") {
		return false
	}

	// 拒绝以空格或句点开头或结尾的文件名
	if strings.HasPrefix(filename, " ") || strings.HasSuffix(filename, " ") ||
		strings.HasPrefix(filename, ".") || strings.HasSuffix(filename, ".") {
		return false
	}

	return true
}

// isValidMimeType 验证 MIME 类型
// 参数：
//   mimeType - MIME 类型字符串
// 返回：
//   bool - MIME 类型是否有效
func isValidMimeType(mimeType string) bool {
	return mimeTypeRegex.MatchString(mimeType)
}

// CreateAttachment 创建新附件
// 参数：
//   ctx - 上下文
//   req - 创建附件请求
// 返回：
//   *apiv1.Attachment - 创建的附件对象
//   error - 错误信息
func (s *APIV1Service) CreateAttachment(ctx context.Context, req *apiv1.CreateAttachmentRequest) (*apiv1.Attachment, error) {
	// 检查认证
	currentUser, err := s.fetchCurrentUser(ctx)
	if err != nil || currentUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	// 验证必需字段
	if req.Attachment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "attachment is required")
	}
	if req.Attachment.Filename == "" {
		return nil, status.Errorf(codes.InvalidArgument, "filename is required")
	}
	if !validateFilename(req.Attachment.Filename) {
		return nil, status.Errorf(codes.InvalidArgument, "filename contains invalid characters")
	}
	if req.Attachment.Type == "" {
		return nil, status.Errorf(codes.InvalidArgument, "type is required")
	}
	if !isValidMimeType(req.Attachment.Type) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid MIME type format")
	}

	// 检查文件大小（对 []byte 使用 len，binary.Size 对切片不能正确工作）
	size := len(req.Attachment.Content)
	if size > MaxUploadBufferSizeBytes {
		return nil, status.Errorf(codes.InvalidArgument, "file size exceeds the limit (%d bytes)", MaxUploadBufferSizeBytes)
	}
	if size == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "file content cannot be empty")
	}

	// 在存储层创建附件
	storeAttachment := &pbstore.Attachment{
		Filename:  req.Attachment.Filename,
		Type:      req.Attachment.Type,
		Size:      int64(size),
		Content:   req.Attachment.Content,
		AuthorId:  fmt.Sprintf("%d", currentUser.ID),
		CreatedAt: 0, // 将由存储层设置
		UpdatedAt: 0, // 将由存储层设置
	}

	if req.Attachment.NoteId != "" {
		storeAttachment.NoteId = req.Attachment.NoteId
	}

	createdAttachment, err := s.Store.CreateAttachment(ctx, storeAttachment)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create attachment: %v", err)
	}

	// 转换为 API 响应（响应中不包含内容）
	return convertAttachmentToAPI(createdAttachment), nil
}

// ListAttachments 列出附件
// 参数：
//   ctx - 上下文
//   req - 列出附件请求
// 返回：
//   *apiv1.ListAttachmentsResponse - 附件列表响应
//   error - 错误信息
func (s *APIV1Service) ListAttachments(ctx context.Context, req *apiv1.ListAttachmentsRequest) (*apiv1.ListAttachmentsResponse, error) {
	// 获取当前用户（可选）
	currentUser, _ := s.fetchCurrentUser(ctx)

	// 设置默认页面大小
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	// Parse note_id if provided
	var noteID *int64
	if req.NoteId != "" {
		// Extract ID from resource name format: notes/{id}
		var id int64
		if _, err := fmt.Sscanf(req.NoteId, "notes/%d", &id); err == nil {
			noteID = &id
		}
	}

	// Permission check
	var authorID *uint
	if noteID != nil {
		// If listing attachments for a specific note
		note, err := s.Store.GetNote(ctx, *noteID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "note not found")
		}

		// Check if user can see this note
		if !s.isNoteVisibleToUser(note, currentUser) {
			return nil, status.Errorf(codes.PermissionDenied, "permission denied")
		}
		// If they can see the note, we don't need to filter by authorID
	} else {
		// If listing all attachments, authentication is required
		if currentUser == nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication required to list all attachments")
		}
		// Filter by authorID if not listing for a specific note
		id := currentUser.ID
		authorID = &id
	}

	attachments, err := s.Store.ListAttachments(ctx, noteID, authorID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list attachments: %v", err)
	}

	// 限制到页面大小
	if len(attachments) > pageSize {
		attachments = attachments[:pageSize]
	}

	// 转换为 API 响应
	apiAttachments := make([]*apiv1.Attachment, len(attachments))
	for i, att := range attachments {
		apiAttachments[i] = convertAttachmentToAPI(att)
	}

	return &apiv1.ListAttachmentsResponse{
		Attachments: apiAttachments,
		TotalSize:   int32(len(apiAttachments)),
	}, nil
}

// GetAttachment 根据名称获取附件
// 参数：
//   ctx - 上下文
//   req - 获取附件请求
// 返回：
//   *apiv1.Attachment - 附件对象
//   error - 错误信息
func (s *APIV1Service) GetAttachment(ctx context.Context, req *apiv1.GetAttachmentRequest) (*apiv1.Attachment, error) {
	// 获取当前用户（可选）
	currentUser, _ := s.fetchCurrentUser(ctx)

	// 从资源名称中提取ID
	var id int64
	if _, err := fmt.Sscanf(req.Name, "attachments/%d", &id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid attachment name: %s", req.Name)
	}

	attachment, err := s.Store.GetAttachment(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "attachment not found: %v", err)
	}

	// 权限检查
	allowed := false

	// 1. 检查它是否属于公开笔记
	if attachment.NoteId != "" {
		var noteID int64
		if _, err := fmt.Sscanf(attachment.NoteId, "notes/%d", &noteID); err == nil {
			note, err := s.Store.GetNote(ctx, noteID)
			if err == nil && s.isNoteVisibleToUser(note, currentUser) {
				allowed = true
			}
		}
	}

	// 2. 检查当前用户是否是作者
	if !allowed && currentUser != nil {
	var authorID uint
		if _, err := fmt.Sscanf(attachment.AuthorId, "%d", &authorID); err == nil {
			if currentUser.ID == authorID || currentUser.Role == store.RoleAdmin || currentUser.Role == store.RoleHost {
				allowed = true
			}
		}
	}

	if !allowed {
		if currentUser == nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication required")
		}
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	return convertAttachmentToAPI(attachment), nil
}

// UpdateAttachment 更新附件（例如，将其链接到笔记）
// 参数：
//   ctx - 上下文
//   req - 更新附件请求
// 返回：
//   *apiv1.Attachment - 更新后的附件对象
//   error - 错误信息
func (s *APIV1Service) UpdateAttachment(ctx context.Context, req *apiv1.UpdateAttachmentRequest) (*apiv1.Attachment, error) {
	// 检查认证
	currentUser, err := s.fetchCurrentUser(ctx)
	if err != nil || currentUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	// 验证必需字段
	if req.Attachment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "attachment is required")
	}
	if req.Attachment.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "attachment name is required")
	}

	// 从资源名称中提取ID
	var id int64
	if _, err := fmt.Sscanf(req.Attachment.Name, "attachments/%d", &id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid attachment name: %s", req.Attachment.Name)
	}

	// 获取附件以检查权限
	existingAttachment, err := s.Store.GetAttachment(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "attachment not found: %v", err)
	}

	// 检查权限：只有作者可以更新
	var authorID uint
	if _, err := fmt.Sscanf(existingAttachment.AuthorId, "%d", &authorID); err != nil {
		return nil, status.Errorf(codes.Internal, "invalid author ID format")
	}
	if currentUser.ID != authorID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	// 转换为存储层附件
	storeAttachment := &pbstore.Attachment{
		Name:   req.Attachment.Name,
		NoteId: req.Attachment.NoteId,
	}

	// 更新附件
	updatedAttachment, err := s.Store.UpdateAttachment(ctx, storeAttachment)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update attachment: %v", err)
	}

	return convertAttachmentToAPI(updatedAttachment), nil
}

// DeleteAttachment 删除附件
// 参数：
//   ctx - 上下文
//   req - 删除附件请求
// 返回：
//   *emptypb.Empty - 空响应
//   error - 错误信息
func (s *APIV1Service) DeleteAttachment(ctx context.Context, req *apiv1.DeleteAttachmentRequest) (*emptypb.Empty, error) {
	// 检查认证
	currentUser, err := s.fetchCurrentUser(ctx)
	if err != nil || currentUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	// 从资源名称中提取ID
	var id int64
	if _, err := fmt.Sscanf(req.Name, "attachments/%d", &id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid attachment name: %s", req.Name)
	}

	// 获取附件以检查权限
	attachment, err := s.Store.GetAttachment(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "attachment not found: %v", err)
	}

	// 检查权限：只有作者可以删除
	var authorID uint
	if _, err := fmt.Sscanf(attachment.AuthorId, "%d", &authorID); err != nil {
		return nil, status.Errorf(codes.Internal, "invalid author ID format")
	}
	if currentUser.ID != authorID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	if err := s.Store.DeleteAttachment(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete attachment: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// convertAttachmentToAPI 将 store.Attachment 转换为 api.v1.Attachment
// 参数：
//   storeAttachment - 存储层附件对象
// 返回：
//   *apiv1.Attachment - API 附件对象
func convertAttachmentToAPI(storeAttachment *pbstore.Attachment) *apiv1.Attachment {
	apiAttachment := &apiv1.Attachment{
		Name:     storeAttachment.Name,
		Filename: storeAttachment.Filename,
		Type:     storeAttachment.Type,
		Size:     storeAttachment.Size,
		NoteId:   storeAttachment.NoteId,
		// 注意：内容仅用于输入，不包含在响应中
		Content: nil,
	}

	// 设置创建时间
	if storeAttachment.CreatedAt > 0 {
		apiAttachment.CreateTime = timestamppb.New(timestampToTime(storeAttachment.CreatedAt))
	}

	return apiAttachment
}

// timestampToTime 将 unix 时间戳（秒）转换为 time.Time
// 参数：
//   ts - unix 时间戳（秒）
// 返回：
//   time.Time - 时间对象
func timestampToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

