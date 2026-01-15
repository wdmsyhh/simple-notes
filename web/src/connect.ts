/**
 * Connect RPC 客户端配置
 * 负责创建和管理与后端的 gRPC-Web 连接
 */
import { Code, ConnectError, createClient, type Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { NoteService } from "./types/proto/api/v1/note_service_pb";
import { CategoryService } from "./types/proto/api/v1/category_service_pb";
import { TagService } from "./types/proto/api/v1/tag_service_pb";
import { UserService } from "./types/proto/api/v1/user_service_pb";
import { AttachmentService } from "./types/proto/api/v1/attachment_service_pb";
import { getAccessToken, setAccessToken } from "./auth-state";

// ============================================================================
// 常量定义
// ============================================================================

/** 重试请求头名称 */
const RETRY_HEADER = "X-Retry";
/** 重试请求头值 */
const RETRY_HEADER_VALUE = "true";

// ============================================================================
// 认证拦截器
// ============================================================================

/**
 * 认证拦截器
 * 功能：
 * 1. 自动在请求头中添加 Bearer token
 * 2. 处理认证失败的情况（token 过期或无效时清除 token）
 * 3. 防止无限重试循环
 */
const authInterceptor: Interceptor = (next) => async (req) => {
  // 获取访问令牌
  const token = getAccessToken();
  if (token) {
    // 在请求头中添加认证信息
    req.header.set("Authorization", `Bearer ${token}`);
  }

  try {
    return await next(req);
  } catch (error) {
    // 如果不是 ConnectError，直接抛出
    if (!(error instanceof ConnectError)) {
      throw error;
    }

    // 如果不是未认证错误，直接抛出
    if (error.code !== Code.Unauthenticated) {
      throw error;
    }

    // 如果已经重试过，避免无限循环
    if (req.header.get(RETRY_HEADER) === RETRY_HEADER_VALUE) {
      throw error;
    }

    // Token 过期或无效，清除它
    setAccessToken(null);
    throw error;
  }
};

// ============================================================================
// 传输配置
// ============================================================================

/**
 * 带凭证的 fetch 函数
 * 确保在跨域请求时包含 cookies 等凭证信息
 */
const fetchWithCredentials: typeof globalThis.fetch = (input, init) => {
  return globalThis.fetch(input, {
    ...init,
    credentials: "include",
  });
};

/**
 * 创建 Connect 传输层
 * 配置：
 * - baseUrl: 使用当前窗口的源地址
 * - useBinaryFormat: 使用 JSON 格式（false）
 * - fetch: 使用带凭证的 fetch
 * - interceptors: 添加认证拦截器
 */
const transport = createConnectTransport({
  baseUrl: window.location.origin,
  useBinaryFormat: false,
  fetch: fetchWithCredentials,
  interceptors: [authInterceptor],
});

// ============================================================================
// 服务客户端
// ============================================================================

/** 笔记服务客户端 */
export const noteServiceClient = createClient(NoteService, transport);
/** 分类服务客户端 */
export const categoryServiceClient = createClient(CategoryService, transport);
/** 标签服务客户端 */
export const tagServiceClient = createClient(TagService, transport);
/** 用户服务客户端 */
export const userServiceClient = createClient(UserService, transport);
/** 附件服务客户端 */
export const attachmentServiceClient = createClient(AttachmentService, transport);

