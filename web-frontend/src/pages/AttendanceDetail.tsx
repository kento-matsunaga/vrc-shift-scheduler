import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  getAttendanceCollection,
  getAttendanceResponses,
  closeAttendanceCollection,
  type AttendanceCollection as AttendanceCollectionType,
  type AttendanceResponse,
} from '../lib/api/attendanceApi';

export default function AttendanceDetail() {
  const { collectionId } = useParams<{ collectionId: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [collection, setCollection] = useState<AttendanceCollectionType | null>(null);
  const [responses, setResponses] = useState<AttendanceResponse[]>([]);
  const [closing, setClosing] = useState(false);

  useEffect(() => {
    if (!collectionId) {
      setError('出欠確認IDが指定されていません');
      setLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setLoading(true);
        const [collectionData, responsesData] = await Promise.all([
          getAttendanceCollection(collectionId),
          getAttendanceResponses(collectionId),
        ]);
        setCollection(collectionData);
        setResponses(responsesData);
      } catch (err) {
        setError(err instanceof Error ? err.message : '取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [collectionId]);

  const handleClose = async () => {
    if (!collectionId) return;
    if (!confirm('この出欠確認を締め切りますか？締め切り後は回答を受け付けなくなります。')) return;

    try {
      setClosing(true);
      await closeAttendanceCollection(collectionId);
      alert('出欠確認を締め切りました');
      // Reload data
      const collectionData = await getAttendanceCollection(collectionId);
      setCollection(collectionData);
    } catch (err) {
      alert(err instanceof Error ? err.message : '締切に失敗しました');
    } finally {
      setClosing(false);
    }
  };

  const handleCopyUrl = () => {
    if (!collection) return;
    const url = `${window.location.origin}/p/attendance/${collection.public_token}`;
    navigator.clipboard.writeText(url);
    alert('URLをコピーしました');
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-2 text-gray-600">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-800">{error}</p>
          <button
            onClick={() => navigate('/attendance')}
            className="mt-2 text-blue-600 hover:underline"
          >
            出欠確認一覧に戻る
          </button>
        </div>
      </div>
    );
  }

  if (!collection) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
          <p className="text-gray-600">出欠確認が見つかりません</p>
        </div>
      </div>
    );
  }

  // 日付ごとに回答を集計
  const aggregateByDate = () => {
    const dateMap = new Map<string, {
      targetDateId: string;
      targetDate: string;
      attending: AttendanceResponse[];
      absent: AttendanceResponse[];
    }>();

    responses.forEach((resp) => {
      if (!dateMap.has(resp.target_date_id)) {
        dateMap.set(resp.target_date_id, {
          targetDateId: resp.target_date_id,
          targetDate: resp.target_date,
          attending: [],
          absent: [],
        });
      }
      const dateData = dateMap.get(resp.target_date_id)!;
      if (resp.response === 'attending') {
        dateData.attending.push(resp);
      } else {
        dateData.absent.push(resp);
      }
    });

    return Array.from(dateMap.values()).sort(
      (a, b) => new Date(a.targetDate).getTime() - new Date(b.targetDate).getTime()
    );
  };

  const aggregatedData = aggregateByDate();
  const publicUrl = `${window.location.origin}/p/attendance/${collection.public_token}`;

  return (
    <div className="max-w-6xl mx-auto">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">出欠確認の詳細</h1>
        <button
          onClick={() => navigate('/attendance')}
          className="text-blue-600 hover:underline"
        >
          一覧に戻る
        </button>
      </div>

      {/* 出欠確認情報 */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <h2 className="text-xl font-semibold text-gray-900 mb-2">{collection.title}</h2>
            {collection.description && (
              <p className="text-gray-600 whitespace-pre-wrap mb-4">{collection.description}</p>
            )}
            <div className="flex items-center gap-4 text-sm text-gray-600">
              <span
                className={`px-3 py-1 rounded-full ${
                  collection.status === 'open'
                    ? 'bg-green-100 text-green-800'
                    : 'bg-gray-100 text-gray-800'
                }`}
              >
                {collection.status === 'open' ? '受付中' : '締切済み'}
              </span>
              {collection.deadline && (
                <span>締切: {new Date(collection.deadline).toLocaleString('ja-JP')}</span>
              )}
            </div>
          </div>
          {collection.status === 'open' && (
            <button
              onClick={handleClose}
              disabled={closing}
              className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition disabled:bg-gray-400"
            >
              {closing ? '処理中...' : '締め切る'}
            </button>
          )}
        </div>

        <div className="border-t pt-4">
          <p className="text-sm font-medium text-gray-700 mb-2">公開URL:</p>
          <div className="flex gap-2">
            <input
              type="text"
              value={publicUrl}
              readOnly
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm"
            />
            <button
              onClick={handleCopyUrl}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
            >
              コピー
            </button>
          </div>
        </div>
      </div>

      {/* 回答集計 */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">回答状況</h2>

        {aggregatedData.length === 0 ? (
          <p className="text-gray-600">まだ回答がありません</p>
        ) : (
          <div className="space-y-6">
            {aggregatedData.map((dateData) => (
              <div key={dateData.targetDateId} className="border-b pb-6 last:border-b-0">
                <h3 className="font-medium text-gray-900 mb-3">
                  {new Date(dateData.targetDate).toLocaleDateString('ja-JP', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                    weekday: 'short',
                  })}
                </h3>

                <div className="grid md:grid-cols-2 gap-4">
                  {/* 参加者 */}
                  <div className="bg-green-50 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-medium text-green-900">参加</h4>
                      <span className="text-sm text-green-700">
                        {dateData.attending.length}人
                      </span>
                    </div>
                    {dateData.attending.length > 0 ? (
                      <ul className="space-y-1">
                        {dateData.attending.map((resp) => (
                          <li key={resp.response_id} className="text-sm text-green-800">
                            • {resp.member_name}
                            {resp.note && (
                              <span className="text-green-600 ml-2">({resp.note})</span>
                            )}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-sm text-green-600">なし</p>
                    )}
                  </div>

                  {/* 不参加者 */}
                  <div className="bg-red-50 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-medium text-red-900">不参加</h4>
                      <span className="text-sm text-red-700">{dateData.absent.length}人</span>
                    </div>
                    {dateData.absent.length > 0 ? (
                      <ul className="space-y-1">
                        {dateData.absent.map((resp) => (
                          <li key={resp.response_id} className="text-sm text-red-800">
                            • {resp.member_name}
                            {resp.note && (
                              <span className="text-red-600 ml-2">({resp.note})</span>
                            )}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-sm text-red-600">なし</p>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
