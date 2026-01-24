import { SaveStatus } from "../hooks/useSnippetAutoSave";

interface SaveStatusBadgeProps {
  status: SaveStatus;
}

export function SaveStatusBadge({ status }: SaveStatusBadgeProps) {
  if (!status) return null;

  const styles = {
    pending: "bg-orange-100 text-orange-600",
    saving: "bg-blue-100 text-blue-600",
    saved: "bg-green-100 text-green-600",
  };

  const labels = {
    pending: "Unsaved",
    saving: "Saving...",
    saved: "Saved",
  };

  return (
    <span className={`px-2 py-1 text-xs rounded ${styles[status]}`}>
      {labels[status]}
    </span>
  );
}
