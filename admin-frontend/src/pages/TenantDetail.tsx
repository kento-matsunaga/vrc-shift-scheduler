import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getTenantDetail, updateTenantStatus, allowPasswordReset, type TenantDetail as TenantDetailType } from '../lib/api';

export default function TenantDetail() {
  const { tenantId } = useParams<{ tenantId: string }>();
  const [tenant, setTenant] = useState<TenantDetailType | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [passwordResetSuccess, setPasswordResetSuccess] = useState<string | null>(null);
  const [allowingResetFor, setAllowingResetFor] = useState<string | null>(null);

  useEffect(() => {
    if (!tenantId) return;

    async function fetchTenant() {
      try {
        setIsLoading(true);
        const response = await getTenantDetail(tenantId!);
        setTenant(response.data);
      } catch (err) {
        setError('テナントの取得に失敗しました');
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    }

    fetchTenant();
  }, [tenantId]);

  const handleStatusChange = async (newStatus: string) => {
    if (!tenantId || !tenant) return;
    const action = newStatus === 'suspended' ? '停止' : newStatus === 'active' ? '有効化' : 'ステータス変更';
    if (!confirm(`このテナントを${action}しますか？`)) return;

    try {
      await updateTenantStatus(tenantId, { status: newStatus });
      const response = await getTenantDetail(tenantId);
      setTenant(response.data);
    } catch (err) {
      setError('ステータスの変更に失敗しました');
      console.error(err);
    }
  };

  const handleAllowPasswordReset = async (adminId: string, displayName: string) => {
    if (!confirm(`「${displayName}」のパスワードリセットを許可しますか？\n（24時間有効）`)) return;

    setAllowingResetFor(adminId);
    setError(null);
    setPasswordResetSuccess(null);

    try {
      const response = await allowPasswordReset(adminId);
      setPasswordResetSuccess(
        `「${displayName}」のパスワードリセットを許可しました。有効期限: ${new Date(response.data.expires_at).toLocaleString('ja-JP')}`
      );
      setTimeout(() => setPasswordResetSuccess(null), 10000);
    } catch (err) {
      setError('パスワードリセットの許可に失敗しました');
      console.error(err);
    } finally {
      setAllowingResetFor(null);
    }
  };

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      active: 'bg-green-100 text-green-800',
      grace: 'bg-yellow-100 text-yellow-800',
      suspended: 'bg-red-100 text-red-800',
      pending_payment: 'bg-blue-100 text-blue-800',
    };
    const labels: Record<string, string> = {
      active: '有効',
      grace: '猶予中',
      suspended: '停止',
      pending_payment: '決済待ち',
    };
    return (
      <span className={`px-3 py-1 text-sm font-medium rounded-full ${colors[status] || 'bg-gray-100'}`}>
        {labels[status] || status}
      </span>
    );
  };

  const getSubscriptionStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      active: 'bg-green-100 text-green-800',
      trialing: 'bg-blue-100 text-blue-800',
      past_due: 'bg-yellow-100 text-yellow-800',
      canceled: 'bg-gray-100 text-gray-800',
      unpaid: 'bg-red-100 text-red-800',
      incomplete: 'bg-orange-100 text-orange-800',
    };
    const labels: Record<string, string> = {
      active: '有効',
      trialing: 'トライアル',
      past_due: '支払い遅延',
      canceled: 'キャンセル済',
      unpaid: '未払い',
      incomplete: '不完全',
    };
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${colors[status] || 'bg-gray-100'}`}>
        {labels[status] || status}
      </span>
    );
  };

  if (isLoading) {
    return <div className="text-center py-8">読み込み中...</div>;
  }

  if (error || !tenant) {
    return (
      <div className="text-center py-8">
        <p className="text-red-600">{error || 'テナントが見つかりません'}</p>
        <Link to="/tenants" className="text-indigo-600 hover:text-indigo-500 mt-4 inline-block">
          ← テナント一覧に戻る
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <Link to="/tenants" className="text-sm text-indigo-600 hover:text-indigo-500">
          ← テナント一覧に戻る
        </Link>
      </div>

      {/* 基本情報 */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="flex justify-between items-start">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">{tenant.tenant_name}</h2>
            <p className="text-sm text-gray-500 font-mono mt-1">{tenant.tenant_id}</p>
          </div>
          <div className="flex items-center space-x-3">
            {getStatusBadge(tenant.status)}
            {tenant.status !== 'active' && (
              <button
                onClick={() => handleStatusChange('active')}
                className="px-3 py-1 text-sm font-medium text-green-700 bg-green-100 rounded-md hover:bg-green-200"
              >
                有効化
              </button>
            )}
            {tenant.status === 'active' && (
              <button
                onClick={() => handleStatusChange('suspended')}
                className="px-3 py-1 text-sm font-medium text-red-700 bg-red-100 rounded-md hover:bg-red-200"
              >
                停止
              </button>
            )}
          </div>
        </div>

        <dl className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-sm font-medium text-gray-500">作成日</dt>
            <dd className="text-sm text-gray-900">
              {new Date(tenant.created_at).toLocaleString('ja-JP')}
            </dd>
          </div>
          {tenant.grace_until && (
            <div>
              <dt className="text-sm font-medium text-gray-500">猶予期限</dt>
              <dd className="text-sm text-gray-900">
                {new Date(tenant.grace_until).toLocaleString('ja-JP')}
              </dd>
            </div>
          )}
        </dl>
      </div>

      {/* サブスクリプション */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Stripeサブスクリプション</h3>
        {tenant.subscription ? (
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div>
              <dt className="text-sm font-medium text-gray-500">ステータス</dt>
              <dd className="text-sm text-gray-900 flex items-center space-x-2">
                {getSubscriptionStatusBadge(tenant.subscription.status)}
                {tenant.subscription.cancel_at_period_end && (
                  <span className="px-2 py-1 text-xs font-medium rounded-full bg-orange-100 text-orange-800">
                    キャンセル予約中
                  </span>
                )}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Stripe Customer ID</dt>
              <dd className="text-sm text-gray-900 font-mono">{tenant.subscription.stripe_customer_id}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Stripe Subscription ID</dt>
              <dd className="text-sm text-gray-900 font-mono">{tenant.subscription.stripe_subscription_id}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">現在の請求期間終了日</dt>
              <dd className="text-sm text-gray-900">
                {tenant.subscription.current_period_end
                  ? new Date(tenant.subscription.current_period_end).toLocaleString('ja-JP')
                  : '未設定'}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">作成日</dt>
              <dd className="text-sm text-gray-900">
                {new Date(tenant.subscription.created_at).toLocaleString('ja-JP')}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">更新日</dt>
              <dd className="text-sm text-gray-900">
                {new Date(tenant.subscription.updated_at).toLocaleString('ja-JP')}
              </dd>
            </div>
            {tenant.subscription.cancel_at_period_end && tenant.subscription.cancel_at && (
              <div>
                <dt className="text-sm font-medium text-gray-500">キャンセル予定日</dt>
                <dd className="text-sm text-orange-600 font-medium">
                  {new Date(tenant.subscription.cancel_at).toLocaleString('ja-JP')}
                </dd>
              </div>
            )}
          </dl>
        ) : (
          <p className="text-sm text-gray-500">サブスクリプションがありません（ライセンスキー利用の可能性があります）</p>
        )}
      </div>

      {/* エンタイトルメント */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">エンタイトルメント</h3>
        {tenant.entitlements && tenant.entitlements.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
            <thead>
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  プランコード
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  開始日
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  有効期限
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  状態
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {tenant.entitlements.map((e) => (
                <tr key={e.entitlement_id}>
                  <td className="px-4 py-2 text-sm font-medium text-gray-900">
                    {e.plan_code}
                  </td>
                  <td className="px-4 py-2 text-sm text-gray-500">
                    {new Date(e.started_at).toLocaleDateString('ja-JP')}
                  </td>
                  <td className="px-4 py-2 text-sm text-gray-500">
                    {e.expires_at ? new Date(e.expires_at).toLocaleDateString('ja-JP') : '無期限'}
                  </td>
                  <td className="px-4 py-2 text-sm">
                    {e.revoked_at ? (
                      <span className="text-red-600">失効済</span>
                    ) : (
                      <span className="text-green-600">有効</span>
                    )}
                  </td>
                </tr>
              ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-gray-500">エンタイトルメントがありません</p>
        )}
      </div>

      {/* 管理者 */}
      <div className="bg-white shadow rounded-lg p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">管理者</h3>

        {passwordResetSuccess && (
          <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-md">
            <p className="text-sm text-green-800">{passwordResetSuccess}</p>
          </div>
        )}

        {tenant.admins && tenant.admins.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
            <thead>
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  表示名
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  メールアドレス
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  ロール
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {tenant.admins.map((a) => (
                <tr key={a.admin_id}>
                  <td className="px-4 py-2 text-sm font-medium text-gray-900">
                    {a.display_name}
                  </td>
                  <td className="px-4 py-2 text-sm text-gray-500">
                    {a.email}
                  </td>
                  <td className="px-4 py-2 text-sm text-gray-500">
                    {a.role === 'owner' ? 'オーナー' : 'マネージャー'}
                  </td>
                  <td className="px-4 py-2 text-sm">
                    <button
                      onClick={() => handleAllowPasswordReset(a.admin_id, a.display_name)}
                      disabled={allowingResetFor === a.admin_id}
                      className="px-3 py-1 text-sm font-medium text-indigo-700 bg-indigo-100 rounded-md hover:bg-indigo-200 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {allowingResetFor === a.admin_id ? '許可中...' : 'PWリセット許可'}
                    </button>
                  </td>
                </tr>
              ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-gray-500">管理者がいません</p>
        )}
      </div>
    </div>
  );
}
