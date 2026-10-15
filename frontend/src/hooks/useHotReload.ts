import { useEffect, useRef, useState } from "react";
import { Snippet } from "../types";
import { app, onSnippetsChanged } from "../bridge";

interface HotReloadOptions {
  selectedSnippet: Snippet | null;
  setSelectedSnippet: (snippet: Snippet | null) => void;
  isDirtyRef: React.MutableRefObject<boolean>;
  refreshList: () => void;
}

interface HotReloadState {
  changedOnDisk: boolean;
  deletedOnDisk: boolean;
}

function sameContent(a: Snippet, b: Snippet): boolean {
  return (
    a.title === b.title &&
    a.description === b.description &&
    a.body === b.body &&
    a.language === b.language &&
    a.is_favorite === b.is_favorite &&
    a.tags.join("\n") === b.tags.join("\n")
  );
}

/**
 * Applies external snippet file changes to the list and the open snippet.
 * Unsaved local edits are never overwritten or dropped; changedOnDisk and deletedOnDisk flag the conflict instead.
 */
export function useHotReload({
  selectedSnippet,
  setSelectedSnippet,
  isDirtyRef,
  refreshList,
}: HotReloadOptions): HotReloadState {
  const [changedOnDisk, setChangedOnDisk] = useState(false);
  const [deletedOnDisk, setDeletedOnDisk] = useState(false);
  const selectedRef = useRef(selectedSnippet);

  useEffect(() => {
    selectedRef.current = selectedSnippet;
  }, [selectedSnippet]);

  useEffect(() => {
    setChangedOnDisk(false);
  }, [selectedSnippet?.id]);

  // Saving replaces the selected snippet object, and a save recreates the deleted file.
  useEffect(() => {
    setDeletedOnDisk(false);
  }, [selectedSnippet]);

  useEffect(() => {
    return onSnippetsChanged(async () => {
      refreshList();

      const current = selectedRef.current;
      if (!current) return;

      let snippets: Snippet[];
      try {
        snippets = await app.GetAllSnippets();
      } catch (err) {
        console.error("Failed to reload snippets:", err);
        return;
      }

      const onDisk = snippets.find((s) => s.id === current.id);
      if (!onDisk) {
        if (isDirtyRef.current) {
          setDeletedOnDisk(true);
        } else {
          setSelectedSnippet(null);
        }
        return;
      }
      setDeletedOnDisk(false);
      if (sameContent(onDisk, current)) return;

      if (isDirtyRef.current) {
        setChangedOnDisk(true);
        return;
      }
      setSelectedSnippet(onDisk);
      setChangedOnDisk(false);
    });
  }, [refreshList, setSelectedSnippet, isDirtyRef]);

  return { changedOnDisk, deletedOnDisk };
}
