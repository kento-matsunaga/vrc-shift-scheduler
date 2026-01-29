import { useMemo } from 'react';

// 日付/候補の共通インターフェース
export interface DateItem {
  id: string;
  date: string;
  startTime?: string;
  endTime?: string;
}

// 回答の共通インターフェース
export interface ResponseItem {
  memberId: string;
  memberName: string;
  dateItemId: string;
  responseValue: string;
  note: string;
  availableFrom?: string;
  availableTo?: string;
}

// 回答タイプの設定
export interface ResponseTypeConfig {
  value: string;
  label: string;
  icon: string;
  iconColor: string;
  bgColor: string;
}

interface BaseResponseTableProps {
  dateItems: DateItem[];
  responses: ResponseItem[];
  responseTypes: ResponseTypeConfig[];
  ariaLabel: string;
  captionText: string;
  /** 時間表示をフォーマットする関数（Schedule用にISO→HH:MM変換が必要な場合） */
  formatTimeDisplay?: (time: string | undefined) => string;
  /** 回答セル内に追加表示するコンテンツ（availableFrom/To用） */
  renderExtraContent?: (response: { availableFrom?: string; availableTo?: string } | undefined) => React.ReactNode;
}

interface MemberResponseMap {
  memberName: string;
  memberId: string;
  responses: Record<string, { responseValue: string; note: string; availableFrom?: string; availableTo?: string }>;
}

export default function BaseResponseTable({
  dateItems,
  responses,
  responseTypes,
  ariaLabel,
  captionText,
  formatTimeDisplay,
  renderExtraContent,
}: BaseResponseTableProps) {
  // メンバーごとに回答をグループ化
  const memberResponses = useMemo(() => {
    const memberMap = new Map<string, MemberResponseMap>();

    responses.forEach((r) => {
      if (!memberMap.has(r.memberId)) {
        memberMap.set(r.memberId, {
          memberName: r.memberName,
          memberId: r.memberId,
          responses: {},
        });
      }

      const member = memberMap.get(r.memberId)!;
      member.responses[r.dateItemId] = {
        responseValue: r.responseValue,
        note: r.note,
        availableFrom: r.availableFrom,
        availableTo: r.availableTo,
      };
    });

    // メンバー名でソート
    return Array.from(memberMap.values()).sort((a, b) =>
      a.memberName.localeCompare(b.memberName, 'ja')
    );
  }, [responses]);

  // 日付ごとの回答数を計算
  const dateSummary = useMemo(() => {
    const summary: Record<string, Record<string, number>> = {};

    dateItems.forEach((item) => {
      summary[item.id] = {};
      responseTypes.forEach((rt) => {
        summary[item.id][rt.value] = 0;
      });
    });

    responses.forEach((r) => {
      if (summary[r.dateItemId] && summary[r.dateItemId][r.responseValue] !== undefined) {
        summary[r.dateItemId][r.responseValue]++;
      }
    });

    return summary;
  }, [dateItems, responses, responseTypes]);

  const getResponseConfig = (responseValue: string | undefined): ResponseTypeConfig | undefined => {
    return responseTypes.find((rt) => rt.value === responseValue);
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('ja-JP', {
      month: 'numeric',
      day: 'numeric',
      weekday: 'short',
    });
  };

  const defaultFormatTime = (timeStr: string | undefined) => timeStr || '';

  const formatTime = formatTimeDisplay || defaultFormatTime;

  if (memberResponses.length === 0) {
    return (
      <div className="bg-gray-50 rounded-lg p-6 text-center">
        <p className="text-gray-500">まだ回答がありません</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse text-sm" aria-label={ariaLabel}>
        <caption className="sr-only">{captionText}</caption>
        <thead>
          <tr className="bg-gray-100">
            <th className="border border-gray-300 px-3 py-2 text-left font-medium text-gray-700 sticky left-0 bg-gray-100 z-10">
              名前
            </th>
            {dateItems.map((item) => (
              <th
                key={item.id}
                className="border border-gray-300 px-3 py-2 text-center font-medium text-gray-700 min-w-[80px]"
              >
                <div>{formatDate(item.date)}</div>
                {(item.startTime || item.endTime) && (
                  <div className="text-xs text-gray-500 font-normal">
                    {formatTime(item.startTime)}〜{formatTime(item.endTime)}
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
              {dateItems.map((item) => {
                const resp = member.responses[item.id];
                const config = getResponseConfig(resp?.responseValue);
                return (
                  <td
                    key={item.id}
                    className={`border border-gray-300 px-3 py-2 text-center ${config?.bgColor || 'bg-gray-50'}`}
                    title={resp?.note || undefined}
                  >
                    <div className="flex flex-col items-center">
                      {config ? (
                        <span className={`${config.iconColor} font-bold text-lg`}>{config.icon}</span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                      {renderExtraContent && renderExtraContent(resp)}
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
            {dateItems.map((item) => {
              const summary = dateSummary[item.id];
              return (
                <td
                  key={item.id}
                  className="border border-gray-300 px-3 py-2 text-center"
                >
                  <div className="flex justify-center gap-2 text-xs">
                    {responseTypes.map((rt) => (
                      <span key={rt.value} className={rt.iconColor}>
                        {rt.icon}{summary?.[rt.value] || 0}
                      </span>
                    ))}
                  </div>
                </td>
              );
            })}
          </tr>
        </tbody>
      </table>
      <div className="mt-3 flex gap-4 text-sm text-gray-600">
        {responseTypes.map((rt) => (
          <span key={rt.value}>
            <span className={`${rt.iconColor} font-bold`}>{rt.icon}</span> {rt.label}
          </span>
        ))}
      </div>
    </div>
  );
}
