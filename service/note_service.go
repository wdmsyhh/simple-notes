package service

import (
	"context"
	"fmt"
	"math"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	pbstore "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"github.com/wdmsyhh/simple-notes/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

// NoteService 处理笔记相关操作的服务
// 实现了 NoteServiceServer 接口
// 封装了笔记的增删改查等业务逻辑

type NoteService struct {
	// store - 数据存储实例，用于数据库操作
	store *store.Store
	// UnimplementedNoteServiceServer - 未实现的 NoteService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedNoteServiceServer
}

// NewNoteService 创建一个新的 NoteService 实例
// 参数：
//
//	store - 数据存储实例
//
// 返回：
//
//	*NoteService - 创建的笔记服务实例
func NewNoteService(store *store.Store) *NoteService {
	return &NoteService{
		store: store,
	}
}

// ListNotes 获取笔记列表，支持分页
// 参数：
//
//	ctx - 上下文
//	req - 笔记列表请求，包含分页和过滤条件
//
// 返回：
//
//	*apiv1.ListNotesResponse - 笔记列表响应，包含笔记列表、分页信息
//	error - 错误信息
func (s *NoteService) ListNotes(ctx context.Context, req *apiv1.ListNotesRequest) (*apiv1.ListNotesResponse, error) {
	// 如果未提供，设置默认值
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}

	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100 // 限制页面大小为 100
	}

	// 创建存储层请求
	storeReq := &store.ListNotesRequest{
		Page:               page,
		PageSize:           pageSize,
		CategoryID:         req.GetCategoryId(),
		TagID:              req.GetTagId(),
		Search:             req.GetSearch(),
		SortBy:             req.GetSortBy(),
		SortDesc:           req.GetSortDesc(),
		IncludeUnpublished: false, // API 仅返回已发布的笔记
	}

	// 调用存储层
	notes, total, err := s.store.ListNotes(ctx, storeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	// 计算总页数
	totalPages := int32(math.Ceil(float64(total) / float64(pageSize)))

	// 创建响应
	response := &apiv1.ListNotesResponse{
		Notes:      notes,
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return response, nil
}

// GetNote 根据ID获取单个笔记
// 参数：
//
//	ctx - 上下文
//	req - 获取笔记请求，包含资源名称
//
// 返回：
//
//	*pbstore.Note - 笔记信息
//	error - 错误信息
func (s *NoteService) GetNote(ctx context.Context, req *apiv1.GetNoteRequest) (*pbstore.Note, error) {
	// 从资源名称中提取笔记ID
	// 资源名称格式："notes/{note}"
	noteID, err := extractIDFromResourceName(req.GetName(), "notes")
	if err != nil {
		return nil, err
	}

	// 调用存储层
	note, err := s.store.GetNote(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// 设置资源名称
	note.Name = fmt.Sprintf("notes/%d", note.Id)

	return note, nil
}

// CreateNote 创建新笔记
// 参数：
//
//	ctx - 上下文
//	req - 创建笔记请求，包含笔记信息
//
// 返回：
//
//	*pbstore.Note - 创建的笔记信息
//	error - 错误信息
func (s *NoteService) CreateNote(ctx context.Context, req *apiv1.CreateNoteRequest) (*pbstore.Note, error) {
	// 从请求中获取笔记
	note := req.GetNote()
	if note == nil {
		return nil, fmt.Errorf("note is required")
	}

	// 验证笔记数据
	if note.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if note.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	// 调用存储层
	createdNote, err := s.store.CreateNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// 设置资源名称
	createdNote.Name = fmt.Sprintf("notes/%d", createdNote.Id)

	return createdNote, nil
}

// UpdateNote 更新现有笔记
// 参数：
//
//	ctx - 上下文
//	req - 更新笔记请求，包含笔记信息
//
// 返回：
//
//	*pbstore.Note - 更新后的笔记信息
//	error - 错误信息
func (s *NoteService) UpdateNote(ctx context.Context, req *apiv1.UpdateNoteRequest) (*pbstore.Note, error) {
	// 从请求中获取笔记
	note := req.GetNote()
	if note == nil {
		return nil, fmt.Errorf("note is required")
	}

	// 从资源名称中提取笔记ID
	if note.Name != "" {
		noteID, err := extractIDFromResourceName(note.Name, "notes")
		if err != nil {
			return nil, err
		}
		note.Id = noteID
	}

	if note.Id == 0 {
		return nil, fmt.Errorf("note ID is required")
	}

	// 调用存储层
	updatedNote, err := s.store.UpdateNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// 设置资源名称
	updatedNote.Name = fmt.Sprintf("notes/%d", updatedNote.Id)

	return updatedNote, nil
}

// DeleteNote 删除笔记
// 参数：
//
//	ctx - 上下文
//	req - 删除笔记请求，包含资源名称
//
// 返回：
//
//	*emptypb.Empty - 空响应
//	error - 错误信息
func (s *NoteService) DeleteNote(ctx context.Context, req *apiv1.DeleteNoteRequest) (*emptypb.Empty, error) {
	// 从资源名称中提取笔记ID
	noteID, err := extractIDFromResourceName(req.GetName(), "notes")
	if err != nil {
		return nil, err
	}

	// 调用存储层
	if err := s.store.DeleteNote(ctx, noteID); err != nil {
		return nil, fmt.Errorf("failed to delete note: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// GetNoteBySlug 已移除，请使用 GetNote 通过 ID 获取笔记

// extractIDFromResourceName 从资源名称中提取数字ID
// 资源名称格式: "{type}/{id}"
// 参数：
//
//	name - 资源名称
//	expectedType - 预期的资源类型
//
// 返回：
//
//	int64 - 提取的ID
//	error - 错误信息
func extractIDFromResourceName(name, expectedType string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("resource name is required")
	}

	// 分割资源名称
	parts := []string{}
	current := ""
	for _, r := range name {
		if r == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid resource name format: expected %s/{id}", expectedType)
	}

	resourceType := parts[0]
	if resourceType != expectedType {
		return 0, fmt.Errorf("invalid resource type: expected %s, got %s", expectedType, resourceType)
	}

	// 将ID解析为int64
	var id int64
	if _, err := fmt.Sscanf(parts[1], "%d", &id); err != nil {
		return 0, fmt.Errorf("invalid resource ID: %w", err)
	}

	if id <= 0 {
		return 0, fmt.Errorf("invalid resource ID: must be positive")
	}

	return id, nil
}
