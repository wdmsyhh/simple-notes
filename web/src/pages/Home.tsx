/**
 * Home 页面组件
 * 公开首页，显示所有公开的笔记列表
 * 支持：
 * - 按标签过滤
 * - 分页浏览
 * - 显示分类和标签信息
 */
import React, { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import MarkdownContent from '../components/MarkdownContent';
import Sidebar from '../components/Sidebar';
import { noteServiceClient, categoryServiceClient, tagServiceClient } from '../connect';
import { create } from '@bufbuild/protobuf';
import { ListNotesRequestSchema } from '../types/proto/api/v1/note_service_pb';
import { ListCategoriesRequestSchema } from '../types/proto/api/v1/category_service_pb';
import { ListTagsRequestSchema } from '../types/proto/api/v1/tag_service_pb';
import type { Note } from '../types/proto/store/note_pb';
import type { Category } from '../types/proto/store/note_pb';
import type { Tag } from '../types/proto/store/note_pb';

/**
 * NoteItem 接口
 * 笔记项的数据结构
 */
interface NoteItem {
  /** 笔记ID */
  id: number;
  /** 标题 */
  title: string;
  /** URL友好的标识符 */
  slug: string;
  /** 摘要 */
  summary: string;
  /** 分类ID */
  category: string;
  /** 标签ID列表 */
  tagIds: string[];
  /** 创建日期（字符串格式） */
  createdAt: string;
  /** 封面图片URL（可选） */
  coverImage?: string;
  /** 阅读时间（分钟） */
  readingTime: number;
  /** 浏览次数 */
  viewCount: number;
}

const Home: React.FC = () => {
  const [searchParams] = useSearchParams();
  const rawTagId = searchParams.get('tagId');
  
  /**
   * 验证ID是否为有效的正整数
   * @param value - 要验证的值
   * @returns 如果是有效的正整数则返回 true
   */
  const isValidPositiveId = (value?: string | null): value is string => {
    if (!value) return false;
    const num = Number(value);
    return Number.isFinite(num) && num > 0;
  };
  
  /** 从URL参数中获取标签ID */
  const tagId = isValidPositiveId(rawTagId) ? rawTagId : '';
  /** 笔记列表 */
  const [notes, setNotes] = useState<NoteItem[]>([]);
  /** 分类列表 */
  const [categories, setCategories] = useState<Category[]>([]);
  /** 标签列表 */
  const [tags, setTags] = useState<Tag[]>([]);
  /** 是否正在加载 */
  const [loading, setLoading] = useState(true);
  /** 错误信息 */
  const [error, setError] = useState<string | null>(null);
  /** 当前页码 */
  const [currentPage, setCurrentPage] = useState(1);
  /** 总页数 */
  const [totalPages, setTotalPages] = useState(1);
  /** 每页显示的笔记数量 */
  const pageSize = 10;

  // 当标签ID改变时，重置到第一页
  useEffect(() => {
    setCurrentPage(1);
  }, [tagId]);

  // 获取分类和标签（用于在笔记项中显示分类/标签名称）
  useEffect(() => {
    const fetchMetaData = async () => {
      try {
        // 获取分类
        const categoriesRequest = create(ListCategoriesRequestSchema, {
          includeHidden: false,
          parentId: BigInt(0),
        });
        const categoriesResponse = await categoryServiceClient.listCategories(categoriesRequest);
        setCategories(categoriesResponse.categories || []);
        
        // 获取标签
        const tagsRequest = create(ListTagsRequestSchema, {
          limit: 20,
          offset: 0,
        });
        const tagsResponse = await tagServiceClient.listTags(tagsRequest);
        setTags(tagsResponse.tags || []);
      } catch (error) {
        console.error('Failed to fetch meta data:', error);
      }
    };
    fetchMetaData();
  }, []);

  // 当页码或标签ID改变时，获取笔记列表
  useEffect(() => {
    /**
     * 获取笔记列表
     */
    const fetchNotes = async () => {
      try {
        setLoading(true);
        
        // 获取笔记
        const notesRequest = create(ListNotesRequestSchema, {
          page: currentPage,
          pageSize: pageSize,
          categoryId: '',
          tagId: tagId,
          search: '',
          sortBy: 'created_at',
          sortDesc: true, // 按创建时间降序：最新的在前
        });

        const notesResponse = await noteServiceClient.listNotes(notesRequest);
        
        // 将 Note 转换为 NoteItem 格式
        const convertedNotes: NoteItem[] = (notesResponse.notes || []).map((note: Note) => {
          // 将 bigint id 转换为 number
          const noteId = Number(note.id);
          // 将时间戳（秒）转换为日期字符串
          const createdAt = note.createdAt 
            ? new Date(Number(note.createdAt) * 1000).toISOString().split('T')[0]
            : new Date().toISOString().split('T')[0];
          
          return {
            id: noteId,
            title: note.title || '',
            slug: note.slug || (noteId ? `note-${noteId}` : ''),
            summary: note.summary || '',
            category: note.categoryId || '',
            tagIds: note.tagIds || [],
            createdAt: createdAt,
            coverImage: note.coverImage || undefined,
            readingTime: note.readingTime || 0,
            viewCount: note.viewCount || 0,
          };
        });

        setNotes(convertedNotes);
        
        // 从API响应中设置总页数
        const pages = notesResponse.totalPages || 1;
        setTotalPages(pages);

        setError(null);
      } catch (error) {
        console.error('Failed to fetch data:', error);
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        setError(errorMessage);
        // 发生错误时设置空数组以显示空状态
        setNotes([]);
      } finally {
        setLoading(false);
      }
    };

    fetchNotes();
  }, [currentPage, tagId]);

  /**
   * 处理页码变化
   * @param newPage - 新的页码
   */
  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setCurrentPage(newPage);
    }
  };

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  return (
    <div className="home">
      {error && (
        <div className="error-message" style={{ padding: '20px', background: '#fee', color: '#c33', margin: '20px' }}>
          <p>错误: {error}</p>
        </div>
      )}
      <div className="home-content">
        <div className="posts-list">
          {notes.length === 0 ? (
            <div className="empty-state">
              <p>暂无文章</p>
            </div>
          ) : (
            notes.map(note => (
            <article key={note.id} className="post-item">
              {note.coverImage && (
                <div className="post-cover">
                  <img src={note.coverImage} alt={note.title} />
                </div>
              )}
              <div className="post-content">
                <div className="post-meta">
                  <span className="date">{note.createdAt}</span>
                  {(() => {
                    const category = note.category ? categories.find(cat => String(cat.id) === note.category) : null;
                    return (
                      <span className="category">
                        分类：{category ? (category.nameText || '未命名分类') : '无'}
                      </span>
                    );
                  })()}
                  <span className="tags">
                    标签：{note.tagIds && note.tagIds.length > 0 ? (
                      note.tagIds.map(tagId => {
                        const tag = tags.find(t => String(t.id) === tagId);
                        if (!tag) return null;
                        return (
                          <Link key={tagId} to={`/?tagId=${tag.id}`} className="tag">
                            #{tag.nameText || '未命名标签'}
                          </Link>
                        );
                      }).filter(Boolean)
                    ) : (
                      '无'
                    )}
                  </span>
                </div>
                <h2 className="post-title">
                  <Link to={note.id ? `/note/${note.id}` : '#'} onClick={(e) => {
                    if (!note.id) {
                      e.preventDefault();
                      alert('笔记链接无效');
                    }
                  }}>{note.title}</Link>
                </h2>
                <div className="post-summary">
                  {note.summary ? (
                    <MarkdownContent content={note.summary} />
                  ) : (
                    <p style={{ color: '#999', fontStyle: 'italic' }}>暂无描述</p>
                  )}
                </div>

              </div>
            </article>
            ))
          )}
        </div>
        <Sidebar />
      </div>
      {totalPages > 1 && (
      <div className="pagination">
          <button 
            className="page-btn" 
            disabled={currentPage === 1}
            onClick={() => handlePageChange(currentPage - 1)}
          >
            上一页
          </button>
          <span className="page-info">第 {currentPage} 页，共 {totalPages} 页</span>
          <button 
            className="page-btn" 
            disabled={currentPage === totalPages}
            onClick={() => handlePageChange(currentPage + 1)}
          >
            下一页
          </button>
      </div>
      )}
    </div>
  );
};

export default Home;