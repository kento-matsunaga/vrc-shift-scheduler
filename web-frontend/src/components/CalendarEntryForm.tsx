import { useState, type FormEvent } from 'react';
import {
  createCalendarEntry,
  updateCalendarEntry,
  type CalendarEntry,
} from '../lib/api/calendarApi';
import { ApiClientError } from '../lib/apiClient';

interface CalendarEntryFormProps {
  calendarId: string;
  entry?: CalendarEntry;
  onSave: () => void;
  onCancel: () => void;
}

export default function CalendarEntryForm({
  calendarId,
  entry,
  onSave,
  onCancel,
}: CalendarEntryFormProps) {
  const isEditing = !!entry;
  const [title, setTitle] = useState(entry?.title || '');
  const [date, setDate] = useState(entry?.date || '');
  const [startTime, setStartTime] = useState(entry?.start_time || '');
  const [endTime, setEndTime] = useState(entry?.end_time || '');
  const [note, setNote] = useState(entry?.note || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');

    if (!title.trim()) {
      setError('タイトルを入力してください');
      return;
    }

    if (!date) {
      setError('日付を選択してください');
      return;
    }

    setLoading(true);
    try {
      const input = {
        title: title.trim(),
        date,
        start_time: startTime || undefined,
        end_time: endTime || undefined,
        note: note.trim() || undefined,
      };

      if (entry) {
        await updateCalendarEntry(calendarId, entry.entry_id, input);
      } else {
        await createCalendarEntry(calendarId, input);
      }
      onSave();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError(isEditing ? '予定の更新に失敗しました' : '予定の作成に失敗しました');
      }
      console.error('Failed to save calendar entry:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <h3 className="text-xl font-bold text-gray-900 mb-4">
          {isEditing ? '予定を編集' : '予定を追加'}
        </h3>

        <form onSubmit={handleSubmit}>
          {/* タイトル */}
          <div className="mb-4">
            <label htmlFor="entryTitle" className="label">
              タイトル <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="entryTitle"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="例: 定例ミーティング"
              className="input-field"
              disabled={loading}
              autoFocus
            />
          </div>

          {/* 日付 */}
          <div className="mb-4">
            <label htmlFor="entryDate" className="label">
              日付 <span className="text-red-500">*</span>
            </label>
            <input
              type="date"
              id="entryDate"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              className="input-field"
              disabled={loading}
            />
          </div>

          {/* 時間 */}
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label htmlFor="entryStartTime" className="label">
                開始時間
              </label>
              <input
                type="time"
                id="entryStartTime"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                className="input-field"
                disabled={loading}
              />
            </div>
            <div>
              <label htmlFor="entryEndTime" className="label">
                終了時間
              </label>
              <input
                type="time"
                id="entryEndTime"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
                className="input-field"
                disabled={loading}
              />
            </div>
          </div>

          <p className="text-xs text-gray-500 mb-4">
            時間は任意です。終日の予定の場合は空欄のままにできます。
          </p>

          {/* 備考 */}
          <div className="mb-4">
            <label htmlFor="entryNote" className="label">
              備考
            </label>
            <textarea
              id="entryNote"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="備考を入力（任意）"
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
              onClick={onCancel}
              className="flex-1 btn-secondary"
              disabled={loading}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="flex-1 btn-primary"
              disabled={loading || !title.trim() || !date}
            >
              {loading
                ? isEditing
                  ? '更新中...'
                  : '作成中...'
                : isEditing
                ? '更新'
                : '追加'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
