/**
 * SignUp 页面组件
 * 用户注册页面，处理新用户注册
 * 注册成功后自动登录
 */
import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { userServiceClient } from "../connect";
import { create } from "@bufbuild/protobuf";
import { RegisterUserRequestSchema } from "../types/proto/api/v1/user_service_pb";
import { LoginUserRequestSchema } from "../types/proto/api/v1/user_service_pb";
import { UserSchema } from "../types/proto/store/note_pb";
import { useAuth } from "../contexts/AuthContext";
import "./SignUp.css";

const SignUp: React.FC = () => {
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

    // 验证密码长度
    if (password.length < 6) {
      setError("密码长度至少为6个字符");
      return;
    }

    if (username === "" || password === "") {
      setError("用户名和密码不能为空");
      return;
    }

    setLoading(true);

    try {
      // 创建用户对象（仅包含用户名，昵称是可选的）
      const user = create(UserSchema, {
        username,
        nickname: username, // 使用用户名作为默认昵称
      });

      // 注册用户
      const registerRequest = create(RegisterUserRequestSchema, {
        user,
        password,
      });

      await userServiceClient.registerUser(registerRequest);

      // 注册成功后自动登录
      const loginRequest = create(LoginUserRequestSchema, {
        username,
        password,
      });

      const loginResponse = await userServiceClient.loginUser(loginRequest);

      if (loginResponse.token) {
        // 计算过期时间（从现在起1天后）
        const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000);
        await login(loginResponse.token, expiresAt);
        navigate("/");
      } else {
        setError("注册成功，但登录失败：未收到认证令牌");
      }
    } catch (err: any) {
      console.error("Signup error:", err);
      setError(err.message || "注册失败，请检查输入信息");
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
    <div className="signup-page">
      <div className="signup-container">
        <h1 className="signup-title">Simple Notes</h1>
        <h2 className="signup-subtitle">注册</h2>
        <form onSubmit={handleSubmit} className="signup-form">
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
              autoCapitalize="off"
              spellCheck={false}
              minLength={3}
              maxLength={50}
              pattern="[a-zA-Z0-9_-]+"
              title="用户名只能包含字母、数字、下划线和连字符"
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
              autoComplete="new-password"
              autoCapitalize="off"
              spellCheck={false}
              minLength={6}
            />
          </div>
          <button type="submit" className="signup-button" disabled={loading}>
            {loading ? "注册中..." : "注册"}
          </button>
        </form>
        <div className="login-link">
          <span>已有账号？</span>
          <Link to="/login" className="login-link-text">立即登录</Link>
        </div>
      </div>
    </div>
  );
};

export default SignUp;

