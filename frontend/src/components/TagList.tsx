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
    <div className="border-b border-gray-100 pb-2 mb-1">
      <div className="flex items-center justify-between px-3 py-1">
        <h2 className="text-[10px] font-semibold text-gray-400 uppercase tracking-wider">
          Tags
        </h2>
        {selectedTag && (
          <button
            onClick={() => onSelectTag(null)}
            className="text-[10px] text-gray-400 hover:text-gray-600"
          >
            Clear
          </button>
        )}
      </div>
      <div className="px-3 flex flex-wrap gap-1">
        {tags.map((tag) => (
          <button
            key={tag}
            onClick={() => onSelectTag(selectedTag === tag ? null : tag)}
            className={`px-1.5 py-0.5 text-[10px] rounded transition-colors ${
              selectedTag === tag
                ? "bg-blue-500 text-white"
                : "bg-blue-100 text-blue-700 hover:bg-blue-200"
            }`}
          >
            {tag}
          </button>
        ))}
      </div>
    </div>
  );
}
