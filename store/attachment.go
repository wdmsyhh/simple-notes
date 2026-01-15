package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/wdmsyhh/simple-notes/proto/gen/store"
)

// CreateAttachment 创建附件
func (s *Store) CreateAttachment(ctx context.Context, attachment *store.Attachment) (*store.Attachment, error) {
	var authorID uint
	fmt.Sscanf(attachment.AuthorId, "%d", &authorID)

	var noteID *int64
	if attachment.NoteId != "" {
		var noteIDInt64 int64
		if _, err := fmt.Sscanf(attachment.NoteId, "notes/%d", &noteIDInt64); err == nil {
			noteID = &noteIDInt64
		}
	}

	query := `
		INSERT INTO attachments (
			filename, type, size, blob, note_id, author_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := s.db.ExecContext(ctx, query,
		attachment.Filename,
		attachment.Type,
		attachment.Size,
		attachment.Content,
		noteID,
		authorID,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attachment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.GetAttachment(ctx, id)
}

// GetAttachment 根据ID获取附件
func (s *Store) GetAttachment(ctx context.Context, id int64) (*store.Attachment, error) {
	query := `SELECT * FROM attachments WHERE id = ? AND deleted_at IS NULL`
	row := s.db.QueryRowContext(ctx, query, id)

	attachment, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("attachment not found: %d", id)
		}
		return nil, err
	}

	return attachment, nil
}

// ListAttachments 获取附件列表
func (s *Store) ListAttachments(ctx context.Context, noteID *int64, authorID *uint) ([]*store.Attachment, error) {
	query := `SELECT * FROM attachments WHERE deleted_at IS NULL`
	params := []interface{}{}

	if noteID != nil {
		query += ` AND note_id = ?`
		params = append(params, *noteID)
	}

	if authorID != nil {
		query += ` AND author_id = ?`
		params = append(params, *authorID)
	}

	// Order by creation time ascending to show attachments in the order they were added
	query += ` ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to list attachments: %w", err)
	}
	defer rows.Close()

	var attachments []*store.Attachment
	for rows.Next() {
		attachment, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// UpdateAttachment 更新附件（主要用于关联到笔记）
func (s *Store) UpdateAttachment(ctx context.Context, attachment *store.Attachment) (*store.Attachment, error) {
	// Extract ID from resource name
	var id int64
	if _, err := fmt.Sscanf(attachment.Name, "attachments/%d", &id); err != nil {
		return nil, fmt.Errorf("invalid attachment name: %s", attachment.Name)
	}

	// Verify attachment exists
	if _, err := s.GetAttachment(ctx, id); err != nil {
		return nil, err
	}

	// Update note_id if provided
	var noteID *int64
	if attachment.NoteId != "" {
		var noteIDInt64 int64
		if _, err := fmt.Sscanf(attachment.NoteId, "notes/%d", &noteIDInt64); err == nil {
			noteID = &noteIDInt64
		}
	}

	query := `UPDATE attachments SET note_id = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, noteID, time.Now(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to update attachment: %w", err)
	}

	// Reload attachment
	return s.GetAttachment(ctx, id)
}

// DeleteAttachment 删除附件
func (s *Store) DeleteAttachment(ctx context.Context, id int64) error {
	query := `UPDATE attachments SET deleted_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}
	return nil
}

// attachmentRow 用于扫描数据库行的临时结构体
type attachmentRow struct {
	// id 附件ID
	id        uint
	// createdAt 创建时间
	createdAt time.Time
	// updatedAt 更新时间
	updatedAt time.Time
	// deletedAt 删除时间（软删除）
	deletedAt sql.NullTime
	// filename 文件名
	filename  string
	// fileType 文件类型（MIME类型）
	fileType  string
	// size 文件大小（字节）
	size      int64
	// blob 文件二进制内容
	blob      []byte
	// noteID 关联的笔记ID（可选）
	noteID    sql.NullInt64
	// authorID 作者ID
	authorID  uint
	}

// scanAttachment 扫描附件数据
// 参数：
//   rows - 数据库行（可以是*sql.Row或*sql.Rows）
// 返回：
//   *store.Attachment - 附件信息
//   error - 错误信息
func scanAttachment(rows interface{}) (*store.Attachment, error) {
	var row attachmentRow

	var err error
	switch v := rows.(type) {
	case *sql.Row:
		err = v.Scan(&row.id, &row.createdAt, &row.updatedAt, &row.deletedAt, &row.filename, &row.fileType, &row.size, &row.blob, &row.noteID, &row.authorID)
	case *sql.Rows:
		err = v.Scan(&row.id, &row.createdAt, &row.updatedAt, &row.deletedAt, &row.filename, &row.fileType, &row.size, &row.blob, &row.noteID, &row.authorID)
	default:
		return nil, fmt.Errorf("unsupported type for scanning")
	}

	if err != nil {
		return nil, err
	}

	attachment := &store.Attachment{
		Name:      fmt.Sprintf("attachments/%d", row.id),
		Id:        int64(row.id),
		Filename:  row.filename,
		Type:      row.fileType,
		Size:      row.size,
		Content:   row.blob,
		AuthorId:  fmt.Sprintf("%d", row.authorID),
		CreatedAt: row.createdAt.Unix(),
		UpdatedAt: row.updatedAt.Unix(),
	}

	if row.noteID.Valid {
		attachment.NoteId = fmt.Sprintf("notes/%d", row.noteID.Int64)
	}

	return attachment, nil
}

