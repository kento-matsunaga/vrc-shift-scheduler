import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  getAttendanceCollection,
  getAttendanceResponses,
  closeAttendanceCollection,
  deleteAttendanceCollection,
  updateAttendanceResponse,
  type AttendanceCollection as AttendanceCollectionType,
  type AttendanceResponse,
  type TargetDate,
} from '../lib/api/attendanceApi';
import { getMembers } from '../lib/api';
import { getMemberGroups, getMemberGroupDetail, type MemberGroup } from '../lib/api/memberGroupApi';
import { listRoles, type Role } from '../lib/api/roleApi';
import { ApiClientError } from '../lib/apiClient';
import { isValidTimeRange } from '../lib/timeUtils';
import type { Member } from '../types/api';
import { SEO } from '../components/seo';

// ソートの種類
type SortKey = 'name' | 'attending_count' | 'date_attending';
type SortDirection = 'asc' | 'desc';

// 編集モーダル用の型
interface EditingData {
  member: Member;
  targetDate: TargetDate;
  response: 'attending' | 'absent' | 'undecided' | '';
  note: string;
  availableFrom: string;
  availableTo: string;
}

export default function AttendanceDetail() {
  const { collectionId } = useParams<{ collectionId: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [collection, setCollection] = useState<AttendanceCollectionType | null>(null);
  const [responses, setResponses] = useState<AttendanceResponse[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [appliedGroups, setAppliedGroups] = useState<MemberGroup[]>([]);
  const [appliedRoles, setAppliedRoles] = useState<Role[]>([]);
  const [allRolesState, setAllRolesState] = useState<Role[]>([]);
  const [closing, setClosing] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [publicUrl, setPublicUrl] = useState('');
  const [copied, setCopied] = useState(false);

  // 編集モーダル状態
  const [editingData, setEditingData] = useState<EditingData | null>(null);
  const [saving, setSaving] = useState(false);

  // ソート状態
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [sortTargetDateId, setSortTargetDateId] = useState<string | null>(null);

  // フィルタ状態（複数選択）
  const [filterRoleIds, setFilterRoleIds] = useState<string[]>([]);

  useEffect(() => {
    if (!collectionId) {
      setError('出欠確認IDが指定されていません');
      setLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setLoading(true);
        const [collectionData, responsesData, membersData, allGroups, allRoles] = await Promise.all([
          getAttendanceCollection(collectionId),
          getAttendanceResponses(collectionId),
          getMembers({ is_active: true }),
          getMemberGroups(),
          listRoles(),
        ]);
        setCollection(collectionData);
        setResponses(responsesData || []);
        setAllRolesState(allRoles || []);

        const groupIds = collectionData.group_ids || [];
        const roleIds = collectionData.role_ids || [];
        const hasGroupFilter = groupIds.length > 0;
        const hasRoleFilter = roleIds.length > 0;

        // グループでフィルタリング
        let allowedMemberIdsByGroup: Set<string> | null = null;
        if (hasGroupFilter) {
          const groups = (allGroups.groups || []).filter((g: MemberGroup) => groupIds.includes(g.group_id));
          setAppliedGroups(groups);

          allowedMemberIdsByGroup = new Set<string>();
          for (const groupId of groupIds) {
            try {
              const groupDetail = await getMemberGroupDetail(groupId);
              (groupDetail.member_ids || []).forEach((memberId: string) => allowedMemberIdsByGroup!.add(memberId));
            } catch (e) {
              console.error('Failed to fetch group members:', e);
            }
          }
        } else {
          setAppliedGroups([]);
        }

        // ロールでフィルタリング
        let allowedMemberIdsByRole: Set<string> | null = null;
        if (hasRoleFilter) {
          const roles = allRoles.filter((r: Role) => roleIds.includes(r.role_id));
          setAppliedRoles(roles);

          allowedMemberIdsByRole = new Set<string>();
          (membersData.members || []).forEach((member: Member) => {
            if (member.role_ids && member.role_ids.some(rid => roleIds.includes(rid))) {
              allowedMemberIdsByRole!.add(member.member_id);
            }
          });
        } else {
          setAppliedRoles([]);
        }

        // AND条件でフィルタリング
        let filteredMembers = membersData.members || [];
        if (allowedMemberIdsByGroup !== null && allowedMemberIdsByRole !== null) {
          // 両方指定: AND条件
          filteredMembers = filteredMembers.filter((m: Member) =>
            allowedMemberIdsByGroup!.has(m.member_id) && allowedMemberIdsByRole!.has(m.member_id)
          );
        } else if (allowedMemberIdsByGroup !== null) {
          // グループのみ
          filteredMembers = filteredMembers.filter((m: Member) =>
            allowedMemberIdsByGroup!.has(m.member_id)
          );
        } else if (allowedMemberIdsByRole !== null) {
          // ロールのみ
          filteredMembers = filteredMembers.filter((m: Member) =>
            allowedMemberIdsByRole!.has(m.member_id)
          );
        }
        setMembers(filteredMembers);

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

  const handleDelete = async () => {
    if (!collectionId) return;
    if (!confirm('この出欠確認を削除しますか？この操作は取り消せません。')) return;

    try {
      setDeleting(true);
      await deleteAttendanceCollection(collectionId);
      alert('出欠確認を削除しました');
      navigate('/attendance');
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

  // 編集モーダルを開く
  const handleOpenEditModal = (member: Member, targetDate: TargetDate) => {
    const memberResponses = responses.filter(
      (r) => r.member_id === member.member_id && r.target_date_id === targetDate.target_date_id
    );
    const existingResponse = memberResponses[0];

    setEditingData({
      member,
      targetDate,
      response: existingResponse?.response || '',
      note: existingResponse?.note || '',
      availableFrom: existingResponse?.available_from || '',
      availableTo: existingResponse?.available_to || '',
    });
  };

  // 編集モーダルを閉じる
  const handleCloseEditModal = () => {
    setEditingData(null);
  };

  // 編集を保存
  const handleSaveEdit = async () => {
    if (!editingData || !collectionId) return;
    if (!editingData.response) {
      alert('回答を選択してください');
      return;
    }

    // 時間バリデーション: 開始時間と終了時間が同じ場合は無効（深夜営業パターンは許可）
    if (!isValidTimeRange(editingData.availableFrom || '', editingData.availableTo || '')) {
      alert('参加可能時間の開始時間と終了時間を異なる時間に設定してください');
      return;
    }

    try {
      setSaving(true);
      await updateAttendanceResponse(collectionId, {
        member_id: editingData.member.member_id,
        target_date_id: editingData.targetDate.target_date_id,
        response: editingData.response as 'attending' | 'absent' | 'undecided',
        note: editingData.note || undefined,
        available_from: editingData.availableFrom || undefined,
        available_to: editingData.availableTo || undefined,
      });

      // 成功したらデータを再取得
      const responsesData = await getAttendanceResponses(collectionId);
      setResponses(responsesData || []);
      setEditingData(null);
    } catch (err) {
      alert(err instanceof Error ? err.message : '保存に失敗しました');
    } finally {
      setSaving(false);
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

  // Create note map for quick lookup: member_id -> note (most recent response's note)
  // Store note with timestamp to track the most recent one efficiently
  const noteDataMap = new Map<string, { note: string; respondedAt: Date }>();
  responses.forEach((resp) => {
    if (resp.note && resp.note.trim()) {
      const existing = noteDataMap.get(resp.member_id);
      const respondedAt = new Date(resp.responded_at);
      // Keep the most recent note per member
      if (!existing || respondedAt > existing.respondedAt) {
        noteDataMap.set(resp.member_id, { note: resp.note, respondedAt });
      }
    }
  });
  // Convert to simple note map for easier access
  const noteMap = new Map<string, string>();
  noteDataMap.forEach((data, memberId) => {
    noteMap.set(memberId, data.note);
  });

  // ロールのインデックスマップ（ロール順ソート用）
  const roleIndexMap = new Map(filterRoleIds.map((id, i) => [id, i]));

  // メンバーの最小ロールインデックスを取得（選択ロール順でグループ化）
  const getMemberRoleIndex = (member: Member): number => {
    if (filterRoleIds.length === 0) return 0;
    let minIndex = filterRoleIds.length;
    for (const rid of member.role_ids || []) {
      const idx = roleIndexMap.get(rid);
      if (idx !== undefined && idx < minIndex) {
        minIndex = idx;
      }
    }
    return minIndex;
  };

  // ソート・フィルタリング処理
  const sortedAndFilteredMembers = [...members]
    // ロールでフィルタ
    .filter((member) => {
      if (filterRoleIds.length === 0) return true;
      return member.role_ids?.some((rid) => filterRoleIds.includes(rid));
    })
    // ソート
    .sort((a, b) => {
      // 複数ロール選択時はロール順でグループ化
      if (filterRoleIds.length >= 2) {
        const roleComp = getMemberRoleIndex(a) - getMemberRoleIndex(b);
        if (roleComp !== 0) return roleComp;
      }

      let comparison = 0;

      if (sortKey === 'name') {
        // 名前でソート（日本語対応）
        comparison = a.display_name.localeCompare(b.display_name, 'ja');
      } else if (sortKey === 'attending_count') {
        // 全体の参加数でソート
        const aCount = sortedTargetDates.filter(
          (td) => responseMap.get(a.member_id)?.get(td.target_date_id) === 'attending'
        ).length;
        const bCount = sortedTargetDates.filter(
          (td) => responseMap.get(b.member_id)?.get(td.target_date_id) === 'attending'
        ).length;
        comparison = aCount - bCount;
      } else if (sortKey === 'date_attending' && sortTargetDateId) {
        // 特定の日付の参加状態でソート
        const aResponse = responseMap.get(a.member_id)?.get(sortTargetDateId);
        const bResponse = responseMap.get(b.member_id)?.get(sortTargetDateId);
        const order = { attending: 0, undecided: 1, absent: 2, undefined: 3 };
        const aOrder = aResponse ? order[aResponse] : order.undefined;
        const bOrder = bResponse ? order[bResponse] : order.undefined;
        comparison = aOrder - bOrder;
      }

      return sortDirection === 'asc' ? comparison : -comparison;
    });

  // ソートハンドラ
  const handleSort = (key: SortKey, targetDateId?: string) => {
    if (key === 'date_attending' && targetDateId) {
      if (sortKey === 'date_attending' && sortTargetDateId === targetDateId) {
        setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
      } else {
        setSortKey('date_attending');
        setSortTargetDateId(targetDateId);
        setSortDirection('asc');
      }
    } else if (key === sortKey && key !== 'date_attending') {
      setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortTargetDateId(null);
      setSortDirection('asc');
    }
  };

  // ソートアイコン
  const SortIcon = ({ active, direction }: { active: boolean; direction: SortDirection }) => (
    <span className={`ml-1 inline-block ${active ? 'text-accent' : 'text-gray-400'}`}>
      {direction === 'asc' ? '↑' : '↓'}
    </span>
  );

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
      <SEO noindex={true} />
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
            <button
              onClick={handleDelete}
              disabled={deleting}
              className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition disabled:bg-red-400 text-sm"
            >
              {deleting ? '削除中...' : '削除'}
            </button>
            {/* シフト調整ボタン: イベントに紐づけられた出欠確認が締め切られた場合のみ表示 */}
            {collection.status === 'closed' && collection.target_type === 'event' && collection.target_id && (
              <button
                onClick={() => navigate(`/attendance/${collectionId}/shift-adjustment`)}
                className="px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition text-sm"
              >
                シフト調整
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
          </div>
        )}

        {/* 対象ロール */}
        {appliedRoles.length > 0 && (
          <div className={`pt-4 ${appliedGroups.length === 0 ? 'border-t border-gray-200' : ''} mb-4`}>
            <h3 className="text-sm font-semibold text-gray-900 mb-2">対象ロール</h3>
            <div className="flex flex-wrap gap-2">
              {appliedRoles.map((role) => (
                <span
                  key={role.role_id}
                  className="px-3 py-1 rounded-full text-sm font-medium text-white"
                  style={{ backgroundColor: role.color || '#6366f1' }}
                >
                  {role.name}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* フィルタ説明 */}
        {(appliedGroups.length > 0 || appliedRoles.length > 0) && (
          <p className="text-xs text-gray-500 mb-4">
            {appliedGroups.length > 0 && appliedRoles.length > 0
              ? '上記グループに属し、かつ上記ロールを持つメンバーのみが回答対象です'
              : appliedGroups.length > 0
                ? '上記グループに属するメンバーのみが回答対象です'
                : '上記ロールを持つメンバーのみが回答対象です'}
          </p>
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
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">出欠確認状況</h2>
              <p className="text-sm text-gray-600 mt-1">
                ○: 参加、△: 未定、×: 不参加、-: 未回答
              </p>
            </div>
            {/* ソート・フィルタコントロール */}
            <div className="flex flex-wrap items-center gap-2">
              {/* ロールフィルタ（複数選択） */}
              {allRolesState.length > 0 && (
                <div className="flex flex-wrap items-center gap-1.5">
                  {allRolesState.map((role) => {
                    const isSelected = filterRoleIds.includes(role.role_id);
                    return (
                      <button
                        key={role.role_id}
                        onClick={() =>
                          setFilterRoleIds((prev) =>
                            isSelected
                              ? prev.filter((id) => id !== role.role_id)
                              : [...prev, role.role_id]
                          )
                        }
                        className={`px-2.5 py-1 text-xs font-medium rounded-full border transition ${
                          isSelected
                            ? 'text-white border-transparent'
                            : 'bg-white text-gray-600 border-gray-300 hover:border-gray-400'
                        }`}
                        style={isSelected ? { backgroundColor: role.color || '#6366f1' } : undefined}
                      >
                        {role.name}
                      </button>
                    );
                  })}
                  {filterRoleIds.length > 0 && (
                    <button
                      onClick={() => setFilterRoleIds([])}
                      className="px-2 py-1 text-xs text-gray-500 hover:text-gray-700 transition"
                    >
                      クリア
                    </button>
                  )}
                </div>
              )}
              {/* ソート選択 */}
              <select
                value={sortKey === 'date_attending' ? `date_${sortTargetDateId}` : sortKey}
                onChange={(e) => {
                  const value = e.target.value;
                  if (value === 'name') {
                    handleSort('name');
                  } else if (value === 'attending_count') {
                    handleSort('attending_count');
                  } else if (value.startsWith('date_')) {
                    handleSort('date_attending', value.replace('date_', ''));
                  }
                }}
                className="px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
              >
                <option value="name">名前順</option>
                <option value="attending_count">参加数順</option>
                {sortedTargetDates.map((td) => (
                  <option key={td.target_date_id} value={`date_${td.target_date_id}`}>
                    {new Date(td.target_date).toLocaleDateString('ja-JP', { month: '2-digit', day: '2-digit' })}の参加状況
                  </option>
                ))}
              </select>
              {/* 昇順・降順 */}
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
                {sortedTargetDates.map((targetDate) => (
                  <th
                    key={targetDate.target_date_id}
                    className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider min-w-[100px] cursor-pointer hover:bg-gray-100"
                    onClick={() => handleSort('date_attending', targetDate.target_date_id)}
                  >
                    <div className="flex flex-col items-center">
                      <span className="flex items-center">
                        {new Date(targetDate.target_date).toLocaleDateString('ja-JP', {
                          month: '2-digit',
                          day: '2-digit',
                        })}
                        {sortKey === 'date_attending' && sortTargetDateId === targetDate.target_date_id && (
                          <SortIcon active={true} direction={sortDirection} />
                        )}
                      </span>
                      <span className="text-xs font-normal normal-case text-gray-400">
                        {new Date(targetDate.target_date).toLocaleDateString('ja-JP', {
                          weekday: 'short',
                        })}
                      </span>
                    </div>
                  </th>
                ))}
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider min-w-[120px]">
                  備考
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {/* 集計行（先頭） */}
              <tr className="bg-gray-50">
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
                <td className="px-4 py-3 text-center text-xs text-gray-500">
                  {noteMap.size > 0 && `${noteMap.size}件`}
                </td>
              </tr>
              {sortedAndFilteredMembers.length === 0 ? (
                <tr>
                  <td colSpan={sortedTargetDates.length + 2} className="px-6 py-12 text-center text-gray-500">
                    {members.length === 0 ? 'メンバーがいません' : 'フィルタ条件に一致するメンバーがいません'}
                  </td>
                </tr>
              ) : (
                sortedAndFilteredMembers.map((member) => {
                  const memberResponses = responseMap.get(member.member_id);
                  const memberNote = noteMap.get(member.member_id);
                  return (
                    <tr key={member.member_id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 sticky left-0 bg-white">
                        <span className="flex items-center gap-1">
                          {member.display_name}
                          {memberNote && (
                            <span className="text-amber-500" title="備考あり">
                              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path fillRule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zM9 9h2v2H9V9z" clipRule="evenodd" />
                              </svg>
                            </span>
                          )}
                        </span>
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
                            className={`px-4 py-4 text-center ${bgColor} cursor-pointer hover:opacity-80 transition-opacity`}
                            title={timeDisplay ? `${timeDisplay} (クリックで編集)` : 'クリックで編集'}
                            onClick={() => handleOpenEditModal(member, targetDate)}
                          >
                            <div className="text-lg font-semibold">{content}</div>
                            {timeDisplay && (
                              <div className="text-xs text-gray-600 mt-1">{timeDisplay}</div>
                            )}
                          </td>
                        );
                      })}
                      <td className="px-4 py-4 text-sm text-gray-600 max-w-xs">
                        {memberNote ? (
                          <div className="truncate" title={memberNote}>
                            {memberNote}
                          </div>
                        ) : (
                          <span className="text-gray-300">-</span>
                        )}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* 編集モーダル */}
      {editingData && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">
                出欠回答を編集
              </h3>
              <p className="text-sm text-gray-600 mt-1">
                {editingData.member.display_name} - {new Date(editingData.targetDate.target_date).toLocaleDateString('ja-JP', {
                  year: 'numeric',
                  month: '2-digit',
                  day: '2-digit',
                  weekday: 'short',
                })}
              </p>
            </div>
            <div className="px-6 py-4 space-y-4">
              {/* 出欠ステータス */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  出欠 <span className="text-red-500">*</span>
                </label>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => setEditingData({ ...editingData, response: 'attending' })}
                    className={`flex-1 px-4 py-3 rounded-md border-2 transition font-medium ${
                      editingData.response === 'attending'
                        ? 'border-green-500 bg-green-50 text-green-800'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    ○ 参加
                  </button>
                  <button
                    type="button"
                    onClick={() => setEditingData({ ...editingData, response: 'undecided' })}
                    className={`flex-1 px-4 py-3 rounded-md border-2 transition font-medium ${
                      editingData.response === 'undecided'
                        ? 'border-yellow-500 bg-yellow-50 text-yellow-800'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    △ 未定
                  </button>
                  <button
                    type="button"
                    onClick={() => setEditingData({ ...editingData, response: 'absent' })}
                    className={`flex-1 px-4 py-3 rounded-md border-2 transition font-medium ${
                      editingData.response === 'absent'
                        ? 'border-red-500 bg-red-50 text-red-800'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    × 不参加
                  </button>
                </div>
              </div>

              {/* 参加可能時間 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  参加可能時間（任意）
                </label>
                <div className="flex items-center gap-2">
                  <input
                    type="time"
                    value={editingData.availableFrom}
                    onChange={(e) => setEditingData({ ...editingData, availableFrom: e.target.value })}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                  />
                  <span className="text-gray-500">〜</span>
                  <input
                    type="time"
                    value={editingData.availableTo}
                    onChange={(e) => setEditingData({ ...editingData, availableTo: e.target.value })}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                  />
                </div>
              </div>

              {/* 備考 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  備考（任意）
                </label>
                <textarea
                  value={editingData.note}
                  onChange={(e) => setEditingData({ ...editingData, note: e.target.value })}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent resize-none"
                  placeholder="メモや補足があれば入力してください"
                />
              </div>
            </div>
            <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
              <button
                type="button"
                onClick={handleCloseEditModal}
                disabled={saving}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition disabled:opacity-50"
              >
                キャンセル
              </button>
              <button
                type="button"
                onClick={handleSaveEdit}
                disabled={saving || !editingData.response}
                className="px-4 py-2 text-sm font-medium text-white bg-accent rounded-md hover:bg-accent-dark transition disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
