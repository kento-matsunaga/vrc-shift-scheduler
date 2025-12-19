import { useState, useEffect } from 'react';
import { getEvents, deleteEvent, getCurrentTenant, updateTenant, changePassword } from '../lib/api';
import type { Event } from '../types/api';
import type { Tenant } from '../lib/api/tenantApi';
import { ApiClientError } from '../lib/apiClient';

export default function Settings() {
  // Tenant state
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [editingTenantName, setEditingTenantName] = useState(false);
  const [tenantName, setTenantName] = useState('');
  const [savingTenant, setSavingTenant] = useState(false);

  // Password change state
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [changingPassword, setChangingPassword] = useState(false);
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');

  // Events state
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<Event | null>(null);
  const [confirmText, setConfirmText] = useState('');
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [tenantData, eventsData] = await Promise.all([
        getCurrentTenant(),
        getEvents({ is_active: true }),
      ]);
      setTenant(tenantData);
      setTenantName(tenantData.tenant_name);
      setEvents(eventsData.events);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('データの取得に失敗しました');
      }
      console.error('Failed to load data:', err);
    } finally {
      setLoading(false);
    }
  };

  // Tenant handlers
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

  // Password change handlers
  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');
    setPasswordSuccess('');

    // Validation
    if (!currentPassword) {
      setPasswordError('現在のパスワードを入力してください');
      return;
    }
    if (!newPassword) {
      setPasswordError('新しいパスワードを入力してください');
      return;
    }
    if (newPassword.length < 8) {
      setPasswordError('新しいパスワードは8文字以上で入力してください');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('新しいパスワードと確認用パスワードが一致しません');
      return;
    }
    if (currentPassword === newPassword) {
      setPasswordError('新しいパスワードは現在のパスワードと異なるものを入力してください');
      return;
    }

    setChangingPassword(true);

    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
        confirm_new_password: confirmPassword,
      });
      setPasswordSuccess('パスワードを変更しました');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      if (err instanceof ApiClientError) {
        if (err.message.includes('incorrect') || err.message.includes('Unauthorized')) {
          setPasswordError('現在のパスワードが正しくありません');
        } else {
          setPasswordError(err.getUserMessage());
        }
      } else {
        setPasswordError('パスワードの変更に失敗しました');
      }
      console.error('Failed to change password:', err);
    } finally {
      setChangingPassword(false);
    }
  };

  // Event deletion handlers
  const handleDeleteClick = (event: Event) => {
    setDeleteTarget(event);
    setConfirmText('');
    setError('');
  };

  const handleCancelDelete = () => {
    setDeleteTarget(null);
    setConfirmText('');
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    if (confirmText !== deleteTarget.event_name) {
      setError('イベント名が一致しません');
      return;
    }

    setDeleting(true);
    setError('');

    try {
      await deleteEvent(deleteTarget.event_id);
      setEvents(events.filter(e => e.event_id !== deleteTarget.event_id));
      setDeleteTarget(null);
      setConfirmText('');
      setSuccess(`「${deleteTarget.event_name}」を削除しました`);
      setTimeout(() => setSuccess(''), 5000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('イベントの削除に失敗しました');
      }
      console.error('Failed to delete event:', err);
    } finally {
      setDeleting(false);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <h2 className="text-2xl font-bold text-gray-900 mb-6">基本設定</h2>

      {error && !deleteTarget && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {success && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-green-800">{success}</p>
        </div>
      )}

      {/* テナント情報セクション */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
          組織情報
        </h3>

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
                  className="p-1 text-gray-400 hover:text-blue-600 transition-colors"
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

      {/* パスワード変更セクション */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
          </svg>
          パスワード変更
        </h3>

        {passwordError && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{passwordError}</p>
          </div>
        )}

        {passwordSuccess && (
          <div className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-green-800">{passwordSuccess}</p>
          </div>
        )}

        <form onSubmit={handleChangePassword} className="space-y-4">
          <div>
            <label htmlFor="currentPassword" className="block text-sm font-medium text-gray-700 mb-1">
              現在のパスワード
            </label>
            <input
              type="password"
              id="currentPassword"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="input-field"
              disabled={changingPassword}
              autoComplete="current-password"
            />
          </div>

          <div>
            <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-1">
              新しいパスワード
            </label>
            <input
              type="password"
              id="newPassword"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="input-field"
              disabled={changingPassword}
              autoComplete="new-password"
            />
            <p className="text-xs text-gray-500 mt-1">8文字以上で入力してください</p>
          </div>

          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
              新しいパスワード（確認）
            </label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="input-field"
              disabled={changingPassword}
              autoComplete="new-password"
            />
          </div>

          <button
            type="submit"
            disabled={changingPassword || !currentPassword || !newPassword || !confirmPassword}
            className="btn-primary"
          >
            {changingPassword ? 'パスワード変更中...' : 'パスワードを変更'}
          </button>
        </form>
      </div>

      {/* イベント削除セクション */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          イベントの削除
        </h3>

        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-4">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div>
              <p className="text-sm font-medium text-amber-800">注意</p>
              <p className="text-sm text-amber-700 mt-1">
                イベントを削除すると、関連する営業日、シフト枠、シフト割り当てなども削除されます。
                この操作は取り消せません。
              </p>
            </div>
          </div>
        </div>

        {events.length === 0 ? (
          <p className="text-gray-600">削除できるイベントがありません</p>
        ) : (
          <div className="space-y-3">
            {events.map((event) => (
              <div
                key={event.event_id}
                className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border border-gray-200"
              >
                <div>
                  <p className="font-medium text-gray-900">{event.event_name}</p>
                  <p className="text-sm text-gray-500">
                    {event.event_type === 'normal' ? '通常イベント' : '特別イベント'}
                    {event.description && ` - ${event.description}`}
                  </p>
                </div>
                <button
                  onClick={() => handleDeleteClick(event)}
                  className="px-3 py-1.5 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded transition-colors"
                >
                  削除
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 削除確認モーダル */}
      {deleteTarget && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center flex-shrink-0">
                <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </div>
              <h3 className="text-xl font-bold text-gray-900">イベントの削除</h3>
            </div>

            <div className="mb-4">
              <p className="text-gray-700 mb-2">
                「<span className="font-bold text-red-600">{deleteTarget.event_name}</span>」を削除しようとしています。
              </p>
              <p className="text-sm text-gray-600">
                このイベントに関連するすべてのデータ（営業日、シフト枠、シフト割り当てなど）も削除されます。
              </p>
            </div>

            <div className="bg-gray-50 rounded-lg p-4 mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                確認のため、イベント名「<span className="font-mono text-red-600">{deleteTarget.event_name}</span>」を入力してください
              </label>
              <input
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder="イベント名を入力"
                className="input-field"
                disabled={deleting}
                autoFocus
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
                <p className="text-sm text-red-800">{error}</p>
              </div>
            )}

            <div className="flex space-x-3">
              <button
                type="button"
                onClick={handleCancelDelete}
                className="flex-1 btn-secondary"
                disabled={deleting}
              >
                キャンセル
              </button>
              <button
                type="button"
                onClick={handleConfirmDelete}
                disabled={deleting || confirmText !== deleteTarget.event_name}
                className="flex-1 bg-red-600 hover:bg-red-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                {deleting ? '削除中...' : '削除する'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
