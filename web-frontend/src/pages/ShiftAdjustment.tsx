import { useEffect, useState, useMemo } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import {
  getAttendanceCollection,
  getAttendanceResponses,
  type AttendanceCollection,
  type AttendanceResponse,
} from '../lib/api/attendanceApi';
import { getShiftSlots } from '../lib/api/shiftSlotApi';
import { confirmAssignment, getAssignments, cancelAssignment } from '../lib/api/shiftAssignmentApi';
import { getEventBusinessDays } from '../lib/api/eventApi';
import { getActualAttendance } from '../lib/api/actualAttendanceApi';
import { getMembers } from '../lib/api/memberApi';
import { listRoles, type Role } from '../lib/api/roleApi';
import { ApiClientError } from '../lib/apiClient';
import type { InstanceData } from '../lib/shiftTextExport';
import ShiftTextPreviewModal from '../components/ShiftTextPreviewModal';
import type { ShiftSlot, ShiftAssignment, BusinessDay, RecentAttendanceResponse } from '../types/api';

interface AttendingMember {
  memberId: string;
  memberName: string;
  availableFrom?: string;
  availableTo?: string;
}

interface SlotWithAssignments {
  slot: ShiftSlot;
  assignments: ShiftAssignment[];
}

interface InstanceGroup {
  instanceName: string;
  instanceId: string | null;
  slots: SlotWithAssignments[];
}

