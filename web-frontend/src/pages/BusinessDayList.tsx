import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { SEO } from '../components/seo';
import { getEventDetail, getBusinessDays, createBusinessDay, getMembers } from '../lib/api';
import { listSchedules, getSchedule, getScheduleResponses, type Schedule, type ScheduleResponse } from '../lib/api/scheduleApi';
import { listTemplates } from '../lib/api/templateApi';
import type { Event, BusinessDay, Member, Template } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function BusinessDayList() {
  const { eventId } = useParams<{ eventId: string }>();
  const [event, setEvent] = useState<Event | null>(null);
  const [businessDays, setBusinessDays] = useState<BusinessDay[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

  // 現在表示中の月を管理（YYYY-MM形式）
  const now = new Date();
  const currentMonthKey = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
  const [selectedMonth, setSelectedMonth] = useState<string>(currentMonthKey);

  useEffect(() => {
    if (eventId) {
      loadData();
    }
  }, [eventId]);

  const loadData = async () => {
    if (!eventId) return;

    try {
      setLoading(true);
      const [eventData, businessDaysData] = await Promise.all([
        getEventDetail(eventId),
        getBusinessDays(eventId),
      ]);
      setEvent(eventData);
      setBusinessDays(businessDaysData.business_days || []);
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

  // 営業日を月ごとにグループ化
  const groupByMonth = (days: BusinessDay[]) => {
    const groups: Record<string, BusinessDay[]> = {};

    days.forEach((day) => {
      const date = new Date(day.target_date);
      const monthKey = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;

      if (!groups[monthKey]) {
        groups[monthKey] = [];
      }
      groups[monthKey].push(day);
    });

    // 各月内で日付順にソート
    Object.keys(groups).forEach((key) => {
      groups[key].sort((a, b) =>
        new Date(a.target_date).getTime() - new Date(b.target_date).getTime()
      );
    });

    return groups;
  };

  // 月キーをソート（時系列順）
  const getSortedMonthKeys = (groups: Record<string, BusinessDay[]>) => {
    return Object.keys(groups).sort((a, b) => a.localeCompare(b));
  };

  // 前の月へ移動
  const goToPreviousMonth = () => {
    const monthGroups = groupByMonth(businessDays);
    const sortedKeys = getSortedMonthKeys(monthGroups);
    const currentIndex = sortedKeys.indexOf(selectedMonth);
    if (currentIndex > 0) {
      setSelectedMonth(sortedKeys[currentIndex - 1]);
    }
  };

  // 次の月へ移動
  const goToNextMonth = () => {
    const monthGroups = groupByMonth(businessDays);
    const sortedKeys = getSortedMonthKeys(monthGroups);
    const currentIndex = sortedKeys.indexOf(selectedMonth);
    if (currentIndex < sortedKeys.length - 1) {
      setSelectedMonth(sortedKeys[currentIndex + 1]);
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

  if (!event) {
    return (
      <div className="card text-center py-12">
        <p className="text-gray-600">イベントが見つかりません</p>
      </div>
    );
  }

  return (
    <div>
      <SEO noindex={true} />
      {/* パンくずリスト */}
      <nav className="mb-6 text-sm text-gray-600">
        <Link to="/events" className="hover:text-gray-900">
          イベント一覧
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">{event.event_name}</span>
      </nav>

      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">{event.event_name}</h2>
          <p className="text-sm text-gray-600 mt-1">{event.description}</p>
        </div>
        <div className="flex gap-2">
          <Link
            to={`/events/${eventId}/templates`}
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
            テンプレート管理
          </Link>
          <Link
            to={`/events/${eventId}/instances`}
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
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
              />
            </svg>
            インスタンス
          </Link>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            ＋ 営業日を追加
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {businessDays.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだ営業日がありません</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初の営業日を追加
          </button>
        </div>
      ) : (() => {
        const monthGroups = groupByMonth(businessDays);
        const sortedKeys = getSortedMonthKeys(monthGroups);

        // 選択された月が存在しない場合は最初の月を選択
        if (!monthGroups[selectedMonth] && sortedKeys.length > 0) {
          setSelectedMonth(sortedKeys[0]);
          return null;
        }

        const monthDays = monthGroups[selectedMonth] || [];
        const [_year, _month] = selectedMonth.split('-');
        const currentIndex = sortedKeys.indexOf(selectedMonth);
        const hasPrevious = currentIndex > 0;
        const hasNext = currentIndex < sortedKeys.length - 1;

        return (
          <div>
            {/* 月選択コントロール */}
            <div className="card mb-6">
              <div className="flex items-center justify-between gap-4">
                {/* 前月ボタン */}
                <button
                  onClick={goToPreviousMonth}
                  disabled={!hasPrevious}
                  className={`p-2 rounded-lg transition-colors ${
                    hasPrevious
                      ? 'text-gray-700 hover:bg-gray-100'
                      : 'text-gray-300 cursor-not-allowed'
                  }`}
                  title="前の月"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                </button>

                {/* 月選択プルダウン */}
                <select
                  value={selectedMonth}
                  onChange={(e) => setSelectedMonth(e.target.value)}
                  className="flex-1 px-4 py-2 text-center text-lg font-bold text-gray-900 bg-white border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-accent"
                >
                  {sortedKeys.map((monthKey) => {
                    const [y, m] = monthKey.split('-');
                    return (
                      <option key={monthKey} value={monthKey}>
                        {y}年{parseInt(m)}月
                      </option>
                    );
                  })}
                </select>

                {/* 次月ボタン */}
                <button
                  onClick={goToNextMonth}
                  disabled={!hasNext}
                  className={`p-2 rounded-lg transition-colors ${
                    hasNext
                      ? 'text-gray-700 hover:bg-gray-100'
                      : 'text-gray-300 cursor-not-allowed'
                  }`}
                  title="次の月"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              </div>

              {/* 営業日数表示 */}
              <div className="text-sm text-gray-600 text-center mt-3">
                {monthDays.length}件の営業日
              </div>
            </div>

            {/* 営業日カード */}
            {monthDays.length === 0 ? (
              <div className="card text-center py-12">
                <p className="text-gray-600">この月には営業日がありません</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {monthDays.map((day) => (
                  <Link
                    key={day.business_day_id}
                    to={`/business-days/${day.business_day_id}/shift-slots`}
                    className="card hover:shadow-lg transition-shadow"
                  >
                    <div className="flex justify-between items-start mb-2">
                      <div>
                        <div className="text-lg font-bold text-gray-900">
                          {new Date(day.target_date).toLocaleDateString('ja-JP', {
                            month: 'long',
                            day: 'numeric',
                            weekday: 'short',
                          })}
                        </div>
                        <div className="text-sm text-gray-600">
                          {day.start_time.slice(0, 5)} 〜 {day.end_time.slice(0, 5)}
                        </div>
                      </div>
                      <span
                        className={`inline-block px-2 py-1 text-xs font-semibold rounded ${
                          day.occurrence_type === 'recurring'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-orange-100 text-orange-800'
                        }`}
                      >
                        {day.occurrence_type === 'recurring' ? '通常営業' : '特別営業'}
                      </span>
                    </div>
                    {!day.is_active && (
                      <div className="mt-2 text-xs text-red-600">（非アクティブ）</div>
                    )}
                  </Link>
                ))}
              </div>
            )}
          </div>
        );
      })()}

      {/* 営業日作成モーダル */}
      {showCreateModal && eventId && (
        <CreateBusinessDayModal
          eventId={eventId}
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}
    </div>
  );
}

// 営業日作成モーダルコンポーネント
function CreateBusinessDayModal({
  eventId,
  onClose,
  onSuccess,
}: {
  eventId: string;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [targetDate, setTargetDate] = useState('');
  const [startTime, setStartTime] = useState('21:30');
  const [endTime, setEndTime] = useState('23:00');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [selectedScheduleId, setSelectedScheduleId] = useState<string>('');
  const [selectedSchedule, setSelectedSchedule] = useState<Schedule | null>(null);
  const [scheduleResponses, setScheduleResponses] = useState<ScheduleResponse[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [loadingSchedule, setLoadingSchedule] = useState(false);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>('');

  // 日程調整一覧を取得
  useEffect(() => {
    loadSchedules();
  }, []);

  // 日程調整を手動で選択したときの処理
  useEffect(() => {
    if (selectedScheduleId) {
      loadScheduleDetail(selectedScheduleId);
    } else {
      setSelectedSchedule(null);
      setScheduleResponses([]);
    }
  }, [selectedScheduleId]);

  const loadSchedules = async () => {
    try {
      const [schedulesData, membersData, templatesData] = await Promise.all([
        listSchedules(),
        getMembers({ is_active: true }),
        listTemplates(eventId),
      ]);
      setSchedules(schedulesData || []);
      setMembers(membersData.members || []);
      setTemplates(templatesData || []);
    } catch (err) {
      console.error('Failed to load schedules:', err);
    }
  };

  const loadScheduleDetail = async (scheduleId: string) => {
    try {
      setLoadingSchedule(true);
      const [scheduleData, responsesData] = await Promise.all([
        getSchedule(scheduleId),
        getScheduleResponses(scheduleId),
      ]);
      setSelectedSchedule(scheduleData);
      setScheduleResponses(responsesData || []);
    } catch (err) {
      console.error('Failed to load schedule detail:', err);
    } finally {
      setLoadingSchedule(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!targetDate) {
      setError('日付を選択してください');
      return;
    }

    if (!startTime || !endTime) {
      setError('時刻を入力してください');
      return;
    }

    setLoading(true);

    try {
      await createBusinessDay(eventId, {
        target_date: targetDate,
        start_time: startTime,
        end_time: endTime,
        occurrence_type: 'special', // 手動作成は常に特別営業
        template_id: selectedTemplateId || undefined,
      });
      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('営業日の作成に失敗しました');
      }
      console.error('Failed to create business day:', err);
    } finally {
      setLoading(false);
    }
  };

  // 回答済みメンバーのユニークIDを取得
  const respondedMemberIds = selectedSchedule ? new Set(scheduleResponses.map((r) => r.member_id)) : new Set();

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto p-6">
        <h3 className="text-xl font-bold text-gray-900 mb-4">営業日を追加</h3>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="targetDate" className="label">
              日付 <span className="text-red-500">*</span>
            </label>
            <input
              type="date"
              id="targetDate"
              value={targetDate}
              onChange={(e) => setTargetDate(e.target.value)}
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

          <div className="bg-accent/10 border border-accent/30 rounded-lg p-3 mb-4">
            <p className="text-xs text-accent-dark">
              💡 深夜営業の場合、終了時刻が開始時刻より前でもOKです（例: 21:30-02:00）
            </p>
            <p className="text-xs text-accent-dark mt-1">
              📋 手動で追加した営業日は「特別営業」として登録されます
            </p>
          </div>

          {/* テンプレート選択 */}
          {templates.length > 0 && (
            <div className="mb-4">
              <label htmlFor="templateSelect" className="label">
                シフトテンプレート（任意）
              </label>
              <select
                id="templateSelect"
                value={selectedTemplateId}
                onChange={(e) => setSelectedTemplateId(e.target.value)}
                className="input-field"
                disabled={loading}
              >
                <option value="">テンプレートを選択しない</option>
                {templates.map((template) => (
                  <option key={template.template_id} value={template.template_id}>
                    {template.template_name} ({(template.items || []).length}個のシフト枠)
                  </option>
                ))}
              </select>
              <p className="text-xs text-gray-500 mt-1">
                テンプレートを選択すると、営業日作成時に自動的にシフト枠が作成されます
              </p>
            </div>
          )}

          {/* 日程調整選択 */}
          {schedules.length > 0 && (
            <div className="mb-4">
              <label htmlFor="scheduleSelect" className="label">
                日程調整を参照（任意）
              </label>
              <select
                id="scheduleSelect"
                value={selectedScheduleId}
                onChange={(e) => setSelectedScheduleId(e.target.value)}
                className="input-field"
                disabled={loading}
              >
                <option value="">日程調整を選択してください</option>
                {schedules.map((schedule) => (
                  <option key={schedule.schedule_id} value={schedule.schedule_id}>
                    {schedule.title}
                  </option>
                ))}
              </select>
              <p className="text-xs text-gray-500 mt-1">
                日程調整の回答状況を確認しながら営業日を追加できます
              </p>
            </div>
          )}

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* 日程調整結果 */}
          {selectedSchedule && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <h4 className="font-semibold text-gray-900 mb-3">
                📅 日程調整結果: {selectedSchedule.title}
              </h4>
              {loadingSchedule ? (
                <div className="text-center py-4 text-gray-600">読み込み中...</div>
              ) : (
                <div>
                  <p className="text-sm text-gray-600 mb-3">
                    回答数: {respondedMemberIds.size}/{members.length}人
                  </p>
                  <div className="max-h-64 overflow-y-auto border border-gray-200 rounded-lg">
                    <table className="min-w-full text-sm">
                      <thead className="bg-gray-50 sticky top-0">
                        <tr>
                          <th className="px-3 py-2 text-left text-xs font-medium text-gray-500">候補日</th>
                          <th className="px-3 py-2 text-center text-xs font-medium text-gray-500">○</th>
                          <th className="px-3 py-2 text-center text-xs font-medium text-gray-500">△</th>
                          <th className="px-3 py-2 text-center text-xs font-medium text-gray-500">×</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-200">
                        {selectedSchedule.candidates?.map((candidate: any) => {
                          const candidateResponses = scheduleResponses.filter(
                            (r) => r.candidate_id === candidate.candidate_id
                          );
                          const availableCount = candidateResponses.filter((r) => r.availability === 'available').length;
                          const maybeCount = candidateResponses.filter((r) => r.availability === 'maybe').length;
                          const unavailableCount = candidateResponses.filter((r) => r.availability === 'unavailable').length;

                          // 選択した日付と候補日が同じかチェック
                          const candidateDateStr = new Date(candidate.date).toISOString().split('T')[0];
                          const isSelected = targetDate === candidateDateStr;

                          return (
                            <tr
                              key={candidate.candidate_id}
                              className={isSelected ? 'bg-accent/10' : 'hover:bg-gray-50'}
                              onClick={() => setTargetDate(candidateDateStr)}
                              style={{ cursor: 'pointer' }}
                            >
                              <td className="px-3 py-2">
                                <div className="flex items-center gap-2">
                                  {isSelected && <span className="text-accent">→</span>}
                                  <span className={isSelected ? 'font-semibold text-accent-dark' : ''}>
                                    {new Date(candidate.date).toLocaleDateString('ja-JP', {
                                      month: '2-digit',
                                      day: '2-digit',
                                      weekday: 'short',
                                    })}
                                  </span>
                                </div>
                              </td>
                              <td className="px-3 py-2 text-center">
                                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800">
                                  {availableCount}
                                </span>
                              </td>
                              <td className="px-3 py-2 text-center">
                                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                                  {maybeCount}
                                </span>
                              </td>
                              <td className="px-3 py-2 text-center">
                                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-red-100 text-red-800">
                                  {unavailableCount}
                                </span>
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                  <p className="text-xs text-gray-500 mt-2">
                    💡 候補日をクリックすると日付欄に自動入力されます。○: 参加可能、△: 不確定、×: 参加不可
                  </p>

                  {/* 選択した日付のメンバー別回答詳細 */}
                  {targetDate && (() => {
                    // 選択した日付の候補日を見つける
                    const selectedCandidate = selectedSchedule.candidates?.find((c: any) => {
                      const candidateDateStr = new Date(c.date).toISOString().split('T')[0];
                      return targetDate === candidateDateStr;
                    });

                    if (!selectedCandidate) return null;

                    // この候補日への回答を取得
                    const candidateResponses = scheduleResponses.filter(
                      (r) => r.candidate_id === selectedCandidate.candidate_id
                    );

                    // メンバーごとの回答状況を作成
                    const memberResponseMap = new Map<string, string>();
                    candidateResponses.forEach((r) => {
                      memberResponseMap.set(r.member_id, r.availability);
                    });

                    return (
                      <div className="mt-4 pt-4 border-t border-gray-200">
                        <h5 className="font-semibold text-gray-900 mb-3">
                          {new Date(targetDate).toLocaleDateString('ja-JP', {
                            month: 'long',
                            day: 'numeric',
                            weekday: 'short',
                          })} のメンバー別回答
                        </h5>
                        <div className="max-h-48 overflow-y-auto border border-gray-200 rounded-lg">
                          <table className="min-w-full text-sm">
                            <thead className="bg-gray-50 sticky top-0">
                              <tr>
                                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500">メンバー</th>
                                <th className="px-3 py-2 text-center text-xs font-medium text-gray-500">回答</th>
                              </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-200">
                              {members.map((member) => {
                                const availability = memberResponseMap.get(member.member_id);
                                let statusText = '-';
                                let statusColor = 'text-gray-400';

                                if (availability === 'available') {
                                  statusText = '○';
                                  statusColor = 'text-green-600 font-bold';
                                } else if (availability === 'maybe') {
                                  statusText = '△';
                                  statusColor = 'text-yellow-600 font-bold';
                                } else if (availability === 'unavailable') {
                                  statusText = '×';
                                  statusColor = 'text-red-600 font-bold';
                                }

                                return (
                                  <tr key={member.member_id} className="hover:bg-gray-50">
                                    <td className="px-3 py-2 text-gray-900">{member.display_name}</td>
                                    <td className={`px-3 py-2 text-center ${statusColor} text-base`}>
                                      {statusText}
                                    </td>
                                  </tr>
                                );
                              })}
                            </tbody>
                          </table>
                        </div>
                        <p className="text-xs text-gray-500 mt-2">
                          ○: 参加可能、△: 不確定、×: 参加不可、-: 未回答
                        </p>
                      </div>
                    );
                  })()}
                </div>
              )}
            </div>
          )}

          <div className="flex space-x-3 mt-6">
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
              disabled={loading || !targetDate || !startTime || !endTime}
            >
              {loading ? '作成中...' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

