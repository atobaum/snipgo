import { useEffect, useState, useCallback, useMemo } from "react";
import { Snippet } from "../types";
import { app } from "../bridge";
import { confirmAndDeleteSnippet } from "../actions/deleteSnippet";
import { ContextMenu, ContextMenuItem } from "./ContextMenu";

type SortField = "name" | "created" | "updated";
type SortOrder = "asc" | "desc";

interface SnippetListProps {
  onSelect: (snippet: Snippet) => void;
  onSnippetDeleted: (id: string) => void;
  searchQuery?: string;
  selectedId?: string;
  refreshKey?: number; // 저장 후 목록 갱신 트리거
  selectedTag?: string | null; // 선택된 태그로 필터링
}

function loadSortPreference(): { field: SortField; order: SortOrder } {
  try {
    const saved = localStorage.getItem("snippetSort");
    if (saved) {
      const { field, order } = JSON.parse(saved);
      return { field, order };
    }
  } catch {
    // Ignore parse errors
  }
  return { field: "updated", order: "desc" };
}

export function SnippetList({
  onSelect,
  onSnippetDeleted,
  searchQuery = "",
  selectedId,
  refreshKey = 0,
  selectedTag = null,
}: SnippetListProps) {
  const [snippets, setSnippets] = useState<Snippet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [contextMenu, setContextMenu] = useState<{ snippet: Snippet; x: number; y: number } | null>(null);
  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  // Sort state with localStorage persistence
  const [sortField, setSortField] = useState<SortField>(() => loadSortPreference().field);
  const [sortOrder, setSortOrder] = useState<SortOrder>(() => loadSortPreference().order);

  // 태그로 필터링된 스니펫
  const filteredSnippets = useMemo(() => {
    if (!selectedTag) return snippets;
    return snippets.filter((snippet) =>
      snippet.tags.some(
        (tag) => tag.toLowerCase() === selectedTag.toLowerCase()
      )
    );
  }, [snippets, selectedTag]);

  // Sort filtered snippets
  const sortedSnippets = useMemo(() => {
    const sorted = [...filteredSnippets].sort((a, b) => {
      let compare = 0;
      switch (sortField) {
        case "name":
          compare = a.title.localeCompare(b.title);
          break;
        case "created":
          compare = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
          break;
        case "updated":
          compare = new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime();
          break;
      }
      return sortOrder === "asc" ? compare : -compare;
    });
    return sorted;
  }, [filteredSnippets, sortField, sortOrder]);

  // Save sort preference to localStorage
  useEffect(() => {
    localStorage.setItem("snippetSort", JSON.stringify({ field: sortField, order: sortOrder }));
  }, [sortField, sortOrder]);

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

  const contextMenuItems = (snippet: Snippet): ContextMenuItem[] => [
    {
      label: "Delete",
      danger: true,
      onSelect: async () => {
        if (await confirmAndDeleteSnippet(snippet.id)) onSnippetDeleted(snippet.id);
      },
    },
  ];

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

  if (sortedSnippets.length === 0) {
    return (
      <div className="p-4">
        <p className="text-gray-500">No snippets found.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Sort selector */}
      <div className="flex items-center gap-1 px-2 py-1 border-b border-gray-200 text-xs flex-shrink-0">
        <select
          value={sortField}
          onChange={(e) => setSortField(e.target.value as SortField)}
          className="bg-transparent border-none outline-none text-gray-600 cursor-pointer"
        >
          <option value="name">Name</option>
          <option value="created">Created</option>
          <option value="updated">Updated</option>
        </select>
        <button
          onClick={() => setSortOrder(sortOrder === "asc" ? "desc" : "asc")}
          className="text-gray-500 hover:text-gray-700"
          title={sortOrder === "asc" ? "Ascending" : "Descending"}
        >
          {sortOrder === "asc" ? "↑" : "↓"}
        </button>
      </div>
      {/* Snippet list */}
      <div
        data-testid="snippet-list-scroll"
        onScroll={closeContextMenu}
        className="flex-1 min-h-0 divide-y divide-gray-100 overflow-y-auto"
      >
      {sortedSnippets.map((snippet) => (
        <div
          key={snippet.id}
          onClick={() => onSelect(snippet)}
          onContextMenu={(e) => {
            e.preventDefault();
            setContextMenu({ snippet, x: e.clientX, y: e.clientY });
          }}
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
      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          items={contextMenuItems(contextMenu.snippet)}
          onClose={closeContextMenu}
        />
      )}
    </div>
  );
}
