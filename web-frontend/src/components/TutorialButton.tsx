import { useState, useEffect } from 'react';
import { getTutorials, type Tutorial } from '../lib/api/tutorialApi';

export function TutorialButton() {
  const [isOpen, setIsOpen] = useState(false);
  const [tutorials, setTutorials] = useState<Tutorial[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && tutorials.length === 0) {
      fetchTutorials();
    }
  }, [isOpen]);

  // ESCキーで閉じる
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }
    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
      document.body.style.overflow = 'hidden';
    }
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  async function fetchTutorials() {
    setLoading(true);
    try {
      const data = await getTutorials();
      setTutorials(data);
      if (data.length > 0) {
        const categories = [...new Set(data.map(t => t.category))];
        setSelectedCategory(categories[0]);
      }
    } catch (error) {
      console.error('Failed to fetch tutorials:', error);
    } finally {
      setLoading(false);
    }
  }

  const categories = [...new Set(tutorials.map(t => t.category))];
  const filteredTutorials = tutorials.filter(t => t.category === selectedCategory);

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-full transition-colors"
        aria-label="ヘルプ"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>

      {isOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            {/* Overlay */}
            <div
              className="fixed inset-0 bg-black/50 transition-opacity"
              onClick={() => setIsOpen(false)}
            />

            {/* Modal */}
            <div className="relative bg-white rounded-lg shadow-xl w-full max-w-3xl max-h-[80vh] overflow-hidden">
              {/* Header */}
              <div className="flex items-center justify-between p-4 border-b border-gray-200">
                <h2 className="text-xl font-semibold text-gray-900">操作ガイド</h2>
                <button
                  onClick={() => setIsOpen(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              {/* Content */}
              <div className="flex h-[60vh]">
                {/* Sidebar */}
                <div className="w-48 border-r border-gray-200 bg-gray-50 overflow-y-auto">
                  {loading ? (
                    <div className="p-4 text-gray-500">読み込み中...</div>
                  ) : (
                    <nav className="p-2">
                      {categories.map(category => (
                        <button
                          key={category}
                          onClick={() => setSelectedCategory(category)}
                          className={`w-full text-left px-3 py-2 rounded-md text-sm ${
                            selectedCategory === category
                              ? 'bg-blue-100 text-blue-700 font-medium'
                              : 'text-gray-700 hover:bg-gray-100'
                          }`}
                        >
                          {category}
                        </button>
                      ))}
                    </nav>
                  )}
                </div>

                {/* Main content */}
                <div className="flex-1 overflow-y-auto p-4">
                  {loading ? (
                    <div className="text-gray-500">読み込み中...</div>
                  ) : filteredTutorials.length === 0 ? (
                    <div className="text-gray-500">チュートリアルがありません</div>
                  ) : (
                    <div className="space-y-6">
                      {filteredTutorials.map(tutorial => (
                        <div key={tutorial.id}>
                          <h3 className="text-lg font-medium text-gray-900 mb-2">{tutorial.title}</h3>
                          <div className="prose prose-sm max-w-none text-gray-600 whitespace-pre-wrap">
                            {tutorial.body}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
