package store

import (
	"database/sql"
	"fmt"

	"github.com/wdmsyhh/simple-notes/internal/profile"
)

// Store 数据存储结构体，封装了数据库连接和操作
// 提供了对分类、标签、笔记等数据的CRUD操作接口
type Store struct {
	// driver 数据库驱动实例
	driver Driver
	// profile 服务器配置
	profile *profile.Profile
	// db 数据库连接实例
	db *sql.DB
}

// NewStore 创建一个新的Store实例
func NewStore(driver Driver, profile *profile.Profile) *Store {
	return &Store{
		driver:  driver,
		profile: profile,
		db:      driver.GetDB(),
	}
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	return s.driver.Close()
}

// GetDB 获取底层数据库连接
func (s *Store) GetDB() *sql.DB {
	return s.db
}

// RunMigrations 执行数据库迁移，创建所需的表结构
func (s *Store) RunMigrations() error {
	// 创建用户表
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		username VARCHAR(100) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		nickname VARCHAR(100),
		avatar VARCHAR(255),
		bio VARCHAR(500),
		role VARCHAR(20) DEFAULT 'USER'
	);`

	// 创建分类表
	categoriesTableSQL := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		name_text VARCHAR(100) NOT NULL,
		description VARCHAR(500),
		parent_id INTEGER,
		"order" INTEGER DEFAULT 0,
		visible BOOLEAN DEFAULT TRUE,
		FOREIGN KEY (parent_id) REFERENCES categories(id)
	);`

	// 创建标签表
	tagsTableSQL := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		name_text VARCHAR(100) NOT NULL,
		description VARCHAR(500),
		count INTEGER DEFAULT 0
	);`

	// 创建笔记表
	notesTableSQL := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		title VARCHAR(255) NOT NULL,
		content TEXT,
		summary VARCHAR(500),
		category_id INTEGER,
		tag_ids VARCHAR(500),
		published BOOLEAN DEFAULT FALSE,
		author_id INTEGER,
		published_at DATETIME,
		cover_image VARCHAR(255),
		reading_time INTEGER DEFAULT 0,
		view_count INTEGER DEFAULT 0,
		visibility VARCHAR(20) DEFAULT 'PUBLIC',
		FOREIGN KEY (category_id) REFERENCES categories(id),
		FOREIGN KEY (author_id) REFERENCES users(id)
	);`

	// 创建笔记标签关联表
	noteTagsTableSQL := `
	CREATE TABLE IF NOT EXISTS note_tags (
		note_id INTEGER,
		tag_id INTEGER,
		PRIMARY KEY (note_id, tag_id),
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (tag_id) REFERENCES tags(id)
	);`

	// 创建评论表
	commentsTableSQL := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		note_id INTEGER NOT NULL,
		author VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL,
		content TEXT NOT NULL,
		parent_id INTEGER,
		approved BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (parent_id) REFERENCES comments(id)
	);`

	// 创建页面表
	pagesTableSQL := `
	CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		title VARCHAR(255) NOT NULL,
		slug VARCHAR(255) NOT NULL UNIQUE,
		content TEXT,
		published BOOLEAN DEFAULT FALSE,
		in_navigation BOOLEAN DEFAULT FALSE,
		"order" INTEGER DEFAULT 0
	);`

	// 创建附件表
	attachmentsTableSQL := `
	CREATE TABLE IF NOT EXISTS attachments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		filename VARCHAR(255) NOT NULL,
		type VARCHAR(100) NOT NULL,
		size INTEGER NOT NULL,
		blob BLOB,
		note_id INTEGER,
		author_id INTEGER NOT NULL,
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (author_id) REFERENCES users(id)
	);`

	// 执行所有迁移SQL语句
	migrations := []string{
		usersTableSQL,
		categoriesTableSQL,
		tagsTableSQL,
		notesTableSQL,
		noteTagsTableSQL,
		commentsTableSQL,
		pagesTableSQL,
		attachmentsTableSQL,
	}

	for _, migration := range migrations {
		_, err := s.db.Exec(migration)
		if err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	// 迁移现有表：移除 slug 字段
	// SQLite 不支持直接删除列，需要重建表
	if err := s.migrateRemoveNotesSlug(); err != nil {
		// 如果迁移失败（可能是表不存在或已经是新结构），记录但不中断
		fmt.Printf("Warning: failed to migrate notes slug removal: %v\n", err)
	}

	// 迁移现有表：移除 tags 表的 slug 字段
	if err := s.migrateRemoveTagsSlug(); err != nil {
		// 如果迁移失败（可能是表不存在或已经是新结构），记录但不中断
		fmt.Printf("Warning: failed to migrate tags slug removal: %v\n", err)
	}

	// 迁移现有表：移除 categories 表的 slug 字段
	if err := s.migrateRemoveCategoriesSlug(); err != nil {
		// 如果迁移失败（可能是表不存在或已经是新结构），记录但不中断
		fmt.Printf("Warning: failed to migrate categories slug removal: %v\n", err)
	}

	return nil
}

// migrateRemoveNotesSlug 从 notes 表中移除 slug 字段
// SQLite 不支持 ALTER TABLE DROP COLUMN，需要重建表
func (s *Store) migrateRemoveNotesSlug() error {
	// 检查表是否存在
	var tableExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='notes'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		return nil // 表不存在，不需要迁移
	}

	// 检查 slug 列是否存在
	var slugColumnExists int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('notes') WHERE name = 'slug'
	`).Scan(&slugColumnExists)
	if err != nil || slugColumnExists == 0 {
		return nil // slug 列不存在，不需要迁移
	}

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 创建新表（没有 slug 字段）
	_, err = tx.Exec(`
		CREATE TABLE notes_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			title VARCHAR(255) NOT NULL,
			content TEXT,
			summary VARCHAR(500),
			category_id INTEGER,
			tag_ids VARCHAR(500),
			published BOOLEAN DEFAULT FALSE,
			author_id INTEGER,
			published_at DATETIME,
			cover_image VARCHAR(255),
			reading_time INTEGER DEFAULT 0,
			view_count INTEGER DEFAULT 0,
			visibility VARCHAR(20) DEFAULT 'PUBLIC',
			FOREIGN KEY (category_id) REFERENCES categories(id),
			FOREIGN KEY (author_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// 复制数据（排除 slug 字段）
	_, err = tx.Exec(`
		INSERT INTO notes_new (
			id, created_at, updated_at, deleted_at, title, content, summary,
			category_id, tag_ids, published, author_id, published_at,
			cover_image, reading_time, view_count, visibility
		)
		SELECT 
			id, created_at, updated_at, deleted_at, title, content, summary,
			category_id, tag_ids, published, author_id, published_at,
			cover_image, reading_time, view_count, visibility
		FROM notes
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 删除旧表
	_, err = tx.Exec(`DROP TABLE notes`)
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// 重命名新表
	_, err = tx.Exec(`ALTER TABLE notes_new RENAME TO notes`)
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// migrateRemoveTagsSlug 从 tags 表中移除 slug 字段
// SQLite 不支持 ALTER TABLE DROP COLUMN，需要重建表
func (s *Store) migrateRemoveTagsSlug() error {
	// 检查表是否存在
	var tableExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='tags'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		return nil // 表不存在，不需要迁移
	}

	// 检查 slug 列是否存在
	var slugColumnExists int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tags') WHERE name = 'slug'
	`).Scan(&slugColumnExists)
	if err != nil || slugColumnExists == 0 {
		return nil // slug 列不存在，不需要迁移
	}

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 创建新表（没有 slug 字段）
	_, err = tx.Exec(`
		CREATE TABLE tags_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			name_text VARCHAR(100) NOT NULL,
			description VARCHAR(500),
			count INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// 复制数据（排除 slug 字段）
	_, err = tx.Exec(`
		INSERT INTO tags_new (id, created_at, updated_at, deleted_at, name_text, description, count)
		SELECT id, created_at, updated_at, deleted_at, name_text, description, count FROM tags
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 删除旧表
	_, err = tx.Exec("DROP TABLE tags")
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// 重命名新表
	_, err = tx.Exec("ALTER TABLE tags_new RENAME TO tags")
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// migrateRemoveCategoriesSlug 从 categories 表中移除 slug 字段
// SQLite 不支持 ALTER TABLE DROP COLUMN，需要重建表
func (s *Store) migrateRemoveCategoriesSlug() error {
	// 检查表是否存在
	var tableExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='categories'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		return nil // 表不存在，不需要迁移
	}

	// 检查 slug 列是否存在
	var slugColumnExists int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('categories') WHERE name = 'slug'
	`).Scan(&slugColumnExists)
	if err != nil || slugColumnExists == 0 {
		return nil // slug 列不存在，不需要迁移
	}

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 创建新表（没有 slug 字段）
	_, err = tx.Exec(`
		CREATE TABLE categories_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			name_text VARCHAR(100) NOT NULL,
			description VARCHAR(500),
			parent_id INTEGER,
			"order" INTEGER DEFAULT 0,
			visible BOOLEAN DEFAULT TRUE,
			FOREIGN KEY (parent_id) REFERENCES categories_new(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// 复制数据（排除 slug 字段）
	_, err = tx.Exec(`
		INSERT INTO categories_new (id, created_at, updated_at, deleted_at, name_text, description, parent_id, "order", visible)
		SELECT id, created_at, updated_at, deleted_at, name_text, description, parent_id, "order", visible FROM categories
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 删除旧表
	_, err = tx.Exec("DROP TABLE categories")
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// 重命名新表
	_, err = tx.Exec("ALTER TABLE categories_new RENAME TO categories")
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
