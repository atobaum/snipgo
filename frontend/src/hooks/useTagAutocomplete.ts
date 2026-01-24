import { useState, useEffect, useMemo } from "react";
import { app } from "../bridge";

export interface TagAutocompleteState {
  tagInput: string;
  setTagInput: (value: string) => void;
  allTags: string[];
  showSuggestions: boolean;
  setShowSuggestions: (value: boolean) => void;
  filteredSuggestions: string[];
}

/**
 * Hook to handle tag input with autocomplete suggestions
 * @param currentTags - Currently selected tags (to exclude from suggestions)
 */
export function useTagAutocomplete(currentTags: string[]): TagAutocompleteState {
  const [tagInput, setTagInput] = useState("");
  const [allTags, setAllTags] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);

  // Load all tags from backend
  useEffect(() => {
    const loadAllTags = async () => {
      try {
        const result = await app.GetAllTags();
        setAllTags(result);
      } catch (err) {
        console.error("Failed to load tags:", err);
      }
    };
    loadAllTags();
  }, []);

  // Filter suggestions based on input
  const filteredSuggestions = useMemo(() => {
    if (!tagInput.trim()) return [];
    const lower = tagInput.toLowerCase();
    return allTags
      .filter((tag) => tag.toLowerCase().includes(lower) && !currentTags.includes(tag))
      .slice(0, 5);
  }, [tagInput, allTags, currentTags]);

  return {
    tagInput,
    setTagInput,
    allTags,
    showSuggestions,
    setShowSuggestions,
    filteredSuggestions,
  };
}
