import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';
import { Snippet } from './types';

const mockSnippets: Snippet[] = [
  {
    id: '1',
    title: 'First Snippet',
    tags: ['tag1'],
    language: 'javascript',
    is_favorite: false,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    body: 'first body',
  },
  {
    id: '2',
    title: 'Second Snippet',
    tags: ['tag2'],
    language: 'python',
    is_favorite: false,
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-02T00:00:00Z',
    body: 'second body',
  },
];

// Mock the bridge module
vi.mock('./bridge', () => ({
  app: {
    GetAllSnippets: vi.fn(),
    GetSnippet: vi.fn(),
    SearchSnippets: vi.fn(),
    SaveSnippet: vi.fn(),
    DeleteSnippet: vi.fn(),
    ReloadSnippets: vi.fn(),
    CopyToClipboard: vi.fn(),
  },
}));

// Mock CodeMirror
vi.mock('@uiw/react-codemirror', () => ({
  default: ({ value, onChange }: { value: string; onChange: (v: string) => void }) => (
    <textarea
      data-testid="codemirror-mock"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

describe('App', () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    // Reset confirm mock
    vi.spyOn(window, 'confirm').mockImplementation(() => true);
    
    // Setup mocks
    const { app } = await import('./bridge');
    vi.mocked(app.GetAllSnippets).mockResolvedValue(mockSnippets);
    vi.mocked(app.GetSnippet).mockImplementation((id: string) => {
      const snippet = mockSnippets.find(s => s.id === id);
      return Promise.resolve(snippet!);
    });
    vi.mocked(app.SearchSnippets).mockResolvedValue([mockSnippets[0]]);
    vi.mocked(app.SaveSnippet).mockResolvedValue(undefined);
    vi.mocked(app.DeleteSnippet).mockResolvedValue(undefined);
    vi.mocked(app.ReloadSnippets).mockResolvedValue(undefined);
    vi.mocked(app.CopyToClipboard).mockResolvedValue(undefined);
  });

  describe('기본 렌더링', () => {
    it('헤더와 검색창을 표시한다', async () => {
      render(<App />);

      expect(screen.getByText('SnipGo')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Search snippets...')).toBeInTheDocument();
    });

    it('snippet 목록을 로드하고 표시한다', async () => {
      render(<App />);

      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
        expect(screen.getByText('Second Snippet')).toBeInTheDocument();
      });
    });

    it('snippet 미선택 시 선택 안내 메시지를 표시한다', async () => {
      render(<App />);

      await waitFor(() => {
        expect(screen.getByText('Select a snippet to edit')).toBeInTheDocument();
      });
    });
  });

  describe('snippet 선택', () => {
    it('아이템 클릭 시 GetSnippet을 호출하여 파일에서 최신 데이터를 가져온다', async () => {
      const { app } = await import('./bridge');
      const user = userEvent.setup();
      render(<App />);

      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
      });

      await user.click(screen.getByText('First Snippet'));

      await waitFor(() => {
        expect(app.GetSnippet).toHaveBeenCalledWith('1');
      });
    });

    it('snippet 선택 후 에디터에 내용이 표시된다', async () => {
      const user = userEvent.setup();
      render(<App />);

      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
      });

      await user.click(screen.getByText('First Snippet'));

      await waitFor(() => {
        // 에디터에 title이 표시됨
        expect(screen.getByDisplayValue('First Snippet')).toBeInTheDocument();
      });
    });
  });

  describe('저장 확인 다이얼로그', () => {
    it('dirty 상태에서 다른 아이템 선택 시 confirm 다이얼로그를 표시한다', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
      const user = userEvent.setup();
      render(<App />);

      // 첫 번째 아이템 선택
      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
      });
      await user.click(screen.getByText('First Snippet'));

      // 에디터가 로드될 때까지 대기
      await waitFor(() => {
        expect(screen.getByDisplayValue('First Snippet')).toBeInTheDocument();
      });

      // title 변경 (dirty 상태로 만듦)
      const titleInput = screen.getByDisplayValue('First Snippet');
      await user.clear(titleInput);
      await user.type(titleInput, 'Modified Title');

      // 두 번째 아이템 선택 시도
      await user.click(screen.getByText('Second Snippet'));

      // confirm이 호출되어야 함
      expect(confirmSpy).toHaveBeenCalledWith(
        '저장하지 않은 변경사항이 있습니다. 저장하지 않고 이동하시겠습니까?'
      );
    });

    it('confirm에서 취소하면 선택이 변경되지 않는다', async () => {
      vi.spyOn(window, 'confirm').mockReturnValue(false);
      const user = userEvent.setup();
      render(<App />);

      // 첫 번째 아이템 선택
      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
      });
      await user.click(screen.getByText('First Snippet'));

      await waitFor(() => {
        expect(screen.getByDisplayValue('First Snippet')).toBeInTheDocument();
      });

      // title 변경
      const titleInput = screen.getByDisplayValue('First Snippet');
      await user.clear(titleInput);
      await user.type(titleInput, 'Modified');

      // 두 번째 아이템 선택 시도 (취소)
      await user.click(screen.getByText('Second Snippet'));

      // 여전히 첫 번째 아이템의 modified title이 표시되어야 함
      await waitFor(() => {
        expect(screen.getByDisplayValue('Modified')).toBeInTheDocument();
      });
    });
  });

  describe('검색', () => {
    it('검색어 입력 시 SearchSnippets를 호출한다', async () => {
      const { app } = await import('./bridge');
      const user = userEvent.setup();
      render(<App />);

      const searchInput = screen.getByPlaceholderText('Search snippets...');
      await user.type(searchInput, 'docker');

      await waitFor(() => {
        expect(app.SearchSnippets).toHaveBeenCalledWith('docker');
      });
    });
  });
});

