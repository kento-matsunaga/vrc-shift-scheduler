import { useMemo } from 'react';
import type { TargetDate, PublicAttendanceResponse } from '../lib/api/publicApi';

interface ResponseTableProps {
  targetDates: TargetDate[];
  responses: PublicAttendanceResponse[];
}

type ResponseType = 'attending' | 'absent' | 'undecided';

// 型ガード関数
function isValidResponseType(value: unknown): value is ResponseType {
  return value === 'attending' || value === 'absent' || value === 'undecided';
}

function isValidPublicAttendanceResponse(obj: unknown): obj is PublicAttendanceResponse {
  if (typeof obj !== 'object' || obj === null) return false;
  const r = obj as Record<string, unknown>;
  return (
    typeof r.member_id === 'string' &&
    typeof r.member_name === 'string' &&
    typeof r.target_date_id === 'string' &&
    isValidResponseType(r.response) &&
    typeof r.note === 'string'
  );
}

interface MemberResponseMap {
  memberName: string;
  memberId: string;
  responses: Record<string, { response: ResponseType; note: string; availableFrom?: string; availableTo?: string }>;
}

export default function ResponseTable({ targetDates, responses }: ResponseTableProps) {
  // 型安全なレスポンスのみをフィルタリング
  const validResponses = useMemo(() =>
    responses.filter(isValidPublicAttendanceResponse),
    [responses]
  );

  // メンバーごとに回答をグループ化
  const memberResponses = useMemo(() => {
    const memberMap = new Map<string, MemberResponseMap>();

    validResponses.forEach((r) => {
      if (!memberMap.has(r.member_id)) {
        memberMap.set(r.member_id, {
          memberName: r.member_name,
          memberId: r.member_id,
          responses: {},
        });
      }

      const member = memberMap.get(r.member_id)!;
      member.responses[r.target_date_id] = {
        response: r.response,
        note: r.note,
        availableFrom: r.available_from,
        availableTo: r.available_to,
      };
    });

    // メンバー名でソート
    return Array.from(memberMap.values()).sort((a, b) =>
      a.memberName.localeCompare(b.memberName, 'ja')
    );
  }, [validResponses]);

  // 日付ごとの参加数を計算
  const dateSummary = useMemo(() => {
    const summary: Record<string, { attending: number; absent: number; undecided: number }> = {};

    targetDates.forEach((td) => {
      summary[td.target_date_id] = { attending: 0, absent: 0, undecided: 0 };
    });

    validResponses.forEach((r) => {
      if (summary[r.target_date_id]) {
        summary[r.target_date_id][r.response]++;
      }
    });

    return summary;
  }, [targetDates, validResponses]);

  const getResponseIcon = (response: ResponseType | undefined) => {
    switch (response) {
      case 'attending':
        return <span className="text-green-600 font-bold text-lg">○</span>;
      case 'absent':
        return <span className="text-red-600 font-bold text-lg">×</span>;
      case 'undecided':
        return <span className="text-yellow-600 font-bold text-lg">△</span>;
      default:
        return <span className="text-gray-400">-</span>;
    }
  };

  const getResponseBgColor = (response: ResponseType | undefined) => {
    switch (response) {
      case 'attending':
        return 'bg-green-50';
      case 'absent':
        return 'bg-red-50';
      case 'undecided':
        return 'bg-yellow-50';
      default:
        return 'bg-gray-50';
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('ja-JP', {
      month: 'numeric',
      day: 'numeric',
      weekday: 'short',
    });
  };

  if (memberResponses.length === 0) {
    return (
      <div className="bg-gray-50 rounded-lg p-6 text-center">
        <p className="text-gray-500">まだ回答がありません</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="bg-gray-100">
            <th className="border border-gray-300 px-3 py-2 text-left font-medium text-gray-700 sticky left-0 bg-gray-100 z-10">
              名前
            </th>
            {targetDates.map((td) => (
              <th
                key={td.target_date_id}
                className="border border-gray-300 px-3 py-2 text-center font-medium text-gray-700 min-w-[80px]"
              >
                <div>{formatDate(td.target_date)}</div>
                {(td.start_time || td.end_time) && (
                  <div className="text-xs text-gray-500 font-normal">
                    {td.start_time || ''}〜{td.end_time || ''}
                  </div>
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {memberResponses.map((member) => (
            <tr key={member.memberId} className="hover:bg-gray-50">
              <td className="border border-gray-300 px-3 py-2 font-medium text-gray-900 sticky left-0 bg-white z-10">
                {member.memberName}
              </td>
              {targetDates.map((td) => {
                const resp = member.responses[td.target_date_id];
                return (
                  <td
                    key={td.target_date_id}
                    className={`border border-gray-300 px-3 py-2 text-center ${getResponseBgColor(resp?.response)}`}
                    title={resp?.note || undefined}
                  >
                    <div className="flex flex-col items-center">
                      {getResponseIcon(resp?.response)}
                      {resp?.availableFrom && resp?.availableTo && (
                        <div className="text-xs text-gray-500 mt-1">
                          {resp.availableFrom}〜{resp.availableTo}
                        </div>
                      )}
                    </div>
                  </td>
                );
              })}
            </tr>
          ))}
          {/* 集計行 */}
          <tr className="bg-gray-100 font-medium">
            <td className="border border-gray-300 px-3 py-2 text-gray-700 sticky left-0 bg-gray-100 z-10">
              集計
            </td>
            {targetDates.map((td) => {
              const summary = dateSummary[td.target_date_id];
              return (
                <td
                  key={td.target_date_id}
                  className="border border-gray-300 px-3 py-2 text-center"
                >
                  <div className="flex justify-center gap-2 text-xs">
                    <span className="text-green-600">○{summary?.attending || 0}</span>
                    <span className="text-yellow-600">△{summary?.undecided || 0}</span>
                    <span className="text-red-600">×{summary?.absent || 0}</span>
                  </div>
                </td>
              );
            })}
          </tr>
        </tbody>
      </table>
      <div className="mt-3 flex gap-4 text-sm text-gray-600">
        <span><span className="text-green-600 font-bold">○</span> 参加</span>
        <span><span className="text-yellow-600 font-bold">△</span> 未定</span>
        <span><span className="text-red-600 font-bold">×</span> 不参加</span>
      </div>
    </div>
  );
}
