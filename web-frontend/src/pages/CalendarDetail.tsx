import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { SEO } from '../components/seo';
import CalendarGrid from '../components/CalendarGrid';
import CalendarEntryForm from '../components/CalendarEntryForm';
import {
  getCalendarById,
  getCalendarEntries,
  deleteCalendarEntry,
  getPublicCalendarUrl,
  type Calendar,
  type CalendarEntry,
} from '../lib/api/calendarApi';
import { getEventDetail, getEventBusinessDays } from '../lib/api';
import type { PublicEvent, PublicCalendarEntry } from '../lib/api/publicApi';
import { ApiClientError } from '../lib/apiClient';

export default function CalendarDetail() {
  const { calendarId } = useParams<{ calendarId: string }>();
  const [calendar, setCalendar] = useState<Calendar | null>(null);
  const [entries, setEntries] = useState<CalendarEntry[]>([]);
  const [previewEvents, setPreviewEvents] = useState<PublicEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showEntryForm, setShowEntryForm] = useState(false);
  const [editingEntry, setEditingEntry] = useState<CalendarEntry | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  useEffect(() => {
    if (calendarId) {
      loadData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- 初回マウント時のみ実行（loadDataは関数定義のため除外）
  }, [calendarId]);

  const loadData = async () => {
    if (!calendarId) return;

    try {
      setLoading(true);
      setError('');

      // カレンダー情報とエントリを取得
      const [calendarData, entriesData] = await Promise.all([
        getCalendarById(calendarId),
        getCalendarEntries(calendarId),
      ]);

      setCalendar(calendarData);
      setEntries(entriesData.entries || []);

      // イベントの詳細を取得してプレビュー用に変換
      const eventDetails = await Promise.all(
        calendarData.event_ids.map(async (eventId) => {
          try {
            const event = await getEventDetail(eventId);
            const businessDays = await getEventBusinessDays(eventId);

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

  const handleEntryFormSave = () => {
    setShowEntryForm(false);
    setEditingEntry(null);
    loadData();
    setSuccess(editingEntry ? '予定を更新しました' : '予定を追加しました');
    setTimeout(() => setSuccess(''), 3000);
  };

  const handleEntryFormCancel = () => {
    setShowEntryForm(false);
    setEditingEntry(null);
  };

  const handleDeleteEntry = async (entryId: string) => {
    if (!calendarId) return;
    if (!confirm('この予定を削除しますか？')) return;

    setDeletingId(entryId);
    setError('');

    try {
      await deleteCalendarEntry(calendarId, entryId);
      setEntries(entries.filter((e) => e.entry_id !== entryId));
      setSuccess('予定を削除しました');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('予定の削除に失敗しました');
      }
      console.error('Failed to delete entry:', err);
    } finally {
      setDeletingId(null);
    }
  };

  const handleCopyUrl = async () => {
    if (!calendar?.public_token) return;

    const url = getPublicCalendarUrl(calendar.public_token);
    try {
      await navigator.clipboard.writeText(url);
      setSuccess('URLをコピーしました');
      setTimeout(() => setSuccess(''), 3000);
    } catch {
      setError('URLのコピーに失敗しました');
    }
  };

  // CalendarEntry を PublicCalendarEntry に変換
  const publicEntries: PublicCalendarEntry[] = entries.map((e) => ({
    entry_id: e.entry_id,
    title: e.title,
    date: e.date,
    start_time: e.start_time,
    end_time: e.end_time,
    note: e.note,
  }));

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  if (!calendar) {
    return (
      <div className="card text-center py-12">
        <p className="text-gray-600">カレンダーが見つかりません</p>
        <Link to="/calendars" className="text-accent hover:underline mt-4 inline-block">
          カレンダー一覧に戻る
        </Link>
      </div>
    );
  }

  return (
    <div>
      <SEO noindex={true} />

      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to="/calendars" className="hover:text-gray-900">
          カレンダー一覧
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">{calendar.title}</span>
      </nav>

      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-start gap-4 mb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <h2 className="text-xl sm:text-2xl font-bold text-gray-900">{calendar.title}</h2>
            <span
              className={`inline-block px-2 py-1 text-xs font-semibold rounded ${
                calendar.is_public
                  ? 'bg-green-100 text-green-800'
                  : 'bg-gray-100 text-gray-800'
              }`}
            >
              {calendar.is_public ? '公開中' : '非公開'}
            </span>
          </div>
          {calendar.description && (
            <p className="text-sm text-gray-600">{calendar.description}</p>
          )}
        </div>
        <div className="flex gap-2">
          {calendar.is_public && calendar.public_token && (
            <button
              onClick={handleCopyUrl}
              className="px-4 py-2 text-sm text-accent bg-accent/10 hover:bg-accent/20 rounded-md transition-colors"
            >
              URLをコピー
            </button>
          )}
          <button
            onClick={() => setShowEntryForm(true)}
            className="btn-primary text-sm"
          >
            ＋ 予定を追加
          </button>
        </div>
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

      {/* カレンダーグリッド */}
      <div className="mb-8">
        <CalendarGrid events={previewEvents} entries={publicEntries} />
      </div>

      {/* 予定一覧（編集/削除可能） */}
      <div className="card">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-bold text-gray-900">自由予定一覧</h3>
          <span className="text-sm text-gray-500">{entries.length}件</span>
        </div>

        {entries.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <p className="mb-4">まだ予定がありません</p>
            <button
              onClick={() => setShowEntryForm(true)}
              className="text-accent hover:underline"
            >
              最初の予定を追加
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {entries
              .sort((a, b) => a.date.localeCompare(b.date))
              .map((entry) => (
                <div
                  key={entry.entry_id}
                  className="flex items-center justify-between p-3 bg-emerald-50 rounded-lg"
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium text-gray-900">
                        {new Date(entry.date + 'T00:00:00').toLocaleDateString('ja-JP', {
                          month: 'short',
                          day: 'numeric',
                          weekday: 'short',
                        })}
                      </span>
                      {(entry.start_time || entry.end_time) && (
                        <span className="text-sm text-emerald-600">
                          {entry.start_time && entry.end_time
                            ? `${entry.start_time} - ${entry.end_time}`
                            : entry.start_time || entry.end_time}
                        </span>
                      )}
                    </div>
                    <div className="text-gray-900 truncate">{entry.title}</div>
                    {entry.note && (
                      <div className="text-sm text-gray-600 truncate mt-1">{entry.note}</div>
                    )}
                  </div>
                  <div className="flex gap-2 ml-4">
                    <button
                      onClick={() => {
                        setEditingEntry(entry);
                        setShowEntryForm(true);
                      }}
                      className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded"
                      title="編集"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>
                    <button
                      onClick={() => handleDeleteEntry(entry.entry_id)}
                      disabled={deletingId === entry.entry_id}
                      className="p-2 text-red-500 hover:text-red-700 hover:bg-red-50 rounded disabled:opacity-50"
                      title="削除"
                    >
                      {deletingId === entry.entry_id ? (
                        <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                      ) : (
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      )}
                    </button>
                  </div>
                </div>
              ))}
          </div>
        )}
      </div>

      {/* 予定追加/編集フォーム */}
      {showEntryForm && calendarId && (
        <CalendarEntryForm
          calendarId={calendarId}
          entry={editingEntry || undefined}
          onSave={handleEntryFormSave}
          onCancel={handleEntryFormCancel}
        />
      )}
    </div>
  );
}
