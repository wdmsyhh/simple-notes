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
	switch s.profile.Driver {
	case "sqlite":
		return s.runSQLiteMigrations()
	case "mysql":
		return s.runMySQLMigrations()
	case "postgres":
		return s.runPostgresMigrations()
	default:
		return fmt.Errorf("unsupported database driver: %s", s.profile.Driver)
	}
}

// runSQLiteMigrations 执行 SQLite 数据库迁移
func (s *Store) runSQLiteMigrations() error {
	// 创建用户表
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 用户ID，主键，自增
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间，默认当前时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间，默认当前时间
		deleted_at DATETIME, -- 删除时间（软删除），NULL表示未删除
		username VARCHAR(100) NOT NULL UNIQUE, -- 用户名，必填，唯一
		password_hash VARCHAR(255) NOT NULL, -- 密码哈希值，必填
		nickname VARCHAR(100), -- 昵称，可选
		avatar VARCHAR(255), -- 头像URL，可选
		bio VARCHAR(500), -- 个人简介，可选
		role VARCHAR(20) DEFAULT 'USER' -- 用户角色，默认普通用户（USER/HOST/ADMIN）
	);`

	// 创建分类表
	categoriesTableSQL := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 分类ID，主键，自增
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间，默认当前时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间，默认当前时间
		deleted_at DATETIME, -- 删除时间（软删除），NULL表示未删除
		name_text VARCHAR(100) NOT NULL, -- 分类名称，必填
		description VARCHAR(500), -- 分类描述，可选
		parent_id INTEGER, -- 父分类ID，可选，用于构建分类树
		"order" INTEGER DEFAULT 0, -- 排序顺序，默认0
		visible BOOLEAN DEFAULT TRUE, -- 是否可见，默认可见
		FOREIGN KEY (parent_id) REFERENCES categories(id) -- 外键，引用父分类
	);`

	// 创建标签表
	tagsTableSQL := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 标签ID，主键，自增
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间，默认当前时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间，默认当前时间
		deleted_at DATETIME, -- 删除时间（软删除），NULL表示未删除
		name_text VARCHAR(100) NOT NULL, -- 标签名称，必填
		description VARCHAR(500), -- 标签描述，可选
		count INTEGER DEFAULT 0 -- 使用次数，默认0
	);`

	// 创建笔记表
	notesTableSQL := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 笔记ID，主键，自增
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间，默认当前时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间，默认当前时间
		deleted_at DATETIME, -- 删除时间（软删除），NULL表示未删除
		title VARCHAR(255) NOT NULL, -- 笔记标题，必填
		content TEXT, -- 笔记内容（Markdown格式），可选
		summary VARCHAR(500), -- 笔记摘要，可选
		category_id INTEGER, -- 分类ID，可选
		tag_ids VARCHAR(500), -- 标签ID列表（逗号分隔），可选
		published BOOLEAN DEFAULT FALSE, -- 是否已发布，默认未发布
		author_id INTEGER, -- 作者ID，可选
		published_at DATETIME, -- 发布时间，可选
		cover_image VARCHAR(255), -- 封面图片URL，可选
		reading_time INTEGER DEFAULT 0, -- 阅读时间（分钟），默认0
		view_count INTEGER DEFAULT 0, -- 浏览次数，默认0
		visibility VARCHAR(20) DEFAULT 'PUBLIC', -- 可见性（PUBLIC/PRIVATE），默认公开
		FOREIGN KEY (category_id) REFERENCES categories(id), -- 外键，引用分类
		FOREIGN KEY (author_id) REFERENCES users(id) -- 外键，引用用户
	);`

	// 创建笔记标签关联表
	noteTagsTableSQL := `
	CREATE TABLE IF NOT EXISTS note_tags (
		note_id INTEGER, -- 笔记ID
		tag_id INTEGER, -- 标签ID
		PRIMARY KEY (note_id, tag_id), -- 联合主键，确保每个笔记-标签组合唯一
		FOREIGN KEY (note_id) REFERENCES notes(id), -- 外键，引用笔记
		FOREIGN KEY (tag_id) REFERENCES tags(id) -- 外键，引用标签
	);`

	// 创建评论表
	commentsTableSQL := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 评论ID，主键，自增
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间，默认当前时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间，默认当前时间
		deleted_at DATETIME, -- 删除时间（软删除），NULL表示未删除
		note_id INTEGER NOT NULL, -- 笔记ID，必填
		author VARCHAR(100) NOT NULL, -- 评论作者名称，必填
		email VARCHAR(100) NOT NULL, -- 评论作者邮箱，必填
		content TEXT NOT NULL, -- 评论内容，必填
		parent_id INTEGER, -- 父评论ID，可选，用于回复功能
		approved BOOLEAN DEFAULT FALSE, -- 是否已审核，默认未审核
		FOREIGN KEY (note_id) REFERENCES notes(id), -- 外键，引用笔记
		FOREIGN KEY (parent_id) REFERENCES comments(id) -- 外键，引用父评论
	);`

	// 创建页面表
	pagesTableSQL := `
	CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 页面ID，主键，自增
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间，默认当前时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间，默认当前时间
		deleted_at DATETIME, -- 删除时间（软删除），NULL表示未删除
		title VARCHAR(255) NOT NULL, -- 页面标题，必填
		slug VARCHAR(255) NOT NULL UNIQUE, -- URL友好的标识符，必填，唯一
		content TEXT, -- 页面内容，可选
		published BOOLEAN DEFAULT FALSE, -- 是否已发布，默认未发布
		in_navigation BOOLEAN DEFAULT FALSE, -- 是否在导航中显示，默认不显示
		"order" INTEGER DEFAULT 0 -- 排序顺序，默认0
	);`

	// 创建附件表
	attachmentsTableSQL := `
	CREATE TABLE IF NOT EXISTS attachments (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 附件ID，主键，自增
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间，默认当前时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间，默认当前时间
		deleted_at DATETIME, -- 删除时间（软删除），NULL表示未删除
		filename VARCHAR(255) NOT NULL, -- 文件名，必填
		type VARCHAR(100) NOT NULL, -- MIME类型，必填
		size INTEGER NOT NULL, -- 文件大小（字节），必填
		blob BLOB, -- 文件二进制内容，可选（可存储在文件系统中）
		note_id INTEGER, -- 关联的笔记ID，可选
		author_id INTEGER NOT NULL, -- 上传者ID，必填
		FOREIGN KEY (note_id) REFERENCES notes(id), -- 外键，引用笔记
		FOREIGN KEY (author_id) REFERENCES users(id) -- 外键，引用用户
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

	// 迁移现有表：移除 users 表的 email 字段
	if err := s.migrateRemoveUsersEmailSQLite(); err != nil {
		fmt.Printf("Warning: failed to migrate users email removal: %v\n", err)
	}

	return nil
}

