import { useState, useEffect } from "react";
import { app } from "../bridge";

interface SettingsPageProps {
  onClose: () => void;
  onSettingsChange: () => void;
}

export function SettingsPage({ onClose, onSettingsChange }: SettingsPageProps) {
  const [configPath, setConfigPath] = useState("");
  const [dataDirectory, setDataDirectory] = useState("");
  const [originalDataDirectory, setOriginalDataDirectory] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadSettings = async () => {
      try {
        setLoading(true);
        const [configPathResult, dataDirectoryResult] = await Promise.all([
          app.GetConfigPath(),
          app.GetDataDirectory(),
        ]);
        setConfigPath(configPathResult);
        setDataDirectory(dataDirectoryResult);
        setOriginalDataDirectory(dataDirectoryResult);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load settings");
      } finally {
        setLoading(false);
      }
    };
    loadSettings();
  }, []);

  const handleBrowse = async () => {
    try {
      const result = await app.BrowseForDirectory();
      if (result) {
        setDataDirectory(result);
      }
    } catch (err) {
      console.error("Failed to open directory picker:", err);
    }
  };

  const handleSave = async () => {
    if (dataDirectory === originalDataDirectory) {
      onClose();
      return;
    }

    try {
      setSaving(true);
      setError(null);
      await app.SetDataDirectory(dataDirectory);
      setOriginalDataDirectory(dataDirectory);
      onSettingsChange();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  const isDirty = dataDirectory !== originalDataDirectory;

  if (loading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg shadow-xl p-6">
          <p className="text-gray-500">Loading settings...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold">Settings</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {error && (
            <div className="p-3 bg-red-50 text-red-600 rounded text-sm">
              {error}
            </div>
          )}

          {/* Config Path (read-only) */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Config File Path
            </label>
            <input
              type="text"
              value={configPath}
              readOnly
              className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
            />
            <p className="mt-1 text-xs text-gray-500">
              Set via SNIPGO_CONFIG_PATH environment variable
            </p>
          </div>

          {/* Data Directory */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Snippets Directory
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={dataDirectory}
                onChange={(e) => setDataDirectory(e.target.value)}
                className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={handleBrowse}
                className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200"
              >
                Browse
              </button>
            </div>
            <p className="mt-1 text-xs text-gray-500">
              Directory where snippet files are stored
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving || !isDirty}
            className={`px-4 py-2 rounded-lg ${
              saving || !isDirty
                ? "bg-gray-300 text-gray-500 cursor-not-allowed"
                : "bg-blue-500 text-white hover:bg-blue-600"
            }`}
          >
            {saving ? "Saving..." : "Save"}
          </button>
        </div>
      </div>
    </div>
  );
}
