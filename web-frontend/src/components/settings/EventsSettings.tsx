import { useState, useEffect } from 'react';
import { AlertTriangle, Trash2 } from 'lucide-react';
import { getEvents, deleteEvent } from '../../lib/api';
import type { Event } from '../../types/api';
import { ApiClientError } from '../../lib/apiClient';

export function EventsSettings() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<Event | null>(null);
  const [confirmText, setConfirmText] = useState('');
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    loadEvents();
  }, []);

  const loadEvents = async () => {
    try {
      setLoading(true);
      const data = await getEvents({ is_active: true });
      setEvents(data.events || []);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベントの取得に失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteClick = (event: Event) => {
    setDeleteTarget(event);
    setConfirmText('');
    setError('');
  };

  const handleCancelDelete = () => {
    setDeleteTarget(null);
    setConfirmText('');
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    if (confirmText !== deleteTarget.event_name) {
      setError('イベント名が一致しません');
      return;
    }

    setDeleting(true);
    setError('');

    try {
      await deleteEvent(deleteTarget.event_id);
      setEvents(events.filter(e => e.event_id !== deleteTarget.event_id));
      setDeleteTarget(null);
      setConfirmText('');
      setSuccess(`「${deleteTarget.event_name}」を削除しました`);
      setTimeout(() => setSuccess(''), 5000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベントの削除に失敗しました');
      }
    } finally {
      setDeleting(false);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-accent mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 危険操作エリア */}
      <div className="bg-white rounded-lg shadow p-6 border-l-4 border-red-500">
        <h2 className="text-lg font-semibold flex items-center gap-2 mb-4 text-red-600">
          <AlertTriangle className="w-5 h-5" />
          危険な操作
        </h2>

        <p className="text-sm text-gray-500 mb-4">
          以下の操作は取り消しできません。慎重に行ってください。
        </p>

        {/* 警告メッセージ */}
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-amber-800">注意</p>
              <p className="text-sm text-amber-700 mt-1">
                イベントを削除すると、関連する営業日、シフト枠、シフト割り当てなども削除されます。
              </p>
            </div>
          </div>
        </div>

        {error && !deleteTarget && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {success && (
          <div className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-green-800">{success}</p>
          </div>
        )}

        {/* イベント一覧 */}
        {events.length === 0 ? (
          <p className="text-gray-600">削除できるイベントがありません</p>
        ) : (
          <div className="space-y-3">
            {events.map((event) => (
              <div
                key={event.event_id}
                className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border border-gray-200"
              >
                <div>
                  <p className="font-medium text-gray-900">{event.event_name}</p>
                  <p className="text-sm text-gray-500">
                    {event.event_type === 'normal' ? '通常イベント' : '特別イベント'}
                    {event.description && ` - ${event.description}`}
                  </p>
                </div>
                <button
                  onClick={() => handleDeleteClick(event)}
                  className="flex items-center gap-2 px-3 py-1.5 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                  削除
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 削除確認モーダル */}
      {deleteTarget && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center flex-shrink-0">
                <AlertTriangle className="w-6 h-6 text-red-600" />
              </div>
              <h3 className="text-xl font-bold text-gray-900">イベントの削除</h3>
            </div>

            <div className="mb-4">
              <p className="text-gray-700 mb-2">
                「<span className="font-bold text-red-600">{deleteTarget.event_name}</span>」を削除しようとしています。
              </p>
              <p className="text-sm text-gray-600">
                このイベントに関連するすべてのデータ（営業日、シフト枠、シフト割り当てなど）も削除されます。
              </p>
            </div>

            <div className="bg-gray-50 rounded-lg p-4 mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                確認のため、イベント名「<span className="font-mono text-red-600">{deleteTarget.event_name}</span>」を入力してください
              </label>
              <input
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder="イベント名を入力"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
                disabled={deleting}
                autoFocus
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
                onClick={handleCancelDelete}
                className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
                disabled={deleting}
              >
                キャンセル
              </button>
              <button
                type="button"
                onClick={handleConfirmDelete}
                disabled={deleting || confirmText !== deleteTarget.event_name}
                className="flex-1 bg-red-600 hover:bg-red-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                {deleting ? '削除中...' : '削除する'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
