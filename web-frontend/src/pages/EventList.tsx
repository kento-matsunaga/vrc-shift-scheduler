import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { SEO } from '../components/seo';
import GenerateBusinessDaysModal from '../components/GenerateBusinessDaysModal';
import {
  getEvents,
  createEvent,
  generateBusinessDays,
  updateEvent,
  getEventGroupAssignments,
  updateEventGroupAssignments,
  getMemberGroups,
  getRoleGroups,
} from '../lib/api';
import type { Event } from '../types/api';
import type { MemberGroup } from '../lib/api/memberGroupApi';
import type { RoleGroup } from '../lib/api/roleGroupApi';
import { ApiClientError } from '../lib/apiClient';

// 曜日名の配列
const DAY_NAMES = ['日', '月', '火', '水', '木', '金', '土'];

// 定期タイプの表示名
const RECURRENCE_TYPE_LABELS: Record<string, string> = {
  none: 'なし',
  weekly: '毎週',
  biweekly: '隔週',
};

export default function EventList() {
  const navigate = useNavigate();
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [generatingEvent, setGeneratingEvent] = useState<Event | null>(null);
  const [generateLoading, setGenerateLoading] = useState(false);
  const [editingEventId, setEditingEventId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [savingEventId, setSavingEventId] = useState<string | null>(null);
  const editInputRef = useRef<HTMLInputElement>(null);
  const [showGroupModal, setShowGroupModal] = useState(false);
  const [selectedEventForGroups, setSelectedEventForGroups] = useState<Event | null>(null);

  useEffect(() => {
    loadEvents();
  }, []);

  const loadEvents = async () => {
    try {
      setLoading(true);
      const data = await getEvents({ is_active: true });
      setEvents(data.events || []);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベント一覧の取得に失敗しました');
      }
      console.error('Failed to load events:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    loadEvents();
  };

  const handleGenerateClick = (event: Event, e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setGeneratingEvent(event);
  };

  const handleGenerateConfirm = async (months: number) => {
    if (!generatingEvent) return;

    setGenerateLoading(true);
    setError('');
    setSuccess('');

    try {
      const result = await generateBusinessDays(generatingEvent.event_id, months);
      setSuccess(result.message || `${months}ヶ月分の営業日を生成しました`);
      setGeneratingEvent(null);
      setTimeout(() => setSuccess(''), 5000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('営業日の生成に失敗しました');
      }
      console.error('Failed to generate business days:', err);
    } finally {
      setGenerateLoading(false);
    }
  };

  // 編集モードを開始
  const handleStartEdit = (event: Event, e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setEditingEventId(event.event_id);
    setEditingName(event.event_name);
    setError('');
    // 次のレンダリングでinputにフォーカス
    setTimeout(() => editInputRef.current?.focus(), 0);
  };

  // 編集をキャンセル
  const handleCancelEdit = (e?: React.MouseEvent | React.KeyboardEvent) => {
    e?.preventDefault();
    e?.stopPropagation();
    setEditingEventId(null);
    setEditingName('');
  };

  // 編集を保存
  const handleSaveEdit = async (eventId: string, e?: React.MouseEvent | React.FormEvent) => {
    e?.preventDefault();
    e?.stopPropagation();

    if (!editingName.trim()) {
      setError('イベント名を入力してください');
      return;
    }

    setSavingEventId(eventId);
    setError('');

    try {
      await updateEvent(eventId, { event_name: editingName.trim() });
      // ローカルのevents配列を更新
      setEvents(events.map(ev =>
        ev.event_id === eventId ? { ...ev, event_name: editingName.trim() } : ev
      ));
      setEditingEventId(null);
      setEditingName('');
      setSuccess('イベント名を更新しました');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベント名の更新に失敗しました');
      }
      console.error('Failed to update event:', err);
    } finally {
      setSavingEventId(null);
    }
  };

  // キーボードイベント処理
  const handleEditKeyDown = (eventId: string, e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveEdit(eventId, e);
    } else if (e.key === 'Escape') {
      handleCancelEdit(e);
    }
  };

  // グループ設定モーダルを開く
  const handleOpenGroupModal = (event: Event, e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setSelectedEventForGroups(event);
    setShowGroupModal(true);
  };

  // グループ設定モーダルを閉じる
  const handleCloseGroupModal = () => {
    setShowGroupModal(false);
    setSelectedEventForGroups(null);
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div>
      <SEO noindex={true} />
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-6">
        <h2 className="text-xl sm:text-2xl font-bold text-gray-900">イベント一覧</h2>
        <button id="btn-create-event" onClick={() => setShowCreateModal(true)} className="btn-primary text-sm sm:text-base w-full sm:w-auto">
          ＋ 新しいイベント
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {success && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-green-800">{success}</p>
        </div>
      )}

      {events.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだイベントがありません</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初のイベントを作成
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {events.map((event) => (
            <div
              key={event.event_id}
              className="card hover:shadow-lg transition-shadow cursor-pointer"
              onClick={() => {
                // 編集中でなければ遷移
                if (editingEventId !== event.event_id) {
                  navigate(`/events/${event.event_id}/business-days`);
                }
              }}
            >
              {/* バッジ類 */}
              <div className="mb-2 flex flex-wrap gap-2">
                <span
                  className={`inline-block px-2 py-1 text-xs font-semibold rounded ${
                    event.event_type === 'normal'
                      ? 'bg-accent/10 text-accent-dark'
                      : 'bg-purple-100 text-purple-800'
                  }`}
                >
                  {event.event_type === 'normal' ? '通常イベント' : '特別イベント'}
                </span>
                {event.recurrence_type !== 'none' && (
                  <span className="inline-block px-2 py-1 text-xs font-semibold rounded bg-green-100 text-green-800">
                    {RECURRENCE_TYPE_LABELS[event.recurrence_type]}
                    {event.recurrence_day_of_week !== undefined && `（${DAY_NAMES[event.recurrence_day_of_week]}）`}
                  </span>
                )}
              </div>

              {/* イベント名（編集可能） */}
              {editingEventId === event.event_id ? (
                <div className="mb-2" onClick={(e) => e.stopPropagation()}>
                  <div className="flex items-center gap-2">
                    <input
                      ref={editInputRef}
                      type="text"
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                      onKeyDown={(e) => handleEditKeyDown(event.event_id, e)}
                      className="flex-1 px-2 py-1 text-lg font-bold border border-accent/50 rounded focus:outline-none focus:ring-2 focus:ring-accent"
                      disabled={savingEventId === event.event_id}
                    />
                    <button
                      onClick={(e) => handleSaveEdit(event.event_id, e)}
                      disabled={savingEventId === event.event_id}
                      className="p-1 text-green-600 hover:text-green-800 disabled:opacity-50"
                      title="保存"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      disabled={savingEventId === event.event_id}
                      className="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50"
                      title="キャンセル"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-2 mb-2">
                  <h3 className="text-xl font-bold text-gray-900">
                    {event.event_name}
                  </h3>
                  <button
                    onClick={(e) => handleStartEdit(event, e)}
                    className="p-1 text-gray-400 hover:text-accent transition-colors"
                    title="イベント名を編集"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                    </svg>
                  </button>
                </div>
              )}

              {/* イベント詳細 */}
              <p className="text-sm text-gray-600 mb-2">{event.description || '説明なし'}</p>
              {event.recurrence_type !== 'none' && event.default_start_time && event.default_end_time && (
                <p className="text-xs text-gray-500 mb-2">
                  時間: {event.default_start_time.slice(0, 5)}〜{event.default_end_time.slice(0, 5)}
                </p>
              )}
              <div className="text-xs text-gray-500 mb-3">
                作成日: {new Date(event.created_at).toLocaleDateString('ja-JP')}
              </div>

              <div className="flex gap-2 mt-2">
                <button
                  onClick={(e) => handleOpenGroupModal(event, e)}
                  className="flex-1 px-3 py-2 text-sm text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
                  title="対象グループを設定"
                >
                  グループ設定
                </button>
                {event.recurrence_type !== 'none' && (
                  <button
                    onClick={(e) => handleGenerateClick(event, e)}
                    className="flex-1 btn-secondary text-sm py-2"
                  >
                    営業日生成
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* イベント作成モーダル */}
      {showCreateModal && (
        <CreateEventModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}

      {/* グループ設定モーダル */}
      {showGroupModal && selectedEventForGroups && (
        <EventGroupModal
          event={selectedEventForGroups}
          onClose={handleCloseGroupModal}
          onSuccess={() => {
            setSuccess('グループ設定を更新しました');
            setTimeout(() => setSuccess(''), 3000);
          }}
        />
      )}

      {/* 営業日生成モーダル */}
      {generatingEvent && (
        <GenerateBusinessDaysModal
          eventName={generatingEvent.event_name}
          onConfirm={handleGenerateConfirm}
          onCancel={() => setGeneratingEvent(null)}
          loading={generateLoading}
        />
      )}
    </div>
  );
}

// イベント作成モーダルコンポーネント
function CreateEventModal({
  onClose,
  onSuccess,
}: {
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [eventName, setEventName] = useState('');
  const [eventType, setEventType] = useState<'normal' | 'special'>('normal');
  const [description, setDescription] = useState('');
  const [recurrenceType, setRecurrenceType] = useState<'none' | 'weekly' | 'biweekly'>('none');
  const [recurrenceDayOfWeek, setRecurrenceDayOfWeek] = useState<number>(6); // デフォルト土曜
  const [recurrenceStartDate, setRecurrenceStartDate] = useState(() => {
    // デフォルトは今日
    return new Date().toISOString().split('T')[0];
  });
  const [defaultStartTime, setDefaultStartTime] = useState('21:00');
  const [defaultEndTime, setDefaultEndTime] = useState('23:30');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!eventName.trim()) {
      setError('イベント名を入力してください');
      return;
    }

    // 定期イベントの場合のバリデーション
    if (recurrenceType !== 'none') {
      if (!defaultStartTime || !defaultEndTime) {
        setError('定期イベントの場合は開始時刻と終了時刻を入力してください');
        return;
      }
      if (recurrenceType === 'biweekly' && !recurrenceStartDate) {
        setError('隔週イベントの場合は開始日を入力してください');
        return;
      }
    }

    setLoading(true);

    try {
      const requestData: Parameters<typeof createEvent>[0] = {
        event_name: eventName.trim(),
        event_type: eventType,
        description: description.trim(),
        recurrence_type: recurrenceType,
      };

      // 定期イベントの場合のみ追加フィールドを送信
      if (recurrenceType !== 'none') {
        requestData.recurrence_day_of_week = recurrenceDayOfWeek;
        requestData.default_start_time = defaultStartTime + ':00'; // HH:MM:SS形式
        requestData.default_end_time = defaultEndTime + ':00';
        requestData.recurrence_start_date = recurrenceStartDate;
      }

      await createEvent(requestData);
      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベントの作成に失敗しました');
      }
      console.error('Failed to create event:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6 max-h-[calc(100dvh-2rem)] overflow-y-auto">
        <h3 className="text-xl font-bold text-gray-900 mb-4">新しいイベントを作成</h3>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="eventName" className="label">
              イベント名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="eventName"
              value={eventName}
              onChange={(e) => setEventName(e.target.value)}
              placeholder="例: VRChat 交流会"
              className="input-field"
              disabled={loading}
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label htmlFor="eventType" className="label">
              イベント種別
            </label>
            <select
              id="eventType"
              value={eventType}
              onChange={(e) => setEventType(e.target.value as 'normal' | 'special')}
              className="input-field"
              disabled={loading}
            >
              <option value="normal">通常イベント</option>
              <option value="special">特別イベント</option>
            </select>
          </div>

          <div className="mb-4">
            <label htmlFor="recurrenceType" className="label">
              開催頻度
            </label>
            <select
              id="recurrenceType"
              value={recurrenceType}
              onChange={(e) => setRecurrenceType(e.target.value as 'none' | 'weekly' | 'biweekly')}
              className="input-field"
              disabled={loading}
            >
              <option value="none">不定期（都度設定）</option>
              <option value="weekly">毎週</option>
              <option value="biweekly">隔週</option>
            </select>
          </div>

          {/* 定期イベントの場合のみ表示 */}
          {recurrenceType !== 'none' && (
            <>
              <div className="mb-4">
                <label htmlFor="recurrenceDayOfWeek" className="label">
                  開催曜日 <span className="text-red-500">*</span>
                </label>
                <select
                  id="recurrenceDayOfWeek"
                  value={recurrenceDayOfWeek}
                  onChange={(e) => setRecurrenceDayOfWeek(Number(e.target.value))}
                  className="input-field"
                  disabled={loading}
                >
                  {DAY_NAMES.map((name, index) => (
                    <option key={index} value={index}>{name}曜日</option>
                  ))}
                </select>
              </div>

              {/* 隔週の場合は開始日を必須で表示 */}
              {recurrenceType === 'biweekly' && (
                <div className="mb-4">
                  <label htmlFor="recurrenceStartDate" className="label">
                    開始日（基準日） <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="date"
                    id="recurrenceStartDate"
                    value={recurrenceStartDate}
                    onChange={(e) => setRecurrenceStartDate(e.target.value)}
                    className="input-field"
                    disabled={loading}
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    この日付を基準に隔週で営業日が生成されます
                  </p>
                </div>
              )}

              <div className="mb-4 grid grid-cols-2 gap-4">
                <div>
                  <label htmlFor="defaultStartTime" className="label">
                    開始時刻 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="time"
                    id="defaultStartTime"
                    value={defaultStartTime}
                    onChange={(e) => setDefaultStartTime(e.target.value)}
                    className="input-field"
                    disabled={loading}
                  />
                </div>
                <div>
                  <label htmlFor="defaultEndTime" className="label">
                    終了時刻 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="time"
                    id="defaultEndTime"
                    value={defaultEndTime}
                    onChange={(e) => setDefaultEndTime(e.target.value)}
                    className="input-field"
                    disabled={loading}
                  />
                </div>
              </div>

              <div className="mb-4 p-3 bg-accent/10 border border-accent/30 rounded-lg">
                <p className="text-sm text-accent-dark">
                  定期イベントを作成後、「営業日を自動生成」ボタンで今月〜来月の営業日をまとめて作成できます。
                </p>
              </div>
            </>
          )}

          <div className="mb-4">
            <label htmlFor="description" className="label">
              説明
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="イベントの説明を入力"
              className="input-field"
              rows={3}
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
              id="btn-submit-event"
              type="submit"
              className="flex-1 btn-primary"
              disabled={loading || !eventName.trim()}
            >
              {loading ? '作成中...' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// イベントグループ設定モーダルコンポーネント
function EventGroupModal({
  event,
  onClose,
  onSuccess,
}: {
  event: Event;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [memberGroups, setMemberGroups] = useState<MemberGroup[]>([]);
  const [roleGroups, setRoleGroups] = useState<RoleGroup[]>([]);
  const [selectedMemberGroups, setSelectedMemberGroups] = useState<string[]>([]);
  const [selectedRoleGroups, setSelectedRoleGroups] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- 初回マウント時のみ実行（loadDataは関数定義のため除外）
  }, [event.event_id]);

  const loadData = async () => {
    try {
      setLoading(true);
      // 並列でデータを取得
      const [memberGroupsData, roleGroupsData, assignmentsData] = await Promise.all([
        getMemberGroups(),
        getRoleGroups(),
        getEventGroupAssignments(event.event_id),
      ]);

      setMemberGroups(memberGroupsData.groups || []);
      setRoleGroups(roleGroupsData.groups || []);
      setSelectedMemberGroups(assignmentsData.member_group_ids || []);
      setSelectedRoleGroups(assignmentsData.role_group_ids || []);
    } catch (err) {
      console.error('Failed to load group data:', err);
      setError('グループデータの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setError('');

    try {
      await updateEventGroupAssignments(event.event_id, {
        member_group_ids: selectedMemberGroups,
        role_group_ids: selectedRoleGroups,
      });
      onSuccess();
      onClose();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('グループ設定の保存に失敗しました');
      }
      console.error('Failed to save group assignments:', err);
    } finally {
      setSaving(false);
    }
  };

  const toggleMemberGroup = (groupId: string) => {
    setSelectedMemberGroups((prev) =>
      prev.includes(groupId)
        ? prev.filter((id) => id !== groupId)
        : [...prev, groupId]
    );
  };

  const toggleRoleGroup = (groupId: string) => {
    setSelectedRoleGroups((prev) =>
      prev.includes(groupId)
        ? prev.filter((id) => id !== groupId)
        : [...prev, groupId]
    );
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-lg w-full p-6 max-h-[calc(100dvh-2rem)] overflow-y-auto">
        <h3 className="text-xl font-bold text-gray-900 mb-2">グループ設定</h3>
        <p className="text-sm text-gray-600 mb-4">
          「{event.event_name}」に参加可能なグループを設定します
        </p>

        {loading ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-accent mx-auto"></div>
            <p className="mt-2 text-gray-600 text-sm">読み込み中...</p>
          </div>
        ) : (
          <>
            {/* メンバーグループ選択 */}
            <div className="mb-6">
              <h4 className="text-sm font-semibold text-gray-700 mb-2">
                メンバーグループ
              </h4>
              <p className="text-xs text-gray-500 mb-2">
                選択したグループに所属するメンバーのみが対象になります（未選択の場合は全メンバー）
              </p>
              {memberGroups.length === 0 ? (
                <p className="text-sm text-gray-500 py-2">
                  メンバーグループがありません
                </p>
              ) : (
                <div className="space-y-2 max-h-40 overflow-y-auto border border-gray-200 rounded-md p-2">
                  {memberGroups.map((group) => (
                    <label
                      key={group.group_id}
                      className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-1 rounded"
                    >
                      <input
                        type="checkbox"
                        checked={selectedMemberGroups.includes(group.group_id)}
                        onChange={() => toggleMemberGroup(group.group_id)}
                        className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                      />
                      <span
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: group.color || '#6B7280' }}
                      />
                      <span className="text-sm text-gray-900">{group.name}</span>
                    </label>
                  ))}
                </div>
              )}
            </div>

            {/* ロールグループ選択 */}
            <div className="mb-6">
              <h4 className="text-sm font-semibold text-gray-700 mb-2">
                ロールグループ
              </h4>
              <p className="text-xs text-gray-500 mb-2">
                選択したロールグループに属するロールを持つメンバーのみが対象になります（未選択の場合は全ロール）
              </p>
              {roleGroups.length === 0 ? (
                <p className="text-sm text-gray-500 py-2">
                  ロールグループがありません
                </p>
              ) : (
                <div className="space-y-2 max-h-40 overflow-y-auto border border-gray-200 rounded-md p-2">
                  {roleGroups.map((group) => (
                    <label
                      key={group.group_id}
                      className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-1 rounded"
                    >
                      <input
                        type="checkbox"
                        checked={selectedRoleGroups.includes(group.group_id)}
                        onChange={() => toggleRoleGroup(group.group_id)}
                        className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                      />
                      <span
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: group.color || '#6B7280' }}
                      />
                      <span className="text-sm text-gray-900">{group.name}</span>
                    </label>
                  ))}
                </div>
              )}
            </div>

            {/* 選択状態のサマリ */}
            <div className="mb-4 p-3 bg-gray-50 rounded-lg">
              <p className="text-xs text-gray-600">
                メンバーグループ: {selectedMemberGroups.length === 0 ? '全て' : `${selectedMemberGroups.length}件選択中`}
              </p>
              <p className="text-xs text-gray-600">
                ロールグループ: {selectedRoleGroups.length === 0 ? '全て' : `${selectedRoleGroups.length}件選択中`}
              </p>
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
                disabled={saving}
              >
                キャンセル
              </button>
              <button
                type="button"
                onClick={handleSave}
                className="flex-1 btn-primary"
                disabled={saving}
              >
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

