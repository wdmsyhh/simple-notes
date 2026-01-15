/**
 * 附件工具函数
 * 提供附件相关的辅助功能
 */
import type { Attachment } from "../types/proto/api/v1/attachment_service_pb";

/**
 * 从资源名称中提取附件ID
 * 格式：attachments/{id}
 * @param name - 资源名称
 * @returns 附件ID，如果格式不正确则返回 null
 */
const extractAttachmentId = (name: string): string | null => {
  if (!name) return null;
  const match = name.match(/^attachments\/(\d+)$/);
  return match ? match[1] : null;
};

/**
 * 获取附件的下载/查看URL
 * 使用新的 HTTP 文件服务器端点以更好地支持 Range 请求
 * @param attachment - 附件对象
 * @returns 附件的URL
 */
export const getAttachmentUrl = (attachment: Attachment): string => {
  if (!attachment.name || !attachment.filename) {
    return "";
  }
  const id = extractAttachmentId(attachment.name);
  if (!id) {
    return "";
  }
  // 使用 HTTP 文件服务器端点：/file/attachments/:id/:filename
  return `/file/attachments/${id}/${encodeURIComponent(attachment.filename)}`;
};

/**
 * 获取图片附件的缩略图URL
 * 目前使用与完整图片相同的URL
 * TODO: 未来添加缩略图支持
 * @param attachment - 附件对象
 * @returns 缩略图URL
 */
export const getAttachmentThumbnailUrl = (attachment: Attachment): string => {
  return getAttachmentUrl(attachment);
};

/**
 * 检查附件是否是图片
 * @param mimeType - MIME 类型
 * @returns 如果是图片则返回 true
 */
export const isImage = (mimeType?: string): boolean => {
  if (!mimeType) return false;
  return mimeType.startsWith("image/");
};

