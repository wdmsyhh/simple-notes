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

// ListCategories 获取分类列表，支持可选的过滤条件
// 参数：
//
//	ctx - 上下文
//	req - 分类列表请求，包含过滤条件
//
// 返回：
//
//	*apiv1.ListCategoriesResponse - 分类列表响应
//	error - 错误信息
func (s *APIV1Service) ListCategories(ctx context.Context, req *apiv1.ListCategoriesRequest) (*apiv1.ListCategoriesResponse, error) {
	// 调用存储层获取分类列表
	categories, err := s.Store.ListCategories(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取分类列表失败: %w", err)
	}

	// 创建响应对象
	response := &apiv1.ListCategoriesResponse{
		Categories: categories,
	}

	return response, nil
}

// GetCategory 根据ID获取分类
// 参数：
//
//	ctx - 上下文
//	req - 获取分类请求，包含资源名称
//
// 返回：
//
//	*pbstore.Category - 分类对象
//	error - 错误信息
func (s *APIV1Service) GetCategory(ctx context.Context, req *apiv1.GetCategoryRequest) (*pbstore.Category, error) {
	// 从资源名称中提取分类ID
	categoryID, err := extractIDFromResourceName(req.GetName(), "categories")
	if err != nil {
		return nil, err
	}

	// 调用存储层获取分类
	category, err := s.Store.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("获取分类失败: %w", err)
	}

	// 设置资源名称
	category.Name = fmt.Sprintf("categories/%d", category.Id)

	return category, nil
}

// CreateCategory 创建新分类
// 参数：
//
//	ctx - 上下文
//	req - 创建分类请求，包含分类信息
//
// 返回：
//
//	*pbstore.Category - 创建的分类对象
//	error - 错误信息
func (s *APIV1Service) CreateCategory(ctx context.Context, req *apiv1.CreateCategoryRequest) (*pbstore.Category, error) {
	// 从请求中获取分类信息
	category := req.GetCategory()
	if category == nil {
		return nil, fmt.Errorf("分类信息不能为空")
	}

	// 验证分类数据
	if category.NameText == "" {
		return nil, fmt.Errorf("分类名称不能为空")
	}

	// slug 字段不需要，使用主键 id 即可，设置为空字符串
	category.Slug = ""

	// 调用存储层创建分类
	createdCategory, err := s.Store.CreateCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}

	// 设置资源名称
	createdCategory.Name = fmt.Sprintf("categories/%d", createdCategory.Id)

	return createdCategory, nil
}

// UpdateCategory 更新现有分类
// 参数：
//
//	ctx - 上下文
//	req - 更新分类请求，包含分类信息
//
// 返回：
//
//	*pbstore.Category - 更新后的分类对象
//	error - 错误信息
func (s *APIV1Service) UpdateCategory(ctx context.Context, req *apiv1.UpdateCategoryRequest) (*pbstore.Category, error) {
	// 从请求中获取分类信息
	category := req.GetCategory()
	if category == nil {
		return nil, fmt.Errorf("分类信息不能为空")
	}

	// 从资源名称中提取分类ID
	if category.Name != "" {
		categoryID, err := extractIDFromResourceName(category.Name, "categories")
		if err != nil {
			return nil, err
		}
		category.Id = categoryID
	}

	if category.Id == 0 {
		return nil, fmt.Errorf("分类ID不能为空")
	}

	// 调用存储层更新分类
	updatedCategory, err := s.Store.UpdateCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("更新分类失败: %w", err)
	}

	// 设置资源名称
	updatedCategory.Name = fmt.Sprintf("categories/%d", updatedCategory.Id)

	return updatedCategory, nil
}

