import { SyncConflict, SyncSettings, SyncStatus } from "../types";

interface SyncSettingsSectionProps {
  value: SyncSettings;
  onChange: (value: SyncSettings) => void;
  status: SyncStatus;
  onSyncNow: () => void;
  syncing: boolean;
}

const CONFLICT_OPTIONS: { value: SyncConflict; label: string }[] = [
  { value: "manual", label: "Stop and let me resolve" },
  { value: "local", label: "Prefer this machine" },
  { value: "remote", label: "Prefer remote" },
];

function lastRunText(status: SyncStatus): string {
  if (!status.lastRunAt) return "Never synced";
  const at = new Date(status.lastRunAt).toLocaleString();
  return status.ok ? `Last sync ${at}: ${status.summary}` : `Last sync ${at} failed: ${status.message}`;
}

export function SyncSettingsSection({ value, onChange, status, onSyncNow, syncing }: SyncSettingsSectionProps) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium text-gray-700">Git Sync</h3>
      <label className="flex items-center gap-2 text-sm text-gray-700">
        <input
          type="checkbox"
          checked={value.enabled}
          onChange={(e) => onChange({ ...value, enabled: e.target.checked })}
        />
        Automatic sync
      </label>
      <div className="flex gap-4">
        <label className="flex flex-col text-sm text-gray-700">
          Interval (minutes)
          <input
            type="number"
            min={1}
            value={Number.isNaN(value.intervalMinutes) ? "" : value.intervalMinutes}
            onChange={(e) => onChange({ ...value, intervalMinutes: e.target.valueAsNumber })}
            className="mt-1 w-28 px-3 py-2 border border-gray-300 rounded-lg"
          />
        </label>
        <label className="flex flex-col text-sm text-gray-700">
          On conflict
          <select
            value={value.conflict}
            onChange={(e) => onChange({ ...value, conflict: e.target.value as SyncConflict })}
            className="mt-1 px-3 py-2 border border-gray-300 rounded-lg"
          >
            {CONFLICT_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </label>
      </div>
      <div className="flex items-center gap-3">
        <button
          onClick={onSyncNow}
          disabled={syncing}
          className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50"
        >
          {syncing ? "Syncing..." : "Sync now"}
        </button>
        <p className="text-xs text-gray-500 flex-1 min-w-0 break-words">{lastRunText(status)}</p>
      </div>
      {status.paused && (
        <div className="p-3 bg-yellow-50 text-yellow-800 rounded text-sm">
          <p>Automatic sync is paused until a manual sync succeeds. Conflicted files:</p>
          <ul className="mt-1 list-disc list-inside">
            {status.conflictFiles.map((f) => (
              <li key={f}>{f}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
