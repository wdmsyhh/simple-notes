// v1 包包含 API V1 版本的服务实现
package v1

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	"github.com/wdmsyhh/simple-notes/service"
	"github.com/wdmsyhh/simple-notes/store"
)

// APIV1Service 是 API V1 版本的服务实现结构体
// 实现了 gRPC 服务接口和 Connect 服务接口

type APIV1Service struct {
	// 未实现的 NoteService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedNoteServiceServer
	// 未实现的 CategoryService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedCategoryServiceServer
	// 未实现的 TagService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedTagServiceServer
	// 未实现的 UserService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedUserServiceServer
	// 未实现的 AttachmentService 服务器（用于 gRPC 兼容性）
	apiv1.UnimplementedAttachmentServiceServer

	// 数据存储实例，用于数据库操作
	Store *store.Store
	// 用户服务实例，用于处理用户相关业务逻辑
	userService *service.UserService
	// Secret 用于 JWT token 签名
	Secret string
}

// NewAPIV1Service 创建一个新的 APIV1Service 实例
func NewAPIV1Service(store *store.Store, secret string) *APIV1Service {
	// 创建用户服务实例
	userService := service.NewUserService(store)

	return &APIV1Service{
		Store:       store,
		userService: userService,
		Secret:      secret,
	}
}

// RegisterGateway 注册 gRPC-Gateway 和 Connect 处理器到给定的 Echo 实例
func (s *APIV1Service) RegisterGateway(ctx context.Context, echoServer *echo.Echo) error {
	// 创建 gRPC-Gateway 多路复用器
	gwMux := runtime.NewServeMux()

	// 注册 NoteService 处理服务器
	if err := apiv1.RegisterNoteServiceHandlerServer(ctx, gwMux, s); err != nil {
		return err
	}

	// 注册 CategoryService 处理服务器
	if err := apiv1.RegisterCategoryServiceHandlerServer(ctx, gwMux, s); err != nil {
		return err
	}

	// 注册 TagService 处理服务器
	if err := apiv1.RegisterTagServiceHandlerServer(ctx, gwMux, s); err != nil {
		return err
	}

	// 注册 UserService 处理服务器
	if err := apiv1.RegisterUserServiceHandlerServer(ctx, gwMux, s); err != nil {
		return err
	}

	// 注册 AttachmentService 处理服务器
	if err := apiv1.RegisterAttachmentServiceHandlerServer(ctx, gwMux, s); err != nil {
		return err
	}

	// 创建 API 网关路由组
	gwGroup := echoServer.Group("")
	// 添加 CORS 中间件
	gwGroup.Use(middleware.CORS())
	// 将 gRPC-Gateway 多路复用器包装为 Echo 处理器
	handler := echo.WrapHandler(gwMux)

	// 注册所有 API V1 路径
	gwGroup.Any("/api/v1/*", handler)

	// 为浏览器客户端创建 Connect 处理器
	connectInterceptors := connect.WithInterceptors(
		NewMetadataInterceptor(),
		NewLoggingInterceptor(true), // 启用日志记录以进行调试
		NewAuthInterceptor(s.Store, s.Secret),
	)

	// 配置 Connect 处理器选项，支持大文件上传（32MB）
	const maxMessageSize = 32 << 20 // 32MB
	connectHandlerOptions := []connect.HandlerOption{
		connectInterceptors,
		connect.WithReadMaxBytes(maxMessageSize),
		connect.WithSendMaxBytes(maxMessageSize),
		// 添加恢复处理器以正确处理 panic 而不关闭连接
		connect.WithRecover(func(ctx context.Context, spec connect.Spec, header http.Header, p any) error {
			log.Printf("Panic recovered in %s: %v", spec.Procedure, p)
			return connect.NewError(connect.CodeInternal, fmt.Errorf("internal server error: %v", p))
		}),
	}

	connectMux := http.NewServeMux()
	// 创建 Connect 服务处理器
	connectHandler := NewConnectServiceHandler(s)
	// 注册 Connect 处理器
	connectHandler.RegisterConnectHandlers(connectMux, connectHandlerOptions...)

	// 为 Connect 处理器添加 CORS 支持
	corsHandler := middleware.CORSWithConfig(middleware.CORSConfig{
		// 允许所有来源
		AllowOriginFunc: func(_ string) (bool, error) {
			return true, nil
		},
		// 允许的 HTTP 方法
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		// 允许的 HTTP 头部
		AllowHeaders: []string{"*"},
		// 允许携带凭证
		AllowCredentials: true,
	})

	// 创建 Connect 路由组
	connectGroup := echoServer.Group("", corsHandler)
	// 注册所有 Connect 服务路径
	// Connect 路径格式: /package.Service/Method (例如: /api.v1.NoteService/ListNotes)
	connectGroup.Any("/api.v1.*", echo.WrapHandler(connectMux))

	return nil
}
