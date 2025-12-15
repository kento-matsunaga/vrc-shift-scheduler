import { useState, useEffect } from 'react';
import { getMembers, createMember } from '../lib/api/memberApi';
import type { Member } from '../types/api';

export default function Members() {
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // 新規登録フォーム
  const [showForm, setShowForm] = useState(false);
  const [displayName, setDisplayName] = useState('');
  const [discordUserId, setDiscordUserId] = useState('');
  const [email, setEmail] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // メンバー一覧を取得
  const fetchMembers = async () => {
    try {
      setLoading(true);
      const response = await getMembers();
      setMembers(response.members || []);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch members:', err);
      setError('メンバー一覧の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMembers();
  }, []);

  // メンバー登録
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!displayName.trim()) {
      alert('表示名は必須です');
      return;
    }

    try {
      setSubmitting(true);
      await createMember({
        display_name: displayName.trim(),
        discord_user_id: discordUserId.trim() || undefined,
        email: email.trim() || undefined,
      });
      
      // フォームをリセット
      setDisplayName('');
      setDiscordUserId('');
      setEmail('');
      setShowForm(false);
      
      // 一覧を再取得
      await fetchMembers();
    } catch (err: any) {
      console.error('Failed to create member:', err);
      if (err?.response?.data?.error?.message) {
        alert(err.response.data.error.message);
      } else {
        alert('メンバーの登録に失敗しました');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">読み込み中...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <p className="text-red-700">{error}</p>
        <button 
          onClick={fetchMembers}
          className="mt-2 text-red-600 hover:text-red-800 text-sm underline"
        >
          再試行
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* ヘッダー */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">メンバー管理</h2>
          <p className="mt-1 text-sm text-gray-500">
            シフトや出欠確認に参加するキャストを管理します
          </p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="btn-primary"
        >
          {showForm ? 'キャンセル' : '+ 新規登録'}
        </button>
      </div>

      {/* 新規登録フォーム */}
      {showForm && (
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">新規メンバー登録</h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="displayName" className="block text-sm font-medium text-gray-700 mb-1">
                表示名 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="displayName"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="例: 山田太郎"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                maxLength={50}
                required
              />
              <p className="mt-1 text-xs text-gray-500">50文字以内</p>
            </div>

            <div>
              <label htmlFor="discordUserId" className="block text-sm font-medium text-gray-700 mb-1">
                Discord ユーザーID（任意）
              </label>
              <input
                type="text"
                id="discordUserId"
                value={discordUserId}
                onChange={(e) => setDiscordUserId(e.target.value)}
                placeholder="例: 123456789012345678"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
              <p className="mt-1 text-xs text-gray-500">Discord連携に使用します</p>
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                メールアドレス（任意）
              </label>
              <input
                type="email"
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="例: example@mail.com"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div className="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                onClick={() => setShowForm(false)}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                キャンセル
              </button>
              <button
                type="submit"
                disabled={submitting || !displayName.trim()}
                className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {submitting ? '登録中...' : '登録'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* メンバー一覧 */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">
            登録済みメンバー（{members.length}名）
          </h3>
        </div>

        {members.length === 0 ? (
          <div className="px-6 py-12 text-center">
            <div className="text-gray-400 text-4xl mb-4">👥</div>
            <p className="text-gray-500">メンバーがまだ登録されていません</p>
            <p className="text-sm text-gray-400 mt-2">
              「新規登録」ボタンからメンバーを追加してください
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {members.map((member) => (
              <div
                key={member.member_id}
                className="px-6 py-4 flex items-center justify-between hover:bg-gray-50"
              >
                <div className="flex items-center space-x-4">
                  <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
                    <span className="text-blue-600 font-semibold">
                      {member.display_name.charAt(0)}
                    </span>
                  </div>
                  <div>
                    <div className="font-medium text-gray-900">{member.display_name}</div>
                    <div className="text-sm text-gray-500 space-x-3">
                      {member.discord_user_id && (
                        <span>Discord: {member.discord_user_id}</span>
                      )}
                      {member.email && (
                        <span>📧 {member.email}</span>
                      )}
                    </div>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <span
                    className={`px-2 py-1 text-xs rounded-full ${
                      member.is_active
                        ? 'bg-green-100 text-green-700'
                        : 'bg-gray-100 text-gray-500'
                    }`}
                  >
                    {member.is_active ? '有効' : '無効'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

