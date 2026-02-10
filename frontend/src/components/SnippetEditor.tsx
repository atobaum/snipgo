import { useState, useCallback, useEffect } from "react";
import { Snippet, Variable } from "../types";
import { app } from "../bridge";
import { useSnippetForm } from "../hooks/useSnippetForm";
import { useSnippetAutoSave } from "../hooks/useSnippetAutoSave";
import { EditorHeader } from "./EditorHeader";
import { TagInput } from "./TagInput";
import { LanguageSelector } from "./LanguageSelector";
import { SnippetMetaInfo } from "./SnippetMetaInfo";
import { ActionButtons } from "./ActionButtons";
import { CodeEditor } from "./CodeEditor";
import { VariablePromptModal } from "./VariablePromptModal";

interface SnippetEditorProps {
  snippet: Snippet | null;
  onSave: (updatedSnippet: Snippet) => void;
  onDelete: () => void;
  onDirtyChange?: (isDirty: boolean) => void;
  onListRefresh?: () => void;
}

export function SnippetEditor({
  snippet,
  onSave,
  onDelete,
  onDirtyChange,
  onListRefresh,
}: SnippetEditorProps) {
  // Form state management
  const form = useSnippetForm(snippet);

  // Local UI state
  const [rawMode, setRawMode] = useState(false);
  const [rawContent, setRawContent] = useState("");
  const [wordWrap, setWordWrap] = useState(true);
  const [showVariableModal, setShowVariableModal] = useState(false);
  const [pendingVariables, setPendingVariables] = useState<Variable[]>([]);

  // Auto-save with debouncing
  const { saveStatus } = useSnippetAutoSave({
    snippet,
    title: form.title,
    description: form.description,
    tags: form.tags,
    language: form.language,
    isFavorite: form.isFavorite,
    body: form.body,
    isDirty: form.isDirty,
    isLoadingRef: form.isLoadingRef,
    onSave,
  });

  // Notify parent of dirty state changes
  useEffect(() => {
    onDirtyChange?.(form.isDirty);
  }, [form.isDirty, onDirtyChange]);

  // Tag/favorite immediate save helper
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
        onListRefresh?.();
      } catch (err) {
        alert(
          "Failed to save: " +
            (err instanceof Error ? err.message : "Unknown error")
        );
      }
    },
    [snippet, onListRefresh]
  );

  const handleAddTag = useCallback(
    (tag: string) => {
      const newTags = [...form.tags, tag];
      form.setTags(newTags);
      saveTagsAndFavorite(newTags, form.isFavorite);
    },
    [form, saveTagsAndFavorite]
  );

  const handleRemoveTag = useCallback(
    (tagToRemove: string) => {
      const newTags = form.tags.filter((tag) => tag !== tagToRemove);
      form.setTags(newTags);
      saveTagsAndFavorite(newTags, form.isFavorite);
    },
    [form, saveTagsAndFavorite]
  );

  const handleToggleFavorite = useCallback(() => {
    const newFavorite = !form.isFavorite;
    form.setIsFavorite(newFavorite);
    saveTagsAndFavorite(form.tags, newFavorite);
  }, [form, saveTagsAndFavorite]);

  const handleDelete = useCallback(async () => {
    if (!snippet) return;
    if (!confirm("Are you sure you want to delete this snippet?")) return;

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
  }, [snippet, onDelete]);

  const handleToggleRawMode = useCallback(() => {
    if (!rawMode && snippet) {
      const frontmatter = `---
id: "${snippet.id}"
title: "${form.title}"
description: "${form.description}"
tags: [${form.tags.map((t) => `"${t}"`).join(", ")}]
language: "${form.language}"
is_favorite: ${form.isFavorite}
created_at: "${snippet.created_at}"
updated_at: "${snippet.updated_at}"
---

${form.body}`;
      setRawContent(frontmatter);
    } else {
      alert(
        "Raw mode editing is read-only in this version. Please use the form fields."
      );
    }
    setRawMode(!rawMode);
  }, [rawMode, snippet, form]);

  const handleCopyWithVariables = useCallback((variables: Variable[]) => {
    setPendingVariables(variables);
    setShowVariableModal(true);
  }, []);

  const handleVariableSubmit = useCallback(
    async (values: Record<string, string>) => {
      if (!snippet) return;

      try {
        // Expand snippet with variable values
        const expanded = await app.ExpandSnippet(snippet.id, values);

        // Copy to clipboard
        await app.CopyToClipboard(expanded);

        // Save variable history
        await app.SaveVariableHistory(values);

        // Close modal
        setShowVariableModal(false);
        setPendingVariables([]);

        alert("Copied expanded snippet to clipboard!");
      } catch (err) {
        alert(
          "Failed to copy expanded snippet: " +
            (err instanceof Error ? err.message : "Unknown error")
        );
      }
    },
    [snippet]
  );

  const handleVariableCancel = useCallback(() => {
    setShowVariableModal(false);
    setPendingVariables([]);
  }, []);

  if (!snippet) {
    return (
      <div className="p-8 text-center text-gray-500">
        <p>Select a snippet to edit</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 bg-white">
        <EditorHeader
          title={form.title}
          onTitleChange={form.setTitle}
          saveStatus={saveStatus}
          isFavorite={form.isFavorite}
          onToggleFavorite={handleToggleFavorite}
          rawMode={rawMode}
          onToggleRawMode={handleToggleRawMode}
          wordWrap={wordWrap}
          onToggleWordWrap={() => setWordWrap(!wordWrap)}
        />

        {/* Description */}
        <div className="mb-2">
          <input
            type="text"
            value={form.description}
            onChange={(e) => form.setDescription(e.target.value)}
            placeholder="Description (optional)"
            className="w-full px-2 py-1 text-sm text-gray-600 border-none outline-none focus:ring-2 focus:ring-blue-500 rounded"
          />
        </div>

        {/* Tags */}
        <TagInput
          key={snippet.id}
          tags={form.tags}
          onAddTag={handleAddTag}
          onRemoveTag={handleRemoveTag}
        />

        {/* Meta Info */}
        <div className="flex items-center gap-4 text-sm text-gray-600">
          <LanguageSelector
            language={form.language}
            onLanguageChange={form.setLanguage}
          />
          <SnippetMetaInfo
            createdAt={snippet.created_at}
            updatedAt={snippet.updated_at}
          />
        </div>

        {/* Actions */}
        <ActionButtons
          body={form.body}
          snippetId={snippet.id}
          onDelete={handleDelete}
          onCopyWithVariables={handleCopyWithVariables}
        />
      </div>

      {/* Editor */}
      <div className="flex-1 min-h-0">
        <CodeEditor
          body={form.body}
          onBodyChange={form.setBody}
          language={form.language}
          wordWrap={wordWrap}
          rawMode={rawMode}
          rawContent={rawContent}
        />
      </div>

      {/* Variable Prompt Modal */}
      {showVariableModal && (
        <VariablePromptModal
          variables={pendingVariables}
          onSubmit={handleVariableSubmit}
          onCancel={handleVariableCancel}
        />
      )}
    </div>
  );
}
