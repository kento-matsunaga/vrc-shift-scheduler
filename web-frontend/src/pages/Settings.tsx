import { useState, useEffect } from 'react';
import { getEvents, deleteEvent } from '../lib/api';
import type { Event } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function Settings() {
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
      console.error('Failed to delete event:', err);
    } finally {
      setDeleting(false);
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
    <div className="max-w-4xl mx-auto">
      <h2 className="text-2xl font-bold text-gray-900 mb-6">基本設定</h2>

      {error && !deleteTarget && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {success && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-green-800">{success}</p>
        </div>
      )}

      {/* イベント削除セクション */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          イベントの削除
        </h3>

        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-4">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div>
              <p className="text-sm font-medium text-amber-800">注意</p>
              <p className="text-sm text-amber-700 mt-1">
                イベントを削除すると、関連する営業日、シフト枠、シフト割り当てなども削除されます。
                この操作は取り消せません。
              </p>
            </div>
          </div>
        </div>

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
                  className="px-3 py-1.5 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded transition-colors"
                >
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
                <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
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
                className="input-field"
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
                className="flex-1 btn-secondary"
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
