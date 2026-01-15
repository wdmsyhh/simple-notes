/**
 * 主应用组件
 * 负责设置路由、认证上下文和整体布局
 */
import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Header from './components/Header';
import ProtectedRoute from './components/ProtectedRoute';
import Home from './pages/Home';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';
import SignUp from './pages/SignUp';
import NoteDetail from './pages/NoteDetail';
import Category from './pages/Category';
import Tag from './pages/Tag';
import Page from './pages/Page';
import Archives from './pages/Archives';
import './App.css';

/**
 * RootRoute 组件
 * 根据认证状态处理根路由：
 * - 如果已登录，重定向到仪表板
 * - 如果未登录，显示公开首页
 */
const RootRoute: React.FC = () => {
  const { currentUser, isInitialized } = useAuth();

  // 如果认证状态尚未初始化，显示加载中
  if (!isInitialized) {
    return <div className="loading">加载中...</div>;
  }

  // 如果已登录，重定向到仪表板；否则显示公开首页
  if (currentUser) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Home />;
};

/**
 * App 组件
 * 应用程序的根组件，包含：
 * - 认证提供者（AuthProvider）
 * - 路由配置（Router）
 * - 页面布局（Header 和主内容区）
 */
const App: React.FC = () => {
  return (
    <AuthProvider>
      <Router>
        <div className="app">
          <Header />
          <main className="main-content">
            <div className="container">
              <Routes>
                {/* 根路由，根据认证状态显示不同内容 */}
                <Route path="/" element={<RootRoute />} />
                {/* 仪表板页面，需要认证 */}
                <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
                {/* 登录页面 */}
                <Route path="/login" element={<Login />} />
                {/* 注册页面 */}
                <Route path="/signup" element={<SignUp />} />
                {/* 归档页面 */}
                <Route path="/archives" element={<Archives />} />
                {/* 笔记详情页面 */}
                <Route path="/note/:id" element={<NoteDetail />} />
                {/* 分类页面 */}
                <Route path="/category/:id" element={<Category />} />
                {/* 标签页面 */}
                <Route path="/tag/:slug" element={<Tag />} />
                {/* 页面详情 */}
                <Route path="/page/:slug" element={<Page />} />
              </Routes>
            </div>
          </main>
        </div>
      </Router>
    </AuthProvider>
  );
};

export default App;