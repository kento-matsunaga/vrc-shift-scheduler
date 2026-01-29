import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { resetPasswordWithToken } from '../../lib/api/authApi';
import { useDocumentTitle } from '../../hooks/useDocumentTitle';
import { SEO } from '../../components/seo';

export default function ResetPasswordWithToken() {
  const navigate = useNavigate();
  const { token } = useParams<{ token: string }>();

  useDocumentTitle('パスワードリセット');

  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  // Password strength state
  const [passwordStrength, setPasswordStrength] = useState({
    hasMinLength: false,
    hasUpper: false,
    hasLower: false,
    hasDigit: false,
  });

  // Update password strength when password changes
  useEffect(() => {
    setPasswordStrength({
      hasMinLength: newPassword.length >= 8,
      hasUpper: /[A-Z]/.test(newPassword),
      hasLower: /[a-z]/.test(newPassword),
      hasDigit: /\d/.test(newPassword),
    });
  }, [newPassword]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!token) {
      setError('トークンが無効です');
      return;
    }

    if (!newPassword) {
      setError('新しいパスワードを入力してください');
      return;
    }

    // Check password complexity
    const allPasswordRequirementsMet =
      passwordStrength.hasMinLength &&
      passwordStrength.hasUpper &&
      passwordStrength.hasLower &&
      passwordStrength.hasDigit;

    if (!allPasswordRequirementsMet) {
      setError('パスワードは8文字以上で、大文字・小文字・数字を含む必要があります');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('パスワードが一致しません');
      return;
    }

    setLoading(true);

    try {
      await resetPasswordWithToken({
        token,
        new_password: newPassword,
        confirm_new_password: confirmPassword,
      });

      setSuccess(true);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('パスワードリセットに失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <>
      <SEO noindex={true} />
      <div className="min-h-screen bg-gray-100 flex flex-col">
        <header className="bg-vrc-dark text-white shadow-soft">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
          </div>
        </header>

        <main className="flex-1 flex items-center justify-center p-4">
          <div className="bg-white rounded-card shadow-soft max-w-md w-full p-10 border border-gray-200">
            <div className="text-center">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
                <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </div>
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                無効なリンク
              </h2>
              <p className="text-sm text-gray-500 mb-6">
                このリンクは無効です。パスワードリセットをもう一度お試しください。
              </p>
              <button
                onClick={() => navigate('/forgot-password')}
                className="w-full py-3 px-4 bg-gradient-to-b from-accent-light to-accent-dark hover:from-accent-hover hover:to-accent text-white font-semibold rounded-lg border border-accent-dark shadow-inset-light transition-all"
              >
                パスワードリセットに戻る
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
      </>
    );
  }

  if (success) {
    return (
      <>
      <SEO noindex={true} />
      <div className="min-h-screen bg-gray-100 flex flex-col">
        <header className="bg-vrc-dark text-white shadow-soft">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
          </div>
        </header>

        <main className="flex-1 flex items-center justify-center p-4">
          <div className="bg-white rounded-card shadow-soft max-w-md w-full p-10 border border-gray-200">
            <div className="text-center">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
                <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                パスワードをリセットしました
              </h2>
              <p className="text-sm text-gray-500 mb-6">
                新しいパスワードでログインしてください
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
      </>
    );
  }

  return (
    <>
    <SEO noindex={true} />
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
              新しいパスワードを設定
            </h2>
            <p className="text-sm text-gray-500">
              新しいパスワードを入力してください
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-1.5">
                新しいパスワード
              </label>
              <input
                type="password"
                id="newPassword"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="8文字以上"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 shadow-inset-input focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={loading}
                autoFocus
              />
              {/* Password strength indicator */}
              {newPassword && (
                <ul className="mt-2 space-y-1 text-xs">
                  <li className={`flex items-center gap-1 ${passwordStrength.hasMinLength ? 'text-green-600' : 'text-gray-400'}`}>
                    {passwordStrength.hasMinLength ? '✓' : '○'} 8文字以上
                  </li>
                  <li className={`flex items-center gap-1 ${passwordStrength.hasUpper ? 'text-green-600' : 'text-gray-400'}`}>
                    {passwordStrength.hasUpper ? '✓' : '○'} 大文字を含む
                  </li>
                  <li className={`flex items-center gap-1 ${passwordStrength.hasLower ? 'text-green-600' : 'text-gray-400'}`}>
                    {passwordStrength.hasLower ? '✓' : '○'} 小文字を含む
                  </li>
                  <li className={`flex items-center gap-1 ${passwordStrength.hasDigit ? 'text-green-600' : 'text-gray-400'}`}>
                    {passwordStrength.hasDigit ? '✓' : '○'} 数字を含む
                  </li>
                </ul>
              )}
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1.5">
                新しいパスワード（確認）
              </label>
              <input
                type="password"
                id="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="もう一度入力"
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
              disabled={loading || !newPassword || !confirmPassword}
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  リセット中...
                </span>
              ) : (
                'パスワードをリセット'
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
    </>
  );
}
