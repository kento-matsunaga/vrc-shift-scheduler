import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { createMember } from '../lib/api';
import { ApiClientError } from '../lib/apiClient';

export default function Login() {
  const [displayName, setDisplayName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  // 招待URLから tenant_id を取得（例: ?tenant_id=xxxxx）
  useEffect(() => {
    const tenantIdFromUrl = searchParams.get('tenant_id');
    if (tenantIdFromUrl) {
      localStorage.setItem('tenant_id', tenantIdFromUrl);
    }
  }, [searchParams]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!displayName.trim()) {
      setError('表示名を入力してください');
      return;
    }

    // tenant_id の確認
    const tenantId = localStorage.getItem('tenant_id') || import.meta.env.VITE_TENANT_ID;
    if (!tenantId) {
      setError('テナントIDが設定されていません。招待URLからアクセスしてください。');
      return;
    }

    setLoading(true);

    try {
      // メンバー作成 API を叩く
      const member = await createMember({
        display_name: displayName.trim(),
      });

      // localStorage に保存
      localStorage.setItem('member_id', member.member_id);
      localStorage.setItem('member_name', member.display_name);

      // イベント一覧に遷移
      navigate('/events');
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('ログインに失敗しました。もう一度お試しください。');
      }
      console.error('Login error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="card max-w-md w-full">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            VRC Shift Scheduler
          </h1>
          <p className="text-sm text-gray-600">
            α 版テスト - シフト管理システム
          </p>
        </div>

        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-yellow-800">
            ⚠️ <strong>テスト版について</strong>
          </p>
          <ul className="text-xs text-yellow-700 mt-2 space-y-1 list-disc list-inside">
            <li>これは α 版のテストです</li>
            <li>データは予告なく消える可能性があります</li>
            <li>バグが見つかった場合は報告をお願いします</li>
          </ul>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="mb-6">
            <label htmlFor="displayName" className="label">
              表示名
            </label>
            <input
              type="text"
              id="displayName"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="例: テスト太郎"
              className="input-field"
              disabled={loading}
              autoFocus
            />
            <p className="text-xs text-gray-500 mt-1">
              他のメンバーに表示される名前を入力してください
            </p>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <button
            type="submit"
            className="w-full btn-primary"
            disabled={loading || !displayName.trim()}
          >
            {loading ? '登録中...' : 'ログイン'}
          </button>
        </form>

        <div className="mt-6 pt-6 border-t border-gray-200">
          <p className="text-xs text-gray-500 text-center">
            初めての方は表示名を入力するだけで参加できます
          </p>
        </div>
      </div>
    </div>
  );
}

