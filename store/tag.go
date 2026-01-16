package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	"github.com/wdmsyhh/simple-notes/proto/gen/store"
)

// ListTags 获取标签列表，支持可选的分页
func (s *Store) ListTags(ctx context.Context, req *apiv1.ListTagsRequest) ([]*store.Tag, int64, error) {
	// 构建计数查询
	countQuery := `SELECT COUNT(*) FROM tags`

	// 获取总记录数
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 构建主查询
	query := `SELECT * FROM tags ORDER BY count desc, name_text asc`

	// 应用分页
	params := []interface{}{}
	if req.Limit > 0 {
		query += " LIMIT ?"
		params = append(params, req.Limit)

		if req.Offset > 0 {
			query += " OFFSET ?"
			params = append(params, req.Offset)
		}
	}

	// 执行查询
	rows, err := s.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// 扫描标签数据
	var tags []*store.Tag
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, 0, err
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// GetTag 根据ID获取标签
func (s *Store) GetTag(ctx context.Context, tagID int64) (*store.Tag, error) {
	// 根据ID查询标签
	query := `SELECT * FROM tags WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, tagID)

	tag, err := scanTag(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("tag not found: %d", tagID)
		}
		return nil, err
	}

	return tag, nil
}

// CreateTag 创建新标签
func (s *Store) CreateTag(ctx context.Context, tag *store.Tag) (*store.Tag, error) {
	now := time.Now()

	// 插入标签（不包含 slug 字段）
	query := `
		INSERT INTO tags (
			name_text, description, count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		tag.NameText,
		tag.Description,
		tag.Count,
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

	return s.GetTag(ctx, id)
}

// UpdateTag 更新现有标签
func (s *Store) UpdateTag(ctx context.Context, tag *store.Tag) (*store.Tag, error) {
	// 更新标签（不包含 slug 字段）
	query := `
		UPDATE tags SET 
			name_text = ?, description = ?, count = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		tag.NameText,
		tag.Description,
		tag.Count,
		time.Now(),
		tag.Id,
	)
	if err != nil {
		return nil, err
	}

	// 检查标签是否存在并已更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("tag not found: %d", tag.Id)
	}

	return s.GetTag(ctx, tag.Id)
}

// DeleteTag 删除标签
func (s *Store) DeleteTag(ctx context.Context, tagID int64) error {
	// 检查标签下是否有文章
	var noteCount int64
	countQuery := `SELECT COUNT(*) FROM note_tags WHERE tag_id = ?`
	err := s.db.QueryRowContext(ctx, countQuery, tagID).Scan(&noteCount)
	if err != nil {
		return fmt.Errorf("failed to check notes count: %w", err)
	}

	if noteCount > 0 {
		return fmt.Errorf("cannot delete tag: tag has %d note(s)", noteCount)
	}

	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除标签-文章关联（虽然已经检查过没有文章，但为了完整性还是删除）
	if _, err := tx.ExecContext(ctx, "DELETE FROM note_tags WHERE tag_id = ?", tagID); err != nil {
		return err
	}

	// 删除标签
	result, err := tx.ExecContext(ctx, "DELETE FROM tags WHERE id = ?", tagID)
	if err != nil {
		return err
	}

	// 检查标签是否存在并已删除
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found: %d", tagID)
	}

	// 提交事务
	return tx.Commit()
}

// GetTagBySlug 通过slug获取标签
func (s *Store) GetTagBySlug(ctx context.Context, slug string) (*store.Tag, error) {
	// 根据slug查询标签
	query := `SELECT * FROM tags WHERE slug = ?`
	row := s.db.QueryRowContext(ctx, query, slug)

	tag, err := scanTag(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("tag not found with slug: %s", slug)
		}
		return nil, err
	}

	return tag, nil
}

// IncrementTagCount 增加标签计数
func (s *Store) IncrementTagCount(ctx context.Context, tagID int64) error {
	// 更新标签计数
	query := `UPDATE tags SET count = count + 1, updated_at = ? WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, time.Now(), tagID)
	if err != nil {
		return err
	}

	// 检查标签是否存在并已更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found: %d", tagID)
	}

	return nil
}

// DecrementTagCount 减少标签计数，确保不小于0
func (s *Store) DecrementTagCount(ctx context.Context, tagID int64) error {
	// 更新标签计数
	query := `UPDATE tags SET count = GREATEST(count - 1, 0), updated_at = ? WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, time.Now(), tagID)
	if err != nil {
		return err
	}

	// 检查标签是否存在并已更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found: %d", tagID)
	}

	return nil
}

// tagRow 用于扫描数据库行的临时结构体
type tagRow struct {
	// id 标签ID
	id uint
	// createdAt 创建时间
	createdAt time.Time
	// updatedAt 更新时间
	updatedAt time.Time
	// deletedAt 删除时间（软删除）
	deletedAt sql.NullTime
	// nameText 标签名称
	nameText string
	// description 标签描述
	description string
	// count 使用次数
	count int
}

// scanTag 将数据库行扫描到store.Tag
func scanTag(rows interface{}) (*store.Tag, error) {
	var row tagRow

	switch v := rows.(type) {
	case *sql.Row:
		if err := v.Scan(
			&row.id,
			&row.createdAt,
			&row.updatedAt,
			&row.deletedAt,
			&row.nameText,
			&row.description,
			&row.count,
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
			&row.count,
		); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported rows type: %T", rows)
	}

	return &store.Tag{
		Id:          int64(row.id),
		NameText:    row.nameText,
		Slug:        "", // slug 字段已移除，使用主键 id
		Description: row.description,
		Count:       int32(row.count),
		CreatedAt:   row.createdAt.Unix(),
		UpdatedAt:   row.updatedAt.Unix(),
	}, nil
}
