import { useState, useRef, useCallback } from "react";
import { useClickOutside } from "../hooks/useClickOutside";
import { app } from "../bridge";
import { Variable } from "../types";

interface ActionButtonsProps {
  body: string;
  snippetId: string;
  onDelete: () => void;
  onCopyWithVariables: (variables: Variable[]) => void;
}

export function ActionButtons({
  body,
  snippetId,
  onDelete,
  onCopyWithVariables,
}: ActionButtonsProps) {
  const [showMoreMenu, setShowMoreMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useClickOutside(
    menuRef,
    useCallback(() => setShowMoreMenu(false), []),
    showMoreMenu
  );

  const handleCopyToClipboard = async () => {
    try {
      // Check if snippet has variables
      const variables = await app.ExtractVariables(snippetId);

      if (variables.length > 0) {
        // Has variables - show modal
        onCopyWithVariables(variables);
      } else {
        // No variables - copy raw body
        await app.CopyToClipboard(body);
        alert("Copied to clipboard!");
      }
    } catch (err) {
      alert(
        "Failed to copy to clipboard: " +
          (err instanceof Error ? err.message : "Unknown error")
      );
    }
  };

  return (
    <div className="mt-4 flex gap-2">
      <button
        onClick={handleCopyToClipboard}
        className="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
      >
        Copy to Clipboard
      </button>
      <div ref={menuRef} className="relative">
        <button
          onClick={() => setShowMoreMenu(!showMoreMenu)}
          className="px-3 py-2 bg-gray-100 text-gray-600 rounded hover:bg-gray-200"
          title="More options"
        >
          &#8942;
        </button>
        {showMoreMenu && (
          <div className="absolute right-0 mt-1 bg-white border border-gray-200 rounded shadow-lg z-10 min-w-[120px]">
            <button
              onClick={() => {
                setShowMoreMenu(false);
                onDelete();
              }}
              className="w-full px-4 py-2 text-left text-red-600 hover:bg-red-50"
            >
              Delete
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
