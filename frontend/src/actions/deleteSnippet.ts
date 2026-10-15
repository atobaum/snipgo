import { app } from "../bridge";

export async function confirmAndDeleteSnippet(id: string): Promise<boolean> {
  try {
    const confirmed = await app.ConfirmDialog(
      "Delete Snippet",
      "Are you sure you want to delete this snippet?"
    );
    if (!confirmed) return false;

    await app.DeleteSnippet(id);
    await app.ReloadSnippets();
    return true;
  } catch (err) {
    app.AlertDialog("SnipGo",
      "Failed to delete snippet: " +
        (err instanceof Error ? err.message : "Unknown error")
    );
    return false;
  }
}
