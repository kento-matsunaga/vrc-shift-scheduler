import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  getAttendanceCollection,
  getAttendanceResponses,
  closeAttendanceCollection,
  type AttendanceCollection as AttendanceCollectionType,
  type AttendanceResponse,
} from '../lib/api/attendanceApi';
import { getMembers } from '../lib/api';
import { getMemberGroups, getMemberGroupDetail, type MemberGroup } from '../lib/api/memberGroupApi';
import type { Member } from '../types/api';

export default function AttendanceDetail() {
  const { collectionId } = useParams<{ collectionId: string }>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [collection, setCollection] = useState<AttendanceCollectionType | null>(null);
  const [responses, setResponses] = useState<AttendanceResponse[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [appliedGroups, setAppliedGroups] = useState<MemberGroup[]>([]);
  const [closing, setClosing] = useState(false);
  const [publicUrl, setPublicUrl] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!collectionId) {
      setError('出欠確認IDが指定されていません');
      setLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setLoading(true);
        const [collectionData, responsesData, membersData, allGroups] = await Promise.all([
          getAttendanceCollection(collectionId),
          getAttendanceResponses(collectionId),
          getMembers({ is_active: true }),
          getMemberGroups(),
        ]);
        setCollection(collectionData);
        setResponses(responsesData || []);

        // グループIDが設定されている場合、そのグループに属するメンバーのみを表示
        const groupIds = collectionData.group_ids || [];
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
        const url = `${baseUrl}/p/attendance/${collectionData.public_token}`;
        setPublicUrl(url);
      } catch (err) {
        setError(err instanceof Error ? err.message : '取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [collectionId]);

  const handleClose = async () => {
    if (!collectionId) return;
    if (!confirm('この出欠確認を締め切りますか？締め切り後は回答を受け付けなくなります。')) return;

    try {
      setClosing(true);
      await closeAttendanceCollection(collectionId);
      alert('出欠確認を締め切りました');
      const collectionData = await getAttendanceCollection(collectionId);
      setCollection(collectionData);
    } catch (err) {
      alert(err instanceof Error ? err.message : '締切に失敗しました');
    } finally {
      setClosing(false);
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
      case 'closed':
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-gray-100 text-gray-800">締切済み</span>;
      default:
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-gray-100 text-gray-800">{status}</span>;
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

  if (error || !collection) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <p className="text-red-800">{error || '出欠確認が見つかりません'}</p>
          <Link to="/attendance" className="text-accent hover:underline mt-4 inline-block">
            ← 出欠確認一覧に戻る
          </Link>
        </div>
      </div>
    );
  }

  // Sort target dates by display_order
  const sortedTargetDates = (collection.target_dates || []).sort((a, b) => a.display_order - b.display_order);

  // Get unique member IDs who responded
  const respondedMemberIds = new Set(responses.map((r) => r.member_id));
  const responseCount = respondedMemberIds.size;
  const totalMembers = members.length;

  // Create response map for quick lookup: member_id -> target_date_id -> response
  const responseMap = new Map<string, Map<string, 'attending' | 'absent' | 'undecided'>>();
  responses.forEach((resp) => {
    if (!responseMap.has(resp.member_id)) {
      responseMap.set(resp.member_id, new Map());
    }
    responseMap.get(resp.member_id)!.set(resp.target_date_id, resp.response);
  });

  // Create time map for quick lookup: member_id -> target_date_id -> { from, to }
  const timeMap = new Map<string, Map<string, { from?: string; to?: string }>>();
  responses.forEach((resp) => {
    if (resp.available_from || resp.available_to) {
      if (!timeMap.has(resp.member_id)) {
        timeMap.set(resp.member_id, new Map());
      }
      timeMap.get(resp.member_id)!.set(resp.target_date_id, {
        from: resp.available_from,
        to: resp.available_to,
      });
    }
  });

  // Calculate stats for each target date
  const dateStats = sortedTargetDates.map((targetDate) => {
    const attendingCount = responses.filter(
      (r) => r.target_date_id === targetDate.target_date_id && r.response === 'attending'
    ).length;
    const undecidedCount = responses.filter(
      (r) => r.target_date_id === targetDate.target_date_id && r.response === 'undecided'
    ).length;
    const absentCount = responses.filter(
      (r) => r.target_date_id === targetDate.target_date_id && r.response === 'absent'
    ).length;
    const noResponseCount = totalMembers - respondedMemberIds.size;

    return {
      targetDateId: targetDate.target_date_id,
      attendingCount,
      undecidedCount,
      absentCount,
      noResponseCount,
    };
  });

  return (
    <div className="max-w-7xl mx-auto">
      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to="/attendance" className="hover:text-gray-900">
          出欠確認一覧
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">{collection.title}</span>
      </nav>

      {/* 基本情報 */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex justify-between items-start mb-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">{collection.title}</h1>
            {collection.description && (
              <p className="text-gray-600 mb-4">{collection.description}</p>
            )}
          </div>
          <div className="flex gap-2">
            {getStatusBadge(collection.status)}
            {collection.status === 'open' && (
              <button
                onClick={handleClose}
                disabled={closing}
                className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition disabled:bg-gray-400 text-sm"
              >
                {closing ? '処理中...' : '締め切る'}
              </button>
            )}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 text-sm mb-4">
          <div>
            <span className="text-gray-500">対象日数:</span>{' '}
            <span className="font-medium">{sortedTargetDates.length}件</span>
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
              {new Date(collection.created_at).toLocaleDateString('ja-JP')}
            </span>
          </div>
          {collection.deadline && (
            <div>
              <span className="text-gray-500">締切:</span>{' '}
              <span className="font-medium">
                {new Date(collection.deadline).toLocaleString('ja-JP', {
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
              className="px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition text-sm whitespace-nowrap"
            >
              {copied ? '✓ コピー済み' : 'URLをコピー'}
            </button>
          </div>
        </div>
      </div>

      {/* 出欠確認表（調整さん形式） */}
      <div className="bg-white rounded-lg shadow overflow-hidden mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">出欠確認状況</h2>
          <p className="text-sm text-gray-600 mt-1">
            ○: 参加、△: 未定、×: 不参加、-: 未回答
          </p>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 z-10">
                  メンバー
                </th>
                {sortedTargetDates.map((targetDate) => (
                  <th
                    key={targetDate.target_date_id}
                    className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider min-w-[100px]"
                  >
                    <div>
                      {new Date(targetDate.target_date).toLocaleDateString('ja-JP', {
                        month: '2-digit',
                        day: '2-digit',
                      })}
                    </div>
                    <div className="text-xs font-normal normal-case text-gray-400">
                      {new Date(targetDate.target_date).toLocaleDateString('ja-JP', {
                        weekday: 'short',
                      })}
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {members.length === 0 ? (
                <tr>
                  <td colSpan={sortedTargetDates.length + 1} className="px-6 py-12 text-center text-gray-500">
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
                      {sortedTargetDates.map((targetDate) => {
                        const response = memberResponses?.get(targetDate.target_date_id);
                        const memberTimes = timeMap.get(member.member_id);
                        const times = memberTimes?.get(targetDate.target_date_id);
                        let content;
                        let bgColor;

                        if (response === 'attending') {
                          content = '○';
                          bgColor = 'bg-green-50 text-green-800';
                        } else if (response === 'undecided') {
                          content = '△';
                          bgColor = 'bg-yellow-50 text-yellow-800';
                        } else if (response === 'absent') {
                          content = '×';
                          bgColor = 'bg-red-50 text-red-800';
                        } else {
                          content = '-';
                          bgColor = 'bg-gray-50 text-gray-400';
                        }

                        // Format time display
                        const timeDisplay = times && (times.from || times.to)
                          ? `${times.from || '?'}〜${times.to || '?'}`
                          : null;

                        return (
                          <td
                            key={targetDate.target_date_id}
                            className={`px-4 py-4 text-center ${bgColor}`}
                            title={timeDisplay || undefined}
                          >
                            <div className="text-lg font-semibold">{content}</div>
                            {timeDisplay && (
                              <div className="text-xs text-gray-600 mt-1">{timeDisplay}</div>
                            )}
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
                {sortedTargetDates.map((targetDate) => {
                  const stats = dateStats.find((s) => s.targetDateId === targetDate.target_date_id);
                  return (
                    <td key={targetDate.target_date_id} className="px-4 py-3 text-center">
                      <div className="text-xs space-y-1">
                        <div className="text-green-700">
                          ○ {stats?.attendingCount || 0}
                        </div>
                        <div className="text-yellow-700">
                          △ {stats?.undecidedCount || 0}
                        </div>
                        <div className="text-red-700">
                          × {stats?.absentCount || 0}
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
