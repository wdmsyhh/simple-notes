package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	"github.com/wdmsyhh/simple-notes/proto/gen/store"
)

// ListCategories 获取分类列表，支持可选的过滤条件
func (s *Store) ListCategories(ctx context.Context, req *apiv1.ListCategoriesRequest) ([]*store.Category, error) {
	// 构建查询语句
	query := `SELECT * FROM categories`
	params := []interface{}{}

	// 构建WHERE条件
	whereConditions := []string{}

	if !req.IncludeHidden {
		whereConditions = append(whereConditions, "visible = ?")
		params = append(params, true)
	}

	if req.ParentId > 0 {
		whereConditions = append(whereConditions, "parent_id = ?")
		params = append(params, req.ParentId)
	}

	// 添加WHERE子句
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 添加ORDER子句
	query += ` ORDER BY "order" asc, created_at desc`

	// 执行查询
	rows, err := s.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 扫描分类数据
	var categories []*store.Category
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// GetCategory 根据ID获取分类
func (s *Store) GetCategory(ctx context.Context, categoryID int64) (*store.Category, error) {
	// 根据ID查询分类
	query := `SELECT * FROM categories WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, categoryID)

	category, err := scanCategory(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category not found: %d", categoryID)
		}
		return nil, err
	}

	return category, nil
}

// CreateCategory 创建新分类
func (s *Store) CreateCategory(ctx context.Context, category *store.Category) (*store.Category, error) {
	now := time.Now()
	parentID := uint(0)
	if category.ParentId > 0 {
		parentID = uint(category.ParentId)
	}

	// 插入分类（不包含 slug 字段）
	query := `
		INSERT INTO categories (
			name_text, description, parent_id, "order", visible, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		category.NameText,
		category.Description,
		parentID,
		category.Order,
		category.Visible,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetCategory(ctx, id)
}

// UpdateCategory 更新现有分类
func (s *Store) UpdateCategory(ctx context.Context, category *store.Category) (*store.Category, error) {
	parentID := uint(0)
	if category.ParentId > 0 {
		parentID = uint(category.ParentId)
	}

	// 更新分类（不包含 slug 字段）
	query := `
		UPDATE categories SET 
			name_text = ?, description = ?, parent_id = ?, "order" = ?, 
			visible = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		category.NameText,
		category.Description,
		parentID,
		category.Order,
		category.Visible,
		time.Now(),
		category.Id,
	)
	if err != nil {
		return nil, err
	}

	// 检查分类是否存在并已更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("category not found: %d", category.Id)
	}

	return s.GetCategory(ctx, category.Id)
}

// DeleteCategory 删除分类
func (s *Store) DeleteCategory(ctx context.Context, categoryID int64) error {
	// 检查分类下是否有文章
	var noteCount int64
	countQuery := `SELECT COUNT(*) FROM notes WHERE category_id = ?`
	err := s.db.QueryRowContext(ctx, countQuery, categoryID).Scan(&noteCount)
	if err != nil {
		return fmt.Errorf("failed to check notes count: %w", err)
	}

	if noteCount > 0 {
		return fmt.Errorf("cannot delete category: category has %d note(s)", noteCount)
	}

	// 删除分类
	query := `DELETE FROM categories WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, categoryID)
	if err != nil {
		return err
	}

	// 检查分类是否存在并已删除
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category not found: %d", categoryID)
	}

	return nil
}

// GetCategoryBySlug 通过slug获取分类
func (s *Store) GetCategoryBySlug(ctx context.Context, slug string) (*store.Category, error) {
	// 根据slug查询分类
	query := `SELECT * FROM categories WHERE slug = ?`
	row := s.db.QueryRowContext(ctx, query, slug)

	category, err := scanCategory(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category not found with slug: %s", slug)
		}
		return nil, err
	}

	return category, nil
}

// categoryRow 用于扫描数据库行的临时结构体
type categoryRow struct {
	// id 分类ID
	id uint
	// createdAt 创建时间
	createdAt time.Time
	// updatedAt 更新时间
	updatedAt time.Time
	// deletedAt 删除时间（软删除）
	deletedAt sql.NullTime
	// nameText 分类名称
	nameText string
	// description 分类描述
	description string
	// parentID 父分类ID
	parentID uint
	// order 排序顺序
	order int
	// visible 是否可见
	visible bool
}

// scanCategory 将数据库行扫描到store.Category
func scanCategory(rows interface{}) (*store.Category, error) {
	var row categoryRow

	switch v := rows.(type) {
	case *sql.Row:
		if err := v.Scan(
			&row.id,
			&row.createdAt,
			&row.updatedAt,
			&row.deletedAt,
			&row.nameText,
			&row.description,
			&row.parentID,
			&row.order,
			&row.visible,
		); err != nil {
			return nil, err
		}
	case *sql.Rows:
		if err := v.Scan(
			&row.id,
			&row.createdAt,
			&row.updatedAt,
			&row.deletedAt,
			&row.nameText,
			&row.description,
			&row.parentID,
			&row.order,
			&row.visible,
		); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported rows type: %T", rows)
	}

	category := &store.Category{
		Id:          int64(row.id),
		NameText:    row.nameText,
		Slug:        "", // slug 字段已移除，使用主键 id
		Description: row.description,
		Order:       int32(row.order),
		Visible:     row.visible,
		CreatedAt:   row.createdAt.Unix(),
		UpdatedAt:   row.updatedAt.Unix(),
	}

	if row.parentID > 0 {
		category.ParentId = int64(row.parentID)
	}

	return category, nil
}
