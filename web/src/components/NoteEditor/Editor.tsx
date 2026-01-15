import React, { forwardRef, useImperativeHandle, useRef, useEffect } from "react";
import { useEditorContext } from "./context";

export interface EditorRefActions {
  focus: () => void;
  getContent: () => string;
  setContent: (content: string) => void;
}

interface EditorProps {
  placeholder?: string;
  autoFocus?: boolean;
  onContentChange?: (content: string) => void;
}

const Editor = forwardRef<EditorRefActions, EditorProps>(
  ({ placeholder = "此刻的想法...", autoFocus = false, onContentChange }, ref) => {
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const { state, dispatch } = useEditorContext();

    useImperativeHandle(ref, () => ({
      focus: () => {
        textareaRef.current?.focus();
      },
      getContent: () => {
        return state.content;
      },
      setContent: (content: string) => {
        dispatch({ type: "UPDATE_CONTENT", payload: content });
      },
    }));

    useEffect(() => {
      if (autoFocus) {
        textareaRef.current?.focus();
      }
    }, [autoFocus]);

    const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const content = e.target.value;
      dispatch({ type: "UPDATE_CONTENT", payload: content });
      onContentChange?.(content);
    };

    return (
      <textarea
        ref={textareaRef}
        className="note-editor-content"
        placeholder={placeholder}
        value={state.content}
        onInput={handleInput}
        style={{ minHeight: "300px", maxHeight: "1200px" }}
      />
    );
  }
);

Editor.displayName = "Editor";

export default Editor;

