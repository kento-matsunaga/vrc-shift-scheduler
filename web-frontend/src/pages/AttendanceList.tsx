import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  listAttendanceCollections,
  createAttendanceCollection,
  type AttendanceCollection,
} from '../lib/api/attendanceApi';
import { getMemberGroups, type MemberGroup } from '../lib/api/memberGroupApi';
import { MobileCard, CardHeader, CardField } from '../components/MobileCard';

export default function AttendanceList() {
  const navigate = useNavigate();
  const [collections, setCollections] = useState<AttendanceCollection[]>([]);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [deadline, setDeadline] = useState('');
  const [targetDates, setTargetDates] = useState<string[]>(['', '', '']);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [createdCollection, setCreatedCollection] = useState<AttendanceCollection | null>(null);
  const [publicUrl, setPublicUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const [submittedDatesCount, setSubmittedDatesCount] = useState(0);
  const [memberGroups, setMemberGroups] = useState<MemberGroup[]>([]);
  const [selectedGroupIds, setSelectedGroupIds] = useState<string[]>([]);

  useEffect(() => {
    loadCollections();
    loadMemberGroups();
  }, []);

  const loadMemberGroups = async () => {
    try {
      const data = await getMemberGroups();
      setMemberGroups(data.groups || []);
    } catch (err) {
      console.error('Failed to load member groups:', err);
    }
  };

  const loadCollections = async () => {
    try {
      setLoading(true);
      const data = await listAttendanceCollections();
      setCollections(data || []);
    } catch (err) {
      console.error('Failed to load collections:', err);
      setError('出欠確認一覧の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleAddDate = () => {
    setTargetDates([...targetDates, '']);
  };

  const handleRemoveDate = (index: number) => {
    if (targetDates.length > 1) {
      setTargetDates(targetDates.filter((_, i) => i !== index));
    }
  };

  const handleDateChange = (index: number, value: string) => {
    const newDates = [...targetDates];
    newDates[index] = value;
    setTargetDates(newDates);
  };

  const toggleGroupSelection = (groupId: string) => {
    setSelectedGroupIds((prev) =>
      prev.includes(groupId)
        ? prev.filter((id) => id !== groupId)
        : [...prev, groupId]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setCreatedCollection(null);

    if (!title.trim()) {
      setError('タイトルを入力してください');
      return;
    }

    const validDates = targetDates.filter((d) => d.trim() !== '');
    if (validDates.length === 0) {
      setError('対象日を1つ以上入力してください');
      return;
    }

    setSubmitting(true);

    try {
      setSubmittedDatesCount(validDates.length);

      const result = await createAttendanceCollection({
        title: title.trim(),
        description: description.trim(),
        target_type: 'business_day',
        target_dates: validDates.map((d) => new Date(d).toISOString()),
        deadline: deadline ? new Date(deadline).toISOString() : undefined,
        group_ids: selectedGroupIds.length > 0 ? selectedGroupIds : undefined,
      });

      const baseUrl = window.location.origin;
      const url = `${baseUrl}/p/attendance/${result.public_token}`;
      setPublicUrl(url);
      setCreatedCollection(result);

      setTitle('');
      setDescription('');
      setDeadline('');
      setTargetDates(['', '', '']);
      setSelectedGroupIds([]);
      setShowCreateForm(false);

      loadCollections();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('出欠確認の作成に失敗しました');
      }
      console.error('Create collection error:', err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(publicUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'open':
        return <span className="px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">受付中</span>;
      case 'closed':
        return <span className="px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">締切済み</span>;
      default:
        return <span className="px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">{status}</span>;
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

  return (
    <div className="max-w-6xl mx-auto">
      <div className="mb-6 flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold text-gray-900">出欠確認</h1>
          <p className="text-xs sm:text-sm text-gray-600 mt-1">
            イベントやシフトの出欠確認を作成して、メンバーに回答してもらいましょう
          </p>
        </div>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent-dark transition-colors font-medium text-sm sm:text-base w-full sm:w-auto"
        >
          {showCreateForm ? 'キャンセル' : '+ 新規作成'}
        </button>
      </div>

      {showCreateForm && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            新しい出欠確認を作成
          </h2>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                タイトル <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="例：12月のシフト出欠確認"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={submitting}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                説明
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                placeholder="詳細な説明や注意事項を入力してください"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={submitting}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                対象日 <span className="text-red-500">*</span>
              </label>
              <div className="space-y-2">
                {targetDates.map((date, index) => (
                  <div key={index} className="flex gap-2">
                    <input
                      type="date"
                      value={date}
                      onChange={(e) => handleDateChange(index, e.target.value)}
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                      disabled={submitting}
                    />
                    {targetDates.length > 1 && (
                      <button
                        type="button"
                        onClick={() => handleRemoveDate(index)}
                        className="px-3 py-2 text-red-600 hover:bg-red-50 rounded-md transition"
                        disabled={submitting}
                      >
                        削除
                      </button>
                    )}
                  </div>
                ))}
              </div>
              <button
                type="button"
                onClick={handleAddDate}
                className="mt-2 px-3 py-1 text-sm text-accent hover:bg-accent/10 rounded-md transition"
                disabled={submitting}
              >
                + 対象日を追加
              </button>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                回答締切（任意）
              </label>
              <input
                type="datetime-local"
                value={deadline}
                onChange={(e) => setDeadline(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={submitting}
              />
            </div>

            {memberGroups.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  対象メンバーグループ（任意）
                </label>
                <p className="text-xs text-gray-500 mb-2">
                  選択すると、そのグループに属するメンバーのみが回答可能になります
                </p>
                <div className="flex flex-wrap gap-2">
                  {memberGroups.map((group) => (
                    <button
                      key={group.group_id}
                      type="button"
                      onClick={() => toggleGroupSelection(group.group_id)}
                      disabled={submitting}
                      className={`px-3 py-1.5 rounded-full text-sm font-medium transition ${
                        selectedGroupIds.includes(group.group_id)
                          ? 'bg-accent text-white'
                          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      }`}
                      style={
                        selectedGroupIds.includes(group.group_id) && group.color
                          ? { backgroundColor: group.color }
                          : undefined
                      }
                    >
                      {group.name}
                    </button>
                  ))}
                </div>
                {selectedGroupIds.length > 0 && (
                  <p className="mt-2 text-xs text-accent">
                    {selectedGroupIds.length}個のグループを選択中
                  </p>
                )}
              </div>
            )}

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3">
                <p className="text-sm text-red-800">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={submitting || !title.trim()}
              className="w-full px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {submitting ? '作成中...' : '出欠確認を作成'}
            </button>
          </form>
        </div>
      )}

      {createdCollection && publicUrl && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-6 mb-6">
          <div className="flex items-start">
            <div className="text-green-500 text-2xl mr-3">✓</div>
            <div className="flex-1">
              <h3 className="text-lg font-semibold text-green-900 mb-2">
                出欠確認を作成しました
              </h3>
              <p className="text-sm text-green-800 mb-4">
                以下のURLをメンバーに送信してください
              </p>

              <div className="bg-white rounded-md p-3 mb-3 border border-green-300">
                <p className="text-xs text-gray-600 mb-1">公開URL:</p>
                <p className="text-sm text-gray-900 font-mono break-all">{publicUrl}</p>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={handleCopy}
                  className="flex-1 px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition text-sm"
                >
                  {copied ? '✓ コピーしました' : 'URLをコピー'}
                </button>
                <a
                  href={publicUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition text-sm text-center"
                >
                  プレビュー
                </a>
              </div>

              <div className="mt-4 pt-4 border-t border-green-200">
                <p className="text-xs text-green-700">
                  <strong>対象日:</strong> {submittedDatesCount}件
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* モバイル用カードビュー */}
      <div className="md:hidden space-y-3">
        {collections.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-8 text-center text-gray-500">
            出欠確認がまだありません。新規作成してください。
          </div>
        ) : (
          collections.map((collection) => (
            <MobileCard
              key={collection.collection_id}
              onClick={() => navigate(`/attendance/${collection.collection_id}`)}
            >
              <CardHeader
                title={collection.title}
                subtitle={collection.description || undefined}
                badge={getStatusBadge(collection.status)}
              />
              <div className="space-y-1">
                <CardField label="対象日数" value={`${collection.target_date_count || 0}件`} />
                <CardField label="回答数" value={`${collection.response_count || 0}人`} />
                <CardField
                  label="締切"
                  value={
                    collection.deadline
                      ? new Date(collection.deadline).toLocaleString('ja-JP', {
                          month: '2-digit',
                          day: '2-digit',
                          hour: '2-digit',
                          minute: '2-digit',
                        })
                      : '-'
                  }
                />
                <CardField
                  label="作成日"
                  value={new Date(collection.created_at).toLocaleDateString('ja-JP')}
                />
              </div>
            </MobileCard>
          ))
        )}
      </div>

      {/* デスクトップ用テーブルビュー */}
      <div className="hidden md:block bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  タイトル
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  ステータス
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  対象日数
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  回答数
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  締切
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  作成日
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {collections.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-12 text-center text-gray-500">
                    出欠確認がまだありません。新規作成してください。
                  </td>
                </tr>
              ) : (
                collections.map((collection) => (
                  <tr key={collection.collection_id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div>
                        <div className="text-sm font-medium text-gray-900">{collection.title}</div>
                        {collection.description && (
                          <div className="text-sm text-gray-500 truncate max-w-md">{collection.description}</div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(collection.status)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {collection.target_date_count || 0}件
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {collection.response_count || 0}人
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {collection.deadline
                        ? new Date(collection.deadline).toLocaleString('ja-JP', {
                            year: 'numeric',
                            month: '2-digit',
                            day: '2-digit',
                            hour: '2-digit',
                            minute: '2-digit',
                          })
                        : '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(collection.created_at).toLocaleDateString('ja-JP')}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => navigate(`/attendance/${collection.collection_id}`)}
                        className="text-accent hover:text-accent-dark transition"
                      >
                        詳細
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      <div className="mt-6 p-4 bg-accent/10 border border-accent/30 rounded-lg">
        <h3 className="text-sm font-semibold text-accent-dark mb-2">💡 使い方</h3>
        <ul className="text-sm text-accent-dark space-y-1 list-disc list-inside">
          <li>出欠確認を作成すると公開URLが発行されます</li>
          <li>複数の対象日を設定して、メンバーに各日の出欠を回答してもらえます</li>
          <li>URLをメンバーに送信して、各日の出欠を回答してもらいましょう</li>
          <li>締切を設定すると、締切後は回答できなくなります</li>
          <li>詳細画面で回答状況を確認できます</li>
        </ul>
      </div>
    </div>
  );
}
