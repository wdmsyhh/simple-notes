import { create } from "@bufbuild/protobuf";
import { attachmentServiceClient } from "../../../connect";
import { CreateAttachmentRequestSchema, UpdateAttachmentRequestSchema } from "../../../types/proto/api/v1/attachment_service_pb";
import { AttachmentSchema } from "../../../types/proto/api/v1/attachment_service_pb";
import type { Attachment } from "../../../types/proto/api/v1/attachment_service_pb";
import type { LocalFile } from "../types";

export const uploadService = {
  async uploadFiles(localFiles: LocalFile[]): Promise<Attachment[]> {
    if (localFiles.length === 0) return [];

    const attachments: Attachment[] = [];

    for (const { file } of localFiles) {
      const buffer = new Uint8Array(await file.arrayBuffer());
      const attachment = await attachmentServiceClient.createAttachment({
        attachment: create(AttachmentSchema, {
          filename: file.name,
          size: BigInt(file.size),
          type: file.type || "application/octet-stream",
          content: buffer,
        }),
      });
      attachments.push(attachment);
    }

    return attachments;
  },

  async linkAttachmentsToNote(attachmentNames: string[], noteId: string): Promise<void> {
    if (attachmentNames.length === 0) {
      return;
    }
    
    console.log(`Linking ${attachmentNames.length} attachments to note ${noteId}`);
    
    const errors: string[] = [];
    for (const attachmentName of attachmentNames) {
      try {
        console.log(`Linking attachment ${attachmentName} to note ${noteId}`);
        await attachmentServiceClient.updateAttachment({
          attachment: create(AttachmentSchema, {
            name: attachmentName,
            noteId: noteId,
          }),
        });
        console.log(`Successfully linked attachment ${attachmentName} to note ${noteId}`);
      } catch (error: any) {
        const errorMsg = `Failed to link attachment ${attachmentName} to note ${noteId}: ${error?.message || error}`;
        console.error(errorMsg, error);
        errors.push(errorMsg);
        // Continue with other attachments even if one fails
      }
    }
    
    if (errors.length > 0) {
      console.warn(`Some attachments failed to link:`, errors);
      // Don't throw, but log the errors
    }
  },

  async deleteAttachment(attachmentName: string): Promise<void> {
    await attachmentServiceClient.deleteAttachment({
      name: attachmentName,
    });
  },
};

