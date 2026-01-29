import { useState, useEffect } from 'react';
import { getManagerPermissions, updateManagerPermissions } from '../../lib/api';
import type { ManagerPermissions } from '../../lib/api/tenantApi';
import { ApiClientError } from '../../lib/apiClient';

export function PermissionsSettings() {
  const [permissions, setPermissions] = useState<ManagerPermissions | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    loadPermissions();
  }, []);

  // REV-002: setTimeout メモリリーク修正
  useEffect(() => {
    if (success) {
      const timer = setTimeout(() => setSuccess(''), 3000);
      return () => clearTimeout(timer);
    }
  }, [success]);

  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => setError(''), 5000);
      return () => clearTimeout(timer);
    }
  }, [error]);

  const loadPermissions = async () => {
    try {
      setLoading(true);
      const data = await getManagerPermissions();
      setPermissions(data);
      setError('');
    } catch (err) {
      console.error('Failed to load manager permissions:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('マネージャー権限の読み込みに失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  const handlePermissionChange = (key: keyof ManagerPermissions, value: boolean) => {
    if (!permissions) return;
    setPermissions({ ...permissions, [key]: value });
  };

  const handleSave = async () => {
    if (!permissions) return;

    setSaving(true);
    setError('');
    setSuccess('');

    try {
      const updated = await updateManagerPermissions(permissions);
      setPermissions(updated);
      setSuccess('マネージャー権限を保存しました');
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('マネージャー権限の保存に失敗しました');
      }
      console.error('Failed to save permissions:', err);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
          <div className="space-y-3">
            <div className="h-4 bg-gray-200 rounded w-full"></div>
            <div className="h-4 bg-gray-200 rounded w-full"></div>
            <div className="h-4 bg-gray-200 rounded w-full"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
          <svg className="w-5 h-5 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
          マネージャー権限の設定
        </h2>

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-sm text-blue-800">
              マネージャーに許可する操作を設定します。オーナーはすべての操作が可能です。
            </p>
          </div>
        </div>

        {error && (
          <div role="alert" className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {success && (
          <div role="status" className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-green-800">{success}</p>
          </div>
        )}

        {permissions && (
          <>
            <div className="space-y-6">
              {/* メンバー管理 */}
              <div>
                <h4 className="text-sm font-medium text-gray-700 mb-3">メンバー管理</h4>
                <div className="space-y-2">
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_add_member}
                      onChange={(e) => handlePermissionChange('can_add_member', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">メンバーの追加</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_edit_member}
                      onChange={(e) => handlePermissionChange('can_edit_member', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">メンバーの編集</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_delete_member}
                      onChange={(e) => handlePermissionChange('can_delete_member', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">メンバーの削除</span>
                  </label>
                </div>
              </div>

              {/* イベント管理 */}
              <div>
                <h4 className="text-sm font-medium text-gray-700 mb-3">イベント管理</h4>
                <div className="space-y-2">
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_create_event}
                      onChange={(e) => handlePermissionChange('can_create_event', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">イベントの作成</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_edit_event}
                      onChange={(e) => handlePermissionChange('can_edit_event', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">イベントの編集</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_delete_event}
                      onChange={(e) => handlePermissionChange('can_delete_event', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">イベントの削除</span>
                  </label>
                </div>
              </div>

              {/* シフト管理 */}
              <div>
                <h4 className="text-sm font-medium text-gray-700 mb-3">シフト管理</h4>
                <div className="space-y-2">
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_assign_shift}
                      onChange={(e) => handlePermissionChange('can_assign_shift', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">シフトの割り当て</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_edit_shift}
                      onChange={(e) => handlePermissionChange('can_edit_shift', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">シフトの編集</span>
                  </label>
                </div>
              </div>

              {/* 出欠・スケジュール管理 */}
              <div>
                <h4 className="text-sm font-medium text-gray-700 mb-3">出欠・スケジュール管理</h4>
                <div className="space-y-2">
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_create_attendance}
                      onChange={(e) => handlePermissionChange('can_create_attendance', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">出欠確認の作成</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_create_schedule}
                      onChange={(e) => handlePermissionChange('can_create_schedule', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">日程調整の作成</span>
                  </label>
                </div>
              </div>

              {/* 組織設定 */}
              <div>
                <h4 className="text-sm font-medium text-gray-700 mb-3">組織設定</h4>
                <div className="space-y-2">
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_manage_roles}
                      onChange={(e) => handlePermissionChange('can_manage_roles', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">ロールの管理</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_manage_groups}
                      onChange={(e) => handlePermissionChange('can_manage_groups', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">グループの管理</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={permissions.can_invite_manager}
                      onChange={(e) => handlePermissionChange('can_invite_manager', e.target.checked)}
                      className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                    />
                    <span className="text-sm text-gray-700">マネージャーの招待</span>
                  </label>
                </div>
              </div>
            </div>

            <div className="mt-6">
              <button
                onClick={handleSave}
                disabled={saving}
                className="btn-primary"
              >
                {saving ? '保存中...' : '権限設定を保存'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
