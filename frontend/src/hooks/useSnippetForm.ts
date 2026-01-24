import { useState, useEffect, useMemo, useRef } from "react";
import { Snippet } from "../types";

export interface SnippetFormState {
  title: string;
  setTitle: (value: string) => void;
  description: string;
  setDescription: (value: string) => void;
  tags: string[];
  setTags: (value: string[]) => void;
  language: string;
  setLanguage: (value: string) => void;
  isFavorite: boolean;
  setIsFavorite: (value: boolean) => void;
  body: string;
  setBody: (value: string) => void;
  isDirty: boolean;
  isLoadingRef: React.RefObject<boolean>;
}

/**
 * Hook to manage snippet form state with auto-reset on snippet change
 */
export function useSnippetForm(snippet: Snippet | null): SnippetFormState {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [language, setLanguage] = useState("");
  const [isFavorite, setIsFavorite] = useState(false);
  const [body, setBody] = useState("");

  // Track if we're loading a new snippet (prevent auto-save during load)
  const isLoadingRef = useRef(false);

  // isDirty calculation (title, description, body, language only - tag/favorite save immediately)
  const isDirty = useMemo(() => {
    if (!snippet) return false;
    return (
      title !== snippet.title ||
      description !== snippet.description ||
      body !== snippet.body ||
      language !== snippet.language
    );
  }, [snippet, title, description, body, language]);

  // Reset form when snippet changes
  useEffect(() => {
    isLoadingRef.current = true;

    if (snippet) {
      setTitle(snippet.title);
      setDescription(snippet.description || "");
      setTags([...snippet.tags]);
      setLanguage(snippet.language);
      setIsFavorite(snippet.is_favorite);
      setBody(snippet.body);
    } else {
      setTitle("");
      setDescription("");
      setTags([]);
      setLanguage("");
      setIsFavorite(false);
      setBody("");
    }

    // Clear loading flag after state updates
    setTimeout(() => {
      isLoadingRef.current = false;
    }, 0);
  }, [snippet]);

  return {
    title,
    setTitle,
    description,
    setDescription,
    tags,
    setTags,
    language,
    setLanguage,
    isFavorite,
    setIsFavorite,
    body,
    setBody,
    isDirty,
    isLoadingRef,
  };
}
