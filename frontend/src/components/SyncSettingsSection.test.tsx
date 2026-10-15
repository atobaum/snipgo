import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SyncSettingsSection } from "./SyncSettingsSection";
import { SyncSettings, SyncStatus } from "../types";

const settings: SyncSettings = { enabled: false, intervalMinutes: 10, conflict: "manual" };
const neverRun: SyncStatus = { lastRunAt: "", ok: false, manual: false, summary: "", errorKind: "", message: "", conflictFiles: [], paused: false };

function setup(overrides: Partial<Parameters<typeof SyncSettingsSection>[0]> = {}) {
  const props = { value: settings, onChange: vi.fn(), status: neverRun, onSyncNow: vi.fn(), syncing: false, ...overrides };
  render(<SyncSettingsSection {...props} />);
  return props;
}

describe("SyncSettingsSection", () => {
  it("toggles enabled", async () => {
    const props = setup();
    await userEvent.click(screen.getByLabelText("Automatic sync"));
    expect(props.onChange).toHaveBeenCalledWith({ ...settings, enabled: true });
  });

  it("changes interval and conflict", async () => {
    const props = setup();
    // The input is controlled by a mock, so set the value in one change event.
    fireEvent.change(screen.getByLabelText("Interval (minutes)"), { target: { value: "5" } });
    expect(props.onChange).toHaveBeenLastCalledWith({ ...settings, intervalMinutes: 5 });
    await userEvent.selectOptions(screen.getByLabelText("On conflict"), "remote");
    expect(props.onChange).toHaveBeenLastCalledWith({ ...settings, conflict: "remote" });
  });

  it("runs sync now", async () => {
    const props = setup();
    await userEvent.click(screen.getByRole("button", { name: "Sync now" }));
    expect(props.onSyncNow).toHaveBeenCalled();
  });

  it("disables sync now while syncing", () => {
    setup({ syncing: true });
    expect(screen.getByRole("button", { name: "Syncing..." })).toBeDisabled();
  });

  it("shows never synced", () => {
    setup();
    expect(screen.getByText("Never synced")).toBeInTheDocument();
  });

  it("shows paused conflict files", () => {
    setup({ status: { ...neverRun, lastRunAt: "2026-10-15T21:50:00Z", errorKind: "conflict", message: "Sync conflict in 1 file(s): a.md.", conflictFiles: ["a.md"], paused: true } });
    expect(screen.getByText(/Automatic sync is paused/)).toBeInTheDocument();
    expect(screen.getByText("a.md")).toBeInTheDocument();
  });
});
