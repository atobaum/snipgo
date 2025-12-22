import { useEffect, useState } from 'react';
import { Snippet } from '../types';
import { app } from '../bridge';

interface SnippetListProps {
  onSelect: (snippet: Snippet) => void;
  searchQuery?: string;
}

export function SnippetList({ onSelect, searchQuery = '' }: SnippetListProps) {
  const [snippets, setSnippets] = useState<Snippet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadSnippets();
  }, [searchQuery]);

  const loadSnippets = async () => {
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
      setError(err instanceof Error ? err.message : 'Failed to load snippets');
    } finally {
      setLoading(false);
    }
  };

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

  if (snippets.length === 0) {
    return (
      <div className="p-4">
        <p className="text-gray-500">No snippets found.</p>
      </div>
    );
  }

  return (
    <div className="divide-y divide-gray-200">
      {snippets.map((snippet) => (
        <div
          key={snippet.id}
          onClick={() => onSelect(snippet)}
          className="p-4 hover:bg-gray-50 cursor-pointer transition-colors"
        >
          <div className="flex items-center justify-between">
            <div className="flex-1">
              <h3 className="font-semibold text-lg">{snippet.title}</h3>
              {snippet.tags.length > 0 && (
                <div className="mt-1 flex flex-wrap gap-1">
                  {snippet.tags.map((tag, idx) => (
                    <span
                      key={idx}
                      className="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}
              {snippet.language && (
                <span className="mt-1 inline-block text-xs text-gray-500">
                  {snippet.language}
                </span>
              )}
            </div>
            {snippet.is_favorite && (
              <span className="text-yellow-500">★</span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}


