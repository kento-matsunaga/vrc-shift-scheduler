import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { SEO } from '../components/seo';
import CalendarGrid from '../components/CalendarGrid';
import {
  getCalendars,
  createCalendar,
  updateCalendar,
  deleteCalendar,
  getPublicCalendarUrl,
} from '../lib/api/calendarApi';
import type { Calendar } from '../lib/api/calendarApi';
import type { PublicEvent } from '../lib/api/publicApi';
import { getEvents, getEventDetail, getEventBusinessDays } from '../lib/api';
import type { Event } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function CalendarList() {
  const [calendars, setCalendars] = useState<Calendar[]>([]);
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingCalendar, setEditingCalendar] = useState<Calendar | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [previewCalendar, setPreviewCalendar] = useState<Calendar | null>(null);
  const [previewEvents, setPreviewEvents] = useState<PublicEvent[]>([]);
  const [previewLoading, setPreviewLoading] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [calendarData, eventData] = await Promise.all([
        getCalendars(),
        getEvents({ is_active: true }),
      ]);
      setCalendars(calendarData.calendars || []);
      setEvents(eventData.events || []);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('データの取得に失敗しました');
      }
      console.error('Failed to load data:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    loadData();
    setSuccess('カレンダーを作成しました');
    setTimeout(() => setSuccess(''), 3000);
  };

  const handleEditSuccess = () => {
    setEditingCalendar(null);
    loadData();
    setSuccess('カレンダーを更新しました');
    setTimeout(() => setSuccess(''), 3000);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('このカレンダーを削除しますか？')) {
      return;
    }

    setDeletingId(id);
    setError('');

    try {
      await deleteCalendar(id);
      setCalendars(calendars.filter((c) => c.id !== id));
      setSuccess('カレンダーを削除しました');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('カレンダーの削除に失敗しました');
      }
      console.error('Failed to delete calendar:', err);
    } finally {
      setDeletingId(null);
    }
  };

  const handleCopyUrl = async (publicToken: string) => {
    const url = getPublicCalendarUrl(publicToken);
    try {
      await navigator.clipboard.writeText(url);
      setSuccess('URLをコピーしました');
      setTimeout(() => setSuccess(''), 3000);
    } catch {
      setError('URLのコピーに失敗しました');
    }
  };

  const handlePreview = async (calendar: Calendar) => {
    setPreviewLoading(true);
    setPreviewCalendar(calendar);

    try {
      // カレンダーに紐づくイベントの詳細を取得
      const eventDetails = await Promise.all(
        calendar.event_ids.map(async (eventId) => {
          try {
            const event = await getEventDetail(eventId);
            const businessDays = await getEventBusinessDays(eventId);

            // PublicEvent形式に変換
            return {
              title: event.event_name,
              description: event.description || '',
              business_days: businessDays.map((bd) => ({
                date: bd.target_date,
                start_time: bd.start_time,
                end_time: bd.end_time,
              })),
            } as PublicEvent;
          } catch {
            return null;
          }
        })
      );

      setPreviewEvents(eventDetails.filter((e): e is PublicEvent => e !== null));
    } catch (err) {
      console.error('Failed to load preview:', err);
      setError('プレビューの読み込みに失敗しました');
    } finally {
      setPreviewLoading(false);
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

  return (
    <div>
      <SEO noindex={true} />
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-6">
        <h2 className="text-xl sm:text-2xl font-bold text-gray-900">カレンダー一覧</h2>
        <button onClick={() => setShowCreateModal(true)} className="btn-primary text-sm sm:text-base w-full sm:w-auto">
          ＋ 新しいカレンダー
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

      {calendars.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだカレンダーがありません</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初のカレンダーを作成
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {calendars.map((calendar) => (
            <div key={calendar.id} className="card">
              {/* 公開状態バッジ */}
              <div className="mb-2 flex flex-wrap gap-2">
                <span
                  className={`inline-block px-2 py-1 text-xs font-semibold rounded ${
                    calendar.is_public
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}
                >
                  {calendar.is_public ? '公開中' : '非公開'}
                </span>
                <span className="inline-block px-2 py-1 text-xs font-semibold rounded bg-accent/10 text-accent-dark">
                  {calendar.event_ids.length} イベント
                </span>
              </div>

              {/* タイトル */}
              <h3 className="text-xl font-bold text-gray-900 mb-2">{calendar.title}</h3>

              {/* 説明 */}
              <p className="text-sm text-gray-600 mb-3">
                {calendar.description || '説明なし'}
              </p>

              {/* 作成日 */}
              <div className="text-xs text-gray-500 mb-4">
                作成日: {new Date(calendar.created_at).toLocaleDateString('ja-JP')}
              </div>

              {/* 公開URL（公開中の場合のみ表示） */}
              {calendar.is_public && calendar.public_token && (
                <div className="mb-4 p-2 bg-gray-50 rounded-md">
                  <p className="text-xs text-gray-500 mb-1">公開URL:</p>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 text-xs bg-white p-1 rounded border truncate">
                      {getPublicCalendarUrl(calendar.public_token)}
                    </code>
                    <button
                      onClick={() => handleCopyUrl(calendar.public_token!)}
                      className="p-1 text-accent hover:text-accent-dark transition-colors"
                      title="URLをコピー"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                    </button>
                  </div>
                </div>
              )}

              {/* アクションボタン */}
              <div className="flex gap-2">
                <Link
                  to={`/calendars/${calendar.id}`}
                  className="flex-1 px-3 py-2 text-sm text-center text-accent bg-accent/10 hover:bg-accent/20 rounded-md transition-colors"
                >
                  詳細
                </Link>
                <button
                  onClick={() => handlePreview(calendar)}
                  className="flex-1 px-3 py-2 text-sm text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
                >
                  プレビュー
                </button>
                <button
                  onClick={() => setEditingCalendar(calendar)}
                  className="px-3 py-2 text-sm text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
                >
                  編集
                </button>
                <button
                  onClick={() => handleDelete(calendar.id)}
                  disabled={deletingId === calendar.id}
                  className="px-3 py-2 text-sm text-red-600 bg-red-50 hover:bg-red-100 rounded-md transition-colors disabled:opacity-50"
                >
                  {deletingId === calendar.id ? '削除中...' : '削除'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 作成モーダル */}
      {showCreateModal && (
        <CalendarFormModal
          events={events}
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}

      {/* 編集モーダル */}
      {editingCalendar && (
        <CalendarFormModal
          calendar={editingCalendar}
          events={events}
          onClose={() => setEditingCalendar(null)}
          onSuccess={handleEditSuccess}
        />
      )}

      {/* プレビューモーダル */}
      {previewCalendar && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between p-4 border-b">
              <h3 className="text-xl font-bold text-gray-900">
                {previewCalendar.title} - プレビュー
              </h3>
              <button
                onClick={() => {
                  setPreviewCalendar(null);
                  setPreviewEvents([]);
                }}
                className="p-2 hover:bg-gray-100 rounded-full"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-4">
              {previewLoading ? (
                <div className="text-center py-12">
                  <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
                  <p className="mt-4 text-gray-600">読み込み中...</p>
                </div>
              ) : previewEvents.length > 0 ? (
                <CalendarGrid events={previewEvents} />
              ) : (
                <div className="text-center py-12 text-gray-500">
                  表示するイベントがありません
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// カレンダー作成/編集モーダル
function CalendarFormModal({
  calendar,
  events,
  onClose,
  onSuccess,
}: {
  calendar?: Calendar;
  events: Event[];
  onClose: () => void;
  onSuccess: () => void;
}) {
  const isEditing = !!calendar;
  const [title, setTitle] = useState(calendar?.title || '');
  const [description, setDescription] = useState(calendar?.description || '');
  const [selectedEventIds, setSelectedEventIds] = useState<string[]>(calendar?.event_ids || []);
  const [isPublic, setIsPublic] = useState(calendar?.is_public || false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!title.trim()) {
      setError('タイトルを入力してください');
      return;
    }

    setLoading(true);

    try {
      if (isEditing && calendar) {
        await updateCalendar(calendar.id, {
          title: title.trim(),
          description: description.trim(),
          event_ids: selectedEventIds,
          is_public: isPublic,
        });
      } else {
        await createCalendar({
          title: title.trim(),
          description: description.trim(),
          event_ids: selectedEventIds,
        });
      }
      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError(isEditing ? 'カレンダーの更新に失敗しました' : 'カレンダーの作成に失敗しました');
      }
      console.error('Failed to save calendar:', err);
    } finally {
      setLoading(false);
    }
  };

  const toggleEvent = (eventId: string) => {
    setSelectedEventIds((prev) =>
      prev.includes(eventId)
        ? prev.filter((id) => id !== eventId)
        : [...prev, eventId]
    );
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6 max-h-[90vh] overflow-y-auto">
        <h3 className="text-xl font-bold text-gray-900 mb-4">
          {isEditing ? 'カレンダーを編集' : '新しいカレンダーを作成'}
        </h3>

        <form onSubmit={handleSubmit}>
          {/* タイトル */}
          <div className="mb-4">
            <label htmlFor="title" className="label">
              タイトル <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="例: ○○イベントのスケジュール"
              className="input-field"
              disabled={loading}
              autoFocus
            />
          </div>

          {/* 説明 */}
          <div className="mb-4">
            <label htmlFor="description" className="label">
              説明
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="カレンダーの説明を入力"
              className="input-field"
              rows={3}
              disabled={loading}
            />
          </div>

          {/* イベント選択 */}
          <div className="mb-4">
            <label className="label">対象イベント</label>
            {events.length === 0 ? (
              <p className="text-sm text-gray-500 py-2">イベントがありません</p>
            ) : (
              <div className="space-y-2 max-h-40 overflow-y-auto border border-gray-200 rounded-md p-2">
                {events.map((event) => (
                  <label
                    key={event.event_id}
                    className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-1 rounded"
                  >
                    <input
                      type="checkbox"
                      checked={selectedEventIds.includes(event.event_id)}
                      onChange={() => toggleEvent(event.event_id)}
                      className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                      disabled={loading}
                    />
                    <span className="text-sm text-gray-900">{event.event_name}</span>
                  </label>
                ))}
              </div>
            )}
            <p className="text-xs text-gray-500 mt-1">
              選択: {selectedEventIds.length}件
            </p>
          </div>

          {/* 公開設定（編集時のみ） */}
          {isEditing && (
            <div className="mb-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={isPublic}
                  onChange={(e) => setIsPublic(e.target.checked)}
                  className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                  disabled={loading}
                />
                <span className="text-sm text-gray-900">公開する</span>
              </label>
              <p className="text-xs text-gray-500 mt-1">
                公開すると、URLを知っている人がカレンダーを閲覧できます
              </p>
            </div>
          )}

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
              disabled={loading || !title.trim()}
            >
              {loading ? (isEditing ? '更新中...' : '作成中...') : (isEditing ? '更新' : '作成')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
