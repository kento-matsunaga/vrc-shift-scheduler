import { useState, useMemo } from 'react';
import type { PublicEvent, PublicBusinessDay } from '../lib/api/publicApi';

interface CalendarGridProps {
  events: PublicEvent[];
}

interface DayEvent {
  eventTitle: string;
  eventDescription: string;
  startTime: string;
  endTime: string;
}

type EventsByDate = Record<string, DayEvent[]>;

const WEEKDAYS = ['日', '月', '火', '水', '木', '金', '土'];

export default function CalendarGrid({ events }: CalendarGridProps) {
  const [currentDate, setCurrentDate] = useState(() => new Date());
  const [selectedDate, setSelectedDate] = useState<string | null>(null);

  // イベントを日付ごとにグループ化
  const eventsByDate = useMemo(() => {
    const grouped: EventsByDate = {};

    events.forEach((event) => {
      event.business_days.forEach((bd: PublicBusinessDay) => {
        const dateKey = bd.date;
        if (!grouped[dateKey]) {
          grouped[dateKey] = [];
        }
        grouped[dateKey].push({
          eventTitle: event.title,
          eventDescription: event.description,
          startTime: bd.start_time,
          endTime: bd.end_time,
        });
      });
    });

    // 各日付のイベントを時間順にソート
    Object.keys(grouped).forEach((dateKey) => {
      grouped[dateKey].sort((a, b) => a.startTime.localeCompare(b.startTime));
    });

    return grouped;
  }, [events]);

  // カレンダーグリッドのデータを生成
  const calendarData = useMemo(() => {
    const year = currentDate.getFullYear();
    const month = currentDate.getMonth();

    // 月の初日と最終日
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);

    // カレンダーの開始日（前月の日曜日から）
    const startDate = new Date(firstDay);
    startDate.setDate(startDate.getDate() - firstDay.getDay());

    // カレンダーの終了日（翌月の土曜日まで）
    const endDate = new Date(lastDay);
    const remainingDays = 6 - lastDay.getDay();
    endDate.setDate(endDate.getDate() + remainingDays);

    const weeks: Date[][] = [];
    let currentWeek: Date[] = [];
    const current = new Date(startDate);

    while (current <= endDate) {
      currentWeek.push(new Date(current));
      if (currentWeek.length === 7) {
        weeks.push(currentWeek);
        currentWeek = [];
      }
      current.setDate(current.getDate() + 1);
    }

    return { year, month, weeks, firstDay, lastDay };
  }, [currentDate]);

  const handlePrevMonth = () => {
    setCurrentDate((prev) => new Date(prev.getFullYear(), prev.getMonth() - 1, 1));
    setSelectedDate(null);
  };

  const handleNextMonth = () => {
    setCurrentDate((prev) => new Date(prev.getFullYear(), prev.getMonth() + 1, 1));
    setSelectedDate(null);
  };

  const handleToday = () => {
    setCurrentDate(new Date());
    setSelectedDate(null);
  };

  const formatDateKey = (date: Date): string => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  const isCurrentMonth = (date: Date): boolean => {
    return date.getMonth() === calendarData.month;
  };

  const isToday = (date: Date): boolean => {
    const today = new Date();
    return (
      date.getFullYear() === today.getFullYear() &&
      date.getMonth() === today.getMonth() &&
      date.getDate() === today.getDate()
    );
  };

  const handleDateClick = (date: Date) => {
    const dateKey = formatDateKey(date);
    if (eventsByDate[dateKey]?.length > 0) {
      setSelectedDate(selectedDate === dateKey ? null : dateKey);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden">
      {/* ヘッダー */}
      <div className="flex items-center justify-between p-4 border-b bg-gray-50">
        <button
          onClick={handlePrevMonth}
          className="p-2 hover:bg-gray-200 rounded-full transition"
          aria-label="前月"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>

        <div className="flex items-center gap-4">
          <h2 className="text-xl font-bold text-gray-900">
            {calendarData.year}年 {calendarData.month + 1}月
          </h2>
          <button
            onClick={handleToday}
            className="px-3 py-1 text-sm bg-accent text-white rounded-md hover:bg-accent-dark transition"
          >
            今日
          </button>
        </div>

        <button
          onClick={handleNextMonth}
          className="p-2 hover:bg-gray-200 rounded-full transition"
          aria-label="次月"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      {/* 曜日ヘッダー */}
      <div className="grid grid-cols-7 border-b">
        {WEEKDAYS.map((day, index) => (
          <div
            key={day}
            className={`p-2 text-center text-sm font-medium ${
              index === 0 ? 'text-red-500' : index === 6 ? 'text-blue-500' : 'text-gray-700'
            }`}
          >
            {day}
          </div>
        ))}
      </div>

      {/* カレンダーグリッド */}
      <div className="grid grid-cols-7">
        {calendarData.weeks.map((week, weekIndex) =>
          week.map((date, dayIndex) => {
            const dateKey = formatDateKey(date);
            const dayEvents = eventsByDate[dateKey] || [];
            const hasEvents = dayEvents.length > 0;
            const isSelected = selectedDate === dateKey;

            return (
              <div
                key={`${weekIndex}-${dayIndex}`}
                onClick={() => handleDateClick(date)}
                className={`
                  min-h-[80px] md:min-h-[100px] border-b border-r p-1 transition
                  ${!isCurrentMonth(date) ? 'bg-gray-50 text-gray-400' : 'bg-white'}
                  ${hasEvents ? 'cursor-pointer hover:bg-blue-50' : ''}
                  ${isSelected ? 'bg-blue-100' : ''}
                `}
              >
                <div
                  className={`
                    text-sm font-medium mb-1 w-7 h-7 flex items-center justify-center rounded-full
                    ${isToday(date) ? 'bg-accent text-white' : ''}
                    ${dayIndex === 0 && !isToday(date) ? 'text-red-500' : ''}
                    ${dayIndex === 6 && !isToday(date) ? 'text-blue-500' : ''}
                  `}
                >
                  {date.getDate()}
                </div>

                {/* イベント表示（最大2件 + more） */}
                <div className="space-y-0.5">
                  {dayEvents.slice(0, 2).map((event, eventIndex) => (
                    <div
                      key={eventIndex}
                      className="text-xs bg-accent/10 text-accent rounded px-1 py-0.5 truncate"
                      title={`${event.eventTitle} ${event.startTime}-${event.endTime}`}
                    >
                      <span className="hidden md:inline">{event.startTime} </span>
                      {event.eventTitle}
                    </div>
                  ))}
                  {dayEvents.length > 2 && (
                    <div className="text-xs text-gray-500 pl-1">
                      +{dayEvents.length - 2}件
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* 選択された日付の詳細 */}
      {selectedDate && eventsByDate[selectedDate] && (
        <div className="p-4 border-t bg-blue-50">
          <h3 className="font-bold text-gray-900 mb-3">
            {new Date(selectedDate + 'T00:00:00').toLocaleDateString('ja-JP', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
              weekday: 'short',
            })}
            のイベント
          </h3>
          <div className="space-y-3">
            {eventsByDate[selectedDate].map((event, index) => (
              <div key={index} className="bg-white rounded-md p-3 shadow-sm">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-accent font-medium">
                    {event.startTime} - {event.endTime}
                  </span>
                </div>
                <div className="font-medium text-gray-900">{event.eventTitle}</div>
                {event.eventDescription && (
                  <div className="text-sm text-gray-600 mt-1 whitespace-pre-wrap">
                    {event.eventDescription}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
