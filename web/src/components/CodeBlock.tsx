/**
 * CodeBlock 组件
 * 用于渲染代码块，支持语法高亮
 * 使用 highlight.js 进行代码高亮
 */
import React, { useMemo } from 'react';
import hljs from 'highlight.js';
import 'highlight.js/styles/github.css';

/**
 * CodeBlock 组件的属性接口
 */
interface CodeBlockProps {
  /** 代码内容 */
  children?: React.ReactNode;
  /** CSS 类名，通常包含语言信息（如 "language-javascript"） */
  className?: string;
}

/**
 * 从 react-markdown 的 code 元素中提取代码内容
 * @param children - React 子节点
 * @returns 提取的代码字符串
 */
const extractCodeContent = (children: React.ReactNode): string => {
  // 在 react-markdown 中，children 可以是：
  // 1. 字符串（直接文本内容）
  // 2. 带有 children 属性的 React 元素
  if (typeof children === 'string') {
    return children;
  }
  if (React.isValidElement(children)) {
    const codeElement = children as React.ReactElement;
    const childrenContent = codeElement?.props?.children;
    if (typeof childrenContent === 'string') {
      return childrenContent;
    }
    return String(childrenContent || '').replace(/\n$/, '');
  }
  return String(children || '').replace(/\n$/, '');
};

/**
 * 从 className 中提取语言名称
 * 格式： "language-xxx"
 * @param className - CSS 类名
 * @returns 提取的语言名称
 */
const extractLanguage = (className: string): string => {
  const match = /language-(\w+)/.exec(className);
  return match ? match[1] : '';
};

export const CodeBlock: React.FC<CodeBlockProps> = ({ children, className, ...props }) => {
  // In react-markdown, children is the code content as string
  // className contains "language-xxx" format
  const codeContent = extractCodeContent(children);
  const language = extractLanguage(className || '');

  // Highlight code using highlight.js
  const highlightedCode = useMemo(() => {
    if (!language) {
      // No language specified, just escape HTML
      return Object.assign(document.createElement('span'), {
        textContent: codeContent,
      }).innerHTML;
    }

    try {
      const lang = hljs.getLanguage(language);
      if (lang) {
        return hljs.highlight(codeContent, {
          language: language,
        }).value;
      }
    } catch {
      // If highlighting fails, just escape HTML
    }

    // Fallback: escape HTML entities
    return Object.assign(document.createElement('span'), {
      textContent: codeContent,
    }).innerHTML;
  }, [language, codeContent]);

  return (
    <pre className="markdown-code-block" {...props}>
      {language && <span className="code-language">{language}</span>}
      <code
        className={language ? `language-${language}` : ''}
        dangerouslySetInnerHTML={{ __html: highlightedCode }}
      />
    </pre>
  );
};

CodeBlock.displayName = 'CodeBlock';
