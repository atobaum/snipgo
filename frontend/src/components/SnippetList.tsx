import { useEffect, useState, useCallback, useMemo } from "react";
import { Snippet } from "../types";
import { app } from "../bridge";

interface SnippetListProps {
  onSelect: (snippet: Snippet) => void;
  searchQuery?: string;
  selectedId?: string;
  refreshKey?: number; // 저장 후 목록 갱신 트리거
  selectedTag?: string | null; // 선택된 태그로 필터링
}

export function SnippetList({
  onSelect,
  searchQuery = "",
  selectedId,
  refreshKey = 0,
  selectedTag = null,
}: SnippetListProps) {
  const [snippets, setSnippets] = useState<Snippet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 태그로 필터링된 스니펫
  const filteredSnippets = useMemo(() => {
    if (!selectedTag) return snippets;
    return snippets.filter((snippet) =>
      snippet.tags.some(
        (tag) => tag.toLowerCase() === selectedTag.toLowerCase()
      )
    );
  }, [snippets, selectedTag]);

  const loadSnippets = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      let result: Snippet[];
      if (searchQuery.trim()) {
        result = await app.SearchSnippets(searchQuery);
      } else {
        result = await app.GetAllSnippets();
      }
      setSnippets(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load snippets");
    } finally {
      setLoading(false);
    }
  }, [searchQuery]);

  useEffect(() => {
    loadSnippets();
  }, [loadSnippets, refreshKey]);

  if (loading) {
    return (
      <div className="p-4">
        <p className="text-gray-500">Loading snippets...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <p className="text-red-500">Error: {error}</p>
        <button
          onClick={loadSnippets}
          className="mt-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Retry
        </button>
      </div>
    );
  }

  if (filteredSnippets.length === 0) {
    return (
      <div className="p-4">
        <p className="text-gray-500">No snippets found.</p>
      </div>
    );
  }

  return (
    <div className="divide-y divide-gray-100">
      {filteredSnippets.map((snippet) => (
        <div
          key={snippet.id}
          onClick={() => onSelect(snippet)}
          className={`px-3 py-2 cursor-pointer transition-colors ${
            selectedId === snippet.id
              ? "bg-blue-50 border-l-2 border-blue-500"
              : "hover:bg-gray-50"
          }`}
        >
          <div className="flex items-start justify-between gap-1">
            <div className="flex-1 min-w-0">
              <h3 className="font-medium text-sm leading-tight truncate">
                {snippet.title}
              </h3>
              {snippet.description && (
                <p className="text-xs text-gray-500 mt-0.5 truncate">
                  {snippet.description}
                </p>
              )}
              {snippet.tags.length > 0 && (
                <div className="mt-1 flex flex-wrap gap-0.5">
                  {snippet.tags.slice(0, 3).map((tag, idx) => (
                    <span
                      key={idx}
                      className="px-1.5 py-0.5 text-[10px] bg-blue-100 text-blue-700 rounded"
                    >
                      {tag}
                    </span>
                  ))}
                  {snippet.tags.length > 3 && (
                    <span className="px-1 py-0.5 text-[10px] text-gray-400">
                      +{snippet.tags.length - 3}
                    </span>
                  )}
                </div>
              )}
            </div>
            <div className="flex items-center gap-1 flex-shrink-0">
              {snippet.language && (
                <span className="text-[10px] text-gray-400">{snippet.language}</span>
              )}
              {snippet.is_favorite && (
                <span className="text-yellow-500 text-xs">★</span>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
