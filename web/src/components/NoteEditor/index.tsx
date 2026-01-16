import React, { useRef, useEffect, useState } from "react";
import { EditorProvider, useEditorContext } from "./context";
import Editor, { type EditorRefActions } from "./Editor";
import { noteServiceClient, attachmentServiceClient, categoryServiceClient, tagServiceClient } from "../../connect";
import { create } from "@bufbuild/protobuf";
import { ConnectError } from "@connectrpc/connect";
import { CreateNoteRequestSchema, UpdateNoteRequestSchema, GetNoteRequestSchema } from "../../types/proto/api/v1/note_service_pb";
import { ListAttachmentsRequestSchema, AttachmentSchema } from "../../types/proto/api/v1/attachment_service_pb";
import { ListCategoriesRequestSchema, CreateCategoryRequestSchema, UpdateCategoryRequestSchema, DeleteCategoryRequestSchema } from "../../types/proto/api/v1/category_service_pb";
import { ListTagsRequestSchema, CreateTagRequestSchema, UpdateTagRequestSchema, DeleteTagRequestSchema } from "../../types/proto/api/v1/tag_service_pb";
import type { Attachment } from "../../types/proto/api/v1/attachment_service_pb";
import { NoteSchema, NoteVisibility, CategorySchema, TagSchema } from "../../types/proto/store/note_pb";
import type { Category } from "../../types/proto/store/note_pb";
import type { Tag } from "../../types/proto/store/note_pb";
import { useAuth } from "../../contexts/AuthContext";
import { uploadService } from "./services/uploadService";
import { useFileUpload } from "./hooks/useFileUpload";
import { useDragAndDrop } from "./hooks/useDragAndDrop";
import AttachmentList from "./components/AttachmentList";
import type { LocalFile } from "./types";
import type { Note } from "../../types/proto/store/note_pb";
import "./NoteEditor.css";

interface NoteEditorProps {
  noteId?: string; // Resource name format: notes/{id}, if provided, load and edit this note
  onSave?: () => void;
  onCancel?: () => void;
}

