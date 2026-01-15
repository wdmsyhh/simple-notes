import React from "react";
import type { Attachment } from "../../../types/proto/api/v1/attachment_service_pb";
import type { LocalFile } from "../types";
import { getAttachmentThumbnailUrl, isImage } from "../../../utils/attachment";
import "../NoteEditor.css";

interface AttachmentListProps {
  attachments: Attachment[];
  localFiles: LocalFile[];
  onRemoveAttachment?: (name: string) => void;
  onRemoveLocalFile?: (previewUrl: string) => void;
}

const AttachmentList: React.FC<AttachmentListProps> = ({
  attachments,
  localFiles,
  onRemoveAttachment,
  onRemoveLocalFile,
}) => {
  if (attachments.length === 0 && localFiles.length === 0) {
    return null;
  }

  const formatFileSize = (bytes: bigint | number): string => {
    const size = typeof bytes === "bigint" ? Number(bytes) : bytes;
    if (size < 1024) return `${size} B`;
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  };

  const getFileTypeIcon = (filename: string, mimeType?: string) => {
    if (mimeType?.startsWith("image/")) {
      return "🖼️";
    }
    return "📄";
  };

  return (
    <div className="note-editor-attachment-list">
      <div className="note-editor-attachment-header">
        <span className="note-editor-attachment-header-icon">📎</span>
        <span className="note-editor-attachment-header-text">
          附件 ({attachments.length + localFiles.length})
        </span>
      </div>
      <div className="note-editor-attachment-items">
        {attachments.map((attachment) => {
          const attachmentIsImage = isImage(attachment.type);
          return (
            <div key={attachment.name} className="note-editor-attachment-item">
              <div className="note-editor-attachment-item-thumbnail">
                {attachmentIsImage && attachment.name ? (
                  <img
                    src={getAttachmentThumbnailUrl(attachment)}
                    alt={attachment.filename}
                    onError={(e) => {
                      (e.target as HTMLImageElement).style.display = "none";
                    }}
                  />
                ) : (
                  <span className="note-editor-attachment-item-thumbnail-icon">📄</span>
                )}
              </div>
              <div className="note-editor-attachment-item-info">
                <span className="note-editor-attachment-item-name" title={attachment.filename}>
                  {attachment.filename}
                </span>
                <div className="note-editor-attachment-item-meta">
                  <span>{attachment.type || "文件"}</span>
                  {attachment.size && (
                    <>
                      <span>•</span>
                      <span>{formatFileSize(attachment.size)}</span>
                    </>
                  )}
                </div>
              </div>
              {onRemoveAttachment && (
                <div className="note-editor-attachment-item-actions">
                  <button
                    type="button"
                    onClick={() => onRemoveAttachment(attachment.name)}
                    className="note-editor-attachment-item-button"
                    title="删除"
                  >
                    <span className="note-editor-attachment-item-button-icon">×</span>
                  </button>
                </div>
              )}
            </div>
          );
        })}
        {localFiles.map((localFile) => {
          const isImage = localFile.file.type.startsWith("image/");
          return (
            <div key={localFile.previewUrl} className="note-editor-attachment-item">
              <div className="note-editor-attachment-item-thumbnail">
                {isImage && localFile.previewUrl ? (
                  <img src={localFile.previewUrl} alt={localFile.file.name} />
                ) : (
                  <span className="note-editor-attachment-item-thumbnail-icon">📄</span>
                )}
              </div>
              <div className="note-editor-attachment-item-info">
                <span className="note-editor-attachment-item-name" title={localFile.file.name}>
                  {localFile.file.name}
                </span>
                <div className="note-editor-attachment-item-meta">
                  <span>{localFile.file.type || "文件"}</span>
                  <span>•</span>
                  <span>{formatFileSize(localFile.file.size)}</span>
                  <span className="note-editor-attachment-item-uploading">上传中...</span>
                </div>
              </div>
              {onRemoveLocalFile && (
                <div className="note-editor-attachment-item-actions">
                  <button
                    type="button"
                    onClick={() => onRemoveLocalFile(localFile.previewUrl)}
                    className="note-editor-attachment-item-button"
                    title="删除"
                  >
                    <span className="note-editor-attachment-item-button-icon">×</span>
                  </button>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default AttachmentList;

