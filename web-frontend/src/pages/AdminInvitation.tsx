import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { inviteAdmin } from '../lib/api/invitationApi';

export default function AdminInvitation() {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('admin');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [inviteUrl, setInviteUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess(false);

    if (!email.trim()) {
      setError('メールアドレスを入力してください');
      return;
    }

    if (!role) {
      setError('ロールを選択してください');
      return;
    }

    setLoading(true);

    try {
      const result = await inviteAdmin({
        email: email.trim(),
        role: role,
      });

      // 招待URLを生成
      const baseUrl = window.location.origin;
      const url = `${baseUrl}/invite/${result.token}`;
      setInviteUrl(url);
      setSuccess(true);

      // フォームをクリア
      setEmail('');
      setRole('admin');
    } catch (err) {
      if (err instanceof Error) {
        // エラーメッセージに基づいて日本語表示
        if (err.message.includes('認証が必要です')) {
          setError('認証が必要です。再度ログインしてください。');
          setTimeout(() => navigate('/admin/login'), 2000);
        } else if (err.message.includes('Invalid or expired token')) {
          setError('トークンが無効または期限切れです。再度ログインしてください。');
        } else {
          setError(err.message);
        }
      } else {
        setError('招待に失敗しました。もう一度お試しください。');
      }
      console.error('Invitation error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(inviteUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  const handleBackToEvents = () => {
    navigate('/events');
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col">
      {/* ヘッダー */}
      <header className="bg-indigo-900 text-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
        </div>
      </header>

      {/* メインコンテンツ */}
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-md max-w-md w-full p-8">
          <div className="text-center mb-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              管理者を招待
            </h2>
            <p className="text-sm text-gray-500">
              新しい管理者を招待して、チームを拡大しましょう
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1.5">
                メールアドレス
              </label>
              <input
                type="email"
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="newadmin@example.com"
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition"
                disabled={loading}
                autoFocus
              />
            </div>

            <div>
              <label htmlFor="role" className="block text-sm font-medium text-gray-700 mb-1.5">
                ロール
              </label>
              <select
                id="role"
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition"
                disabled={loading}
              >
                <option value="admin">管理者 (Admin)</option>
                <option value="manager">マネージャー (Manager)</option>
              </select>
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            {success && inviteUrl && (
              <div className="bg-green-50 border border-green-200 rounded-md p-4 space-y-3">
                <p className="text-sm text-green-700 font-medium">
                  招待メールを送信しました！
                </p>
                <div className="bg-gray-50 rounded p-2 break-all">
                  <p className="text-xs text-gray-500 mb-2">招待URL:</p>
                  <p className="text-sm text-gray-900 font-mono">{inviteUrl}</p>
                </div>
                <button
                  type="button"
                  onClick={handleCopy}
                  className="w-full py-2 px-4 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-md transition-colors"
                >
                  {copied ? '✓ コピーしました' : 'URLをコピー'}
                </button>
              </div>
            )}

            <button
              type="submit"
              className="w-full py-3 px-4 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-400 text-white font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
              disabled={loading || !email.trim() || !role}
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  招待中...
                </span>
              ) : (
                '招待を送信'
              )}
            </button>
          </form>

          <div className="mt-6 pt-6 border-t border-gray-200">
            <button
              onClick={handleBackToEvents}
              className="w-full py-2 px-4 text-sm text-gray-600 hover:text-gray-900 transition-colors"
            >
              ← イベント一覧に戻る
            </button>
          </div>
        </div>
      </main>

      {/* フッター */}
      <footer className="bg-white border-t">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-500">
            VRC Shift Scheduler
          </p>
        </div>
      </footer>
    </div>
  );
}
