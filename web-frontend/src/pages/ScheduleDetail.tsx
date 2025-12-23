import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getSchedule, getScheduleResponses, type Schedule, type ScheduleResponse } from '../lib/api/scheduleApi';
import { getMembers } from '../lib/api';
import { getMemberGroups, getMemberGroupDetail, type MemberGroup } from '../lib/api/memberGroupApi';
import type { Member } from '../types/api';

interface Candidate {
  candidate_id: string;
  date: string;
  start_time?: string;
  end_time?: string;
}

export default function ScheduleDetail() {
  const { scheduleId } = useParams<{ scheduleId: string }>();
  const [schedule, setSchedule] = useState<Schedule | null>(null);
  const [responses, setResponses] = useState<ScheduleResponse[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [appliedGroups, setAppliedGroups] = useState<MemberGroup[]>([]);
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
      const [scheduleData, responsesData, membersData, allGroups] = await Promise.all([
        getSchedule(scheduleId),
        getScheduleResponses(scheduleId),
        getMembers({ is_active: true }),
        getMemberGroups(),
      ]);

      setSchedule(scheduleData);
      setResponses(responsesData || []);

      // グループIDが設定されている場合、そのグループに属するメンバーのみを表示
      const groupIds = scheduleData.group_ids || [];
      if (groupIds.length > 0) {
        // 適用グループ情報を取得
        const groups = (allGroups.groups || []).filter((g: MemberGroup) => groupIds.includes(g.group_id));
        setAppliedGroups(groups);

        // グループに属するメンバーIDを集める
        const allowedMemberIds = new Set<string>();
        for (const groupId of groupIds) {
          try {
            const groupDetail = await getMemberGroupDetail(groupId);
            (groupDetail.member_ids || []).forEach((memberId: string) => allowedMemberIds.add(memberId));
          } catch (e) {
            console.error('Failed to fetch group members:', e);
          }
        }

        // フィルタリング
        const filteredMembers = (membersData.members || []).filter((m: Member) =>
          allowedMemberIds.has(m.member_id)
        );
        setMembers(filteredMembers);
      } else {
        setAppliedGroups([]);
        setMembers(membersData.members || []);
      }

      const baseUrl = window.location.origin;
      const url = `${baseUrl}/p/schedule/${scheduleData.public_token}`;
      setPublicUrl(url);
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
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-indigo-100 text-indigo-800">決定済み</span>;
      case 'closed':
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-gray-100 text-gray-800">締切済み</span>;
      default:
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-gray-100 text-gray-800">{status}</span>;
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  if (error || !schedule) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <p className="text-red-800">{error || '日程調整が見つかりません'}</p>
          <Link to="/schedules" className="text-indigo-600 hover:underline mt-4 inline-block">
            ← 日程調整一覧に戻る
          </Link>
        </div>
      </div>
    );
  }

  const candidates = (schedule.candidates || []) as Candidate[];

  // Get unique member IDs who responded
  const respondedMemberIds = new Set(responses.map((r) => r.member_id));
  const responseCount = respondedMemberIds.size;
  const totalMembers = members.length;

  // Create response map for quick lookup: member_id -> candidate_id -> availability
  const responseMap = new Map<string, Map<string, 'available' | 'maybe' | 'unavailable'>>();
  responses.forEach((resp) => {
    if (!responseMap.has(resp.member_id)) {
      responseMap.set(resp.member_id, new Map());
    }
    responseMap.get(resp.member_id)!.set(resp.candidate_id, resp.availability);
  });

  // Calculate stats for each candidate
  const candidateStats = candidates.map((candidate) => {
    const availableCount = responses.filter(
      (r) => r.candidate_id === candidate.candidate_id && r.availability === 'available'
    ).length;
    const maybeCount = responses.filter(
      (r) => r.candidate_id === candidate.candidate_id && r.availability === 'maybe'
    ).length;
    const unavailableCount = responses.filter(
      (r) => r.candidate_id === candidate.candidate_id && r.availability === 'unavailable'
    ).length;
    const noResponseCount = totalMembers - respondedMemberIds.size;

    return {
      candidateId: candidate.candidate_id,
      availableCount,
      maybeCount,
      unavailableCount,
      noResponseCount,
    };
  });

  return (
    <div className="max-w-7xl mx-auto">
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

        <div className="grid grid-cols-2 gap-4 text-sm mb-4">
          <div>
            <span className="text-gray-500">候補日数:</span>{' '}
            <span className="font-medium">{candidates.length}件</span>
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

        {/* 対象グループ */}
        {appliedGroups.length > 0 && (
          <div className="pt-4 border-t border-gray-200 mb-4">
            <h3 className="text-sm font-semibold text-gray-900 mb-2">対象グループ</h3>
            <div className="flex flex-wrap gap-2">
              {appliedGroups.map((group) => (
                <span
                  key={group.group_id}
                  className="px-3 py-1 rounded-full text-sm font-medium text-white"
                  style={{ backgroundColor: group.color || '#6366f1' }}
                >
                  {group.name}
                </span>
              ))}
            </div>
            <p className="text-xs text-gray-500 mt-2">
              上記グループに属するメンバーのみが回答対象です
            </p>
          </div>
        )}

        {/* 公開URL */}
        <div className="pt-4 border-t border-gray-200">
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
              className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition text-sm whitespace-nowrap"
            >
              {copied ? '✓ コピー済み' : 'URLをコピー'}
            </button>
          </div>
        </div>
      </div>

      {/* 日程調整表（調整さん形式） */}
      <div className="bg-white rounded-lg shadow overflow-hidden mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">回答状況</h2>
          <p className="text-sm text-gray-600 mt-1">
            ○: 参加可能、△: 不確定、×: 参加不可、-: 未回答
          </p>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 z-10">
                  メンバー
                </th>
                {candidates.map((candidate) => (
                  <th
                    key={candidate.candidate_id}
                    className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider min-w-[120px]"
                  >
                    <div>
                      {new Date(candidate.date).toLocaleDateString('ja-JP', {
                        month: '2-digit',
                        day: '2-digit',
                      })}
                    </div>
                    <div className="text-xs font-normal normal-case text-gray-400">
                      {new Date(candidate.date).toLocaleDateString('ja-JP', {
                        weekday: 'short',
                      })}
                    </div>
                    {candidate.start_time && candidate.end_time && (
                      <div className="text-xs font-normal normal-case text-gray-400 mt-1">
                        {new Date(candidate.start_time).toLocaleTimeString('ja-JP', {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                        -
                        {new Date(candidate.end_time).toLocaleTimeString('ja-JP', {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
                    )}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {members.length === 0 ? (
                <tr>
                  <td colSpan={candidates.length + 1} className="px-6 py-12 text-center text-gray-500">
                    メンバーがいません
                  </td>
                </tr>
              ) : (
                members.map((member) => {
                  const memberResponses = responseMap.get(member.member_id);
                  return (
                    <tr key={member.member_id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 sticky left-0 bg-white">
                        {member.display_name}
                      </td>
                      {candidates.map((candidate) => {
                        const availability = memberResponses?.get(candidate.candidate_id);
                        let content;
                        let bgColor;

                        if (availability === 'available') {
                          content = '○';
                          bgColor = 'bg-green-50 text-green-800';
                        } else if (availability === 'maybe') {
                          content = '△';
                          bgColor = 'bg-yellow-50 text-yellow-800';
                        } else if (availability === 'unavailable') {
                          content = '×';
                          bgColor = 'bg-red-50 text-red-800';
                        } else {
                          content = '-';
                          bgColor = 'bg-gray-50 text-gray-400';
                        }

                        return (
                          <td
                            key={candidate.candidate_id}
                            className={`px-4 py-4 text-center text-lg font-semibold ${bgColor}`}
                          >
                            {content}
                          </td>
                        );
                      })}
                    </tr>
                  );
                })
              )}
            </tbody>
            <tfoot className="bg-gray-50">
              <tr>
                <td className="px-6 py-3 text-sm font-medium text-gray-700 sticky left-0 bg-gray-50">
                  集計
                </td>
                {candidates.map((candidate) => {
                  const stats = candidateStats.find((s) => s.candidateId === candidate.candidate_id);
                  return (
                    <td key={candidate.candidate_id} className="px-4 py-3 text-center">
                      <div className="text-xs space-y-1">
                        <div className="text-green-700">
                          ○ {stats?.availableCount || 0}
                        </div>
                        <div className="text-yellow-700">
                          △ {stats?.maybeCount || 0}
                        </div>
                        <div className="text-red-700">
                          × {stats?.unavailableCount || 0}
                        </div>
                        <div className="text-gray-500">
                          - {stats?.noResponseCount || 0}
                        </div>
                      </div>
                    </td>
                  );
                })}
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>
  );
}
