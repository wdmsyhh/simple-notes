import type { EditorAction, EditorState } from "./types";
import { initialState } from "./types";

export function editorReducer(state: EditorState, action: EditorAction): EditorState {
  switch (action.type) {
    case "LOAD_NOTE":
      return {
        ...state,
        noteId: action.payload.noteId,
        title: action.payload.title,
        summary: action.payload.summary,
        content: action.payload.content,
        categoryId: action.payload.categoryId,
        tagIds: action.payload.tagIds,
        visibility: action.payload.visibility,
        attachments: action.payload.attachments,
        localFiles: [],
      };

    case "UPDATE_TITLE":
      return {
        ...state,
        title: action.payload,
      };

    case "UPDATE_SUMMARY":
      return {
        ...state,
        summary: action.payload,
      };

    case "UPDATE_CONTENT":
      return {
        ...state,
        content: action.payload,
      };

    case "SET_CATEGORY":
      return {
        ...state,
        categoryId: action.payload,
      };

    case "SET_TAGS":
      return {
        ...state,
        tagIds: action.payload,
      };

    case "SET_VISIBILITY":
      return {
        ...state,
        visibility: action.payload,
      };

    case "ADD_ATTACHMENT":
      return {
        ...state,
        attachments: [...state.attachments, action.payload],
      };

    case "REMOVE_ATTACHMENT":
      return {
        ...state,
        attachments: state.attachments.filter((a) => a.name !== action.payload),
      };

    case "ADD_LOCAL_FILE":
      return {
        ...state,
        localFiles: [...state.localFiles, action.payload],
      };

    case "REMOVE_LOCAL_FILE":
      return {
        ...state,
        localFiles: state.localFiles.filter((f) => f.previewUrl !== action.payload),
      };

    case "CLEAR_LOCAL_FILES":
      return {
        ...state,
        localFiles: [],
      };

    case "SET_LOADING":
      return {
        ...state,
        isLoading: {
          ...state.isLoading,
          [action.payload.key]: action.payload.value,
        },
      };

    case "RESET":
      return initialState;

    default:
      return state;
  }
}

