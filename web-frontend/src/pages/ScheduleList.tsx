import { useState } from 'react';
import { createSchedule, type Schedule } from '../lib/api/scheduleApi';

export default function ScheduleList() {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [deadline, setDeadline] = useState('');
  const [candidateDates, setCandidateDates] = useState<string[]>(['', '', '']);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [createdSchedule, setCreatedSchedule] = useState<Schedule | null>(null);
  const [publicUrl, setPublicUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const [submittedCandidatesCount, setSubmittedCandidatesCount] = useState(0);

  const handleAddDate = () => {
    setCandidateDates([...candidateDates, '']);
  };

  const handleRemoveDate = (index: number) => {
    if (candidateDates.length > 1) {
      setCandidateDates(candidateDates.filter((_, i) => i !== index));
    }
  };

  const handleDateChange = (index: number, value: string) => {
    const newDates = [...candidateDates];
    newDates[index] = value;
    setCandidateDates(newDates);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setCreatedSchedule(null);

    if (!title.trim()) {
      setError('タイトルを入力してください');
      return;
    }

    const validDates = candidateDates.filter((d) => d.trim() !== '');
    if (validDates.length === 0) {
      setError('候補日を1つ以上入力してください');
      return;
    }

    setLoading(true);

    try {
      // 候補日の数を保存
      setSubmittedCandidatesCount(validDates.length);

      const result = await createSchedule({
        title: title.trim(),
        description: description.trim(),
        candidates: validDates.map((d) => ({
          date: new Date(d).toISOString(),
        })),
        deadline: deadline ? new Date(deadline).toISOString() : undefined,
      });

      // 公開URLを生成
      const baseUrl = window.location.origin;
      const url = `${baseUrl}/p/schedule/${result.public_token}`;
      setPublicUrl(url);
      setCreatedSchedule(result);

      // フォームをクリア
      setTitle('');
      setDescription('');
      setDeadline('');
      setCandidateDates(['', '', '']);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('日程調整の作成に失敗しました');
      }
      console.error('Create schedule error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(publicUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">日程調整</h1>
        <p className="text-sm text-gray-600 mt-1">
          複数の候補日から、メンバーが参加可能な日程を回答してもらいましょう
        </p>
      </div>

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          新しい日程調整を作成
        </h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              タイトル <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="例：忘年会の日程調整"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              説明
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              placeholder="詳細な説明や注意事項を入力してください"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              候補日 <span className="text-red-500">*</span>
            </label>
            <div className="space-y-2">
              {candidateDates.map((date, index) => (
                <div key={index} className="flex gap-2">
                  <input
                    type="datetime-local"
                    value={date}
                    onChange={(e) => handleDateChange(index, e.target.value)}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={loading}
                  />
                  {candidateDates.length > 1 && (
                    <button
                      type="button"
                      onClick={() => handleRemoveDate(index)}
                      className="px-3 py-2 text-red-600 hover:bg-red-50 rounded-md transition"
                      disabled={loading}
                    >
                      削除
                    </button>
                  )}
                </div>
              ))}
            </div>
            <button
              type="button"
              onClick={handleAddDate}
              className="mt-2 px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded-md transition"
              disabled={loading}
            >
              + 候補日を追加
            </button>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              回答締切（任意）
            </label>
            <input
              type="datetime-local"
              value={deadline}
              onChange={(e) => setDeadline(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={loading}
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-3">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !title.trim()}
            className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {loading ? '作成中...' : '日程調整を作成'}
          </button>
        </form>
      </div>

      {createdSchedule && publicUrl && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-6">
          <div className="flex items-start">
            <div className="text-green-500 text-2xl mr-3">✓</div>
            <div className="flex-1">
              <h3 className="text-lg font-semibold text-green-900 mb-2">
                日程調整を作成しました
              </h3>
              <p className="text-sm text-green-800 mb-4">
                以下のURLをメンバーに送信してください
              </p>

              <div className="bg-white rounded-md p-3 mb-3 border border-green-300">
                <p className="text-xs text-gray-600 mb-1">公開URL:</p>
                <p className="text-sm text-gray-900 font-mono break-all">{publicUrl}</p>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={handleCopy}
                  className="flex-1 px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition text-sm"
                >
                  {copied ? '✓ コピーしました' : 'URLをコピー'}
                </button>
                <a
                  href={publicUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition text-sm text-center"
                >
                  プレビュー
                </a>
              </div>

              <div className="mt-4 pt-4 border-t border-green-200">
                <p className="text-xs text-green-700">
                  <strong>Schedule ID:</strong> {createdSchedule.schedule_id}
                </p>
                <p className="text-xs text-green-700 mt-1">
                  <strong>ステータス:</strong>{' '}
                  {createdSchedule.status === 'open'
                    ? '受付中'
                    : createdSchedule.status === 'decided'
                    ? '日程決定済み'
                    : '締切済み'}
                </p>
                <p className="text-xs text-green-700 mt-1">
                  <strong>候補日:</strong> {submittedCandidatesCount}件
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <h3 className="text-sm font-semibold text-blue-900 mb-2">💡 使い方</h3>
        <ul className="text-sm text-blue-800 space-y-1 list-disc list-inside">
          <li>日程調整を作成すると公開URLが発行されます</li>
          <li>URLをメンバーに送信して、参加可能な日程を回答してもらいましょう</li>
          <li>メンバーは候補日の中から参加可能な日程を複数選択できます</li>
          <li>締切を設定すると、締切後は回答できなくなります</li>
        </ul>
      </div>
    </div>
  );
}
