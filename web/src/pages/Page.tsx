import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

interface PageType {
  id: number;
  title: string;
  content: string;
  createdAt: string;
  updatedAt: string;
}

const Page: React.FC = () => {
  const { slug } = useParams<{ slug: string }>();
  const [page, setPage] = useState<PageType | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Fetch page by slug
    const fetchPage = async () => {
      try {
        // Replace with actual API call
        const mockPage: PageType = {
          id: 1,
          title: "关于我们",
          content: `<h2>简单笔记</h2><p>简单笔记是一个轻量级的博客系统，使用 Go 和 React 技术栈开发。</p><h2>技术栈</h2><ul><li>后端：Go + gRPC</li><li>前端：React + TypeScript</li><li>数据库：SQLite</li></ul>`,
          createdAt: "2023-12-01",
          updatedAt: "2023-12-01"
        };

        setPage(mockPage);
      } catch (error) {
        console.error('Failed to fetch page:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchPage();
  }, [slug]);

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  if (!page) {
    return <div className="not-found">页面不存在</div>;
  }

  return (
    <div className="page-detail">
      <article className="page">
        <div className="page-header">
          <h1 className="page-title">{page.title}</h1>
          <div className="page-meta">
            <span className="created-at">创建时间：{page.createdAt}</span>
            <span className="updated-at">更新时间：{page.updatedAt}</span>
          </div>
        </div>
        <div className="page-body" dangerouslySetInnerHTML={{ __html: page.content }} />
      </article>
    </div>
  );
};

export default Page;