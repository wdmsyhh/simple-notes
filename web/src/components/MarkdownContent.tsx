/**
 * MarkdownContent 组件
 * 用于渲染 Markdown 内容，支持：
 * - GitHub Flavored Markdown (GFM)
 * - 代码块语法高亮
 * - 任务列表
 * - HTML 标签（经过安全过滤）
 */
import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import rehypeRaw from 'rehype-raw';
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize';
import type { Element } from 'hast';
import { CodeBlock } from './CodeBlock';

/**
 * MarkdownContent 组件的属性接口
 */
interface MarkdownContentProps {
  /** 要渲染的 Markdown 内容 */
  content: string;
}

/**
 * 检测是否是任务列表项的 checkbox
 * @param node - HTML 元素节点
 * @returns 如果是任务列表项的 checkbox 则返回 true
 */
function isTaskListItemNode(node: Element | undefined): boolean {
  if (!node) return false;
  const type = (node.properties as any)?.type;
  return typeof type === 'string' && type === 'checkbox';
}

/**
 * 安全过滤配置
 * 允许 HTML 标签如 <font>，但过滤不安全的标签和属性
 */
const sanitizeSchema = {
  ...defaultSchema,
  tagNames: [
    ...(defaultSchema.tagNames || []),
    'font', // 允许 font 标签
  ],
  attributes: {
    ...defaultSchema.attributes,
    font: ['face', 'size', 'color'], // 允许 font 标签的属性
  },
};

const MarkdownContent: React.FC<MarkdownContentProps> = ({ content }) => {
  return (
    <div className="markdown-content">
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkBreaks]}
        rehypePlugins={[rehypeRaw, [rehypeSanitize, sanitizeSchema]]}
        components={{
          // 自定义代码块渲染，支持语法高亮
          code: ((codeProps: React.ComponentProps<'code'> & { node?: Element }) => {
            const { className, children, ...props } = codeProps;
            const isInline = !className || !className.includes('language-');
            
            if (isInline) {
              // 内联代码
              return <code className="markdown-inline-code" {...props}>{children}</code>;
            }
            
            // 代码块 - CodeBlock 会自己处理 pre 标签
            return <CodeBlock className={className} {...props}>{children}</CodeBlock>;
          }) as React.ComponentType<React.ComponentProps<'code'> & { node?: Element }>,
          
          // 自定义 pre 标签，配合 code 组件使用
          pre: ((preProps: React.ComponentProps<'pre'>) => {
            const { children, ...props } = preProps;
            // 检查 children 是否是 CodeBlock
            if (React.isValidElement(children)) {
              const childType = children.type as any;
              if (childType?.displayName === 'CodeBlock' || childType === CodeBlock) {
                // CodeBlock 已经包含了 pre，直接返回
                return <>{children}</>;
              }
            }
            // 否则使用默认渲染
            return <pre {...props}>{children}</pre>;
          }) as React.ComponentType<React.ComponentProps<'pre'>>,
          
          // 自定义复选框的渲染，确保任务列表项正确显示
          input: ((inputProps: React.ComponentProps<'input'> & { node?: Element }) => {
            const { node, checked, type, ...props } = inputProps;
            
            // 如果是任务列表项的 checkbox
            if (type === 'checkbox' && isTaskListItemNode(node)) {
              return (
                <input
                  type="checkbox"
                  checked={checked}
                  disabled
                  {...props}
                />
              );
            }
            
            // 其他类型的 input
            return <input type={type} {...props} />;
          }) as React.ComponentType<React.ComponentProps<'input'> & { node?: Element }>,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};

export default MarkdownContent;
