import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { listTemplates, deleteTemplate } from '../lib/api/templateApi';
import type { Template } from '../types/api';

const TemplateList = () => {
  const { eventId } = useParams<{ eventId: string }>();
  const navigate = useNavigate();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchTemplates();
  }, [eventId]);

  const fetchTemplates = async () => {
    if (!eventId) {
      setError('イベントIDが指定されていません');
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      const data = await listTemplates(eventId);
      setTemplates(data);
      setError(null);
    } catch (err: any) {
      console.error('Failed to fetch templates:', err);
      setError(err.response?.data?.error?.message || 'テンプレートの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (templateId: string, templateName: string) => {
    if (!eventId) return;

    if (!confirm(`テンプレート「${templateName}」を削除しますか？\nこの操作は取り消せません。`)) {
      return;
    }

    try {
      await deleteTemplate(eventId, templateId);
      alert('テンプレートを削除しました');
      fetchTemplates();
    } catch (err: any) {
      console.error('Failed to delete template:', err);
      alert(err.response?.data?.error?.message || 'テンプレートの削除に失敗しました');
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-600">読み込み中...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">シフトテンプレート管理</h1>
          <p className="text-sm text-gray-600 mt-1">
            営業日作成時に使用できるシフト枠のテンプレートを管理します
          </p>
        </div>
        <button
          onClick={() => navigate(`/events/${eventId}/templates/new`)}
          className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center"
        >
          <svg
            className="w-5 h-5 mr-2"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          新規テンプレート作成
        </button>
      </div>

      {templates.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">テンプレートがありません</h3>
          <p className="mt-1 text-sm text-gray-500">
            新しいテンプレートを作成して、営業日作成を効率化しましょう
          </p>
          <div className="mt-6">
            <button
              onClick={() => navigate(`/events/${eventId}/templates/new`)}
              className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg"
            >
              テンプレートを作成
            </button>
          </div>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {templates.map((template) => (
            <div
              key={template.template_id}
              className="bg-white rounded-lg shadow hover:shadow-md transition-shadow p-5"
            >
              <div className="flex justify-between items-start mb-3">
                <h3 className="text-lg font-semibold text-gray-900 flex-1">
                  {template.template_name}
                </h3>
                <span className="bg-blue-100 text-blue-800 text-xs font-medium px-2.5 py-0.5 rounded">
                  {template.items.length} 枠
                </span>
              </div>

              {template.description && (
                <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                  {template.description}
                </p>
              )}

              <div className="mb-4">
                <h4 className="text-xs font-semibold text-gray-700 mb-2">シフト枠:</h4>
                <div className="space-y-1">
                  {template.items.slice(0, 3).map((item, index) => (
                    <div key={index} className="text-xs text-gray-600 flex items-center">
                      <span className="w-2 h-2 bg-blue-400 rounded-full mr-2"></span>
                      {item.slot_name} ({item.instance_name}) - {item.start_time.substring(0, 5)}~
                      {item.end_time.substring(0, 5)} ({item.required_count}名)
                    </div>
                  ))}
                  {template.items.length > 3 && (
                    <div className="text-xs text-gray-500 ml-4">
                      他 {template.items.length - 3} 件
                    </div>
                  )}
                </div>
              </div>

              <div className="flex gap-2 pt-3 border-t">
                <button
                  onClick={() => navigate(`/events/${eventId}/templates/${template.template_id}`)}
                  className="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 px-3 py-2 rounded text-sm font-medium"
                >
                  詳細
                </button>
                <button
                  onClick={() => navigate(`/events/${eventId}/templates/${template.template_id}/edit`)}
                  className="flex-1 bg-blue-100 hover:bg-blue-200 text-blue-700 px-3 py-2 rounded text-sm font-medium"
                >
                  編集
                </button>
                <button
                  onClick={() => handleDelete(template.template_id, template.template_name)}
                  className="bg-red-100 hover:bg-red-200 text-red-700 px-3 py-2 rounded text-sm font-medium"
                >
                  削除
                </button>
              </div>

              <div className="text-xs text-gray-500 mt-3">
                作成日: {new Date(template.created_at).toLocaleDateString('ja-JP')}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default TemplateList;
