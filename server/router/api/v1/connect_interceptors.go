package v1

import (
	"context"
	"errors"
	"log"
	"runtime/debug"

	"connectrpc.com/connect"

	"github.com/wdmsyhh/simple-notes/server/auth"
	"github.com/wdmsyhh/simple-notes/store"
)

// NewRecoveryInterceptor 创建一个新的恢复拦截器，用于捕获panic并返回适当的错误
// 参数：
//
//	logStacktraces - 是否记录堆栈跟踪
//
// 返回：
//
//	connect.Interceptor - Connect拦截器
func NewRecoveryInterceptor(logStacktraces bool) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			defer func() {
				if r := recover(); r != nil {
					stackTrace := string(debug.Stack())
					log.Printf("Panic recovered in %s: %v\n%s", req.Spec().Procedure, r, stackTrace)
					// 不要重新 panic，而是返回错误以防止连接关闭
				}
			}()
			return next(ctx, req)
		}
	})
}

// NewMetadataInterceptor 创建一个新的元数据拦截器，用于将HTTP头转换为gRPC元数据
// 返回：
//
//	connect.Interceptor - Connect拦截器
func NewMetadataInterceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return next(ctx, req)
		}
	})
}

// NewLoggingInterceptor 创建一个新的日志拦截器，用于记录请求和响应
// 参数：
//
//	logStacktraces - 是否记录堆栈跟踪
//
// 返回：
//
//	connect.Interceptor - Connect拦截器
func NewLoggingInterceptor(logStacktraces bool) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			log.Printf("Request: %s", req.Spec().Procedure)
			resp, err := next(ctx, req)
			if err != nil {
				log.Printf("Response: %s, Error: %v", req.Spec().Procedure, err)
				if logStacktraces {
					log.Printf("Error stack trace: %s", string(debug.Stack()))
				}
			} else {
				log.Printf("Response: %s, Status: OK", req.Spec().Procedure)
			}
			return resp, err
		}
	})
}

// AuthInterceptor 处理 Connect 处理器的认证
type AuthInterceptor struct {
	// authenticator 认证器实例
	authenticator *auth.Authenticator
}

// NewAuthInterceptor 创建新的认证拦截器
// 参数：
//   store - 数据存储实例
//   secret - JWT 密钥
// 返回：
//   *AuthInterceptor - 认证拦截器实例
func NewAuthInterceptor(store *store.Store, secret string) *AuthInterceptor {
	return &AuthInterceptor{
		authenticator: auth.NewAuthenticator(store, secret),
	}
}

// WrapUnary 包装一元函数以进行认证
// 参数：
//   next - 下一个处理函数
// 返回：
//   connect.UnaryFunc - 包装后的处理函数
func (in *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		header := req.Header()
		authHeader := header.Get("Authorization")

		result := in.authenticator.Authenticate(ctx, authHeader)

		// 对非公开方法强制执行认证
		if result == nil && !IsPublicMethod(req.Spec().Procedure) {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
		}

		// 根据认证结果设置上下文
		if result != nil {
			if result.Claims != nil {
				// 访问令牌 V2 - 无状态，使用声明
				ctx = auth.SetUserClaimsInContext(ctx, result.Claims)
			}
		}

		return next(ctx, req)
	}
}

// WrapStreamingClient 包装流式客户端函数
// 参数：
//   next - 下一个流式客户端函数
// 返回：
//   connect.StreamingClientFunc - 包装后的流式客户端函数
func (*AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler 包装流式处理器函数
// 参数：
//   next - 下一个流式处理器函数
// 返回：
//   connect.StreamingHandlerFunc - 包装后的流式处理器函数
func (*AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
