import { useMemo } from 'react';
import type { ScheduleCandidate, PublicScheduleResponse } from '../lib/api/publicApi';

interface ScheduleResponseTableProps {
  candidates: ScheduleCandidate[];
  responses: PublicScheduleResponse[];
}

type AvailabilityType = 'available' | 'unavailable' | 'maybe';

// 型ガード関数
function isValidAvailabilityType(value: unknown): value is AvailabilityType {
  return value === 'available' || value === 'unavailable' || value === 'maybe';
}

function isValidPublicScheduleResponse(obj: unknown): obj is PublicScheduleResponse {
  if (typeof obj !== 'object' || obj === null) return false;
  const r = obj as Record<string, unknown>;
  return (
    typeof r.member_id === 'string' &&
    typeof r.member_name === 'string' &&
    typeof r.candidate_id === 'string' &&
    isValidAvailabilityType(r.availability) &&
    typeof r.note === 'string'
  );
}

interface MemberResponseMap {
  memberName: string;
  memberId: string;
  responses: Record<string, { availability: AvailabilityType; note: string }>;
}

export default function ScheduleResponseTable({ candidates, responses }: ScheduleResponseTableProps) {
  // 型安全なレスポンスのみをフィルタリング
  const validResponses = useMemo(() =>
    responses.filter(isValidPublicScheduleResponse),
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
      member.responses[r.candidate_id] = {
        availability: r.availability,
        note: r.note,
      };
    });

    // メンバー名でソート
    return Array.from(memberMap.values()).sort((a, b) =>
      a.memberName.localeCompare(b.memberName, 'ja')
    );
  }, [validResponses]);

  // 候補日ごとの参加数を計算
  const candidateSummary = useMemo(() => {
    const summary: Record<string, { available: number; unavailable: number; maybe: number }> = {};

    candidates.forEach((c) => {
      summary[c.candidate_id] = { available: 0, unavailable: 0, maybe: 0 };
    });

    validResponses.forEach((r) => {
      if (summary[r.candidate_id]) {
        summary[r.candidate_id][r.availability]++;
      }
    });

    return summary;
  }, [candidates, validResponses]);

  const getAvailabilityIcon = (availability: AvailabilityType | undefined) => {
    switch (availability) {
      case 'available':
        return <span className="text-green-600 font-bold text-lg">○</span>;
      case 'unavailable':
        return <span className="text-red-600 font-bold text-lg">×</span>;
      case 'maybe':
        return <span className="text-yellow-600 font-bold text-lg">△</span>;
      default:
        return <span className="text-gray-400">-</span>;
    }
  };

  const getAvailabilityBgColor = (availability: AvailabilityType | undefined) => {
    switch (availability) {
      case 'available':
        return 'bg-green-50';
      case 'unavailable':
        return 'bg-red-50';
      case 'maybe':
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

  const formatTime = (timeStr: string | undefined) => {
    if (!timeStr) return '';
    const date = new Date(timeStr);
    return date.toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
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
            {candidates.map((c) => (
              <th
                key={c.candidate_id}
                className="border border-gray-300 px-3 py-2 text-center font-medium text-gray-700 min-w-[80px]"
              >
                <div>{formatDate(c.date)}</div>
                {(c.start_time || c.end_time) && (
                  <div className="text-xs text-gray-500 font-normal">
                    {formatTime(c.start_time) || ''}〜{formatTime(c.end_time) || ''}
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
              {candidates.map((c) => {
                const resp = member.responses[c.candidate_id];
                return (
                  <td
                    key={c.candidate_id}
                    className={`border border-gray-300 px-3 py-2 text-center ${getAvailabilityBgColor(resp?.availability)}`}
                    title={resp?.note || undefined}
                  >
                    {getAvailabilityIcon(resp?.availability)}
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
            {candidates.map((c) => {
              const summary = candidateSummary[c.candidate_id];
              return (
                <td
                  key={c.candidate_id}
                  className="border border-gray-300 px-3 py-2 text-center"
                >
                  <div className="flex justify-center gap-2 text-xs">
                    <span className="text-green-600">○{summary?.available || 0}</span>
                    <span className="text-yellow-600">△{summary?.maybe || 0}</span>
                    <span className="text-red-600">×{summary?.unavailable || 0}</span>
                  </div>
                </td>
              );
            })}
          </tr>
        </tbody>
      </table>
      <div className="mt-3 flex gap-4 text-sm text-gray-600">
        <span><span className="text-green-600 font-bold">○</span> 参加可能</span>
        <span><span className="text-yellow-600 font-bold">△</span> 微妙</span>
        <span><span className="text-red-600 font-bold">×</span> 参加不可</span>
      </div>
    </div>
  );
}
