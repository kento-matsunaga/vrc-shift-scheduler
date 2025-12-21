import { useState, useEffect } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { getTemplate, deleteTemplate } from '../lib/api/templateApi';
import type { Template } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function TemplateDetail() {
  const { eventId, templateId } = useParams<{ eventId: string; templateId: string }>();
  const navigate = useNavigate();
  const [template, setTemplate] = useState<Template | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (eventId && templateId) {
      loadTemplate();
    }
  }, [eventId, templateId]);

  const loadTemplate = async () => {
    if (!eventId || !templateId) return;

    try {
      setLoading(true);
      const data = await getTemplate(eventId, templateId);
      setTemplate(data);
    } catch (err) {
      console.error('Failed to load template:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('テンプレートの読み込みに失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!eventId || !templateId || !template) return;

    if (!confirm(`テンプレート「${template.template_name}」を削除しますか？\nこの操作は取り消せません。`)) {
      return;
    }

    try {
      await deleteTemplate(eventId, templateId);
      navigate(`/events/${eventId}/templates`);
    } catch (err) {
      console.error('Failed to delete template:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('テンプレートの削除に失敗しました');
      }
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  if (error || !template) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-sm text-red-800">{error || 'テンプレートが見つかりません'}</p>
        </div>
        <div className="mt-4">
          <Link to={`/events/${eventId}/templates`} className="text-indigo-600 hover:text-indigo-800">
            ← テンプレート一覧に戻る
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to={`/events/${eventId}/templates`} className="hover:text-gray-900">
          テンプレート一覧
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">{template.template_name}</span>
      </nav>

      {/* ヘッダー */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex justify-between items-start mb-4">
          <div className="flex-1">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">{template.template_name}</h2>
            {template.description && (
              <p className="text-gray-600">{template.description}</p>
            )}
          </div>
          <div className="flex gap-2 ml-4">
            <Link
              to={`/events/${eventId}/templates/${templateId}/edit`}
              className="bg-indigo-100 hover:bg-indigo-200 text-indigo-700 px-4 py-2 rounded-lg text-sm font-medium"
            >
              編集
            </Link>
            <button
              onClick={handleDelete}
              className="bg-red-100 hover:bg-red-200 text-red-700 px-4 py-2 rounded-lg text-sm font-medium"
            >
              削除
            </button>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 pt-4 border-t border-gray-200">
          <div>
            <p className="text-sm text-gray-600">シフト枠数</p>
            <p className="text-lg font-semibold text-gray-900">{(template.items || []).length} 枠</p>
          </div>
          <div>
            <p className="text-sm text-gray-600">作成日</p>
            <p className="text-lg font-semibold text-gray-900">
              {new Date(template.created_at).toLocaleDateString('ja-JP')}
            </p>
          </div>
        </div>
      </div>

      {/* シフト枠一覧 */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">シフト枠一覧</h3>

        {(template.items || []).length === 0 ? (
          <p className="text-gray-600 text-center py-8">シフト枠がありません</p>
        ) : (
          <div className="space-y-4">
            {(template.items || []).map((item, index) => (
              <div key={index} className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h4 className="font-semibold text-gray-900 text-lg">
                      {item.slot_name} ({item.instance_name})
                    </h4>
                    <p className="text-sm text-gray-600 mt-1">
                      優先度: {item.priority}
                    </p>
                  </div>
                  <span className="bg-indigo-100 text-indigo-800 text-xs font-medium px-2.5 py-0.5 rounded">
                    {item.required_count}名
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <p className="text-gray-600">開始時刻</p>
                    <p className="font-medium text-gray-900">{item.start_time.substring(0, 5)}</p>
                  </div>
                  <div>
                    <p className="text-gray-600">終了時刻</p>
                    <p className="font-medium text-gray-900">{item.end_time.substring(0, 5)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 使用方法の説明 */}
      <div className="bg-indigo-50 border border-indigo-200 rounded-lg p-4 mt-6">
        <h4 className="font-semibold text-indigo-900 mb-2">💡 テンプレートの使い方</h4>
        <p className="text-sm text-indigo-800">
          このテンプレートは営業日作成時に選択することで、登録されているシフト枠を自動的に作成します。
          営業日一覧ページから「営業日を追加」を選択し、テンプレートを選んでください。
        </p>
      </div>
    </div>
  );
}
