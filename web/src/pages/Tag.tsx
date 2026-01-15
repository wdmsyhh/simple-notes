import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

interface NoteItem {
  id: number;
  title: string;
  slug: string;
  summary: string;
  createdAt: string;
  readingTime: number;
  viewCount: number;
}

interface TagType {
  id: number;
  name: string;
  noteCount: number;
}

const Tag: React.FC = () => {
  const { slug } = useParams<{ slug: string }>();
  const [tag, setTag] = useState<TagType | null>(null);
  const [notes, setNotes] = useState<NoteItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Fetch tag and its notes
    const fetchTagData = async () => {
      try {
        // Replace with actual API call
        const mockTag: TagType = {
          id: 1,
          name: "Go",
          noteCount: 1
        };

        const mockNotes: NoteItem[] = [
          {
            id: 1,
            title: "Go 语言入门教程",
            slug: "go-language-introduction",
            summary: "Go 语言是一门开源的编程语言，它能让构造简单、可靠且高效的软件变得容易。",
            createdAt: "2023-12-01",
            readingTime: 5,
            viewCount: 1234
          }
        ];

        setTag(mockTag);
        setNotes(mockNotes);
      } catch (error) {
        console.error('Failed to fetch tag data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchTagData();
  }, [slug]);

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  if (!tag) {
    return <div className="not-found">标签不存在</div>;
  }

  return (
    <div className="tag-page">
      <div className="tag-header">
        <h1>标签：{tag.name}</h1>
        <p className="tag-post-count">共 {tag.noteCount} 篇文章</p>
      </div>

      <div className="posts-list">
        {notes.map(note => (
          <article key={note.id} className="post-item">
            <div className="post-content">
              <div className="post-meta">
                <span className="date">{note.createdAt}</span>
                <span className="reading-time">{note.readingTime}分钟</span>
                <span className="view-count">{note.viewCount}阅读</span>
              </div>
              <h2 className="post-title">
                <Link to={`/note/${note.id}`}>{note.title}</Link>
              </h2>
              <p className="post-summary">{note.summary}</p>
              <Link to={`/note/${note.id}`} className="read-more">
                阅读全文 →
              </Link>
            </div>
          </article>
        ))}
      </div>

      <div className="pagination">
        <button disabled className="page-btn">上一页</button>
        <span className="page-info">第 1 页，共 1 页</span>
        <button disabled className="page-btn">下一页</button>
      </div>
    </div>
  );
};

export default Tag;