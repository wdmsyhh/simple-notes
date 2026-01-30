# Simple Notes

一个简洁优雅的笔记管理系统，支持 Markdown 编辑、分类管理、标签系统等功能。

## 项目介绍

Simple Notes 是一个全栈笔记应用，采用前后端分离架构。后端使用 Go 语言开发，提供高性能的 API 服务；前端使用 React + TypeScript 构建，提供现代化的用户界面。

### 主要功能

- 📝 **笔记管理**：支持 Markdown 格式的笔记创建、编辑、删除
- 📁 **分类管理**：为笔记添加分类，方便组织和管理
- 🏷️ **标签系统**：使用标签对笔记进行分类和检索
- 📎 **附件管理**：支持上传和管理笔记附件

### 技术栈

**后端：**
- Go 1.25+
- Echo Web Framework
- ConnectRPC / gRPC
- Protocol Buffers
- SQLite / MySQL / PostgreSQL
- JWT 认证

**前端：**
- React 18
- TypeScript
- Vite
- React Router
- ConnectRPC Web Client
- React Markdown

## 环境要求

- Go 1.25 或更高版本
- Node.js 18+ 和 npm（用于前端开发）
- SQLite（默认）或 MySQL / PostgreSQL（可选）

## 安装和运行

### 1. 克隆项目

```bash
git clone git@github.com:wdmsyhh/simple-notes.git
cd simple-notes
```

### 2. 后端运行

#### 安装依赖

```bash
go mod download
```

#### 运行服务器

**使用默认配置（SQLite）：**

```bash
go run cmd/notes/main.go
```

**使用命令行参数：**

```bash
# 指定端口
go run cmd/notes/main.go --port 3000

# 使用 MySQL
go run cmd/notes/main.go --db-driver mysql --db-dsn "user:password@tcp(localhost:3306)/simple_notes"

# 使用 PostgreSQL
go run cmd/notes/main.go --db-driver postgres --db-dsn "host=localhost user=postgres password=password dbname=simple_notes sslmode=disable"
```

**使用环境变量：**

```bash
export NOTES_PORT=3000
export NOTES_DB_DRIVER=sqlite
export NOTES_DB_DSN=./data/simple-notes.db
go run cmd/notes/main.go
```

**编译并运行：**

```bash
# 编译（当前平台）
go build -o notes cmd/notes/main.go

# 指定 Linux 平台构建（如在 macOS/Windows 下交叉编译用于 Linux 部署）
GOOS=linux GOARCH=amd64 go build -o notes cmd/notes/main.go

# 查询系统架构
uname -m
# 若系统架构是amd64（也叫 x86_64），属于英特尔 / AMD 的 64 位 x86 架构，不是 arm64（aarch64）
# 若目标为 ARM64（如部分云服务器）
# GOOS=linux GOARCH=arm64 go build -o notes cmd/notes/main.go

# 运行
./notes --port 8080
```

**后台运行（nohup）：**

关闭终端后仍保持运行，可将输出重定向到日志文件：

```bash
# 后台运行，输出写入 notes.log
nohup ./notes --port 8010 > notes.log 2>&1 &

# 查看进程
ps aux | grep notes
# 或按端口查看
lsof -i :8010

# 停止进程（将 <PID> 替换为实际进程 ID）
kill <PID>
```

### 3. 前端运行

#### 开发模式

```bash
cd web
npm install
npm run dev
```

前端开发服务器将在 `http://localhost:5173` 启动（Vite 默认端口）。

#### 生产构建

```bash
cd web
npm install
npm run release
```

构建完成后，前端文件将输出到 `server/router/frontend/dist` 目录，后端会自动提供静态文件服务。

### 4. 访问应用

- **开发模式**：前端 `http://localhost:5173`，后端 API `http://localhost:8080`
- **生产模式**：访问 `http://localhost:8080`（前后端集成）

## 配置说明

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--port` | 服务器监听端口 | 8080 |
| `--db-driver` | 数据库驱动类型（sqlite/mysql/postgres） | sqlite |
| `--db-dsn` | 数据库连接字符串 | ./data/simple-notes.db |

### 环境变量

所有配置项也可以通过环境变量设置，环境变量前缀为 `NOTES_`：

- `NOTES_PORT`：服务器端口
- `NOTES_DB_DRIVER`：数据库驱动
- `NOTES_DB_DSN`：数据库连接字符串

### 数据库配置示例

**SQLite（默认）：**
```
./data/simple-notes.db
```

**MySQL：**
```
user:password@tcp(localhost:3306)/simple_notes
```

**PostgreSQL：**
```
host=localhost user=postgres password=password dbname=simple_notes sslmode=disable
```

## 项目结构

```
simple-notes/
├── cmd/notes/          # 应用程序入口
├── internal/           # 内部工具包
│   ├── profile/        # 配置管理
│   ├── util/           # 工具函数
│   └── version/         # 版本信息
├── proto/              # Protocol Buffers 定义
│   ├── api/v1/         # API 服务定义
│   └── store/          # 数据模型定义
├── server/              # 服务器相关
│   ├── auth/           # 认证模块
│   └── router/          # 路由处理
│       ├── api/v1/     # API 路由
│       ├── fileserver/ # 文件服务
│       └── frontend/   # 前端静态文件
├── service/            # 业务逻辑层
├── store/               # 数据存储层
│   └── db/             # 数据库驱动
└── web/                 # 前端应用
    └── src/
        ├── components/ # React 组件
        ├── pages/      # 页面组件
        └── utils/      # 工具函数
```

## 开发说明

### 生成 Protocol Buffers 代码

```bash
cd proto
buf generate
```

### 数据库迁移

数据库表结构会在首次启动时自动创建，无需手动迁移。

