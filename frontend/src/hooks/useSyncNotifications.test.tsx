import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import { Toast } from "./useToasts";
import { renderHook, act } from "@testing-library/react";
import { useSyncNotifications } from "./useSyncNotifications";
import { SyncStatus } from "../types";

let emit: (s: SyncStatus) => void = () => {};
vi.mock("../bridge", () => ({
  onSyncStatus: vi.fn((cb: (s: SyncStatus) => void) => {
    emit = cb;
    return () => {};
  }),
}));

const base: SyncStatus = {
  lastRunAt: "2026-10-15T21:50:00Z",
  ok: true,
  manual: false,
  summary: "Already up to date",
  errorKind: "",
  message: "",
  conflictFiles: [],
  paused: false,
};

describe("useSyncNotifications", () => {
  let push: Mock<(toast: Omit<Toast, "id">) => void>;
  beforeEach(() => {
    push = vi.fn();
    renderHook(() => useSyncNotifications(push));
  });

  it("shows an error toast for a failed sync", () => {
    act(() => emit({ ...base, ok: false, errorKind: "push", message: "git push failed: denied" }));
    expect(push).toHaveBeenCalledWith({ kind: "error", title: "Sync failed", message: "git push failed: denied" });
  });

  it("does not repeat the same automatic error", () => {
    const offline = { ...base, ok: false, errorKind: "pull", message: "git pull failed: offline" };
    act(() => emit(offline));
    act(() => emit(offline));
    expect(push).toHaveBeenCalledTimes(1);
  });

  it("shows the error again after a success in between", () => {
    const offline = { ...base, ok: false, errorKind: "pull", message: "git pull failed: offline" };
    act(() => emit(offline));
    act(() => emit(base));
    act(() => emit(offline));
    expect(push).toHaveBeenCalledTimes(2);
  });

  it("always shows manual errors", () => {
    const offline = { ...base, ok: false, manual: true, errorKind: "pull", message: "git pull failed: offline" };
    act(() => emit(offline));
    act(() => emit(offline));
    expect(push).toHaveBeenCalledTimes(2);
  });

  it("shows a success toast only for manual syncs", () => {
    act(() => emit(base));
    expect(push).not.toHaveBeenCalled();
    act(() => emit({ ...base, manual: true, summary: "Synced: pushed" }));
    expect(push).toHaveBeenCalledWith({ kind: "success", title: "Synced", message: "Synced: pushed" });
  });
});
