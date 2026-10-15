import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SettingsPage } from "./SettingsPage";
import { SyncStatus } from "../types";

let emitSyncStatus: (s: SyncStatus) => void = () => {};
const unsubscribeSyncStatus = vi.fn();

vi.mock("../bridge", () => ({
  onSyncStatus: vi.fn((cb: (s: SyncStatus) => void) => {
    emitSyncStatus = cb;
    return unsubscribeSyncStatus;
  }),
  app: {
    GetConfigPath: vi.fn().mockResolvedValue("/c/config.yaml"),
    GetDataDirectory: vi.fn().mockResolvedValue("/data"),
    SetDataDirectory: vi.fn().mockResolvedValue(undefined),
    BrowseForDirectory: vi.fn(),
    GetSyncSettings: vi.fn(),
    SaveSyncSettings: vi.fn().mockResolvedValue(undefined),
    GetSyncStatus: vi.fn(),
    SyncNow: vi.fn(),
  },
}));

describe("SettingsPage sync", () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    const { app } = await import("../bridge");
    vi.mocked(app.GetSyncSettings).mockResolvedValue({ enabled: false, intervalMinutes: 10, conflict: "manual" });
    vi.mocked(app.GetSyncStatus).mockResolvedValue({ lastRunAt: "", ok: false, manual: false, summary: "", errorKind: "", message: "", conflictFiles: [], paused: false });
    vi.mocked(app.SyncNow).mockResolvedValue({ lastRunAt: "2026-10-15T21:50:00Z", ok: true, manual: true, summary: "Synced: pushed", errorKind: "", message: "", conflictFiles: [], paused: false });
  });

  it("saves changed sync settings without touching the data directory", async () => {
    const { app } = await import("../bridge");
    const onClose = vi.fn();
    render(<SettingsPage onClose={onClose} onSettingsChange={vi.fn()} />);
    await userEvent.click(await screen.findByLabelText("Automatic sync"));
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(app.SaveSyncSettings).toHaveBeenCalledWith({ enabled: true, intervalMinutes: 10, conflict: "manual" }));
    expect(app.SetDataDirectory).not.toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it("shows a validation error for an interval below 1", async () => {
    const { app } = await import("../bridge");
    render(<SettingsPage onClose={vi.fn()} onSettingsChange={vi.fn()} />);
    const interval = await screen.findByLabelText("Interval (minutes)");
    await userEvent.clear(interval);
    await userEvent.type(interval, "0");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Interval must be at least 1 minute")).toBeInTheDocument();
    expect(app.SaveSyncSettings).not.toHaveBeenCalled();
  });

  it("shows an error and re-enables sync now when sync now rejects", async () => {
    const { app } = await import("../bridge");
    vi.mocked(app.SyncNow).mockRejectedValue("boom");
    render(<SettingsPage onClose={vi.fn()} onSettingsChange={vi.fn()} />);
    await userEvent.click(await screen.findByRole("button", { name: "Sync now" }));
    expect(await screen.findByText("boom")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sync now" })).toBeEnabled();
  });

  it("keeps the data directory usable when sync settings fail to load", async () => {
    const { app } = await import("../bridge");
    vi.mocked(app.GetSyncSettings).mockRejectedValue("no git");
    render(<SettingsPage onClose={vi.fn()} onSettingsChange={vi.fn()} />);
    expect(await screen.findByDisplayValue("/data")).toBeInTheDocument();
    expect(screen.getByText("Failed to load sync settings: no git")).toBeInTheDocument();
    expect(screen.queryByLabelText("Automatic sync")).not.toBeInTheDocument();
  });

  it("runs sync now and shows the result", async () => {
    const { app } = await import("../bridge");
    render(<SettingsPage onClose={vi.fn()} onSettingsChange={vi.fn()} />);
    await userEvent.click(await screen.findByRole("button", { name: "Sync now" }));
    expect(app.SyncNow).toHaveBeenCalled();
    expect(await screen.findByText(/Synced: pushed/)).toBeInTheDocument();
  });

  it("updates the last run line and paused notice from sync events while open", async () => {
    const { unmount } = render(<SettingsPage onClose={vi.fn()} onSettingsChange={vi.fn()} />);
    expect(await screen.findByText("Never synced")).toBeInTheDocument();
    act(() =>
      emitSyncStatus({
        lastRunAt: "2026-10-15T21:50:00Z",
        ok: false,
        manual: false,
        summary: "",
        errorKind: "conflict",
        message: "Sync conflict in 1 file(s): a.md.",
        conflictFiles: ["a.md"],
        paused: true,
      }),
    );
    expect(screen.getByText(/failed: Sync conflict in 1 file\(s\): a\.md\./)).toBeInTheDocument();
    expect(screen.getByText(/Automatic sync is paused/)).toBeInTheDocument();
    expect(screen.getByText("a.md")).toBeInTheDocument();
    unmount();
    expect(unsubscribeSyncStatus).toHaveBeenCalled();
  });
});
