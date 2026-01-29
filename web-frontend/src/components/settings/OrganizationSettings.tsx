import { useState, useEffect } from 'react';
import { getCurrentTenant, updateTenant } from '../../lib/api';
import type { Tenant } from '../../lib/api/tenantApi';
import { ApiClientError } from '../../lib/apiClient';

export function OrganizationSettings() {
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [editingTenantName, setEditingTenantName] = useState(false);
  const [tenantName, setTenantName] = useState('');
  const [savingTenant, setSavingTenant] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    loadTenant();
  }, []);

  const loadTenant = async () => {
    try {
      setLoading(true);
      const tenantData = await getCurrentTenant();
      setTenant(tenantData);
      setTenantName(tenantData.tenant_name);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('組織情報の取得に失敗しました');
      }
      console.error('Failed to load tenant:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveTenantName = async () => {
    if (!tenantName.trim()) {
      setError('組織名を入力してください');
      return;
    }

    setSavingTenant(true);
    setError('');

    try {
      const updated = await updateTenant({ tenant_name: tenantName.trim() });
      setTenant(updated);
      setEditingTenantName(false);
      setSuccess('組織名を更新しました');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('組織名の更新に失敗しました');
      }
      console.error('Failed to update tenant:', err);
    } finally {
      setSavingTenant(false);
    }
  };

  const handleCancelTenantEdit = () => {
    setTenantName(tenant?.tenant_name || '');
    setEditingTenantName(false);
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
          <svg className="w-5 h-5 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
          組織情報
        </h2>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {success && (
          <div className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-green-800">{success}</p>
          </div>
        )}

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">組織名</label>
            {editingTenantName ? (
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={tenantName}
                  onChange={(e) => setTenantName(e.target.value)}
                  className="input-field flex-1"
                  disabled={savingTenant}
                  autoFocus
                />
                <button
                  onClick={handleSaveTenantName}
                  disabled={savingTenant}
                  className="btn-primary"
                >
                  {savingTenant ? '保存中...' : '保存'}
                </button>
                <button
                  onClick={handleCancelTenantEdit}
                  disabled={savingTenant}
                  className="btn-secondary"
                >
                  キャンセル
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <span className="text-gray-900">{tenant?.tenant_name}</span>
                <button
                  onClick={() => setEditingTenantName(true)}
                  className="p-1 text-gray-400 hover:text-accent transition-colors"
                  title="組織名を編集"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
              </div>
            )}
          </div>
          {tenant && (
            <div className="text-sm text-gray-500">
              タイムゾーン: {tenant.timezone}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
