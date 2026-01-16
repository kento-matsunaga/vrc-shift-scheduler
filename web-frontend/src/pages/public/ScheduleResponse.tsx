import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  getScheduleByToken,
  getMembers,
  submitScheduleResponse,
  getAllScheduleResponses,
  type DateSchedule,
  type Member,
  type ScheduleResponseInput,
  type PublicScheduleResponse,
  PublicApiError,
} from '../../lib/api/publicApi';
import SearchableSelect from '../../components/SearchableSelect';
import ScheduleResponseTable from '../../components/ScheduleResponseTable';

export default function ScheduleResponse() {
  const { token } = useParams<{ token: string }>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [schedule, setSchedule] = useState<DateSchedule | null>(null);
  const [members, setMembers] = useState<Member[]>([]);

  // フォーム状態
  const [selectedMemberId, setSelectedMemberId] = useState('');
  const [responses, setResponses] = useState<Record<string, { availability: 'available' | 'unavailable' | 'maybe'; note: string }>>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  // 全回答一覧（調整さん形式表示用）
  const [allResponses, setAllResponses] = useState<PublicScheduleResponse[]>([]);
  const [loadingAllResponses, setLoadingAllResponses] = useState(false);

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

        // 日程調整情報を取得
        const scheduleData = await getScheduleByToken(token);
        setSchedule(scheduleData);

        // メンバー一覧を取得（グループでフィルタリング）
        const membersData = await getMembers(scheduleData.tenant_id, scheduleData.group_ids);
        setMembers(membersData.data?.members || []);

        // 初期値設定（全候補に対してmaybeを設定）
        const initialResponses: Record<string, { availability: 'available' | 'unavailable' | 'maybe'; note: string }> = {};
        scheduleData.candidates.forEach((candidate) => {
          initialResponses[candidate.candidate_id] = {
            availability: 'maybe',
            note: '',
          };
        });
        setResponses(initialResponses);

        // 全回答一覧を取得
        setLoadingAllResponses(true);
        try {
          const allResponsesData = await getAllScheduleResponses(token);
          setAllResponses(allResponsesData.responses || []);
        } catch {
          // 回答一覧の取得に失敗してもエラー表示はしない
          console.warn('Failed to load all responses');
        } finally {
          setLoadingAllResponses(false);
        }
      } catch (err) {
        if (err instanceof PublicApiError) {
          if (err.isNotFound()) {
            setError('日程調整が見つかりません。URLをご確認ください。');
          } else if (err.isForbidden()) {
            setError('この日程調整は既に締め切られています。');
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

  const updateResponse = (candidateId: string, availability: 'available' | 'unavailable' | 'maybe', note?: string) => {
    setResponses((prev) => ({
      ...prev,
      [candidateId]: {
        availability,
        note: note !== undefined ? note : prev[candidateId]?.note || '',
      },
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedMemberId) {
      alert('お名前を選択してください');
      return;
    }

    if (!token) return;

    try {
      setSubmitting(true);
      setError(null);

      // レスポンスを配列に変換
      const responseArray: ScheduleResponseInput[] = Object.entries(responses).map(
        ([candidateId, data]) => ({
          candidate_id: candidateId,
          availability: data.availability,
          note: data.note || '',
        })
      );

      const requestData = {
        member_id: selectedMemberId,
        responses: responseArray,
      };

      await submitScheduleResponse(token, requestData);

      // 送信成功後に全回答を再取得
      try {
        const allResponsesData = await getAllScheduleResponses(token);
        setAllResponses(allResponsesData.responses || []);
      } catch {
        // 回答一覧の取得に失敗してもエラー表示はしない
      }

      setSubmitted(true);
    } catch (err) {
      if (err instanceof PublicApiError) {
        if (err.isNotFound()) {
          setError('日程調整が見つかりません。');
        } else if (err.isBadRequest()) {
          setError('入力内容に誤りがあります。');
        } else if (err.isForbidden()) {
          setError('この日程調整は既に締め切られています。');
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

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      weekday: 'short',
    });
  };

  const formatTime = (timeStr?: string) => {
    if (!timeStr) return '';
    // ISO 8601形式（例: "0001-01-01T14:30:00Z"）からHH:MM形式を抽出
    const date = new Date(timeStr);
    if (!isNaN(date.getTime())) {
      return date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit', hour12: false });
    }
    // フォールバック: HH:MM:SS形式の場合
    return timeStr.substring(0, 5);
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

  if (error && !schedule) {
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
              日程回答を受け付けました。
            </p>
            <button
              onClick={() => {
                setSubmitted(false);
                setSelectedMemberId('');
                // 初期値にリセット
                const initialResponses: Record<string, { availability: 'available' | 'unavailable' | 'maybe'; note: string }> = {};
                schedule?.candidates.forEach((candidate) => {
                  initialResponses[candidate.candidate_id] = {
                    availability: 'maybe',
                    note: '',
                  };
                });
                setResponses(initialResponses);
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
      <div className="max-w-3xl mx-auto">
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {schedule?.title}
          </h1>
          {schedule?.description && (
            <p className="text-gray-600 mb-4 whitespace-pre-wrap">
              {schedule.description}
            </p>
          )}
          {schedule?.deadline && (
            <p className="text-sm text-gray-500">
              締切: {new Date(schedule.deadline).toLocaleString('ja-JP')}
            </p>
          )}
          {schedule?.status === 'closed' && (
            <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <p className="text-yellow-800 text-sm">
                この日程調整は締め切られています
              </p>
            </div>
          )}
          {schedule?.status === 'decided' && schedule.decided_candidate_id && (
            <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-md">
              <p className="text-green-800 text-sm font-semibold">
                日程が確定しました
              </p>
            </div>
          )}
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            日程の都合を回答する
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
              <SearchableSelect
                options={members.map((member) => ({
                  value: member.member_id,
                  label: member.display_name,
                }))}
                value={selectedMemberId}
                onChange={setSelectedMemberId}
                placeholder="名前を検索して選択..."
                disabled={schedule?.status !== 'open'}
              />
              <p className="mt-1 text-xs text-gray-500">
                お名前が見つからない場合は、管理者にお問い合わせください
              </p>
            </div>

            <div className="border-t pt-4">
              <h3 className="text-md font-medium text-gray-900 mb-3">
                候補日ごとの都合
              </h3>
              <div className="space-y-4">
                {schedule?.candidates.map((candidate) => {
                  const isDecided = schedule.decided_candidate_id === candidate.candidate_id;
                  return (
                    <div
                      key={candidate.candidate_id}
                      className={`p-4 border rounded-md ${
                        isDecided ? 'border-green-500 bg-green-50' : 'border-gray-200'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <div>
                          <p className="font-medium text-gray-900">
                            {formatDate(candidate.date)}
                            {isDecided && (
                              <span className="ml-2 text-xs bg-green-600 text-white px-2 py-1 rounded">
                                確定
                              </span>
                            )}
                          </p>
                          {(candidate.start_time || candidate.end_time) && (
                            <p className="text-sm text-gray-600">
                              {formatTime(candidate.start_time)} 〜 {formatTime(candidate.end_time)}
                            </p>
                          )}
                        </div>
                      </div>

                      <div className="space-y-2">
                        <label className="flex items-center">
                          <input
                            type="radio"
                            name={`candidate-${candidate.candidate_id}`}
                            value="available"
                            checked={responses[candidate.candidate_id]?.availability === 'available'}
                            onChange={() => updateResponse(candidate.candidate_id, 'available')}
                            className="mr-2"
                            disabled={schedule?.status !== 'open'}
                          />
                          <span className="text-sm text-green-700">⭕ 参加できる</span>
                        </label>
                        <label className="flex items-center">
                          <input
                            type="radio"
                            name={`candidate-${candidate.candidate_id}`}
                            value="unavailable"
                            checked={responses[candidate.candidate_id]?.availability === 'unavailable'}
                            onChange={() => updateResponse(candidate.candidate_id, 'unavailable')}
                            className="mr-2"
                            disabled={schedule?.status !== 'open'}
                          />
                          <span className="text-sm text-red-700">❌ 参加できない</span>
                        </label>
                        <label className="flex items-center">
                          <input
                            type="radio"
                            name={`candidate-${candidate.candidate_id}`}
                            value="maybe"
                            checked={responses[candidate.candidate_id]?.availability === 'maybe'}
                            onChange={() => updateResponse(candidate.candidate_id, 'maybe')}
                            className="mr-2"
                            disabled={schedule?.status !== 'open'}
                          />
                          <span className="text-sm text-gray-700">△ 未定・要相談</span>
                        </label>
                      </div>

                      <div className="mt-2">
                        <input
                          type="text"
                          placeholder="備考（任意）"
                          value={responses[candidate.candidate_id]?.note || ''}
                          onChange={(e) =>
                            updateResponse(
                              candidate.candidate_id,
                              responses[candidate.candidate_id]?.availability || 'maybe',
                              e.target.value
                            )
                          }
                          className="w-full px-2 py-1 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-accent"
                          disabled={schedule?.status !== 'open'}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            <button
              type="submit"
              disabled={submitting || schedule?.status !== 'open'}
              className="w-full px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {submitting ? '送信中...' : '回答を送信'}
            </button>
          </form>
        </div>

        {/* 回答一覧（調整さん形式） */}
        <div className="bg-white rounded-lg shadow-md p-6 mt-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            回答一覧
          </h2>
          {loadingAllResponses ? (
            <div className="flex items-center justify-center py-8">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-accent"></div>
              <span className="ml-3 text-gray-600">読み込み中...</span>
            </div>
          ) : (
            <ScheduleResponseTable candidates={schedule?.candidates || []} responses={allResponses} />
          )}
        </div>
      </div>
    </div>
  );
}
