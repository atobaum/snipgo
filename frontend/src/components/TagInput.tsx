import { useRef, useCallback } from "react";
import { useClickOutside } from "../hooks/useClickOutside";
import { useTagAutocomplete } from "../hooks/useTagAutocomplete";

interface TagInputProps {
  tags: string[];
  onAddTag: (tag: string) => void;
  onRemoveTag: (tag: string) => void;
}

export function TagInput({ tags, onAddTag, onRemoveTag }: TagInputProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const {
    tagInput,
    setTagInput,
    showSuggestions,
    setShowSuggestions,
    filteredSuggestions,
  } = useTagAutocomplete(tags);

  useClickOutside(
    containerRef,
    useCallback(() => setShowSuggestions(false), [setShowSuggestions]),
    showSuggestions
  );

  const handleAddTag = () => {
    const trimmed = tagInput.trim();
    if (trimmed && !tags.includes(trimmed)) {
      onAddTag(trimmed);
      setTagInput("");
    }
  };

  const handleSelectSuggestion = (tag: string) => {
    onAddTag(tag);
    setTagInput("");
    setShowSuggestions(false);
    inputRef.current?.focus();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      if (filteredSuggestions.length > 0) {
        handleSelectSuggestion(filteredSuggestions[0]);
      } else {
        handleAddTag();
      }
      setShowSuggestions(false);
    } else if (e.key === "Escape") {
      setShowSuggestions(false);
    } else if (e.key === "Backspace" && tagInput === "" && tags.length > 0) {
      onRemoveTag(tags[tags.length - 1]);
    }
  };

  return (
    <div ref={containerRef} className="mb-2 relative tag-input-container">
      <div className="flex flex-wrap items-center gap-1 px-2 py-1.5 border border-gray-200 rounded bg-gray-50 focus-within:border-blue-400 focus-within:ring-1 focus-within:ring-blue-400 transition-all">
        {tags.map((tag, idx) => (
          <span
            key={idx}
            className="inline-flex items-center px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded text-xs group"
          >
            {tag}
            <button
              onClick={() => onRemoveTag(tag)}
              className="ml-0.5 text-blue-500 hover:text-blue-700 opacity-60 group-hover:opacity-100"
            >
              <svg
                className="w-3 h-3"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          </span>
        ))}
        <input
          ref={inputRef}
          type="text"
          value={tagInput}
          onChange={(e) => {
            setTagInput(e.target.value);
            setShowSuggestions(true);
          }}
          onFocus={() => setShowSuggestions(true)}
          onKeyDown={handleKeyDown}
          placeholder={tags.length === 0 ? "Add tags..." : ""}
          className="flex-1 min-w-[60px] px-1 py-0.5 bg-transparent border-none outline-none text-xs placeholder-gray-400"
        />
      </div>
      {/* Tag Suggestions Dropdown */}
      {showSuggestions && filteredSuggestions.length > 0 && (
        <div className="absolute left-0 right-0 top-full mt-1 bg-white border border-gray-200 rounded shadow-lg z-20 max-h-32 overflow-y-auto">
          {filteredSuggestions.map((tag) => (
            <button
              key={tag}
              onClick={() => handleSelectSuggestion(tag)}
              className="w-full px-3 py-1.5 text-left text-xs hover:bg-blue-50 text-gray-700"
            >
              {tag}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
