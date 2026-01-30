/**
 * Header 组件
 * 应用程序的顶部导航栏，包含：
 * - Logo 和导航链接
 * - 用户信息（如果已登录）：下拉菜单（设置、登出）
 * - 登录按钮（未登录且允许登录时）
 */
import React, { useState, useRef, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useSystemSettings } from '../contexts/SystemSettingsContext';
import { UserRole } from '../types/proto/store/note_pb';
import { getAccessToken } from '../auth-state';
import './Header.css';

const SYSTEM_SETTINGS_URL = '/api/v1/system/settings';

const Header: React.FC = () => {
  const { currentUser, logout } = useAuth();
  const { loginEnabled, refreshSettings } = useSystemSettings();
  const navigate = useNavigate();
  const location = useLocation();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [settingsLoginEnabled, setSettingsLoginEnabled] = useState(true);
  const [settingsSaving, setSettingsSaving] = useState(false);
  const [settingsError, setSettingsError] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const isHost = currentUser?.role === UserRole.HOST;

  // 显示用户区域：允许登录时显示，或当前已登录（HOST 在关闭登录后仍可看到菜单）
  const showUserArea = loginEnabled !== false || !!currentUser;

  const handleLogout = async () => {
    setDropdownOpen(false);
    await logout();
    navigate("/");
  };

  const handleHomeClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    if (location.pathname === '/') {
      e.preventDefault();
      window.location.href = '/';
    }
  };

  const openSettings = () => {
    setDropdownOpen(false);
    setSettingsLoginEnabled(loginEnabled !== false);
    setSettingsError(null);
    setSettingsOpen(true);
  };

  const saveSettings = async () => {
    setSettingsSaving(true);
    setSettingsError(null);
    const token = getAccessToken();
    if (!token) {
      setSettingsError('未登录');
      setSettingsSaving(false);
      return;
    }
    try {
      const res = await fetch(SYSTEM_SETTINGS_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        credentials: 'include',
        body: JSON.stringify({ login_enabled: settingsLoginEnabled }),
      });
      const data = await res.json();
      if (!res.ok) {
        setSettingsError(data?.error || '保存失败');
        setSettingsSaving(false);
        return;
      }
      await refreshSettings();
      setSettingsOpen(false);
    } catch (e) {
      setSettingsError('网络错误');
    } finally {
      setSettingsSaving(false);
    }
  };

  useEffect(() => {
    const onOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener('click', onOutside);
    return () => document.removeEventListener('click', onOutside);
  }, []);

  return (
    <header className="header">
      <div className="container">
        <div className="header-content">
          <div className="logo">
            <Link to="/" onClick={handleHomeClick}>Simple Notes</Link>
          </div>
          <nav className="nav">
            <ul>
              <li><Link to="/" onClick={handleHomeClick}>首页</Link></li>
              <li><Link to="/archives">归档</Link></li>
              <li><Link to="/page/about">关于</Link></li>
            </ul>
          </nav>
          {showUserArea && (
            <div className="header-actions">
              {currentUser ? (
                <div className={`header-user-dropdown ${dropdownOpen ? 'is-open' : ''}`} ref={dropdownRef}>
                  <button
                    type="button"
                    className="header-user-trigger"
                    onClick={() => setDropdownOpen((o) => !o)}
                    aria-expanded={dropdownOpen}
                    aria-haspopup="true"
                  >
                    <span className="user-name">{currentUser.nickname || currentUser.username}</span>
                    <span className="header-user-caret" aria-hidden>▼</span>
                  </button>
                  {dropdownOpen && (
                    <ul className="header-user-menu" role="menu">
                      {isHost && (
                        <li role="none">
                          <button type="button" className="header-user-menu-item" role="menuitem" onClick={openSettings}>
                            设置
                          </button>
                        </li>
                      )}
                      <li role="none">
                        <button type="button" className="header-user-menu-item" role="menuitem" onClick={handleLogout}>
                          登出
                        </button>
                      </li>
                    </ul>
                  )}
                </div>
              ) : (
                <Link to="/login" className="login-link">登录</Link>
              )}
            </div>
          )}
        </div>
      </div>

      {/* 设置弹框：仅 HOST 可关闭/开启登录 */}
      {settingsOpen && (
        <div className="header-settings-overlay" role="dialog" aria-modal="true" aria-labelledby="settings-title">
          <div className="header-settings-modal">
            <h2 id="settings-title" className="header-settings-title">设置</h2>
            <div className="header-settings-row">
              <label className="header-settings-label">
                <input
                  type="checkbox"
                  checked={settingsLoginEnabled}
                  onChange={(e) => setSettingsLoginEnabled(e.target.checked)}
                />
                <span>允许登录与注册</span>
              </label>
              <p className="header-settings-hint">关闭后，仅系统管理员（首个注册用户）可登录。</p>
            </div>
            {settingsError && <p className="header-settings-error">{settingsError}</p>}
            <div className="header-settings-actions">
              <button type="button" className="header-settings-btn header-settings-btn-secondary" onClick={() => setSettingsOpen(false)}>
                取消
              </button>
              <button type="button" className="header-settings-btn header-settings-btn-primary" onClick={saveSettings} disabled={settingsSaving}>
                {settingsSaving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </header>
  );
};

export default Header;
