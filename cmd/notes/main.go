package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wdmsyhh/simple-notes/internal/profile"
	"github.com/wdmsyhh/simple-notes/internal/version"
	"github.com/wdmsyhh/simple-notes/server"
	"github.com/wdmsyhh/simple-notes/store"
	"github.com/wdmsyhh/simple-notes/store/db"
)

var (
	// rootCmd 根命令，用于启动 Simple Notes 服务器
	rootCmd = &cobra.Command{
		Use:   "notes",
		Short: "Simple Notes server",
		Run: func(_ *cobra.Command, _ []string) {
			log.Printf("正在启动 Simple Notes v%s", version.Version)

			instanceProfile := &profile.Profile{
				Driver: viper.GetString("db-driver"),
				DSN:    viper.GetString("db-dsn"),
			}

			dbDriverInstance, err := db.NewDBDriver(instanceProfile)
			if err != nil {
				log.Fatalf("初始化数据库驱动失败: %v", err)
			}
			defer dbDriverInstance.Close()

			storeInstance := store.NewStore(dbDriverInstance, instanceProfile)
			if err := storeInstance.RunMigrations(); err != nil {
				log.Fatalf("运行数据库迁移失败: %v", err)
			}
			// 初次启动时插入允许登录的默认数据到 system_settings
			if err := storeInstance.EnsureDefaultSystemSettings(context.Background()); err != nil {
				log.Printf("Warning: 确保系统设置默认值失败: %v", err)
			}

			port := viper.GetInt("port")
			s := server.NewServer(storeInstance, instanceProfile, port)

			ctx := context.Background()
			if err := s.SetupRoutes(ctx); err != nil {
				log.Fatalf("设置路由失败: %v", err)
			}

			go func() {
				if err := s.Start(); err != nil {
					log.Fatalf("启动服务器失败: %v", err)
				}
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			log.Println("正在关闭服务器...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.Shutdown(ctx); err != nil {
				log.Fatalf("服务器强制关闭: %v", err)
			}

			log.Println("服务器已成功停止")
		},
	}
)

// init 初始化命令行参数和配置
func init() {
	// 设置默认配置值
	viper.SetDefault("port", 8080)
	viper.SetDefault("db-driver", "sqlite")
	viper.SetDefault("db-dsn", "./data/simple-notes.db")

	// 定义命令行标志
	rootCmd.PersistentFlags().Int("port", 8080, "服务器端口")
	rootCmd.PersistentFlags().String("db-driver", "sqlite", "数据库驱动 (sqlite 或 mysql)")
	rootCmd.PersistentFlags().String("db-dsn", "./data/simple-notes.db", "数据库连接字符串")

	// 绑定命令行标志到 viper
	if err := viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("db-driver", rootCmd.PersistentFlags().Lookup("db-driver")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("db-dsn", rootCmd.PersistentFlags().Lookup("db-dsn")); err != nil {
		panic(err)
	}

	// 设置环境变量前缀和自动读取环境变量
	viper.SetEnvPrefix("notes")
	viper.AutomaticEnv()
	if err := viper.BindEnv("db-driver", "NOTES_DB_DRIVER"); err != nil {
		panic(err)
	}
	if err := viper.BindEnv("db-dsn", "NOTES_DB_DSN"); err != nil {
		panic(err)
	}
	if err := viper.BindEnv("port", "NOTES_PORT"); err != nil {
		panic(err)
	}
}

// main 应用程序入口点
func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
