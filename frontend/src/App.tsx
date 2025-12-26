import { useState, useRef, useCallback } from 'react';
import { Snippet } from './types';
import { SnippetList } from './components/SnippetList';
import { SnippetEditor } from './components/SnippetEditor';
import { app } from './bridge';

function App() {
  const [selectedSnippet, setSelectedSnippet] = useState<Snippet | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [listRefreshKey, setListRefreshKey] = useState(0);
  const isDirtyRef = useRef(false);

  const handleDirtyChange = useCallback((dirty: boolean) => {
    isDirtyRef.current = dirty;
  }, []);

  const handleSelectSnippet = async (snippet: Snippet) => {
    if (isDirtyRef.current) {
      const result = confirm('저장하지 않은 변경사항이 있습니다. 저장하지 않고 이동하시겠습니까?');
      if (!result) {
        return; // 사용자가 취소함
      }
    }
    // 파일에서 최신 데이터 읽어오기
    try {
      const freshSnippet = await app.GetSnippet(snippet.id);
      setSelectedSnippet(freshSnippet);
    } catch (err) {
      console.error('Failed to load snippet:', err);
      setSelectedSnippet(snippet); // fallback
    }
  };

  const handleSave = (updatedSnippet: Snippet) => {
    // 저장 후 선택 유지 (업데이트된 snippet으로 갱신)
    setSelectedSnippet(updatedSnippet);
    setListRefreshKey((k) => k + 1); // 목록 갱신
  };

  const handleListRefresh = useCallback(() => {
    setListRefreshKey((k) => k + 1);
  }, []);

  const handleDelete = () => {
    setSelectedSnippet(null);
    setListRefreshKey((k) => k + 1); // 목록 갱신
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-800">SnipGo</h1>
          <div className="flex-1 max-w-md ml-8">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search snippets..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar - Snippet List */}
        <aside className="w-80 bg-white border-r border-gray-200 overflow-y-auto">
          <SnippetList
            onSelect={handleSelectSnippet}
            searchQuery={searchQuery}
            selectedId={selectedSnippet?.id}
            refreshKey={listRefreshKey}
          />
        </aside>

        {/* Main - Editor */}
        <main className="flex-1 overflow-hidden">
          <SnippetEditor
            snippet={selectedSnippet}
            onSave={handleSave}
            onDelete={handleDelete}
            onDirtyChange={handleDirtyChange}
            onListRefresh={handleListRefresh}
          />
        </main>
      </div>
    </div>
  );
}

export default App;
