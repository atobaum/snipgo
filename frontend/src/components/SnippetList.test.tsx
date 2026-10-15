import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SnippetList } from './SnippetList';
import { Snippet } from '../types';

const mockSnippets: Snippet[] = [
  {
    id: '1',
    title: 'First Snippet',
    description: '',
    tags: ['tag1'],
    language: 'javascript',
    is_favorite: true,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    body: 'first body',
  },
  {
    id: '2',
    title: 'Second Snippet',
    description: '',
    tags: ['tag2', 'tag3'],
    language: 'python',
    is_favorite: false,
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-02T00:00:00Z',
    body: 'second body',
  },
];

// Mock the bridge module
vi.mock('../bridge', () => ({
  app: {
    GetAllSnippets: vi.fn(),
    SearchSnippets: vi.fn(),
    ConfirmDialog: vi.fn(),
    DeleteSnippet: vi.fn(),
    ReloadSnippets: vi.fn(),
    AlertDialog: vi.fn(),
  },
}));

describe('SnippetList', () => {
  const defaultProps = {
    onSelect: vi.fn(),
    onSnippetDeleted: vi.fn(),
    searchQuery: '',
    selectedId: undefined,
    refreshKey: 0,
  };

  beforeEach(async () => {
    vi.clearAllMocks();
    const { app } = await import('../bridge');
    vi.mocked(app.GetAllSnippets).mockResolvedValue(mockSnippets);
    vi.mocked(app.SearchSnippets).mockResolvedValue([mockSnippets[0]]);
    vi.mocked(app.ConfirmDialog).mockResolvedValue(true);
    vi.mocked(app.DeleteSnippet).mockResolvedValue(undefined);
    vi.mocked(app.ReloadSnippets).mockResolvedValue(undefined);
    vi.mocked(app.AlertDialog).mockResolvedValue(undefined);
  });

  describe('렌더링', () => {
    it('snippets 목록을 표시한다', async () => {
      render(<SnippetList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
        expect(screen.getByText('Second Snippet')).toBeInTheDocument();
      });
    });

    it('태그를 표시한다', async () => {
      render(<SnippetList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('tag1')).toBeInTheDocument();
        expect(screen.getByText('tag2')).toBeInTheDocument();
        expect(screen.getByText('tag3')).toBeInTheDocument();
      });
    });

    it('language를 표시한다', async () => {
      render(<SnippetList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('javascript')).toBeInTheDocument();
        expect(screen.getByText('python')).toBeInTheDocument();
      });
    });

    it('favorite 아이템에 별 표시를 한다', async () => {
      render(<SnippetList {...defaultProps} />);

      await waitFor(() => {
        // First snippet은 favorite이므로 별이 있어야 함
        expect(screen.getByText('★')).toBeInTheDocument();
      });
    });
  });

  describe('선택 하이라이트', () => {
    it('선택된 아이템에 하이라이트 스타일이 적용된다', async () => {
      render(<SnippetList {...defaultProps} selectedId="1" />);

      await waitFor(() => {
        const firstItem = screen.getByText('First Snippet').closest('div[class*="cursor-pointer"]');
        expect(firstItem).toHaveClass('bg-blue-50');
        expect(firstItem).toHaveClass('border-l-2');
        expect(firstItem).toHaveClass('border-blue-500');
      });
    });

    it('선택되지 않은 아이템은 hover 스타일만 있다', async () => {
      render(<SnippetList {...defaultProps} selectedId="1" />);

      await waitFor(() => {
        const secondItem = screen.getByText('Second Snippet').closest('div[class*="cursor-pointer"]');
        expect(secondItem).toHaveClass('hover:bg-gray-50');
        expect(secondItem).not.toHaveClass('bg-blue-50');
      });
    });
  });

  describe('아이템 선택', () => {
    it('아이템 클릭 시 onSelect 콜백이 호출된다', async () => {
      const user = userEvent.setup();
      render(<SnippetList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('First Snippet')).toBeInTheDocument();
      });

      await user.click(screen.getByText('First Snippet'));
      expect(defaultProps.onSelect).toHaveBeenCalledWith(mockSnippets[0]);
    });
  });

  describe('refreshKey', () => {
    it('refreshKey 변경 시 목록을 다시 로드한다', async () => {
      const { app } = await import('../bridge');
      const { rerender } = render(<SnippetList {...defaultProps} refreshKey={0} />);

      await waitFor(() => {
        expect(app.GetAllSnippets).toHaveBeenCalledTimes(1);
      });

      // refreshKey 변경
      rerender(<SnippetList {...defaultProps} refreshKey={1} />);

      await waitFor(() => {
        expect(app.GetAllSnippets).toHaveBeenCalledTimes(2);
      });
    });
  });

  describe('검색', () => {
    it('searchQuery가 있으면 SearchSnippets를 호출한다', async () => {
      const { app } = await import('../bridge');
      render(<SnippetList {...defaultProps} searchQuery="docker" />);

      await waitFor(() => {
        expect(app.SearchSnippets).toHaveBeenCalledWith('docker');
      });
    });

    it('searchQuery가 비어있으면 GetAllSnippets를 호출한다', async () => {
      const { app } = await import('../bridge');
      render(<SnippetList {...defaultProps} searchQuery="" />);

      await waitFor(() => {
        expect(app.GetAllSnippets).toHaveBeenCalled();
      });
    });
  });

  describe('로딩 상태', () => {
    it('로딩 중일 때 로딩 메시지를 표시한다', () => {
      render(<SnippetList {...defaultProps} />);
      expect(screen.getByText('Loading snippets...')).toBeInTheDocument();
    });
  });

  describe('빈 목록', () => {
    it('snippets가 없으면 안내 메시지를 표시한다', async () => {
      const { app } = await import('../bridge');
      vi.mocked(app.GetAllSnippets).mockResolvedValueOnce([]);

      render(<SnippetList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('No snippets found.')).toBeInTheDocument();
      });
    });
  });
  describe('컨텍스트 메뉴', () => {
    const openContextMenu = async (title: string) => {
      await waitFor(() => expect(screen.getByText(title)).toBeInTheDocument());
      fireEvent.contextMenu(screen.getByText(title), { clientX: 40, clientY: 60 });
    };

    it('우클릭 시 Delete 항목이 있는 메뉴를 커서 위치에 연다', async () => {
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      const menu = screen.getByRole('menu');
      expect(menu).toHaveStyle({ left: '40px', top: '60px' });
      expect(screen.getByRole('menuitem', { name: 'Delete' })).toHaveClass('text-red-600');
    });

    it('우클릭은 기본 메뉴를 막는다', async () => {
      render(<SnippetList {...defaultProps} />);
      await waitFor(() => expect(screen.getByText('First Snippet')).toBeInTheDocument());

      const notCancelled = fireEvent.contextMenu(screen.getByText('First Snippet'));
      expect(notCancelled).toBe(false);
    });

    it('우클릭은 선택을 바꾸지 않는다', async () => {
      render(<SnippetList {...defaultProps} selectedId="2" />);
      await openContextMenu('First Snippet');

      expect(defaultProps.onSelect).not.toHaveBeenCalled();
      const firstItem = screen.getByText('First Snippet').closest('div[class*="cursor-pointer"]');
      expect(firstItem).not.toHaveClass('bg-blue-50');
    });

    it('Delete 확인 시 삭제 후 onSnippetDeleted를 호출한다', async () => {
      const { app } = await import('../bridge');
      const user = userEvent.setup();
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      await user.click(screen.getByRole('menuitem', { name: 'Delete' }));

      await waitFor(() => {
        expect(app.DeleteSnippet).toHaveBeenCalledWith('1');
        expect(app.ReloadSnippets).toHaveBeenCalled();
        expect(defaultProps.onSnippetDeleted).toHaveBeenCalledWith('1');
      });
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('Delete 취소 시 삭제하지 않는다', async () => {
      const { app } = await import('../bridge');
      vi.mocked(app.ConfirmDialog).mockResolvedValue(false);
      const user = userEvent.setup();
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      await user.click(screen.getByRole('menuitem', { name: 'Delete' }));

      await waitFor(() => expect(app.ConfirmDialog).toHaveBeenCalled());
      expect(app.DeleteSnippet).not.toHaveBeenCalled();
      expect(defaultProps.onSnippetDeleted).not.toHaveBeenCalled();
    });

    it('삭제 실패 시 AlertDialog로 알린다', async () => {
      const { app } = await import('../bridge');
      vi.mocked(app.DeleteSnippet).mockRejectedValue(new Error('boom'));
      const user = userEvent.setup();
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      await user.click(screen.getByRole('menuitem', { name: 'Delete' }));

      await waitFor(() => {
        expect(app.AlertDialog).toHaveBeenCalledWith('SnipGo', 'Failed to delete snippet: boom');
      });
      expect(defaultProps.onSnippetDeleted).not.toHaveBeenCalled();
    });

    it('Escape를 누르면 닫힌다', async () => {
      const user = userEvent.setup();
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      await user.keyboard('{Escape}');
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('바깥을 클릭하면 닫힌다', async () => {
      const user = userEvent.setup();
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      await user.click(document.body);
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('목록을 스크롤하면 닫힌다', async () => {
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      fireEvent.scroll(screen.getByTestId('snippet-list-scroll'));
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('창이 포커스를 잃으면 닫힌다', async () => {
      render(<SnippetList {...defaultProps} />);
      await openContextMenu('First Snippet');

      fireEvent.blur(window);
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });
  });
});

