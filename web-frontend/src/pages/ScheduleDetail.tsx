import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getSchedule, getScheduleResponses, type Schedule, type ScheduleResponse } from '../lib/api/scheduleApi';
import { getMembers } from '../lib/api';
import type { Member } from '../types/api';

interface CandidateWithResponses {
  candidate_id: string;
  date: string;
  start_time?: string;
  end_time?: string;
  availableCount: number;
  maybeCount: number;
  unavailableCount: number;
  noResponseCount: number;
}

export default function ScheduleDetail() {
  const { scheduleId } = useParams<{ scheduleId: string }>();
  const [schedule, setSchedule] = useState<Schedule | null>(null);
  const [responses, setResponses] = useState<ScheduleResponse[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [candidatesWithResponses, setCandidatesWithResponses] = useState<CandidateWithResponses[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [publicUrl, setPublicUrl] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (scheduleId) {
      loadData();
    }
  }, [scheduleId]);

  const loadData = async () => {
    if (!scheduleId) return;

    try {
      setLoading(true);
      const [scheduleData, responsesData, membersData] = await Promise.all([
        getSchedule(scheduleId),
        getScheduleResponses(scheduleId),
        getMembers({ is_active: true }),
      ]);

      setSchedule(scheduleData);
      setResponses(responsesData);
      setMembers(membersData.members);

      // 公開URLを生成
      const baseUrl = window.location.origin;
      const url = `${baseUrl}/p/schedule/${scheduleData.public_token}`;
      setPublicUrl(url);

      // 候補日ごとの集計を作成
      if (scheduleData.candidates) {
        const candidatesMap = scheduleData.candidates.map((candidate: any) => {
          // この候補日への回答を集計
          const candidateResponses = responsesData.filter(
            (r) => r.candidate_id === candidate.candidate_id
          );

          const availableCount = candidateResponses.filter((r) => r.availability === 'available').length;
          const maybeCount = candidateResponses.filter((r) => r.availability === 'maybe').length;
          const unavailableCount = candidateResponses.filter((r) => r.availability === 'unavailable').length;

          // 回答済みメンバーのユニークIDを取得
          const respondedMemberIds = new Set(responsesData.map((r) => r.member_id));
          const noResponseCount = membersData.members.length - respondedMemberIds.size;

          return {
            candidate_id: candidate.candidate_id,
            date: candidate.date,
            start_time: candidate.start_time,
            end_time: candidate.end_time,
            availableCount,
            maybeCount,
            unavailableCount,
            noResponseCount,
          };
        });

        setCandidatesWithResponses(candidatesMap);
      }
    } catch (err) {
      console.error('Failed to load schedule detail:', err);
      setError('データの取得に失敗しました');
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

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'open':
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-green-100 text-green-800">受付中</span>;
      case 'decided':
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-blue-100 text-blue-800">決定済み</span>;
      case 'closed':
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-gray-100 text-gray-800">締切済み</span>;
      default:
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-gray-100 text-gray-800">{status}</span>;
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

  if (error || !schedule) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <p className="text-red-800">{error || '日程調整が見つかりません'}</p>
          <Link to="/schedules" className="text-blue-600 hover:underline mt-4 inline-block">
            ← 日程調整一覧に戻る
          </Link>
        </div>
      </div>
    );
  }

  // 回答済みメンバーのユニークIDを取得
  const respondedMemberIds = new Set(responses.map((r) => r.member_id));
  const responseCount = respondedMemberIds.size;
  const totalMembers = members.length;

  return (
    <div className="max-w-6xl mx-auto">
      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to="/schedules" className="hover:text-gray-900">
          日程調整一覧
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">{schedule.title}</span>
      </nav>

      {/* 基本情報 */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex justify-between items-start mb-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">{schedule.title}</h1>
            {schedule.description && (
              <p className="text-gray-600 mb-4">{schedule.description}</p>
            )}
          </div>
          {getStatusBadge(schedule.status)}
        </div>

        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">候補日数:</span>{' '}
            <span className="font-medium">{schedule.candidates?.length || 0}件</span>
          </div>
          <div>
            <span className="text-gray-500">回答数:</span>{' '}
            <span className="font-medium">
              {responseCount}/{totalMembers}人
            </span>
          </div>
          <div>
            <span className="text-gray-500">作成日:</span>{' '}
            <span className="font-medium">
              {new Date(schedule.created_at).toLocaleDateString('ja-JP')}
            </span>
          </div>
          {schedule.deadline && (
            <div>
              <span className="text-gray-500">締切:</span>{' '}
              <span className="font-medium">
                {new Date(schedule.deadline).toLocaleString('ja-JP', {
                  year: 'numeric',
                  month: '2-digit',
                  day: '2-digit',
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </span>
            </div>
          )}
        </div>

        {/* 公開URL */}
        <div className="mt-6 pt-6 border-t border-gray-200">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">公開URL</h3>
          <div className="flex gap-2">
            <input
              type="text"
              value={publicUrl}
              readOnly
              className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-md bg-gray-50 font-mono"
            />
            <button
              onClick={handleCopy}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition text-sm whitespace-nowrap"
            >
              {copied ? '✓ コピー済み' : 'URLをコピー'}
            </button>
            <a
              href={publicUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition text-sm whitespace-nowrap"
            >
              プレビュー
            </a>
          </div>
        </div>
      </div>

      {/* 候補日ごとの回答状況 */}
      <div className="bg-white rounded-lg shadow overflow-hidden mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">候補日ごとの回答状況</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  候補日
                </th>
                <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  ○ 参加可能
                </th>
                <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  △ 不確定
                </th>
                <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  × 参加不可
                </th>
                <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  - 未回答
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {candidatesWithResponses.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-gray-500">
                    候補日がありません
                  </td>
                </tr>
              ) : (
                candidatesWithResponses.map((candidate) => (
                  <tr key={candidate.candidate_id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">
                        {new Date(candidate.date).toLocaleString('ja-JP', {
                          year: 'numeric',
                          month: '2-digit',
                          day: '2-digit',
                          hour: '2-digit',
                          minute: '2-digit',
                          weekday: 'short',
                        })}
                      </div>
                      {candidate.start_time && candidate.end_time && (
                        <div className="text-xs text-gray-500">
                          {new Date(candidate.start_time).toLocaleTimeString('ja-JP', {
                            hour: '2-digit',
                            minute: '2-digit',
                          })}{' '}
                          〜{' '}
                          {new Date(candidate.end_time).toLocaleTimeString('ja-JP', {
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                        {candidate.availableCount}人
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800">
                        {candidate.maybeCount}人
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800">
                        {candidate.unavailableCount}人
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-600">
                        {candidate.noResponseCount}人
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* 個別回答一覧 */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">個別回答一覧</h2>
          <p className="text-sm text-gray-600 mt-1">
            {responseCount > 0 ? `${responseCount}人のメンバーが回答しました` : 'まだ回答がありません'}
          </p>
        </div>
        <div className="divide-y divide-gray-200">
          {responseCount === 0 ? (
            <div className="px-6 py-12 text-center text-gray-500">
              まだ回答がありません
            </div>
          ) : (
            // メンバーIDでグループ化して表示
            Array.from(respondedMemberIds).map((memberId) => {
              const memberResponses = responses.filter((r) => r.member_id === memberId);
              const member = members.find((m) => m.member_id === memberId);

              return (
                <div key={memberId} className="px-6 py-4">
                  <div className="font-medium text-gray-900 mb-2">
                    {member?.display_name || memberId}
                  </div>
                  <div className="space-y-1">
                    {memberResponses.map((response) => {
                      const candidate = schedule.candidates?.find(
                        (c: any) => c.candidate_id === response.candidate_id
                      );
                      if (!candidate) return null;

                      let statusBadge;
                      if (response.availability === 'available') {
                        statusBadge = (
                          <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800">
                            ○ 参加可能
                          </span>
                        );
                      } else if (response.availability === 'maybe') {
                        statusBadge = (
                          <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                            △ 不確定
                          </span>
                        );
                      } else {
                        statusBadge = (
                          <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-red-100 text-red-800">
                            × 参加不可
                          </span>
                        );
                      }

                      return (
                        <div
                          key={response.response_id}
                          className="flex items-center gap-3 text-sm text-gray-600"
                        >
                          {statusBadge}
                          <span>
                            {new Date(candidate.date).toLocaleString('ja-JP', {
                              month: '2-digit',
                              day: '2-digit',
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </span>
                          {response.note && (
                            <span className="text-gray-500">（{response.note}）</span>
                          )}
                        </div>
                      );
                    })}
                  </div>
                  <div className="mt-1 text-xs text-gray-400">
                    回答日時:{' '}
                    {new Date(memberResponses[0].responded_at).toLocaleString('ja-JP')}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
