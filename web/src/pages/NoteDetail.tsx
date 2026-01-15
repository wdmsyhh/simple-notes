import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import MarkdownContent from '../components/MarkdownContent';
import Sidebar from '../components/Sidebar';
import { noteServiceClient, attachmentServiceClient } from '../connect';
import { create } from '@bufbuild/protobuf';
import { GetNoteRequestSchema, DeleteNoteRequestSchema } from '../types/proto/api/v1/note_service_pb';
import { ListAttachmentsRequestSchema } from '../types/proto/api/v1/attachment_service_pb';
import type { Note } from '../types/proto/store/note_pb';
import type { Attachment } from '../types/proto/api/v1/attachment_service_pb';
import { NoteVisibility } from '../types/proto/store/note_pb';
import { useAuth } from '../contexts/AuthContext';
import NoteEditor from '../components/NoteEditor';
import './NoteDetail.css';

const NoteDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { currentUser } = useAuth();
  const [note, setNote] = useState<Note | null>(null);
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [previewImage, setPreviewImage] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);

  const fetchNote = async () => {
    if (!id) {
      console.error('No id provided');
      setError('缺少笔记ID');
      setLoading(false);
      return;
    }

    console.log('Fetching note with id:', id);
      try {
        setLoading(true);
        setError(null);

      // Parse id (could be just number or "note-{id}" format for backward compatibility)
      let noteIdStr = id;
      if (id.startsWith('note-')) {
        noteIdStr = id.replace('note-', '');
      }
      
      const noteId = BigInt(noteIdStr);
      console.log('Using note ID:', noteId);
      
      // Use GetNote to fetch by ID
      const getNoteRequest = create(GetNoteRequestSchema, {
        name: `notes/${noteId}`,
        });
      const fetchedNote = await noteServiceClient.getNote(getNoteRequest);
      
      console.log('Note fetched:', fetchedNote);
      
      if (!fetchedNote) {
        throw new Error('笔记数据为空');
      }
      
        setNote(fetchedNote);

        // Fetch attachments for this note
        if (fetchedNote.name) {
          try {
            const attachmentsRequest = create(ListAttachmentsRequestSchema, {
              noteId: fetchedNote.name,
              pageSize: 100,
            });
            const attachmentsResponse = await attachmentServiceClient.listAttachments(attachmentsRequest);
            setAttachments(attachmentsResponse.attachments || []);
          } catch (err) {
            console.error('Failed to load attachments:', err);
          // Don't fail the whole page if attachments fail to load
          }
        }
      } catch (err: any) {
        console.error('Failed to fetch note:', err);
      const errorMessage = err?.message || err?.toString() || '加载笔记失败';
      setError(errorMessage);
      setNote(null);
      } finally {
        setLoading(false);
      }
    };

  useEffect(() => {
    fetchNote();
  }, [id]);

  const formatDate = (timestamp: bigint | undefined): string => {
    if (!timestamp) return '未知时间';
    const date = new Date(Number(timestamp) * 1000);
    return date.toLocaleString('zh-CN');
  };

  const handleDownloadAttachment = async (attachment: Attachment) => {
    try {
      const attachmentResponse = await attachmentServiceClient.getAttachment({
        name: attachment.name || '',
      });

      if (attachmentResponse.content) {
        const content = attachmentResponse.content instanceof Uint8Array 
          ? attachmentResponse.content 
          : new Uint8Array(attachmentResponse.content);
        const blob = new Blob([content as BlobPart], { type: attachment.type || 'application/octet-stream' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = attachment.filename || 'download';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      }
    } catch (error: any) {
      console.error('Failed to download attachment:', error);
      alert(`下载失败: ${error.message || '未知错误'}`);
    }
  };

  const handlePreviewImage = async (attachment: Attachment) => {
    try {
      const attachmentResponse = await attachmentServiceClient.getAttachment({
        name: attachment.name || '',
      });

      if (attachmentResponse.content) {
        const content = attachmentResponse.content instanceof Uint8Array 
          ? attachmentResponse.content 
          : new Uint8Array(attachmentResponse.content);
        const blob = new Blob([content as BlobPart], { type: attachment.type || 'image/png' });
        const url = URL.createObjectURL(blob);
        setPreviewImage(url);
      }
    } catch (error: any) {
      console.error('Failed to preview image:', error);
      alert(`预览失败: ${error.message || '未知错误'}`);
    }
  };

  const handleDelete = async () => {
    if (!note || !note.name) return;
    if (!window.confirm('确定要删除这篇笔记吗？此操作无法撤销。')) return;

    try {
      const request = create(DeleteNoteRequestSchema, {
        name: note.name,
      });
      await noteServiceClient.deleteNote(request);
      navigate('/');
    } catch (err: any) {
      console.error('Failed to delete note:', err);
      alert(`删除失败: ${err.message || '未知错误'}`);
    }
  };

  const handleSave = () => {
    setIsEditing(false);
    fetchNote();
  };

  const isImage = (type: string | undefined): boolean => {
    if (!type) return false;
    return type.startsWith('image/');
  };

  // Check if current user is the author
  const isAuthor = currentUser && note && String(currentUser.id) === note.authorId;
  
  // Debug logging
  useEffect(() => {
    console.log('=== NoteDetail State ===');
    console.log('id:', id);
    console.log('loading:', loading);
    console.log('error:', error);
    console.log('hasNote:', !!note);
    console.log('isEditing:', isEditing);
    console.log('note:', note);
    console.log('currentUser:', currentUser);
    console.log('isAuthor:', isAuthor);
    console.log('=======================');
  }, [id, loading, error, note, isEditing, currentUser, isAuthor]);

  // Show loading state
  if (loading) {
    console.log('Rendering: Loading state');
    return <div className="note-detail-loading">加载中...</div>;
  }

  // If editing but note doesn't exist, show error with cancel option
  if (isEditing && !note) {
    console.log('Rendering: Editing mode but no note');
    return (
      <div className="note-detail">
        <div className="note-edit-container">
          <div className="edit-header">
            <h2>编辑笔记</h2>
            <button className="btn-close" onClick={() => setIsEditing(false)}>关闭</button>
          </div>
          <div className="note-detail-error">{(error || '笔记不存在')}</div>
        </div>
      </div>
    );
  }

  // If no note and not editing, show error
  if (!note && !isEditing) {
    console.log('Rendering: No note and not editing, showing error');
    return <div className="note-detail-error">{(error || '笔记不存在')}</div>;
  }

  // If we reach here, we should have a note (either editing or viewing)
  // If note is still null, show error as fallback
  if (!note) {
    console.log('Rendering: Fallback error (note is null)');
    return <div className="note-detail-error">{(error || '笔记不存在')}</div>;
  }

  console.log('Rendering: Main content with note:', note);

  return (
    <div className="note-detail">
      {isEditing ? (
        <div className="note-edit-container">
          <div className="edit-header">
            <h2>编辑笔记</h2>
            <button className="btn-close" onClick={() => setIsEditing(false)}>关闭</button>
          </div>
          {note?.name ? (
            <NoteEditor 
              noteId={note.name} 
              onSave={handleSave} 
              onCancel={() => setIsEditing(false)} 
            />
          ) : (
            <div className="note-detail-loading">加载笔记信息中...</div>
          )}
        </div>
      ) : (
        <div className="home-content">
          <div className="posts-list">
      <article className="note">
        <div className="note-header">
                  <div className="note-title-row">
                    <h1 className="note-title">{note?.title || '无标题'}</h1>
                    {isAuthor && (
                      <div className="note-actions">
                        <button 
                          className="btn-edit" 
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            console.log('Edit button clicked, setting isEditing to true');
                            setIsEditing(true);
                          }}
                        >
                          编辑
                        </button>
                        <button 
                          className="btn-delete" 
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            handleDelete();
                          }}
                        >
                          删除
                        </button>
                      </div>
                    )}
                  </div>
          <div className="note-meta">
                    <span className="note-date">发布时间：{formatDate(note?.publishedAt || note?.createdAt)}</span>
                    {note && note.readingTime > 0 && (
              <span className="reading-time">阅读时间：{note.readingTime}分钟</span>
            )}
                    {note && note.viewCount > 0 && (
              <span className="view-count">阅读量：{note.viewCount}</span>
            )}
          </div>
                  {note?.tagIds && note.tagIds.length > 0 && (
            <div className="note-tags">
              {note.tagIds.map((tagId, index) => (
                <span key={index} className="tag">#{tagId}</span>
              ))}
            </div>
          )}
        </div>

        {/* Markdown Content */}
        <div className="note-body">
                  {note?.content ? (
                    <MarkdownContent content={note.content} />
                  ) : (
                    <p style={{ color: '#999', fontStyle: 'italic' }}>暂无内容</p>
                  )}
              </div>

        {/* Attachments */}
        {attachments.length > 0 && (
          <div className="note-attachments">
            <h3>附件 ({attachments.length})</h3>
            <div className="attachments-list">
              {attachments.map((attachment) => (
                <div key={attachment.name} className="attachment-item">
                  <span className="attachment-name">{attachment.filename}</span>
                  <span className="attachment-size">
                    ({(Number(attachment.size) / 1024).toFixed(2)} KB)
                  </span>
                  <div className="attachment-actions">
                    {isImage(attachment.type) && (
                      <button
                        className="btn-preview"
                        onClick={() => handlePreviewImage(attachment)}
                      >
                        预览
                      </button>
                    )}
                    <button
                      className="btn-download"
                      onClick={() => handleDownloadAttachment(attachment)}
                    >
                      下载
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
            </article>
          </div>
          <Sidebar activeCategoryId={note?.categoryId} />
        </div>
      )}

      {/* Image Preview Modal */}
      {previewImage && (
        <div className="image-preview-modal" onClick={() => setPreviewImage(null)}>
          <div className="image-preview-content" onClick={(e) => e.stopPropagation()}>
            <button className="close-btn" onClick={() => setPreviewImage(null)}>×</button>
            <img src={previewImage} alt="Preview" />
          </div>
        </div>
      )}
    </div>
  );
};

export default NoteDetail;
