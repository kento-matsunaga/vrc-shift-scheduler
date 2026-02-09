import type { ScheduleCandidate, PublicScheduleResponse } from '../lib/api/publicApi';
import BaseResponseTable, { type DateItem, type ResponseItem, type ResponseTypeConfig } from './BaseResponseTable';

interface ScheduleResponseTableProps {
  candidates: ScheduleCandidate[];
  responses: PublicScheduleResponse[];
}

// 日程調整用の回答タイプ設定
const scheduleResponseTypes: ResponseTypeConfig[] = [
  { value: 'available', label: '参加可能', icon: '○', iconColor: 'text-green-600', bgColor: 'bg-green-50' },
  { value: 'maybe', label: '微妙', icon: '△', iconColor: 'text-yellow-600', bgColor: 'bg-yellow-50' },
  { value: 'unavailable', label: '参加不可', icon: '×', iconColor: 'text-red-600', bgColor: 'bg-red-50' },
];

// ISO 形式の時間を HH:MM に変換
const formatTimeDisplay = (timeStr: string | undefined): string => {
  if (!timeStr) return '';
  const date = new Date(timeStr);
  return date.toLocaleTimeString('ja-JP', {
    hour: '2-digit',
    minute: '2-digit',
  });
};

export default function ScheduleResponseTable({ candidates, responses }: ScheduleResponseTableProps) {
  // ScheduleCandidate を DateItem に変換
  const dateItems: DateItem[] = candidates.map((c) => ({
    id: c.candidate_id,
    date: c.date,
    startTime: c.start_time,
    endTime: c.end_time,
  }));

  // PublicScheduleResponse を ResponseItem に変換
  const responseItems: ResponseItem[] = responses.map((r) => ({
    memberId: r.member_id,
    memberName: r.member_name,
    dateItemId: r.candidate_id,
    responseValue: r.availability,
    note: r.note,
  }));

  return (
    <BaseResponseTable
      dateItems={dateItems}
      responses={responseItems}
      responseTypes={scheduleResponseTypes}
      ariaLabel="日程調整回答一覧"
      captionText="メンバーごとの日程調整回答を候補日別に表示しています。○は参加可能、△は微妙、×は参加不可を表します。"
      formatTimeDisplay={formatTimeDisplay}
    />
  );
}
