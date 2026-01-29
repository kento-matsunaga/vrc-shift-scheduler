import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { listTenants, updateTenantStatus, type TenantListItem } from '../lib/api';

export default function Tenants() {
  const [tenants, setTenants] = useState<TenantListItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState('');
  const [totalCount, setTotalCount] = useState(0);

  const fetchTenants = useCallback(async () => {
    try {
      setIsLoading(true);
      const params: { status?: string; search?: string; limit?: number } = { limit: 100 };
      if (statusFilter) params.status = statusFilter;
      if (searchQuery) params.search = searchQuery;
      const response = await listTenants(params);
      setTenants(response.data.tenants);
      setTotalCount(response.data.total_count);
    } catch (err) {
      setError('テナントの取得に失敗しました');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, [statusFilter, searchQuery]);

  useEffect(() => {
    fetchTenants();
  }, [fetchTenants]);

  const handleStatusChange = async (tenantId: string, newStatus: string) => {
    const action = newStatus === 'suspended' ? '停止' : newStatus === 'active' ? '有効化' : 'ステータス変更';
    if (!confirm(`このテナントを${action}しますか？`)) return;

    try {
      await updateTenantStatus(tenantId, { status: newStatus });
      await fetchTenants();
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
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${colors[status] || 'bg-gray-100'}`}>
        {labels[status] || status}
      </span>
    );
  };

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">テナント管理</h2>
        <p className="text-sm text-gray-500 mt-1">
          登録されているテナントの状態を管理します
        </p>
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4">
          <div className="text-sm text-red-700">{error}</div>
        </div>
      )}

      {/* フィルター */}
      <div className="mb-4 flex flex-wrap items-center gap-4">
        <input
          type="text"
          placeholder="テナント名で検索..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="block rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="block rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        >
          <option value="">すべてのステータス</option>
          <option value="active">有効</option>
          <option value="grace">猶予中</option>
          <option value="suspended">停止</option>
        </select>
        <span className="text-sm text-gray-500">
          全 {totalCount} 件
        </span>
      </div>

      {/* テーブル */}
      {isLoading ? (
        <div className="text-center py-8">読み込み中...</div>
      ) : (
        <div className="bg-white shadow overflow-hidden rounded-lg">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  テナント名
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  ステータス
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  猶予期限
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  作成日
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {tenants.map((tenant) => (
                <tr key={tenant.tenant_id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      to={`/tenants/${tenant.tenant_id}`}
                      className="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                    >
                      {tenant.tenant_name}
                    </Link>
                    <div className="text-xs text-gray-500 font-mono">
                      {tenant.tenant_id.slice(0, 12)}...
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getStatusBadge(tenant.status)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {tenant.grace_until
                      ? new Date(tenant.grace_until).toLocaleDateString('ja-JP')
                      : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(tenant.created_at).toLocaleDateString('ja-JP')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                    {tenant.status !== 'active' && (
                      <button
                        onClick={() => handleStatusChange(tenant.tenant_id, 'active')}
                        className="text-green-600 hover:text-green-900"
                      >
                        有効化
                      </button>
                    )}
                    {tenant.status === 'active' && (
                      <button
                        onClick={() => handleStatusChange(tenant.tenant_id, 'suspended')}
                        className="text-red-600 hover:text-red-900"
                      >
                        停止
                      </button>
                    )}
                    <Link
                      to={`/tenants/${tenant.tenant_id}`}
                      className="text-indigo-600 hover:text-indigo-900"
                    >
                      詳細
                    </Link>
                  </td>
                </tr>
              ))}
              {tenants.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-6 py-8 text-center text-gray-500">
                    テナントがありません
                  </td>
                </tr>
              )}
            </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
