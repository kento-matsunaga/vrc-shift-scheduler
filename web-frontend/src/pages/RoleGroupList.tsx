import { useState, useEffect } from 'react';
import {
  getRoleGroups,
  createRoleGroup,
  updateRoleGroup,
  deleteRoleGroup,
  assignRolesToGroup,
  type RoleGroup,
} from '../lib/api/roleGroupApi';
import { listRoles, type Role } from '../lib/api/roleApi';
import { ApiClientError } from '../lib/apiClient';

export default function RoleGroupList() {
  const [groups, setGroups] = useState<RoleGroup[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingGroup, setEditingGroup] = useState<RoleGroup | null>(null);
  const [assigningGroup, setAssigningGroup] = useState<RoleGroup | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [groupsData, rolesData] = await Promise.all([
        getRoleGroups(),
        listRoles(),
      ]);
      setGroups(groupsData.groups || []);
      setRoles(rolesData || []);
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

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    loadData();
  };

  const handleUpdateSuccess = () => {
    setEditingGroup(null);
    loadData();
  };

  const handleAssignSuccess = () => {
    setAssigningGroup(null);
    loadData();
  };

  const handleDelete = async (groupId: string) => {
    if (!confirm('このグループを削除してもよろしいですか？')) {
      return;
    }

    try {
      await deleteRoleGroup(groupId);
      loadData();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('グループの削除に失敗しました');
      }
      console.error('Failed to delete group:', err);
    }
  };

  // ロールIDから名前を取得
  const getRoleName = (roleId: string) => {
    const role = roles.find((r) => r.role_id === roleId);
    return role?.name || 'Unknown';
  };

  // ロールの色を取得
  const getRoleColor = (roleId: string) => {
    const role = roles.find((r) => r.role_id === roleId);
    return role?.color || '#6B7280';
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
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-6">
        <div>
          <h2 className="text-xl sm:text-2xl font-bold text-gray-900">ロールグループ管理</h2>
          <p className="text-xs sm:text-sm text-gray-600 mt-1">ロールをグループ分けして管理します</p>
        </div>
        <button onClick={() => setShowCreateModal(true)} className="btn-primary text-sm sm:text-base w-full sm:w-auto">
          ＋ グループ追加
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {groups.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだロールグループがありません</p>
          <p className="text-sm text-gray-500 mb-4">ロールをグループ分けして、イベントごとに使い分けできます</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初のグループを追加
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {groups.map((group) => (
            <div key={group.group_id} className="card hover:shadow-lg transition-shadow">
              <div className="flex justify-between items-start mb-2">
                <div className="flex items-center gap-2">
                  {group.color && (
                    <div
                      className="w-4 h-4 rounded"
                      style={{ backgroundColor: group.color }}
                    ></div>
                  )}
                  <div className="text-lg font-bold text-gray-900">{group.name}</div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => setAssigningGroup(group)}
                    className="text-green-600 hover:text-green-800 text-sm"
                  >
                    ロール
                  </button>
                  <button
                    onClick={() => setEditingGroup(group)}
                    className="text-accent hover:text-accent-dark text-sm"
                  >
                    編集
                  </button>
                  <button
                    onClick={() => handleDelete(group.group_id)}
                    className="text-red-600 hover:text-red-800 text-sm"
                  >
                    削除
                  </button>
                </div>
              </div>
              {group.description && (
                <p className="text-sm text-gray-600 mt-2">{group.description}</p>
              )}
              {/* ロール一覧 */}
              <div className="mt-3">
                <div className="text-xs text-gray-500 mb-1">
                  ロール ({group.role_ids?.length || 0}個)
                </div>
                {group.role_ids && group.role_ids.length > 0 ? (
                  <div className="flex flex-wrap gap-1">
                    {group.role_ids.slice(0, 5).map((roleId) => (
                      <span
                        key={roleId}
                        className="inline-block px-2 py-0.5 text-xs rounded-full"
                        style={{
                          backgroundColor: `${getRoleColor(roleId)}20`,
                          color: getRoleColor(roleId),
                        }}
                      >
                        {getRoleName(roleId)}
                      </span>
                    ))}
                    {group.role_ids.length > 5 && (
                      <span className="inline-block px-2 py-0.5 text-xs text-gray-500">
                        +{group.role_ids.length - 5}個
                      </span>
                    )}
                  </div>
                ) : (
                  <p className="text-xs text-gray-400">ロールなし</p>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* グループ作成モーダル */}
      {showCreateModal && (
        <RoleGroupFormModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}

      {/* グループ編集モーダル */}
      {editingGroup && (
        <RoleGroupFormModal
          group={editingGroup}
          onClose={() => setEditingGroup(null)}
          onSuccess={handleUpdateSuccess}
        />
      )}

      {/* ロール割り当てモーダル */}
      {assigningGroup && (
        <AssignRolesModal
          group={assigningGroup}
          roles={roles}
          onClose={() => setAssigningGroup(null)}
          onSuccess={handleAssignSuccess}
        />
      )}
    </div>
  );
}

// グループ作成・編集モーダル
function RoleGroupFormModal({
  group,
  onClose,
  onSuccess,
}: {
  group?: RoleGroup;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [name, setName] = useState(group?.name || '');
  const [description, setDescription] = useState(group?.description || '');
  const [color, setColor] = useState(group?.color || '#3B82F6');
  const [displayOrder, setDisplayOrder] = useState(group?.display_order || 0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name) {
      setError('グループ名を入力してください');
      return;
    }

    setLoading(true);

    try {
      const input = {
        name,
        description,
        color,
        display_order: displayOrder,
      };

      if (group) {
        await updateRoleGroup(group.group_id, input);
      } else {
        await createRoleGroup(input);
      }

      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError(group ? 'グループの更新に失敗しました' : 'グループの作成に失敗しました');
      }
      console.error('Failed to save group:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <h3 className="text-xl font-bold text-gray-900 mb-4">
          {group ? 'グループを編集' : 'グループを追加'}
        </h3>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="name" className="label">
              グループ名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="input-field"
              disabled={loading}
              autoFocus
              placeholder="例: イベントA用、定例イベント用"
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
              placeholder="このグループの説明を入力してください"
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
              {loading ? '処理中...' : group ? '更新' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ロール割り当てモーダル
function AssignRolesModal({
  group,
  roles,
  onClose,
  onSuccess,
}: {
  group: RoleGroup;
  roles: Role[];
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>(group.role_ids || []);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');

  // ロール検索
  const filteredRoles = roles.filter((r) =>
    r.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const toggleRole = (roleId: string) => {
    if (selectedRoleIds.includes(roleId)) {
      setSelectedRoleIds(selectedRoleIds.filter((id) => id !== roleId));
    } else {
      setSelectedRoleIds([...selectedRoleIds, roleId]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await assignRolesToGroup(group.group_id, selectedRoleIds);
      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('ロールの割り当てに失敗しました');
      }
      console.error('Failed to assign roles:', err);
    } finally {
      setLoading(false);
    }
  };

  // 全選択・全解除
  const selectAll = () => {
    setSelectedRoleIds(filteredRoles.map((r) => r.role_id));
  };

  const deselectAll = () => {
    setSelectedRoleIds([]);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-lg w-full p-6 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center gap-2 mb-4">
          {group.color && (
            <div
              className="w-4 h-4 rounded"
              style={{ backgroundColor: group.color }}
            ></div>
          )}
          <h3 className="text-xl font-bold text-gray-900">
            {group.name} のロール
          </h3>
        </div>

        <form onSubmit={handleSubmit}>
          {/* 検索 */}
          <div className="mb-4">
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="input-field"
              placeholder="ロールを検索..."
            />
          </div>

          {/* 全選択・全解除 */}
          <div className="flex gap-2 mb-3">
            <button
              type="button"
              onClick={selectAll}
              className="text-xs text-accent hover:text-accent-dark"
            >
              全選択
            </button>
            <button
              type="button"
              onClick={deselectAll}
              className="text-xs text-gray-600 hover:text-gray-800"
            >
              全解除
            </button>
            <span className="text-xs text-gray-500 ml-auto">
              {selectedRoleIds.length}個選択中
            </span>
          </div>

          {/* ロール一覧 */}
          <div className="border border-gray-300 rounded-lg p-3 max-h-64 overflow-y-auto mb-4">
            {filteredRoles.length === 0 ? (
              <p className="text-sm text-gray-500 text-center py-4">
                ロールが見つかりません
              </p>
            ) : (
              <div className="space-y-1">
                {filteredRoles.map((role) => (
                  <label
                    key={role.role_id}
                    className="flex items-center p-2 hover:bg-gray-50 rounded cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={selectedRoleIds.includes(role.role_id)}
                      onChange={() => toggleRole(role.role_id)}
                      disabled={loading}
                      className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                    />
                    <div className="flex items-center gap-2 ml-3">
                      {role.color && (
                        <div
                          className="w-3 h-3 rounded"
                          style={{ backgroundColor: role.color }}
                        ></div>
                      )}
                      <span className="text-sm text-gray-900">{role.name}</span>
                    </div>
                  </label>
                ))}
              </div>
            )}
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
              disabled={loading}
            >
              {loading ? '処理中...' : '保存'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
