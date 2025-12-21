import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { getTemplate, createTemplate, updateTemplate } from '../lib/api/templateApi';
import type { TemplateItem, CreateTemplateRequest } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function TemplateForm() {
  const { eventId, templateId } = useParams<{ eventId: string; templateId?: string }>();
  const navigate = useNavigate();
  const isEditMode = !!templateId;

  const [templateName, setTemplateName] = useState('');
  const [description, setDescription] = useState('');
  const [items, setItems] = useState<TemplateItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingData, setLoadingData] = useState(isEditMode);
  const [error, setError] = useState('');

  // デフォルトのポジションID（既存のShiftSlotListと同じ固定値）
  const defaultPositionId = '01KCMHNFRVKWY3SY44BXSBNVTT';

  useEffect(() => {
    if (isEditMode && templateId && eventId) {
      loadTemplate();
    }
  }, [isEditMode, templateId, eventId]);

  const loadTemplate = async () => {
    if (!eventId || !templateId) return;

    try {
      setLoadingData(true);
      const template = await getTemplate(eventId, templateId);
      setTemplateName(template.template_name);
      setDescription(template.description);
      setItems(template.items || []);
    } catch (err) {
      console.error('Failed to load template:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('テンプレートの読み込みに失敗しました');
      }
    } finally {
      setLoadingData(false);
    }
  };

  const addItem = () => {
    setItems([
      ...items,
      {
        position_id: defaultPositionId,
        slot_name: '',
        instance_name: '',
        start_time: '21:30:00',
        end_time: '23:00:00',
        required_count: 1,
        priority: items.length + 1,
      },
    ]);
  };

  const removeItem = (index: number) => {
    setItems(items.filter((_, i) => i !== index));
  };

  const updateItem = (index: number, field: keyof TemplateItem, value: string | number) => {
    const newItems = [...items];
    newItems[index] = { ...newItems[index], [field]: value };
    setItems(newItems);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!eventId) {
      setError('イベントIDが指定されていません');
      return;
    }

    if (!templateName.trim()) {
      setError('テンプレート名を入力してください');
      return;
    }

    if (items.length === 0) {
      setError('少なくとも1つのシフト枠を追加してください');
      return;
    }

    // バリデーション
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (!item.slot_name.trim()) {
        setError(`シフト枠 ${i + 1}: シフト名を入力してください`);
        return;
      }
      if (!item.instance_name.trim()) {
        setError(`シフト枠 ${i + 1}: インスタンス名を入力してください`);
        return;
      }
      if (item.required_count < 1) {
        setError(`シフト枠 ${i + 1}: 必要人数は1以上にしてください`);
        return;
      }
    }

    setLoading(true);

    try {
      const data: CreateTemplateRequest = {
        template_name: templateName.trim(),
        description: description.trim(),
        items: items.map((item) => ({
          position_id: defaultPositionId,
          slot_name: item.slot_name.trim(),
          instance_name: item.instance_name.trim(),
          start_time: item.start_time,
          end_time: item.end_time,
          required_count: item.required_count,
          priority: item.priority,
        })),
      };

      if (isEditMode && templateId) {
        await updateTemplate(eventId, templateId, data);
      } else {
        await createTemplate(eventId, data);
      }

      navigate(`/events/${eventId}/templates`);
    } catch (err) {
      console.error('Failed to save template:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError(isEditMode ? 'テンプレートの更新に失敗しました' : 'テンプレートの作成に失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  if (loadingData) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">
          {isEditMode ? 'テンプレートを編集' : '新規テンプレート作成'}
        </h2>
        <p className="text-sm text-gray-600 mt-1">
          営業日作成時に使用するシフト枠のテンプレートを{isEditMode ? '編集' : '作成'}します
        </p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">基本情報</h3>

          <div className="mb-4">
            <label htmlFor="templateName" className="label">
              テンプレート名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="templateName"
              value={templateName}
              onChange={(e) => setTemplateName(e.target.value)}
              className="input-field"
              disabled={loading}
              placeholder="例: 土曜日テンプレート、通常営業日テンプレート"
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label htmlFor="description" className="label">
              説明
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="input-field"
              disabled={loading}
              rows={3}
              placeholder="このテンプレートの用途や特徴を入力してください"
            />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-gray-900">シフト枠設定</h3>
            <button
              type="button"
              onClick={addItem}
              disabled={loading}
              className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg text-sm flex items-center"
            >
              <svg
                className="w-5 h-5 mr-1"
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
              シフト枠を追加
            </button>
          </div>

          {items.length === 0 ? (
            <div className="text-center py-8 bg-gray-50 rounded-lg">
              <p className="text-gray-600 mb-4">まだシフト枠が追加されていません</p>
              <button
                type="button"
                onClick={addItem}
                disabled={loading}
                className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg text-sm"
              >
                最初のシフト枠を追加
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {items.map((item, index) => (
                <div key={index} className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                  <div className="flex justify-between items-start mb-3">
                    <h4 className="font-semibold text-gray-900">シフト枠 {index + 1}</h4>
                    <button
                      type="button"
                      onClick={() => removeItem(index)}
                      disabled={loading}
                      className="text-red-600 hover:text-red-800 text-sm"
                    >
                      削除
                    </button>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="label text-sm">
                        シフト名 <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="text"
                        value={item.slot_name}
                        onChange={(e) => updateItem(index, 'slot_name', e.target.value)}
                        className="input-field text-sm"
                        disabled={loading}
                        placeholder="例: 受付、案内、配信"
                      />
                    </div>

                    <div>
                      <label className="label text-sm">
                        インスタンス名 <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="text"
                        value={item.instance_name}
                        onChange={(e) => updateItem(index, 'instance_name', e.target.value)}
                        className="input-field text-sm"
                        disabled={loading}
                        placeholder="例: 午前、午後、終日、A、B"
                      />
                    </div>

                    <div>
                      <label className="label text-sm">
                        開始時刻 <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="time"
                        value={item.start_time.substring(0, 5)}
                        onChange={(e) => updateItem(index, 'start_time', e.target.value + ':00')}
                        className="input-field text-sm"
                        disabled={loading}
                      />
                    </div>

                    <div>
                      <label className="label text-sm">
                        終了時刻 <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="time"
                        value={item.end_time.substring(0, 5)}
                        onChange={(e) => updateItem(index, 'end_time', e.target.value + ':00')}
                        className="input-field text-sm"
                        disabled={loading}
                      />
                    </div>

                    <div>
                      <label className="label text-sm">
                        必要人数 <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="number"
                        value={item.required_count}
                        onChange={(e) => updateItem(index, 'required_count', Number(e.target.value))}
                        className="input-field text-sm"
                        disabled={loading}
                        min="1"
                      />
                    </div>

                    <div>
                      <label className="label text-sm">優先度</label>
                      <input
                        type="number"
                        value={item.priority}
                        onChange={(e) => updateItem(index, 'priority', Number(e.target.value))}
                        className="input-field text-sm"
                        disabled={loading}
                        min="1"
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex gap-4">
          <button
            type="button"
            onClick={() => navigate(`/events/${eventId}/templates`)}
            disabled={loading}
            className="flex-1 bg-gray-200 hover:bg-gray-300 text-gray-700 px-6 py-3 rounded-lg font-medium"
          >
            キャンセル
          </button>
          <button
            type="submit"
            disabled={loading || items.length === 0}
            className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white px-6 py-3 rounded-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '保存中...' : isEditMode ? '更新する' : '作成する'}
          </button>
        </div>
      </form>
    </div>
  );
}
