import { useState, useEffect } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { getBusinessDayDetail, getShiftSlots, createShiftSlot } from '../lib/api';
import type { BusinessDay, ShiftSlot } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function ShiftSlotList() {
  const { businessDayId } = useParams<{ businessDayId: string }>();
  const navigate = useNavigate();
  const [businessDay, setBusinessDay] = useState<BusinessDay | null>(null);
  const [shiftSlots, setShiftSlots] = useState<ShiftSlot[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    if (businessDayId) {
      loadData();
    }
  }, [businessDayId]);

  const loadData = async () => {
    if (!businessDayId) return;

    try {
      setLoading(true);
      const [businessDayData, shiftSlotsData] = await Promise.all([
        getBusinessDayDetail(businessDayId),
        getShiftSlots(businessDayId),
      ]);
      setBusinessDay(businessDayData);
      setShiftSlots(shiftSlotsData.shift_slots);
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

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    loadData();
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  if (!businessDay) {
    return (
      <div className="card text-center py-12">
        <p className="text-gray-600">営業日が見つかりません</p>
      </div>
    );
  }

  return (
    <div>
      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to="/events" className="hover:text-gray-900">
          イベント一覧
        </Link>
        <span className="mx-2">/</span>
        <Link to={`/events/${businessDay.event_id}/business-days`} className="hover:text-gray-900">
          営業日一覧
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">
          {new Date(businessDay.target_date).toLocaleDateString('ja-JP')}
        </span>
      </nav>

      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">
            {new Date(businessDay.target_date).toLocaleDateString('ja-JP', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
              weekday: 'short',
            })}
          </h2>
          <p className="text-sm text-gray-600 mt-1">
            {businessDay.start_time.slice(0, 5)} 〜 {businessDay.end_time.slice(0, 5)}
          </p>
        </div>
        <button onClick={() => setShowCreateModal(true)} className="btn-primary">
          ＋ シフト枠を追加
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {shiftSlots.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだシフト枠がありません</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初のシフト枠を追加
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          {shiftSlots.map((slot) => (
            <div key={slot.slot_id} className="card">
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <h3 className="text-lg font-bold text-gray-900">
                    {slot.slot_name} - {slot.instance_name}
                  </h3>
                  <p className="text-sm text-gray-600 mt-1">
                    {slot.start_time.slice(0, 5)} 〜 {slot.end_time.slice(0, 5)}
                    {slot.is_overnight && ' （深夜営業）'}
                  </p>
                  <div className="mt-2">
                    <span
                      className={`inline-block px-2 py-1 text-xs font-semibold rounded ${
                        (slot.assigned_count || 0) >= slot.required_count
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {slot.assigned_count || 0} / {slot.required_count} 人
                    </span>
                  </div>
                </div>
                <button
                  onClick={() => navigate(`/shift-slots/${slot.slot_id}/assign`)}
                  className="btn-primary ml-4"
                  disabled={(slot.assigned_count || 0) >= slot.required_count}
                >
                  {(slot.assigned_count || 0) >= slot.required_count ? '満員' : '割り当て'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* シフト枠作成モーダル */}
      {showCreateModal && businessDayId && (
        <CreateShiftSlotModal
          businessDayId={businessDayId}
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}
    </div>
  );
}

// シフト枠作成モーダルコンポーネント
function CreateShiftSlotModal({
  businessDayId,
  onClose,
  onSuccess,
}: {
  businessDayId: string;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [slotName, setSlotName] = useState('');
  const [instanceName, setInstanceName] = useState('');
  const [startTime, setStartTime] = useState('21:30');
  const [endTime, setEndTime] = useState('23:00');
  const [requiredCount, setRequiredCount] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Position は固定値として扱う（テスト用）
  const dummyPositionId = '01HXX00000000000000000000'; // TODO: Position 一覧から選択できるようにする

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!slotName.trim() || !instanceName.trim()) {
      setError('役職名とインスタンス名を入力してください');
      return;
    }

    if (!startTime || !endTime) {
      setError('時刻を入力してください');
      return;
    }

    if (requiredCount < 1) {
      setError('必要人数は1人以上で入力してください');
      return;
    }

    setLoading(true);

    try {
      await createShiftSlot(businessDayId, {
        position_id: dummyPositionId,
        slot_name: slotName.trim(),
        instance_name: instanceName.trim(),
        start_time: startTime,
        end_time: endTime,
        required_count: requiredCount,
      });
      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('シフト枠の作成に失敗しました');
      }
      console.error('Failed to create shift slot:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <h3 className="text-xl font-bold text-gray-900 mb-4">シフト枠を追加</h3>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="slotName" className="label">
              役職名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="slotName"
              value={slotName}
              onChange={(e) => setSlotName(e.target.value)}
              placeholder="例: 受付"
              className="input-field"
              disabled={loading}
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label htmlFor="instanceName" className="label">
              インスタンス名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="instanceName"
              value={instanceName}
              onChange={(e) => setInstanceName(e.target.value)}
              placeholder="例: 受付1"
              className="input-field"
              disabled={loading}
            />
          </div>

          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label htmlFor="startTime" className="label">
                開始時刻 <span className="text-red-500">*</span>
              </label>
              <input
                type="time"
                id="startTime"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                className="input-field"
                disabled={loading}
              />
            </div>
            <div>
              <label htmlFor="endTime" className="label">
                終了時刻 <span className="text-red-500">*</span>
              </label>
              <input
                type="time"
                id="endTime"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
                className="input-field"
                disabled={loading}
              />
            </div>
          </div>

          <div className="mb-4">
            <label htmlFor="requiredCount" className="label">
              必要人数 <span className="text-red-500">*</span>
            </label>
            <input
              type="number"
              id="requiredCount"
              value={requiredCount}
              onChange={(e) => setRequiredCount(parseInt(e.target.value, 10))}
              min="1"
              className="input-field"
              disabled={loading}
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 btn-secondary"
              disabled={loading}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="flex-1 btn-primary"
              disabled={loading || !slotName.trim() || !instanceName.trim()}
            >
              {loading ? '作成中...' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

