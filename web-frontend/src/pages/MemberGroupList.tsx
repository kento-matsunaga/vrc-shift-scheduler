import { useState, useEffect } from 'react';
import {
  getMemberGroups,
  createMemberGroup,
  updateMemberGroup,
  deleteMemberGroup,
  assignMembersToGroup,
  type MemberGroup,
} from '../lib/api/memberGroupApi';
import { getMembers } from '../lib/api/memberApi';
import type { Member } from '../types/api';
import { ApiClientError } from '../lib/apiClient';
import { SEO } from '../components/seo';

// プリセットカラーパレット
const PRESET_COLORS = {
  basic: [
    { name: '赤', color: '#EF4444' },
    { name: 'オレンジ', color: '#F97316' },
    { name: '黄', color: '#EAB308' },
    { name: '緑', color: '#22C55E' },
    { name: '青', color: '#3B82F6' },
    { name: '藍', color: '#6366F1' },
    { name: '紫', color: '#A855F7' },
    { name: 'ピンク', color: '#EC4899' },
  ],
  pastel: [
    { name: 'ピンク', color: '#FCA5A5' },
    { name: 'オレンジ', color: '#FDBA74' },
    { name: 'イエロー', color: '#FDE047' },
    { name: 'ライム', color: '#BEF264' },
    { name: 'スカイ', color: '#7DD3FC' },
    { name: 'ラベンダー', color: '#C4B5FD' },
    { name: 'ローズ', color: '#F9A8D4' },
    { name: 'グレー', color: '#D1D5DB' },
  ],
};

