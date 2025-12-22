import { useState } from 'react';
import { Snippet } from './types';
import { SnippetList } from './components/SnippetList';
import { SnippetEditor } from './components/SnippetEditor';

function App() {
  const [selectedSnippet, setSelectedSnippet] = useState<Snippet | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const handleSelectSnippet = (snippet: Snippet) => {
    setSelectedSnippet(snippet);
  };

  const handleSave = () => {
    // Reload will be handled by components
    setSelectedSnippet(null);
  };

  const handleDelete = () => {
    setSelectedSnippet(null);
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
          <SnippetList onSelect={handleSelectSnippet} searchQuery={searchQuery} />
        </aside>

        {/* Main - Editor */}
        <main className="flex-1 overflow-hidden">
          <SnippetEditor
            snippet={selectedSnippet}
            onSave={handleSave}
            onDelete={handleDelete}
          />
        </main>
      </div>
    </div>
  );
}

export default App;
