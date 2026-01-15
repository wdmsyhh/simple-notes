import { createContext, useContext, useReducer, useMemo } from "react";
import { editorReducer } from "./reducer";
import type { EditorAction, EditorState } from "./types";
import { initialState } from "./types";

interface EditorContextValue {
  state: EditorState;
  dispatch: React.Dispatch<EditorAction>;
}

const EditorContext = createContext<EditorContextValue | null>(null);

export const useEditorContext = () => {
  const context = useContext(EditorContext);
  if (!context) {
    throw new Error("useEditorContext must be used within EditorProvider");
  }
  return context;
};

interface EditorProviderProps {
  children: React.ReactNode;
}

export const EditorProvider: React.FC<EditorProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(editorReducer, initialState);

  const value = useMemo<EditorContextValue>(
    () => ({
      state,
      dispatch,
    }),
    [state],
  );

  return <EditorContext.Provider value={value}>{children}</EditorContext.Provider>;
};

