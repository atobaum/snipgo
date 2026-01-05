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

export function SnippetEditor({
  snippet,
  onSave,
  onDelete,
  onDirtyChange,
  onListRefresh,
}: SnippetEditorProps) {
  const [title, setTitle] = useState("");
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

  // 스니펫 로딩 중인지 추적 (로딩 중에는 auto-save 방지)
  const isLoadingSnippetRef = useRef(false);
  // 현재 스니펫 참조 (auto-save 콜백에서 사용)
  const snippetRef = useRef(snippet);

  // isDirty 계산 (title, body, language만 - tag/favorite는 즉시 저장됨)
  const isDirty = useMemo(() => {
    if (!snippet) return false;
    return (
      title !== snippet.title ||
      body !== snippet.body ||
      language !== snippet.language
    );
  }, [snippet, title, body, language]);

  // isDirty 변경 시 부모에게 알림
  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

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
  }, [title, tags, language, isFavorite, body, onSave]);

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
  }, [isDirty, title, body, language, debouncedSave]);

  // 컴포넌트 언마운트 시 flush
  useEffect(() => {
    return () => {
      flushSave();
    };
  }, [flushSave]);

  // 스니펫 전환 시 상태 초기화
  useEffect(() => {
    // 스니펫 전환 시 pending save 취소
    cancelPendingSave();
    isLoadingSnippetRef.current = true;

    if (snippet) {
      setTitle(snippet.title);
      setTags([...snippet.tags]);
      setTagInput(""); // tagInput 초기화
      setLanguage(snippet.language);
      setIsFavorite(snippet.is_favorite);
      setBody(snippet.body);
      setSaveStatus(null);
    } else {
      setTitle("");
      setTags([]);
      setTagInput("");
      setLanguage("");
      setIsFavorite(false);
      setBody("");
      setSaveStatus(null);
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

  const handleSave = async () => {
    if (!snippet) return;

    try {
      const updatedSnippet: Snippet = {
        ...snippet,
        title,
        tags,
        language,
        is_favorite: isFavorite,
        body,
      };
      await app.SaveSnippet(updatedSnippet);
      await app.ReloadSnippets();
      onSave(updatedSnippet);
    } catch (err) {
      alert(
        "Failed to save snippet: " +
          (err instanceof Error ? err.message : "Unknown error")
      );
    }
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
        <div className="flex items-center justify-between mb-4">
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

        {/* Tags */}
        <div className="mb-2">
          <div className="flex flex-wrap gap-2 items-center">
            {tags.map((tag, idx) => (
              <span
                key={idx}
                className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-sm flex items-center gap-1"
              >
                {tag}
                <button
                  onClick={() => handleRemoveTag(tag)}
                  className="text-blue-600 hover:text-blue-800"
                >
                  ×
                </button>
              </span>
            ))}
            <input
              type="text"
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyPress={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddTag();
                }
              }}
              placeholder="Add tag..."
              className="px-2 py-1 border border-gray-300 rounded text-sm"
            />
          </div>
        </div>

        {/* Meta Info */}
        <div className="flex items-center gap-4 text-sm text-gray-600">
          <div>
            <label className="mr-2">Language:</label>
            <input
              type="text"
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
              placeholder="e.g., javascript, python"
              className="px-2 py-1 border border-gray-300 rounded"
            />
          </div>
          <div>
            <span className="text-gray-500">
              Created: {new Date(snippet.created_at).toLocaleDateString()}
            </span>
          </div>
          <div>
            <span className="text-gray-500">
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
          <button
            onClick={handleDelete}
            className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Delete
          </button>
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
