package v1

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	pbstore "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"github.com/wdmsyhh/simple-notes/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListNotes 获取笔记列表，支持分页
func (s *APIV1Service) ListNotes(ctx context.Context, req *apiv1.ListNotesRequest) (*apiv1.ListNotesResponse, error) {
	// 如果未提供，设置默认分页参数
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}

	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100 // 将分页大小限制为100
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
		IncludeUnpublished: false, // API只返回已发布的帖子
	}

	// 调用存储层获取笔记列表
	notes, totalCount, err := s.Store.ListNotes(ctx, storeReq)
	if err != nil {
		return nil, fmt.Errorf("获取笔记列表失败: %w", err)
	}

	// 获取当前用户
	currentUser, _ := s.fetchCurrentUser(ctx)

	// 根据可见性过滤笔记并设置资源名称
	visibleNotes := []*pbstore.Note{}
	for _, note := range notes {
		// 检查笔记可见性
		if s.isNoteVisibleToUser(note, currentUser) {
			// 设置资源名称
			note.Name = fmt.Sprintf("notes/%d", note.Id)
			visibleNotes = append(visibleNotes, note)
		}
	}

	// 计算总页数（使用数据库返回的总记录数）
	// 注意：totalCount 是数据库查询的总记录数，可能包含不可见的笔记
	// 但对于首页和分类页面，通常都是公开笔记，所以可以使用 totalCount
	totalPages := int32(math.Ceil(float64(totalCount) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	// 创建响应对象
	response := &apiv1.ListNotesResponse{
		Notes:      visibleNotes,
		Total:      int32(totalCount), // 使用数据库返回的总记录数，而不是当前页的数量
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return response, nil
}

// isNoteVisibleToUser 检查用户是否有权访问指定笔记
func (s *APIV1Service) isNoteVisibleToUser(note *pbstore.Note, user *store.User) bool {
	// 公共笔记对所有人可见
	if note.Visibility == pbstore.NoteVisibility_NOTE_VISIBILITY_PUBLIC {
		return true
	}

	// 私有笔记仅对作者和管理员可见
	if user == nil {
		return false
	}

	// 检查用户是否为笔记作者
	authorID, _ := strconv.ParseUint(note.AuthorId, 10, 32)
	if user.ID == uint(authorID) {
		return true
	}

	// 检查用户是否为管理员或主机
	return user.Role == store.RoleAdmin || user.Role == store.RoleHost
}

// GetNote 根据ID获取笔记
func (s *APIV1Service) GetNote(ctx context.Context, req *apiv1.GetNoteRequest) (*pbstore.Note, error) {
	// 从资源名称中提取笔记ID
	// 资源名称格式: "notes/{note}"
	noteID, err := extractIDFromResourceName(req.GetName(), "notes")
	if err != nil {
		return nil, err
	}

	// 调用存储层获取笔记
	note, err := s.Store.GetNote(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("获取笔记失败: %w", err)
	}

	// 检查可见性权限
	currentUser, _ := s.fetchCurrentUser(ctx)
	if !s.isNoteVisibleToUser(note, currentUser) {
		return nil, fmt.Errorf("没有权限访问该笔记")
	}

	// 设置资源名称
	note.Name = fmt.Sprintf("notes/%d", note.Id)

	return note, nil
}

// CreateNote 创建新笔记
func (s *APIV1Service) CreateNote(ctx context.Context, req *apiv1.CreateNoteRequest) (*pbstore.Note, error) {
	// 检查认证
	currentUser, err := s.fetchCurrentUser(ctx)
	if err != nil || currentUser == nil {
		return nil, fmt.Errorf("authentication required")
	}

	// 从请求中获取笔记信息
	note := req.GetNote()
	if note == nil {
		return nil, fmt.Errorf("笔记信息不能为空")
	}

	// 验证笔记数据：标题、描述、内容都是必填
	if note.Title == "" {
		return nil, fmt.Errorf("标题不能为空")
	}
	if note.Summary == "" {
		return nil, fmt.Errorf("描述不能为空")
	}
	if note.Content == "" {
		return nil, fmt.Errorf("内容不能为空")
	}

	// Slug 是可选的，如果为空就不设置（使用 ID 访问）
	// 如果提供了 slug，可以保留用于 SEO，但不强制

	// 设置作者ID
	note.AuthorId = fmt.Sprintf("%d", currentUser.ID)

	// 如果设置为发布，设置发布时间
	if note.Published {
		note.PublishedAt = time.Now().Unix()
	}

	// 调用存储层创建笔记
	createdNote, err := s.Store.CreateNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("创建笔记失败: %w", err)
	}

	// 设置资源名称
	createdNote.Name = fmt.Sprintf("notes/%d", createdNote.Id)

	return createdNote, nil
}

// UpdateNote 更新现有笔记
func (s *APIV1Service) UpdateNote(ctx context.Context, req *apiv1.UpdateNoteRequest) (*pbstore.Note, error) {
	// 检查认证
	currentUser, err := s.fetchCurrentUser(ctx)
	if err != nil || currentUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	// 从请求中获取笔记信息
	note := req.GetNote()
	if note == nil {
		return nil, fmt.Errorf("笔记信息不能为空")
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
		return nil, fmt.Errorf("笔记ID不能为空")
	}

	// 获取现有笔记以检查权限
	existingNote, err := s.Store.GetNote(ctx, note.Id)
	if err != nil {
		return nil, fmt.Errorf("获取笔记失败: %w", err)
	}

	// 检查权限：只有作者或管理员可以更新
	authorID, _ := strconv.ParseUint(existingNote.AuthorId, 10, 32)
	if currentUser.ID != uint(authorID) && currentUser.Role != store.RoleAdmin && currentUser.Role != store.RoleHost {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied: only author or admin can update note")
	}

	// 验证笔记数据：标题、描述、内容都是必填
	if note.Title == "" {
		return nil, fmt.Errorf("标题不能为空")
	}
	if note.Summary == "" {
		return nil, fmt.Errorf("描述不能为空")
	}
	if note.Content == "" {
		return nil, fmt.Errorf("内容不能为空")
	}

	// 如果从未发布变为发布，设置发布时间
	if !existingNote.Published && note.Published {
		note.PublishedAt = time.Now().Unix()
	}

	// 调用存储层更新笔记
	updatedNote, err := s.Store.UpdateNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("更新笔记失败: %w", err)
	}

	// 设置资源名称
	updatedNote.Name = fmt.Sprintf("notes/%d", updatedNote.Id)

	return updatedNote, nil
}

// DeleteNote 删除笔记
func (s *APIV1Service) DeleteNote(ctx context.Context, req *apiv1.DeleteNoteRequest) (*emptypb.Empty, error) {
	// 检查认证
	currentUser, err := s.fetchCurrentUser(ctx)
	if err != nil || currentUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	// 从资源名称中提取笔记ID
	noteID, err := extractIDFromResourceName(req.GetName(), "notes")
	if err != nil {
		return nil, err
	}

	// 获取现有笔记以检查权限
	existingNote, err := s.Store.GetNote(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("获取笔记失败: %w", err)
	}

	// 检查权限：只有作者或管理员可以删除
	authorID, _ := strconv.ParseUint(existingNote.AuthorId, 10, 32)
	if currentUser.ID != uint(authorID) && currentUser.Role != store.RoleAdmin && currentUser.Role != store.RoleHost {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied: only author or admin can delete note")
	}

	// 调用存储层删除笔记
	if err := s.Store.DeleteNote(ctx, noteID); err != nil {
		return nil, fmt.Errorf("删除笔记失败: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// GetNoteBySlug 已移除，请使用 GetNote 通过 ID 获取笔记
