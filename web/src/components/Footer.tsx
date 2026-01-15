/**
 * Footer 组件
 * 应用程序的页脚，显示版权信息
 */
import React from 'react';
import './Footer.css';

const Footer: React.FC = () => {
  return (
    <footer className="footer">
      <div className="container">
        <div className="footer-content">
          <p>© Simple Notes.</p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;