const NoteEditorImpl: React.FC<NoteEditorProps> = ({ noteId, onSave, onCancel }) => {
  const { state, dispatch } = useEditorContext();
  const editorRef = useRef<EditorRefActions>(null);
  const { currentUser } = useAuth();
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [showCreateCategory, setShowCreateCategory] = useState(false);
  const [showEditCategory, setShowEditCategory] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);
  const [showCreateTag, setShowCreateTag] = useState(false);
  const [showEditTag, setShowEditTag] = useState(false);
  const [editingTag, setEditingTag] = useState<Tag | null>(null);
  const [showTagDropdown, setShowTagDropdown] = useState(false);
  const [showCategoryDropdown, setShowCategoryDropdown] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [newTagName, setNewTagName] = useState('');
  const tagDropdownRef = useRef<HTMLDivElement>(null);
  const categoryDropdownRef = useRef<HTMLDivElement>(null);

  // Load categories
  useEffect(() => {
    const loadCategories = async () => {
      try {
        const request = create(ListCategoriesRequestSchema, {
          includeHidden: false,
          parentId: BigInt(0),
        });
        const response = await categoryServiceClient.listCategories(request);
        setCategories(response.categories || []);
      } catch (error) {
        console.error("Failed to load categories:", error);
      }
    };
    loadCategories();
  }, []);

  // Load tags
  useEffect(() => {
    const loadTags = async () => {
      try {
        const request = create(ListTagsRequestSchema, {
          limit: 100,
          offset: 0,
        });
        const response = await tagServiceClient.listTags(request);
        setTags(response.tags || []);
      } catch (error) {
        console.error("Failed to load tags:", error);
      }
    };
    loadTags();
  }, []);

  // Close tag dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (tagDropdownRef.current && !tagDropdownRef.current.contains(event.target as Node)) {
        setShowTagDropdown(false);
      }
      if (categoryDropdownRef.current && !categoryDropdownRef.current.contains(event.target as Node)) {
        setShowCategoryDropdown(false);
      }
    };

    if (showTagDropdown || showCategoryDropdown) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showTagDropdown, showCategoryDropdown]);

  // Load note if noteId is provided
  useEffect(() => {
    if (!noteId) {
      // Reset if noteId is removed
      if (state.noteId) {
        dispatch({ type: "RESET" });
      }
      return;
    }

    // Only load if noteId changed and doesn't match current state
    if (noteId !== state.noteId) {
      const loadNote = async () => {
        dispatch({ type: "SET_LOADING", payload: { key: "loading", value: true } });
        try {
          const request = create(GetNoteRequestSchema, {
            name: noteId,
          });
          const note = await noteServiceClient.getNote(request);
          
          // Extract ID from resource name
          const noteIdFromName = note.name || noteId;
          
      // Load attachments for this note
      let attachments: Attachment[] = [];
      try {
        const attachmentsRequest = create(ListAttachmentsRequestSchema, {
          noteId: noteIdFromName,
          pageSize: 100,
        });
        const attachmentsResponse = await attachmentServiceClient.listAttachments(attachmentsRequest);
        attachments = attachmentsResponse.attachments || [];
      } catch (error) {
        console.error("Failed to load attachments:", error);
        // Continue even if loading attachments fails
      }

      dispatch({
        type: "LOAD_NOTE",
        payload: {
          noteId: noteIdFromName,
          title: note.title || "",
          summary: note.summary || "",
          content: note.content || "",
          categoryId: note.categoryId || "",
          tagIds: note.tagIds || [],
          visibility: note.visibility || NoteVisibility.PUBLIC,
          attachments: attachments,
        },
      });
        } catch (error: any) {
          console.error("Failed to load note:", error);
          alert(`加载笔记失败: ${error.message || "未知错误"}`);
        } finally {
          dispatch({ type: "SET_LOADING", payload: { key: "loading", value: false } });
        }
      };

      loadNote();
    }
  }, [noteId, state.noteId, dispatch]);

  // File upload handlers
  const handleFilesSelected = (localFiles: LocalFile[]) => {
    localFiles.forEach((localFile) => {
      dispatch({ type: "ADD_LOCAL_FILE", payload: localFile });
    });
  };

  const { fileInputRef, handleFileInputChange, handleUploadClick } = useFileUpload(handleFilesSelected);

  const { dragHandlers } = useDragAndDrop((files: FileList) => {
    const localFiles: LocalFile[] = Array.from(files).map((file) => ({
      file,
      previewUrl: URL.createObjectURL(file),
    }));
    handleFilesSelected(localFiles);
  });

  const handleRemoveLocalFile = (previewUrl: string) => {
    dispatch({ type: "REMOVE_LOCAL_FILE", payload: previewUrl });
    // Revoke object URL to free memory
    URL.revokeObjectURL(previewUrl);
  };

  const handleRemoveAttachment = async (name: string) => {
    if (!confirm('确定要删除这个附件吗？')) {
      return;
    }

    try {
      await uploadService.deleteAttachment(name);
      dispatch({ type: "REMOVE_ATTACHMENT", payload: name });
    } catch (error: any) {
      console.error('Failed to delete attachment:', error);
      alert(`删除附件失败: ${error.message || '未知错误'}`);
    }
  };

  const handleSave = async () => {
    // 验证必填字段
    const title = state.title.trim();
    const summary = state.summary.trim();
    const content = state.content.trim();

    if (!title) {
      alert("请输入标题");
      return;
    }

    if (!summary) {
      alert("请输入描述");
      return;
    }

    if (!content) {
      alert("请输入内容");
      return;
    }

    dispatch({ type: "SET_LOADING", payload: { key: "saving", value: true } });

    try {
      // 1. Upload local files first and collect all attachments
      let allAttachments = [...state.attachments]; // Start with existing attachments
      
      if (state.localFiles.length > 0) {
        dispatch({ type: "SET_LOADING", payload: { key: "uploading", value: true } });
        try {
          const newAttachments = await uploadService.uploadFiles(state.localFiles);
          console.log(`Uploaded ${newAttachments.length} new attachments:`, newAttachments.map(att => ({ name: att.name, filename: att.filename })));
          
          // Add new attachments to the list
          allAttachments = [...allAttachments, ...newAttachments];
          
          // Update state
          newAttachments.forEach((attachment) => {
            dispatch({ type: "ADD_ATTACHMENT", payload: attachment });
          });
          dispatch({ type: "CLEAR_LOCAL_FILES" });
        } catch (error: any) {
          console.error("Failed to upload files:", error);
          alert(`文件上传失败: ${error.message || "未知错误"}`);
          dispatch({ type: "SET_LOADING", payload: { key: "uploading", value: false } });
          dispatch({ type: "SET_LOADING", payload: { key: "saving", value: false } });
          return;
        } finally {
          dispatch({ type: "SET_LOADING", payload: { key: "uploading", value: false } });
        }
      }
      
      console.log(`Total attachments to link: ${allAttachments.length}`, allAttachments.map(att => ({ name: att.name, filename: att.filename })));

      if (state.noteId) {
        // Update existing note
        // Extract ID from resource name
        const idMatch = state.noteId.match(/notes\/(\d+)/);
        const noteIdNum = idMatch ? BigInt(idMatch[1]) : BigInt(0);

        const note = create(NoteSchema, {
          name: state.noteId,
          id: noteIdNum,
          title: title,
          slug: "", // Slug will be preserved by backend
          content: content,
          summary: summary,
          categoryId: state.categoryId,
          tagIds: state.tagIds,
          published: true,
          authorId: "",
          createdAt: BigInt(0),
          updatedAt: BigInt(Math.floor(Date.now() / 1000)),
          publishedAt: BigInt(0),
          coverImage: "",
          readingTime: 0,
          viewCount: 0,
          visibility: state.visibility,
        });

        const request = create(UpdateNoteRequestSchema, {
          note,
        });

        const updatedNote = await noteServiceClient.updateNote(request);
        
        console.log(`Updating note ${updatedNote.name}, attachments count: ${allAttachments.length}`);
        console.log(`Attachments:`, allAttachments.map(att => ({ name: att.name, filename: att.filename })));
        
        // Sync attachments: unlink removed attachments and link new ones
        if (updatedNote.name) {
          // Get current attachments linked to this note
          let currentAttachments: Attachment[] = [];
          try {
            const currentAttachmentsRequest = create(ListAttachmentsRequestSchema, {
              noteId: updatedNote.name,
              pageSize: 100,
            });
            const currentAttachmentsResponse = await attachmentServiceClient.listAttachments(currentAttachmentsRequest);
            currentAttachments = currentAttachmentsResponse.attachments || [];
            console.log(`Current attachments linked to note: ${currentAttachments.length}`, currentAttachments.map(att => ({ name: att.name, filename: att.filename })));
          } catch (error) {
            console.error("Failed to load current attachments:", error);
          }

          // Find attachments to unlink (in current but not in new list)
          const currentAttachmentNames = new Set(currentAttachments.map(att => att.name).filter((name): name is string => !!name));
          const newAttachmentNames = new Set(allAttachments.map(att => att.name).filter((name): name is string => !!name));
          
          // Unlink removed attachments
          for (const currentAttachmentName of currentAttachmentNames) {
            if (!newAttachmentNames.has(currentAttachmentName)) {
              try {
                await attachmentServiceClient.updateAttachment({
                  attachment: create(AttachmentSchema, {
                    name: currentAttachmentName,
                    noteId: "", // Empty noteId to unlink
                  }),
                });
                console.log(`Unlinked attachment ${currentAttachmentName} from note ${updatedNote.name}`);
              } catch (error) {
                console.error(`Failed to unlink attachment ${currentAttachmentName}:`, error);
              }
            }
          }

          // Link new attachments (not already linked)
          const attachmentsToLink = allAttachments
            .map(att => att.name)
            .filter((name): name is string => !!name && !currentAttachmentNames.has(name));
          
          console.log(`Attachments to link: ${attachmentsToLink.length}`, attachmentsToLink);
          if (attachmentsToLink.length > 0) {
            await uploadService.linkAttachmentsToNote(attachmentsToLink, updatedNote.name);
          }
        }
      } else {
        // Create new note
        // Slug 是可选的，留空让后端使用 ID
        const note = create(NoteSchema, {
          name: "",
          id: BigInt(0),
          title: title,
          slug: "", // 不使用 slug，直接使用 ID
          content: content,
          summary: summary,
          categoryId: state.categoryId,
          tagIds: state.tagIds,
          published: true,
          authorId: "",
          createdAt: BigInt(0),
          updatedAt: BigInt(0),
          publishedAt: BigInt(0),
          coverImage: "",
          readingTime: 0,
          viewCount: 0,
          visibility: state.visibility,
        });

        const request = create(CreateNoteRequestSchema, {
          note,
        });

        const createdNote = await noteServiceClient.createNote(request);
        
        // Link attachments to note
        console.log(`Creating note ${createdNote.name}, attachments count: ${allAttachments.length}`);
        console.log(`Attachments:`, allAttachments.map(att => ({ name: att.name, filename: att.filename })));
        
        if (allAttachments.length > 0 && createdNote.name) {
          const attachmentNames = allAttachments.map((att) => att.name).filter((name): name is string => !!name);
          console.log(`Linking ${attachmentNames.length} attachments to note ${createdNote.name}:`, attachmentNames);
          await uploadService.linkAttachmentsToNote(attachmentNames, createdNote.name);
        } else {
          console.warn(`Skipping attachment linking: attachments.length=${allAttachments.length}, note.name=${createdNote.name}`);
        }
      }

      // Reset editor
      dispatch({ type: "RESET" });
      onSave?.();
    } catch (error: any) {
      console.error("Failed to save note:", error);
      alert(error.message || "保存失败");
    } finally {
      dispatch({ type: "SET_LOADING", payload: { key: "saving", value: false } });
    }
  };

  const handleCategoryChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    dispatch({ type: "SET_CATEGORY", payload: e.target.value });
  };

  const handleCreateCategory = async () => {
    if (!newCategoryName.trim()) {
      alert('请输入分类名称');
      return;
    }

    try {
      // 只传分类名称，slug 由后端自动生成
      const category = create(CategorySchema, {
        name: '',
        id: BigInt(0),
        nameText: newCategoryName.trim(),
        slug: '', // 留空，让后端自动生成
        description: '',
        parentId: BigInt(0),
        order: 0,
        visible: true,
        createdAt: BigInt(0),
        updatedAt: BigInt(0),
      });

      const request = create(CreateCategoryRequestSchema, {
        category,
      });

      const createdCategory = await categoryServiceClient.createCategory(request);
      
      // 刷新分类列表
      const categoriesRequest = create(ListCategoriesRequestSchema, {
        includeHidden: false,
        parentId: BigInt(0),
      });
      const categoriesResponse = await categoryServiceClient.listCategories(categoriesRequest);
      setCategories(categoriesResponse.categories || []);
      
      // 自动选择新创建的分类
      dispatch({ type: "SET_CATEGORY", payload: String(createdCategory.id) });
      
      // 关闭弹窗并清空输入
      setShowCreateCategory(false);
      setNewCategoryName('');
    } catch (error: any) {
      console.error('Failed to create category:', error);
      alert(`创建分类失败: ${error.message || '未知错误'}`);
    }
  };

  const handleEditCategory = (category: Category) => {
    setEditingCategory(category);
    setNewCategoryName(category.nameText || '');
    setShowEditCategory(true);
  };

  const handleUpdateCategory = async () => {
    if (!editingCategory || !newCategoryName.trim()) {
      alert('请输入分类名称');
      return;
    }

    try {
      const category = create(CategorySchema, {
        name: editingCategory.name || `categories/${editingCategory.id}`,
        id: editingCategory.id,
        nameText: newCategoryName.trim(),
        slug: editingCategory.slug || '',
        description: editingCategory.description || '',
        parentId: editingCategory.parentId || BigInt(0),
        order: editingCategory.order || 0,
        visible: editingCategory.visible !== undefined ? editingCategory.visible : true,
        createdAt: editingCategory.createdAt || BigInt(0),
        updatedAt: editingCategory.updatedAt || BigInt(0),
      });

      const request = create(UpdateCategoryRequestSchema, {
        category,
      });

      await categoryServiceClient.updateCategory(request);
      
      // 刷新分类列表
      const categoriesRequest = create(ListCategoriesRequestSchema, {
        includeHidden: false,
        parentId: BigInt(0),
      });
      const categoriesResponse = await categoryServiceClient.listCategories(categoriesRequest);
      setCategories(categoriesResponse.categories || []);
      
      // 关闭弹窗并清空输入
      setShowEditCategory(false);
      setEditingCategory(null);
      setNewCategoryName('');
    } catch (error: any) {
      console.error('Failed to update category:', error);
      alert(`更新分类失败: ${error.message || '未知错误'}`);
    }
  };

  const handleDeleteCategory = async (category: Category) => {
    if (!category.name && !category.id) {
      alert('分类信息不完整，无法删除');
      return;
    }

    if (!window.confirm(`确定要删除分类"${category.nameText || '未命名分类'}"吗？此操作无法撤销。`)) {
      return;
    }

    try {
      const categoryName = category.name || `categories/${category.id}`;
      const request = create(DeleteCategoryRequestSchema, {
        name: categoryName,
      });

      await categoryServiceClient.deleteCategory(request);
      
      // 刷新分类列表
      const categoriesRequest = create(ListCategoriesRequestSchema, {
        includeHidden: false,
        parentId: BigInt(0),
      });
      const categoriesResponse = await categoryServiceClient.listCategories(categoriesRequest);
      setCategories(categoriesResponse.categories || []);
      
      // 如果删除的是当前选中的分类，清空选择
      if (state.categoryId === String(category.id)) {
        dispatch({ type: "SET_CATEGORY", payload: "" });
      }
    } catch (error: any) {
      console.error('Failed to delete category:', error);
      let errorMessage = '删除分类失败';
      if (error instanceof ConnectError) {
        // 提取 ConnectError 的错误信息
        const message = error.message || "";
        // 匹配 "desc = 该分类下还有文章，无法删除" 格式
        const descMatch = message.match(/desc\s*=\s*(.+)$/);
        if (descMatch && descMatch[1]) {
          errorMessage = descMatch[1].trim();
        } else {
          // 如果没有匹配到，尝试直接使用 message，但清理格式
          errorMessage = message.replace(/^\[unknown\]\s*rpc error:\s*code\s*=\s*\w+\s*desc\s*=\s*/i, "").trim() || errorMessage;
        }
      } else if (error?.message) {
        errorMessage = error.message;
      }
      alert(errorMessage);
    }
  };

  const handleTagToggle = (tagId: string) => {
    const currentTagIds = state.tagIds || [];
    const isSelected = currentTagIds.includes(tagId);
    const newTagIds = isSelected
      ? currentTagIds.filter(id => id !== tagId)
      : [...currentTagIds, tagId];
    dispatch({ type: "SET_TAGS", payload: newTagIds });
  };

  const handleCreateTag = async () => {
    if (!newTagName.trim()) {
      alert('请输入标签名称');
      return;
    }

    try {
      // 只传标签名称，slug 由后端自动生成
      const tag = create(TagSchema, {
        name: '',
        id: BigInt(0),
        nameText: newTagName.trim(),
        slug: '', // 留空，让后端处理
        description: '',
        count: 0,
        createdAt: BigInt(0),
        updatedAt: BigInt(0),
      });

      const request = create(CreateTagRequestSchema, {
        tag,
      });

      const createdTag = await tagServiceClient.createTag(request);
      
      // 刷新标签列表
      const tagsRequest = create(ListTagsRequestSchema, {
        limit: 100,
        offset: 0,
      });
      const tagsResponse = await tagServiceClient.listTags(tagsRequest);
      setTags(tagsResponse.tags || []);
      
      // 自动选择新创建的标签
      const currentTagIds = state.tagIds || [];
      if (!currentTagIds.includes(String(createdTag.id))) {
        dispatch({ type: "SET_TAGS", payload: [...currentTagIds, String(createdTag.id)] });
      }
      
      // 关闭弹窗并清空输入
      setShowCreateTag(false);
      setNewTagName('');
    } catch (error: any) {
      console.error('Failed to create tag:', error);
      alert(`创建标签失败: ${error.message || '未知错误'}`);
    }
  };

  const handleEditTag = (tag: Tag) => {
    setEditingTag(tag);
    setNewTagName(tag.nameText || '');
    setShowEditTag(true);
  };

  const handleUpdateTag = async () => {
    if (!editingTag || !newTagName.trim()) {
      alert('请输入标签名称');
      return;
    }
    try {
      const tag = create(TagSchema, {
        name: editingTag.name || `tags/${editingTag.id}`,
        id: editingTag.id,
        nameText: newTagName.trim(),
        slug: editingTag.slug || '',
        description: editingTag.description || '',
        count: editingTag.count || 0,
        createdAt: editingTag.createdAt || BigInt(0),
        updatedAt: editingTag.updatedAt || BigInt(0),
      });
      const request = create(UpdateTagRequestSchema, { tag });
      await tagServiceClient.updateTag(request);
      // 刷新标签列表
      const tagsRequest = create(ListTagsRequestSchema, {
        limit: 100,
        offset: 0,
      });
      const tagsResponse = await tagServiceClient.listTags(tagsRequest);
      setTags(tagsResponse.tags || []);
      setShowEditTag(false);
      setEditingTag(null);
      setNewTagName('');
    } catch (error: any) {
      console.error('Failed to update tag:', error);
      alert(`更新标签失败: ${error.message || '未知错误'}`);
    }
  };

  const handleDeleteTag = async (tag: Tag) => {
    if (!tag.name && !tag.id) {
      alert('标签信息不完整，无法删除');
      return;
    }
    if (!window.confirm(`确定要删除标签"${tag.nameText || '未命名标签'}"吗？此操作无法撤销。`)) {
      return;
    }
    try {
      const tagName = tag.name || `tags/${tag.id}`;
      const request = create(DeleteTagRequestSchema, {
        name: tagName,
      });
      await tagServiceClient.deleteTag(request);
      // 刷新标签列表
      const tagsRequest = create(ListTagsRequestSchema, {
        limit: 100,
        offset: 0,
      });
      const tagsResponse = await tagServiceClient.listTags(tagsRequest);
      setTags(tagsResponse.tags || []);
      // 如果删除的是当前选中的标签，从选中列表中移除
      const currentTagIds = state.tagIds || [];
      if (currentTagIds.includes(String(tag.id))) {
        dispatch({ type: "SET_TAGS", payload: currentTagIds.filter(id => id !== String(tag.id)) });
      }
    } catch (error: any) {
      console.error('Failed to delete tag:', error);
      let errorMessage = '删除标签失败';
      if (error instanceof ConnectError) {
        const message = error.message || "";
        const descMatch = message.match(/desc\s*=\s*(.+)$/);
        if (descMatch && descMatch[1]) {
          errorMessage = descMatch[1].trim();
        } else {
          errorMessage = message.replace(/^\[unknown\]\s*rpc error:\s*code\s*=\s*\w+\s*desc\s*=\s*/i, "").trim() || errorMessage;
        }
      } else if (error?.message) {
        errorMessage = error.message;
      }
      alert(errorMessage);
    }
  };

  const handleVisibilityChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const visibility =
      e.target.value === "PUBLIC"
        ? NoteVisibility.PUBLIC
        : NoteVisibility.PRIVATE;
    dispatch({ type: "SET_VISIBILITY", payload: visibility });
  };

  const canSave = state.title.trim().length > 0 && state.summary.trim().length > 0 && state.content.trim().length > 0;
  const isLoading = state.isLoading.loading || state.isLoading.saving || state.isLoading.uploading;

  if (state.isLoading.loading) {
    return (
      <div className="note-editor loading">
        <div style={{ textAlign: "center", padding: "32px 0", color: "#6b7280" }}>加载中...</div>
      </div>
    );
  }

  return (
    <div className="note-editor" {...dragHandlers}>
      {/* Hidden file input */}
      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={handleFileInputChange}
      />

      {/* Title Input */}
      <input
        type="text"
        className="note-editor-title-input"
        placeholder="标题（必填）..."
        value={state.title}
        onChange={(e) => dispatch({ type: "UPDATE_TITLE", payload: e.target.value })}
        required
      />

      {/* Summary Input */}
      <textarea
        className="note-editor-summary-input"
        placeholder="描述（必填，用于列表展示）..."
        value={state.summary}
        onChange={(e) => dispatch({ type: "UPDATE_SUMMARY", payload: e.target.value })}
        rows={2}
        required
      />

      {/* Editor */}
      <Editor ref={editorRef} placeholder="内容（必填）..." autoFocus />

      {/* Attachment List */}
      <AttachmentList
        attachments={state.attachments}
        localFiles={state.localFiles}
        onRemoveAttachment={handleRemoveAttachment}
        onRemoveLocalFile={handleRemoveLocalFile}
      />

      {/* Toolbar */}
      <div className="note-editor-toolbar">
        <div className="note-editor-toolbar-left">
          <button
            type="button"
            onClick={handleUploadClick}
            disabled={isLoading}
            className="note-editor-button"
            title="上传文件"
          >
            {state.isLoading.uploading ? "上传中..." : "📎 附件"}
          </button>
          <button
            type="button"
            className="note-editor-button"
            onClick={() => setShowCreateCategory(true)}
            title="新建分类"
          >
            + 新建分类
          </button>
          <div className="note-editor-category-select-wrapper" ref={categoryDropdownRef}>
            <select
              className="note-editor-select note-editor-select-category"
              value={state.categoryId}
              onChange={(e) => {
                handleCategoryChange(e);
                if (e.target.value) {
                  setShowCategoryDropdown(true);
                } else {
                  setShowCategoryDropdown(false);
                }
              }}
            >
              <option value="">选择分类</option>
              {categories.map((category) => (
                <option key={category.id} value={String(category.id)}>
                  {category.nameText || '未命名分类'}
                </option>
              ))}
            </select>
            {state.categoryId && showCategoryDropdown && categories.length > 0 && (
              <div className="note-editor-category-dropdown">
                <div className="note-editor-category-actions">
                  <button
                    type="button"
                    className="note-editor-tag-action-button"
                    onClick={(e) => {
                      e.stopPropagation();
                      const category = categories.find(c => String(c.id) === state.categoryId);
                      if (category) {
                        handleEditCategory(category);
                        setShowCategoryDropdown(false);
                      }
                    }}
                    title="编辑分类"
                  >
                    编辑
                  </button>
                  <button
                    type="button"
                    className="note-editor-tag-action-button"
                    onClick={(e) => {
                      e.stopPropagation();
                      const category = categories.find(c => String(c.id) === state.categoryId);
                      if (category) {
                        handleDeleteCategory(category);
                        setShowCategoryDropdown(false);
                      }
                    }}
                    title="删除分类"
                    style={{ color: '#d32f2f' }}
                  >
                    删除
                  </button>
                </div>
              </div>
            )}
          </div>
          <button
            type="button"
            className="note-editor-button"
            onClick={() => setShowCreateTag(true)}
            title="新建标签"
          >
            + 新建标签
          </button>
          <div className="note-editor-tag-select-wrapper" ref={tagDropdownRef}>
            <button
              type="button"
              className="note-editor-select note-editor-tag-select-button"
              onClick={() => setShowTagDropdown(!showTagDropdown)}
            >
              {state.tagIds.length > 0 ? `标签 (${state.tagIds.length})` : "选择标签"}
            </button>
            {showTagDropdown && (
              <div className="note-editor-tag-dropdown">
                {tags.length === 0 ? (
                  <div className="note-editor-tag-dropdown-empty">暂无标签</div>
                ) : (
                  tags.map((tag) => {
                    const isSelected = state.tagIds.includes(String(tag.id));
                    return (
                      <div key={tag.id} className="note-editor-tag-item-wrapper">
                        <label className="note-editor-tag-item">
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => handleTagToggle(String(tag.id))}
                          />
                          <span>{tag.nameText || '未命名标签'}</span>
                        </label>
                        <div className="note-editor-tag-actions">
                          <button
                            type="button"
                            className="note-editor-tag-action-button"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleEditTag(tag);
                            }}
                            title="编辑标签"
                          >
                            编辑
                          </button>
                          <button
                            type="button"
                            className="note-editor-tag-action-button"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteTag(tag);
                            }}
                            title="删除标签"
                            style={{ color: '#d32f2f' }}
                          >
                            删除
                          </button>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            )}
          </div>
          <select
            className="note-editor-select"
            value={state.visibility === NoteVisibility.PUBLIC ? "PUBLIC" : "PRIVATE"}
            onChange={handleVisibilityChange}
          >
            <option value="PUBLIC">公开</option>
            <option value="PRIVATE">私有</option>
          </select>
        </div>

        <div className="note-editor-toolbar-right">
          {onCancel && (
            <button
              className="note-editor-button-ghost"
              onClick={onCancel}
              disabled={isLoading}
            >
              取消
            </button>
          )}
          <button
            className="note-editor-button note-editor-button-primary"
            onClick={handleSave}
            disabled={!canSave || isLoading}
          >
            {state.isLoading.saving ? "保存中..." : state.noteId ? "更新" : "发布"}
          </button>
        </div>
      </div>

      {/* Create Category Modal */}
      {showCreateCategory && (
        <div className="create-category-modal-overlay" onClick={() => setShowCreateCategory(false)}>
          <div className="create-category-modal" onClick={(e) => e.stopPropagation()}>
            <div className="create-category-modal-header">
              <h3>新建分类</h3>
              <button
                className="create-category-modal-close"
                onClick={() => {
                  setShowCreateCategory(false);
                  setNewCategoryName('');
                }}
              >
                ×
              </button>
            </div>
            <div className="create-category-modal-body">
              <label>分类名称</label>
              <input
                type="text"
                value={newCategoryName}
                onChange={(e) => setNewCategoryName(e.target.value)}
                placeholder="请输入分类名称"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    handleCreateCategory();
                  } else if (e.key === 'Escape') {
                    setShowCreateCategory(false);
                    setNewCategoryName('');
                  }
                }}
              />
            </div>
            <div className="create-category-modal-footer">
              <button
                className="create-category-btn-primary"
                onClick={handleCreateCategory}
              >
                新建
              </button>
              <button
                className="create-category-btn-cancel"
                onClick={() => {
                  setShowCreateCategory(false);
                  setNewCategoryName('');
                }}
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Category Modal */}
      {showEditCategory && editingCategory && (
        <div className="create-category-modal-overlay" onClick={() => {
          setShowEditCategory(false);
          setEditingCategory(null);
          setNewCategoryName('');
        }}>
          <div className="create-category-modal" onClick={(e) => e.stopPropagation()}>
            <div className="create-category-modal-header">
              <h3>编辑分类</h3>
              <button
                className="create-category-modal-close"
                onClick={() => {
                  setShowEditCategory(false);
                  setEditingCategory(null);
                  setNewCategoryName('');
                }}
              >
                ×
              </button>
            </div>
            <div className="create-category-modal-body">
              <label>分类名称</label>
              <input
                type="text"
                value={newCategoryName}
                onChange={(e) => setNewCategoryName(e.target.value)}
                placeholder="请输入分类名称"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    handleUpdateCategory();
                  } else if (e.key === 'Escape') {
                    setShowEditCategory(false);
                    setEditingCategory(null);
                    setNewCategoryName('');
                  }
                }}
              />
            </div>
            <div className="create-category-modal-footer">
              <button
                className="create-category-btn-primary"
                onClick={handleUpdateCategory}
              >
                更新
              </button>
              <button
                className="create-category-btn-cancel"
                onClick={() => {
                  setShowEditCategory(false);
                  setEditingCategory(null);
                  setNewCategoryName('');
                }}
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Tag Modal */}
      {showEditTag && editingTag && (
        <div className="create-category-modal-overlay" onClick={() => {
          setShowEditTag(false);
          setEditingTag(null);
          setNewTagName('');
        }}>
          <div className="create-category-modal" onClick={(e) => e.stopPropagation()}>
            <div className="create-category-modal-header">
              <h3>编辑标签</h3>
              <button
                className="create-category-modal-close"
                onClick={() => {
                  setShowEditTag(false);
                  setEditingTag(null);
                  setNewTagName('');
                }}
              >
                ×
              </button>
            </div>
            <div className="create-category-modal-body">
              <label>标签名称</label>
              <input
                type="text"
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                placeholder="请输入标签名称"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    handleUpdateTag();
                  } else if (e.key === 'Escape') {
                    setShowEditTag(false);
                    setEditingTag(null);
                    setNewTagName('');
                  }
                }}
              />
            </div>
            <div className="create-category-modal-footer">
              <button
                className="create-category-btn-primary"
                onClick={handleUpdateTag}
              >
                更新
              </button>
              <button
                className="create-category-btn-cancel"
                onClick={() => {
                  setShowEditTag(false);
                  setEditingTag(null);
                  setNewTagName('');
                }}
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Tag Modal */}
      {showCreateTag && (
        <div className="create-category-modal-overlay" onClick={() => setShowCreateTag(false)}>
          <div className="create-category-modal" onClick={(e) => e.stopPropagation()}>
            <div className="create-category-modal-header">
              <h3>新建标签</h3>
              <button
                className="create-category-modal-close"
                onClick={() => {
                  setShowCreateTag(false);
                  setNewTagName('');
                }}
              >
                ×
              </button>
            </div>
            <div className="create-category-modal-body">
              <label>标签名称</label>
              <input
                type="text"
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                placeholder="请输入标签名称"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    handleCreateTag();
                  } else if (e.key === 'Escape') {
                    setShowCreateTag(false);
                    setNewTagName('');
                  }
                }}
              />
            </div>
            <div className="create-category-modal-footer">
              <button
                className="create-category-btn-primary"
                onClick={handleCreateTag}
              >
                新建
              </button>
              <button
                className="create-category-btn-cancel"
                onClick={() => {
                  setShowCreateTag(false);
                  setNewTagName('');
                }}
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const NoteEditor: React.FC<NoteEditorProps> = ({ noteId, ...props }) => {
  return (
    <EditorProvider>
      <NoteEditorImpl noteId={noteId} {...props} />
    </EditorProvider>
  );
};

export default NoteEditor;

