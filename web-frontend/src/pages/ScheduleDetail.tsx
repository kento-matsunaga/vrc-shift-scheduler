import { useState, useEffect } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { getSchedule, getScheduleResponses, deleteSchedule, convertToAttendance, type Schedule, type ScheduleResponse } from '../lib/api/scheduleApi';
import { getMembers } from '../lib/api';
import { getMemberGroups, getMemberGroupDetail, type MemberGroup } from '../lib/api/memberGroupApi';
import { ApiClientError } from '../lib/apiClient';
import type { Member } from '../types/api';
import { formatTimeRange } from '../lib/timeUtils';

interface Candidate {
  candidate_id: string;
  date: string;
  start_time?: string;
  end_time?: string;
}

// ソートの種類
type SortKey = 'name' | 'available_count' | 'date_available';
type SortDirection = 'asc' | 'desc';

// formatTime関数は共通ユーティリティ（lib/timeUtils.ts）に移行済み

export default function ScheduleDetail() {
  const { scheduleId } = useParams<{ scheduleId: string }>();
  const navigate = useNavigate();
  const [schedule, setSchedule] = useState<Schedule | null>(null);
  const [responses, setResponses] = useState<ScheduleResponse[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [appliedGroups, setAppliedGroups] = useState<MemberGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [publicUrl, setPublicUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const [deleting, setDeleting] = useState(false);

  // 出欠確認変換モーダル状態
  const [showConvertModal, setShowConvertModal] = useState(false);
  const [selectedCandidateIds, setSelectedCandidateIds] = useState<string[]>([]);
  const [convertTitle, setConvertTitle] = useState('');
  const [converting, setConverting] = useState(false);

  // ソート状態
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [sortCandidateId, setSortCandidateId] = useState<string | null>(null);

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

  const handleDelete = async () => {
    if (!scheduleId) return;
    if (!confirm('この日程調整を削除しますか？この操作は取り消せません。')) return;

    try {
      setDeleting(true);
      await deleteSchedule(scheduleId);
      alert('日程調整を削除しました');
      navigate('/schedules');
    } catch (err) {
      if (err instanceof ApiClientError) {
        alert(err.getUserMessage());
      } else {
        alert(err instanceof Error ? err.message : '削除に失敗しました');
      }
    } finally {
      setDeleting(false);
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

  const handleOpenConvertModal = () => {
    setSelectedCandidateIds([]);
    setConvertTitle(schedule?.title || '');
    setShowConvertModal(true);
  };

  const handleToggleCandidate = (candidateId: string) => {
    setSelectedCandidateIds((prev) =>
      prev.includes(candidateId)
        ? prev.filter((id) => id !== candidateId)
        : [...prev, candidateId]
    );
  };

  const handleConvertToAttendance = async () => {
    if (!scheduleId || selectedCandidateIds.length === 0) return;

    try {
      setConverting(true);
      const result = await convertToAttendance(scheduleId, {
        candidate_ids: selectedCandidateIds,
        title: convertTitle || undefined,
      });
      setShowConvertModal(false);
      alert('出欠確認に変換しました');
      navigate(`/attendance/${result.collection_id}`);
    } catch (err) {
      if (err instanceof ApiClientError) {
        alert(err.getUserMessage());
      } else {
        alert(err instanceof Error ? err.message : '変換に失敗しました');
      }
    } finally {
      setConverting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'open':
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-green-100 text-green-800">受付中</span>;
      case 'decided':
        return <span className="px-3 py-1 text-sm font-semibold rounded-full bg-accent/10 text-accent-dark">決定済み</span>;
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

  if (error || !schedule) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <p className="text-red-800">{error || '日程調整が見つかりません'}</p>
          <Link to="/schedules" className="text-accent hover:underline mt-4 inline-block">
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

  // ソート・フィルタリング処理
  const sortedMembers = [...members].sort((a, b) => {
    let comparison = 0;

    if (sortKey === 'name') {
      // 名前でソート（日本語対応）
      comparison = a.display_name.localeCompare(b.display_name, 'ja');
    } else if (sortKey === 'available_count') {
      // 全体の参加可能数でソート
      const aCount = candidates.filter(
        (c) => responseMap.get(a.member_id)?.get(c.candidate_id) === 'available'
      ).length;
      const bCount = candidates.filter(
        (c) => responseMap.get(b.member_id)?.get(c.candidate_id) === 'available'
      ).length;
      comparison = aCount - bCount;
    } else if (sortKey === 'date_available' && sortCandidateId) {
      // 特定の日付の参加可能状態でソート
      const aAvailability = responseMap.get(a.member_id)?.get(sortCandidateId);
      const bAvailability = responseMap.get(b.member_id)?.get(sortCandidateId);
      const order = { available: 0, maybe: 1, unavailable: 2, undefined: 3 };
      const aOrder = aAvailability ? order[aAvailability] : order.undefined;
      const bOrder = bAvailability ? order[bAvailability] : order.undefined;
      comparison = aOrder - bOrder;
    }

    return sortDirection === 'asc' ? comparison : -comparison;
  });

  // ソートハンドラ
  const handleSort = (key: SortKey, candidateId?: string) => {
    if (key === 'date_available' && candidateId) {
      if (sortKey === 'date_available' && sortCandidateId === candidateId) {
        setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
      } else {
        setSortKey('date_available');
        setSortCandidateId(candidateId);
        setSortDirection('asc');
      }
    } else if (key === sortKey && key !== 'date_available') {
      setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortCandidateId(null);
      setSortDirection('asc');
    }
  };

  // ソートアイコン
  const SortIcon = ({ active, direction }: { active: boolean; direction: SortDirection }) => (
    <span className={`ml-1 inline-block ${active ? 'text-accent' : 'text-gray-400'}`}>
      {direction === 'asc' ? '↑' : '↓'}
    </span>
  );

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
          <div className="flex gap-2 items-center">
            {getStatusBadge(schedule.status)}
            <button
              onClick={handleOpenConvertModal}
              className="px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition text-sm"
            >
              出欠確認に変換
            </button>
            <button
              onClick={handleDelete}
              disabled={deleting}
              className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition disabled:bg-red-400 text-sm"
            >
              {deleting ? '削除中...' : '削除'}
            </button>
          </div>
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
              className="px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition text-sm whitespace-nowrap"
            >
              {copied ? '✓ コピー済み' : 'URLをコピー'}
            </button>
          </div>
        </div>
      </div>

      {/* 日程調整表（調整さん形式） */}
      <div className="bg-white rounded-lg shadow overflow-hidden mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">回答状況</h2>
              <p className="text-sm text-gray-600 mt-1">
                ○: 参加可能、△: 不確定、×: 参加不可、-: 未回答
              </p>
            </div>
            {/* ソートコントロール */}
            <div className="flex flex-wrap items-center gap-2">
              <select
                value={sortKey === 'date_available' ? `date_${sortCandidateId}` : sortKey}
                onChange={(e) => {
                  const value = e.target.value;
                  if (value === 'name') {
                    handleSort('name');
                  } else if (value === 'available_count') {
                    handleSort('available_count');
                  } else if (value.startsWith('date_')) {
                    handleSort('date_available', value.replace('date_', ''));
                  }
                }}
                className="px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
              >
                <option value="name">名前順</option>
                <option value="available_count">参加可能数順</option>
                {candidates.map((c) => (
                  <option key={c.candidate_id} value={`date_${c.candidate_id}`}>
                    {new Date(c.date).toLocaleDateString('ja-JP', { month: '2-digit', day: '2-digit' })}の参加状況
                  </option>
                ))}
              </select>
              <button
                onClick={() => setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'))}
                className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-accent"
                title={sortDirection === 'asc' ? '昇順' : '降順'}
              >
                {sortDirection === 'asc' ? '↑ 昇順' : '↓ 降順'}
              </button>
            </div>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 z-10 cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('name')}
                >
                  <span className="flex items-center">
                    メンバー
                    <SortIcon active={sortKey === 'name'} direction={sortDirection} />
                  </span>
                </th>
                {candidates.map((candidate) => (
                  <th
                    key={candidate.candidate_id}
                    className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider min-w-[120px] cursor-pointer hover:bg-gray-100"
                    onClick={() => handleSort('date_available', candidate.candidate_id)}
                  >
                    <div className="flex flex-col items-center">
                      <span className="flex items-center">
                        {new Date(candidate.date).toLocaleDateString('ja-JP', {
                          month: '2-digit',
                          day: '2-digit',
                        })}
                        {sortKey === 'date_available' && sortCandidateId === candidate.candidate_id && (
                          <SortIcon active={true} direction={sortDirection} />
                        )}
                      </span>
                      <span className="text-xs font-normal normal-case text-gray-400">
                        {new Date(candidate.date).toLocaleDateString('ja-JP', {
                          weekday: 'short',
                        })}
                      </span>
                      {(candidate.start_time || candidate.end_time) && (
                        <span className="text-xs font-normal normal-case text-gray-400 mt-1">
                          {formatTimeRange(candidate.start_time, candidate.end_time, '-')}
                        </span>
                      )}
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {sortedMembers.length === 0 ? (
                <tr>
                  <td colSpan={candidates.length + 1} className="px-6 py-12 text-center text-gray-500">
                    メンバーがいません
                  </td>
                </tr>
              ) : (
                sortedMembers.map((member) => {
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

      {/* 出欠確認変換モーダル */}
      {showConvertModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">出欠確認に変換</h3>
            </div>
            <div className="p-6 space-y-4">
              {/* タイトル入力 */}
              <div>
                <label htmlFor="convert-title" className="block text-sm font-medium text-gray-700 mb-1">
                  タイトル
                </label>
                <input
                  id="convert-title"
                  type="text"
                  value={convertTitle}
                  onChange={(e) => setConvertTitle(e.target.value)}
                  placeholder={schedule.title}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                />
                <p className="text-xs text-gray-500 mt-1">空の場合は元のタイトルが使用されます</p>
              </div>

              {/* 候補日選択 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  対象日を選択してください
                </label>
                <div className="space-y-2 max-h-60 overflow-y-auto border border-gray-200 rounded-md p-3">
                  {candidates.map((candidate) => {
                    const stats = candidateStats.find((s) => s.candidateId === candidate.candidate_id);
                    const isSelected = selectedCandidateIds.includes(candidate.candidate_id);
                    return (
                      <label
                        key={candidate.candidate_id}
                        className={`flex items-center justify-between p-3 rounded-md cursor-pointer transition ${
                          isSelected ? 'bg-accent/10 border border-accent' : 'bg-gray-50 hover:bg-gray-100'
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => handleToggleCandidate(candidate.candidate_id)}
                            className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                          />
                          <div>
                            <span className="font-medium">
                              {new Date(candidate.date).toLocaleDateString('ja-JP', {
                                month: '2-digit',
                                day: '2-digit',
                                weekday: 'short',
                              })}
                            </span>
                            {(candidate.start_time || candidate.end_time) && (
                              <span className="text-sm text-gray-500 ml-2">
                                {formatTimeRange(candidate.start_time, candidate.end_time, '-')}
                              </span>
                            )}
                          </div>
                        </div>
                        <span className="text-sm text-gray-600">
                          {stats?.availableCount || 0}名が参加可能
                        </span>
                      </label>
                    );
                  })}
                </div>
                {selectedCandidateIds.length === 0 && (
                  <p className="text-xs text-red-500 mt-1">少なくとも1つの日付を選択してください</p>
                )}
              </div>
            </div>
            <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
              <button
                onClick={() => setShowConvertModal(false)}
                className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-md transition"
              >
                キャンセル
              </button>
              <button
                onClick={handleConvertToAttendance}
                disabled={converting || selectedCandidateIds.length === 0}
                className="px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition disabled:bg-gray-400"
              >
                {converting ? '変換中...' : '変換する'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
