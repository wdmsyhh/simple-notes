package v1

import (
	"net/http"

	"connectrpc.com/connect"

	apiv1connect "github.com/wdmsyhh/simple-notes/proto/gen/api/v1/apiv1connect"
)

// ConnectServiceHandler 包装 APIV1Service 以实现 Connect 处理器接口
// 它将现有的 gRPC 服务实现适配为使用 Connect 的请求/响应包装类型
//
// 这种包装模式允许我们：
// - 重用现有的 gRPC 服务实现
// - 同时支持原生 gRPC 和 Connect 协议
// - 保持业务逻辑的单一来源
type ConnectServiceHandler struct {
	*APIV1Service
}

// NewConnectServiceHandler 创建一个新的 Connect 服务处理器
func NewConnectServiceHandler(svc *APIV1Service) *ConnectServiceHandler {
	return &ConnectServiceHandler{APIV1Service: svc}
}

// RegisterConnectHandlers 在给定的多路复用器上注册所有 Connect 服务处理器
func (s *ConnectServiceHandler) RegisterConnectHandlers(mux *http.ServeMux, opts ...connect.HandlerOption) {
	// 直接注册所有服务处理器
	mux.Handle(apiv1connect.NewNoteServiceHandler(s, opts...))
	mux.Handle(apiv1connect.NewCategoryServiceHandler(s, opts...))
	mux.Handle(apiv1connect.NewTagServiceHandler(s, opts...))
	mux.Handle(apiv1connect.NewUserServiceHandler(s, opts...))
	mux.Handle(apiv1connect.NewAttachmentServiceHandler(s, opts...))
}

// wrap 将 (path, handler) 返回值转换为结构体，以便更清晰地迭代
func wrap(path string, handler http.Handler) struct {
	path    string
	handler http.Handler
} {
	return struct {
		path    string
		handler http.Handler
	}{path, handler}
}
