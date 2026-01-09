import { useState, useEffect, useMemo, useCallback, useRef } from "react";
import CodeMirror from "@uiw/react-codemirror";
import { javascript } from "@codemirror/lang-javascript";
import { python } from "@codemirror/lang-python";
import { yaml } from "@codemirror/lang-yaml";
import { json } from "@codemirror/lang-json";
import { markdown } from "@codemirror/lang-markdown";
import type { Extension } from "@codemirror/state";
import { Snippet } from "../types";
import { app } from "../bridge";
import { useDebouncedSave } from "../hooks/useDebounce";

const AUTO_SAVE_DELAY = 1500; // 1.5초

interface SnippetEditorProps {
  snippet: Snippet | null;
  onSave: (updatedSnippet: Snippet) => void;
  onDelete: () => void;
  onDirtyChange?: (isDirty: boolean) => void;
  onListRefresh?: () => void;
}

const languageExtensions: Record<string, Extension> = {
  javascript: javascript(),
  typescript: javascript({ jsx: true }),
  python: python(),
  yaml: yaml(),
  json: json(),
  markdown: markdown(),
};

const SUPPORTED_LANGUAGES = [
  "plaintext",
  "javascript",
  "typescript",
  "python",
  "yaml",
  "json",
  "markdown",
  "bash",
  "go",
  "rust",
  "java",
  "cpp",
  "c",
  "html",
  "css",
  "sql",
];

