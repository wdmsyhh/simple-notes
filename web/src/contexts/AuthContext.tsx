/**
 * 认证上下文（AuthContext）
 * 提供全局的认证状态管理和相关操作
 */
import { createContext, useContext, useState, useCallback, ReactNode, useEffect } from "react";
import { userServiceClient } from "../connect";
import { clearAccessToken, getAccessToken, setAccessToken } from "../auth-state";
import type { User } from "../types/proto/store/note_pb";

/**
 * 认证状态接口
 */
interface AuthState {
  /** 当前登录的用户 */
  currentUser: User | undefined;
  /** 是否已完成初始化 */
  isInitialized: boolean;
  /** 是否正在加载 */
  isLoading: boolean;
}

/**
 * 认证上下文值接口
 * 扩展了 AuthState，添加了操作方法
 */
interface AuthContextValue extends AuthState {
  /** 初始化认证状态 */
  initialize: () => Promise<void>;
  /** 登出 */
  logout: () => Promise<void>;
  /** 登录 */
  login: (token: string, expiresAt?: Date) => Promise<void>;
}

/** 创建认证上下文 */
const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * AuthProvider 组件
 * 提供认证上下文给子组件
 * @param children - 子组件
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    currentUser: undefined,
    isInitialized: false,
    isLoading: true,
  });

  /**
   * 初始化认证状态
   * 检查是否有有效的访问令牌，如果有则获取当前用户信息
   */
  const initialize = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      const token = getAccessToken();
      if (!token) {
        // 没有令牌，设置为未登录状态
        setState({
          currentUser: undefined,
          isInitialized: true,
          isLoading: false,
        });
        return;
      }

      // 使用令牌获取当前用户信息
      const currentUser = await userServiceClient.getCurrentUser({});

      setState({
        currentUser,
        isInitialized: true,
        isLoading: false,
      });
    } catch (error) {
      // 初始化失败，清除无效的令牌
      console.error("Failed to initialize auth:", error);
      clearAccessToken();
      setState({
        currentUser: undefined,
        isInitialized: true,
        isLoading: false,
      });
    }
  }, []);

  /**
   * 登录
   * @param token - 访问令牌
   * @param expiresAt - 令牌过期时间（可选）
   */
  const login = useCallback(async (token: string, expiresAt?: Date) => {
    setAccessToken(token, expiresAt);
    await initialize();
  }, [initialize]);

  /**
   * 登出
   * 清除令牌和用户信息
   */
  const logout = useCallback(async () => {
    clearAccessToken();
    setState({
      currentUser: undefined,
      isInitialized: true,
      isLoading: false,
    });
  }, []);

  // 组件挂载时初始化认证状态
  useEffect(() => {
    initialize();
  }, [initialize]);

  return (
    <AuthContext.Provider
      value={{
        ...state,
        initialize,
        logout,
        login,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

/**
 * useAuth Hook
 * 用于在组件中访问认证上下文
 * @returns 认证上下文值
 * @throws 如果不在 AuthProvider 内使用则抛出错误
 */
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}

