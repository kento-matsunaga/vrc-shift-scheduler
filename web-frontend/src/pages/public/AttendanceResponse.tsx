import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  getAttendanceByToken,
  getMembers,
  submitAttendanceResponse,
  type AttendanceCollection,
  type Member,
  PublicApiError,
} from '../../lib/api/publicApi';

export default function AttendanceResponse() {
  const { token } = useParams<{ token: string }>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [collection, setCollection] = useState<AttendanceCollection | null>(null);
  const [members, setMembers] = useState<Member[]>([]);

  // フォーム状態
  const [selectedMemberId, setSelectedMemberId] = useState('');
  const [response, setResponse] = useState<'attending' | 'absent'>('attending');
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

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

        // 出欠確認情報を取得
        const collectionData = await getAttendanceByToken(token);
        setCollection(collectionData);

        // メンバー一覧を取得
        const membersData = await getMembers(collectionData.tenant_id);
        setMembers(membersData.data?.members || []);
      } catch (err) {
        if (err instanceof PublicApiError) {
          if (err.isNotFound()) {
            setError('出欠確認が見つかりません。URLをご確認ください。');
          } else if (err.isForbidden()) {
            setError('この出欠確認は既に締め切られています。');
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedMemberId) {
      alert('お名前を選択してください');
      return;
    }

    if (!token) return;

    try {
      setSubmitting(true);
      setError(null);

      await submitAttendanceResponse(token, {
        member_id: selectedMemberId,
        response,
        note,
      });

      setSubmitted(true);
    } catch (err) {
      if (err instanceof PublicApiError) {
        if (err.isNotFound()) {
          setError('出欠確認が見つかりません。');
        } else if (err.isBadRequest()) {
          setError('入力内容に誤りがあります。');
        } else if (err.isForbidden()) {
          setError('この出欠確認は既に締め切られています。');
        } else {
          setError('送信に失敗しました。');
        }
      } else {
        setError('通信エラーが発生しました。');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (error && !collection) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-6">
          <div className="text-center">
            <div className="text-red-500 text-5xl mb-4">⚠️</div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">エラー</h2>
            <p className="text-gray-600">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (submitted) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-6">
          <div className="text-center">
            <div className="text-green-500 text-5xl mb-4">✓</div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">送信完了</h2>
            <p className="text-gray-600 mb-4">
              出欠回答を受け付けました。
            </p>
            <button
              onClick={() => {
                setSubmitted(false);
                setSelectedMemberId('');
                setResponse('attending');
                setNote('');
              }}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
            >
              別の回答を送信
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4">
      <div className="max-w-2xl mx-auto">
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {collection?.title}
          </h1>
          {collection?.description && (
            <p className="text-gray-600 mb-4 whitespace-pre-wrap">
              {collection.description}
            </p>
          )}
          {collection?.deadline && (
            <p className="text-sm text-gray-500">
              締切: {new Date(collection.deadline).toLocaleString('ja-JP')}
            </p>
          )}
          {collection?.status === 'closed' && (
            <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <p className="text-yellow-800 text-sm">
                この出欠確認は締め切られています
              </p>
            </div>
          )}
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            出欠を回答する
          </h2>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-red-800 text-sm">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                お名前 <span className="text-red-500">*</span>
              </label>
              <select
                value={selectedMemberId}
                onChange={(e) => setSelectedMemberId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
                disabled={collection?.status === 'closed'}
              >
                <option value="">選択してください</option>
                {members.map((member) => (
                  <option key={member.member_id} value={member.member_id}>
                    {member.display_name}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                お名前が見つからない場合は、管理者にお問い合わせください
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                出欠 <span className="text-red-500">*</span>
              </label>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="radio"
                    value="attending"
                    checked={response === 'attending'}
                    onChange={(e) => setResponse(e.target.value as 'attending')}
                    className="mr-2"
                    disabled={collection?.status === 'closed'}
                  />
                  <span className="text-gray-700">参加</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    value="absent"
                    checked={response === 'absent'}
                    onChange={(e) => setResponse(e.target.value as 'absent')}
                    className="mr-2"
                    disabled={collection?.status === 'closed'}
                  />
                  <span className="text-gray-700">不参加</span>
                </label>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                備考（任意）
              </label>
              <textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="補足事項があれば入力してください"
                disabled={collection?.status === 'closed'}
              />
            </div>

            <button
              type="submit"
              disabled={submitting || collection?.status === 'closed'}
              className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {submitting ? '送信中...' : '回答を送信'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
