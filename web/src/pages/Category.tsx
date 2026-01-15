import React, { useEffect, useState } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import MarkdownContent from '../components/MarkdownContent';
import Sidebar from '../components/Sidebar';
import { categoryServiceClient, noteServiceClient, tagServiceClient } from '../connect';
import { create } from '@bufbuild/protobuf';
import { GetCategoryRequestSchema } from '../types/proto/api/v1/category_service_pb';
import { ListNotesRequestSchema } from '../types/proto/api/v1/note_service_pb';
import { ListTagsRequestSchema } from '../types/proto/api/v1/tag_service_pb';
import type { Category } from '../types/proto/store/note_pb';
import type { Note } from '../types/proto/store/note_pb';
import type { Tag } from '../types/proto/store/note_pb';

interface NoteItem {
  id: number;
  title: string;
  slug: string;
  summary: string;
  tagIds: string[];
  createdAt: string;
  readingTime: number;
  viewCount: number;
}

const Category: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const rawTagId = searchParams.get('tagId');
  const isValidPositiveId = (value?: string | null): value is string => {
    if (!value) return false;
    const num = Number(value);
    return Number.isFinite(num) && num > 0;
  };
  const tagId = isValidPositiveId(rawTagId) ? rawTagId : '';
  const [category, setCategory] = useState<Category | null>(null);
  const [tags, setTags] = useState<Tag[]>([]);
  const [notes, setNotes] = useState<NoteItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const pageSize = 10;

  useEffect(() => {
    // Reset page when id or tagId changes
    setCurrentPage(1);
    
    // Fetch category (only once when id changes)
    const fetchCategory = async () => {
      if (!id) {
        // If no category ID, we can still filter by tag
        setCategory(null);
        setLoading(false);
        return;
      }

      try {
        const categoryId = parseInt(id, 10);
        if (isNaN(categoryId) || categoryId <= 0) {
          // Invalid category id, allow tag-only filtering without error
          setCategory(null);
          setLoading(false);
          return;
        }
        
        const categoryRequest = create(GetCategoryRequestSchema, {
          name: `categories/${categoryId}`,
        });
        const fetchedCategory = await categoryServiceClient.getCategory(categoryRequest);
        setCategory(fetchedCategory);
      } catch (error) {
        console.error('Failed to fetch category:', error);
        const errorMessage = error instanceof Error ? error.message : '加载分类失败';
        setError(errorMessage);
        setCategory(null);
      }
    };

    fetchCategory();
  }, [id]);

  useEffect(() => {
    // Fetch notes when category, tagId, or page changes
    const fetchNotes = async () => {
      // Allow fetching notes even without category if tagId is provided
      if (!category && !tagId) return;

      try {
        setLoading(true);
        setError(null);

        const categoryId = category?.id ? String(category.id) : '';
        const notesRequest = create(ListNotesRequestSchema, {
          page: currentPage,
          pageSize: pageSize,
          categoryId: categoryId,
          tagId: tagId,
          search: '',
          sortBy: 'created_at',
          sortDesc: true, // 按创建时间降序：最新的在前
        });

        const notesResponse = await noteServiceClient.listNotes(notesRequest);
        
        // Convert Note to NoteItem format
        const convertedNotes: NoteItem[] = (notesResponse.notes || []).map((note: Note) => {
          const noteId = Number(note.id);
          const createdAt = note.createdAt 
            ? new Date(Number(note.createdAt) * 1000).toISOString().split('T')[0]
            : new Date().toISOString().split('T')[0];
          
          return {
            id: noteId,
            title: note.title || '',
            slug: note.slug || (noteId ? `note-${noteId}` : ''),
            summary: note.summary || '',
            tagIds: note.tagIds || [],
            createdAt: createdAt,
            readingTime: note.readingTime || 0,
            viewCount: note.viewCount || 0,
          };
        });

        setNotes(convertedNotes);
        
        // Set total pages from API response
        const pages = notesResponse.totalPages || 1;
        setTotalPages(pages);
      } catch (error) {
        console.error('Failed to fetch notes:', error);
        const errorMessage = error instanceof Error ? error.message : '加载文章列表失败';
        setError(errorMessage);
        setNotes([]);
      } finally {
        setLoading(false);
      }
    };

    fetchNotes();
  }, [category, currentPage, tagId]);

  // Fetch tags for post meta display (needed for showing tag names in post items)
  useEffect(() => {
    const fetchTags = async () => {
      try {
        const tagsRequest = create(ListTagsRequestSchema, {
          limit: 20,
          offset: 0,
        });
        const tagsResponse = await tagServiceClient.listTags(tagsRequest);
        setTags(tagsResponse.tags || []);
      } catch (error) {
        console.error('Failed to fetch tags:', error);
      }
    };

    fetchTags();
  }, []);

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setCurrentPage(newPage);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  // Allow rendering even without category if tagId is provided
  if (error && !tagId) {
    return <div className="not-found">{error || '分类不存在'}</div>;
  }
  
  if (!category && !tagId) {
    return <div className="not-found">分类不存在</div>;
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
              <p>
                {tagId && category 
                  ? '该分类和标签下暂无文章' 
                  : tagId 
                  ? '该标签下暂无文章'
                  : '该分类下暂无文章'}
              </p>
            </div>
          ) : (
            notes.map(note => (
              <article key={note.id} className="post-item">
            <div className="post-content">
              <div className="post-meta">
                    <span className="date">{note.createdAt}</span>
                    <span className="category">
                      分类：{category ? (category.nameText || '未命名分类') : '无'}
                    </span>
                    <span className="tags">
                      标签：{note.tagIds && note.tagIds.length > 0 ? (
                        note.tagIds.map(tagId => {
                          const tag = tags.find(t => String(t.id) === tagId);
                          if (!tag) return null;
                          const tagLink = id 
                            ? `/category/${id}?tagId=${tag.id}` 
                            : `/?tagId=${tag.id}`;
                          return (
                            <Link key={tagId} to={tagLink} className="tag">
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
                    <Link to={`/note/${note.id}`}>{note.title}</Link>
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
        <Sidebar activeCategoryId={id} />
      </div>
      <div className="pagination">
        <button 
          className="page-btn" 
          disabled={currentPage === 1 || totalPages <= 1}
          onClick={() => handlePageChange(currentPage - 1)}
        >
          上一页
        </button>
        <span className="page-info">第 {currentPage} 页，共 {totalPages} 页</span>
        <button 
          className="page-btn" 
          disabled={currentPage === totalPages || totalPages <= 1}
          onClick={() => handlePageChange(currentPage + 1)}
        >
          下一页
        </button>
      </div>
    </div>
  );
};

export default Category;