// DeleteCategory 删除分类
// 参数：
//
//	ctx - 上下文
//	req - 删除分类请求，包含资源名称
//
// 返回：
//
//	*emptypb.Empty - 空响应
//	error - 错误信息
func (s *APIV1Service) DeleteCategory(ctx context.Context, req *apiv1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	// 从资源名称中提取分类ID
	categoryID, err := extractIDFromResourceName(req.GetName(), "categories")
	if err != nil {
		return nil, err
	}

	// 调用存储层删除分类
	if err := s.Store.DeleteCategory(ctx, categoryID); err != nil {
		// 检查是否是"分类下有文章"的错误
		if strings.Contains(err.Error(), "category has") {
			return nil, fmt.Errorf("该分类下还有文章，无法删除")
		}
		return nil, fmt.Errorf("删除分类失败: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// GetCategoryBySlug 根据标识获取分类
// 参数：
//
//	ctx - 上下文
//	req - 根据标识获取分类请求，包含分类标识
//
// 返回：
//
//	*pbstore.Category - 分类对象
//	error - 错误信息
func (s *APIV1Service) GetCategoryBySlug(ctx context.Context, req *apiv1.GetCategoryBySlugRequest) (*pbstore.Category, error) {
	slug := req.GetSlug()
	if slug == "" {
		return nil, fmt.Errorf("分类标识不能为空")
	}

	// 调用存储层根据标识获取分类
	category, err := s.Store.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("根据标识获取分类失败: %w", err)
	}

	// 设置资源名称
	category.Name = fmt.Sprintf("categories/%d", category.Id)

	return category, nil
}

// generateSlugFromName 从分类名称生成 slug
// 将中文和特殊字符转换为 URL 友好的格式
func generateSlugFromName(name string) string {
	// 简单的 slug 生成：将名称转换为小写，替换空格为连字符
	slug := strings.ToLower(strings.TrimSpace(name))
	// 移除特殊字符，只保留字母、数字和连字符
	slug = strings.ReplaceAll(slug, " ", "-")
	// 移除所有非字母数字和连字符的字符
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	// 清理连续的连字符
	slug = strings.ReplaceAll(slug, "--", "-")
	slug = strings.Trim(slug, "-")
	// 如果处理后仍为空（比如纯中文），使用默认值
	if slug == "" {
		slug = "category"
	}
	return slug
}

// CategoryService 的 Connect 处理器实现

// ListCategoriesHandler 实现 ListCategories 方法的 Connect 处理器
// 参数：
//
//	ctx - 上下文
//	req - Connect请求，包含分类列表请求信息
//
// 返回：
//
//	*connect.Response[apiv1.ListCategoriesResponse] - Connect响应，包含分类列表响应
//	error - 错误信息
func (s *APIV1Service) ListCategoriesHandler(ctx context.Context, req *connect.Request[apiv1.ListCategoriesRequest]) (*connect.Response[apiv1.ListCategoriesResponse], error) {
	// 调用gRPC方法
	nativeResp, err := s.ListCategories(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// GetCategoryHandler 实现 GetCategory 方法的 Connect 处理器
// 参数：
//
//	ctx - 上下文
//	req - Connect请求，包含获取分类请求信息
//
// 返回：
//
//	*connect.Response[pbstore.Category] - Connect响应，包含分类对象
//	error - 错误信息
func (s *APIV1Service) GetCategoryHandler(ctx context.Context, req *connect.Request[apiv1.GetCategoryRequest]) (*connect.Response[pbstore.Category], error) {
	// 调用gRPC方法
	nativeResp, err := s.GetCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// CreateCategoryHandler 实现 CreateCategory 方法的 Connect 处理器
// 参数：
//
//	ctx - 上下文
//	req - Connect请求，包含创建分类请求信息
//
// 返回：
//
//	*connect.Response[pbstore.Category] - Connect响应，包含创建的分类对象
//	error - 错误信息
func (s *APIV1Service) CreateCategoryHandler(ctx context.Context, req *connect.Request[apiv1.CreateCategoryRequest]) (*connect.Response[pbstore.Category], error) {
	// 调用gRPC方法
	nativeResp, err := s.CreateCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// UpdateCategoryHandler 实现 UpdateCategory 方法的 Connect 处理器
// 参数：
//
//	ctx - 上下文
//	req - Connect请求，包含更新分类请求信息
//
// 返回：
//
//	*connect.Response[pbstore.Category] - Connect响应，包含更新后的分类对象
//	error - 错误信息
func (s *APIV1Service) UpdateCategoryHandler(ctx context.Context, req *connect.Request[apiv1.UpdateCategoryRequest]) (*connect.Response[pbstore.Category], error) {
	// 调用gRPC方法
	nativeResp, err := s.UpdateCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// DeleteCategoryHandler 实现 DeleteCategory 方法的 Connect 处理器
// 参数：
//
//	ctx - 上下文
//	req - Connect请求，包含删除分类请求信息
//
// 返回：
//
//	*connect.Response[emptypb.Empty] - Connect响应，包含空响应
//	error - 错误信息
func (s *APIV1Service) DeleteCategoryHandler(ctx context.Context, req *connect.Request[apiv1.DeleteCategoryRequest]) (*connect.Response[emptypb.Empty], error) {
	// 调用gRPC方法
	nativeResp, err := s.DeleteCategory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}

// GetCategoryBySlugHandler 实现 GetCategoryBySlug 方法的 Connect 处理器
// 参数：
//
//	ctx - 上下文
//	req - Connect请求，包含根据标识获取分类请求信息
//
// 返回：
//
//	*connect.Response[pbstore.Category] - Connect响应，包含分类对象
//	error - 错误信息
func (s *APIV1Service) GetCategoryBySlugHandler(ctx context.Context, req *connect.Request[apiv1.GetCategoryBySlugRequest]) (*connect.Response[pbstore.Category], error) {
	// 调用gRPC方法
	nativeResp, err := s.GetCategoryBySlug(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	// 转换为Connect响应
	return connect.NewResponse(nativeResp), nil
}
