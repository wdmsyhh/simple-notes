/**
 * Sidebar 组件
 * 侧边栏组件，显示：
 * - 分类列表（带笔记数量）
 * - 标签云
 * - 最新文章列表
 */
import React, { useEffect, useState, memo } from 'react';
import { Link, useSearchParams, useParams } from 'react-router-dom';
import { categoryServiceClient, tagServiceClient, noteServiceClient } from '../connect';
import { create } from '@bufbuild/protobuf';
import { ListCategoriesRequestSchema } from '../types/proto/api/v1/category_service_pb';
import { ListTagsRequestSchema } from '../types/proto/api/v1/tag_service_pb';
import { ListNotesRequestSchema } from '../types/proto/api/v1/note_service_pb';
import type { Category } from '../types/proto/store/note_pb';
import type { Tag } from '../types/proto/store/note_pb';
import type { Note } from '../types/proto/store/note_pb';

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

/**
 * Sidebar 组件的属性接口
 */
interface SidebarProps {
  /** 当前激活的分类ID（可选） */
  activeCategoryId?: string;
}

/**
 * Sidebar 组件
 * 使用 memo 优化性能，避免不必要的重新渲染
 * @param activeCategoryId - 当前激活的分类ID
 */
const Sidebar: React.FC<SidebarProps> = memo(({ activeCategoryId }) => {
  const [searchParams] = useSearchParams();
  const params = useParams<{ id?: string }>();
  const rawCategoryId = activeCategoryId || params.id;
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
  
  /** 当前分类ID（已验证） */
  const categoryId = isValidPositiveId(rawCategoryId) ? rawCategoryId : undefined;
  /** 当前标签ID（已验证） */
  const tagId = isValidPositiveId(rawTagId) ? rawTagId : '';
  /** 分类列表 */
  const [categories, setCategories] = useState<Category[]>([]);
  /** 每个分类的笔记数量 */
  const [categoryCounts, setCategoryCounts] = useState<Record<string, number>>({});
  /** 标签列表 */
  const [tags, setTags] = useState<Tag[]>([]);
  /** 最新笔记列表 */
  const [latestNotes, setLatestNotes] = useState<NoteItem[]>([]);

  // 获取分类列表（仅执行一次）
  useEffect(() => {
    /**
     * 获取分类列表
     */
    const fetchCategories = async () => {
      try {
        const categoriesRequest = create(ListCategoriesRequestSchema, {
          includeHidden: false,
          parentId: BigInt(0),
        });
        const categoriesResponse = await categoryServiceClient.listCategories(categoriesRequest);
        const fetchedCategories = categoriesResponse.categories || [];
        // 按创建时间升序排序（最早的在最前，最新的在最后）
        const sortedCategories = [...fetchedCategories].sort((a, b) => {
          const aTime = Number(a.createdAt || 0);
          const bTime = Number(b.createdAt || 0);
          return aTime - bTime;
        });
        setCategories(sortedCategories);

        // 获取每个分类的笔记数量
        const counts: Record<string, number> = {};
        for (const category of sortedCategories) {
          try {
            const notesRequest = create(ListNotesRequestSchema, {
              page: 1,
              pageSize: 1,
              categoryId: String(category.id),
              tagId: '',
              search: '',
              sortBy: 'created_at',
              sortDesc: true,
            });
            const notesResponse = await noteServiceClient.listNotes(notesRequest);
            counts[String(category.id)] = notesResponse.total || 0;
          } catch (error) {
            console.error(`Failed to fetch count for category ${category.id}:`, error);
            counts[String(category.id)] = 0;
          }
        }
        setCategoryCounts(counts);
      } catch (error) {
        console.error('Failed to fetch categories:', error);
      }
    };

    fetchCategories();
  }, []);

  // 获取标签列表（仅执行一次）
  useEffect(() => {
    /**
     * 获取标签列表
     */
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

  // 获取最新3篇笔记用于侧边栏显示（仅执行一次）
  useEffect(() => {
    /**
     * 获取最新笔记列表
     */
    const fetchLatestNotes = async () => {
      try {
        const notesRequest = create(ListNotesRequestSchema, {
          page: 1,
          pageSize: 3,
          categoryId: '',
          tagId: '',
          search: '',
          sortBy: 'created_at',
          sortDesc: true,
        });
        const notesResponse = await noteServiceClient.listNotes(notesRequest);
        
        // 将 Note 转换为 NoteItem 格式
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
            category: note.categoryId || '',
            tagIds: note.tagIds || [],
            createdAt: createdAt,
            coverImage: note.coverImage || undefined,
            readingTime: note.readingTime || 0,
            viewCount: note.viewCount || 0,
          };
        });

        setLatestNotes(convertedNotes);
      } catch (error) {
        console.error('Failed to fetch latest notes:', error);
      }
    };

    fetchLatestNotes();
  }, []);

  return (
    <aside className="sidebar">
      <div className="sidebar-section">
        <h3>分类</h3>
        <ul className="category-list">
          {categories.length === 0 ? (
            <li>暂无分类</li>
          ) : (
            categories.map((category) => {
              const count = categoryCounts[String(category.id)] || 0;
              const isActive = activeCategoryId === String(category.id);
              return (
                <li key={category.id}>
                  <Link 
                    to={`/category/${category.id}`}
                    style={{
                      color: isActive ? '#47698C' : '#333',
                    }}
                  >
                    {category.nameText || '未命名分类'} ({count})
                  </Link>
                </li>
              );
            })
          )}
        </ul>
      </div>
      <div className="sidebar-section">
        <h3>标签</h3>
        <div className="tag-cloud">
          {tags.length === 0 ? (
            <span style={{ color: '#999', fontStyle: 'italic' }}>暂无标签</span>
          ) : (
            tags.map((tag) => {
              const tagIdStr = typeof tag.id === 'bigint' ? tag.id.toString() : String(tag.id);
              if (!isValidPositiveId(tagIdStr)) {
                return null;
              }
              const isActive = tagId === tagIdStr;
              const tagLink = categoryId
                ? `/category/${categoryId}?tagId=${tagIdStr}`
                : `/?tagId=${tagIdStr}`;
              return (
                <Link 
                  key={tag.id} 
                  to={tagLink}
                  style={{
                    backgroundColor: isActive ? '#47698C' : '#f0f0f0',
                    color: isActive ? 'white' : '#666',
                  }}
                >
                  {tag.nameText || '未命名标签'}
                </Link>
              );
            })
          )}
        </div>
      </div>
      <div className="sidebar-section">
        <h3>最新文章</h3>
        <ul className="latest-posts">
          {latestNotes.length === 0 ? (
            <li style={{ color: '#999', fontStyle: 'italic' }}>暂无文章</li>
          ) : (
            latestNotes.map((note) => (
              <li key={note.id}>
                <Link 
                  to={note.id ? `/note/${note.id}` : '#'} 
                  onClick={(e) => {
                    if (!note.id) {
                      e.preventDefault();
                      alert('笔记链接无效');
                    }
                  }}
                >
                  {note.title}
                </Link>
              </li>
            ))
          )}
        </ul>
      </div>
    </aside>
  );
});

Sidebar.displayName = 'Sidebar';

export default Sidebar;
