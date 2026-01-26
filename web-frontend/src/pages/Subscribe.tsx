import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useDocumentTitle } from '../hooks/useDocumentTitle';
import { getErrorMessage, isRateLimitError } from '../utils/errorHandler';

interface SubscribeResponse {
  checkout_url: string;
  session_id: string;
  tenant_id: string;
  expires_at: number;
  message: string;
}

export default function Subscribe() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [tenantName, setTenantName] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [agreedToTerms, setAgreedToTerms] = useState(false);

  useDocumentTitle('新規登録');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email.trim()) {
      setError('メールアドレスを入力してください');
      return;
    }

    if (!password) {
      setError('パスワードを入力してください');
      return;
    }

    if (password.length < 8) {
      setError('パスワードは8文字以上で入力してください');
      return;
    }

    if (password !== passwordConfirm) {
      setError('パスワードが一致しません');
      return;
    }

    if (!tenantName.trim()) {
      setError('組織名を入力してください');
      return;
    }

    if (!displayName.trim()) {
      setError('表示名を入力してください');
      return;
    }

    if (!agreedToTerms) {
      setError('利用規約とプライバシーポリシーに同意してください');
      return;
    }

    setLoading(true);

    try {
      const response = await fetch('/api/v1/public/subscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: email.trim(),
          password: password,
          tenant_name: tenantName.trim(),
          display_name: displayName.trim(),
          timezone: 'Asia/Tokyo',
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        // Use centralized error handler for better error messages
        const errorMessage = getErrorMessage(data, '登録に失敗しました');

        // Check for rate limiting - provide more specific guidance
        if (isRateLimitError(data)) {
          setError('リクエストが多すぎます。1分ほど待ってから再度お試しください。');
        } else {
          setError(errorMessage);
        }
        console.error('Subscribe error:', data);
        return;
      }

      const result: SubscribeResponse = data.data;

      // Redirect to Stripe Checkout
      window.location.href = result.checkout_url;
    } catch (err) {
      // Network errors or other unexpected errors
      console.error('Subscribe error:', err);
      setError('通信エラーが発生しました。ネットワーク接続を確認して再度お試しください。');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="min-h-screen text-white"
      style={{
        background: 'linear-gradient(180deg, #0a0a0f 0%, #0f0f1a 50%, #0a0a0f 100%)',
        fontFamily: '"Noto Sans JP", "Inter", system-ui, sans-serif',
      }}
    >
      {/* Header */}
      <header
        className="fixed top-0 left-0 right-0 z-50 py-4"
        style={{
          background: 'rgba(10, 10, 15, 0.85)',
          backdropFilter: 'blur(20px)',
          WebkitBackdropFilter: 'blur(20px)',
          borderBottom: '1px solid rgba(139, 92, 246, 0.1)',
        }}
      >
        <div className="max-w-4xl mx-auto px-4 sm:px-6 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 sm:gap-3 group">
            <div
              className="w-8 h-8 sm:w-10 sm:h-10 rounded-xl flex items-center justify-center"
              style={{
                background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
              }}
            >
              <span className="text-base sm:text-xl">📅</span>
            </div>
            <span className="font-bold text-sm sm:text-lg text-white">VRC Shift Scheduler</span>
          </Link>
          <Link
            to="/"
            className="text-gray-400 hover:text-white transition-colors text-sm"
          >
            トップに戻る
          </Link>
        </div>
      </header>

      {/* Content */}
      <main className="pt-24 pb-16 px-4 sm:px-6">
        <div className="max-w-md mx-auto">
          <div className="text-center mb-8">
            <h1 className="text-2xl sm:text-3xl font-bold mb-2">新規登録</h1>
            <p className="text-gray-400 text-sm">
              アカウントを作成して、シフト管理を始めましょう
            </p>
            <div
              className="inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm mt-4"
              style={{
                background: 'rgba(139, 92, 246, 0.15)',
                border: '1px solid rgba(139, 92, 246, 0.3)',
              }}
            >
              <span className="text-violet-300">初期キャンペーン: 月額200円</span>
            </div>
          </div>

          <div
            className="rounded-2xl p-6 sm:p-8"
            style={{
              background: 'rgba(255, 255, 255, 0.03)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
            }}
          >
            <form onSubmit={handleSubmit} className="space-y-5">
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-300 mb-1.5">
                  メールアドレス
                </label>
                <input
                  type="email"
                  id="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="example@example.com"
                  className="w-full px-4 py-3 rounded-lg bg-white/5 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent transition"
                  disabled={loading}
                  autoFocus
                />
              </div>

              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-300 mb-1.5">
                  パスワード
                </label>
                <input
                  type="password"
                  id="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="8文字以上"
                  className="w-full px-4 py-3 rounded-lg bg-white/5 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent transition"
                  disabled={loading}
                />
              </div>

              <div>
                <label htmlFor="passwordConfirm" className="block text-sm font-medium text-gray-300 mb-1.5">
                  パスワード（確認）
                </label>
                <input
                  type="password"
                  id="passwordConfirm"
                  value={passwordConfirm}
                  onChange={(e) => setPasswordConfirm(e.target.value)}
                  placeholder="パスワードを再入力"
                  className="w-full px-4 py-3 rounded-lg bg-white/5 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent transition"
                  disabled={loading}
                />
              </div>

              <div>
                <label htmlFor="tenantName" className="block text-sm font-medium text-gray-300 mb-1.5">
                  組織名（イベント名など）
                </label>
                <input
                  type="text"
                  id="tenantName"
                  value={tenantName}
                  onChange={(e) => setTenantName(e.target.value)}
                  placeholder="例: VRCイベント運営チーム"
                  className="w-full px-4 py-3 rounded-lg bg-white/5 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent transition"
                  disabled={loading}
                />
              </div>

              <div>
                <label htmlFor="displayName" className="block text-sm font-medium text-gray-300 mb-1.5">
                  あなたの表示名
                </label>
                <input
                  type="text"
                  id="displayName"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="例: 山田 太郎"
                  className="w-full px-4 py-3 rounded-lg bg-white/5 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent transition"
                  disabled={loading}
                />
              </div>

              <div className="flex items-start gap-3 pt-2">
                <input
                  type="checkbox"
                  id="terms"
                  checked={agreedToTerms}
                  onChange={(e) => setAgreedToTerms(e.target.checked)}
                  className="mt-1 w-4 h-4 rounded border-gray-600 text-violet-500 focus:ring-violet-500 focus:ring-offset-gray-900"
                  disabled={loading}
                />
                <label htmlFor="terms" className="text-sm text-gray-400">
                  <Link to="/terms" className="text-violet-400 hover:text-violet-300 underline underline-offset-2" target="_blank">
                    利用規約
                  </Link>
                  と
                  <Link to="/privacy" className="text-violet-400 hover:text-violet-300 underline underline-offset-2" target="_blank">
                    プライバシーポリシー
                  </Link>
                  に同意します
                </label>
              </div>

              {error && (
                <div
                  className="rounded-lg p-3"
                  style={{
                    background: 'rgba(239, 68, 68, 0.1)',
                    border: '1px solid rgba(239, 68, 68, 0.3)',
                  }}
                >
                  <p className="text-sm text-red-400">{error}</p>
                </div>
              )}

              <button
                type="submit"
                className="w-full py-4 px-4 rounded-xl font-semibold text-white transition-all duration-300 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
                style={{
                  background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
                  boxShadow: '0 8px 40px rgba(79, 70, 229, 0.3)',
                }}
                disabled={loading || !email.trim() || !password || !passwordConfirm || !tenantName.trim() || !displayName.trim() || !agreedToTerms}
              >
                {loading ? (
                  <span className="flex items-center justify-center gap-2">
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    処理中...
                  </span>
                ) : (
                  '決済に進む'
                )}
              </button>
            </form>
          </div>

          <div className="mt-6 text-center text-sm text-gray-500">
            既にアカウントをお持ちの方は{' '}
            <Link to="/admin/login" className="text-violet-400 hover:text-violet-300 underline underline-offset-2">
              ログイン
            </Link>
          </div>

          <div className="mt-8 text-center">
            <Link
              to="/"
              className="inline-flex items-center gap-2 text-violet-400 hover:text-violet-300 transition-colors"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              トップページに戻る
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
