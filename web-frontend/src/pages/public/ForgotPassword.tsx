import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { forgotPassword } from '../../lib/api/authApi';
import { useDocumentTitle } from '../../hooks/useDocumentTitle';

export default function ForgotPassword() {
  const navigate = useNavigate();

  useDocumentTitle('パスワードを忘れた場合');

  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email.trim()) {
      setError('メールアドレスを入力してください');
      return;
    }

    setLoading(true);

    try {
      await forgotPassword({ email: email.trim() });
      setSubmitted(true);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('リクエストに失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  if (submitted) {
    return (
      <div className="min-h-screen bg-gray-100 flex flex-col">
        <header className="bg-vrc-dark text-white shadow-soft">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
          </div>
        </header>

        <main className="flex-1 flex items-center justify-center p-4">
          <div className="bg-white rounded-card shadow-soft max-w-md w-full p-10 border border-gray-200">
            <div className="text-center">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 mb-4">
                <svg className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              </div>
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                メールを送信しました
              </h2>
              <p className="text-sm text-gray-500 mb-6">
                入力されたメールアドレスにパスワードリセット用のリンクを送信しました。
                メールをご確認ください。
              </p>
              <p className="text-sm text-gray-400 mb-6">
                ※ メールが届かない場合は、迷惑メールフォルダをご確認ください。
              </p>
              <button
                onClick={() => navigate('/admin/login')}
                className="w-full py-3 px-4 bg-gradient-to-b from-accent-light to-accent-dark hover:from-accent-hover hover:to-accent text-white font-semibold rounded-lg border border-accent-dark shadow-inset-light transition-all"
              >
                ログイン画面へ
              </button>
            </div>
          </div>
        </main>

        <footer className="bg-vrc-dark shadow-footer">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <p className="text-center text-sm text-gray-400">
              VRC Shift Scheduler
            </p>
          </div>
        </footer>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col">
      <header className="bg-vrc-dark text-white shadow-soft">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
        </div>
      </header>

      <main className="flex-1 flex items-center justify-center p-4">
        <div className="bg-white rounded-card shadow-soft max-w-md w-full p-10 border border-gray-200">
          <div className="text-center mb-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              パスワードを忘れた場合
            </h2>
            <p className="text-sm text-gray-500">
              登録済みのメールアドレスを入力してください。
              パスワードリセット用のリンクをお送りします。
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

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <button
              type="submit"
              className="w-full py-3 px-4 bg-gradient-to-b from-accent-light to-accent-dark hover:from-accent-hover hover:to-accent disabled:from-gray-400 disabled:to-gray-500 text-white font-semibold rounded-lg border border-accent-dark shadow-inset-light transition-all focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2"
              disabled={loading || !email.trim()}
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  送信中...
                </span>
              ) : (
                'リセットメールを送信'
              )}
            </button>

            <div className="text-center">
              <button
                type="button"
                onClick={() => navigate('/admin/login')}
                className="text-sm text-accent hover:underline"
              >
                ログイン画面に戻る
              </button>
            </div>
          </form>
        </div>
      </main>

      <footer className="bg-vrc-dark shadow-footer">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-400">
            VRC Shift Scheduler
          </p>
        </div>
      </footer>
    </div>
  );
}
