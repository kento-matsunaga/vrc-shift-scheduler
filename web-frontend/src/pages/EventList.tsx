import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { getEvents, createEvent, generateBusinessDays, updateEvent } from '../lib/api';
import type { Event } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

// 曜日名の配列
const DAY_NAMES = ['日', '月', '火', '水', '木', '金', '土'];

// 定期タイプの表示名
const RECURRENCE_TYPE_LABELS: Record<string, string> = {
  none: 'なし',
  weekly: '毎週',
  biweekly: '隔週',
};

export default function EventList() {
  const navigate = useNavigate();
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [generatingEventId, setGeneratingEventId] = useState<string | null>(null);
  const [editingEventId, setEditingEventId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [savingEventId, setSavingEventId] = useState<string | null>(null);
  const editInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    loadEvents();
  }, []);

  const loadEvents = async () => {
    try {
      setLoading(true);
      const data = await getEvents({ is_active: true });
      setEvents(data.events);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベント一覧の取得に失敗しました');
      }
      console.error('Failed to load events:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    loadEvents();
  };

  const handleGenerateBusinessDays = async (eventId: string, e: React.MouseEvent) => {
    e.preventDefault(); // Link のクリックを防止
    e.stopPropagation();

    setGeneratingEventId(eventId);
    setError('');
    setSuccess('');

    try {
      const result = await generateBusinessDays(eventId);
      setSuccess(result.message);
      // 3秒後にメッセージを消す
      setTimeout(() => setSuccess(''), 5000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('営業日の生成に失敗しました');
      }
      console.error('Failed to generate business days:', err);
    } finally {
      setGeneratingEventId(null);
    }
  };

  // 編集モードを開始
  const handleStartEdit = (event: Event, e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setEditingEventId(event.event_id);
    setEditingName(event.event_name);
    setError('');
    // 次のレンダリングでinputにフォーカス
    setTimeout(() => editInputRef.current?.focus(), 0);
  };

  // 編集をキャンセル
  const handleCancelEdit = (e?: React.MouseEvent | React.KeyboardEvent) => {
    e?.preventDefault();
    e?.stopPropagation();
    setEditingEventId(null);
    setEditingName('');
  };

  // 編集を保存
  const handleSaveEdit = async (eventId: string, e?: React.MouseEvent | React.FormEvent) => {
    e?.preventDefault();
    e?.stopPropagation();

    if (!editingName.trim()) {
      setError('イベント名を入力してください');
      return;
    }

    setSavingEventId(eventId);
    setError('');

    try {
      await updateEvent(eventId, { event_name: editingName.trim() });
      // ローカルのevents配列を更新
      setEvents(events.map(ev =>
        ev.event_id === eventId ? { ...ev, event_name: editingName.trim() } : ev
      ));
      setEditingEventId(null);
      setEditingName('');
      setSuccess('イベント名を更新しました');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベント名の更新に失敗しました');
      }
      console.error('Failed to update event:', err);
    } finally {
      setSavingEventId(null);
    }
  };

  // キーボードイベント処理
  const handleEditKeyDown = (eventId: string, e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveEdit(eventId, e);
    } else if (e.key === 'Escape') {
      handleCancelEdit(e);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">イベント一覧</h2>
        <button onClick={() => setShowCreateModal(true)} className="btn-primary">
          ＋ 新しいイベントを作成
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {success && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-green-800">{success}</p>
        </div>
      )}

      {events.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだイベントがありません</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初のイベントを作成
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {events.map((event) => (
            <div
              key={event.event_id}
              className="card hover:shadow-lg transition-shadow cursor-pointer"
              onClick={() => {
                // 編集中でなければ遷移
                if (editingEventId !== event.event_id) {
                  navigate(`/events/${event.event_id}/business-days`);
                }
              }}
            >
              {/* バッジ類 */}
              <div className="mb-2 flex flex-wrap gap-2">
                <span
                  className={`inline-block px-2 py-1 text-xs font-semibold rounded ${
                    event.event_type === 'normal'
                      ? 'bg-blue-100 text-blue-800'
                      : 'bg-purple-100 text-purple-800'
                  }`}
                >
                  {event.event_type === 'normal' ? '通常イベント' : '特別イベント'}
                </span>
                {event.recurrence_type !== 'none' && (
                  <span className="inline-block px-2 py-1 text-xs font-semibold rounded bg-green-100 text-green-800">
                    {RECURRENCE_TYPE_LABELS[event.recurrence_type]}
                    {event.recurrence_day_of_week !== undefined && `（${DAY_NAMES[event.recurrence_day_of_week]}）`}
                  </span>
                )}
              </div>

              {/* イベント名（編集可能） */}
              {editingEventId === event.event_id ? (
                <div className="mb-2" onClick={(e) => e.stopPropagation()}>
                  <div className="flex items-center gap-2">
                    <input
                      ref={editInputRef}
                      type="text"
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                      onKeyDown={(e) => handleEditKeyDown(event.event_id, e)}
                      className="flex-1 px-2 py-1 text-lg font-bold border border-blue-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                      disabled={savingEventId === event.event_id}
                    />
                    <button
                      onClick={(e) => handleSaveEdit(event.event_id, e)}
                      disabled={savingEventId === event.event_id}
                      className="p-1 text-green-600 hover:text-green-800 disabled:opacity-50"
                      title="保存"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      disabled={savingEventId === event.event_id}
                      className="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50"
                      title="キャンセル"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-2 mb-2">
                  <h3 className="text-xl font-bold text-gray-900">
                    {event.event_name}
                  </h3>
                  <button
                    onClick={(e) => handleStartEdit(event, e)}
                    className="p-1 text-gray-400 hover:text-blue-600 transition-colors"
                    title="イベント名を編集"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                    </svg>
                  </button>
                </div>
              )}

              {/* イベント詳細 */}
              <p className="text-sm text-gray-600 mb-2">{event.description || '説明なし'}</p>
              {event.recurrence_type !== 'none' && event.default_start_time && event.default_end_time && (
                <p className="text-xs text-gray-500 mb-2">
                  時間: {event.default_start_time.slice(0, 5)}〜{event.default_end_time.slice(0, 5)}
                </p>
              )}
              <div className="text-xs text-gray-500 mb-3">
                作成日: {new Date(event.created_at).toLocaleDateString('ja-JP')}
              </div>

              {event.recurrence_type !== 'none' && (
                <button
                  onClick={(e) => handleGenerateBusinessDays(event.event_id, e)}
                  disabled={generatingEventId === event.event_id}
                  className="w-full btn-secondary text-sm py-2"
                >
                  {generatingEventId === event.event_id ? '生成中...' : '営業日を自動生成'}
                </button>
              )}
            </div>
          ))}
        </div>
      )}

      {/* イベント作成モーダル */}
      {showCreateModal && (
        <CreateEventModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}
    </div>
  );
}

