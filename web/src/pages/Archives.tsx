/**
 * Archives 页面组件
 * 显示所有已发布笔记的归档，按年份分组
 */
import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { noteServiceClient } from '../connect';
import { create } from '@bufbuild/protobuf';
import { ListNotesRequestSchema } from '../types/proto/api/v1/note_service_pb';
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
  /** 创建日期（格式：YYYY/MM/DD） */
  createdAt: string;
}

/**
 * ArchiveGroup 接口
 * 按年份分组的归档数据结构
 */
interface ArchiveGroup {
  /** 年份 */
  year: string;
  /** 该年份的笔记列表 */
  notes: NoteItem[];
}

const Archives: React.FC = () => {
  /** 归档数据，按年份分组 */
  const [archiveData, setArchiveData] = useState<ArchiveGroup[]>([]);
  /** 加载状态 */
  const [loading, setLoading] = useState(true);
  /** 错误信息 */
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    /**
     * 从 API 获取所有已发布的笔记并按年份分组
     */
    const fetchArchives = async () => {
      try {
        setLoading(true);
        setError(null);

        // 获取所有已发布的笔记（使用较大的 pageSize 获取所有数据）
        const notesRequest = create(ListNotesRequestSchema, {
          page: 1,
          pageSize: 1000, // 使用较大的值获取所有笔记
          categoryId: '',
          tagId: '',
          search: '',
          sortBy: 'created_at',
          sortDesc: true, // 按创建时间降序：最新的在前
        });

        const notesResponse = await noteServiceClient.listNotes(notesRequest);
        
        // 将 Note 转换为 NoteItem 格式
        const convertedNotes: NoteItem[] = (notesResponse.notes || []).map((note: Note) => {
          // 将 bigint id 转换为 number
          const noteId = Number(note.id);
          // 将时间戳（秒）转换为日期字符串（格式：YYYY/MM/DD）
          const createdAtTimestamp = note.createdAt 
            ? Number(note.createdAt) * 1000
            : Date.now();
          const date = new Date(createdAtTimestamp);
          const year = date.getFullYear();
          const month = String(date.getMonth() + 1).padStart(2, '0');
          const day = String(date.getDate()).padStart(2, '0');
          const createdAt = `${year}/${month}/${day}`;
          
          return {
            id: noteId,
            title: note.title || '未命名笔记',
            slug: note.slug || (noteId ? `note-${noteId}` : ''),
            createdAt: createdAt,
          };
        });

        // 按年份分组
        const grouped = convertedNotes.reduce((acc, note) => {
          const year = note.createdAt.substring(0, 4);

          // 查找年份分组
          let yearGroup = acc.find(group => group.year === year);
          if (!yearGroup) {
            yearGroup = { year, notes: [] };
            acc.push(yearGroup);
          }

          // 将笔记添加到年份分组
          yearGroup.notes.push(note);

          return acc;
        }, [] as ArchiveGroup[]);

        // 按年份降序排序
        grouped.sort((a, b) => parseInt(b.year) - parseInt(a.year));

        // 在每个年份分组内，按日期降序排序
        grouped.forEach(yearGroup => {
          yearGroup.notes.sort((a, b) => {
            return new Date(b.createdAt.replace(/\//g, '-')).getTime() - 
                   new Date(a.createdAt.replace(/\//g, '-')).getTime();
          });
        });

        setArchiveData(grouped);
      } catch (error) {
        console.error('Failed to fetch archives:', error);
        const errorMessage = error instanceof Error ? error.message : '获取归档数据失败';
        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    fetchArchives();
  }, []);

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  if (error) {
    return (
      <div className="archives">
        <h1 className="archives-title">文章归档</h1>
        <div className="error-message" style={{ padding: '20px', background: '#fee', color: '#c33', margin: '20px' }}>
          <p>错误: {error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="archives">
      <h1 className="archives-title">文章归档</h1>
      <div className="archives-content">
        {archiveData.length === 0 ? (
          <div style={{ padding: '20px', textAlign: 'center', color: '#999' }}>
            暂无归档数据
          </div>
        ) : (
          archiveData.map(yearGroup => (
            <div key={yearGroup.year} className="archive-year">
              <h2 className="year-title">{yearGroup.year}</h2>
              <ul className="post-list">
                {yearGroup.notes.map(note => (
                  <li key={note.id} className="post-item">
                    <Link to={`/note/${note.id}`}>
                      <span className="post-date">{note.createdAt}</span>
                      <span className="post-link">{note.title}</span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Archives;