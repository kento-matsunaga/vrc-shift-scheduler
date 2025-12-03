import { useState, useEffect } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { getShiftSlotDetail, getMembers, confirmAssignment } from '../lib/api';
import type { ShiftSlot, Member } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function AssignShift() {
  const { slotId } = useParams<{ slotId: string }>();
  const navigate = useNavigate();
  const [shiftSlot, setShiftSlot] = useState<ShiftSlot | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [selectedMemberId, setSelectedMemberId] = useState('');
  const [note, setNote] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (slotId) {
      loadData();
    }
  }, [slotId]);

  const loadData = async () => {
    if (!slotId) return;

    try {
      setLoading(true);
      const [shiftSlotData, membersData] = await Promise.all([
        getShiftSlotDetail(slotId),
        getMembers({ is_active: true }),
      ]);
      setShiftSlot(shiftSlotData);
      setMembers(membersData.members);
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!slotId) return;

    if (!selectedMemberId) {
      setError('メンバーを選択してください');
      return;
    }

    setSubmitting(true);

    try {
      await confirmAssignment({
        slot_id: slotId,
        member_id: selectedMemberId,
        note: note.trim() || undefined,
      });
      setSuccess('シフトを確定しました！');
      
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
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
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

        {success ? (
          <div className="bg-green-50 border border-green-200 rounded-lg p-4 text-center">
            <p className="text-green-800 font-bold mb-2">✅ {success}</p>
            <p className="text-sm text-green-700">シフト枠一覧に戻っています...</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <div className="mb-6">
              <label htmlFor="member" className="label">
                メンバーを選択 <span className="text-red-500">*</span>
              </label>
              <select
                id="member"
                value={selectedMemberId}
                onChange={(e) => setSelectedMemberId(e.target.value)}
                className="input-field"
                disabled={submitting}
              >
                <option value="">-- メンバーを選択してください --</option>
                {members.map((member) => (
                  <option key={member.member_id} value={member.member_id}>
                    {member.display_name}
                  </option>
                ))}
              </select>
              {members.length === 0 && (
                <p className="text-xs text-red-600 mt-1">
                  メンバーが登録されていません。先にメンバーを登録してください。
                </p>
              )}
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
                disabled={submitting || !selectedMemberId || members.length === 0}
              >
                {submitting ? '確定中...' : 'シフトを確定'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}

