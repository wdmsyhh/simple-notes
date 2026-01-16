package service

import (
	"context"
	"fmt"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	pbstore "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"github.com/wdmsyhh/simple-notes/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CategoryService 处理分类相关操作的服务
// 实现了 CategoryServiceServer 接口
// 封装了分类的增删改查等业务逻辑

type CategoryService struct {
	// store - 数据存储实例，用于数据库操作
	store *store.Store
	// UnimplementedCategoryServiceServer - 未实现的 CategoryService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedCategoryServiceServer
}

// NewCategoryService 创建一个新的 CategoryService 实例
func NewCategoryService(store *store.Store) *CategoryService {
	return &CategoryService{
		store: store,
	}
}

// ListCategories 获取分类列表，支持可选的过滤条件
func (s *CategoryService) ListCategories(ctx context.Context, req *apiv1.ListCategoriesRequest) (*apiv1.ListCategoriesResponse, error) {
	// 调用存储层
	categories, err := s.store.ListCategories(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// 创建响应
	response := &apiv1.ListCategoriesResponse{
		Categories: categories,
	}

	return response, nil
}

// GetCategory 根据ID获取分类
func (s *CategoryService) GetCategory(ctx context.Context, req *apiv1.GetCategoryRequest) (*pbstore.Category, error) {
	// 从资源名称中提取分类ID
	categoryID, err := extractIDFromResourceName(req.GetName(), "categories")
	if err != nil {
		return nil, err
	}

	// 调用存储层
	category, err := s.store.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// 设置资源名称
	category.Name = fmt.Sprintf("categories/%d", category.Id)

	return category, nil
}

// CreateCategory 创建新分类
func (s *CategoryService) CreateCategory(ctx context.Context, req *apiv1.CreateCategoryRequest) (*pbstore.Category, error) {
	// 从请求中获取分类
	category := req.GetCategory()
	if category == nil {
		return nil, fmt.Errorf("category is required")
	}

	// 验证分类数据
	if category.NameText == "" {
		return nil, fmt.Errorf("name is required")
	}

	if category.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	// 调用存储层
	createdCategory, err := s.store.CreateCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// 设置资源名称
	createdCategory.Name = fmt.Sprintf("categories/%d", createdCategory.Id)

	return createdCategory, nil
}

// UpdateCategory 更新现有分类
func (s *CategoryService) UpdateCategory(ctx context.Context, req *apiv1.UpdateCategoryRequest) (*pbstore.Category, error) {
	// 从请求中获取分类
	category := req.GetCategory()
	if category == nil {
		return nil, fmt.Errorf("category is required")
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
		return nil, fmt.Errorf("category ID is required")
	}

	// 调用存储层
	updatedCategory, err := s.store.UpdateCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	// 设置资源名称
	updatedCategory.Name = fmt.Sprintf("categories/%d", updatedCategory.Id)

	return updatedCategory, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(ctx context.Context, req *apiv1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	// 从资源名称中提取分类ID
	categoryID, err := extractIDFromResourceName(req.GetName(), "categories")
	if err != nil {
		return nil, err
	}

	// 调用存储层
	if err := s.store.DeleteCategory(ctx, categoryID); err != nil {
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// GetCategoryBySlug 通过slug获取分类
func (s *CategoryService) GetCategoryBySlug(ctx context.Context, req *apiv1.GetCategoryBySlugRequest) (*pbstore.Category, error) {
	slug := req.GetSlug()
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	// 调用存储层
	category, err := s.store.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	// 设置资源名称
	category.Name = fmt.Sprintf("categories/%d", category.Id)

	return category, nil
}
