import { SaveStatusBadge } from "./SaveStatusBadge";
import { SaveStatus } from "../hooks/useSnippetAutoSave";

interface EditorHeaderProps {
  title: string;
  onTitleChange: (value: string) => void;
  saveStatus: SaveStatus;
  changedOnDisk: boolean;
  deletedOnDisk: boolean;
  isFavorite: boolean;
  onToggleFavorite: () => void;
  rawMode: boolean;
  onToggleRawMode: () => void;
  wordWrap: boolean;
  onToggleWordWrap: () => void;
}

export function EditorHeader({
  title,
  onTitleChange,
  saveStatus,
  changedOnDisk,
  deletedOnDisk,
  isFavorite,
  onToggleFavorite,
  rawMode,
  onToggleRawMode,
  wordWrap,
  onToggleWordWrap,
}: EditorHeaderProps) {
  return (
    <div className="flex items-center justify-between mb-2">
      <div className="flex items-center flex-1 gap-2">
        <input
          type="text"
          value={title}
          onChange={(e) => onTitleChange(e.target.value)}
          placeholder="Title"
          className="flex-1 text-xl font-semibold border-none outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
        />
        <SaveStatusBadge status={saveStatus} />
        {changedOnDisk && (
          <span
            className="px-2 py-1 text-xs rounded bg-yellow-100 text-yellow-700"
            title="This snippet was changed outside SnipGo. Your unsaved edits are kept and will overwrite it when saved."
          >
            Changed on disk
          </span>
        )}
        {deletedOnDisk && (
          <span
            className="px-2 py-1 text-xs rounded bg-yellow-100 text-yellow-700"
            title="This snippet's file was deleted outside SnipGo. Your unsaved edits are kept and saving will recreate the file."
          >
            Deleted on disk
          </span>
        )}
      </div>
      <div className="flex items-center gap-2 ml-4">
        <button
          onClick={onToggleFavorite}
          className={`px-3 py-1 rounded ${
            isFavorite
              ? "bg-yellow-100 text-yellow-600"
              : "bg-gray-100 text-gray-600"
          }`}
        >
          {isFavorite ? "★ Favorite" : "☆ Favorite"}
        </button>
        <button
          onClick={onToggleRawMode}
          className="px-3 py-1 bg-gray-100 text-gray-600 rounded hover:bg-gray-200"
        >
          {rawMode ? "Form Mode" : "Raw Mode"}
        </button>
        <button
          onClick={onToggleWordWrap}
          className={`px-3 py-1 rounded ${
            wordWrap
              ? "bg-blue-100 text-blue-600"
              : "bg-gray-100 text-gray-600"
          }`}
          title={wordWrap ? "Word wrap enabled" : "Word wrap disabled"}
        >
          Wrap
        </button>
      </div>
    </div>
  );
}
