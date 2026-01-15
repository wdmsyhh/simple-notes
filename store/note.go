package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wdmsyhh/simple-notes/proto/gen/store"
)

// ListNotes 获取笔记列表，支持分页和过滤
// 参数：
//
//	ctx - 上下文
//	req - 笔记列表请求，包含分页和过滤条件
//
// 返回：
//
//	[]*store.Note - 笔记列表
//	int64 - 总记录数
//	error - 错误信息
func (s *Store) ListNotes(ctx context.Context, req *ListNotesRequest) ([]*store.Note, int64, error) {
	// 构建基础查询
	query := `SELECT DISTINCT p.* FROM notes p`
	countQuery := `SELECT COUNT(DISTINCT p.id) FROM notes p`
	params := []interface{}{}

	// 构建WHERE条件
	whereConditions := []string{}

	if req.CategoryID != "" {
		whereConditions = append(whereConditions, "p.category_id = ?")
		params = append(params, req.CategoryID)
	}

	if req.TagID != "" {
		query += ` JOIN note_tags pt ON p.id = pt.note_id`
		countQuery += ` JOIN note_tags pt ON p.id = pt.note_id`
		whereConditions = append(whereConditions, "pt.tag_id = ?")
		params = append(params, req.TagID)
	}

	if req.Search != "" {
		whereConditions = append(whereConditions, "p.title LIKE ?")
		params = append(params, "%"+req.Search+"%")
	}

	if !req.IncludeUnpublished {
		whereConditions = append(whereConditions, "p.published = 1")
	}

	// 为查询添加WHERE子句
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
		countQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 计算总数
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, params...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 应用排序
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "p.published_at"
	} else {
		sortBy = "p." + sortBy
	}

	sortOrder := "DESC"
	if !req.SortDesc {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// 应用分页
	offset := (req.Page - 1) * req.PageSize
	query += " LIMIT ? OFFSET ?"
	params = append(params, req.PageSize, offset)

	// 执行查询
	rows, err := s.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// 扫描笔记数据
	var notes []*store.Note
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, 0, err
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

// noteRow 用于扫描数据库行的临时结构体
type noteRow struct {
	// id 笔记ID
	id          uint
	// createdAt 创建时间
	createdAt   time.Time
	// updatedAt 更新时间
	updatedAt   time.Time
	// deletedAt 删除时间（软删除）
	deletedAt   sql.NullTime
	// title 标题
	title       string
	// content 内容
	content     sql.NullString
	// summary 摘要
	summary     sql.NullString
	// categoryID 分类ID
	categoryID  uint
	// tagIDs 标签ID列表（逗号分隔）
	tagIDs      string
	// published 是否已发布
	published   bool
	// authorID 作者ID
	authorID    uint
	// publishedAt 发布时间
	publishedAt time.Time
	// coverImage 封面图片URL
	coverImage  sql.NullString
	// readingTime 阅读时间（分钟）
	readingTime int
	// viewCount 浏览次数
	viewCount   int
	// visibility 可见性
	visibility  string
}

// scanNote 将数据库行扫描到store.Note
// 参数：
//
//	rows - 数据库行（可以是*sql.Row或*sql.Rows）
//
// 返回：
//
//	*store.Note - 笔记信息
//	error - 错误信息
func scanNote(rows interface{}) (*store.Note, error) {
	var row noteRow

	switch v := rows.(type) {
	case *sql.Row:
		if err := v.Scan(
			&row.id,
			&row.createdAt,
			&row.updatedAt,
			&row.deletedAt,
			&row.title,
			&row.content,
			&row.summary,
			&row.categoryID,
			&row.tagIDs,
			&row.published,
			&row.authorID,
			&row.publishedAt,
			&row.coverImage,
			&row.readingTime,
			&row.viewCount,
			&row.visibility,
		); err != nil {
			return nil, err
		}
	case *sql.Rows:
		if err := v.Scan(
			&row.id,
			&row.createdAt,
			&row.updatedAt,
			&row.deletedAt,
			&row.title,
			&row.content,
			&row.summary,
			&row.categoryID,
			&row.tagIDs,
			&row.published,
			&row.authorID,
			&row.publishedAt,
			&row.coverImage,
			&row.readingTime,
			&row.viewCount,
			&row.visibility,
		); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported rows type: %T", rows)
	}

	// 从逗号分隔的字符串解析标签ID
	tagIDs := []string{}
	if row.tagIDs != "" {
		tagIDs = strings.Split(row.tagIDs, ",")
	}

	// 将数据库中的可见性转换为protobuf枚举
	visibility := store.NoteVisibility_NOTE_VISIBILITY_PUBLIC
	if row.visibility == "PRIVATE" {
		visibility = store.NoteVisibility_NOTE_VISIBILITY_PRIVATE
	}

	// 处理可能为 NULL 的字段
	content := ""
	if row.content.Valid {
		content = row.content.String
	}
	summary := ""
	if row.summary.Valid {
		summary = row.summary.String
	}
	coverImage := ""
	if row.coverImage.Valid {
		coverImage = row.coverImage.String
	}

	note := &store.Note{
		Id:          int64(row.id),
		Title:       row.title,
		Slug:        "", // slug 字段已从数据库移除，但 protobuf 定义中仍存在，保持为空
		Content:     content,
		Summary:     summary,
		CategoryId:  fmt.Sprintf("%d", row.categoryID),
		TagIds:      tagIDs,
		Published:   row.published,
		AuthorId:    fmt.Sprintf("%d", row.authorID),
		CreatedAt:   row.createdAt.Unix(),
		UpdatedAt:   row.updatedAt.Unix(),
		PublishedAt: row.publishedAt.Unix(),
		CoverImage:  coverImage,
		ReadingTime: int32(row.readingTime),
		ViewCount:   int32(row.viewCount),
		Visibility:  visibility,
	}

	return note, nil
}

// GetNote 根据ID获取笔记
// 参数：
//
//	ctx - 上下文
//	id - 笔记ID
//
// 返回：
//
//	*store.Note - 笔记信息
//	error - 错误信息
func (s *Store) GetNote(ctx context.Context, id int64) (*store.Note, error) {
	// 查询笔记
	query := `SELECT * FROM notes WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	note, err := scanNote(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("note not found: %d", id)
		}
		return nil, err
	}

	return note, nil
}


// CreateNote 创建新笔记
// 参数：
//
//	ctx - 上下文
//	note - 笔记信息
//
// 返回：
//
//	*store.Note - 创建的笔记信息
//	error - 错误信息
func (s *Store) CreateNote(ctx context.Context, note *store.Note) (*store.Note, error) {
	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 解析分类ID和作者ID
	categoryID := uint(0)
	if note.CategoryId != "" {
		categoryID = parseUint(note.CategoryId)
	}

	authorID := uint(0)
	if note.AuthorId != "" {
		authorID = parseUint(note.AuthorId)
	}

	// 将标签ID转换为逗号分隔的字符串
	tagIDs := ""
	if len(note.TagIds) > 0 {
		tagIDs = strings.Join(note.TagIds, ",")
	}

	// 将protobuf枚举转换为数据库中的可见性
	visibility := "PUBLIC"
	if note.Visibility == store.NoteVisibility_NOTE_VISIBILITY_PRIVATE {
		visibility = "PRIVATE"
	}

	now := time.Now()
	publishedAt := now
	if note.PublishedAt > 0 {
		publishedAt = time.Unix(note.PublishedAt, 0)
	}

	// 插入笔记
	query := `
		INSERT INTO notes (
			title, content, summary, category_id, tag_ids, published, 
			author_id, published_at, cover_image, reading_time, view_count, visibility,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.ExecContext(ctx, query,
		note.Title,
		note.Content,
		note.Summary,
		categoryID,
		tagIDs,
		note.Published,
		authorID,
		publishedAt,
		note.CoverImage,
		note.ReadingTime,
		note.ViewCount,
		visibility,
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

	// 处理标签
	if len(note.TagIds) > 0 {
		for _, tagIDStr := range note.TagIds {
			tagID, err := strconv.ParseUint(tagIDStr, 10, 32)
			if err != nil {
				return nil, err
			}

			// 将标签添加到笔记（插入到note_tags）
			_, err = tx.ExecContext(ctx,
				"INSERT INTO note_tags (note_id, tag_id) VALUES (?, ?)",
				id, tagID,
			)
			if err != nil {
				return nil, err
			}

			// 增加标签计数
			_, err = tx.ExecContext(ctx,
				"UPDATE tags SET count = count + 1 WHERE id = ?",
				tagID,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 重新加载带有关联数据的笔记
	return s.GetNote(ctx, id)
}

// UpdateNote 更新现有笔记
// 参数：
//
//	ctx - 上下文
//	note - 笔记信息
//
// 返回：
//
//	*store.Note - 更新后的笔记信息
//	error - 错误信息
func (s *Store) UpdateNote(ctx context.Context, note *store.Note) (*store.Note, error) {
	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 检查笔记是否存在
	_, err = scanNote(tx.QueryRowContext(ctx, `SELECT * FROM notes WHERE id = ?`, note.Id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("note not found: %d", note.Id)
		}
		return nil, err
	}

	// 加载现有标签
	var existingTagIDs []string
	tagQuery := `SELECT t.id FROM tags t JOIN note_tags pt ON t.id = pt.tag_id WHERE pt.note_id = ?`
	tagRows, err := tx.QueryContext(ctx, tagQuery, note.Id)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var tagID uint
		if err := tagRows.Scan(&tagID); err != nil {
			return nil, err
		}
		existingTagIDs = append(existingTagIDs, fmt.Sprintf("%d", tagID))
	}

	// 解析分类ID和作者ID
	categoryID := uint(0)
	if note.CategoryId != "" {
		categoryID = parseUint(note.CategoryId)
	}

	authorID := uint(0)
	if note.AuthorId != "" {
		authorID = parseUint(note.AuthorId)
	}

	// 将标签ID转换为逗号分隔的字符串
	tagIDs := ""
	if len(note.TagIds) > 0 {
		tagIDs = strings.Join(note.TagIds, ",")
	}

	// 将protobuf枚举转换为数据库中的可见性
	visibility := "PUBLIC"
	if note.Visibility == store.NoteVisibility_NOTE_VISIBILITY_PRIVATE {
		visibility = "PRIVATE"
	}

	publishedAt := time.Now()
	if note.PublishedAt > 0 {
		publishedAt = time.Unix(note.PublishedAt, 0)
	}

	updateQuery := `
		UPDATE notes SET 
			title = ?, content = ?, summary = ?, category_id = ?, tag_ids = ?, 
			published = ?, author_id = ?, published_at = ?, cover_image = ?, reading_time = ?, 
			view_count = ?, visibility = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = tx.ExecContext(ctx, updateQuery,
		note.Title,
		note.Content,
		note.Summary,
		categoryID,
		tagIDs,
		note.Published,
		authorID,
		publishedAt,
		note.CoverImage,
		note.ReadingTime,
		note.ViewCount,
		visibility,
		time.Now(),
		note.Id,
	)
	if err != nil {
		return nil, err
	}

	// 处理标签（如果提供）
	if len(note.TagIds) > 0 {
		// 新标签ID
		newTagIDs := note.TagIds

		// 找出需要添加和删除的标签
		tagsToAdd := make([]string, 0)
		tagsToRemove := make([]string, 0)

		// 需要添加的标签
		for _, newID := range newTagIDs {
			found := false
			for _, existingID := range existingTagIDs {
				if newID == existingID {
					found = true
					break
				}
			}
			if !found {
				tagsToAdd = append(tagsToAdd, newID)
			}
		}

		// 需要删除的标签
		for _, existingID := range existingTagIDs {
			found := false
			for _, newID := range newTagIDs {
				if existingID == newID {
					found = true
					break
				}
			}
			if !found {
				tagsToRemove = append(tagsToRemove, existingID)
			}
		}

		// 删除标签
		for _, tagID := range tagsToRemove {
			tagIDUint, err := strconv.ParseUint(tagID, 10, 32)
			if err != nil {
				return nil, err
			}

			// 从note_tags中删除标签
			_, err = tx.ExecContext(ctx,
				"DELETE FROM note_tags WHERE note_id = ? AND tag_id = ?",
				note.Id, tagIDUint,
			)
			if err != nil {
				return nil, err
			}

			// 减少标签计数
			_, err = tx.ExecContext(ctx,
				"UPDATE tags SET count = GREATEST(count - 1, 0) WHERE id = ?",
				tagIDUint,
			)
			if err != nil {
				return nil, err
			}
		}

		// 添加标签
		for _, tagID := range tagsToAdd {
			tagIDUint, err := strconv.ParseUint(tagID, 10, 32)
			if err != nil {
				return nil, err
			}

			// 将标签添加到note_tags
			_, err = tx.ExecContext(ctx,
				"INSERT INTO note_tags (note_id, tag_id) VALUES (?, ?)",
				note.Id, tagIDUint,
			)
			if err != nil {
				return nil, err
			}

			// 增加标签计数
			_, err = tx.ExecContext(ctx,
				"UPDATE tags SET count = count + 1 WHERE id = ?",
				tagIDUint,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 重新加载带有关联数据的笔记
	return s.GetNote(ctx, note.Id)
}

// DeleteNote 根据ID删除笔记
// 参数：
//
//	ctx - 上下文
//	id - 笔记ID
//
// 返回：
//
//	error - 错误信息
func (s *Store) DeleteNote(ctx context.Context, id int64) error {
	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 获取笔记标签ID
	tagQuery := `SELECT tag_id FROM note_tags WHERE note_id = ?`
	tagRows, err := tx.QueryContext(ctx, tagQuery, id)
	if err != nil {
		return err
	}
	defer tagRows.Close()

	var tagIDs []uint
	for tagRows.Next() {
		var tagID uint
		if err := tagRows.Scan(&tagID); err != nil {
			return err
		}
		tagIDs = append(tagIDs, tagID)
	}

	if err := tagRows.Err(); err != nil {
		return err
	}

	// 减少标签计数
	for _, tagID := range tagIDs {
		_, err = tx.ExecContext(ctx,
			"UPDATE tags SET count = GREATEST(count - 1, 0) WHERE id = ?",
			tagID,
		)
		if err != nil {
			return err
		}
	}

	// 删除note_tags条目
	_, err = tx.ExecContext(ctx, "DELETE FROM note_tags WHERE note_id = ?", id)
	if err != nil {
		return err
	}

	// 删除笔记
	_, err = tx.ExecContext(ctx, "DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}

// ListNotesRequest 笔记列表请求结构体
// 包含分页和过滤条件

type ListNotesRequest struct {
	// Page - 页码
	Page int32
	// PageSize - 每页大小
	PageSize int32
	// CategoryID - 分类ID
	CategoryID string
	// TagID - 标签ID
	TagID string
	// Search - 搜索关键词
	Search string
	// SortBy - 排序字段
	SortBy string
	// SortDesc - 是否降序排序
	SortDesc bool
	// IncludeUnpublished - 是否包含未发布的笔记
	IncludeUnpublished bool
}


// parseUint 将字符串转换为uint
// 参数：
//
//	s - 字符串值
//
// 返回：
//
//	uint - 转换后的uint值
func parseUint(s string) uint {
	if s == "" {
		return 0
	}
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}

// convertTimestampToTime 将int64时间戳转换为time.Time
// 参数：
//
//	ts - 时间戳
//
// 返回：
//
//	time.Time - 转换后的时间
func convertTimestampToTime(ts int64) time.Time {
	if ts == 0 {
		return time.Now()
	}
	return time.Unix(ts, 0)
}
