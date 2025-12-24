import { useState, useEffect } from 'react';
import { listRoles, createRole, updateRole, deleteRole, type Role, type CreateRoleInput, type UpdateRoleInput } from '../lib/api/roleApi';
import { ApiClientError } from '../lib/apiClient';

export default function RoleList() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);

  useEffect(() => {
    loadRoles();
  }, []);

  const loadRoles = async () => {
    try {
      setLoading(true);
      const data = await listRoles();
      setRoles(data || []);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('ロール一覧の取得に失敗しました');
      }
      console.error('Failed to load roles:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    loadRoles();
  };

  const handleUpdateSuccess = () => {
    setEditingRole(null);
    loadRoles();
  };

  const handleDelete = async (roleId: string) => {
    if (!confirm('このロールを削除してもよろしいですか？')) {
      return;
    }

    try {
      await deleteRole(roleId);
      loadRoles();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('ロールの削除に失敗しました');
      }
      console.error('Failed to delete role:', err);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">ロール管理</h2>
          <p className="text-sm text-gray-600 mt-1">メンバーに付与する役割・属性を管理します</p>
        </div>
        <button onClick={() => setShowCreateModal(true)} className="btn-primary">
          ＋ ロールを追加
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {roles.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだロールがありません</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初のロールを追加
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {roles.map((role) => (
            <div key={role.role_id} className="card hover:shadow-lg transition-shadow">
              <div className="flex justify-between items-start mb-2">
                <div className="flex items-center gap-2">
                  {role.color && (
                    <div
                      className="w-4 h-4 rounded"
                      style={{ backgroundColor: role.color }}
                    ></div>
                  )}
                  <div className="text-lg font-bold text-gray-900">{role.name}</div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => setEditingRole(role)}
                    className="text-accent hover:text-accent-dark text-sm"
                  >
                    編集
                  </button>
                  <button
                    onClick={() => handleDelete(role.role_id)}
                    className="text-red-600 hover:text-red-800 text-sm"
                  >
                    削除
                  </button>
                </div>
              </div>
              {role.description && (
                <p className="text-sm text-gray-600 mt-2">{role.description}</p>
              )}
              <div className="text-xs text-gray-400 mt-2">表示順序: {role.display_order}</div>
            </div>
          ))}
        </div>
      )}

      {/* ロール作成モーダル */}
      {showCreateModal && (
        <RoleFormModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}

      {/* ロール編集モーダル */}
      {editingRole && (
        <RoleFormModal
          role={editingRole}
          onClose={() => setEditingRole(null)}
          onSuccess={handleUpdateSuccess}
        />
      )}
    </div>
  );
}

// ロール作成・編集モーダルコンポーネント
function RoleFormModal({
  role,
  onClose,
  onSuccess,
}: {
  role?: Role;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [name, setName] = useState(role?.name || '');
  const [description, setDescription] = useState(role?.description || '');
  const [color, setColor] = useState(role?.color || '#3B82F6');
  const [displayOrder, setDisplayOrder] = useState(role?.display_order || 0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name) {
      setError('ロール名を入力してください');
      return;
    }

    setLoading(true);

    try {
      const input: CreateRoleInput | UpdateRoleInput = {
        name,
        description,
        color,
        display_order: displayOrder,
      };

      if (role) {
        await updateRole(role.role_id, input);
      } else {
        await createRole(input);
      }

      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError(role ? 'ロールの更新に失敗しました' : 'ロールの作成に失敗しました');
      }
      console.error('Failed to save role:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <h3 className="text-xl font-bold text-gray-900 mb-4">
          {role ? 'ロールを編集' : 'ロールを追加'}
        </h3>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="name" className="label">
              ロール名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="input-field"
              disabled={loading}
              autoFocus
              placeholder="例: リーダー、サブリーダー、新人"
            />
          </div>

          <div className="mb-4">
            <label htmlFor="description" className="label">
              説明
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="input-field"
              disabled={loading}
              rows={3}
              placeholder="このロールの説明を入力してください"
            />
          </div>

          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label htmlFor="color" className="label">
                カラー
              </label>
              <input
                type="color"
                id="color"
                value={color}
                onChange={(e) => setColor(e.target.value)}
                className="h-10 w-full rounded border border-gray-300"
                disabled={loading}
              />
            </div>
            <div>
              <label htmlFor="displayOrder" className="label">
                表示順序
              </label>
              <input
                type="number"
                id="displayOrder"
                value={displayOrder}
                onChange={(e) => setDisplayOrder(Number(e.target.value))}
                className="input-field"
                disabled={loading}
                placeholder="0"
              />
            </div>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 btn-secondary"
              disabled={loading}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="flex-1 btn-primary"
              disabled={loading || !name}
            >
              {loading ? '処理中...' : role ? '更新' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
