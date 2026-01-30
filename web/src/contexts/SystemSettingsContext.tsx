/**
 * 系统设置上下文
 * 提供 login_enabled 等设置，供 Header、登录/注册页判断是否显示登录入口
 */
import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";

interface SystemSettings {
  login_enabled: boolean;
}

interface SystemSettingsContextValue {
  /** 是否允许登录/注册，undefined 表示尚未加载 */
  loginEnabled: boolean | undefined;
  /** 是否正在加载 */
  isLoading: boolean;
  /** 重新拉取系统设置（如修改「关闭登陆」后刷新） */
  refreshSettings: () => Promise<void>;
}

const SystemSettingsContext = createContext<SystemSettingsContextValue | null>(null);

const SYSTEM_SETTINGS_URL = "/api/v1/system/settings";

export function SystemSettingsProvider({ children }: { children: ReactNode }) {
  const [loginEnabled, setLoginEnabled] = useState<boolean | undefined>(undefined);
  const [isLoading, setIsLoading] = useState(true);

  const fetchSettings = useCallback(async () => {
    try {
      const res = await fetch(SYSTEM_SETTINGS_URL, { credentials: "include" });
      const data: SystemSettings = res.ok ? await res.json() : { login_enabled: true };
      setLoginEnabled(!!data.login_enabled);
    } catch {
      setLoginEnabled(true);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  return (
    <SystemSettingsContext.Provider value={{ loginEnabled, isLoading, refreshSettings: fetchSettings }}>
      {children}
    </SystemSettingsContext.Provider>
  );
}

export function useSystemSettings() {
  const ctx = useContext(SystemSettingsContext);
  if (!ctx) {
    throw new Error("useSystemSettings must be used within SystemSettingsProvider");
  }
  return ctx;
}
