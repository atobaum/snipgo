import { useState, useRef, useCallback } from 'react';
import { Snippet } from './types';
import { SnippetList } from './components/SnippetList';
import { SnippetEditor } from './components/SnippetEditor';
import { TagList } from './components/TagList';
import { SettingsPage } from './components/SettingsPage';
import { app } from './bridge';

function App() {
  const [selectedSnippet, setSelectedSnippet] = useState<Snippet | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [listRefreshKey, setListRefreshKey] = useState(0);
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [showSettings, setShowSettings] = useState(false);
  const isDirtyRef = useRef(false);

  const handleSelectTag = useCallback((tag: string | null) => {
    setSelectedTag(tag);
  }, []);

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

  const handleSettingsChange = useCallback(() => {
    // 설정 변경 후 목록 새로고침
    setSelectedSnippet(null);
    setListRefreshKey((k) => k + 1);
  }, []);

  const handleCreateNew = async () => {
    try {
      const newSnippet = await app.CreateSnippet('Untitled');
      setSelectedSnippet(newSnippet);
      setListRefreshKey((k) => k + 1);
    } catch (err) {
      console.error('Failed to create snippet:', err);
      alert('Failed to create snippet: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
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
          <div className="flex items-center gap-2">
            <button
              onClick={handleCreateNew}
              className="px-3 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 flex items-center gap-1"
              title="New Snippet"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              New
            </button>
            <button
              onClick={() => setShowSettings(true)}
              className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg"
              title="Settings"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar - Tag List + Snippet List */}
        <aside className="w-80 bg-white border-r border-gray-200 overflow-y-auto">
          <TagList
            selectedTag={selectedTag}
            onSelectTag={handleSelectTag}
            refreshKey={listRefreshKey}
          />
          <SnippetList
            onSelect={handleSelectSnippet}
            searchQuery={searchQuery}
            selectedId={selectedSnippet?.id}
            refreshKey={listRefreshKey}
            selectedTag={selectedTag}
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

      {/* Settings Modal */}
      {showSettings && (
        <SettingsPage
          onClose={() => setShowSettings(false)}
          onSettingsChange={handleSettingsChange}
        />
      )}
    </div>
  );
}

export default App;