export default function MemberGroupList() {
  const [groups, setGroups] = useState<MemberGroup[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingGroup, setEditingGroup] = useState<MemberGroup | null>(null);
  const [assigningGroup, setAssigningGroup] = useState<MemberGroup | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [groupsData, membersData] = await Promise.all([
        getMemberGroups(),
        getMembers({ is_active: true }),
      ]);
      setGroups(groupsData.groups || []);
      setMembers(membersData.members || []);
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
      await deleteMemberGroup(groupId);
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

  // メンバーIDから名前を取得
  const getMemberName = (memberId: string) => {
    const member = members.find((m) => m.member_id === memberId);
    return member?.display_name || 'Unknown';
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
      <SEO noindex={true} />
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-6">
        <div>
          <h2 className="text-xl sm:text-2xl font-bold text-gray-900">グループ管理</h2>
          <p className="text-xs sm:text-sm text-gray-600 mt-1">メンバーをグループ分けして管理します</p>
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
          <p className="text-gray-600 mb-4">まだグループがありません</p>
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
                    メンバー
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
              {/* メンバー一覧 */}
              <div className="mt-3">
                <div className="text-xs text-gray-500 mb-1">
                  メンバー ({group.member_ids?.length || 0}人)
                </div>
                {group.member_ids && group.member_ids.length > 0 ? (
                  <div className="flex flex-wrap gap-1">
                    {group.member_ids.slice(0, 5).map((memberId) => (
                      <span
                        key={memberId}
                        className="inline-block px-2 py-0.5 text-xs rounded-full"
                        style={{
                          backgroundColor: group.color ? `${group.color}20` : '#E5E7EB',
                          color: group.color || '#374151',
                        }}
                      >
                        {getMemberName(memberId)}
                      </span>
                    ))}
                    {group.member_ids.length > 5 && (
                      <span className="inline-block px-2 py-0.5 text-xs text-gray-500">
                        +{group.member_ids.length - 5}人
                      </span>
                    )}
                  </div>
                ) : (
                  <p className="text-xs text-gray-400">メンバーなし</p>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* グループ作成モーダル */}
      {showCreateModal && (
        <GroupFormModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}

      {/* グループ編集モーダル */}
      {editingGroup && (
        <GroupFormModal
          group={editingGroup}
          onClose={() => setEditingGroup(null)}
          onSuccess={handleUpdateSuccess}
        />
      )}

      {/* メンバー割り当てモーダル */}
      {assigningGroup && (
        <AssignMembersModal
          group={assigningGroup}
          members={members}
          onClose={() => setAssigningGroup(null)}
          onSuccess={handleAssignSuccess}
        />
      )}
    </div>
  );
}

// グループ作成・編集モーダル
function GroupFormModal({
  group,
  onClose,
  onSuccess,
}: {
  group?: MemberGroup;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [name, setName] = useState(group?.name || '');
  const [description, setDescription] = useState(group?.description || '');
  const [color, setColor] = useState(group?.color || '#10B981');
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

    // HEXカラーコードのバリデーション
    if (color && !/^#[0-9A-Fa-f]{6}$/.test(color)) {
      setError('カラーコードは #RRGGBB 形式で入力してください（例: #FF0000）');
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
        await updateMemberGroup(group.group_id, input);
      } else {
        await createMemberGroup(input);
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
              placeholder="例: Aチーム、初心者グループ"
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

          {/* カラー選択 */}
          <div className="mb-4">
            <label className="label">カラー</label>

            {/* 基本色 */}
            <div className="mb-2">
              <span className="text-xs text-gray-500 mb-1 block">基本色</span>
              <div className="flex flex-wrap gap-2" role="radiogroup" aria-label="基本色の選択">
                {PRESET_COLORS.basic.map((preset) => (
                  <button
                    key={preset.color}
                    type="button"
                    onClick={() => setColor(preset.color)}
                    className={`w-8 h-8 rounded-lg border-2 transition-all ${
                      color === preset.color
                        ? 'border-gray-800 scale-110 shadow-md'
                        : 'border-transparent hover:border-gray-300'
                    }`}
                    style={{ backgroundColor: preset.color }}
                    disabled={loading}
                    title={preset.name}
                    aria-label={`${preset.name} (${preset.color})`}
                    aria-pressed={color === preset.color}
                  />
                ))}
              </div>
            </div>

            {/* パステルカラー */}
            <div className="mb-2">
              <span className="text-xs text-gray-500 mb-1 block">パステル</span>
              <div className="flex flex-wrap gap-2" role="radiogroup" aria-label="パステルカラーの選択">
                {PRESET_COLORS.pastel.map((preset) => (
                  <button
                    key={preset.color}
                    type="button"
                    onClick={() => setColor(preset.color)}
                    className={`w-8 h-8 rounded-lg border-2 transition-all ${
                      color === preset.color
                        ? 'border-gray-800 scale-110 shadow-md'
                        : 'border-transparent hover:border-gray-300'
                    }`}
                    style={{ backgroundColor: preset.color }}
                    disabled={loading}
                    title={preset.name}
                    aria-label={`${preset.name} (${preset.color})`}
                    aria-pressed={color === preset.color}
                  />
                ))}
              </div>
            </div>

            {/* カスタムカラー入力 */}
            <div className="flex items-center gap-2 mt-2">
              <span className="text-xs text-gray-500">カスタム:</span>
              <div
                className="w-6 h-6 rounded border border-gray-300"
                style={{ backgroundColor: color }}
                aria-label={`現在の色: ${color}`}
              ></div>
              <input
                type="text"
                id="color"
                value={color}
                onChange={(e) => setColor(e.target.value)}
                className="input-field flex-1"
                disabled={loading}
                placeholder="#10B981"
                pattern="^#[0-9A-Fa-f]{6}$"
                aria-label="カスタムカラーコード"
              />
            </div>
          </div>

          {/* 表示順序 */}
          <div className="mb-4">
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

// メンバー割り当てモーダル
function AssignMembersModal({
  group,
  members,
  onClose,
  onSuccess,
}: {
  group: MemberGroup;
  members: Member[];
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [selectedMemberIds, setSelectedMemberIds] = useState<string[]>(group.member_ids || []);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');

  // メンバー検索
  const filteredMembers = members.filter((m) =>
    m.display_name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const toggleMember = (memberId: string) => {
    if (selectedMemberIds.includes(memberId)) {
      setSelectedMemberIds(selectedMemberIds.filter((id) => id !== memberId));
    } else {
      setSelectedMemberIds([...selectedMemberIds, memberId]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await assignMembersToGroup(group.group_id, selectedMemberIds);
      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('メンバーの割り当てに失敗しました');
      }
      console.error('Failed to assign members:', err);
    } finally {
      setLoading(false);
    }
  };

  // 全選択・全解除
  const selectAll = () => {
    setSelectedMemberIds(filteredMembers.map((m) => m.member_id));
  };

  const deselectAll = () => {
    setSelectedMemberIds([]);
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
            {group.name} のメンバー
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
              placeholder="メンバーを検索..."
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
              {selectedMemberIds.length}人選択中
            </span>
          </div>

          {/* メンバー一覧 */}
          <div className="border border-gray-300 rounded-lg p-3 max-h-64 overflow-y-auto mb-4">
            {filteredMembers.length === 0 ? (
              <p className="text-sm text-gray-500 text-center py-4">
                メンバーが見つかりません
              </p>
            ) : (
              <div className="space-y-1">
                {filteredMembers.map((member) => (
                  <label
                    key={member.member_id}
                    className="flex items-center p-2 hover:bg-gray-50 rounded cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={selectedMemberIds.includes(member.member_id)}
                      onChange={() => toggleMember(member.member_id)}
                      disabled={loading}
                      className="w-4 h-4 text-accent border-gray-300 rounded focus:ring-accent"
                    />
                    <span className="ml-3 text-sm text-gray-900">{member.display_name}</span>
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