// イベント作成モーダルコンポーネント
function CreateEventModal({
  onClose,
  onSuccess,
}: {
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [eventName, setEventName] = useState('');
  const [eventType, setEventType] = useState<'normal' | 'special'>('normal');
  const [description, setDescription] = useState('');
  const [recurrenceType, setRecurrenceType] = useState<'none' | 'weekly' | 'biweekly'>('none');
  const [recurrenceDayOfWeek, setRecurrenceDayOfWeek] = useState<number>(6); // デフォルト土曜
  const [recurrenceStartDate, setRecurrenceStartDate] = useState(() => {
    // デフォルトは今日
    return new Date().toISOString().split('T')[0];
  });
  const [defaultStartTime, setDefaultStartTime] = useState('21:00');
  const [defaultEndTime, setDefaultEndTime] = useState('23:30');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!eventName.trim()) {
      setError('イベント名を入力してください');
      return;
    }

    // 定期イベントの場合のバリデーション
    if (recurrenceType !== 'none') {
      if (!defaultStartTime || !defaultEndTime) {
        setError('定期イベントの場合は開始時刻と終了時刻を入力してください');
        return;
      }
      if (recurrenceType === 'biweekly' && !recurrenceStartDate) {
        setError('隔週イベントの場合は開始日を入力してください');
        return;
      }
    }

    setLoading(true);

    try {
      const requestData: Parameters<typeof createEvent>[0] = {
        event_name: eventName.trim(),
        event_type: eventType,
        description: description.trim(),
        recurrence_type: recurrenceType,
      };

      // 定期イベントの場合のみ追加フィールドを送信
      if (recurrenceType !== 'none') {
        requestData.recurrence_day_of_week = recurrenceDayOfWeek;
        requestData.default_start_time = defaultStartTime + ':00'; // HH:MM:SS形式
        requestData.default_end_time = defaultEndTime + ':00';
        requestData.recurrence_start_date = recurrenceStartDate;
      }

      await createEvent(requestData);
      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベントの作成に失敗しました');
      }
      console.error('Failed to create event:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6 max-h-[90vh] overflow-y-auto">
        <h3 className="text-xl font-bold text-gray-900 mb-4">新しいイベントを作成</h3>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="eventName" className="label">
              イベント名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="eventName"
              value={eventName}
              onChange={(e) => setEventName(e.target.value)}
              placeholder="例: VRChat 交流会"
              className="input-field"
              disabled={loading}
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label htmlFor="eventType" className="label">
              イベント種別
            </label>
            <select
              id="eventType"
              value={eventType}
              onChange={(e) => setEventType(e.target.value as 'normal' | 'special')}
              className="input-field"
              disabled={loading}
            >
              <option value="normal">通常イベント</option>
              <option value="special">特別イベント</option>
            </select>
          </div>

          <div className="mb-4">
            <label htmlFor="recurrenceType" className="label">
              開催頻度
            </label>
            <select
              id="recurrenceType"
              value={recurrenceType}
              onChange={(e) => setRecurrenceType(e.target.value as 'none' | 'weekly' | 'biweekly')}
              className="input-field"
              disabled={loading}
            >
              <option value="none">不定期（都度設定）</option>
              <option value="weekly">毎週</option>
              <option value="biweekly">隔週</option>
            </select>
          </div>

          {/* 定期イベントの場合のみ表示 */}
          {recurrenceType !== 'none' && (
            <>
              <div className="mb-4">
                <label htmlFor="recurrenceDayOfWeek" className="label">
                  開催曜日 <span className="text-red-500">*</span>
                </label>
                <select
                  id="recurrenceDayOfWeek"
                  value={recurrenceDayOfWeek}
                  onChange={(e) => setRecurrenceDayOfWeek(Number(e.target.value))}
                  className="input-field"
                  disabled={loading}
                >
                  {DAY_NAMES.map((name, index) => (
                    <option key={index} value={index}>{name}曜日</option>
                  ))}
                </select>
              </div>

              {/* 隔週の場合は開始日を必須で表示 */}
              {recurrenceType === 'biweekly' && (
                <div className="mb-4">
                  <label htmlFor="recurrenceStartDate" className="label">
                    開始日（基準日） <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="date"
                    id="recurrenceStartDate"
                    value={recurrenceStartDate}
                    onChange={(e) => setRecurrenceStartDate(e.target.value)}
                    className="input-field"
                    disabled={loading}
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    この日付を基準に隔週で営業日が生成されます
                  </p>
                </div>
              )}

              <div className="mb-4 grid grid-cols-2 gap-4">
                <div>
                  <label htmlFor="defaultStartTime" className="label">
                    開始時刻 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="time"
                    id="defaultStartTime"
                    value={defaultStartTime}
                    onChange={(e) => setDefaultStartTime(e.target.value)}
                    className="input-field"
                    disabled={loading}
                  />
                </div>
                <div>
                  <label htmlFor="defaultEndTime" className="label">
                    終了時刻 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="time"
                    id="defaultEndTime"
                    value={defaultEndTime}
                    onChange={(e) => setDefaultEndTime(e.target.value)}
                    className="input-field"
                    disabled={loading}
                  />
                </div>
              </div>

              <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                <p className="text-sm text-blue-800">
                  定期イベントを作成後、「営業日を自動生成」ボタンで今月〜来月の営業日をまとめて作成できます。
                </p>
              </div>
            </>
          )}

          <div className="mb-4">
            <label htmlFor="description" className="label">
              説明
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="イベントの説明を入力"
              className="input-field"
              rows={3}
              disabled={loading}
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 btn-secondary"
              disabled={loading}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="flex-1 btn-primary"
              disabled={loading || !eventName.trim()}
            >
              {loading ? '作成中...' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

