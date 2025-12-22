import { useState, useEffect } from 'react';
import CodeMirror from '@uiw/react-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { python } from '@codemirror/lang-python';
import { yaml } from '@codemirror/lang-yaml';
import { json } from '@codemirror/lang-json';
import { markdown } from '@codemirror/lang-markdown';
import { Snippet } from '../types';
import { app } from '../bridge';

interface SnippetEditorProps {
  snippet: Snippet | null;
  onSave: () => void;
  onDelete: () => void;
}

const languageExtensions: Record<string, any> = {
  javascript: javascript(),
  typescript: javascript({ jsx: true }),
  python: python(),
  yaml: yaml(),
  json: json(),
  markdown: markdown(),
};

export function SnippetEditor({ snippet, onSave, onDelete }: SnippetEditorProps) {
  const [title, setTitle] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [language, setLanguage] = useState('');
  const [isFavorite, setIsFavorite] = useState(false);
  const [body, setBody] = useState('');
  const [rawMode, setRawMode] = useState(false);
  const [rawContent, setRawContent] = useState('');

  useEffect(() => {
    if (snippet) {
      setTitle(snippet.title);
      setTags([...snippet.tags]);
      setLanguage(snippet.language);
      setIsFavorite(snippet.is_favorite);
      setBody(snippet.body);
    } else {
      setTitle('');
      setTags([]);
      setLanguage('');
      setIsFavorite(false);
      setBody('');
    }
  }, [snippet]);

  const handleAddTag = () => {
    const trimmed = tagInput.trim();
    if (trimmed && !tags.includes(trimmed)) {
      setTags([...tags, trimmed]);
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTags(tags.filter((tag) => tag !== tagToRemove));
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
      onSave();
    } catch (err) {
      alert('Failed to save snippet: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleDelete = async () => {
    if (!snippet) return;

    if (!confirm('Are you sure you want to delete this snippet?')) {
      return;
    }

    try {
      await app.DeleteSnippet(snippet.id);
      await app.ReloadSnippets();
      onDelete();
    } catch (err) {
      alert('Failed to delete snippet: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleCopyToClipboard = async () => {
    try {
      await app.CopyToClipboard(body);
      alert('Copied to clipboard!');
    } catch (err) {
      alert('Failed to copy to clipboard: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleToggleRawMode = () => {
    if (!rawMode) {
      // Enter raw mode - serialize current snippet
      const frontmatter = `---
id: "${snippet?.id || ''}"
title: "${title}"
tags: [${tags.map((t) => `"${t}"`).join(', ')}]
language: "${language}"
is_favorite: ${isFavorite}
created_at: "${snippet?.created_at || ''}"
updated_at: "${snippet?.updated_at || ''}"
---

${body}`;
      setRawContent(frontmatter);
    } else {
      // Exit raw mode - parse raw content (simplified, would need proper parsing)
      // For now, just show a warning
      alert('Raw mode editing is read-only in this version. Please use the form fields.');
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

  const languageExtension = language ? languageExtensions[language.toLowerCase()] : undefined;

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 bg-white">
        <div className="flex items-center justify-between mb-4">
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Title"
            className="flex-1 text-xl font-semibold border-none outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
          />
          <div className="flex items-center gap-2">
            <button
              onClick={() => setIsFavorite(!isFavorite)}
              className={`px-3 py-1 rounded ${
                isFavorite ? 'bg-yellow-100 text-yellow-600' : 'bg-gray-100 text-gray-600'
              }`}
            >
              {isFavorite ? '★ Favorite' : '☆ Favorite'}
            </button>
            <button
              onClick={handleToggleRawMode}
              className="px-3 py-1 bg-gray-100 text-gray-600 rounded hover:bg-gray-200"
            >
              {rawMode ? 'Form Mode' : 'Raw Mode'}
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
                if (e.key === 'Enter') {
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
            <span className="text-gray-500">Created: {new Date(snippet.created_at).toLocaleDateString()}</span>
          </div>
          <div>
            <span className="text-gray-500">Updated: {new Date(snippet.updated_at).toLocaleDateString()}</span>
          </div>
        </div>

        {/* Actions */}
        <div className="mt-4 flex gap-2">
          <button
            onClick={handleSave}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Save
          </button>
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