// runMySQLMigrations 执行 MySQL 数据库迁移
func (s *Store) runMySQLMigrations() error {
	// 创建用户表
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID，主键，自增',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间，默认当前时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间，默认当前时间',
		deleted_at DATETIME NULL COMMENT '删除时间（软删除），NULL表示未删除',
		username VARCHAR(100) NOT NULL UNIQUE COMMENT '用户名，必填，唯一',
		password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希值，必填',
		nickname VARCHAR(100) NULL COMMENT '昵称，可选',
		avatar VARCHAR(255) NULL COMMENT '头像URL，可选',
		bio VARCHAR(500) NULL COMMENT '个人简介，可选',
		role VARCHAR(20) DEFAULT 'USER' COMMENT '用户角色，默认普通用户（USER/HOST/ADMIN）'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 创建分类表
	categoriesTableSQL := `
	CREATE TABLE IF NOT EXISTS categories (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '分类ID，主键，自增',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间，默认当前时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间，默认当前时间',
		deleted_at DATETIME NULL COMMENT '删除时间（软删除），NULL表示未删除',
		name_text VARCHAR(100) NOT NULL COMMENT '分类名称，必填',
		description VARCHAR(500) NULL COMMENT '分类描述，可选',
		parent_id INT NULL COMMENT '父分类ID，可选，用于构建分类树',
		` + "`order`" + ` INT DEFAULT 0 COMMENT '排序顺序，默认0',
		visible BOOLEAN DEFAULT TRUE COMMENT '是否可见，默认可见',
		FOREIGN KEY (parent_id) REFERENCES categories(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 创建标签表
	tagsTableSQL := `
	CREATE TABLE IF NOT EXISTS tags (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '标签ID，主键，自增',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间，默认当前时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间，默认当前时间',
		deleted_at DATETIME NULL COMMENT '删除时间（软删除），NULL表示未删除',
		name_text VARCHAR(100) NOT NULL COMMENT '标签名称，必填',
		description VARCHAR(500) NULL COMMENT '标签描述，可选',
		count INT DEFAULT 0 COMMENT '使用次数，默认0'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 创建笔记表
	notesTableSQL := `
	CREATE TABLE IF NOT EXISTS notes (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '笔记ID，主键，自增',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间，默认当前时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间，默认当前时间',
		deleted_at DATETIME NULL COMMENT '删除时间（软删除），NULL表示未删除',
		title VARCHAR(255) NOT NULL COMMENT '笔记标题，必填',
		content TEXT NULL COMMENT '笔记内容（Markdown格式），可选',
		summary VARCHAR(500) NULL COMMENT '笔记摘要，可选',
		category_id INT NULL COMMENT '分类ID，可选',
		tag_ids VARCHAR(500) NULL COMMENT '标签ID列表（逗号分隔），可选',
		published BOOLEAN DEFAULT FALSE COMMENT '是否已发布，默认未发布',
		author_id INT NULL COMMENT '作者ID，可选',
		published_at DATETIME NULL COMMENT '发布时间，可选',
		cover_image VARCHAR(255) NULL COMMENT '封面图片URL，可选',
		reading_time INT DEFAULT 0 COMMENT '阅读时间（分钟），默认0',
		view_count INT DEFAULT 0 COMMENT '浏览次数，默认0',
		visibility VARCHAR(20) DEFAULT 'PUBLIC' COMMENT '可见性（PUBLIC/PRIVATE），默认公开',
		FOREIGN KEY (category_id) REFERENCES categories(id),
		FOREIGN KEY (author_id) REFERENCES users(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 创建笔记标签关联表
	noteTagsTableSQL := `
	CREATE TABLE IF NOT EXISTS note_tags (
		note_id INT NOT NULL COMMENT '笔记ID',
		tag_id INT NOT NULL COMMENT '标签ID',
		PRIMARY KEY (note_id, tag_id),
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (tag_id) REFERENCES tags(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 创建评论表
	commentsTableSQL := `
	CREATE TABLE IF NOT EXISTS comments (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '评论ID，主键，自增',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间，默认当前时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间，默认当前时间',
		deleted_at DATETIME NULL COMMENT '删除时间（软删除），NULL表示未删除',
		note_id INT NOT NULL COMMENT '笔记ID，必填',
		author VARCHAR(100) NOT NULL COMMENT '评论作者名称，必填',
		email VARCHAR(100) NOT NULL COMMENT '评论作者邮箱，必填',
		content TEXT NOT NULL COMMENT '评论内容，必填',
		parent_id INT NULL COMMENT '父评论ID，可选，用于回复功能',
		approved BOOLEAN DEFAULT FALSE COMMENT '是否已审核，默认未审核',
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (parent_id) REFERENCES comments(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 创建页面表
	pagesTableSQL := `
	CREATE TABLE IF NOT EXISTS pages (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '页面ID，主键，自增',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间，默认当前时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间，默认当前时间',
		deleted_at DATETIME NULL COMMENT '删除时间（软删除），NULL表示未删除',
		title VARCHAR(255) NOT NULL COMMENT '页面标题，必填',
		slug VARCHAR(255) NOT NULL UNIQUE COMMENT 'URL友好的标识符，必填，唯一',
		content TEXT NULL COMMENT '页面内容，可选',
		published BOOLEAN DEFAULT FALSE COMMENT '是否已发布，默认未发布',
		in_navigation BOOLEAN DEFAULT FALSE COMMENT '是否在导航中显示，默认不显示',
		` + "`order`" + ` INT DEFAULT 0 COMMENT '排序顺序，默认0'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 创建附件表
	attachmentsTableSQL := `
	CREATE TABLE IF NOT EXISTS attachments (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '附件ID，主键，自增',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间，默认当前时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间，默认当前时间',
		deleted_at DATETIME NULL COMMENT '删除时间（软删除），NULL表示未删除',
		filename VARCHAR(255) NOT NULL COMMENT '文件名，必填',
		type VARCHAR(100) NOT NULL COMMENT 'MIME类型，必填',
		size INT NOT NULL COMMENT '文件大小（字节），必填',
		blob LONGBLOB NULL COMMENT '文件二进制内容，可选（可存储在文件系统中）',
		note_id INT NULL COMMENT '关联的笔记ID，可选',
		author_id INT NOT NULL COMMENT '上传者ID，必填',
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (author_id) REFERENCES users(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

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

	// 迁移现有表：移除 users 表的 email 字段
	if err := s.migrateRemoveUsersEmailMySQL(); err != nil {
		fmt.Printf("Warning: failed to migrate users email removal: %v\n", err)
	}

	return nil
}

// runPostgresMigrations 执行 PostgreSQL 数据库迁移
func (s *Store) runPostgresMigrations() error {
	// 创建用户表
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
		username VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		nickname VARCHAR(100),
		avatar VARCHAR(255),
		bio VARCHAR(500),
		role VARCHAR(20) DEFAULT 'USER'
	);`

	// 创建分类表
	categoriesTableSQL := `
	CREATE TABLE IF NOT EXISTS categories (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
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
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
		name_text VARCHAR(100) NOT NULL,
		description VARCHAR(500),
		count INTEGER DEFAULT 0
	);`

	// 创建笔记表
	notesTableSQL := `
	CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
		title VARCHAR(255) NOT NULL,
		content TEXT,
		summary VARCHAR(500),
		category_id INTEGER,
		tag_ids VARCHAR(500),
		published BOOLEAN DEFAULT FALSE,
		author_id INTEGER,
		published_at TIMESTAMP,
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
		note_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (note_id, tag_id),
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (tag_id) REFERENCES tags(id)
	);`

	// 创建评论表
	commentsTableSQL := `
	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
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
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
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
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
		filename VARCHAR(255) NOT NULL,
		type VARCHAR(100) NOT NULL,
		size INTEGER NOT NULL,
		blob BYTEA,
		note_id INTEGER,
		author_id INTEGER NOT NULL,
		FOREIGN KEY (note_id) REFERENCES notes(id),
		FOREIGN KEY (author_id) REFERENCES users(id)
	);`

	// 执行所有迁移SQL语句
	migrations := []struct {
		tableSQL string
		comments []string
	}{
		{
			tableSQL: usersTableSQL,
			comments: []string{
				"COMMENT ON COLUMN users.id IS '用户ID，主键，自增'",
				"COMMENT ON COLUMN users.created_at IS '创建时间，默认当前时间'",
				"COMMENT ON COLUMN users.updated_at IS '更新时间，默认当前时间'",
				"COMMENT ON COLUMN users.deleted_at IS '删除时间（软删除），NULL表示未删除'",
				"COMMENT ON COLUMN users.username IS '用户名，必填，唯一'",
				"COMMENT ON COLUMN users.password_hash IS '密码哈希值，必填'",
				"COMMENT ON COLUMN users.nickname IS '昵称，可选'",
				"COMMENT ON COLUMN users.avatar IS '头像URL，可选'",
				"COMMENT ON COLUMN users.bio IS '个人简介，可选'",
				"COMMENT ON COLUMN users.role IS '用户角色，默认普通用户（USER/HOST/ADMIN）'",
			},
		},
		{
			tableSQL: categoriesTableSQL,
			comments: []string{
				"COMMENT ON COLUMN categories.id IS '分类ID，主键，自增'",
				"COMMENT ON COLUMN categories.created_at IS '创建时间，默认当前时间'",
				"COMMENT ON COLUMN categories.updated_at IS '更新时间，默认当前时间'",
				"COMMENT ON COLUMN categories.deleted_at IS '删除时间（软删除），NULL表示未删除'",
				"COMMENT ON COLUMN categories.name_text IS '分类名称，必填'",
				"COMMENT ON COLUMN categories.description IS '分类描述，可选'",
				"COMMENT ON COLUMN categories.parent_id IS '父分类ID，可选，用于构建分类树'",
				"COMMENT ON COLUMN categories.\"order\" IS '排序顺序，默认0'",
				"COMMENT ON COLUMN categories.visible IS '是否可见，默认可见'",
			},
		},
		{
			tableSQL: tagsTableSQL,
			comments: []string{
				"COMMENT ON COLUMN tags.id IS '标签ID，主键，自增'",
				"COMMENT ON COLUMN tags.created_at IS '创建时间，默认当前时间'",
				"COMMENT ON COLUMN tags.updated_at IS '更新时间，默认当前时间'",
				"COMMENT ON COLUMN tags.deleted_at IS '删除时间（软删除），NULL表示未删除'",
				"COMMENT ON COLUMN tags.name_text IS '标签名称，必填'",
				"COMMENT ON COLUMN tags.description IS '标签描述，可选'",
				"COMMENT ON COLUMN tags.count IS '使用次数，默认0'",
			},
		},
		{
			tableSQL: notesTableSQL,
			comments: []string{
				"COMMENT ON COLUMN notes.id IS '笔记ID，主键，自增'",
				"COMMENT ON COLUMN notes.created_at IS '创建时间，默认当前时间'",
				"COMMENT ON COLUMN notes.updated_at IS '更新时间，默认当前时间'",
				"COMMENT ON COLUMN notes.deleted_at IS '删除时间（软删除），NULL表示未删除'",
				"COMMENT ON COLUMN notes.title IS '笔记标题，必填'",
				"COMMENT ON COLUMN notes.content IS '笔记内容（Markdown格式），可选'",
				"COMMENT ON COLUMN notes.summary IS '笔记摘要，可选'",
				"COMMENT ON COLUMN notes.category_id IS '分类ID，可选'",
				"COMMENT ON COLUMN notes.tag_ids IS '标签ID列表（逗号分隔），可选'",
				"COMMENT ON COLUMN notes.published IS '是否已发布，默认未发布'",
				"COMMENT ON COLUMN notes.author_id IS '作者ID，可选'",
				"COMMENT ON COLUMN notes.published_at IS '发布时间，可选'",
				"COMMENT ON COLUMN notes.cover_image IS '封面图片URL，可选'",
				"COMMENT ON COLUMN notes.reading_time IS '阅读时间（分钟），默认0'",
				"COMMENT ON COLUMN notes.view_count IS '浏览次数，默认0'",
				"COMMENT ON COLUMN notes.visibility IS '可见性（PUBLIC/PRIVATE），默认公开'",
			},
		},
		{
			tableSQL: noteTagsTableSQL,
			comments: []string{
				"COMMENT ON COLUMN note_tags.note_id IS '笔记ID'",
				"COMMENT ON COLUMN note_tags.tag_id IS '标签ID'",
			},
		},
		{
			tableSQL: commentsTableSQL,
			comments: []string{
				"COMMENT ON COLUMN comments.id IS '评论ID，主键，自增'",
				"COMMENT ON COLUMN comments.created_at IS '创建时间，默认当前时间'",
				"COMMENT ON COLUMN comments.updated_at IS '更新时间，默认当前时间'",
				"COMMENT ON COLUMN comments.deleted_at IS '删除时间（软删除），NULL表示未删除'",
				"COMMENT ON COLUMN comments.note_id IS '笔记ID，必填'",
				"COMMENT ON COLUMN comments.author IS '评论作者名称，必填'",
				"COMMENT ON COLUMN comments.email IS '评论作者邮箱，必填'",
				"COMMENT ON COLUMN comments.content IS '评论内容，必填'",
				"COMMENT ON COLUMN comments.parent_id IS '父评论ID，可选，用于回复功能'",
				"COMMENT ON COLUMN comments.approved IS '是否已审核，默认未审核'",
			},
		},
		{
			tableSQL: pagesTableSQL,
			comments: []string{
				"COMMENT ON COLUMN pages.id IS '页面ID，主键，自增'",
				"COMMENT ON COLUMN pages.created_at IS '创建时间，默认当前时间'",
				"COMMENT ON COLUMN pages.updated_at IS '更新时间，默认当前时间'",
				"COMMENT ON COLUMN pages.deleted_at IS '删除时间（软删除），NULL表示未删除'",
				"COMMENT ON COLUMN pages.title IS '页面标题，必填'",
				"COMMENT ON COLUMN pages.slug IS 'URL友好的标识符，必填，唯一'",
				"COMMENT ON COLUMN pages.content IS '页面内容，可选'",
				"COMMENT ON COLUMN pages.published IS '是否已发布，默认未发布'",
				"COMMENT ON COLUMN pages.in_navigation IS '是否在导航中显示，默认不显示'",
				"COMMENT ON COLUMN pages.\"order\" IS '排序顺序，默认0'",
			},
		},
		{
			tableSQL: attachmentsTableSQL,
			comments: []string{
				"COMMENT ON COLUMN attachments.id IS '附件ID，主键，自增'",
				"COMMENT ON COLUMN attachments.created_at IS '创建时间，默认当前时间'",
				"COMMENT ON COLUMN attachments.updated_at IS '更新时间，默认当前时间'",
				"COMMENT ON COLUMN attachments.deleted_at IS '删除时间（软删除），NULL表示未删除'",
				"COMMENT ON COLUMN attachments.filename IS '文件名，必填'",
				"COMMENT ON COLUMN attachments.type IS 'MIME类型，必填'",
				"COMMENT ON COLUMN attachments.size IS '文件大小（字节），必填'",
				"COMMENT ON COLUMN attachments.blob IS '文件二进制内容，可选（可存储在文件系统中）'",
				"COMMENT ON COLUMN attachments.note_id IS '关联的笔记ID，可选'",
				"COMMENT ON COLUMN attachments.author_id IS '上传者ID，必填'",
			},
		},
	}

	for _, migration := range migrations {
		// 创建表
		_, err := s.db.Exec(migration.tableSQL)
		if err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}

		// 立即添加字段注释
		for _, comment := range migration.comments {
			_, err := s.db.Exec(comment)
			if err != nil {
				// 如果注释已存在或表不存在，忽略错误
				// PostgreSQL 允许重复执行 COMMENT ON COLUMN，会更新注释
				continue
			}
		}
	}

	// 迁移现有表：移除 users 表的 email 字段
	if err := s.migrateRemoveUsersEmailPostgres(); err != nil {
		fmt.Printf("Warning: failed to migrate users email removal: %v\n", err)
	}

	return nil
}

// migrateRemoveUsersEmailSQLite 从 users 表中移除 email 字段（SQLite）
// SQLite 不支持 ALTER TABLE DROP COLUMN，需要重建表
func (s *Store) migrateRemoveUsersEmailSQLite() error {
	// 检查表是否存在
	var tableExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		return nil // 表不存在，不需要迁移
	}

	// 检查 email 列是否存在
	var emailColumnExists int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = 'email'
	`).Scan(&emailColumnExists)
	if err != nil || emailColumnExists == 0 {
		return nil // email 列不存在，不需要迁移
	}

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 创建新表（没有 email 字段）
	_, err = tx.Exec(`
		CREATE TABLE users_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			username VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			nickname VARCHAR(100),
			avatar VARCHAR(255),
			bio VARCHAR(500),
			role VARCHAR(20) DEFAULT 'USER'
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// 复制数据（排除 email 字段）
	_, err = tx.Exec(`
		INSERT INTO users_new (
			id, created_at, updated_at, deleted_at, username, password_hash,
			nickname, avatar, bio, role
		)
		SELECT 
			id, created_at, updated_at, deleted_at, username, password_hash,
			nickname, avatar, bio, role
		FROM users
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 删除旧表
	_, err = tx.Exec(`DROP TABLE users`)
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// 重命名新表
	_, err = tx.Exec(`ALTER TABLE users_new RENAME TO users`)
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// migrateRemoveUsersEmailMySQL 从 users 表中移除 email 字段（MySQL）
func (s *Store) migrateRemoveUsersEmailMySQL() error {
	// 检查表是否存在
	var tableExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		return nil // 表不存在，不需要迁移
	}

	// 检查 email 列是否存在
	var emailColumnExists int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_schema = DATABASE() AND table_name = 'users' AND column_name = 'email'
	`).Scan(&emailColumnExists)
	if err != nil || emailColumnExists == 0 {
		return nil // email 列不存在，不需要迁移
	}

	// MySQL 支持直接删除列
	_, err = s.db.Exec(`ALTER TABLE users DROP COLUMN email`)
	if err != nil {
		return fmt.Errorf("failed to drop email column: %w", err)
	}

	return nil
}

// migrateRemoveUsersEmailPostgres 从 users 表中移除 email 字段（PostgreSQL）
func (s *Store) migrateRemoveUsersEmailPostgres() error {
	// 检查表是否存在
	var tableExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		return nil // 表不存在，不需要迁移
	}

	// 检查 email 列是否存在
	var emailColumnExists int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'email'
	`).Scan(&emailColumnExists)
	if err != nil || emailColumnExists == 0 {
		return nil // email 列不存在，不需要迁移
	}

	// PostgreSQL 支持直接删除列
	_, err = s.db.Exec(`ALTER TABLE users DROP COLUMN email`)
	if err != nil {
		return fmt.Errorf("failed to drop email column: %w", err)
	}

	return nil
}
