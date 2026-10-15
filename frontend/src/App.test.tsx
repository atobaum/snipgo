import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, act, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';
import { Snippet } from './types';

const mockSnippets: Snippet[] = [
  {
    id: '1',
    title: 'First Snippet',
    description: '',
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
    description: '',
    tags: ['tag2'],
    language: 'python',
    is_favorite: false,
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-02T00:00:00Z',
    body: 'second body',
  },
];

let emitSnippetsChanged: () => void | Promise<void> = () => {};

// Mock the bridge module
vi.mock('./bridge', () => ({
  onSnippetsChanged: vi.fn((callback: () => void | Promise<void>) => {
    emitSnippetsChanged = callback;
    return () => {};
  }),
  app: {
    GetAllSnippets: vi.fn(),
    GetSnippet: vi.fn(),
    SearchSnippets: vi.fn(),
    SaveSnippet: vi.fn(),
    DeleteSnippet: vi.fn(),
    ReloadSnippets: vi.fn(),
    CopyToClipboard: vi.fn(),
    GetAllTags: vi.fn().mockResolvedValue([]),
    ConfirmDialog: vi.fn(),
    AlertDialog: vi.fn().mockResolvedValue(undefined),
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
    // Setup mocks
    const { app } = await import('./bridge');
    vi.mocked(app.ConfirmDialog).mockResolvedValue(true);
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
      const { app } = await import('./bridge');
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
      await waitFor(() => {
        expect(app.ConfirmDialog).toHaveBeenCalledWith(
          'Unsaved Changes',
          '저장하지 않은 변경사항이 있습니다. 저장하지 않고 이동하시겠습니까?'
        );
        expect(screen.getByDisplayValue('Second Snippet')).toBeInTheDocument();
      });
    });

    it('confirm에서 취소하면 선택이 변경되지 않는다', async () => {
      const { app } = await import('./bridge');
      vi.mocked(app.ConfirmDialog).mockResolvedValue(false);
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
      await waitFor(() => expect(app.ConfirmDialog).toHaveBeenCalled());
      expect(screen.getByDisplayValue('Modified')).toBeInTheDocument();
      expect(app.GetSnippet).not.toHaveBeenCalledWith('2');
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

  describe('외부 변경 (Hot Reload)', () => {
    const selectFirstSnippet = async () => {
      const user = userEvent.setup();
      render(<App />);
      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
      });
      await user.click(screen.getByText('First Snippet'));
      await waitFor(() => {
        expect(screen.getByDisplayValue('First Snippet')).toBeInTheDocument();
      });
      return user;
    };

    it('snippets:changed 이벤트를 받으면 목록을 다시 불러온다', async () => {
      const { app } = await import('./bridge');
      render(<App />);
      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
      });

      const added: Snippet = { ...mockSnippets[1], id: '3', title: 'Third Snippet' };
      vi.mocked(app.GetAllSnippets).mockResolvedValue([...mockSnippets, added]);
      await act(async () => {
        await emitSnippetsChanged();
      });

      await waitFor(() => {
        expect(screen.getByText('Third Snippet')).toBeInTheDocument();
      });
    });

    it('수정 중이 아닌 열린 snippet은 디스크 내용으로 갱신한다', async () => {
      const { app } = await import('./bridge');
      await selectFirstSnippet();

      const changed: Snippet = { ...mockSnippets[0], title: 'Changed On Disk' };
      vi.mocked(app.GetAllSnippets).mockResolvedValue([changed, mockSnippets[1]]);
      await act(async () => {
        await emitSnippetsChanged();
      });

      await waitFor(() => {
        expect(screen.getByDisplayValue('Changed On Disk')).toBeInTheDocument();
      });
      expect(screen.queryByText('Changed on disk')).not.toBeInTheDocument();
    });

    it('수정 중인 snippet은 덮어쓰지 않고 변경 안내를 표시한다', async () => {
      const { app } = await import('./bridge');
      const user = await selectFirstSnippet();

      const titleInput = screen.getByDisplayValue('First Snippet');
      await user.clear(titleInput);
      await user.type(titleInput, 'Local Edit');

      const changed: Snippet = { ...mockSnippets[0], title: 'Changed On Disk' };
      vi.mocked(app.GetAllSnippets).mockResolvedValue([changed, mockSnippets[1]]);
      await act(async () => {
        await emitSnippetsChanged();
      });

      await waitFor(() => {
        expect(screen.getByText('Changed on disk')).toBeInTheDocument();
      });
      expect(screen.getByDisplayValue('Local Edit')).toBeInTheDocument();
      expect(screen.queryByDisplayValue('Changed On Disk')).not.toBeInTheDocument();
    });

    it('열린 snippet이 외부에서 삭제되면 선택을 해제한다', async () => {
      const { app } = await import('./bridge');
      await selectFirstSnippet();

      vi.mocked(app.GetAllSnippets).mockResolvedValue([mockSnippets[1]]);
      await act(async () => {
        await emitSnippetsChanged();
      });

      await waitFor(() => {
        expect(screen.getByText('Select a snippet to edit')).toBeInTheDocument();
      });
    });

    const deleteWhileEditing = async () => {
      const { app } = await import('./bridge');
      const user = await selectFirstSnippet();

      const titleInput = screen.getByDisplayValue('First Snippet');
      await user.clear(titleInput);
      await user.type(titleInput, 'Local Edit');

      vi.mocked(app.GetAllSnippets).mockResolvedValue([mockSnippets[1]]);
      await act(async () => {
        await emitSnippetsChanged();
      });
      return user;
    };

    it('수정 중인 snippet이 외부에서 삭제되면 편집 내용을 유지하고 삭제 안내를 표시한다', async () => {
      await deleteWhileEditing();

      await waitFor(() => {
        expect(screen.getByText('Deleted on disk')).toBeInTheDocument();
      });
      expect(screen.getByDisplayValue('Local Edit')).toBeInTheDocument();
      expect(screen.queryByText('Select a snippet to edit')).not.toBeInTheDocument();
    });

    it('삭제 안내가 뜬 snippet을 저장하면 다시 기록하고 안내를 지운다', async () => {
      const { app } = await import('./bridge');
      const user = await deleteWhileEditing();
      await waitFor(() => {
        expect(screen.getByText('Deleted on disk')).toBeInTheDocument();
      });

      await user.keyboard('{Meta>}s{/Meta}');

      await waitFor(() => {
        expect(app.SaveSnippet).toHaveBeenCalledWith(
          expect.objectContaining({ id: '1', title: 'Local Edit' })
        );
        expect(screen.queryByText('Deleted on disk')).not.toBeInTheDocument();
      });
      expect(screen.getByDisplayValue('Local Edit')).toBeInTheDocument();
    });
  });
  describe('컨텍스트 메뉴 삭제', () => {
    it('열려 있는 snippet을 삭제하면 에디터를 비우고 목록을 갱신한다', async () => {
      const { app } = await import('./bridge');
      const user = userEvent.setup();
      render(<App />);

      await waitFor(() => expect(screen.getByText('First Snippet')).toBeInTheDocument());
      await user.click(screen.getByText('First Snippet'));
      await waitFor(() => expect(screen.getByDisplayValue('First Snippet')).toBeInTheDocument());
      const loadCount = vi.mocked(app.GetAllSnippets).mock.calls.length;

      fireEvent.contextMenu(screen.getByRole('heading', { name: 'First Snippet' }));
      await user.click(screen.getByRole('menuitem', { name: 'Delete' }));

      await waitFor(() => {
        expect(app.DeleteSnippet).toHaveBeenCalledWith('1');
        expect(screen.getByText('Select a snippet to edit')).toBeInTheDocument();
      });
      await waitFor(() => {
        expect(vi.mocked(app.GetAllSnippets).mock.calls.length).toBeGreaterThan(loadCount);
      });
    });

    it('다른 snippet을 삭제하면 열린 snippet은 유지한다', async () => {
      const { app } = await import('./bridge');
      const user = userEvent.setup();
      render(<App />);

      await waitFor(() => expect(screen.getByText('First Snippet')).toBeInTheDocument());
      await user.click(screen.getByText('First Snippet'));
      await waitFor(() => expect(screen.getByDisplayValue('First Snippet')).toBeInTheDocument());

      fireEvent.contextMenu(screen.getByText('Second Snippet'));
      await user.click(screen.getByRole('menuitem', { name: 'Delete' }));

      await waitFor(() => expect(app.DeleteSnippet).toHaveBeenCalledWith('2'));
      expect(screen.getByDisplayValue('First Snippet')).toBeInTheDocument();
    });
  });
});
