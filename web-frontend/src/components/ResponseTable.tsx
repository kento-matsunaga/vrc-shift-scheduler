import type { TargetDate, PublicAttendanceResponse } from '../lib/api/publicApi';
import BaseResponseTable, { type DateItem, type ResponseItem, type ResponseTypeConfig } from './BaseResponseTable';

interface ResponseTableProps {
  targetDates: TargetDate[];
  responses: PublicAttendanceResponse[];
}

// 出欠回答用の回答タイプ設定
const attendanceResponseTypes: ResponseTypeConfig[] = [
  { value: 'attending', label: '参加', icon: '○', iconColor: 'text-green-600', bgColor: 'bg-green-50' },
  { value: 'undecided', label: '未定', icon: '△', iconColor: 'text-yellow-600', bgColor: 'bg-yellow-50' },
  { value: 'absent', label: '不参加', icon: '×', iconColor: 'text-red-600', bgColor: 'bg-red-50' },
];

export default function ResponseTable({ targetDates, responses }: ResponseTableProps) {
  // TargetDate を DateItem に変換
  const dateItems: DateItem[] = targetDates.map((td) => ({
    id: td.target_date_id,
    date: td.target_date,
    startTime: td.start_time,
    endTime: td.end_time,
  }));

  // PublicAttendanceResponse を ResponseItem に変換
  const responseItems: ResponseItem[] = responses.map((r) => ({
    memberId: r.member_id,
    memberName: r.member_name,
    dateItemId: r.target_date_id,
    responseValue: r.response,
    note: r.note,
    availableFrom: r.available_from,
    availableTo: r.available_to,
  }));

  // 参加可能時間の表示
  const renderExtraContent = (response: { availableFrom?: string; availableTo?: string } | undefined) => {
    if (response?.availableFrom && response?.availableTo) {
      return (
        <div className="text-xs text-gray-500 mt-1">
          {response.availableFrom}〜{response.availableTo}
        </div>
      );
    }
    return null;
  };

  return (
    <BaseResponseTable
      dateItems={dateItems}
      responses={responseItems}
      responseTypes={attendanceResponseTypes}
      ariaLabel="出欠回答一覧"
      captionText="メンバーごとの出欠回答を日付別に表示しています。○は参加、△は未定、×は不参加を表します。"
      renderExtraContent={renderExtraContent}
    />
  );
}
