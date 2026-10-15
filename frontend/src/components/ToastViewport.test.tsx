import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, act, renderHook } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ToastViewport } from "./ToastViewport";
import { useToasts } from "../hooks/useToasts";

describe("ToastViewport", () => {
  it("renders toasts and dismisses on click", async () => {
    const onDismiss = vi.fn();
    render(
      <ToastViewport
        toasts={[{ id: 1, kind: "error", title: "Sync failed", message: "git push failed" }]}
        onDismiss={onDismiss}
      />
    );
    expect(screen.getByRole("alert")).toHaveTextContent("Sync failed");
    expect(screen.getByRole("alert")).toHaveTextContent("git push failed");
    await userEvent.click(screen.getByRole("button", { name: "Dismiss" }));
    expect(onDismiss).toHaveBeenCalledWith(1);
  });
});

describe("useToasts", () => {
  afterEach(() => vi.useRealTimers());

  it("auto-dismisses success toasts but keeps errors", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useToasts());
    act(() => {
      result.current.push({ kind: "success", title: "Synced", message: "ok" });
      result.current.push({ kind: "error", title: "Sync failed", message: "bad" });
    });
    expect(result.current.toasts).toHaveLength(2);
    act(() => vi.advanceTimersByTime(4000));
    expect(result.current.toasts.map((t) => t.kind)).toEqual(["error"]);
  });
});
