import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getTenantDetail, updateTenantStatus, type TenantDetail as TenantDetailType } from '../lib/api';

export default function TenantDetail() {
  const { tenantId } = useParams<{ tenantId: string }>();
  const [tenant, setTenant] = useState<TenantDetailType | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      active: 'bg-green-100 text-green-800',
      grace: 'bg-yellow-100 text-yellow-800',
      suspended: 'bg-red-100 text-red-800',
    };
    const labels: Record<string, string> = {
      active: '有効',
      grace: '猶予中',
      suspended: '停止',
    };
    return (
      <span className={`px-3 py-1 text-sm font-medium rounded-full ${colors[status] || 'bg-gray-100'}`}>
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

      {/* エンタイトルメント */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">エンタイトルメント</h3>
        {tenant.entitlements && tenant.entitlements.length > 0 ? (
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
        ) : (
          <p className="text-sm text-gray-500">エンタイトルメントがありません</p>
        )}
      </div>

      {/* 管理者 */}
      <div className="bg-white shadow rounded-lg p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">管理者</h3>
        {tenant.admins && tenant.admins.length > 0 ? (
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
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <p className="text-sm text-gray-500">管理者がいません</p>
        )}
      </div>
    </div>
  );
}
