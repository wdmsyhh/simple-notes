package service

import (
	"context"
	"fmt"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	pbstore "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"github.com/wdmsyhh/simple-notes/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TagService 处理标签相关操作的服务
// 实现了 TagServiceServer 接口
// 封装了标签的增删改查等业务逻辑

type TagService struct {
	// store - 数据存储实例，用于数据库操作
	store *store.Store
	// UnimplementedTagServiceServer - 未实现的 TagService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedTagServiceServer
}

// NewTagService 创建一个新的 TagService 实例
// 参数：
//
//	store - 数据存储实例
//
// 返回：
//
//	*TagService - 创建的标签服务实例
func NewTagService(store *store.Store) *TagService {
	return &TagService{
		store: store,
	}
}

// ListTags 获取标签列表，支持可选的分页
// 参数：
//
//	ctx - 上下文
//	req - 标签列表请求，包含分页和过滤条件
//
// 返回：
//
//	*apiv1.ListTagsResponse - 标签列表响应，包含标签列表和总数
//	error - 错误信息
func (s *TagService) ListTags(ctx context.Context, req *apiv1.ListTagsRequest) (*apiv1.ListTagsResponse, error) {
	// 调用存储层
	tags, total, err := s.store.ListTags(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	// 创建响应
	response := &apiv1.ListTagsResponse{
		Tags:  tags,
		Total: int32(total),
	}

	return response, nil
}

// GetTag 根据ID获取标签
// 参数：
//
//	ctx - 上下文
//	req - 获取标签请求，包含资源名称
//
// 返回：
//
//	*pbstore.Tag - 标签信息
//	error - 错误信息
func (s *TagService) GetTag(ctx context.Context, req *apiv1.GetTagRequest) (*pbstore.Tag, error) {
	// 从资源名称中提取标签ID
	tagID, err := extractIDFromResourceName(req.GetName(), "tags")
	if err != nil {
		return nil, err
	}

	// 调用存储层
	tag, err := s.store.GetTag(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// 设置资源名称
	tag.Name = fmt.Sprintf("tags/%d", tag.Id)

	return tag, nil
}

// CreateTag 创建新标签
// 参数：
//
//	ctx - 上下文
//	req - 创建标签请求，包含标签信息
//
// 返回：
//
//	*pbstore.Tag - 创建的标签信息
//	error - 错误信息
func (s *TagService) CreateTag(ctx context.Context, req *apiv1.CreateTagRequest) (*pbstore.Tag, error) {
	// 从请求中获取标签
	tag := req.GetTag()
	if tag == nil {
		return nil, fmt.Errorf("tag is required")
	}

	// 验证标签数据
	if tag.NameText == "" {
		return nil, fmt.Errorf("name is required")
	}

	if tag.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	// 调用存储层
	createdTag, err := s.store.CreateTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	// 设置资源名称
	createdTag.Name = fmt.Sprintf("tags/%d", createdTag.Id)

	return createdTag, nil
}

// UpdateTag 更新现有标签
// 参数：
//
//	ctx - 上下文
//	req - 更新标签请求，包含标签信息
//
// 返回：
//
//	*pbstore.Tag - 更新后的标签信息
//	error - 错误信息
func (s *TagService) UpdateTag(ctx context.Context, req *apiv1.UpdateTagRequest) (*pbstore.Tag, error) {
	// 从请求中获取标签
	tag := req.GetTag()
	if tag == nil {
		return nil, fmt.Errorf("tag is required")
	}

	// 从资源名称中提取标签ID
	if tag.Name != "" {
		tagID, err := extractIDFromResourceName(tag.Name, "tags")
		if err != nil {
			return nil, err
		}
		tag.Id = tagID
	}

	if tag.Id == 0 {
		return nil, fmt.Errorf("tag ID is required")
	}

	// 调用存储层
	updatedTag, err := s.store.UpdateTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	// 设置资源名称
	updatedTag.Name = fmt.Sprintf("tags/%d", updatedTag.Id)

	return updatedTag, nil
}

// DeleteTag 删除标签
// 参数：
//
//	ctx - 上下文
//	req - 删除标签请求，包含资源名称
//
// 返回：
//
//	*emptypb.Empty - 空响应
//	error - 错误信息
func (s *TagService) DeleteTag(ctx context.Context, req *apiv1.DeleteTagRequest) (*emptypb.Empty, error) {
	// 从资源名称中提取标签ID
	tagID, err := extractIDFromResourceName(req.GetName(), "tags")
	if err != nil {
		return nil, err
	}

	// 调用存储层
	if err := s.store.DeleteTag(ctx, tagID); err != nil {
		return nil, fmt.Errorf("failed to delete tag: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// GetTagBySlug 通过slug获取标签
// 参数：
//
//	ctx - 上下文
//	req - 通过slug获取标签请求
//
// 返回：
//
//	*pbstore.Tag - 标签信息
//	error - 错误信息
func (s *TagService) GetTagBySlug(ctx context.Context, req *apiv1.GetTagBySlugRequest) (*pbstore.Tag, error) {
	slug := req.GetSlug()
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	// 调用存储层
	tag, err := s.store.GetTagBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}

	// 设置资源名称
	tag.Name = fmt.Sprintf("tags/%d", tag.Id)

	return tag, nil
}
