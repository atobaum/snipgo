import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SnippetEditor } from "./SnippetEditor";
import { Snippet } from "../types";

// Mock the bridge module
vi.mock("../bridge", () => ({
  app: {
    SaveSnippet: vi.fn().mockResolvedValue(undefined),
    DeleteSnippet: vi.fn().mockResolvedValue(undefined),
    ReloadSnippets: vi.fn().mockResolvedValue(undefined),
    CopyToClipboard: vi.fn().mockResolvedValue(undefined),
  },
}));

// Mock CodeMirror (it doesn't work well in jsdom)
vi.mock("@uiw/react-codemirror", () => ({
  default: ({
    value,
    onChange,
  }: {
    value: string;
    onChange: (v: string) => void;
  }) => (
    <textarea
      data-testid="codemirror-mock"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

const mockSnippet: Snippet = {
  id: "test-id",
  title: "Test Snippet",
  tags: ["tag1", "tag2"],
  language: "javascript",
  is_favorite: false,
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
  body: 'console.log("hello");',
};

describe("SnippetEditor", () => {
  const defaultProps = {
    snippet: mockSnippet,
    onSave: vi.fn(),
    onDelete: vi.fn(),
    onDirtyChange: vi.fn(),
    onListRefresh: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("렌더링", () => {
    it("snippet이 없으면 선택 안내 메시지를 표시한다", () => {
      render(<SnippetEditor {...defaultProps} snippet={null} />);
      expect(screen.getByText("Select a snippet to edit")).toBeInTheDocument();
    });

    it("snippet이 있으면 제목, 태그, 언어를 표시한다", () => {
      render(<SnippetEditor {...defaultProps} />);

      expect(screen.getByDisplayValue("Test Snippet")).toBeInTheDocument();
      expect(screen.getByText("tag1")).toBeInTheDocument();
      expect(screen.getByText("tag2")).toBeInTheDocument();
      expect(screen.getByDisplayValue("javascript")).toBeInTheDocument();
    });

    it("favorite 상태를 올바르게 표시한다", () => {
      render(<SnippetEditor {...defaultProps} />);
      expect(screen.getByText("☆ Favorite")).toBeInTheDocument();

      render(
        <SnippetEditor
          {...defaultProps}
          snippet={{ ...mockSnippet, is_favorite: true }}
        />
      );
      expect(screen.getByText("★ Favorite")).toBeInTheDocument();
    });
  });

  describe("isDirty 상태", () => {
    it("title 변경 시 수정됨 표시가 나타난다", async () => {
      const user = userEvent.setup();
      render(<SnippetEditor {...defaultProps} />);

      const titleInput = screen.getByDisplayValue("Test Snippet");
      await user.clear(titleInput);
      await user.type(titleInput, "New Title");

      expect(screen.getByText("수정됨")).toBeInTheDocument();
      expect(defaultProps.onDirtyChange).toHaveBeenCalledWith(true);
    });

    it("body 변경 시 수정됨 표시가 나타난다", async () => {
      render(<SnippetEditor {...defaultProps} />);

      const bodyInput = screen.getByTestId("codemirror-mock");
      fireEvent.change(bodyInput, { target: { value: "new body content" } });

      expect(screen.getByText("수정됨")).toBeInTheDocument();
      expect(defaultProps.onDirtyChange).toHaveBeenCalledWith(true);
    });

    it("language 변경 시 수정됨 표시가 나타난다", async () => {
      const user = userEvent.setup();
      render(<SnippetEditor {...defaultProps} />);

      const languageInput = screen.getByDisplayValue("javascript");
      await user.clear(languageInput);
      await user.type(languageInput, "python");

      expect(screen.getByText("수정됨")).toBeInTheDocument();
    });

    it("tag 변경은 isDirty에 영향을 주지 않는다 (즉시 저장되므로)", async () => {
      const user = userEvent.setup();
      render(<SnippetEditor {...defaultProps} />);

      const tagInput = screen.getByPlaceholderText("Add tag...");
      await user.type(tagInput, "newtag{enter}");

      // 수정됨 표시가 나타나지 않아야 함
      expect(screen.queryByText("수정됨")).not.toBeInTheDocument();
    });
  });

  describe("tagInput 초기화", () => {
    it("snippet 변경 시 tagInput이 초기화된다", async () => {
      const user = userEvent.setup();
      const { rerender } = render(<SnippetEditor {...defaultProps} />);

      // 태그 입력 중
      const tagInput = screen.getByPlaceholderText("Add tag...");
      await user.type(tagInput, "typing...");
      expect(tagInput).toHaveValue("typing...");

      // 다른 snippet 선택
      const newSnippet = { ...mockSnippet, id: "new-id", title: "New Snippet" };
      rerender(<SnippetEditor {...defaultProps} snippet={newSnippet} />);

      // tagInput이 초기화되어야 함
      expect(screen.getByPlaceholderText("Add tag...")).toHaveValue("");
    });
  });

  describe("태그 즉시 저장", () => {
    it("태그 추가 시 즉시 저장된다", async () => {
      const { app } = await import("../bridge");
      const user = userEvent.setup();
      render(<SnippetEditor {...defaultProps} />);

      const tagInput = screen.getByPlaceholderText("Add tag...");
      await user.type(tagInput, "newtag{enter}");

      await waitFor(() => {
        expect(app.SaveSnippet).toHaveBeenCalled();
        expect(app.ReloadSnippets).toHaveBeenCalled();
        expect(defaultProps.onListRefresh).toHaveBeenCalled();
      });
    });

    it("태그 삭제 시 즉시 저장된다", async () => {
      const { app } = await import("../bridge");
      const user = userEvent.setup();
      render(<SnippetEditor {...defaultProps} />);

      // tag1 삭제 버튼 클릭
      const removeButtons = screen.getAllByText("×");
      await user.click(removeButtons[0]);

      await waitFor(() => {
        expect(app.SaveSnippet).toHaveBeenCalled();
        expect(defaultProps.onListRefresh).toHaveBeenCalled();
      });
    });
  });

  describe("favorite 즉시 저장", () => {
    it("favorite 토글 시 즉시 저장된다", async () => {
      const { app } = await import("../bridge");
      const user = userEvent.setup();
      render(<SnippetEditor {...defaultProps} />);

      const favoriteButton = screen.getByText("☆ Favorite");
      await user.click(favoriteButton);

      await waitFor(() => {
        expect(app.SaveSnippet).toHaveBeenCalled();
        expect(defaultProps.onListRefresh).toHaveBeenCalled();
      });

      // UI도 변경되어야 함
      expect(screen.getByText("★ Favorite")).toBeInTheDocument();
    });
  });
});
