import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { claimLicense } from '../../lib/api/billingApi';
import { useDocumentTitle } from '../../hooks/useDocumentTitle';

export default function LicenseClaim() {
  const navigate = useNavigate();

  useDocumentTitle('ライセンス登録');

  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    display_name: '',
    tenant_name: '',
    license_key: '',
  });
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const formatLicenseKey = (value: string) => {
    // Remove any existing hyphens and non-alphanumeric characters
    const cleaned = value.replace(/[^A-Fa-f0-9]/g, '').toUpperCase();
    // Add hyphens every 4 characters
    const parts = [];
    for (let i = 0; i < cleaned.length && i < 16; i += 4) {
      parts.push(cleaned.slice(i, i + 4));
    }
    return parts.join('-');
  };

  const handleLicenseKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatLicenseKey(e.target.value);
    setFormData((prev) => ({ ...prev, license_key: formatted }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validation
    if (formData.password !== formData.confirmPassword) {
      setError('パスワードが一致しません');
      return;
    }

    if (formData.password.length < 8) {
      setError('パスワードは8文字以上で入力してください');
      return;
    }

    // Check password complexity
    if (!/[A-Z]/.test(formData.password)) {
      setError('パスワードには大文字を1文字以上含めてください');
      return;
    }
    if (!/[a-z]/.test(formData.password)) {
      setError('パスワードには小文字を1文字以上含めてください');
      return;
    }
    if (!/[0-9]/.test(formData.password)) {
      setError('パスワードには数字を1文字以上含めてください');
      return;
    }

    if (!formData.license_key.match(/^[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{4}$/)) {
      setError('ライセンスキーの形式が正しくありません（形式: XXXX-XXXX-XXXX-XXXX）');
      return;
    }

    setIsLoading(true);

    try {
      await claimLicense({
        email: formData.email,
        password: formData.password,
        display_name: formData.display_name,
        tenant_name: formData.tenant_name,
        license_key: formData.license_key,
      });

      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'ライセンスの登録に失敗しました');
    } finally {
      setIsLoading(false);
    }
  };

  if (success) {
    return (
      <div className="min-h-screen bg-gray-100 flex flex-col">
        {/* ヘッダー */}
        <header className="bg-vrc-dark text-white shadow">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
          </div>
        </header>

        {/* メインコンテンツ */}
        <main className="flex-1 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-md max-w-md w-full p-8">
            <div className="text-center mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">登録完了！</h2>
              <p className="text-sm text-gray-500">
                アカウントが正常に作成されました。メールアドレスとパスワードでログインできます。
              </p>
            </div>
            <button
              onClick={() => navigate('/login')}
              className="w-full py-3 px-4 bg-accent hover:bg-accent-dark text-white font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2"
            >
              ログインへ
            </button>
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

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col">
      {/* ヘッダー */}
      <header className="bg-vrc-dark text-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
        </div>
      </header>

      {/* メインコンテンツ */}
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-md max-w-md w-full p-8">
          <div className="text-center mb-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              ライセンスキーで登録
            </h2>
            <p className="text-sm text-gray-500">
              BOOTHで購入したライセンスキーを入力してアカウントを作成してください
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label htmlFor="license_key" className="block text-sm font-medium text-gray-700 mb-1.5">
                ライセンスキー
              </label>
              <input
                id="license_key"
                name="license_key"
                type="text"
                required
                value={formData.license_key}
                onChange={handleLicenseKeyChange}
                placeholder="XXXX-XXXX-XXXX-XXXX"
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition font-mono"
                disabled={isLoading}
              />
            </div>

            <div>
              <label htmlFor="tenant_name" className="block text-sm font-medium text-gray-700 mb-1.5">
                組織名
              </label>
              <input
                id="tenant_name"
                name="tenant_name"
                type="text"
                required
                value={formData.tenant_name}
                onChange={handleChange}
                placeholder="VRCイベント名"
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={isLoading}
              />
            </div>

            <div>
              <label htmlFor="display_name" className="block text-sm font-medium text-gray-700 mb-1.5">
                表示名
              </label>
              <input
                id="display_name"
                name="display_name"
                type="text"
                required
                value={formData.display_name}
                onChange={handleChange}
                placeholder="管理者"
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={isLoading}
              />
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1.5">
                メールアドレス
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                value={formData.email}
                onChange={handleChange}
                placeholder="admin@example.com"
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={isLoading}
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1.5">
                パスワード
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="new-password"
                required
                value={formData.password}
                onChange={handleChange}
                placeholder="••••••••"
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={isLoading}
              />
              <p className="mt-1 text-xs text-gray-500">8文字以上、大文字・小文字・数字を含む</p>
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1.5">
                パスワード（確認）
              </label>
              <input
                id="confirmPassword"
                name="confirmPassword"
                type="password"
                autoComplete="new-password"
                required
                value={formData.confirmPassword}
                onChange={handleChange}
                placeholder="••••••••"
                className="w-full px-4 py-3 border border-gray-300 rounded-md text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition"
                disabled={isLoading}
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <button
              type="submit"
              className="w-full py-3 px-4 bg-accent hover:bg-accent-dark disabled:bg-accent/70 text-white font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2"
              disabled={isLoading}
            >
              {isLoading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  登録中...
                </span>
              ) : (
                '登録'
              )}
            </button>

            <div className="text-center">
              <a href="/login" className="text-sm text-accent hover:text-accent">
                アカウントをお持ちですか？ログイン
              </a>
            </div>
          </form>
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
