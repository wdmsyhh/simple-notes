import type { Attachment } from "../../types/proto/api/v1/attachment_service_pb";
import { NoteVisibility } from "../../types/proto/store/note_pb";

export interface LocalFile {
  file: File;
  previewUrl: string;
}

export interface EditorState {
  noteId: string | null; // Resource name format: notes/{id}
  title: string;
  summary: string;
  content: string;
  categoryId: string;
  tagIds: string[]; // Array of tag IDs
  visibility: NoteVisibility;
  attachments: Attachment[];
  localFiles: LocalFile[];
  isLoading: {
    saving: boolean;
    uploading: boolean;
    loading: boolean;
  };
}

export type EditorAction =
  | { type: "LOAD_NOTE"; payload: { noteId: string; title: string; summary: string; content: string; categoryId: string; tagIds: string[]; visibility: NoteVisibility; attachments: Attachment[] } }
  | { type: "UPDATE_TITLE"; payload: string }
  | { type: "UPDATE_SUMMARY"; payload: string }
  | { type: "UPDATE_CONTENT"; payload: string }
  | { type: "SET_CATEGORY"; payload: string }
  | { type: "SET_TAGS"; payload: string[] }
  | { type: "SET_VISIBILITY"; payload: NoteVisibility }
  | { type: "ADD_ATTACHMENT"; payload: Attachment }
  | { type: "REMOVE_ATTACHMENT"; payload: string }
  | { type: "ADD_LOCAL_FILE"; payload: LocalFile }
  | { type: "REMOVE_LOCAL_FILE"; payload: string }
  | { type: "CLEAR_LOCAL_FILES" }
  | { type: "SET_LOADING"; payload: { key: "saving" | "uploading" | "loading"; value: boolean } }
  | { type: "RESET" };

export const initialState: EditorState = {
  noteId: null,
  title: "",
  summary: "",
  content: "",
  categoryId: "",
  tagIds: [],
  visibility: NoteVisibility.PUBLIC,
  attachments: [],
  localFiles: [],
  isLoading: {
    saving: false,
    uploading: false,
    loading: false,
  },
};
