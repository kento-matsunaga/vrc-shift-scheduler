import { useState, useEffect, useCallback } from 'react';
import {
  listTutorials,
  createTutorial,
  updateTutorial,
  deleteTutorial,
  type Tutorial,
} from '../lib/api';

export default function Tutorials() {
  const [tutorials, setTutorials] = useState<Tutorial[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  // モーダル関連
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    category: '',
    title: '',
    body: '',
    display_order: 0,
    is_published: true,
  });

  const fetchTutorials = useCallback(async () => {
    try {
      setIsLoading(true);
      const response = await listTutorials({ limit: 100 });
      setTutorials(response.data.tutorials);
      setTotalCount(response.data.total_count);
    } catch (err) {
      setError('チュートリアルの取得に失敗しました');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTutorials();
  }, [fetchTutorials]);

  const handleOpenCreate = () => {
    setEditingId(null);
    setFormData({
      category: '',
      title: '',
      body: '',
      display_order: tutorials.length,
      is_published: true,
    });
    setIsModalOpen(true);
  };

  const handleOpenEdit = (tutorial: Tutorial) => {
    setEditingId(tutorial.id);
    setFormData({
      category: tutorial.category,
      title: tutorial.title,
      body: tutorial.body,
      display_order: tutorial.display_order,
      is_published: tutorial.is_published,
    });
    setIsModalOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        await updateTutorial(editingId, formData);
      } else {
        await createTutorial(formData);
      }
      setIsModalOpen(false);
      await fetchTutorials();
    } catch (err) {
      setError(editingId ? 'チュートリアルの更新に失敗しました' : 'チュートリアルの作成に失敗しました');
      console.error(err);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('このチュートリアルを削除しますか？')) return;

    try {
      await deleteTutorial(id);
      await fetchTutorials();
    } catch (err) {
      setError('チュートリアルの削除に失敗しました');
      console.error(err);
    }
  };

  const handleTogglePublish = async (tutorial: Tutorial) => {
    try {
      await updateTutorial(tutorial.id, {
        is_published: !tutorial.is_published,
      });
      await fetchTutorials();
    } catch (err) {
      setError('公開状態の変更に失敗しました');
      console.error(err);
    }
  };

  // カテゴリでグループ化
  const categories = [...new Set(tutorials.map(t => t.category))].sort();

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">チュートリアル管理</h2>
          <p className="text-sm text-gray-500 mt-1">
            ユーザーに表示する操作ガイドを管理します
          </p>
        </div>
        <button
          onClick={handleOpenCreate}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          新規作成
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4">
          <div className="text-sm text-red-700">{error}</div>
        </div>
      )}

      <div className="text-sm text-gray-500 mb-4">全 {totalCount} 件</div>

      {isLoading ? (
        <div className="text-center py-8">読み込み中...</div>
      ) : tutorials.length === 0 ? (
        <div className="text-center py-8 text-gray-500">チュートリアルがありません</div>
      ) : (
        <div className="space-y-6">
          {categories.map((category) => (
            <div key={category} className="bg-white shadow overflow-hidden rounded-lg">
              <div className="px-6 py-4 bg-gray-50 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">{category}</h3>
              </div>
              <ul className="divide-y divide-gray-200">
                {tutorials
                  .filter((t) => t.category === category)
                  .sort((a, b) => a.display_order - b.display_order)
                  .map((tutorial) => (
                    <li key={tutorial.id} className="px-6 py-4">
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-sm font-medium text-gray-900">
                              {tutorial.title}
                            </span>
                            <span className={`px-2 py-0.5 text-xs font-medium rounded-full ${
                              tutorial.is_published
                                ? 'bg-green-100 text-green-800'
                                : 'bg-gray-100 text-gray-800'
                            }`}>
                              {tutorial.is_published ? '公開中' : '非公開'}
                            </span>
                            <span className="text-xs text-gray-400">
                              順序: {tutorial.display_order}
                            </span>
                          </div>
                          <p className="mt-1 text-sm text-gray-500 line-clamp-2">
                            {tutorial.body}
                          </p>
                        </div>
                        <div className="ml-4 flex-shrink-0 flex items-center space-x-2">
                          <button
                            onClick={() => handleTogglePublish(tutorial)}
                            className={`text-sm ${
                              tutorial.is_published
                                ? 'text-gray-600 hover:text-gray-900'
                                : 'text-green-600 hover:text-green-900'
                            }`}
                          >
                            {tutorial.is_published ? '非公開に' : '公開する'}
                          </button>
                          <button
                            onClick={() => handleOpenEdit(tutorial)}
                            className="text-sm text-indigo-600 hover:text-indigo-900"
                          >
                            編集
                          </button>
                          <button
                            onClick={() => handleDelete(tutorial.id)}
                            className="text-sm text-red-600 hover:text-red-900"
                          >
                            削除
                          </button>
                        </div>
                      </div>
                    </li>
                  ))}
              </ul>
            </div>
          ))}
        </div>
      )}

      {/* モーダル */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div
              className="fixed inset-0 bg-black/50"
              onClick={() => setIsModalOpen(false)}
            />
            <div className="relative bg-white rounded-lg shadow-xl w-full max-w-lg">
              <form onSubmit={handleSubmit}>
                <div className="p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">
                    {editingId ? 'チュートリアルを編集' : 'チュートリアルを作成'}
                  </h3>

                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700">カテゴリ</label>
                      <input
                        type="text"
                        required
                        value={formData.category}
                        onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                        placeholder="例: 基本操作、シフト管理、メンバー管理"
                        list="category-list"
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                      />
                      <datalist id="category-list">
                        {categories.map((cat) => (
                          <option key={cat} value={cat} />
                        ))}
                      </datalist>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">タイトル</label>
                      <input
                        type="text"
                        required
                        value={formData.title}
                        onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">本文</label>
                      <textarea
                        required
                        rows={8}
                        value={formData.body}
                        onChange={(e) => setFormData({ ...formData, body: e.target.value })}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                      />
                    </div>

                    <div className="flex items-center gap-4">
                      <div className="flex-1">
                        <label className="block text-sm font-medium text-gray-700">表示順</label>
                        <input
                          type="number"
                          min="0"
                          value={formData.display_order}
                          onChange={(e) => setFormData({ ...formData, display_order: parseInt(e.target.value) || 0 })}
                          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                        />
                      </div>
                      <div className="flex-1">
                        <label className="flex items-center gap-2">
                          <input
                            type="checkbox"
                            checked={formData.is_published}
                            onChange={(e) => setFormData({ ...formData, is_published: e.target.checked })}
                            className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                          />
                          <span className="text-sm font-medium text-gray-700">公開する</span>
                        </label>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="px-6 py-4 bg-gray-50 flex justify-end space-x-3 rounded-b-lg">
                  <button
                    type="button"
                    onClick={() => setIsModalOpen(false)}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    キャンセル
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md hover:bg-indigo-700"
                  >
                    {editingId ? '更新' : '作成'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
