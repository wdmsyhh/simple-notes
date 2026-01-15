/**
 * Header 组件
 * 应用程序的顶部导航栏，包含：
 * - Logo 和导航链接
 * - 用户信息（如果已登录）
 * - 登录/登出按钮
 */
import React from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import './Header.css';

const Header: React.FC = () => {
  const { currentUser, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  /**
   * 处理登出操作
   * 登出后重定向到首页
   */
  const handleLogout = async () => {
    await logout();
    navigate("/");
  };

  /**
   * 处理首页点击事件
   * 如果在首页，强制完整页面刷新
   * @param e - 鼠标事件
   */
  const handleHomeClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    // 如果在首页，强制完整页面刷新
    if (location.pathname === '/') {
      e.preventDefault();
      window.location.href = '/';
    }
  };

  return (
    <header className="header">
      <div className="container">
        <div className="header-content">
          {/* Logo */}
          <div className="logo">
            <Link to="/" onClick={handleHomeClick}>Simple Notes</Link>
          </div>
          {/* 导航菜单 */}
          <nav className="nav">
            <ul>
              <li><Link to="/" onClick={handleHomeClick}>首页</Link></li>
              <li><Link to="/archives">归档</Link></li>
              <li><Link to="/page/about">关于</Link></li>
            </ul>
          </nav>
          {/* 用户操作区域 */}
          <div className="header-actions">
            {currentUser ? (
              <>
                {/* 显示用户昵称或用户名 */}
                <span className="user-name">{currentUser.nickname || currentUser.username}</span>
                {/* 登出按钮 */}
                <button onClick={handleLogout} className="logout-button">登出</button>
              </>
            ) : (
              /* 登录链接 */
              <Link to="/login" className="login-link">登录</Link>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
