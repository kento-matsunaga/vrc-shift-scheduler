import { useEffect, useState } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import {
  getAttendanceCollection,
  getAttendanceResponses,
  type AttendanceCollection,
  type AttendanceResponse,
} from '../lib/api/attendanceApi';
import { getShiftSlots } from '../lib/api/shiftSlotApi';
import { confirmAssignment, getAssignments, cancelAssignment } from '../lib/api/shiftAssignmentApi';
import { getEventBusinessDays, type BusinessDay } from '../lib/api/eventApi';
import type { ShiftSlot, ShiftAssignment } from '../types/api';

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

        // Set initial selected date if not set
        if (!selectedDateId && collectionData.target_dates && collectionData.target_dates.length > 0) {
          const sortedDates = [...collectionData.target_dates].sort(
            (a, b) => a.display_order - b.display_order
          );
          setSelectedDateId(sortedDates[0].target_date_id);
        }

        // Load business days if event-based
        if (collectionData.target_id) {
          try {
            const bds = await getEventBusinessDays(collectionData.target_id);
            setBusinessDays(bds || []);
          } catch {
            // Event might not exist or no business days
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

  const handleAssign = async (slotId: string, memberId: string) => {
    if (!memberId) return;

    setAssigning(slotId);
    try {
      await confirmAssignment({
        slot_id: slotId,
        member_id: memberId,
      });
      setRefreshKey((k) => k + 1);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'アサインに失敗しました');
    } finally {
      setAssigning(null);
    }
  };

  const handleRemoveAssignment = async (assignmentId: string) => {
    if (!confirm('このアサインを取り消しますか？')) return;

    try {
      await cancelAssignment(assignmentId);
      setRefreshKey((k) => k + 1);
    } catch (err) {
      alert(err instanceof Error ? err.message : '取り消しに失敗しました');
    }
  };

  // Get members already assigned to any slot
  const assignedMemberIds = new Set(
    slots.flatMap((s) => s.assignments.map((a) => a.member_id))
  );

  // Get available members (attending but not yet assigned)
  const availableMembers = attendingMembers.filter((m) => !assignedMemberIds.has(m.memberId));

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
        <h1 className="text-2xl font-bold text-gray-900 mb-2">シフト調整</h1>
        <p className="text-gray-600">{collection.title}の出欠データをもとにシフトを調整</p>
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
                const isAssigned = assignedMemberIds.has(member.memberId);
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
                        <span className="text-xs text-gray-400">配置済</span>
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
                <span className="font-medium">{assignedMemberIds.size}名</span>
              </div>
            </div>
          </div>
        </div>

        {/* 右: シフト枠 */}
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
            <div className="space-y-4">
              {slots.map(({ slot, assignments }) => {
                const remainingCount = slot.required_count - (slot.assigned_count || assignments.length);
                const isFull = remainingCount <= 0;

                return (
                  <div key={slot.slot_id} className="bg-white rounded-lg shadow p-4">
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <h3 className="font-semibold text-gray-900">{slot.slot_name}</h3>
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
                          className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-accent bg-white"
                          disabled={assigning === slot.slot_id || availableMembers.length === 0}
                          onChange={(e) => {
                            if (e.target.value) {
                              handleAssign(slot.slot_id, e.target.value);
                              e.target.value = '';
                            }
                          }}
                          defaultValue=""
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
                        {assigning === slot.slot_id && (
                          <div className="flex items-center text-gray-500 text-sm">
                            処理中...
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