export function SnippetEditor({
  snippet,
  onSave,
  onDelete,
  onDirtyChange,
  onListRefresh,
}: SnippetEditorProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");
  const [language, setLanguage] = useState("");
  const [isFavorite, setIsFavorite] = useState(false);
  const [body, setBody] = useState("");
  const [rawMode, setRawMode] = useState(false);
  const [rawContent, setRawContent] = useState("");
  const [saveStatus, setSaveStatus] = useState<
    "saved" | "saving" | "pending" | null
  >(null);
  const [showMoreMenu, setShowMoreMenu] = useState(false);
  // Tag autocomplete
  const [allTags, setAllTags] = useState<string[]>([]);
  const [showTagSuggestions, setShowTagSuggestions] = useState(false);
  const tagInputRef = useRef<HTMLInputElement>(null);
  // Language selector
  const [showLanguageDropdown, setShowLanguageDropdown] = useState(false);
  const [languageFilter, setLanguageFilter] = useState("");

  // 스니펫 로딩 중인지 추적 (로딩 중에는 auto-save 방지)
  const isLoadingSnippetRef = useRef(false);
  // 현재 스니펫 참조 (auto-save 콜백에서 사용)
  const snippetRef = useRef(snippet);

  // isDirty 계산 (title, description, body, language만 - tag/favorite는 즉시 저장됨)
  const isDirty = useMemo(() => {
    if (!snippet) return false;
    return (
      title !== snippet.title ||
      description !== snippet.description ||
      body !== snippet.body ||
      language !== snippet.language
    );
  }, [snippet, title, description, body, language]);

  // 필터링된 태그 제안
  const filteredTagSuggestions = useMemo(() => {
    if (!tagInput.trim()) return [];
    const lower = tagInput.toLowerCase();
    return allTags
      .filter((tag) => tag.toLowerCase().includes(lower) && !tags.includes(tag))
      .slice(0, 5);
  }, [tagInput, allTags, tags]);

  // 필터링된 언어 목록
  const filteredLanguages = useMemo(() => {
    if (!languageFilter) return SUPPORTED_LANGUAGES;
    const lower = languageFilter.toLowerCase();
    return SUPPORTED_LANGUAGES.filter((lang) =>
      lang.toLowerCase().includes(lower)
    );
  }, [languageFilter]);

  // isDirty 변경 시 부모에게 알림
  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  // 모든 태그 로드
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

  // snippetRef 업데이트
  useEffect(() => {
    snippetRef.current = snippet;
  }, [snippet]);

  // 자동 저장 함수
  const performAutoSave = useCallback(async () => {
    const currentSnippet = snippetRef.current;
    if (!currentSnippet || isLoadingSnippetRef.current) return;

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

      // 2초 후 "saved" 상태 제거
      setTimeout(() => setSaveStatus(null), 2000);
    } catch (err) {
      console.error("Auto-save failed:", err);
      setSaveStatus(null);
    }
  }, [title, description, tags, language, isFavorite, body, onSave]);

  // Debounced save 훅
  const { debouncedSave, cancelPendingSave, flushSave } = useDebouncedSave(
    performAutoSave,
    AUTO_SAVE_DELAY
  );

  // isDirty 변경 시 debounced save 트리거
  useEffect(() => {
    if (isDirty && !isLoadingSnippetRef.current) {
      setSaveStatus("pending");
      debouncedSave();
    }
  }, [isDirty, title, description, body, language, debouncedSave]);

  // 컴포넌트 언마운트 시 flush
  useEffect(() => {
    return () => {
      flushSave();
    };
  }, [flushSave]);

  // 외부 클릭 시 메뉴 닫기
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (showMoreMenu && !(e.target as Element).closest(".more-menu")) {
        setShowMoreMenu(false);
      }
      if (
        showTagSuggestions &&
        !(e.target as Element).closest(".tag-input-container")
      ) {
        setShowTagSuggestions(false);
      }
      if (
        showLanguageDropdown &&
        !(e.target as Element).closest(".language-selector")
      ) {
        setShowLanguageDropdown(false);
      }
    };
    document.addEventListener("click", handleClickOutside);
    return () => document.removeEventListener("click", handleClickOutside);
  }, [showMoreMenu, showTagSuggestions, showLanguageDropdown]);

  // 스니펫 전환 시 상태 초기화
  useEffect(() => {
    // 스니펫 전환 시 pending save 취소
    cancelPendingSave();
    isLoadingSnippetRef.current = true;

    if (snippet) {
      setTitle(snippet.title);
      setDescription(snippet.description || "");
      setTags([...snippet.tags]);
      setTagInput("");
      setLanguage(snippet.language);
      setLanguageFilter("");
      setIsFavorite(snippet.is_favorite);
      setBody(snippet.body);
      setSaveStatus(null);
      setShowTagSuggestions(false);
      setShowLanguageDropdown(false);
    } else {
      setTitle("");
      setDescription("");
      setTags([]);
      setTagInput("");
      setLanguage("");
      setLanguageFilter("");
      setIsFavorite(false);
      setBody("");
      setSaveStatus(null);
      setShowTagSuggestions(false);
      setShowLanguageDropdown(false);
    }

    // 상태 업데이트 후 로딩 플래그 해제
    setTimeout(() => {
      isLoadingSnippetRef.current = false;
    }, 0);
  }, [snippet, cancelPendingSave]);

  // tag/favorite 즉시 저장 헬퍼
  const saveTagsAndFavorite = useCallback(
    async (newTags: string[], newFavorite: boolean) => {
      if (!snippet) return;
      try {
        const updatedSnippet: Snippet = {
          ...snippet,
          tags: newTags,
          is_favorite: newFavorite,
        };
        await app.SaveSnippet(updatedSnippet);
        await app.ReloadSnippets();
        onListRefresh?.(); // 목록 갱신
      } catch (err) {
        alert(
          "Failed to save: " +
            (err instanceof Error ? err.message : "Unknown error")
        );
      }
    },
    [snippet, onListRefresh]
  );

  const handleAddTag = () => {
    const trimmed = tagInput.trim();
    if (trimmed && !tags.includes(trimmed)) {
      const newTags = [...tags, trimmed];
      setTags(newTags);
      setTagInput("");
      saveTagsAndFavorite(newTags, isFavorite); // 즉시 저장
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    const newTags = tags.filter((tag) => tag !== tagToRemove);
    setTags(newTags);
    saveTagsAndFavorite(newTags, isFavorite); // 즉시 저장
  };

  const handleToggleFavorite = () => {
    const newFavorite = !isFavorite;
    setIsFavorite(newFavorite);
    saveTagsAndFavorite(tags, newFavorite); // 즉시 저장
  };

  const handleDelete = async () => {
    if (!snippet) return;

    if (!confirm("Are you sure you want to delete this snippet?")) {
      return;
    }

    try {
      await app.DeleteSnippet(snippet.id);
      await app.ReloadSnippets();
      onDelete();
    } catch (err) {
      alert(
        "Failed to delete snippet: " +
          (err instanceof Error ? err.message : "Unknown error")
      );
    }
  };

  const handleCopyToClipboard = async () => {
    try {
      await app.CopyToClipboard(body);
      alert("Copied to clipboard!");
    } catch (err) {
      alert(
        "Failed to copy to clipboard: " +
          (err instanceof Error ? err.message : "Unknown error")
      );
    }
  };

  const handleToggleRawMode = () => {
    if (!rawMode) {
      // Enter raw mode - serialize current snippet
      const frontmatter = `---
id: "${snippet?.id || ""}"
title: "${title}"
description: "${description}"
tags: [${tags.map((t) => `"${t}"`).join(", ")}]
language: "${language}"
is_favorite: ${isFavorite}
created_at: "${snippet?.created_at || ""}"
updated_at: "${snippet?.updated_at || ""}"
---

${body}`;
      setRawContent(frontmatter);
    } else {
      // Exit raw mode - parse raw content (simplified, would need proper parsing)
      // For now, just show a warning
      alert(
        "Raw mode editing is read-only in this version. Please use the form fields."
      );
    }
    setRawMode(!rawMode);
  };

  if (!snippet) {
    return (
      <div className="p-8 text-center text-gray-500">
        <p>Select a snippet to edit</p>
      </div>
    );
  }

  const languageExtension = language
    ? languageExtensions[language.toLowerCase()]
    : undefined;

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 bg-white">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center flex-1 gap-2">
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Title"
              className="flex-1 text-xl font-semibold border-none outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
            />
            {saveStatus === "pending" && (
              <span className="px-2 py-1 text-xs bg-orange-100 text-orange-600 rounded">
                Unsaved
              </span>
            )}
            {saveStatus === "saving" && (
              <span className="px-2 py-1 text-xs bg-blue-100 text-blue-600 rounded">
                Saving...
              </span>
            )}
            {saveStatus === "saved" && (
              <span className="px-2 py-1 text-xs bg-green-100 text-green-600 rounded">
                Saved
              </span>
            )}
          </div>
          <div className="flex items-center gap-2 ml-4">
            <button
              onClick={handleToggleFavorite}
              className={`px-3 py-1 rounded ${
                isFavorite
                  ? "bg-yellow-100 text-yellow-600"
                  : "bg-gray-100 text-gray-600"
              }`}
            >
              {isFavorite ? "★ Favorite" : "☆ Favorite"}
            </button>
            <button
              onClick={handleToggleRawMode}
              className="px-3 py-1 bg-gray-100 text-gray-600 rounded hover:bg-gray-200"
            >
              {rawMode ? "Form Mode" : "Raw Mode"}
            </button>
          </div>
        </div>

        {/* Description */}
        <div className="mb-2">
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description (optional)"
            className="w-full px-2 py-1 text-sm text-gray-600 border-none outline-none focus:ring-2 focus:ring-blue-500 rounded"
          />
        </div>

        {/* Tags - Compact Inline Design with Autocomplete */}
        <div className="mb-2 relative tag-input-container">
          <div className="flex flex-wrap items-center gap-1 px-2 py-1.5 border border-gray-200 rounded bg-gray-50 focus-within:border-blue-400 focus-within:ring-1 focus-within:ring-blue-400 transition-all">
            {tags.map((tag, idx) => (
              <span
                key={idx}
                className="inline-flex items-center px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded text-xs group"
              >
                {tag}
                <button
                  onClick={() => handleRemoveTag(tag)}
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
              ref={tagInputRef}
              type="text"
              value={tagInput}
              onChange={(e) => {
                setTagInput(e.target.value);
                setShowTagSuggestions(true);
              }}
              onFocus={() => setShowTagSuggestions(true)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  if (filteredTagSuggestions.length > 0) {
                    const newTags = [...tags, filteredTagSuggestions[0]];
                    setTags(newTags);
                    setTagInput("");
                    saveTagsAndFavorite(newTags, isFavorite);
                  } else {
                    handleAddTag();
                  }
                  setShowTagSuggestions(false);
                } else if (e.key === "Escape") {
                  setShowTagSuggestions(false);
                } else if (
                  e.key === "Backspace" &&
                  tagInput === "" &&
                  tags.length > 0
                ) {
                  handleRemoveTag(tags[tags.length - 1]);
                }
              }}
              placeholder={tags.length === 0 ? "Add tags..." : ""}
              className="flex-1 min-w-[60px] px-1 py-0.5 bg-transparent border-none outline-none text-xs placeholder-gray-400"
            />
          </div>
          {/* Tag Suggestions Dropdown */}
          {showTagSuggestions && filteredTagSuggestions.length > 0 && (
            <div className="absolute left-0 right-0 top-full mt-1 bg-white border border-gray-200 rounded shadow-lg z-20 max-h-32 overflow-y-auto">
              {filteredTagSuggestions.map((tag) => (
                <button
                  key={tag}
                  onClick={() => {
                    const newTags = [...tags, tag];
                    setTags(newTags);
                    setTagInput("");
                    saveTagsAndFavorite(newTags, isFavorite);
                    setShowTagSuggestions(false);
                    tagInputRef.current?.focus();
                  }}
                  className="w-full px-3 py-1.5 text-left text-xs hover:bg-blue-50 text-gray-700"
                >
                  {tag}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Meta Info */}
        <div className="flex items-center gap-4 text-sm text-gray-600">
          <div className="relative language-selector">
            <label className="mr-2">Language:</label>
            <input
              type="text"
              value={showLanguageDropdown ? languageFilter : language}
              onChange={(e) => {
                setLanguageFilter(e.target.value);
                if (!showLanguageDropdown) setShowLanguageDropdown(true);
              }}
              onFocus={() => {
                setLanguageFilter(language);
                setShowLanguageDropdown(true);
              }}
              onKeyDown={(e) => {
                if (e.key === "Escape") {
                  setShowLanguageDropdown(false);
                  setLanguageFilter("");
                } else if (e.key === "Enter" && filteredLanguages.length > 0) {
                  e.preventDefault();
                  setLanguage(filteredLanguages[0]);
                  setShowLanguageDropdown(false);
                  setLanguageFilter("");
                }
              }}
              placeholder="Select..."
              className="px-2 py-1 border border-gray-300 rounded w-32 text-xs"
            />
            {showLanguageDropdown && (
              <div className="absolute left-16 top-full mt-1 bg-white border border-gray-200 rounded shadow-lg z-20 max-h-48 overflow-y-auto min-w-[140px]">
                {filteredLanguages.length > 0 ? (
                  filteredLanguages.map((lang) => (
                    <button
                      key={lang}
                      onClick={() => {
                        setLanguage(lang);
                        setShowLanguageDropdown(false);
                        setLanguageFilter("");
                      }}
                      className={`w-full px-3 py-1.5 text-left text-xs hover:bg-blue-50 ${
                        language === lang
                          ? "bg-blue-100 text-blue-800"
                          : "text-gray-700"
                      }`}
                    >
                      {lang}
                    </button>
                  ))
                ) : (
                  <div className="px-3 py-2 text-xs text-gray-500">
                    Press Enter to use &quot;{languageFilter}&quot;
                  </div>
                )}
                {languageFilter &&
                  !filteredLanguages.includes(languageFilter.toLowerCase()) &&
                  filteredLanguages.length > 0 && (
                    <button
                      onClick={() => {
                        setLanguage(languageFilter);
                        setShowLanguageDropdown(false);
                        setLanguageFilter("");
                      }}
                      className="w-full px-3 py-1.5 text-left text-xs hover:bg-blue-50 text-blue-600 border-t border-gray-100"
                    >
                      Use &quot;{languageFilter}&quot;
                    </button>
                  )}
              </div>
            )}
          </div>
          <div>
            <span className="text-gray-500 text-xs">
              Created: {new Date(snippet.created_at).toLocaleDateString()}
            </span>
          </div>
          <div>
            <span className="text-gray-500 text-xs">
              Updated: {new Date(snippet.updated_at).toLocaleDateString()}
            </span>
          </div>
        </div>

        {/* Actions */}
        <div className="mt-4 flex gap-2">
          <button
            onClick={handleCopyToClipboard}
            className="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
          >
            Copy to Clipboard
          </button>
          <div className="relative more-menu">
            <button
              onClick={() => setShowMoreMenu(!showMoreMenu)}
              className="px-3 py-2 bg-gray-100 text-gray-600 rounded hover:bg-gray-200"
              title="More options"
            >
              &#8942;
            </button>
            {showMoreMenu && (
              <div className="absolute right-0 mt-1 bg-white border border-gray-200 rounded shadow-lg z-10 min-w-[120px]">
                <button
                  onClick={() => {
                    setShowMoreMenu(false);
                    handleDelete();
                  }}
                  className="w-full px-4 py-2 text-left text-red-600 hover:bg-red-50"
                >
                  Delete
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Editor */}
      <div className="flex-1 overflow-auto">
        {rawMode ? (
          <textarea
            value={rawContent}
            onChange={(e) => setRawContent(e.target.value)}
            className="w-full h-full p-4 font-mono text-sm border-none outline-none resize-none"
            readOnly
          />
        ) : (
          <CodeMirror
            value={body}
            onChange={(value) => setBody(value)}
            extensions={languageExtension ? [languageExtension] : []}
            theme="light"
            basicSetup={{
              lineNumbers: true,
              foldGutter: true,
              dropCursor: false,
              allowMultipleSelections: false,
            }}
            className="h-full"
          />
        )}
      </div>
    </div>
  );
}
