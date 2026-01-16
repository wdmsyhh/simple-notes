/**
 * Dashboard 页面组件
 * 用户仪表板，显示用户的笔记列表，支持：
 * - 创建和编辑笔记
 * - 搜索笔记
 * - 分页浏览
 * - 删除笔记
 */
import React, { useEffect, useState, useRef } from "react";
import { Link } from "react-router-dom";
import MarkdownContent from '../components/MarkdownContent';
import { noteServiceClient } from "../connect";
import { create } from "@bufbuild/protobuf";
import { ListNotesRequestSchema, DeleteNoteRequestSchema } from "../types/proto/api/v1/note_service_pb";
import type { Note } from "../types/proto/store/note_pb";
import { NoteVisibility } from "../types/proto/store/note_pb";
import { useAuth } from "../contexts/AuthContext";
import NoteEditor from "../components/NoteEditor";
import "./Dashboard.css";

const Dashboard: React.FC = () => {
  const { currentUser } = useAuth();
  /** 笔记列表 */
  const [notes, setNotes] = useState<Note[]>([]);
  /** 是否正在加载 */
  const [loading, setLoading] = useState(true);
  /** 错误信息 */
  const [error, setError] = useState<string | null>(null);
  /** 正在编辑的笔记ID */
  const [editingNoteId, setEditingNoteId] = useState<string | null>(null);
  /** 当前页码 */
  const [currentPage, setCurrentPage] = useState(1);
  /** 总页数 */
  const [totalPages, setTotalPages] = useState(1);
  /** 搜索查询字符串（实际用于搜索） */
  const [searchQuery, setSearchQuery] = useState("");
  /** 搜索输入框的值（用于输入框的值，点击搜索时才更新 searchQuery） */
  const [searchInput, setSearchInput] = useState("");
  /** 每页显示的笔记数量 */
  const pageSize = 20;
  /** 保存滚动位置的引用 */
  const scrollPositionRef = useRef<number>(0);
  /** 是否正在搜索的引用 */
  const isSearchingRef = useRef<boolean>(false);

  // 调试：记录 editingNoteId 的变化
  useEffect(() => {
    console.log('editingNoteId changed to:', editingNoteId);
  }, [editingNoteId]);

  // 当页码或搜索查询改变时，重新获取笔记列表
  useEffect(() => {
    fetchNotes();
  }, [currentPage, searchQuery]);

  // 恢复滚动位置（仅在搜索后）
  useEffect(() => {
    if (isSearchingRef.current && scrollPositionRef.current > 0) {
      // 使用 setTimeout 确保 DOM 更新完成后再恢复滚动位置
      setTimeout(() => {
        window.scrollTo(0, scrollPositionRef.current);
        isSearchingRef.current = false;
      }, 0);
    }
  }, [notes]);

  /**
   * 获取笔记列表
   * 根据当前页码和搜索查询从服务器获取笔记
   */
  const fetchNotes = async () => {
    try {
      setLoading(true);
      const request = create(ListNotesRequestSchema, {
        page: currentPage,
        pageSize: pageSize,
        categoryId: "",
        tagId: "",
        search: searchQuery,
        sortBy: "created_at",
        sortDesc: true,
      });

      const response = await noteServiceClient.listNotes(request);
      const notesList = response.notes || [];
      const total = response.total || 0;
      const calculatedTotalPages = Math.ceil(total / pageSize);
      
      // 确保最多只显示 pageSize 条数据（防止后端返回过多数据）
      const limitedNotesList = notesList.slice(0, pageSize);
      
      console.log('Fetched notes:', notesList.length, 'Limited to:', limitedNotesList.length);
      console.log('Total:', total, 'Total pages:', calculatedTotalPages);
      
      setNotes(limitedNotesList);
      setTotalPages(calculatedTotalPages);
      setError(null);
    } catch (err: any) {
      console.error("Failed to fetch notes:", err);
      setError(err.message || "获取笔记列表失败");
    } finally {
      setLoading(false);
    }
  };

  /**
   * 处理页码变化
   * @param page - 新的页码
   */
  const handlePageChange = (page: number) => {
    if (page >= 1 && page <= totalPages) {
      setCurrentPage(page);
    }
  };

  /**
   * 处理搜索输入框变化
   * @param e - 输入事件
   */
  const handleSearchInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchInput(e.target.value);
  };

  /**
   * 处理搜索操作
   * 保存当前滚动位置，更新搜索查询，重置到第一页
   */
  const handleSearch = () => {
    // 保存当前滚动位置
    scrollPositionRef.current = window.scrollY || window.pageYOffset || document.documentElement.scrollTop;
    isSearchingRef.current = true;
    setSearchQuery(searchInput);
    setCurrentPage(1); // 搜索时重置到第一页
  };

  /**
   * 处理搜索输入框的键盘事件
   * 按 Enter 键时触发搜索
   * @param e - 键盘事件
   */
  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  /**
   * 处理保存操作
   * 保存后刷新笔记列表
   */
  const handleSave = () => {
    // 保存后刷新笔记列表
    setEditingNoteId(null);
    fetchNotes();
  };

  /**
   * 处理编辑操作
   * @param note - 要编辑的笔记
   */
  const handleEdit = (note: Note) => {
    console.log('Edit button clicked for note:', note);
    console.log('Note name:', note.name);
    console.log('Note id:', note.id);
    
    // 如果 name 缺失，从 id 生成 name
    let noteName = note.name;
    if (!noteName && note.id) {
      noteName = `notes/${note.id}`;
      console.log('Generated note name from id:', noteName);
    }
    
    if (!noteName) {
      console.error('Note name and id are both missing, cannot edit');
      alert('无法编辑：笔记信息不完整');
      return;
    }
    
    console.log('Setting editingNoteId to:', noteName);
    setEditingNoteId(noteName);
  };

  /**
   * 处理取消编辑操作
   */
  const handleCancelEdit = () => {
    setEditingNoteId(null);
  };

  /**
   * 处理删除操作
   * @param note - 要删除的笔记
   */
  const handleDelete = async (note: Note) => {
    if (!note.name) {
      alert("无法删除：笔记名称无效");
      return;
    }

    if (!confirm(`确定要删除笔记 "${note.title || "无标题"}" 吗？此操作无法撤销。`)) {
      return;
    }

    try {
      const request = create(DeleteNoteRequestSchema, {
        name: note.name,
      });
      await noteServiceClient.deleteNote(request);
      
      // 如果正在编辑被删除的笔记，清除编辑状态
      if (editingNoteId === note.name) {
        setEditingNoteId(null);
      }
      
      // 删除后刷新笔记列表
      fetchNotes();
    } catch (error: any) {
      console.error("Failed to delete note:", error);
      alert(`删除失败: ${error.message || "未知错误"}`);
    }
  };

  /**
   * 格式化日期
   * 将时间戳转换为相对时间（如"5分钟前"）或日期字符串
   * @param timestamp - Unix 时间戳（秒）
   * @returns 格式化后的时间字符串
   */
  const formatDate = (timestamp: bigint | undefined) => {
    if (!timestamp) return "";
    const date = new Date(Number(timestamp) * 1000);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 8640000);

    if (minutes < 1) return "刚刚";
    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    if (days < 7) return `${days}天前`;
    return date.toLocaleDateString("zh-CN");
  };

  if (loading) {
    return <div className="dashboard-loading">加载中...</div>;
  }

  return (
    <div className="dashboard">
      <div className="dashboard-container">
        <div className="dashboard-header">
          <h1>我的笔记</h1>
          <p className="dashboard-subtitle">欢迎回来，{currentUser?.username}</p>
        </div>

        {/* Note Editor */}
        {editingNoteId ? (
          <div className="note-editor-section">
            <div className="mb-2 text-sm text-gray-600">编辑笔记 (ID: {editingNoteId})</div>
            <NoteEditor noteId={editingNoteId} onSave={handleSave} onCancel={handleCancelEdit} />
          </div>
        ) : (
          <div className="note-editor-section">
            <NoteEditor onSave={handleSave} />
          </div>
        )}

        {/* Search Bar */}
        <div className="dashboard-search">
          <input
            type="text"
            className="dashboard-search-input"
            placeholder="搜索笔记名称..."
            value={searchInput}
            onChange={handleSearchInputChange}
            onKeyDown={handleSearchKeyDown}
          />
          <button
            type="button"
            className="dashboard-search-button"
            onClick={handleSearch}
          >
            搜索
          </button>
        </div>

        {/* Notes List */}
        {error && (
          <div className="error-message" style={{ padding: "20px", background: "#fee", color: "#c33", margin: "20px 0" }}>
            {error}
          </div>
        )}

        <div className="notes-list">
          {notes.length === 0 ? (
            <div className="empty-state">
              <p>还没有笔记，开始创建你的第一篇笔记吧！</p>
            </div>
          ) : (
            notes.map((note) => (
              <div key={note.id?.toString()} className="note-card">
                <div className="note-header">
                  <h3 className="note-title">
                    <Link to={note.id ? `/note/${note.id}` : '#'} onClick={(e) => {
                      if (!note.id) {
                        e.preventDefault();
                        alert('笔记链接无效');
                      }
                    }}>{note.title || "无标题"}</Link>
                  </h3>
                  <div className="note-meta">
                    <span className="note-visibility">
                      {note.visibility === NoteVisibility.PUBLIC ? "🌐 公开" : "🔒 私有"}
                    </span>
                    <span className="note-time">{formatDate(note.createdAt)}</span>
                  </div>
                </div>
                <div className="note-content">
                  {note.summary ? (
                    <MarkdownContent content={note.summary} />
                  ) : note.content ? (
                    <MarkdownContent content={note.content} />
                  ) : (
                    <p style={{ color: '#999', fontStyle: 'italic' }}>暂无内容</p>
                  )}
                </div>
                {note.tagIds && note.tagIds.length > 0 && (
                  <div className="note-tags">
                    {note.tagIds.map((tagId, idx) => (
                      <span key={idx} className="tag">
                        #{tagId}
                      </span>
                    ))}
                  </div>
                )}
                <div className="note-actions">
                  <a
                    href="#"
                    className="edit-button"
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      console.log('Edit button clicked, note:', note);
                      if (editingNoteId !== note.name) {
                        handleEdit(note);
                      }
                    }}
                  >
                    编辑
                  </a>
                  <a
                    href="#"
                    className="delete-button"
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      handleDelete(note);
                    }}
                  >
                    删除
                  </a>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Pagination */}
        {totalPages > 0 && (
          <div className="dashboard-pagination">
            <button
              className="pagination-button"
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage === 1}
            >
              上一页
            </button>
            <div className="pagination-info">
              第 {currentPage} / {totalPages} 页
            </div>
            <button
              className="pagination-button"
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
            >
              下一页
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Dashboard;

