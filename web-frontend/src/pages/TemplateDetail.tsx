import { useState, useEffect, useMemo } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { getTemplate, deleteTemplate } from '../lib/api/templateApi';
import type { Template, TemplateItem } from '../types/api';
import { ApiClientError } from '../lib/apiClient';
import { SEO } from '../components/seo';

// インスタンスごとにグループ化した構造
interface InstanceGroup {
  instanceName: string;
  items: TemplateItem[];
}

export default function TemplateDetail() {
  const { eventId, templateId } = useParams<{ eventId: string; templateId: string }>();
  const navigate = useNavigate();
  const [template, setTemplate] = useState<Template | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [expandedInstances, setExpandedInstances] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (eventId && templateId) {
      loadTemplate();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- 初回マウント時のみ実行（loadTemplateは関数定義のため除外）
  }, [eventId, templateId]);

  const loadTemplate = async () => {
    if (!eventId || !templateId) return;

    try {
      setLoading(true);
      const data = await getTemplate(eventId, templateId);
      setTemplate(data);
      // 全インスタンスを展開
      const instanceNames = new Set((data.items || []).map((item) => item.instance_name));
      setExpandedInstances(instanceNames);
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

  // インスタンスごとにグループ化
  const instanceGroups = useMemo((): InstanceGroup[] => {
    if (!template || !template.items) return [];

    const groupMap = new Map<string, TemplateItem[]>();
    template.items.forEach((item) => {
      if (!groupMap.has(item.instance_name)) {
        groupMap.set(item.instance_name, []);
      }
      groupMap.get(item.instance_name)!.push(item);
    });

    // 各グループ内をpriorityでソート
    const groups: InstanceGroup[] = [];
    groupMap.forEach((items, instanceName) => {
      items.sort((a, b) => a.priority - b.priority);
      groups.push({ instanceName, items });
    });

    return groups;
  }, [template]);

  const toggleInstanceExpand = (instanceName: string) => {
    setExpandedInstances((prev) => {
      const next = new Set(prev);
      if (next.has(instanceName)) {
        next.delete(instanceName);
      } else {
        next.add(instanceName);
      }
      return next;
    });
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
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
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
          <Link to={`/events/${eventId}/templates`} className="text-accent hover:text-accent-dark">
            ← テンプレート一覧に戻る
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <SEO noindex={true} />
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
              className="bg-accent/10 hover:bg-accent/20 text-accent-dark px-4 py-2 rounded-lg text-sm font-medium"
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

        <div className="grid grid-cols-3 gap-4 pt-4 border-t border-gray-200">
          <div>
            <p className="text-sm text-gray-600">インスタンス数</p>
            <p className="text-lg font-semibold text-gray-900">{instanceGroups.length} 個</p>
          </div>
          <div>
            <p className="text-sm text-gray-600">役職（シフト枠）数</p>
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

      {/* インスタンス・役職一覧 */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">インスタンス・役職一覧</h3>

        {instanceGroups.length === 0 ? (
          <p className="text-gray-600 text-center py-8">シフト枠がありません</p>
        ) : (
          <div className="space-y-4">
            {instanceGroups.map((group) => (
              <div
                key={group.instanceName}
                className="border border-gray-200 rounded-lg overflow-hidden"
              >
                {/* インスタンスヘッダー */}
                <div
                  className="bg-gray-100 px-4 py-3 flex justify-between items-center cursor-pointer"
                  onClick={() => toggleInstanceExpand(group.instanceName)}
                >
                  <div className="flex items-center gap-3">
                    <svg
                      className={`w-5 h-5 text-gray-500 transition-transform ${
                        expandedInstances.has(group.instanceName) ? 'rotate-90' : ''
                      }`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 5l7 7-7 7"
                      />
                    </svg>
                    <span className="font-semibold text-gray-900">{group.instanceName}</span>
                    <span className="text-sm text-gray-500">
                      ({group.items.length}役職)
                    </span>
                  </div>
                  <span className="bg-accent/10 text-accent-dark text-xs font-medium px-2.5 py-0.5 rounded">
                    計 {group.items.reduce((sum, item) => sum + item.required_count, 0)}名
                  </span>
                </div>

                {/* 役職リスト */}
                {expandedInstances.has(group.instanceName) && (
                  <div className="p-4 bg-white">
                    <div className="space-y-3">
                      {group.items.map((item, index) => (
                        <div
                          key={index}
                          className="border border-gray-200 rounded-lg p-4 bg-gray-50 hover:shadow-md transition-shadow"
                        >
                          <div className="flex items-start justify-between mb-2">
                            <div>
                              <h4 className="font-semibold text-gray-900">{item.slot_name}</h4>
                              <p className="text-xs text-gray-500 mt-0.5">優先度: {item.priority}</p>
                            </div>
                            <span className="bg-accent/10 text-accent-dark text-xs font-medium px-2.5 py-0.5 rounded">
                              {item.required_count}名
                            </span>
                          </div>

                          <div className="flex items-center gap-4 text-sm text-gray-600">
                            <div className="flex items-center gap-1">
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                              </svg>
                              <span>{item.start_time.substring(0, 5)} - {item.end_time.substring(0, 5)}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 使用方法の説明 */}
      <div className="bg-accent/10 border border-accent/30 rounded-lg p-4 mt-6">
        <h4 className="font-semibold text-accent-dark mb-2">テンプレートの使い方</h4>
        <p className="text-sm text-accent-dark">
          このテンプレートは営業日作成時に選択することで、登録されているインスタンスと役職を自動的に作成します。
          営業日一覧ページから「営業日を追加」を選択し、テンプレートを選んでください。
        </p>
      </div>
    </div>
  );
}
