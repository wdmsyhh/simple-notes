/**
 * Login 页面组件
 * 用户登录页面，处理用户认证
 */
import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { ConnectError } from "@connectrpc/connect";
import { userServiceClient } from "../connect";
import { create } from "@bufbuild/protobuf";
import { LoginUserRequestSchema } from "../types/proto/api/v1/user_service_pb";
import { useAuth } from "../contexts/AuthContext";
import "./Login.css";

const Login: React.FC = () => {
  const navigate = useNavigate();
  const { currentUser, login, isInitialized } = useAuth();
  /** 用户名输入 */
  const [username, setUsername] = useState("");
  /** 密码输入 */
  const [password, setPassword] = useState("");
  /** 是否正在提交 */
  const [loading, setLoading] = useState(false);
  /** 错误信息 */
  const [error, setError] = useState<string | null>(null);

  // 如果已经登录，重定向到首页
  useEffect(() => {
    if (isInitialized && currentUser) {
      navigate("/");
    }
  }, [currentUser, isInitialized, navigate]);

  /**
   * 处理表单提交
   * @param e - 表单提交事件
   */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const request = create(LoginUserRequestSchema, {
        username,
        password,
      });

      const response = await userServiceClient.loginUser(request);

      if (response.token) {
        // 计算过期时间（从现在起1天后）
        const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000);
        await login(response.token, expiresAt);
        navigate("/");
      } else {
        setError("登录失败：未收到认证令牌");
      }
    } catch (err: any) {
      console.error("Login error:", err);
      // 提取 ConnectError 的错误信息
      let errorMessage = "登录失败，请检查用户名和密码";
      if (err instanceof ConnectError) {
        // ConnectError 的 message 可能包含 "[unknown] rpc error: code = Unauthenticated desc = 用户名或密码错误"
        // 尝试提取 desc 部分
        const message = err.message || "";
        // 匹配 "desc = 用户名或密码错误" 格式，desc 后面直到字符串末尾的内容
        const descMatch = message.match(/desc\s*=\s*(.+)$/);
        if (descMatch && descMatch[1]) {
          errorMessage = descMatch[1].trim();
        } else {
          // 如果没有匹配到，尝试直接使用 message，但清理格式
          errorMessage = message.replace(/^\[unknown\]\s*rpc error:\s*code\s*=\s*\w+\s*desc\s*=\s*/i, "").trim() || errorMessage;
        }
      } else if (err?.message) {
        errorMessage = err.message;
      }
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  if (!isInitialized) {
    return <div className="loading">加载中...</div>;
  }

  if (currentUser) {
    return null; // Will redirect
  }

  return (
    <div className="login-page">
      <div className="login-container">
        <h1 className="login-title">Simple Notes</h1>
        <h2 className="login-subtitle">登录</h2>
        <form onSubmit={handleSubmit} className="login-form">
          {error && <div className="error-message">{error}</div>}
          <div className="form-group">
            <label htmlFor="username">用户名</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={loading}
              autoComplete="username"
            />
          </div>
          <div className="form-group">
            <label htmlFor="password">密码</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={loading}
              autoComplete="current-password"
            />
          </div>
          <button type="submit" className="login-button" disabled={loading}>
            {loading ? "登录中..." : "登录"}
          </button>
        </form>
        <div className="signup-link">
          <span>还没有账号？</span>
          <Link to="/signup" className="signup-link-text">立即注册</Link>
        </div>
      </div>
    </div>
  );
};

export default Login;

