package profile

// Profile 是启动服务器的配置
type Profile struct {
	// Driver 是数据库驱动类型 (sqlite, mysql, postgres)
	Driver string
	// DSN 是数据库连接字符串
	DSN string
}
