import { useEffect, useState } from "react";
import { app } from "../bridge";

interface TagListProps {
  selectedTag: string | null;
  onSelectTag: (tag: string | null) => void;
  refreshKey?: number;
}

export function TagList({
  selectedTag,
  onSelectTag,
  refreshKey = 0,
}: TagListProps) {
  const [tags, setTags] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadTags = async () => {
      try {
        setLoading(true);
        const result = await app.GetAllTags();
        setTags(result);
      } catch (err) {
        console.error("Failed to load tags:", err);
      } finally {
        setLoading(false);
      }
    };
    loadTags();
  }, [refreshKey]);

  if (loading) {
    return null;
  }

  if (tags.length === 0) {
    return null;
  }

  return (
    <div className="border-b border-gray-200 pb-4 mb-2">
      <div className="flex items-center justify-between px-4 py-2">
        <h2 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">
          Tags
        </h2>
        {selectedTag && (
          <button
            onClick={() => onSelectTag(null)}
            className="text-xs text-gray-400 hover:text-gray-600"
          >
            Clear
          </button>
        )}
      </div>
      <div className="px-4 flex flex-wrap gap-2">
        {tags.map((tag) => (
          <button
            key={tag}
            onClick={() => onSelectTag(selectedTag === tag ? null : tag)}
            className={`px-2 py-1 text-xs rounded transition-colors ${
              selectedTag === tag
                ? "bg-blue-500 text-white"
                : "bg-blue-100 text-blue-800 hover:bg-blue-200"
            }`}
          >
            {tag}
          </button>
        ))}
      </div>
    </div>
  );
}