export default function ShiftAdjustment() {
  const { collectionId } = useParams<{ collectionId: string }>();
  const [searchParams] = useSearchParams();
  const initialDateId = searchParams.get('date');

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [collection, setCollection] = useState<AttendanceCollection | null>(null);
  const [responses, setResponses] = useState<AttendanceResponse[]>([]);
  const [selectedDateId, setSelectedDateId] = useState<string>(initialDateId || '');
  const [businessDays, setBusinessDays] = useState<BusinessDay[]>([]);
  const [slots, setSlots] = useState<SlotWithAssignments[]>([]);
  const [attendingMembers, setAttendingMembers] = useState<AttendingMember[]>([]);
  const [assigning, setAssigning] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  // アクションエラー/成功メッセージ（UI表示用）
  const [actionError, setActionError] = useState<string | null>(null);
  const [actionSuccess, setActionSuccess] = useState<string | null>(null);
  // スロットごとの選択メンバー（競合状態回避用）
  const [selectedMembers, setSelectedMembers] = useState<Record<string, string>>({});
  // 本出席状況（参考用）
  const [actualAttendance, setActualAttendance] = useState<RecentAttendanceResponse | null>(null);
  // 未来の日付を含めるかどうか
  const [includeFuture, setIncludeFuture] = useState(false);
  // ロールフィルター用
  const [roles, setRoles] = useState<Role[]>([]);
  const [memberRoleMap, setMemberRoleMap] = useState<Map<string, string[]>>(new Map());
  const [selectedRoleIds, setSelectedRoleIds] = useState<Set<string>>(new Set());
  // インスタンス表プレビュー用
  const [showPreviewModal, setShowPreviewModal] = useState(false);

  // Load collection and responses
  useEffect(() => {
    if (!collectionId) {
      setError('出欠確認IDが指定されていません');
      setLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setLoading(true);
        const [collectionData, responsesData] = await Promise.all([
          getAttendanceCollection(collectionId),
          getAttendanceResponses(collectionId),
        ]);
        setCollection(collectionData);
        setResponses(responsesData || []);

        // イベントに紐づいていない場合はエラー
        if (collectionData.target_type !== 'event' || !collectionData.target_id) {
          setError('この出欠確認はイベントに紐づけられていないため、シフト調整を行えません。');
          setLoading(false);
          return;
        }

        // Set initial selected date if not set
        if (!selectedDateId && collectionData.target_dates && collectionData.target_dates.length > 0) {
          const sortedDates = [...collectionData.target_dates].sort(
            (a, b) => a.display_order - b.display_order
          );
          setSelectedDateId(sortedDates[0].target_date_id);
        }

        // Load business days for the linked event
        if (collectionData.target_id) {
          try {
            const bds = await getEventBusinessDays(collectionData.target_id);
            setBusinessDays(bds || []);
          } catch (err) {
            console.error('Failed to load business days:', err);
            setActionError('イベントの営業日を読み込めませんでした。シフト枠が表示されない可能性があります。');
            setBusinessDays([]);
          }
        }

      } catch (err) {
        setError(err instanceof Error ? err.message : '取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [collectionId, refreshKey]);

  // 本出席状況を取得（includeFutureが変更されたら再取得）
  useEffect(() => {
    if (!collection?.target_id) return;

    const fetchActualAttendance = async () => {
      try {
        const actualAttendanceData = await getActualAttendance({
          // 未来を含める場合はイベントの全営業日を表示（上限100件）
          // 過去のみの場合は10件
          limit: includeFuture ? 100 : 10,
          event_id: collection.target_id,
          include_future: includeFuture,
        });
        // 空配列の場合はnullを設定して不要な描画を避ける
        if (actualAttendanceData.target_dates && actualAttendanceData.target_dates.length > 0) {
          setActualAttendance(actualAttendanceData);
        } else {
          setActualAttendance(null);
        }
      } catch {
        // エラーでも続行
      }
    };

    fetchActualAttendance();
  }, [collection?.target_id, includeFuture]);

  // ロールとメンバーのロール情報を取得
  useEffect(() => {
    const fetchRolesAndMembers = async () => {
      try {
        const [rolesData, membersData] = await Promise.all([
          listRoles(),
          getMembers({ is_active: true }),
        ]);
        setRoles(rolesData);

        // メンバーIDからロールIDsへのマップを作成
        const roleMap = new Map<string, string[]>();
        for (const member of membersData.members) {
          roleMap.set(member.member_id, member.role_ids || []);
        }
        setMemberRoleMap(roleMap);
      } catch {
        // エラーでも続行
      }
    };

    fetchRolesAndMembers();
  }, []);

  // refreshKey変更時にselectedMembersをリセット（競合状態回避）
  useEffect(() => {
    setSelectedMembers({});
  }, [refreshKey]);

  // Load slots and assignments when date changes
  useEffect(() => {
    if (!selectedDateId || !collection) return;

    const loadSlotsForDate = async () => {
      // Find the target date
      const targetDate = collection.target_dates?.find((d) => d.target_date_id === selectedDateId);
      if (!targetDate) return;

      // Find matching business day by date
      const targetDateStr = targetDate.target_date.split('T')[0];
      const matchingBD = businessDays.find((bd) => bd.target_date.split('T')[0] === targetDateStr);

      if (matchingBD) {
        try {
          // Load shift slots for this business day
          const slotsData = await getShiftSlots(matchingBD.business_day_id);
          const shiftSlots = slotsData.shift_slots || [];

          // Load assignments for each slot
          const slotsWithAssignments: SlotWithAssignments[] = await Promise.all(
            shiftSlots.map(async (slot) => {
              const assignmentsData = await getAssignments({
                slot_id: slot.slot_id,
                assignment_status: 'confirmed',
              });
              return {
                slot,
                assignments: assignmentsData.assignments || [],
              };
            })
          );

          setSlots(slotsWithAssignments);
        } catch (err) {
          console.error('Failed to load slots:', err);
          setSlots([]);
        }
      } else {
        setSlots([]);
      }

      // Get attending members for this date
      const attending = responses
        .filter((r) => r.target_date_id === selectedDateId && r.response === 'attending')
        .map((r) => ({
          memberId: r.member_id,
          memberName: r.member_name,
          availableFrom: r.available_from,
          availableTo: r.available_to,
        }));
      setAttendingMembers(attending);
    };

    loadSlotsForDate();
  }, [selectedDateId, collection, businessDays, responses, refreshKey]);

  // メッセージを一定時間後にクリア
  const clearMessages = () => {
    setTimeout(() => {
      setActionError(null);
      setActionSuccess(null);
    }, 5000);
  };

  const handleAssign = async (slotId: string) => {
    const memberId = selectedMembers[slotId];
    if (!memberId) return;

    setAssigning(slotId);
    setActionError(null);
    setActionSuccess(null);

    try {
      await confirmAssignment({
        slot_id: slotId,
        member_id: memberId,
      });
      // 成功時に選択状態をクリアしてリフレッシュ
      setSelectedMembers((prev) => ({ ...prev, [slotId]: '' }));
      setActionSuccess('メンバーをアサインしました');
      setRefreshKey((k) => k + 1);
      clearMessages();
    } catch (err) {
      const message = err instanceof ApiClientError
        ? err.getUserMessage()
        : (err instanceof Error ? err.message : 'アサインに失敗しました');
      setActionError(message);
      clearMessages();
    } finally {
      setAssigning(null);
    }
  };

  const handleRemoveAssignment = async (assignmentId: string) => {
    if (!confirm('このアサインを取り消しますか？')) return;

    setActionError(null);
    setActionSuccess(null);

    try {
      await cancelAssignment(assignmentId);
      setActionSuccess('アサインを取り消しました');
      setRefreshKey((k) => k + 1);
      clearMessages();
    } catch (err) {
      const message = err instanceof ApiClientError
        ? err.getUserMessage()
        : (err instanceof Error ? err.message : '取り消しに失敗しました');
      setActionError(message);
      clearMessages();
    }
  };

  // Get members already assigned to any slot (member_id -> "instance_name-slot_name")
  const assignedMemberSlots = useMemo(
    () => new Map<string, string>(
      slots.flatMap((s) => s.assignments.map((a) => [
        a.member_id,
        `${s.slot.instance_name}-${s.slot.slot_name}`
      ] as [string, string]))
    ),
    [slots]
  );

  // Get available members (attending but not yet assigned)
  const availableMembers = useMemo(
    () => attendingMembers.filter((m) => !assignedMemberSlots.has(m.memberId)),
    [attendingMembers, assignedMemberSlots]
  );

  // スロットをインスタンスごとにグループ化し、priority昇順でソート
  const groupSlotsByInstance = useMemo((): InstanceGroup[] => {
    const instanceMap = new Map<string, InstanceGroup>();

    slots.forEach((slotWithAssignments) => {
      const { slot } = slotWithAssignments;
      const instanceId = slot.instance_id || null;
      const instanceName = slot.instance_name || '未分類';
      const key = instanceId || `__name__${instanceName}`;

      if (!instanceMap.has(key)) {
        instanceMap.set(key, {
          instanceName,
          instanceId,
          slots: [],
        });
      }
      instanceMap.get(key)!.slots.push(slotWithAssignments);
    });

    // 結果を配列に変換
    const result = Array.from(instanceMap.values());

    // 各インスタンス内のスロットをpriority昇順でソート（小さいほど優先）
    result.forEach((group) => {
      group.slots.sort((a, b) => a.slot.priority - b.slot.priority);
    });

    // インスタンスを名前でソート（未分類は最後）
    result.sort((a, b) => {
      if (a.instanceName === '未分類') return 1;
      if (b.instanceName === '未分類') return -1;
      return a.instanceName.localeCompare(b.instanceName, 'ja');
    });

    return result;
  }, [slots]);

  // プレビューモーダル用のインスタンスデータを生成
  const getInstanceDataForPreview = (): InstanceData[] => {
    return groupSlotsByInstance.map((group) => ({
      instanceName: group.instanceName,
      slots: group.slots.map(({ slot, assignments }) => ({
        slotName: slot.slot_name,
        assignments: assignments.map((a) => ({
          memberName: a.member_display_name || a.member_id,
        })),
      })),
    }));
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
            出欠確認一覧に戻る
          </Link>
        </div>
      </div>
    );
  }

  const sortedDates = (collection.target_dates || []).sort((a, b) => a.display_order - b.display_order);
  const selectedTargetDate = sortedDates.find((d) => d.target_date_id === selectedDateId);

  return (
    <div className="max-w-7xl mx-auto">
      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to="/attendance" className="hover:text-gray-900">
          出欠確認一覧
        </Link>
        <span className="mx-2">/</span>
        <Link to={`/attendance/${collectionId}`} className="hover:text-gray-900">
          {collection.title}
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">シフト調整</span>
      </nav>

      {/* ヘッダー */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">シフト調整</h1>
            <p className="text-gray-600">{collection.title}の出欠データをもとにシフトを調整</p>
          </div>
          <div className="flex gap-2 items-center">
            {/* インスタンス表プレビューボタン */}
            <button
              onClick={() => setShowPreviewModal(true)}
              disabled={slots.length === 0}
              className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-lg flex items-center disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg
                className="w-5 h-5 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                />
              </svg>
              インスタンス表を出力
            </button>
          </div>
        </div>

        {/* アクションメッセージ */}
        {actionError && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
            <p className="text-red-800 text-sm">{actionError}</p>
          </div>
        )}
        {actionSuccess && (
          <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-md">
            <p className="text-green-800 text-sm">{actionSuccess}</p>
          </div>
        )}
      </div>

      {/* 日付タブ */}
      <div className="bg-white rounded-lg shadow mb-6">
        <div className="border-b border-gray-200">
          <nav className="flex overflow-x-auto -mb-px">
            {sortedDates.map((date) => {
              const attendingCount = responses.filter(
                (r) => r.target_date_id === date.target_date_id && r.response === 'attending'
              ).length;
              const isSelected = date.target_date_id === selectedDateId;
              const dateObj = new Date(date.target_date);

              return (
                <button
                  key={date.target_date_id}
                  onClick={() => setSelectedDateId(date.target_date_id)}
                  className={`px-4 py-3 text-sm font-medium border-b-2 whitespace-nowrap transition ${
                    isSelected
                      ? 'border-accent text-accent'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  <div>
                    {dateObj.toLocaleDateString('ja-JP', { month: '2-digit', day: '2-digit' })}
                  </div>
                  <div className="text-xs text-gray-400">
                    {dateObj.toLocaleDateString('ja-JP', { weekday: 'short' })}
                  </div>
                  <div className={`text-xs ${isSelected ? 'text-accent' : 'text-gray-400'}`}>
                    参加: {attendingCount}名
                  </div>
                </button>
              );
            })}
          </nav>
        </div>
      </div>

      {/* メインコンテンツ */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 左: 参加者一覧 */}
        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            参加者
            <span className="ml-2 text-sm font-normal text-gray-500">
              ({attendingMembers.length}名)
            </span>
          </h2>

          {attendingMembers.length === 0 ? (
            <p className="text-gray-500 text-sm">参加者がいません</p>
          ) : (
            <div className="space-y-2">
              {attendingMembers.map((member) => {
                const assignedSlotName = assignedMemberSlots.get(member.memberId);
                const isAssigned = !!assignedSlotName;
                return (
                  <div
                    key={member.memberId}
                    className={`p-2 rounded-md ${
                      isAssigned ? 'bg-gray-100 text-gray-500' : 'bg-green-50'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className={isAssigned ? 'line-through' : ''}>
                        {member.memberName}
                      </span>
                      {isAssigned && (
                        <span className="text-xs text-gray-400 truncate max-w-[120px]" title={assignedSlotName}>
                          {assignedSlotName}
                        </span>
                      )}
                    </div>
                    {member.availableFrom || member.availableTo ? (
                      <div className="text-xs text-gray-500 mt-1">
                        {member.availableFrom || '?'}〜{member.availableTo || '?'}
                      </div>
                    ) : null}
                  </div>
                );
              })}
            </div>
          )}

          <div className="mt-4 pt-4 border-t border-gray-200">
            <div className="text-sm text-gray-600">
              <div className="flex justify-between">
                <span>未配置:</span>
                <span className="font-medium">{availableMembers.length}名</span>
              </div>
              <div className="flex justify-between">
                <span>配置済:</span>
                <span className="font-medium">{assignedMemberSlots.size}名</span>
              </div>
            </div>
          </div>

          {/* 本出席状況（参考） */}
          {actualAttendance && actualAttendance.target_dates && actualAttendance.target_dates.length > 0 && attendingMembers.length > 0 && (
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-sm font-semibold text-gray-900">直近の本出席状況（参考）</h3>
                <label className="flex items-center text-xs text-gray-600 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={includeFuture}
                    onChange={(e) => setIncludeFuture(e.target.checked)}
                    className="mr-1.5 rounded border-gray-300 text-accent focus:ring-accent"
                  />
                  未来の日付を含める
                </label>
              </div>
              {/* ロールフィルター */}
              {roles.length > 0 && (
                <div className="mb-2 flex flex-wrap gap-1">
                  {roles.map((role) => (
                    <label
                      key={role.role_id}
                      className={`flex items-center text-xs px-2 py-1 rounded cursor-pointer border transition ${
                        selectedRoleIds.has(role.role_id)
                          ? 'bg-accent text-white border-accent'
                          : 'bg-gray-100 text-gray-600 border-gray-300 hover:bg-gray-200'
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={selectedRoleIds.has(role.role_id)}
                        onChange={(e) => {
                          setSelectedRoleIds((prev) => {
                            const next = new Set(prev);
                            if (e.target.checked) {
                              next.add(role.role_id);
                            } else {
                              next.delete(role.role_id);
                            }
                            return next;
                          });
                        }}
                        className="sr-only"
                      />
                      {role.name}
                    </label>
                  ))}
                  {selectedRoleIds.size > 0 && (
                    <button
                      type="button"
                      onClick={() => setSelectedRoleIds(new Set())}
                      className="text-xs text-gray-500 hover:text-gray-700 underline ml-1"
                    >
                      クリア
                    </button>
                  )}
                </div>
              )}
              <div className="overflow-x-auto">
                <table className="min-w-full text-xs border-collapse border border-gray-300">
                  <thead>
                    <tr className="bg-gray-100">
                      <th className="border border-gray-300 px-2 py-1 text-left font-semibold sticky left-0 bg-gray-100 z-10">
                        メンバー
                      </th>
                      {actualAttendance.target_dates.map((td) => (
                        <th key={td.target_date_id} className="border border-gray-300 px-2 py-1 text-center font-semibold whitespace-nowrap">
                          {new Date(td.target_date).toLocaleDateString('ja-JP', {
                            month: 'numeric',
                            day: 'numeric',
                          })}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {actualAttendance.member_attendances
                      .filter((memberAtt) => {
                        // 参加者のみ表示
                        const isAttending = attendingMembers.some((am) => am.memberId === memberAtt.member_id);
                        if (!isAttending) return false;

                        // ロールフィルター（選択がなければ全員表示）
                        if (selectedRoleIds.size === 0) return true;

                        // メンバーのロールと選択ロールの交差をチェック
                        const memberRoles = memberRoleMap.get(memberAtt.member_id) || [];
                        return memberRoles.some((roleId) => selectedRoleIds.has(roleId));
                      })
                      .map((memberAtt) => (
                        <tr key={memberAtt.member_id} className="hover:bg-gray-50">
                          <td className="border border-gray-300 px-2 py-1 font-medium sticky left-0 bg-white z-10">
                            {memberAtt.member_name}
                          </td>
                          {actualAttendance.target_dates.map((td) => {
                            const status = memberAtt.attendance_map[td.target_date_id] || '';
                            let symbol = '×';
                            let color = 'text-red-600';
                            if (status === 'attended') {
                              symbol = '○';
                              color = 'text-green-600';
                            }
                            return (
                              <td key={td.target_date_id} className={`border border-gray-300 px-2 py-1 text-center ${color} font-bold`}>
                                {symbol}
                              </td>
                            );
                          })}
                        </tr>
                      ))}
                  </tbody>
                </table>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                ○: シフト割り当てあり、×: シフト割り当てなし
              </p>
            </div>
          )}
        </div>

        {/* 右: シフト枠（インスタンスごとにグループ化） */}
        <div className="lg:col-span-2">
          {slots.length === 0 ? (
            <div className="bg-white rounded-lg shadow p-6 text-center">
              <p className="text-gray-500 mb-4">この日のシフト枠がまだ作成されていません</p>
              {selectedTargetDate && (
                <p className="text-sm text-gray-400">
                  営業日詳細画面からシフト枠を作成してください
                </p>
              )}
            </div>
          ) : (
            <div className="space-y-6">
              {groupSlotsByInstance.map((group) => (
                <div key={group.instanceId || group.instanceName} className="bg-white rounded-lg shadow overflow-hidden">
                  {/* インスタンスヘッダー */}
                  <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
                    <h3 className="font-semibold text-gray-900">{group.instanceName}</h3>
                    <p className="text-xs text-gray-500 mt-0.5">
                      {group.slots.length}枠 / 配置済み: {group.slots.reduce((sum, s) => sum + s.assignments.length, 0)}人
                    </p>
                  </div>

                  {/* インスタンス内のスロット一覧 */}
                  <div className="divide-y divide-gray-100">
                    {group.slots.map(({ slot, assignments }) => {
                      const remainingCount = slot.required_count - (slot.assigned_count || assignments.length);
                      const isFull = remainingCount <= 0;

                      return (
                        <div key={slot.slot_id} className="p-4">
                          <div className="flex justify-between items-start mb-3">
                            <div>
                              <h4 className="font-medium text-gray-900">{slot.slot_name}</h4>
                              <div className="text-sm text-gray-500">
                                {slot.start_time?.substring(0, 5)} - {slot.end_time?.substring(0, 5)}
                              </div>
                            </div>
                            <div className={`text-sm font-medium ${isFull ? 'text-green-600' : 'text-amber-600'}`}>
                              {assignments.length} / {slot.required_count}
                            </div>
                          </div>

                          {/* 配置済みメンバー */}
                          {assignments.length > 0 && (
                            <div className="mb-3 space-y-1">
                              {assignments.map((assignment) => (
                                <div
                                  key={assignment.assignment_id}
                                  className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded"
                                >
                                  <span className="text-sm">{assignment.member_display_name}</span>
                                  <button
                                    onClick={() => handleRemoveAssignment(assignment.assignment_id)}
                                    className="text-red-500 hover:text-red-700 text-sm"
                                  >
                                    取消
                                  </button>
                                </div>
                              ))}
                            </div>
                          )}

                          {/* メンバー追加 */}
                          {!isFull && (
                            <div className="flex gap-2">
                              <select
                                aria-label={`${slot.instance_name}-${slot.slot_name}にアサインするメンバーを選択`}
                                className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-accent bg-white"
                                disabled={assigning === slot.slot_id || availableMembers.length === 0}
                                value={selectedMembers[slot.slot_id] || ''}
                                onChange={(e) => {
                                  setSelectedMembers((prev) => ({
                                    ...prev,
                                    [slot.slot_id]: e.target.value,
                                  }));
                                }}
                              >
                                <option value="">
                                  {availableMembers.length === 0 ? '未配置メンバーなし' : 'メンバーを選択...'}
                                </option>
                                {availableMembers.map((member) => (
                                  <option key={member.memberId} value={member.memberId}>
                                    {member.memberName}
                                    {member.availableFrom || member.availableTo
                                      ? ` (${member.availableFrom || '?'}〜${member.availableTo || '?'})`
                                      : ''}
                                  </option>
                                ))}
                              </select>
                              <button
                                type="button"
                                onClick={() => handleAssign(slot.slot_id)}
                                disabled={assigning === slot.slot_id || !selectedMembers[slot.slot_id]}
                                className="px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition text-sm disabled:bg-gray-400 disabled:cursor-not-allowed"
                              >
                                {assigning === slot.slot_id ? '処理中...' : '追加'}
                              </button>
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* インスタンス表プレビューモーダル */}
      <ShiftTextPreviewModal
        isOpen={showPreviewModal}
        onClose={() => setShowPreviewModal(false)}
        instanceData={getInstanceDataForPreview()}
      />
    </div>
  );
}
