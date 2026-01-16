import { useState, useEffect } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { getBusinessDayDetail, getShiftSlots, createShiftSlot, getAssignments, applyTemplateToBusinessDay } from '../lib/api';
import { listTemplates } from '../lib/api/templateApi';
import { listInstances } from '../lib/api/instanceApi';
import type { BusinessDay, ShiftSlot, ShiftAssignment, Template } from '../types/api';
import type { Instance } from '../lib/api/instanceApi';
import { ApiClientError } from '../lib/apiClient';
import type { InstanceData } from '../lib/shiftTextExport';
import ShiftTextPreviewModal from '../components/ShiftTextPreviewModal';

interface InstanceWithSlots {
  instance: Instance | null; // null for slots without instance
  slots: ShiftSlot[];
}

export default function ShiftSlotList() {
  const { businessDayId } = useParams<{ businessDayId: string }>();
  const navigate = useNavigate();
  const [businessDay, setBusinessDay] = useState<BusinessDay | null>(null);
  const [shiftSlots, setShiftSlots] = useState<ShiftSlot[]>([]);
  const [instances, setInstances] = useState<Instance[]>([]);
  const [slotAssignments, setSlotAssignments] = useState<Record<string, ShiftAssignment[]>>({});
  const [expandedInstances, setExpandedInstances] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showTemplateModal, setShowTemplateModal] = useState(false);
  const [showPreviewModal, setShowPreviewModal] = useState(false);

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

      // イベントのインスタンス一覧を取得
      if (businessDayData.event_id) {
        try {
          const instancesData = await listInstances(businessDayData.event_id);
          setInstances(instancesData || []);
        } catch {
          setInstances([]);
        }
      }

      // 各シフト枠の割り当てを取得
      const assignmentsMap: Record<string, ShiftAssignment[]> = {};
      await Promise.all(
        shiftSlotsData.shift_slots.map(async (slot) => {
          try {
            const assignmentsData = await getAssignments({ slot_id: slot.slot_id, assignment_status: 'confirmed' });
            assignmentsMap[slot.slot_id] = assignmentsData.assignments || [];
          } catch {
            assignmentsMap[slot.slot_id] = [];
          }
        })
      );
      setSlotAssignments(assignmentsMap);

      // 最初は全てのインスタンスを展開
      const allInstanceIds = new Set<string>();
      shiftSlotsData.shift_slots.forEach((slot) => {
        if (slot.instance_id) {
          allInstanceIds.add(slot.instance_id);
        } else {
          allInstanceIds.add('__no_instance__');
        }
      });
      setExpandedInstances(allInstanceIds);
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

  const toggleInstance = (instanceId: string) => {
    setExpandedInstances((prev) => {
      const next = new Set(prev);
      if (next.has(instanceId)) {
        next.delete(instanceId);
      } else {
        next.add(instanceId);
      }
      return next;
    });
  };

  // シフト枠をインスタンスごとにグループ化
  const groupSlotsByInstance = (): InstanceWithSlots[] => {
    const instanceMap = new Map<string, Instance>();
    instances.forEach((inst) => instanceMap.set(inst.instance_id, inst));

    const groups = new Map<string, InstanceWithSlots>();

    shiftSlots.forEach((slot) => {
      const instanceId = slot.instance_id || '__no_instance__';
      if (!groups.has(instanceId)) {
        groups.set(instanceId, {
          instance: slot.instance_id ? instanceMap.get(slot.instance_id) || null : null,
          slots: [],
        });
      }
      groups.get(instanceId)!.slots.push(slot);
    });

    // インスタンスの表示順でソート
    const result = Array.from(groups.values());
    result.sort((a, b) => {
      if (a.instance === null && b.instance === null) return 0;
      if (a.instance === null) return 1;
      if (b.instance === null) return -1;
      return a.instance.display_order - b.instance.display_order;
    });

    // 各インスタンス内のスロットをpriority昇順でソート（小さいほど優先）
    // バックエンドでもソート済みだが、フロントエンドでも一貫性を保証
    result.forEach((group) => {
      group.slots.sort((a, b) => a.priority - b.priority);
    });

    return result;
  };

  // プレビューモーダル用のインスタンスデータを生成
  const getInstanceDataForPreview = (): InstanceData[] => {
    const groups = groupSlotsByInstance();
    return groups.map((group) => ({
      instanceName: group.instance?.name || '未分類',
      slots: group.slots.map((slot) => ({
        slotName: slot.slot_name,
        assignments: (slotAssignments[slot.slot_id] || []).map((a) => ({
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

  if (!businessDay) {
    return (
      <div className="card text-center py-12">
        <p className="text-gray-600">営業日が見つかりません</p>
      </div>
    );
  }

  const instanceGroups = groupSlotsByInstance();

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
        <div className="flex gap-2 items-center">
          {/* インスタンス表プレビューボタン */}
          <button
            onClick={() => setShowPreviewModal(true)}
            disabled={shiftSlots.length === 0}
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
            + シフト枠を追加
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
          {instanceGroups.map((group) => {
            const instanceId = group.instance?.instance_id || '__no_instance__';
            const isExpanded = expandedInstances.has(instanceId);
            const instanceName = group.instance?.name || '未分類';

            // インスタンス内の統計
            const totalRequired = group.slots.reduce((sum, slot) => sum + slot.required_count, 0);
            const totalAssigned = group.slots.reduce((sum, slot) => sum + (slot.assigned_count || 0), 0);

            return (
              <div key={instanceId} className="card p-0 overflow-hidden">
                {/* インスタンスヘッダー（クリックで開閉） */}
                <button
                  onClick={() => toggleInstance(instanceId)}
                  className="w-full px-6 py-4 flex items-center justify-between bg-gray-50 hover:bg-gray-100 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <svg
                      className={`w-5 h-5 text-gray-500 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                    <div className="text-left">
                      <h3 className="text-lg font-bold text-gray-900">{instanceName}</h3>
                      <p className="text-sm text-gray-500">
                        {group.slots.length}個の役職
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className={`px-3 py-1 text-sm font-semibold rounded-full ${
                        totalAssigned >= totalRequired
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {totalAssigned} / {totalRequired} 人
                    </span>
                  </div>
                </button>

                {/* シフト枠リスト（展開時のみ表示） */}
                {isExpanded && (
                  <div className="divide-y divide-gray-100">
                    {group.slots.map((slot) => {
                      const assignments = slotAssignments[slot.slot_id] || [];
                      return (
                        <div key={slot.slot_id} className="px-6 py-4 hover:bg-gray-50">
                          <div className="flex justify-between items-start">
                            <div className="flex-1">
                              <h4 className="text-base font-semibold text-gray-900">
                                {slot.slot_name}
                              </h4>
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
                              className="btn-primary ml-4 text-sm"
                            >
                              {(slot.assigned_count || 0) >= slot.required_count ? '編集' : '割り当て'}
                            </button>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* シフト枠作成モーダル */}
      {showCreateModal && businessDayId && businessDay && (
        <CreateShiftSlotModal
          businessDayId={businessDayId}
          instances={instances}
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

      {/* インスタンス表プレビューモーダル */}
      <ShiftTextPreviewModal
        isOpen={showPreviewModal}
        onClose={() => setShowPreviewModal(false)}
        instanceData={getInstanceDataForPreview()}
      />
    </div>
  );
}

// シフト枠作成モーダルコンポーネント
function CreateShiftSlotModal({
  businessDayId,
  instances,
  onClose,
  onSuccess,
}: {
  businessDayId: string;
  instances: Instance[];
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [slotName, setSlotName] = useState('');
  const [selectedInstanceId, setSelectedInstanceId] = useState('');
  const [instanceName, setInstanceName] = useState('');
  const [startTime, setStartTime] = useState('21:30');
  const [endTime, setEndTime] = useState('23:00');
  const [requiredCount, setRequiredCount] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // 選択されたインスタンスが変更されたらinstanceNameを更新
  useEffect(() => {
    if (selectedInstanceId === '__new__') {
      setInstanceName('');
    } else if (selectedInstanceId) {
      const instance = instances.find((i) => i.instance_id === selectedInstanceId);
      if (instance) {
        setInstanceName(instance.name);
      }
    }
  }, [selectedInstanceId, instances]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    const finalInstanceName = selectedInstanceId === '__new__' ? instanceName.trim() : instanceName;

    if (!slotName.trim() || !finalInstanceName) {
      setError('役職名とインスタンスを入力してください');
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
        instance_name: finalInstanceName,
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
            <label htmlFor="instanceSelect" className="label">
              インスタンス <span className="text-red-500">*</span>
            </label>
            {instances.length > 0 ? (
              <select
                id="instanceSelect"
                value={selectedInstanceId}
                onChange={(e) => setSelectedInstanceId(e.target.value)}
                className="input-field"
                disabled={loading}
              >
                <option value="">インスタンスを選択してください</option>
                {instances.map((instance) => (
                  <option key={instance.instance_id} value={instance.instance_id}>
                    {instance.name}
                  </option>
                ))}
                <option value="__new__">+ 新しいインスタンスを作成</option>
              </select>
            ) : (
              <>
                <input type="hidden" value="__new__" />
                <p className="text-sm text-gray-500 mb-2">インスタンスがまだありません。新規作成してください。</p>
              </>
            )}
            {(selectedInstanceId === '__new__' || instances.length === 0) && (
              <input
                type="text"
                value={instanceName}
                onChange={(e) => setInstanceName(e.target.value)}
                placeholder="新しいインスタンス名（例: 第一インスタンス）"
                className="input-field mt-2"
                disabled={loading}
              />
            )}
          </div>

          <div className="mb-4">
            <label htmlFor="slotName" className="label">
              役職名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="slotName"
              value={slotName}
              onChange={(e) => setSlotName(e.target.value)}
              placeholder="例: インスタンスリーダー、サポート、MC など"
              className="input-field"
              disabled={loading}
              autoFocus
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
              disabled={loading || !slotName.trim() || (!instanceName && selectedInstanceId !== '__new__')}
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
                      ・ {item.instance_name} / {item.slot_name} - {item.start_time.substring(0, 5)}~
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
