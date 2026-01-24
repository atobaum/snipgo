import { useState, useEffect, useCallback, useRef } from "react";
import { Snippet } from "../types";
import { app } from "../bridge";
import { useDebouncedSave } from "./useDebounce";

export type SaveStatus = "saved" | "saving" | "pending" | null;

export interface SnippetAutoSaveOptions {
  snippet: Snippet | null;
  title: string;
  description: string;
  tags: string[];
  language: string;
  isFavorite: boolean;
  body: string;
  isDirty: boolean;
  isLoadingRef: React.RefObject<boolean>;
  onSave: (updatedSnippet: Snippet) => void;
  delay?: number;
}

export interface SnippetAutoSaveState {
  saveStatus: SaveStatus;
  cancelPendingSave: () => void;
  flushSave: () => void;
}

const DEFAULT_AUTO_SAVE_DELAY = 1500;

/**
 * Hook to handle auto-save with debouncing for snippet editor
 */
export function useSnippetAutoSave({
  snippet,
  title,
  description,
  tags,
  language,
  isFavorite,
  body,
  isDirty,
  isLoadingRef,
  onSave,
  delay = DEFAULT_AUTO_SAVE_DELAY,
}: SnippetAutoSaveOptions): SnippetAutoSaveState {
  const [saveStatus, setSaveStatus] = useState<SaveStatus>(null);
  const snippetRef = useRef(snippet);

  // Keep snippet ref up to date
  useEffect(() => {
    snippetRef.current = snippet;
  }, [snippet]);

  // Auto-save function
  const performAutoSave = useCallback(async () => {
    const currentSnippet = snippetRef.current;
    if (!currentSnippet || isLoadingRef.current) return;

    try {
      setSaveStatus("saving");
      const updatedSnippet: Snippet = {
        ...currentSnippet,
        title,
        description,
        tags,
        language,
        is_favorite: isFavorite,
        body,
      };
      await app.SaveSnippet(updatedSnippet);
      await app.ReloadSnippets();
      onSave(updatedSnippet);
      setSaveStatus("saved");

      // Clear "saved" status after 2 seconds
      setTimeout(() => setSaveStatus(null), 2000);
    } catch (err) {
      console.error("Auto-save failed:", err);
      setSaveStatus(null);
    }
  }, [title, description, tags, language, isFavorite, body, onSave, isLoadingRef]);

  // Debounced save hook
  const { debouncedSave, cancelPendingSave, flushSave } = useDebouncedSave(
    performAutoSave,
    delay
  );

  // Trigger debounced save when dirty
  useEffect(() => {
    if (isDirty && !isLoadingRef.current) {
      setSaveStatus("pending");
      debouncedSave();
    }
  }, [isDirty, title, description, body, language, debouncedSave, isLoadingRef]);

  // Flush save on unmount
  useEffect(() => {
    return () => {
      flushSave();
    };
  }, [flushSave]);

  // Reset save status when snippet changes
  useEffect(() => {
    cancelPendingSave();
    setSaveStatus(null);
  }, [snippet, cancelPendingSave]);

  return {
    saveStatus,
    cancelPendingSave,
    flushSave,
  };
}
