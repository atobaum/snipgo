import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SettingsPage } from "./SettingsPage";

vi.mock("../bridge", () => ({
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

  it("runs sync now and shows the result", async () => {
    const { app } = await import("../bridge");
    render(<SettingsPage onClose={vi.fn()} onSettingsChange={vi.fn()} />);
    await userEvent.click(await screen.findByRole("button", { name: "Sync now" }));
    expect(app.SyncNow).toHaveBeenCalled();
    expect(await screen.findByText(/Synced: pushed/)).toBeInTheDocument();
  });
});
