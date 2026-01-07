import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  getAttendanceByToken,
  getMembers,
  submitAttendanceResponse,
  type AttendanceCollection,
  type Member,
  type TargetDate,
  PublicApiError,
} from '../../lib/api/publicApi';

export default function AttendanceResponse() {
  const { token } = useParams<{ token: string }>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [collection, setCollection] = useState<AttendanceCollection | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [targetDates, setTargetDates] = useState<TargetDate[]>([]);

  // フォーム状態
  const [selectedMemberId, setSelectedMemberId] = useState('');
  // 各対象日ごとの出欠状態: { target_date_id: 'attending' | 'absent' | 'undecided' }
  const [responses, setResponses] = useState<Record<string, 'attending' | 'absent' | 'undecided'>>({});
  // 各対象日ごとの参加可能時間: { target_date_id: { from: string, to: string } }
  const [availableTimes, setAvailableTimes] = useState<Record<string, { from: string; to: string }>>({});
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => {
    if (!token) {
      setError('URLが無効です');
      setLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        // 出欠確認情報を取得
        const collectionData = await getAttendanceByToken(token);
        console.log('Attendance collection data:', collectionData);
        setCollection(collectionData);

        // メンバー一覧を取得（グループとロールでフィルタリング）
        const membersData = await getMembers(
          collectionData.tenant_id,
          collectionData.group_ids,
          collectionData.role_ids
        );
        setMembers(membersData.data?.members || []);

        // Target dates を設定
        const targetDatesList = collectionData.target_dates || [];
        setTargetDates(targetDatesList);

        // 初期状態として全て「参加」を設定
        const initialResponses: Record<string, 'attending' | 'absent' | 'undecided'> = {};
        const initialTimes: Record<string, { from: string; to: string }> = {};
        targetDatesList.forEach((td) => {
          initialResponses[td.target_date_id] = 'attending';
          initialTimes[td.target_date_id] = { from: '', to: '' };
        });
        setResponses(initialResponses);
        setAvailableTimes(initialTimes);
      } catch (err) {
        if (err instanceof PublicApiError) {
          if (err.isNotFound()) {
            setError('出欠確認が見つかりません。URLをご確認ください。');
          } else if (err.isForbidden()) {
            setError('この出欠確認は既に締め切られています。');
          } else {
            setError('データの取得に失敗しました。');
          }
        } else {
          setError('通信エラーが発生しました。');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [token]);

  const handleResponseChange = (targetDateId: string, response: 'attending' | 'absent' | 'undecided') => {
    setResponses((prev) => ({
      ...prev,
      [targetDateId]: response,
    }));
  };

  const handleTimeChange = (targetDateId: string, field: 'from' | 'to', value: string) => {
    setAvailableTimes((prev) => ({
      ...prev,
      [targetDateId]: {
        ...prev[targetDateId],
        [field]: value,
      },
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedMemberId) {
      alert('お名前を選択してください');
      return;
    }

    if (targetDates.length === 0) {
      alert('対象日が設定されていません');
      return;
    }

    if (!token) return;

    try {
      setSubmitting(true);
      setError(null);

      // 各対象日ごとに回答を送信
      const submitPromises = targetDates.map((td) => {
        const times = availableTimes[td.target_date_id];
        return submitAttendanceResponse(token, {
          member_id: selectedMemberId,
          target_date_id: td.target_date_id,
          response: responses[td.target_date_id] || 'attending',
          note,
          available_from: times?.from || undefined,
          available_to: times?.to || undefined,
        });
      });

      await Promise.all(submitPromises);

      setSubmitted(true);
    } catch (err) {
      if (err instanceof PublicApiError) {
        if (err.isNotFound()) {
          setError('出欠確認が見つかりません。');
        } else if (err.isBadRequest()) {
          setError('入力内容に誤りがあります。');
        } else if (err.isForbidden()) {
          setError('この出欠確認は既に締め切られています。');
        } else {
          setError('送信に失敗しました。');
        }
      } else {
        setError('通信エラーが発生しました。');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-accent"></div>
          <p className="mt-4 text-gray-600">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (error && !collection) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-6">
          <div className="text-center">
            <div className="text-red-500 text-5xl mb-4">⚠️</div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">エラー</h2>
            <p className="text-gray-600">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (submitted) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-6">
          <div className="text-center">
            <div className="text-green-500 text-5xl mb-4">✓</div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">送信完了</h2>
            <p className="text-gray-600 mb-4">
              出欠回答を受け付けました。
            </p>
            <button
              onClick={() => {
                setSubmitted(false);
                setSelectedMemberId('');
                const initialResponses: Record<string, 'attending' | 'absent' | 'undecided'> = {};
                const initialTimes: Record<string, { from: string; to: string }> = {};
                targetDates.forEach((td) => {
                  initialResponses[td.target_date_id] = 'attending';
                  initialTimes[td.target_date_id] = { from: '', to: '' };
                });
                setResponses(initialResponses);
                setAvailableTimes(initialTimes);
                setNote('');
              }}
              className="px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition"
            >
              別の回答を送信
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4">
      <div className="max-w-2xl mx-auto">
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {collection?.title || '(タイトルなし)'}
          </h1>
          {collection?.description && (
            <p className="text-gray-600 mb-4 whitespace-pre-wrap">
              {collection.description}
            </p>
          )}
          {collection?.deadline && (
            <p className="text-sm text-gray-500">
              締切: {new Date(collection.deadline).toLocaleString('ja-JP')}
            </p>
          )}
          {collection?.status === 'closed' && (
            <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <p className="text-yellow-800 text-sm">
                この出欠確認は締め切られています
              </p>
            </div>
          )}
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            出欠を回答する
          </h2>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-red-800 text-sm">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                お名前 <span className="text-red-500">*</span>
              </label>
              <select
                value={selectedMemberId}
                onChange={(e) => setSelectedMemberId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                required
                disabled={collection?.status === 'closed'}
              >
                <option value="">選択してください</option>
                {members.map((member) => (
                  <option key={member.member_id} value={member.member_id}>
                    {member.display_name}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                お名前が見つからない場合は、管理者にお問い合わせください
              </p>
            </div>

            {/* 各対象日ごとの出欠選択 */}
            {targetDates.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-3">
                  各日程の出欠 <span className="text-red-500">*</span>
                </label>
                <div className="space-y-4">
                  {targetDates.map((td) => (
                    <div
                      key={td.target_date_id}
                      className="border border-gray-200 rounded-md p-4 bg-gray-50"
                    >
                      <div className="font-medium text-gray-900 mb-2">
                        {new Date(td.target_date).toLocaleDateString('ja-JP', {
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric',
                          weekday: 'short',
                        })}
                        {(td.start_time || td.end_time) && (
                          <span className="ml-2 text-accent font-normal">
                            {td.start_time || ''}〜{td.end_time || ''}
                          </span>
                        )}
                      </div>
                      <div className="flex gap-4 flex-wrap">
                        <label className="flex items-center cursor-pointer">
                          <input
                            type="radio"
                            value="attending"
                            checked={responses[td.target_date_id] === 'attending'}
                            onChange={() => handleResponseChange(td.target_date_id, 'attending')}
                            className="mr-2"
                            disabled={collection?.status === 'closed'}
                          />
                          <span className="text-gray-700">参加</span>
                        </label>
                        <label className="flex items-center cursor-pointer">
                          <input
                            type="radio"
                            value="undecided"
                            checked={responses[td.target_date_id] === 'undecided'}
                            onChange={() => handleResponseChange(td.target_date_id, 'undecided')}
                            className="mr-2"
                            disabled={collection?.status === 'closed'}
                          />
                          <span className="text-gray-700">未定</span>
                        </label>
                        <label className="flex items-center cursor-pointer">
                          <input
                            type="radio"
                            value="absent"
                            checked={responses[td.target_date_id] === 'absent'}
                            onChange={() => handleResponseChange(td.target_date_id, 'absent')}
                            className="mr-2"
                            disabled={collection?.status === 'closed'}
                          />
                          <span className="text-gray-700">不参加</span>
                        </label>
                      </div>
                      {/* 参加または未定の場合に時間指定を表示 */}
                      {(responses[td.target_date_id] === 'attending' || responses[td.target_date_id] === 'undecided') && (
                        <div className="mt-3 pt-3 border-t border-gray-200">
                          <p className="text-sm text-gray-600 mb-2">参加可能な時間帯（任意）</p>
                          <div className="flex items-center gap-2 flex-wrap">
                            <input
                              type="time"
                              value={availableTimes[td.target_date_id]?.from || ''}
                              onChange={(e) => handleTimeChange(td.target_date_id, 'from', e.target.value)}
                              className="px-2 py-1 border border-gray-300 rounded-md text-sm"
                              disabled={collection?.status === 'closed'}
                            />
                            <span className="text-gray-500">〜</span>
                            <input
                              type="time"
                              value={availableTimes[td.target_date_id]?.to || ''}
                              onChange={(e) => handleTimeChange(td.target_date_id, 'to', e.target.value)}
                              className="px-2 py-1 border border-gray-300 rounded-md text-sm"
                              disabled={collection?.status === 'closed'}
                            />
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                備考（任意）
              </label>
              <textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                placeholder="補足事項があれば入力してください"
                disabled={collection?.status === 'closed'}
              />
            </div>

            <button
              type="submit"
              disabled={submitting || collection?.status === 'closed' || targetDates.length === 0}
              className="w-full px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {submitting ? '送信中...' : '回答を送信'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
