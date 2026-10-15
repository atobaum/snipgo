import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ContextMenu } from "./ContextMenu";

describe("ContextMenu", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("뷰포트를 넘지 않도록 위치를 조정한다", () => {
    vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockReturnValue({
      width: 120,
      height: 40,
    } as DOMRect);
    const right = window.innerWidth - 10;
    const bottom = window.innerHeight - 10;

    render(
      <ContextMenu
        x={right}
        y={bottom}
        items={[{ label: "Delete", onSelect: vi.fn() }]}
        onClose={vi.fn()}
      />
    );

    const menu = screen.getByRole("menu");
    expect(menu).toHaveStyle({
      left: `${window.innerWidth - 120}px`,
      top: `${window.innerHeight - 40}px`,
    });
  });

  it("항목을 고르면 메뉴를 닫고 항목을 실행한다", async () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <ContextMenu x={0} y={0} items={[{ label: "Delete", onSelect }]} onClose={onClose} />
    );

    await user.click(screen.getByRole("menuitem", { name: "Delete" }));

    expect(onClose).toHaveBeenCalled();
    expect(onSelect).toHaveBeenCalled();
  });
});
