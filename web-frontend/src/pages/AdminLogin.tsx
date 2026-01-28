import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login } from '../lib/api/authApi';
import { useDocumentTitle } from '../hooks/useDocumentTitle';
import { SEO } from '../components/seo/SEO';

export default function AdminLogin() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useDocumentTitle('ログイン');

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

    setLoading(true);

    try {
      // ログインAPI呼び出し
      const result = await login({
        email: email.trim(),
        password: password,
      });

      // JWTトークンを localStorage に保存
      localStorage.setItem('auth_token', result.token);
      localStorage.setItem('admin_id', result.admin_id);
      localStorage.setItem('tenant_id', result.tenant_id);
      localStorage.setItem('admin_role', result.role);

      // 管理画面に遷移（ページリロードで認証状態を再初期化）
      window.location.href = '/events';
    } catch (err) {
      if (err instanceof Error) {
        // エラーメッセージに基づいて日本語表示
        if (err.message.includes('Invalid email or password')) {
          setError('メールアドレスまたはパスワードが正しくありません');
        } else if (err.message.includes('Account is disabled')) {
          setError('このアカウントは無効化されています');
        } else {
          setError(err.message);
        }
      } else {
        setError('ログインに失敗しました。もう一度お試しください。');
      }
      console.error('Login error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <SEO noindex={true} />
      <div className="min-h-screen bg-gray-100 flex flex-col">
        {/* ヘッダー */}
      <header className="bg-vrc-dark text-white shadow-soft">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
        </div>
      </header>

      {/* メインコンテンツ */}
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="bg-white rounded-card shadow-soft max-w-md w-full p-10 border border-gray-200">
          <div className="text-center mb-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              ログイン
            </h2>
            <p className="text-sm text-gray-500">
              管理者アカウントでログインしてください
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
                placeholder="admin@example.com"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 shadow-inset-input focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={loading}
                autoFocus
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1.5">
                パスワード
              </label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 shadow-inset-input focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={loading}
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <button
              type="submit"
              className="w-full py-3 px-4 bg-gradient-to-b from-accent-light to-accent-dark hover:from-accent-hover hover:to-accent disabled:from-gray-400 disabled:to-gray-500 text-white font-semibold rounded-lg border border-accent-dark shadow-inset-light transition-all focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2"
              disabled={loading || !email.trim() || !password}
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  ログイン中...
                </span>
              ) : (
                'ログイン'
              )}
            </button>

            <div className="text-center">
              <button
                type="button"
                onClick={() => navigate('/forgot-password')}
                className="text-sm text-accent hover:underline"
              >
                パスワードを忘れた場合
              </button>
            </div>
          </form>
        </div>
      </main>

      {/* フッター */}
      <footer className="bg-vrc-dark shadow-footer">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-400">
            VRC Shift Scheduler
          </p>
        </div>
      </footer>
      </div>
    </>
  );
}
