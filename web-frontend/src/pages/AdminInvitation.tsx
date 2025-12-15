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
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center p-4">
      <div className="bg-white/10 backdrop-blur-lg border border-white/20 rounded-2xl shadow-2xl max-w-md w-full p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">
            管理者を招待
          </h1>
          <p className="text-sm text-gray-300">
            新しい管理者を招待して、チームを拡大しましょう
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-200 mb-1.5">
              メールアドレス
            </label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="newadmin@example.com"
              className="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition"
              disabled={loading}
              autoFocus
            />
          </div>

          <div>
            <label htmlFor="role" className="block text-sm font-medium text-gray-200 mb-1.5">
              ロール
            </label>
            <select
              id="role"
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition"
              disabled={loading}
            >
              <option value="admin" className="bg-slate-800">管理者 (Admin)</option>
              <option value="manager" className="bg-slate-800">マネージャー (Manager)</option>
            </select>
          </div>

          {error && (
            <div className="bg-red-500/20 border border-red-500/50 rounded-lg p-3">
              <p className="text-sm text-red-200">{error}</p>
            </div>
          )}

          {success && inviteUrl && (
            <div className="bg-green-500/20 border border-green-500/50 rounded-lg p-4 space-y-3">
              <p className="text-sm text-green-200 font-medium">
                招待メールを送信しました！
              </p>
              <div className="bg-white/10 rounded p-2 break-all">
                <p className="text-xs text-gray-300 mb-2">招待URL:</p>
                <p className="text-sm text-white font-mono">{inviteUrl}</p>
              </div>
              <button
                type="button"
                onClick={handleCopy}
                className="w-full py-2 px-4 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-lg transition-colors"
              >
                {copied ? '✓ コピーしました' : 'URLをコピー'}
              </button>
            </div>
          )}

          <button
            type="submit"
            className="w-full py-3 px-4 bg-purple-600 hover:bg-purple-700 disabled:bg-purple-600/50 text-white font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 focus:ring-offset-slate-900"
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

        <div className="mt-6 pt-6 border-t border-white/10">
          <button
            onClick={handleBackToEvents}
            className="w-full py-2 px-4 text-sm text-gray-300 hover:text-white transition-colors"
          >
            ← イベント一覧に戻る
          </button>
        </div>
      </div>
    </div>
  );
}
