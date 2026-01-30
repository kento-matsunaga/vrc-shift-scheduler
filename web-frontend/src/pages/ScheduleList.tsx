import { useEffect, useState, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

import { SEO } from '../components/seo';
import { DateRangePicker, type DateInput } from '../components/DateRangePicker';
import { MobileCard, CardField, CardHeader } from '../components/MobileCard';

import {
  createSchedule,
  getSchedule,
  listSchedules,
  updateSchedule,
  type Schedule,
} from '../lib/api/scheduleApi';
import { getMemberGroups, type MemberGroup } from '../lib/api/memberGroupApi';
import { ApiClientError } from '../lib/apiClient';
import { isValidTimeRange, toApiTimeFormat } from '../lib/timeUtils';

// APIから返ってくる time を input[type="time"] 向けに HH:MM にする（"HH:MM:SS" や ISO を想定）
const formatTime = (t?: string | null) => {
  if (!t) return '';
  const m = String(t).match(/^(\d{2}:\d{2})/);
  return m ? m[1] : '';
};

// 候補日の入力データ型
interface CandidateDateInput {
  date: string; // YYYY-MM-DD形式
  startTime: string; // HH:MM形式（任意）
  endTime: string; // HH:MM形式（任意）
}

const emptyCandidateDate = (): CandidateDateInput => ({
  date: '',
  startTime: '',
  endTime: '',
});

export default function ScheduleList() {
  const navigate = useNavigate();

  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [showCreateForm, setShowCreateForm] = useState(false);

  const [isEditing, setIsEditing] = useState(false);
  const [editingScheduleId, setEditingScheduleId] = useState<string | null>(null);
  const [loadingEdit, setLoadingEdit] = useState(false);

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [deadline, setDeadline] = useState('');
  const [candidateDates, setCandidateDates] = useState<CandidateDateInput[]>([
    emptyCandidateDate(),
    emptyCandidateDate(),
    emptyCandidateDate(),
  ]);

  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const [error, setError] = useState('');
  const [conflictMessage, setConflictMessage] = useState('');
  const [pendingUpdatePayload, setPendingUpdatePayload] = useState<{
    scheduleId: string;
    title: string;
    description: string;
    candidates: { date: string; start_time?: string; end_time?: string }[];
    deadline?: string;
  } | null>(null);

  const [createdSchedule, setCreatedSchedule] = useState<Schedule | null>(null);
  const [publicUrl, setPublicUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const [submittedCandidatesCount, setSubmittedCandidatesCount] = useState(0);

  const [memberGroups, setMemberGroups] = useState<MemberGroup[]>([]);
  const [selectedGroupIds, setSelectedGroupIds] = useState<string[]>([]);

  const loadMemberGroups = useCallback(async () => {
    try {
      const data = await getMemberGroups();
      setMemberGroups(data.groups || []);
    } catch (err) {
      console.error('Failed to load member groups:', err);
    }
  }, []);

  const loadSchedules = useCallback(async () => {
    try {
      setLoading(true);
      const data = await listSchedules();
      setSchedules(data || []);
    } catch (err) {
      console.error('Failed to load schedules:', err);
      setError('日程調整一覧の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSchedules();
    loadMemberGroups();
  }, [loadSchedules, loadMemberGroups]);

  const handleAddDate = () => {
    setCandidateDates([...candidateDates, emptyCandidateDate()]);
  };

  const handleRemoveDate = (index: number) => {
    if (candidateDates.length > 1) {
      setCandidateDates(candidateDates.filter((_, i) => i !== index));
    }
  };

  const handleDateChange = (index: number, field: keyof CandidateDateInput, value: string) => {
    const newDates = [...candidateDates];
    newDates[index] = { ...newDates[index], [field]: value };
    setCandidateDates(newDates);
  };

  const toggleGroupSelection = (groupId: string) => {
    setSelectedGroupIds((prev) =>
      prev.includes(groupId) ? prev.filter((id) => id !== groupId) : [...prev, groupId],
    );
  };

  const resetForm = () => {
    setTitle('');
    setDescription('');
    setDeadline('');
    setCandidateDates([emptyCandidateDate(), emptyCandidateDate(), emptyCandidateDate()]);
    setSelectedGroupIds([]);
    setIsEditing(false);
    setEditingScheduleId(null);
    setConflictMessage('');
    setPendingUpdatePayload(null);
  };

  const toInputDate = (isoDate: string) => isoDate.split('T')[0];
  const toInputDateTime = (isoDate?: string) =>
    isoDate ? new Date(isoDate).toISOString().slice(0, 16) : '';

  // DateRangePickerからの一括追加
  const handleAddDatesFromPicker = (dates: DateInput[]) => {
    // 既存の空でない日付を保持
    const existingDates = candidateDates.filter((d) => d.date.trim() !== '');
    const existingDateStringsLocal = existingDates.map((d) => d.date);

    // 重複を除いて新しい日付を追加
    const newDates = dates.filter((d) => !existingDateStringsLocal.includes(d.date));

    // マージして日付順にソート
    const mergedDates = [...existingDates, ...newDates].sort((a, b) => a.date.localeCompare(b.date));

    // 日付がない場合は空欄を追加
    setCandidateDates(mergedDates.length > 0 ? mergedDates : [emptyCandidateDate()]);
  };

  // 既存の日付リスト（重複チェック用）- useMemoでメモ化
  const existingDateStrings = useMemo(() =>
    candidateDates
      .filter((d) => d.date.trim() !== '')
      .map((d) => d.date),
    [candidateDates]
  );

  const handleEditClick = async (scheduleId: string) => {
    setError('');
    setCreatedSchedule(null);
    setShowCreateForm(true);
    setLoadingEdit(true);

    try {
      const schedule = await getSchedule(scheduleId);
      setIsEditing(true);
      setEditingScheduleId(scheduleId);

      setTitle(schedule.title);
      setDescription(schedule.description || '');
      setDeadline(toInputDateTime(schedule.deadline));

      const candidates = schedule.candidates ?? [];
      setCandidateDates(
        candidates.length > 0
          ? candidates.map((candidate) => ({
              date: toInputDate(candidate.date),
              startTime: formatTime(candidate.start_time),
              endTime: formatTime(candidate.end_time),
            }))
          : [emptyCandidateDate()],
      );
    } catch (err) {
      console.error('Failed to load schedule for edit:', err);
      setError('日程調整の取得に失敗しました');
    } finally {
      setLoadingEdit(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setError('');
    setCreatedSchedule(null);
    setConflictMessage('');
    setPendingUpdatePayload(null);

    if (!title.trim()) {
      setError('タイトルを入力してください');
      return;
    }

    const validDates = candidateDates.filter((d) => d.date.trim() !== '');
    if (validDates.length === 0) {
      setError('候補日を1つ以上入力してください');
      return;
    }

    // 時間バリデーション
    for (const candidate of validDates) {
      if (!isValidTimeRange(candidate.startTime, candidate.endTime)) {
        setError('開始時間は終了時間より前に設定してください');
        return;
      }
    }

    const candidatePayload = validDates.map((d) => ({
      date: new Date(d.date).toISOString(),
      start_time: toApiTimeFormat(d.startTime),
      end_time: toApiTimeFormat(d.endTime),
    }));

    setSubmitting(true);

    try {
      setSubmittedCandidatesCount(validDates.length);
      const basePayload = {
        title: title.trim(),
        description: description.trim(),
        candidates: candidatePayload,
        deadline: deadline ? new Date(deadline).toISOString() : undefined,
      };

      const result =
        isEditing && editingScheduleId
          ? await updateSchedule(editingScheduleId, basePayload)
          : await createSchedule({
              ...basePayload,
              group_ids: selectedGroupIds.length > 0 ? selectedGroupIds : undefined,
            });

      const baseUrl = window.location.origin;
      if (!isEditing) {
        const url = `${baseUrl}/p/schedule/${result.public_token}`;
        setPublicUrl(url);
        setCreatedSchedule(result);
      }

      resetForm();
      setShowCreateForm(false);

      loadSchedules();
    } catch (err) {
      if (err instanceof ApiClientError && err.isConflictError() && editingScheduleId) {
        const message = err.getUserMessage();
        setConflictMessage(message);
        setPendingUpdatePayload({
          scheduleId: editingScheduleId,
          title: title.trim(),
          description: description.trim(),
          candidates: candidatePayload,
          deadline: deadline ? new Date(deadline).toISOString() : undefined,
        });
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError(isEditing ? '日程調整の更新に失敗しました' : '日程調整の作成に失敗しました');
      }
      console.error('Create schedule error:', err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleConfirmForceDelete = async () => {
    if (!pendingUpdatePayload) return;

    setSubmitting(true);
    setError('');

    try {
      await updateSchedule(pendingUpdatePayload.scheduleId, {
        title: pendingUpdatePayload.title,
        description: pendingUpdatePayload.description,
        candidates: pendingUpdatePayload.candidates,
        deadline: pendingUpdatePayload.deadline,
        force_delete_candidate_responses: true,
      });

      resetForm();
      setShowCreateForm(false);
      loadSchedules();
    } catch (forceErr) {
      setError(forceErr instanceof Error ? forceErr.message : '日程調整の更新に失敗しました');
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancelForceDelete = () => {
    setConflictMessage('');
    setPendingUpdatePayload(null);
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

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'open':
        return (
          <span className="px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
            受付中
          </span>
        );
      case 'decided':
        return (
          <span className="px-2 py-1 text-xs font-semibold rounded-full bg-accent/10 text-accent-dark">
            決定済み
          </span>
        );
      case 'closed':
        return (
          <span className="px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
            締切済み
          </span>
        );
      default:
        return (
          <span className="px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
            {status}
          </span>
        );
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

  return (
    <div className="max-w-6xl mx-auto">
      <SEO noindex={true} />

      <div className="mb-6 flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold text-gray-900">日程調整</h1>
          <p className="text-xs sm:text-sm text-gray-600 mt-1">
            複数の候補日から、メンバーが参加可能な日程を回答してもらいましょう
          </p>
        </div>

        <button
          onClick={() => {
            if (showCreateForm) {
              resetForm();
              setShowCreateForm(false);
            } else {
              setShowCreateForm(true);
            }
          }}
          className="px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent-dark transition-colors font-medium text-sm sm:text-base w-full sm:w-auto"
        >
          {showCreateForm ? (isEditing ? '編集をキャンセル' : 'キャンセル') : '+ 新規作成'}
        </button>
      </div>

      {showCreateForm && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            {isEditing ? '日程調整を編集' : '新しい日程調整を作成'}
          </h2>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                タイトル <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="例：忘年会の日程調整"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={submitting || loadingEdit}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">説明</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                placeholder="詳細な説明や注意事項を入力してください"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={submitting || loadingEdit}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                候補日 <span className="text-red-500">*</span>
              </label>

              <p className="text-xs text-gray-500 mb-3">
                時間は任意です。設定すると公開ページで回答者に表示されます。
              </p>

              {/* 期間から一括追加 */}
              <div className="mb-4">
                <DateRangePicker
                  onAddDates={handleAddDatesFromPicker}
                  existingDates={existingDateStrings}
                  disabled={submitting || loadingEdit}
                />
              </div>

              {/* 個別の候補日入力 */}
              <div className="space-y-3">
                {candidateDates.map((candidate, index) => (
                  <div key={index} className="p-3 bg-gray-50 rounded-lg border border-gray-200">
                    <div className="flex flex-col sm:flex-row gap-2">
                      <div className="flex-1">
                        <label className="block text-xs text-gray-600 mb-1">日付 *</label>
                        <input
                          type="date"
                          value={candidate.date}
                          onChange={(e) => handleDateChange(index, 'date', e.target.value)}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                          disabled={submitting || loadingEdit}
                        />
                      </div>

                      <div className="w-full sm:w-28">
                        <label className="block text-xs text-gray-600 mb-1">開始時間</label>
                        <input
                          type="time"
                          value={candidate.startTime}
                          onChange={(e) => handleDateChange(index, 'startTime', e.target.value)}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                          disabled={submitting || loadingEdit}
                        />
                      </div>

                      <div className="w-full sm:w-28">
                        <label className="block text-xs text-gray-600 mb-1">終了時間</label>
                        <input
                          type="time"
                          value={candidate.endTime}
                          onChange={(e) => handleDateChange(index, 'endTime', e.target.value)}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                          disabled={submitting || loadingEdit}
                        />
                      </div>

                      {candidateDates.length > 1 && (
                        <div className="flex items-end">
                          <button
                            type="button"
                            onClick={() => handleRemoveDate(index)}
                            className="px-3 py-2 text-red-600 hover:bg-red-100 rounded-md transition text-sm"
                            disabled={submitting || loadingEdit}
                            aria-label={`候補日${index + 1}を削除`}
                          >
                            削除
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>

              <button
                type="button"
                onClick={handleAddDate}
                className="mt-2 px-3 py-1 text-sm text-accent hover:bg-accent/10 rounded-md transition"
                disabled={submitting || loadingEdit}
              >
                + 候補日を追加
              </button>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">回答締切（任意）</label>
              <input
                type="datetime-local"
                value={deadline}
                onChange={(e) => setDeadline(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={submitting || loadingEdit}
              />
            </div>

            {!isEditing && memberGroups.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  対象メンバーグループ（任意）
                </label>
                <p className="text-xs text-gray-500 mb-2">
                  選択すると、そのグループに属するメンバーのみが回答可能になります
                </p>

                <div className="flex flex-wrap gap-2">
                  {memberGroups.map((group) => (
                    <button
                      key={group.group_id}
                      type="button"
                      onClick={() => toggleGroupSelection(group.group_id)}
                      disabled={submitting || loadingEdit}
                      className={`px-3 py-1.5 rounded-full text-sm font-medium transition ${
                        selectedGroupIds.includes(group.group_id)
                          ? 'bg-accent text-white'
                          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      }`}
                      style={
                        selectedGroupIds.includes(group.group_id) && group.color
                          ? { backgroundColor: group.color }
                          : undefined
                      }
                    >
                      {group.name}
                    </button>
                  ))}
                </div>

                {selectedGroupIds.length > 0 && (
                  <p className="mt-2 text-xs text-accent">{selectedGroupIds.length}個のグループを選択中</p>
                )}
              </div>
            )}

            {(error || conflictMessage) && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3 space-y-3">
                {error && <p className="text-sm text-red-800">{error}</p>}

                {conflictMessage && (
                  <div className="space-y-2">
                    <p className="text-sm text-red-800">{conflictMessage}</p>

                    <div className="flex flex-col sm:flex-row gap-2">
                      <button
                        type="button"
                        onClick={handleConfirmForceDelete}
                        className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition text-sm"
                        disabled={submitting}
                      >
                        はい
                      </button>
                      <button
                        type="button"
                        onClick={handleCancelForceDelete}
                        className="px-4 py-2 bg-white text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50 transition text-sm"
                        disabled={submitting}
                      >
                        いいえ
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            <button
              type="submit"
              disabled={submitting || loadingEdit || !title.trim()}
              className="w-full px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {submitting ? (isEditing ? '更新中...' : '作成中...') : isEditing ? '日程調整を更新' : '日程調整を作成'}
            </button>
          </form>
        </div>
      )}

      {createdSchedule && publicUrl && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-6 mb-6">
          <div className="flex items-start">
            <div className="text-green-500 text-2xl mr-3">✓</div>

            <div className="flex-1">
              <h3 className="text-lg font-semibold text-green-900 mb-2">日程調整を作成しました</h3>
              <p className="text-sm text-green-800 mb-4">以下のURLをメンバーに送信してください</p>

              <div className="bg-white rounded-md p-3 mb-3 border border-green-300">
                <p className="text-xs text-gray-600 mb-1">公開URL:</p>
                <p className="text-sm text-gray-900 font-mono break-all">{publicUrl}</p>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={handleCopy}
                  className="flex-1 px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition text-sm"
                >
                  {copied ? '✓ コピーしました' : 'URLをコピー'}
                </button>

                <a
                  href={publicUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition text-sm text-center"
                >
                  プレビュー
                </a>
              </div>

              <div className="mt-4 pt-4 border-t border-green-200">
                <p className="text-xs text-green-700">
                  <strong>候補日:</strong> {submittedCandidatesCount}件
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* モバイル用カードビュー */}
      <div className="md:hidden space-y-3">
        {schedules.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-8 text-center text-gray-500">
            日程調整がまだありません。新規作成してください。
          </div>
        ) : (
          schedules.map((schedule) => (
            <MobileCard
              key={schedule.schedule_id}
              onClick={() => navigate(`/schedules/${schedule.schedule_id}`)}
            >
              <CardHeader
                title={schedule.title}
                subtitle={schedule.description || undefined}
                badge={getStatusBadge(schedule.status)}
              />
              <div className="space-y-1">
                <CardField label="候補日数" value={`${schedule.candidate_count || 0}件`} />
                <CardField label="回答数" value={`${schedule.response_count || 0}人`} />
                <CardField
                  label="締切"
                  value={
                    schedule.deadline
                      ? new Date(schedule.deadline).toLocaleString('ja-JP', {
                          month: '2-digit',
                          day: '2-digit',
                          hour: '2-digit',
                          minute: '2-digit',
                        })
                      : '-'
                  }
                />
                <CardField
                  label="作成日"
                  value={new Date(schedule.created_at).toLocaleDateString('ja-JP')}
                />
                <div className="pt-2">
                  <button
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      handleEditClick(schedule.schedule_id);
                    }}
                    className="text-xs text-accent hover:text-accent-dark"
                  >
                    編集
                  </button>
                </div>
              </div>
            </MobileCard>
          ))
        )}
      </div>

      {/* デスクトップ用テーブルビュー */}
      <div className="hidden md:block bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  タイトル
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  ステータス
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  候補日数
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  回答数
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  締切
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  作成日
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  操作
                </th>
              </tr>
            </thead>

            <tbody className="bg-white divide-y divide-gray-200">
              {schedules.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-12 text-center text-gray-500">
                    日程調整がまだありません。新規作成してください。
                  </td>
                </tr>
              ) : (
                schedules.map((schedule) => (
                  <tr key={schedule.schedule_id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div>
                        <div className="text-sm font-medium text-gray-900">{schedule.title}</div>
                        {schedule.description && (
                          <div className="text-sm text-gray-500 truncate max-w-md">
                            {schedule.description}
                          </div>
                        )}
                      </div>
                    </td>

                    <td className="px-6 py-4 whitespace-nowrap">{getStatusBadge(schedule.status)}</td>

                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {schedule.candidate_count || 0}件
                    </td>

                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {schedule.response_count || 0}人
                    </td>

                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {schedule.deadline
                        ? new Date(schedule.deadline).toLocaleString('ja-JP', {
                            year: 'numeric',
                            month: '2-digit',
                            day: '2-digit',
                            hour: '2-digit',
                            minute: '2-digit',
                          })
                        : '-'}
                    </td>

                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(schedule.created_at).toLocaleDateString('ja-JP')}
                    </td>

                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end gap-3">
                        <button
                          onClick={() => navigate(`/schedules/${schedule.schedule_id}`)}
                          className="text-accent hover:text-accent-dark transition"
                        >
                          詳細
                        </button>
                        <button
                          onClick={() => handleEditClick(schedule.schedule_id)}
                          className="text-gray-600 hover:text-gray-800 transition"
                        >
                          編集
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      <div className="mt-6 p-4 bg-accent/10 border border-accent/30 rounded-lg">
        <h3 className="text-sm font-semibold text-accent-dark mb-2">💡 使い方</h3>
        <ul className="text-sm text-accent-dark space-y-1 list-disc list-inside">
          <li>日程調整を作成すると公開URLが発行されます</li>
          <li>URLをメンバーに送信して、参加可能な日程を回答してもらいましょう</li>
          <li>メンバーは候補日の中から参加可能な日程を複数選択できます</li>
          <li>締切を設定すると、締切後は回答できなくなります</li>
          <li>詳細画面で回答状況を確認し、イベントの営業日を決定できます</li>
        </ul>
      </div>
    </div>
  );
}
