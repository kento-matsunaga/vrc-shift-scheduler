import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  getCalendarByToken,
  type PublicCalendar as PublicCalendarType,
  PublicApiError,
} from '../../lib/api/publicApi';
import CalendarGrid from '../../components/CalendarGrid';
import { useDocumentTitle } from '../../hooks/useDocumentTitle';
import { SEO } from '../../components/seo';

export default function PublicCalendar() {
  const { token } = useParams<{ token: string }>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [calendar, setCalendar] = useState<PublicCalendarType | null>(null);

  useDocumentTitle(calendar?.title || 'カレンダー');

  useEffect(() => {
    if (!token) {
      setError('URLが無効です');
      setLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const calendarData = await getCalendarByToken(token);
        setCalendar(calendarData);
      } catch (err) {
        if (err instanceof PublicApiError) {
          if (err.isNotFound()) {
            setError('カレンダーが見つかりません。URLをご確認ください。');
          } else if (err.isForbidden()) {
            setError('このカレンダーは公開されていません。');
          } else {
            setError('データの取得に失敗しました。');
          }
        } else {
          setError('通信エラーが発生しました。');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [token]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-accent"></div>
          <p className="mt-4 text-gray-600">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-6">
          <div className="text-center">
            <div className="text-red-500 text-5xl mb-4">&#9888;&#65039;</div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">エラー</h2>
            <p className="text-gray-600">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (!calendar) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4">
      <SEO noindex={true} />
      <div className="max-w-4xl mx-auto">
        {/* ヘッダー情報 */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {calendar.title}
          </h1>
          {calendar.description && (
            <p className="text-gray-600 whitespace-pre-wrap">
              {calendar.description}
            </p>
          )}
        </div>

        {/* カレンダーグリッド */}
        {calendar.events.length > 0 ? (
          <CalendarGrid events={calendar.events} />
        ) : (
          <div className="bg-white rounded-lg shadow-md p-6 text-center">
            <p className="text-gray-500">
              表示するイベントがありません
            </p>
          </div>
        )}

        {/* イベント一覧（サブセクション） */}
        {calendar.events.length > 0 && (
          <div className="mt-6 bg-white rounded-lg shadow-md p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              イベント一覧
            </h2>
            <div className="space-y-4">
              {calendar.events.map((event, index) => (
                <div
                  key={index}
                  className="border border-gray-200 rounded-md p-4"
                >
                  <h3 className="font-medium text-gray-900 mb-1">
                    {event.title}
                  </h3>
                  {event.description && (
                    <p className="text-sm text-gray-600 mb-2 whitespace-pre-wrap">
                      {event.description}
                    </p>
                  )}
                  {event.business_days.length > 0 && (
                    <div className="mt-2">
                      <p className="text-sm text-gray-500 mb-1">開催日:</p>
                      <div className="flex flex-wrap gap-2">
                        {event.business_days.slice(0, 5).map((bd, bdIndex) => (
                          <span
                            key={bdIndex}
                            className="text-xs bg-gray-100 text-gray-700 rounded px-2 py-1"
                          >
                            {new Date(bd.date + 'T00:00:00').toLocaleDateString('ja-JP', {
                              month: 'short',
                              day: 'numeric',
                              weekday: 'short',
                            })}{' '}
                            {bd.start_time}-{bd.end_time}
                          </span>
                        ))}
                        {event.business_days.length > 5 && (
                          <span className="text-xs text-gray-500">
                            +{event.business_days.length - 5}件
                          </span>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
