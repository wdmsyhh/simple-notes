/**
 * 访问令牌状态管理
 * 使用 sessionStorage 实现跨页面刷新的持久化存储
 */

/** 内存中的访问令牌 */
let accessToken: string | null = null;
/** 令牌过期时间 */
let tokenExpiresAt: Date | null = null;

/** sessionStorage 中存储令牌的键名 */
const SESSION_TOKEN_KEY = "simple_notes_access_token";
/** sessionStorage 中存储过期时间的键名 */
const SESSION_EXPIRES_KEY = "simple_notes_token_expires_at";

/**
 * 获取访问令牌
 * 如果内存中没有，则尝试从 sessionStorage 恢复
 * @returns 访问令牌，如果不存在或已过期则返回 null
 */
export const getAccessToken = (): string | null => {
  // 如果内存中没有，尝试从 sessionStorage 恢复
  if (!accessToken) {
    try {
      const storedToken = sessionStorage.getItem(SESSION_TOKEN_KEY);
      const storedExpires = sessionStorage.getItem(SESSION_EXPIRES_KEY);

      if (storedToken && storedExpires) {
        const expiresAt = new Date(storedExpires);
        // 只有在令牌未过期时才恢复
        if (expiresAt > new Date()) {
          accessToken = storedToken;
          tokenExpiresAt = expiresAt;
        } else {
          // 令牌已过期，清理 sessionStorage
          sessionStorage.removeItem(SESSION_TOKEN_KEY);
          sessionStorage.removeItem(SESSION_EXPIRES_KEY);
        }
      }
    } catch (e) {
      // sessionStorage 可能不可用（例如在某些隐私模式下）
      console.warn("Failed to access sessionStorage:", e);
    }
  }
  return accessToken;
};

/**
 * 设置访问令牌
 * @param token - 访问令牌，如果为 null 则清除令牌
 * @param expiresAt - 令牌过期时间（可选）
 */
export const setAccessToken = (token: string | null, expiresAt?: Date): void => {
  accessToken = token;
  tokenExpiresAt = expiresAt || null;

  try {
    if (token && expiresAt) {
      // 存储到 sessionStorage 以实现跨页面刷新的持久化
      sessionStorage.setItem(SESSION_TOKEN_KEY, token);
      sessionStorage.setItem(SESSION_EXPIRES_KEY, expiresAt.toISOString());
    } else {
      // 如果要清除令牌，同时清除 sessionStorage
      sessionStorage.removeItem(SESSION_TOKEN_KEY);
      sessionStorage.removeItem(SESSION_EXPIRES_KEY);
    }
  } catch (e) {
    // sessionStorage 可能不可用（例如在某些隐私模式下）
    console.warn("Failed to write to sessionStorage:", e);
  }
};

/**
 * 检查令牌是否已过期
 * 为了安全起见，在实际过期时间前 30 秒就认为已过期
 * @returns 如果令牌已过期或不存在则返回 true
 */
export const isTokenExpired = (): boolean => {
  if (!tokenExpiresAt) return true;
  // 为了安全起见，在实际过期时间前 30 秒就认为已过期
  return new Date() >= new Date(tokenExpiresAt.getTime() - 30000);
};

/**
 * 清除访问令牌
 * 清除内存和 sessionStorage 中的令牌信息
 */
export const clearAccessToken = (): void => {
  accessToken = null;
  tokenExpiresAt = null;

  try {
    sessionStorage.removeItem(SESSION_TOKEN_KEY);
    sessionStorage.removeItem(SESSION_EXPIRES_KEY);
  } catch (e) {
    console.warn("Failed to clear sessionStorage:", e);
  }
};

