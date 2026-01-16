package v1

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	pbstore "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 此文件包含所有 Connect 服务处理器方法的实现
// 每个方法都委托给底层的 gRPC 服务实现，
// 在 Connect 和 gRPC 请求/响应类型之间进行转换

// NoteService

// ListNotes 获取笔记列表的 Connect 处理器
func (s *ConnectServiceHandler) ListNotes(ctx context.Context, req *connect.Request[apiv1.ListNotesRequest]) (*connect.Response[apiv1.ListNotesResponse], error) {
	resp, err := s.APIV1Service.ListNotes(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetNote 获取单个笔记的 Connect 处理器
func (s *ConnectServiceHandler) GetNote(ctx context.Context, req *connect.Request[apiv1.GetNoteRequest]) (*connect.Response[pbstore.Note], error) {
	resp, err := s.APIV1Service.GetNote(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// CreateNote 创建笔记的 Connect 处理器
func (s *ConnectServiceHandler) CreateNote(ctx context.Context, req *connect.Request[apiv1.CreateNoteRequest]) (*connect.Response[pbstore.Note], error) {
	resp, err := s.APIV1Service.CreateNote(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// UpdateNote 更新笔记的 Connect 处理器
func (s *ConnectServiceHandler) UpdateNote(ctx context.Context, req *connect.Request[apiv1.UpdateNoteRequest]) (*connect.Response[pbstore.Note], error) {
	resp, err := s.APIV1Service.UpdateNote(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// DeleteNote 删除笔记的 Connect 处理器
func (s *ConnectServiceHandler) DeleteNote(ctx context.Context, req *connect.Request[apiv1.DeleteNoteRequest]) (*connect.Response[emptypb.Empty], error) {
	resp, err := s.APIV1Service.DeleteNote(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetNoteBySlug 已废弃，请使用 GetNote 通过 ID 获取笔记
func (s *ConnectServiceHandler) GetNoteBySlug(ctx context.Context, req *connect.Request[apiv1.GetNoteBySlugRequest]) (*connect.Response[pbstore.Note], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("GetNoteBySlug 已废弃，请使用 GetNote 通过 ID 获取笔记"))
}

// CategoryService

// ListCategories 获取分类列表的 Connect 处理器
func (s *ConnectServiceHandler) ListCategories(ctx context.Context, req *connect.Request[apiv1.ListCategoriesRequest]) (*connect.Response[apiv1.ListCategoriesResponse], error) {
	resp, err := s.APIV1Service.ListCategories(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetCategory 获取单个分类的 Connect 处理器
func (s *ConnectServiceHandler) GetCategory(ctx context.Context, req *connect.Request[apiv1.GetCategoryRequest]) (*connect.Response[pbstore.Category], error) {
	resp, err := s.APIV1Service.GetCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// CreateCategory 创建分类的 Connect 处理器
func (s *ConnectServiceHandler) CreateCategory(ctx context.Context, req *connect.Request[apiv1.CreateCategoryRequest]) (*connect.Response[pbstore.Category], error) {
	resp, err := s.APIV1Service.CreateCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// UpdateCategory 更新分类的 Connect 处理器
func (s *ConnectServiceHandler) UpdateCategory(ctx context.Context, req *connect.Request[apiv1.UpdateCategoryRequest]) (*connect.Response[pbstore.Category], error) {
	resp, err := s.APIV1Service.UpdateCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// DeleteCategory 删除分类的 Connect 处理器
func (s *ConnectServiceHandler) DeleteCategory(ctx context.Context, req *connect.Request[apiv1.DeleteCategoryRequest]) (*connect.Response[emptypb.Empty], error) {
	resp, err := s.APIV1Service.DeleteCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetCategoryBySlug 通过 slug 获取分类的 Connect 处理器
func (s *ConnectServiceHandler) GetCategoryBySlug(ctx context.Context, req *connect.Request[apiv1.GetCategoryBySlugRequest]) (*connect.Response[pbstore.Category], error) {
	resp, err := s.APIV1Service.GetCategoryBySlug(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// TagService

// ListTags 获取标签列表的 Connect 处理器
func (s *ConnectServiceHandler) ListTags(ctx context.Context, req *connect.Request[apiv1.ListTagsRequest]) (*connect.Response[apiv1.ListTagsResponse], error) {
	resp, err := s.APIV1Service.ListTags(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetTag 获取单个标签的 Connect 处理器
func (s *ConnectServiceHandler) GetTag(ctx context.Context, req *connect.Request[apiv1.GetTagRequest]) (*connect.Response[pbstore.Tag], error) {
	resp, err := s.APIV1Service.GetTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// CreateTag 创建标签的 Connect 处理器
func (s *ConnectServiceHandler) CreateTag(ctx context.Context, req *connect.Request[apiv1.CreateTagRequest]) (*connect.Response[pbstore.Tag], error) {
	resp, err := s.APIV1Service.CreateTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// UpdateTag 更新标签的 Connect 处理器
func (s *ConnectServiceHandler) UpdateTag(ctx context.Context, req *connect.Request[apiv1.UpdateTagRequest]) (*connect.Response[pbstore.Tag], error) {
	resp, err := s.APIV1Service.UpdateTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// DeleteTag 删除标签的 Connect 处理器
func (s *ConnectServiceHandler) DeleteTag(ctx context.Context, req *connect.Request[apiv1.DeleteTagRequest]) (*connect.Response[emptypb.Empty], error) {
	resp, err := s.APIV1Service.DeleteTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetTagBySlug 通过 slug 获取标签的 Connect 处理器
func (s *ConnectServiceHandler) GetTagBySlug(ctx context.Context, req *connect.Request[apiv1.GetTagBySlugRequest]) (*connect.Response[pbstore.Tag], error) {
	resp, err := s.APIV1Service.GetTagBySlug(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// UserService

// RegisterUser 注册新用户
func (s *ConnectServiceHandler) RegisterUser(ctx context.Context, req *connect.Request[apiv1.RegisterUserRequest]) (*connect.Response[pbstore.User], error) {
	resp, err := s.APIV1Service.RegisterUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// LoginUser 认证用户
func (s *ConnectServiceHandler) LoginUser(ctx context.Context, req *connect.Request[apiv1.LoginUserRequest]) (*connect.Response[apiv1.LoginUserResponse], error) {
	resp, err := s.APIV1Service.LoginUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetUser 根据ID检索用户
func (s *ConnectServiceHandler) GetUser(ctx context.Context, req *connect.Request[apiv1.GetUserRequest]) (*connect.Response[pbstore.User], error) {
	resp, err := s.APIV1Service.GetUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetCurrentUser 检索当前已认证的用户
func (s *ConnectServiceHandler) GetCurrentUser(ctx context.Context, req *connect.Request[apiv1.GetCurrentUserRequest]) (*connect.Response[pbstore.User], error) {
	resp, err := s.APIV1Service.GetCurrentUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// UpdateUser 更新现有用户
func (s *ConnectServiceHandler) UpdateUser(ctx context.Context, req *connect.Request[apiv1.UpdateUserRequest]) (*connect.Response[pbstore.User], error) {
	resp, err := s.APIV1Service.UpdateUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// DeleteUser 删除用户
func (s *ConnectServiceHandler) DeleteUser(ctx context.Context, req *connect.Request[apiv1.DeleteUserRequest]) (*connect.Response[emptypb.Empty], error) {
	resp, err := s.APIV1Service.DeleteUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// ListUsers 检索用户列表
func (s *ConnectServiceHandler) ListUsers(ctx context.Context, req *connect.Request[apiv1.ListUsersRequest]) (*connect.Response[apiv1.ListUsersResponse], error) {
	resp, err := s.APIV1Service.ListUsers(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// AttachmentService 附件服务

// CreateAttachment 创建新附件
func (s *ConnectServiceHandler) CreateAttachment(ctx context.Context, req *connect.Request[apiv1.CreateAttachmentRequest]) (*connect.Response[apiv1.Attachment], error) {
	resp, err := s.APIV1Service.CreateAttachment(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// ListAttachments 列出附件
func (s *ConnectServiceHandler) ListAttachments(ctx context.Context, req *connect.Request[apiv1.ListAttachmentsRequest]) (*connect.Response[apiv1.ListAttachmentsResponse], error) {
	resp, err := s.APIV1Service.ListAttachments(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetAttachment 根据名称获取附件
func (s *ConnectServiceHandler) GetAttachment(ctx context.Context, req *connect.Request[apiv1.GetAttachmentRequest]) (*connect.Response[apiv1.Attachment], error) {
	resp, err := s.APIV1Service.GetAttachment(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// UpdateAttachment 更新附件
func (s *ConnectServiceHandler) UpdateAttachment(ctx context.Context, req *connect.Request[apiv1.UpdateAttachmentRequest]) (*connect.Response[apiv1.Attachment], error) {
	resp, err := s.APIV1Service.UpdateAttachment(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// DeleteAttachment 删除附件
func (s *ConnectServiceHandler) DeleteAttachment(ctx context.Context, req *connect.Request[apiv1.DeleteAttachmentRequest]) (*connect.Response[emptypb.Empty], error) {
	resp, err := s.APIV1Service.DeleteAttachment(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}
