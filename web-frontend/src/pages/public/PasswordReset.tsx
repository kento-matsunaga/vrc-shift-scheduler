import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { checkPasswordResetStatus, resetPassword } from '../../lib/api/authApi';

export default function PasswordReset() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [email, setEmail] = useState(searchParams.get('email') || '');
  const [licenseKey, setLicenseKey] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const [loading, setLoading] = useState(false);
  const [checkingStatus, setCheckingStatus] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const [resetAllowed, setResetAllowed] = useState<boolean | null>(null);
  const [expiresAt, setExpiresAt] = useState<string | null>(null);

  // Check reset status when email changes
  useEffect(() => {
    if (!email.trim()) {
      setResetAllowed(null);
      return;
    }

    const timer = setTimeout(async () => {
      setCheckingStatus(true);
      try {
        const status = await checkPasswordResetStatus(email.trim());
        setResetAllowed(status.allowed);
        setExpiresAt(status.expires_at || null);
      } catch (err) {
        console.error('Failed to check status:', err);
        setResetAllowed(false);
      } finally {
        setCheckingStatus(false);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [email]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email.trim()) {
      setError('メールアドレスを入力してください');
      return;
    }

    if (!licenseKey.trim()) {
      setError('ライセンスキーを入力してください');
      return;
    }

    if (!newPassword) {
      setError('新しいパスワードを入力してください');
      return;
    }

    if (newPassword.length < 8) {
      setError('パスワードは8文字以上で入力してください');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('パスワードが一致しません');
      return;
    }

    setLoading(true);

    try {
      await resetPassword({
        email: email.trim(),
        license_key: licenseKey.trim(),
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

  if (success) {
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
              パスワードリセット
            </h2>
            <p className="text-sm text-gray-500">
              オーナーから許可を受けた後、ライセンスキーで本人確認を行いパスワードをリセットします
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
              {checkingStatus && (
                <p className="mt-1 text-sm text-gray-500">確認中...</p>
              )}
              {!checkingStatus && resetAllowed === false && email.trim() && (
                <p className="mt-1 text-sm text-red-600">
                  パスワードリセットが許可されていません。オーナーに依頼してください。
                </p>
              )}
              {!checkingStatus && resetAllowed === true && (
                <p className="mt-1 text-sm text-green-600">
                  パスワードリセットが許可されています
                  {expiresAt && ` (${new Date(expiresAt).toLocaleString()}まで)`}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="licenseKey" className="block text-sm font-medium text-gray-700 mb-1.5">
                ライセンスキー
              </label>
              <input
                type="text"
                id="licenseKey"
                value={licenseKey}
                onChange={(e) => setLicenseKey(e.target.value)}
                placeholder="XXXX-XXXX-XXXX-XXXX"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 shadow-inset-input focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition font-mono"
                disabled={loading}
              />
              <p className="mt-1 text-sm text-gray-500">
                登録時に使用したライセンスキーを入力してください
              </p>
            </div>

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
              />
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
              disabled={loading || !resetAllowed || !email.trim() || !licenseKey.trim() || !newPassword || !confirmPassword}
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
  );
}
