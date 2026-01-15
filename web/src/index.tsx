/**
 * 应用程序入口文件
 * 负责初始化 React 应用并挂载到 DOM
 */
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.css';

// 获取根 DOM 元素
const rootElement = document.getElementById('root');

// 检查根元素是否存在
if (!rootElement) {
  throw new Error('Failed to find the root element');
}

// 创建 React 根节点
const root = ReactDOM.createRoot(rootElement);

// 渲染应用组件，使用严格模式以帮助发现潜在问题
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);