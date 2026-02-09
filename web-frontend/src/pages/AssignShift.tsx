import { useState, useEffect } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { SEO } from '../components/seo';
import { getShiftSlotDetail, getMembers, confirmAssignment, getRecentAttendance, getActualAttendance, getBusinessDayDetail, getAssignments, cancelAssignment } from '../lib/api';
import { listRoles, type Role } from '../lib/api/roleApi';
import { getMemberGroups, type MemberGroup } from '../lib/api/memberGroupApi';
import { getRoleGroups, type RoleGroup } from '../lib/api/roleGroupApi';
import { getEventGroupAssignments, type EventGroupAssignments } from '../lib/api/eventApi';
import type { ShiftSlot, Member, RecentAttendanceResponse } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function AssignShift() {
  const { slotId } = useParams<{ slotId: string }>();
  const navigate = useNavigate();
  const [shiftSlot, setShiftSlot] = useState<ShiftSlot | null>(null);
  const [businessDay, setBusinessDay] = useState<Record<string, unknown> | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [memberGroups, setMemberGroups] = useState<MemberGroup[]>([]);
  const [roleGroups, setRoleGroups] = useState<RoleGroup[]>([]);
  const [eventGroupAssignments, setEventGroupAssignments] = useState<EventGroupAssignments | null>(null);
  const [actualAttendance, setActualAttendance] = useState<RecentAttendanceResponse | null>(null);
  const [todayAttendance, setTodayAttendance] = useState<string[]>([]);
  const [todayAttendingMemberIds, setTodayAttendingMemberIds] = useState<string[]>([]);
  const [existingAssignmentIds, setExistingAssignmentIds] = useState<string[]>([]);
  const [selectedMemberIds, setSelectedMemberIds] = useState<string[]>([]);
  const [note, setNote] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // ロールフィルター（表用とメンバー選択用で別々に管理）
  const [tableFilterRoleIds, setTableFilterRoleIds] = useState<string[]>([]);
  const [memberFilterRoleIds, setMemberFilterRoleIds] = useState<string[]>([]);
  // 参加回答者のみ表示フィルター
  const [showOnlyAttending, setShowOnlyAttending] = useState(false);

  // ロールのカラーを取得
  const getRoleColor = (roleId: string) => {
    const role = roles.find((r) => r.role_id === roleId);
    return role?.color || '#6B7280';
  };

  // ロール名を取得
  const getRoleName = (roleId: string) => {
    const role = roles.find((r) => r.role_id === roleId);
    return role?.name || 'Unknown';
  };

  // メンバーIDからロールIDリストを取得（membersから）
  const getMemberRoleIds = (memberId: string): string[] => {
    const member = members.find((m) => m.member_id === memberId);
    return member?.role_ids || [];
  };

  // イベントのグループ設定に基づいて許可されたメンバーIDリストを取得
  const getAllowedMemberIds = (): string[] | null => {
    if (!eventGroupAssignments || eventGroupAssignments.member_group_ids.length === 0) {
      return null; // グループ設定がなければ全員許可
    }
    const allowedIds = new Set<string>();
    for (const groupId of eventGroupAssignments.member_group_ids) {
      const group = memberGroups.find((g) => g.group_id === groupId);
      if (group?.member_ids) {
        group.member_ids.forEach((id) => allowedIds.add(id));
      }
    }
    return Array.from(allowedIds);
  };

  // イベントのグループ設定に基づいて許可されたロールIDリストを取得
  const getAllowedRoleIds = (): string[] | null => {
    if (!eventGroupAssignments || eventGroupAssignments.role_group_ids.length === 0) {
      return null; // グループ設定がなければ全ロール許可
    }
    const allowedIds = new Set<string>();
    for (const groupId of eventGroupAssignments.role_group_ids) {
      const group = roleGroups.find((g) => g.group_id === groupId);
      if (group?.role_ids) {
        group.role_ids.forEach((id) => allowedIds.add(id));
      }
    }
    return Array.from(allowedIds);
  };

  // 許可されたメンバー一覧
  const allowedMemberIds = getAllowedMemberIds();
  const allowedMembers = allowedMemberIds
    ? members.filter((m) => allowedMemberIds.includes(m.member_id))
    : members;

  // 許可されたロール一覧
  const allowedRoleIds = getAllowedRoleIds();
  const allowedRoles = allowedRoleIds
    ? roles.filter((r) => allowedRoleIds.includes(r.role_id))
    : roles;

  // フィルタリングされたメンバー一覧（メンバー選択用）
  const filteredMembers = allowedMembers.filter((m) => {
    // ロールフィルター（許可されたロールのみ）
    if (memberFilterRoleIds.length > 0) {
      const memberAllowedRoleIds = allowedRoleIds
        ? m.role_ids?.filter((rid) => allowedRoleIds.includes(rid)) || []
        : m.role_ids || [];
      if (!memberAllowedRoleIds.some((roleId) => memberFilterRoleIds.includes(roleId))) {
        return false;
      }
    }
    // 参加回答者フィルター
    if (showOnlyAttending && todayAttendingMemberIds.length > 0) {
      if (!todayAttendingMemberIds.includes(m.member_id)) {
        return false;
      }
    }
    return true;
  });

  // フィルタリングされた本出席データ（表用）
  const filteredActualAttendance = actualAttendance
    ? {
        ...actualAttendance,
        member_attendances: actualAttendance.member_attendances.filter((memberAtt) => {
          // グループフィルター（イベントに割り当てられたメンバーグループのみ）
          if (allowedMemberIds && !allowedMemberIds.includes(memberAtt.member_id)) {
            return false;
          }
          // ロールフィルター（許可されたロールのみ）
          if (tableFilterRoleIds.length > 0) {
            const memberRoleIds = getMemberRoleIds(memberAtt.member_id);
            const memberAllowedRoleIds = allowedRoleIds
              ? memberRoleIds.filter((rid) => allowedRoleIds.includes(rid))
              : memberRoleIds;
            if (!memberAllowedRoleIds.some((roleId) => tableFilterRoleIds.includes(roleId))) {
              return false;
            }
          }
          // 参加回答者フィルター
          if (showOnlyAttending && todayAttendingMemberIds.length > 0) {
            if (!todayAttendingMemberIds.includes(memberAtt.member_id)) {
              return false;
            }
          }
          return true;
        }),
      }
    : null;

  useEffect(() => {
    if (slotId) {
      loadData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- 初回マウント時のみ実行（loadDataは関数定義のため除外）
  }, [slotId]);

  const loadData = async () => {
    if (!slotId) return;

    try {
      setLoading(true);
      const shiftSlotData = await getShiftSlotDetail(slotId);
      setShiftSlot(shiftSlotData);

      const [businessDayData, membersData, rolesData, memberGroupsData, roleGroupsData, recentAttendanceData, actualAttendanceData, existingAssignments] = await Promise.all([
        getBusinessDayDetail(shiftSlotData.business_day_id),
        getMembers({ is_active: true }),
        listRoles(),
        getMemberGroups(),
        getRoleGroups(),
        getRecentAttendance({ limit: 30 }),
        getActualAttendance({ limit: 30 }),
        getAssignments({ slot_id: slotId, assignment_status: 'confirmed' }),
      ]);

      setBusinessDay(businessDayData);
      setMembers(membersData.members || []);
      setRoles(rolesData || []);
      setMemberGroups(memberGroupsData.groups || []);
      setRoleGroups(roleGroupsData.groups || []);
      setActualAttendance(actualAttendanceData);

      // イベントのグループ割り当てを取得
      let groupAssignments: EventGroupAssignments | null = null;
      if (businessDayData.event_id) {
        try {
          groupAssignments = await getEventGroupAssignments(businessDayData.event_id);
          setEventGroupAssignments(groupAssignments);
        } catch (err) {
          console.warn('Failed to load event group assignments:', err);
          // グループ割り当てがなくても続行
        }
      }

      // 既存の割り当てを初期選択状態にする
      const assignments = existingAssignments.assignments || [];
      const assignedMemberIds = assignments.map(a => a.member_id);
      const assignmentIds = assignments.map(a => a.assignment_id);
      setSelectedMemberIds(assignedMemberIds);
      setExistingAssignmentIds(assignmentIds);

      // イベントのグループ設定に基づいて許可されたメンバーIDを計算
      let allowedMemberIdsForAttendance: string[] | null = null;
      if (groupAssignments && groupAssignments.member_group_ids.length > 0) {
        const allowedIds = new Set<string>();
        for (const groupId of groupAssignments.member_group_ids) {
          const group = memberGroupsData.groups?.find((g) => g.group_id === groupId);
          if (group?.member_ids) {
            group.member_ids.forEach((id) => allowedIds.add(id));
          }
        }
        allowedMemberIdsForAttendance = Array.from(allowedIds);
      }

      // この営業日と同じ日付の出欠確認データを集計（参加予定者のみ）
      // 同じ日付の複数のtarget_dateをすべて集計する
      const targetDateStr = businessDayData.target_date.split('T')[0]; // YYYY-MM-DD
      const matchingTargetDates = recentAttendanceData.target_dates.filter((td) => {
        const tdStr = td.target_date.split('T')[0];
        return tdStr === targetDateStr;
      });

      if (matchingTargetDates.length > 0) {
        const attendingMemberNamesSet = new Set<string>();
        const attendingMemberIdSet = new Set<string>();
        // グループ設定がある場合は許可されたメンバーのみをフィルター
        const filteredMemberAttendances = allowedMemberIdsForAttendance
          ? recentAttendanceData.member_attendances.filter((ma) => allowedMemberIdsForAttendance!.includes(ma.member_id))
          : recentAttendanceData.member_attendances;

        filteredMemberAttendances.forEach((memberAtt) => {
          // 複数のtarget_dateをチェックし、いずれかで"attending"なら参加予定
          for (const matchingTargetDate of matchingTargetDates) {
            const response = memberAtt.attendance_map[matchingTargetDate.target_date_id];
            if (response === 'attending') {
              attendingMemberNamesSet.add(memberAtt.member_name);
              attendingMemberIdSet.add(memberAtt.member_id);
              break; // 1つでも参加なら追加してループ終了
            }
          }
        });
        setTodayAttendance(Array.from(attendingMemberNamesSet));
        setTodayAttendingMemberIds(Array.from(attendingMemberIdSet));
      }
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('データの取得に失敗しました');
      }
      console.error('Failed to load data:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleMember = (memberId: string) => {
    setSelectedMemberIds((prev) =>
      prev.includes(memberId)
        ? prev.filter((id) => id !== memberId)
        : [...prev, memberId]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!slotId) return;

    if (shiftSlot && selectedMemberIds.length > shiftSlot.required_count) {
      setError(`必要人数は${shiftSlot.required_count}人です。${selectedMemberIds.length}人選択されています。`);
      return;
    }

    setSubmitting(true);

    try {
      // 1. 既存の割り当てを全て削除
      for (const assignmentId of existingAssignmentIds) {
        try {
          await cancelAssignment(assignmentId);
        } catch (err) {
          console.error('Failed to cancel assignment:', err);
        }
      }

      // 2. 選択された全メンバーを新規に割り当て
      if (selectedMemberIds.length > 0) {
        for (const memberId of selectedMemberIds) {
          await confirmAssignment({
            slot_id: slotId,
            member_id: memberId,
            note: note.trim() || undefined,
          });
        }
        setSuccess(`${selectedMemberIds.length}人のシフトを確定しました！`);
      } else {
        setSuccess('シフト割り当てを解除しました！');
      }

      // 2秒後に営業日のシフト一覧に戻る
      setTimeout(() => {
        if (shiftSlot) {
          navigate(`/business-days/${shiftSlot.business_day_id}/shift-slots`);
        }
      }, 2000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        if (err.isConflictError()) {
          setError('この枠は既に満員です。他の枠を選択してください。');
        } else {
          setError(err.getUserMessage());
        }
      } else {
        setError('シフトの確定に失敗しました');
      }
      console.error('Failed to confirm assignment:', err);
    } finally {
      setSubmitting(false);
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

  if (!shiftSlot) {
    return (
      <div className="card text-center py-12">
        <p className="text-gray-600">シフト枠が見つかりません</p>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <SEO noindex={true} />
      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to="/events" className="hover:text-gray-900">
          イベント一覧
        </Link>
        <span className="mx-2">/</span>
        <Link to={`/business-days/${shiftSlot.business_day_id}/shift-slots`} className="hover:text-gray-900">
          シフト枠一覧
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">シフト割り当て</span>
      </nav>

      <div className="card">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">シフト割り当て</h2>

        {/* シフト枠情報 */}
        <div className="bg-gray-50 rounded-lg p-4 mb-6">
          <h3 className="font-bold text-gray-900 mb-2">
            {shiftSlot.slot_name} - {shiftSlot.instance_name}
          </h3>
          <p className="text-sm text-gray-600">
            {shiftSlot.start_time.slice(0, 5)} 〜 {shiftSlot.end_time.slice(0, 5)}
            {shiftSlot.is_overnight && ' （深夜営業）'}
          </p>
          <div className="mt-2">
            <span className="inline-block px-2 py-1 text-xs font-semibold rounded bg-yellow-100 text-yellow-800">
              {shiftSlot.assigned_count || 0} / {shiftSlot.required_count} 人
            </span>
          </div>
        </div>

        {/* この日の参加予定メンバー */}
        {businessDay && todayAttendance.length > 0 && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-bold text-gray-900">
                {new Date(businessDay.target_date).toLocaleDateString('ja-JP')} の参加予定メンバー
              </h3>
              <button
                type="button"
                onClick={() => setShowOnlyAttending(!showOnlyAttending)}
                className={`inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium transition-all ${
                  showOnlyAttending
                    ? 'bg-green-600 text-white ring-2 ring-green-600 ring-offset-2'
                    : 'bg-white text-green-700 border border-green-300 hover:bg-green-50'
                }`}
              >
                {showOnlyAttending ? (
                  <>
                    <svg className="w-4 h-4 mr-1.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                    参加者のみ表示中
                  </>
                ) : (
                  '参加者のみ表示'
                )}
              </button>
            </div>
            <p className="text-sm text-green-700 mb-2">出欠確認で「参加」と回答したメンバー ({todayAttendance.length}人)</p>
            <div className="flex flex-wrap gap-2">
              {todayAttendance.map((name, idx) => (
                <span key={idx} className="inline-block px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm">
                  {name}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* 直近の本出席状況（全体） */}
        {actualAttendance && actualAttendance.target_dates && actualAttendance.target_dates.length > 0 && (
          <div className="mb-6">
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-bold text-gray-900">直近の本出席状況（参考）</h3>
            </div>

            {/* ロールフィルター（表用） */}
            {allowedRoles.length > 0 && (
              <div className="mb-3">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-gray-600">ロールでフィルター</span>
                  {tableFilterRoleIds.length > 0 && (
                    <button
                      onClick={() => setTableFilterRoleIds([])}
                      className="text-xs text-accent hover:text-accent-dark"
                    >
                      クリア
                    </button>
                  )}
                </div>
                <div className="flex flex-wrap gap-1">
                  {allowedRoles.map((role) => {
                    const isSelected = tableFilterRoleIds.includes(role.role_id);
                    return (
                      <button
                        key={role.role_id}
                        onClick={() => {
                          if (isSelected) {
                            setTableFilterRoleIds(tableFilterRoleIds.filter((id) => id !== role.role_id));
                          } else {
                            setTableFilterRoleIds([...tableFilterRoleIds, role.role_id]);
                          }
                        }}
                        className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium transition-all ${
                          isSelected
                            ? 'ring-2 ring-offset-1 ring-accent'
                            : 'opacity-60 hover:opacity-100'
                        }`}
                        style={{
                          backgroundColor: role.color || '#6B7280',
                          color: 'white',
                        }}
                      >
                        {role.name}
                        {isSelected && (
                          <svg className="w-3 h-3 ml-1" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                          </svg>
                        )}
                      </button>
                    );
                  })}
                </div>
                {tableFilterRoleIds.length > 0 && (
                  <p className="text-xs text-gray-500 mt-1">
                    {filteredActualAttendance?.member_attendances?.length || 0}人表示中
                  </p>
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
                    {(actualAttendance.target_dates || []).map((td) => (
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
                  {(filteredActualAttendance?.member_attendances || []).map((memberAtt) => (
                    <tr key={memberAtt.member_id} className="hover:bg-gray-50">
                      <td className="border border-gray-300 px-2 py-1 font-medium sticky left-0 bg-white z-10">
                        <div className="flex items-center gap-1">
                          <span>{memberAtt.member_name}</span>
                          {getMemberRoleIds(memberAtt.member_id)
                            .filter((roleId) => !allowedRoleIds || allowedRoleIds.includes(roleId))
                            .map((roleId) => (
                            <span
                              key={roleId}
                              className="inline-block w-2 h-2 rounded-full"
                              style={{ backgroundColor: getRoleColor(roleId) }}
                              title={getRoleName(roleId)}
                            />
                          ))}
                        </div>
                      </td>
                      {(actualAttendance.target_dates || []).map((td) => {
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
            <p className="text-xs text-gray-500 mt-2">
              ○: シフト割り当てあり、×: シフト割り当てなし
            </p>
          </div>
        )}

        {success ? (
          <div className="bg-green-50 border border-green-200 rounded-lg p-4 text-center">
            <p className="text-green-800 font-bold mb-2">✅ {success}</p>
            <p className="text-sm text-green-700">シフト枠一覧に戻っています...</p>
          </div>
        ) : members.length === 0 ? (
          /* メンバーが0人の場合の警告表示 */
          <div className="bg-yellow-50 border border-yellow-300 rounded-lg p-6 mb-6 text-center">
            <div className="text-yellow-600 mb-3">
              <svg className="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 className="text-lg font-bold text-yellow-800 mb-2">メンバーが登録されていません</h3>
            <p className="text-yellow-700 mb-4">
              シフト割り当てを行うには、先にメンバーを登録してください。
            </p>
            <Link
              to="/members"
              className="inline-flex items-center px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 transition-colors"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              メンバーを登録する
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <div className="mb-6">
              <label className="label">
                メンバーを選択 <span className="text-red-500">*</span>
                {shiftSlot && (
                  <span className="ml-2 text-sm font-normal text-gray-600">
                    （必要人数: {shiftSlot.required_count}人、選択中: {selectedMemberIds.length}人）
                  </span>
                )}
              </label>

              {/* ロールフィルター（メンバー選択用） */}
              {allowedRoles.length > 0 && (
                <div className="mb-3">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs text-gray-600">ロールでフィルター</span>
                    {memberFilterRoleIds.length > 0 && (
                      <button
                        type="button"
                        onClick={() => setMemberFilterRoleIds([])}
                        className="text-xs text-accent hover:text-accent-dark"
                      >
                        クリア
                      </button>
                    )}
                  </div>
                  <div className="flex flex-wrap gap-1">
                    {allowedRoles.map((role) => {
                      const isSelected = memberFilterRoleIds.includes(role.role_id);
                      return (
                        <button
                          type="button"
                          key={role.role_id}
                          onClick={() => {
                            if (isSelected) {
                              setMemberFilterRoleIds(memberFilterRoleIds.filter((id) => id !== role.role_id));
                            } else {
                              setMemberFilterRoleIds([...memberFilterRoleIds, role.role_id]);
                            }
                          }}
                          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium transition-all ${
                            isSelected
                              ? 'ring-2 ring-offset-1 ring-accent'
                              : 'opacity-60 hover:opacity-100'
                          }`}
                          style={{
                            backgroundColor: role.color || '#6B7280',
                            color: 'white',
                          }}
                        >
                          {role.name}
                          {isSelected && (
                            <svg className="w-3 h-3 ml-1" fill="currentColor" viewBox="0 0 20 20">
                              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                            </svg>
                          )}
                        </button>
                      );
                    })}
                  </div>
                  {memberFilterRoleIds.length > 0 && (
                    <p className="text-xs text-gray-500 mt-1">
                      {filteredMembers.length}人表示中
                    </p>
                  )}
                </div>
              )}

              <div className="border border-gray-300 rounded-lg p-4 max-h-64 overflow-y-auto bg-white">
                {filteredMembers.length === 0 ? (
                  <p className="text-sm text-gray-600">
                    選択したロールのメンバーがいません。
                  </p>
                ) : (
                  <div className="space-y-2">
                    {filteredMembers.map((member) => (
                      <label
                        key={member.member_id}
                        className="flex items-center p-2 hover:bg-gray-50 rounded cursor-pointer"
                      >
                        <input
                          type="checkbox"
                          checked={selectedMemberIds.includes(member.member_id)}
                          onChange={() => handleToggleMember(member.member_id)}
                          disabled={submitting}
                          className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                        />
                        <div className="ml-3 flex items-center gap-2">
                          <span className="text-sm text-gray-900">{member.display_name}</span>
                          {member.role_ids && member.role_ids.length > 0 && (
                            <div className="flex gap-1">
                              {member.role_ids
                                .filter((roleId) => !allowedRoleIds || allowedRoleIds.includes(roleId))
                                .map((roleId) => (
                                <span
                                  key={roleId}
                                  className="inline-block w-2 h-2 rounded-full"
                                  style={{ backgroundColor: getRoleColor(roleId) }}
                                  title={getRoleName(roleId)}
                                />
                              ))}
                            </div>
                          )}
                        </div>
                      </label>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <div className="mb-6">
              <label htmlFor="note" className="label">
                備考（任意）
              </label>
              <textarea
                id="note"
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="例: 急遽対応"
                className="input-field"
                rows={3}
                disabled={submitting}
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                <p className="text-sm text-red-800">{error}</p>
              </div>
            )}

            <div className="flex space-x-3">
              <button
                type="button"
                onClick={() => navigate(`/business-days/${shiftSlot.business_day_id}/shift-slots`)}
                className="flex-1 btn-secondary"
                disabled={submitting}
              >
                キャンセル
              </button>
              <button
                type="submit"
                className="flex-1 btn-primary"
                disabled={submitting || members.length === 0}
              >
                {submitting ? '更新中...' : existingAssignmentIds.length > 0 ? 'シフトを更新' : 'シフトを確定'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}

