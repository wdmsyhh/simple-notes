/**
 * ProtectedRoute 组件
 * 受保护的路由组件，用于保护需要认证才能访问的页面
 * 如果用户未登录，将重定向到登录页面
 */
import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

/**
 * ProtectedRoute 组件的属性接口
 */
interface ProtectedRouteProps {
  /** 子组件（需要保护的页面内容） */
  children: React.ReactNode;
}

/**
 * ProtectedRoute 组件
 * @param children - 需要保护的子组件
 * @returns 如果已认证则渲染子组件，否则重定向到登录页
 */
const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { currentUser, isInitialized } = useAuth();

  // 如果认证状态尚未初始化，显示加载中
  if (!isInitialized) {
    return <div className="loading">加载中...</div>;
  }

  // 如果用户未登录，重定向到登录页面
  if (!currentUser) {
    return <Navigate to="/login" replace />;
  }

  // 用户已登录，渲染受保护的内容
  return <>{children}</>;
};

export default ProtectedRoute;

