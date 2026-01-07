import { useState, useEffect } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { getBusinessDayDetail, getShiftSlots, createShiftSlot, getAssignments, applyTemplateToBusinessDay } from '../lib/api';
import { listTemplates } from '../lib/api/templateApi';
import type { BusinessDay, ShiftSlot, ShiftAssignment, Template } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function ShiftSlotList() {
  const { businessDayId } = useParams<{ businessDayId: string }>();
  const navigate = useNavigate();
  const [businessDay, setBusinessDay] = useState<BusinessDay | null>(null);
  const [shiftSlots, setShiftSlots] = useState<ShiftSlot[]>([]);
  const [slotAssignments, setSlotAssignments] = useState<Record<string, ShiftAssignment[]>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showTemplateModal, setShowTemplateModal] = useState(false);

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
      setShiftSlots(shiftSlotsData.shift_slots || []);

      // 各シフト枠の割り当てを取得
      const assignmentsMap: Record<string, ShiftAssignment[]> = {};
      await Promise.all(
        shiftSlotsData.shift_slots.map(async (slot) => {
          try {
            const assignmentsData = await getAssignments({ slot_id: slot.slot_id, assignment_status: 'confirmed' });
            assignmentsMap[slot.slot_id] = assignmentsData.assignments || [];
          } catch (err) {
            assignmentsMap[slot.slot_id] = [];
          }
        })
      );
      setSlotAssignments(assignmentsMap);
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
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
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
        <div className="flex gap-2">
          <button
            onClick={() => setShowTemplateModal(true)}
            className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-lg flex items-center"
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
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            テンプレートから追加
          </button>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            ＋ シフト枠を追加
          </button>
        </div>
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
          {shiftSlots.map((slot) => {
            const assignments = slotAssignments[slot.slot_id] || [];
            return (
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
                    {assignments.length > 0 && (
                      <div className="mt-3">
                        <p className="text-xs text-gray-500 mb-1">割り当て済み:</p>
                        <div className="flex flex-wrap gap-1">
                          {assignments.map((assignment) => (
                            <span
                              key={assignment.assignment_id}
                              className="inline-block px-2 py-1 bg-accent/10 text-accent-dark rounded text-xs"
                            >
                              {assignment.member_display_name || assignment.member_id}
                            </span>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                  <button
                    onClick={() => navigate(`/shift-slots/${slot.slot_id}/assign`)}
                    className="btn-primary ml-4"
                  >
                    {(slot.assigned_count || 0) >= slot.required_count ? '編集' : '割り当て'}
                  </button>
                </div>
              </div>
            );
          })}
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

      {/* テンプレート適用モーダル */}
      {showTemplateModal && businessDayId && businessDay && (
        <ApplyTemplateModal
          businessDayId={businessDayId}
          eventId={businessDay.event_id}
          onClose={() => setShowTemplateModal(false)}
          onSuccess={() => {
            setShowTemplateModal(false);
            loadData();
          }}
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
      <div className="bg-white rounded-lg max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
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
              placeholder="例: 受付、案内、配信、MC など"
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

// テンプレート適用モーダルコンポーネント
function ApplyTemplateModal({
  businessDayId,
  eventId,
  onClose,
  onSuccess,
}: {
  businessDayId: string;
  eventId: string;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [selectedTemplateId, setSelectedTemplateId] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingTemplates, setLoadingTemplates] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadTemplates();
  }, [eventId]);

  const loadTemplates = async () => {
    try {
      setLoadingTemplates(true);
      const data = await listTemplates(eventId);
      setTemplates(data || []);
    } catch (err) {
      console.error('Failed to load templates:', err);
      setTemplates([]);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('テンプレート一覧の取得に失敗しました');
      }
    } finally {
      setLoadingTemplates(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!selectedTemplateId) {
      setError('テンプレートを選択してください');
      return;
    }

    setLoading(true);

    try {
      await applyTemplateToBusinessDay(businessDayId, selectedTemplateId);
      onSuccess();
    } catch (err) {
      console.error('Failed to apply template:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('テンプレートの適用に失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  const selectedTemplate = templates.find((t) => t.template_id === selectedTemplateId);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
        <h3 className="text-xl font-bold text-gray-900 mb-4">テンプレートから追加</h3>

        {loadingTemplates ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
            <p className="mt-4 text-gray-600">読み込み中...</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label htmlFor="templateSelect" className="label">
                テンプレートを選択 <span className="text-red-500">*</span>
              </label>
              {templates.length === 0 ? (
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <p className="text-sm text-yellow-800">
                    テンプレートがまだ作成されていません。
                    <br />
                    テンプレート管理ページから先にテンプレートを作成してください。
                  </p>
                </div>
              ) : (
                <select
                  id="templateSelect"
                  value={selectedTemplateId}
                  onChange={(e) => setSelectedTemplateId(e.target.value)}
                  className="input-field"
                  disabled={loading}
                  autoFocus
                >
                  <option value="">テンプレートを選択してください</option>
                  {templates.map((template) => (
                    <option key={template.template_id} value={template.template_id}>
                      {template.template_name} ({(template.items || []).length}個のシフト枠)
                    </option>
                  ))}
                </select>
              )}
            </div>

            {selectedTemplate && (
              <div className="mb-4 bg-accent/10 border border-accent/30 rounded-lg p-4">
                <h4 className="font-semibold text-accent-dark mb-2">
                  {selectedTemplate.template_name}
                </h4>
                {selectedTemplate.description && (
                  <p className="text-sm text-accent-dark mb-3">{selectedTemplate.description}</p>
                )}
                <div className="space-y-2">
                  <p className="text-xs font-semibold text-accent-dark">作成されるシフト枠:</p>
                  {(selectedTemplate.items || []).map((item, index) => (
                    <div key={index} className="text-xs text-accent-dark">
                      • {item.slot_name} ({item.instance_name}) - {item.start_time.substring(0, 5)}~
                      {item.end_time.substring(0, 5)} ({item.required_count}名)
                    </div>
                  ))}
                </div>
              </div>
            )}

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
                disabled={loading || !selectedTemplateId || templates.length === 0}
              >
                {loading ? '適用中...' : 'テンプレートを適用'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}

