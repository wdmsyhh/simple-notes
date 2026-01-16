package v1

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	pbstore "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListTags 获取标签列表，支持可选的分页
func (s *APIV1Service) ListTags(ctx context.Context, req *apiv1.ListTagsRequest) (*apiv1.ListTagsResponse, error) {
	// 调用存储层获取标签列表
	tags, total, err := s.Store.ListTags(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取标签列表失败: %w", err)
	}

	// 创建响应对象
	response := &apiv1.ListTagsResponse{
		Tags:  tags,
		Total: int32(total),
	}

	return response, nil
}

// GetTag 根据ID获取标签
func (s *APIV1Service) GetTag(ctx context.Context, req *apiv1.GetTagRequest) (*pbstore.Tag, error) {
	// 从资源名称中提取标签ID
	tagID, err := extractIDFromResourceName(req.GetName(), "tags")
	if err != nil {
		return nil, err
	}

	// 调用存储层获取标签
	tag, err := s.Store.GetTag(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("获取标签失败: %w", err)
	}

	// 设置资源名称
	tag.Name = fmt.Sprintf("tags/%d", tag.Id)

	return tag, nil
}

// CreateTag 创建新标签
func (s *APIV1Service) CreateTag(ctx context.Context, req *apiv1.CreateTagRequest) (*pbstore.Tag, error) {
	// 从请求中获取标签信息
	tag := req.GetTag()
	if tag == nil {
		return nil, fmt.Errorf("标签信息不能为空")
	}

	// 验证标签数据
	if tag.NameText == "" {
		return nil, fmt.Errorf("标签名称不能为空")
	}

	// slug 字段不需要，使用主键 id 即可，设置为空字符串
	tag.Slug = ""

	// 调用存储层创建标签
	createdTag, err := s.Store.CreateTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("创建标签失败: %w", err)
	}

	// 设置资源名称
	createdTag.Name = fmt.Sprintf("tags/%d", createdTag.Id)

	return createdTag, nil
}

// UpdateTag 更新现有标签
func (s *APIV1Service) UpdateTag(ctx context.Context, req *apiv1.UpdateTagRequest) (*pbstore.Tag, error) {
	// 从请求中获取标签信息
	tag := req.GetTag()
	if tag == nil {
		return nil, fmt.Errorf("标签信息不能为空")
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
		return nil, fmt.Errorf("标签ID不能为空")
	}

	// 调用存储层更新标签
	updatedTag, err := s.Store.UpdateTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("更新标签失败: %w", err)
	}

	// 设置资源名称
	updatedTag.Name = fmt.Sprintf("tags/%d", updatedTag.Id)

	return updatedTag, nil
}

// DeleteTag 删除标签
func (s *APIV1Service) DeleteTag(ctx context.Context, req *apiv1.DeleteTagRequest) (*emptypb.Empty, error) {
	// 从资源名称中提取标签ID
	tagID, err := extractIDFromResourceName(req.GetName(), "tags")
	if err != nil {
		return nil, err
	}

	// 调用存储层删除标签
	if err := s.Store.DeleteTag(ctx, tagID); err != nil {
		// 检查是否是"标签下有文章"的错误
		if strings.Contains(err.Error(), "tag has") {
			return nil, fmt.Errorf("该标签下还有文章，无法删除")
		}
		return nil, fmt.Errorf("删除标签失败: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// GetTagBySlug 根据标识获取标签
func (s *APIV1Service) GetTagBySlug(ctx context.Context, req *apiv1.GetTagBySlugRequest) (*pbstore.Tag, error) {
	slug := req.GetSlug()
	if slug == "" {
		return nil, fmt.Errorf("标签标识不能为空")
	}

	// 调用存储层根据标识获取标签
	tag, err := s.Store.GetTagBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("根据标识获取标签失败: %w", err)
	}

	// 设置资源名称
	tag.Name = fmt.Sprintf("tags/%d", tag.Id)

	return tag, nil
}

// TagService 的 Connect 处理器实现

// ListTagsHandler 实现 ListTags 方法的 Connect 处理器
func (s *APIV1Service) ListTagsHandler(ctx context.Context, req *connect.Request[apiv1.ListTagsRequest]) (*connect.Response[apiv1.ListTagsResponse], error) {
	// 调用gRPC方法
	nativeResp, err := s.ListTags(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// GetTagHandler 实现 GetTag 方法的 Connect 处理器
func (s *APIV1Service) GetTagHandler(ctx context.Context, req *connect.Request[apiv1.GetTagRequest]) (*connect.Response[pbstore.Tag], error) {
	// 调用gRPC方法
	nativeResp, err := s.GetTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// CreateTagHandler 实现 CreateTag 方法的 Connect 处理器
func (s *APIV1Service) CreateTagHandler(ctx context.Context, req *connect.Request[apiv1.CreateTagRequest]) (*connect.Response[pbstore.Tag], error) {
	// 调用gRPC方法
	nativeResp, err := s.CreateTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// UpdateTagHandler 实现 UpdateTag 方法的 Connect 处理器
func (s *APIV1Service) UpdateTagHandler(ctx context.Context, req *connect.Request[apiv1.UpdateTagRequest]) (*connect.Response[pbstore.Tag], error) {
	// 调用gRPC方法
	nativeResp, err := s.UpdateTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// DeleteTagHandler 实现 DeleteTag 方法的 Connect 处理器
func (s *APIV1Service) DeleteTagHandler(ctx context.Context, req *connect.Request[apiv1.DeleteTagRequest]) (*connect.Response[emptypb.Empty], error) {
	// 调用gRPC方法
	nativeResp, err := s.DeleteTag(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// GetTagBySlugHandler 实现 GetTagBySlug 方法的 Connect 处理器
func (s *APIV1Service) GetTagBySlugHandler(ctx context.Context, req *connect.Request[apiv1.GetTagBySlugRequest]) (*connect.Response[pbstore.Tag], error) {
	// 调用gRPC方法
	nativeResp, err := s.GetTagBySlug(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}
