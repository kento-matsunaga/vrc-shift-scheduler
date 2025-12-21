import { useState, useEffect } from 'react';
import { getMembers, createMember, updateMember, getRecentAttendance, deleteMember, bulkImportMembers, type BulkImportResponse } from '../lib/api/memberApi';
import { getActualAttendance } from '../lib/api/actualAttendanceApi';
import { listRoles, type Role } from '../lib/api/roleApi';
import type { Member, RecentAttendanceResponse } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function Members() {
  const [members, setMembers] = useState<Member[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // フィルター（複数選択）
  const [filterRoleIds, setFilterRoleIds] = useState<string[]>([]);

  // 新規登録・編集フォーム
  const [showForm, setShowForm] = useState(false);
  const [editingMember, setEditingMember] = useState<Member | null>(null);
  const [displayName, setDisplayName] = useState('');
  const [discordUserId, setDiscordUserId] = useState('');
  const [email, setEmail] = useState('');
  const [isActive, setIsActive] = useState(true);
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);

  // 本出席モーダル
  const [showActualAttendanceModal, setShowActualAttendanceModal] = useState(false);
  const [actualAttendanceData, setActualAttendanceData] = useState<RecentAttendanceResponse | null>(null);
  const [loadingActualAttendance, setLoadingActualAttendance] = useState(false);

  // 出欠確認モーダル
  const [showAttendanceConfirmationModal, setShowAttendanceConfirmationModal] = useState(false);
  const [attendanceConfirmationData, setAttendanceConfirmationData] = useState<RecentAttendanceResponse | null>(null);
  const [loadingAttendanceConfirmation, setLoadingAttendanceConfirmation] = useState(false);

  // 一括登録モーダル
  const [showBulkImportModal, setShowBulkImportModal] = useState(false);
  const [bulkImportText, setBulkImportText] = useState('');
  const [bulkImportSubmitting, setBulkImportSubmitting] = useState(false);
  const [bulkImportResult, setBulkImportResult] = useState<BulkImportResponse | null>(null);

  // データ取得
  const fetchData = async () => {
    try {
      setLoading(true);
      const [membersResponse, rolesData] = await Promise.all([
        getMembers(),
        listRoles(),
      ]);
      setMembers(membersResponse.members || []);
      setRoles(rolesData || []);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch data:', err);
      setError('データの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  // フィルター後のメンバー（選択したロールのいずれかを持つメンバーを表示）
  const filteredMembers = filterRoleIds.length > 0
    ? members.filter((m) => m.role_ids?.some((roleId) => filterRoleIds.includes(roleId)))
    : members;

  // 本出席データを取得
  const fetchActualAttendance = async () => {
    try {
      setLoadingActualAttendance(true);
      const data = await getActualAttendance({ limit: 30 });
      setActualAttendanceData(data);
    } catch (err) {
      console.error('Failed to fetch actual attendance:', err);
      alert('本出席データの取得に失敗しました');
    } finally {
      setLoadingActualAttendance(false);
    }
  };

  // 本出席モーダルを開く
  const handleOpenActualAttendance = async () => {
    setShowActualAttendanceModal(true);
    if (!actualAttendanceData) {
      await fetchActualAttendance();
    }
  };

  // 出欠確認データを取得
  const fetchAttendanceConfirmation = async () => {
    try {
      setLoadingAttendanceConfirmation(true);
      const data = await getRecentAttendance({ limit: 30 });
      setAttendanceConfirmationData(data);
    } catch (err) {
      console.error('Failed to fetch attendance confirmation:', err);
      alert('出欠確認データの取得に失敗しました');
    } finally {
      setLoadingAttendanceConfirmation(false);
    }
  };

  // 出欠確認モーダルを開く
  const handleOpenAttendanceConfirmation = async () => {
    setShowAttendanceConfirmationModal(true);
    if (!attendanceConfirmationData) {
      await fetchAttendanceConfirmation();
    }
  };

  // 一括登録モーダルを開く
  const handleOpenBulkImport = () => {
    setBulkImportText('');
    setBulkImportResult(null);
    setShowBulkImportModal(true);
  };

  // 一括登録を実行
  const handleBulkImport = async () => {
    const lines = bulkImportText
      .split('\n')
      .map((line) => line.trim())
      .filter((line) => line.length > 0);

    if (lines.length === 0) {
      alert('メンバー名を入力してください');
      return;
    }

    if (lines.length > 100) {
      alert('一度に登録できるのは100名までです');
      return;
    }

    try {
      setBulkImportSubmitting(true);
      const result = await bulkImportMembers(lines);
      setBulkImportResult(result);
      if (result.success_count > 0) {
        await fetchData();
      }
    } catch (err) {
      if (err instanceof ApiClientError) {
        alert(err.getUserMessage());
      } else {
        alert('一括登録に失敗しました');
      }
      console.error('Bulk import failed:', err);
    } finally {
      setBulkImportSubmitting(false);
    }
  };

  // 新規登録フォームを開く
  const handleOpenCreateForm = () => {
    setEditingMember(null);
    setDisplayName('');
    setDiscordUserId('');
    setEmail('');
    setIsActive(true);
    setSelectedRoleIds([]);
    setShowForm(true);
  };

  // 編集フォームを開く
  const handleOpenEditForm = (member: Member) => {
    setEditingMember(member);
    setDisplayName(member.display_name);
    setDiscordUserId(member.discord_user_id || '');
    setEmail(member.email || '');
    setIsActive(member.is_active);
    setSelectedRoleIds(member.role_ids || []);
    setShowForm(true);
  };

  // メンバー登録・更新
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!displayName.trim()) {
      alert('表示名は必須です');
      return;
    }

    try {
      setSubmitting(true);

      if (editingMember) {
        // 更新
        await updateMember(editingMember.member_id, {
          display_name: displayName,
          discord_user_id: discordUserId,
          email: email,
          is_active: isActive,
          role_ids: selectedRoleIds,
        });
      } else {
        // 新規作成
        await createMember({
          display_name: displayName,
          discord_user_id: discordUserId,
          email: email,
        });
      }

      await fetchData();
      setShowForm(false);
    } catch (err) {
      if (err instanceof ApiClientError) {
        alert(err.getUserMessage());
      } else {
        alert(editingMember ? 'メンバーの更新に失敗しました' : 'メンバーの登録に失敗しました');
      }
      console.error('Failed to save member:', err);
    } finally {
      setSubmitting(false);
    }
  };

  // ロールIDからロール名を取得
  const getRoleName = (roleId: string) => {
    const role = roles.find((r) => r.role_id === roleId);
    return role?.name || 'Unknown';
  };

  // ロールIDからロールカラーを取得
  const getRoleColor = (roleId: string) => {
    const role = roles.find((r) => r.role_id === roleId);
    return role?.color || '#6B7280';
  };

  // メンバー削除
  const handleDeleteMember = async (member: Member) => {
    if (!confirm(`「${member.display_name}」を削除しますか？\nこの操作は取り消せません。`)) {
      return;
    }

    try {
      await deleteMember(member.member_id);
      await fetchData();
    } catch (err) {
      if (err instanceof ApiClientError) {
        alert(err.getUserMessage());
      } else {
        alert('メンバーの削除に失敗しました');
      }
      console.error('Failed to delete member:', err);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">メンバー管理</h2>
        <div className="flex gap-3">
          <button onClick={handleOpenActualAttendance} className="btn-secondary text-sm">
            本出席を見る
          </button>
          <button onClick={handleOpenAttendanceConfirmation} className="btn-secondary text-sm">
            出欠確認を見る
          </button>
          <button onClick={handleOpenBulkImport} className="btn-secondary">
            一括登録
          </button>
          <button onClick={handleOpenCreateForm} className="btn-primary">
            ＋ メンバーを追加
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {/* ロールフィルター（複数選択） */}
      {roles.length > 0 && (
        <div className="mb-6">
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-medium text-gray-700">
              ロールでフィルター
            </label>
            {filterRoleIds.length > 0 && (
              <button
                onClick={() => setFilterRoleIds([])}
                className="text-xs text-indigo-600 hover:text-indigo-800"
              >
                クリア
              </button>
            )}
          </div>
          <div className="flex flex-wrap gap-2">
            {roles.map((role) => {
              const isSelected = filterRoleIds.includes(role.role_id);
              return (
                <button
                  key={role.role_id}
                  onClick={() => {
                    if (isSelected) {
                      setFilterRoleIds(filterRoleIds.filter((id) => id !== role.role_id));
                    } else {
                      setFilterRoleIds([...filterRoleIds, role.role_id]);
                    }
                  }}
                  className={`inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium transition-all ${
                    isSelected
                      ? 'ring-2 ring-offset-1 ring-indigo-500'
                      : 'opacity-60 hover:opacity-100'
                  }`}
                  style={{
                    backgroundColor: role.color || '#6B7280',
                    color: 'white',
                  }}
                >
                  {role.name}
                  {isSelected && (
                    <svg className="w-4 h-4 ml-1" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  )}
                </button>
              );
            })}
          </div>
          {filterRoleIds.length > 0 && (
            <p className="text-xs text-gray-500 mt-2">
              {filterRoleIds.length}個のロールでフィルター中（{filteredMembers.length}人表示）
            </p>
          )}
        </div>
      )}

      {/* メンバー一覧 */}
      {filteredMembers.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">
            {filterRoleIds.length > 0 ? '選択したロールのメンバーはいません' : 'まだメンバーがいません'}
          </p>
          {filterRoleIds.length === 0 && (
            <button onClick={handleOpenCreateForm} className="btn-primary">
              最初のメンバーを追加
            </button>
          )}
        </div>
      ) : (
        <div className="card overflow-x-auto">
          <table className="min-w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="px-4 py-3 text-left text-sm font-semibold text-gray-900">名前</th>
                <th className="px-4 py-3 text-left text-sm font-semibold text-gray-900">ロール</th>
                <th className="px-4 py-3 text-left text-sm font-semibold text-gray-900">Discord ID</th>
                <th className="px-4 py-3 text-left text-sm font-semibold text-gray-900">Email</th>
                <th className="px-4 py-3 text-left text-sm font-semibold text-gray-900">ステータス</th>
                <th className="px-4 py-3 text-right text-sm font-semibold text-gray-900">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filteredMembers.map((member) => (
                <tr key={member.member_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm text-gray-900">{member.display_name}</td>
                  <td className="px-4 py-3 text-sm">
                    <div className="flex flex-wrap gap-1">
                      {member.role_ids && member.role_ids.length > 0 ? (
                        member.role_ids.map((roleId) => (
                          <span
                            key={roleId}
                            className="inline-flex items-center px-2 py-1 rounded text-xs font-medium text-white"
                            style={{ backgroundColor: getRoleColor(roleId) }}
                          >
                            {getRoleName(roleId)}
                          </span>
                        ))
                      ) : (
                        <span className="text-gray-400">なし</span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {member.discord_user_id || '-'}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">{member.email || '-'}</td>
                  <td className="px-4 py-3 text-sm">
                    <span
                      className={`inline-flex px-2 py-1 text-xs font-semibold rounded ${
                        member.is_active
                          ? 'bg-green-100 text-green-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {member.is_active ? 'アクティブ' : '非アクティブ'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-right space-x-3">
                    <button
                      onClick={() => handleOpenEditForm(member)}
                      className="text-indigo-600 hover:text-indigo-800 font-medium"
                    >
                      編集
                    </button>
                    <button
                      onClick={() => handleDeleteMember(member)}
                      className="text-red-600 hover:text-red-800 font-medium"
                    >
                      削除
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* メンバー登録・編集フォーム */}
      {showForm && (
        <MemberFormModal
          member={editingMember}
          roles={roles}
          displayName={displayName}
          setDisplayName={setDisplayName}
          discordUserId={discordUserId}
          setDiscordUserId={setDiscordUserId}
          email={email}
          setEmail={setEmail}
          isActive={isActive}
          setIsActive={setIsActive}
          selectedRoleIds={selectedRoleIds}
          setSelectedRoleIds={setSelectedRoleIds}
          submitting={submitting}
          onSubmit={handleSubmit}
          onClose={() => setShowForm(false)}
        />
      )}

      {/* 本出席モーダル */}
      {showActualAttendanceModal && (
        <ActualAttendanceModal
          data={actualAttendanceData}
          loading={loadingActualAttendance}
          onClose={() => setShowActualAttendanceModal(false)}
        />
      )}

      {/* 出欠確認モーダル */}
      {showAttendanceConfirmationModal && (
        <AttendanceConfirmationModal
          data={attendanceConfirmationData}
          loading={loadingAttendanceConfirmation}
          onClose={() => setShowAttendanceConfirmationModal(false)}
        />
      )}

      {/* 一括登録モーダル */}
      {showBulkImportModal && (
        <BulkImportModal
          text={bulkImportText}
          setText={setBulkImportText}
          submitting={bulkImportSubmitting}
          result={bulkImportResult}
          onSubmit={handleBulkImport}
          onClose={() => setShowBulkImportModal(false)}
        />
      )}
    </div>
  );
}

// メンバーフォームモーダル
function MemberFormModal({
  member,
  roles,
  displayName,
  setDisplayName,
  discordUserId,
  setDiscordUserId,
  email,
  setEmail,
  isActive,
  setIsActive,
  selectedRoleIds,
  setSelectedRoleIds,
  submitting,
  onSubmit,
  onClose,
}: {
  member: Member | null;
  roles: Role[];
  displayName: string;
  setDisplayName: (v: string) => void;
  discordUserId: string;
  setDiscordUserId: (v: string) => void;
  email: string;
  setEmail: (v: string) => void;
  isActive: boolean;
  setIsActive: (v: boolean) => void;
  selectedRoleIds: string[];
  setSelectedRoleIds: (v: string[]) => void;
  submitting: boolean;
  onSubmit: (e: React.FormEvent) => void;
  onClose: () => void;
}) {
  const toggleRole = (roleId: string) => {
    if (selectedRoleIds.includes(roleId)) {
      setSelectedRoleIds(selectedRoleIds.filter((id) => id !== roleId));
    } else {
      setSelectedRoleIds([...selectedRoleIds, roleId]);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6 max-h-[90vh] overflow-y-auto">
        <h3 className="text-xl font-bold text-gray-900 mb-4">
          {member ? 'メンバーを編集' : 'メンバーを追加'}
        </h3>

        <form onSubmit={onSubmit}>
          <div className="mb-4">
            <label className="label">
              表示名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              className="input-field"
              required
              disabled={submitting}
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label className="label">Discord User ID</label>
            <input
              type="text"
              value={discordUserId}
              onChange={(e) => setDiscordUserId(e.target.value)}
              className="input-field"
              disabled={submitting}
              placeholder="オプション"
            />
          </div>

          <div className="mb-4">
            <label className="label">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="input-field"
              disabled={submitting}
              placeholder="オプション"
            />
          </div>

          {member && (
            <div className="mb-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={isActive}
                  onChange={(e) => setIsActive(e.target.checked)}
                  className="mr-2"
                  disabled={submitting}
                />
                <span className="text-sm font-medium text-gray-700">アクティブ</span>
              </label>
            </div>
          )}

          {/* ロール選択 */}
          {roles.length > 0 && (
            <div className="mb-4">
              <label className="label">ロール</label>
              <div className="border border-gray-300 rounded-md p-3 max-h-40 overflow-y-auto">
                {roles.map((role) => (
                  <label key={role.role_id} className="flex items-center mb-2 last:mb-0">
                    <input
                      type="checkbox"
                      checked={selectedRoleIds.includes(role.role_id)}
                      onChange={() => toggleRole(role.role_id)}
                      className="mr-2"
                      disabled={submitting}
                    />
                    <div className="flex items-center gap-2">
                      {role.color && (
                        <div
                          className="w-3 h-3 rounded"
                          style={{ backgroundColor: role.color }}
                        ></div>
                      )}
                      <span className="text-sm">{role.name}</span>
                    </div>
                  </label>
                ))}
              </div>
            </div>
          )}

          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 btn-secondary"
              disabled={submitting}
            >
              キャンセル
            </button>
            <button type="submit" className="flex-1 btn-primary" disabled={submitting}>
              {submitting ? '処理中...' : member ? '更新' : '登録'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// 本出席モーダル
function ActualAttendanceModal({
  data,
  loading,
  onClose,
}: {
  data: RecentAttendanceResponse | null;
  loading: boolean;
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-6xl w-full p-6 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h3 className="text-xl font-bold text-gray-900">本出席（実績）</h3>
            <p className="text-sm text-gray-600 mt-1">
              実際にシフトに割り当てられた実績データです。○: シフト割り当てあり、×: シフト割り当てなし
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-2xl"
          >
            ×
          </button>
        </div>

        {loading ? (
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">読み込み中...</p>
          </div>
        ) : data && data.target_dates && data.target_dates.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full text-xs border-collapse border border-gray-300">
              <thead>
                <tr className="bg-gray-100">
                  <th className="border border-gray-300 px-2 py-1 text-left font-semibold sticky left-0 bg-gray-100 z-10">
                    メンバー
                  </th>
                  {(data.target_dates || []).map((td) => (
                    <th key={td.target_date_id} className="border border-gray-300 px-2 py-1 text-center font-semibold whitespace-nowrap">
                      {new Date(td.target_date).toLocaleDateString('ja-JP', {
                        month: 'numeric',
                        day: 'numeric',
                      })}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(data.member_attendances || []).map((memberAtt) => (
                  <tr key={memberAtt.member_id} className="hover:bg-gray-50">
                    <td className="border border-gray-300 px-2 py-1 font-medium sticky left-0 bg-white z-10">
                      {memberAtt.member_name}
                    </td>
                    {(data.target_dates || []).map((td) => {
                      const status = memberAtt.attendance_map[td.target_date_id] || '';
                      let symbol = '×';
                      let color = 'text-red-600';
                      if (status === 'attended') {
                        symbol = '○';
                        color = 'text-green-600';
                      }
                      return (
                        <td key={td.target_date_id} className={`border border-gray-300 px-2 py-1 text-center ${color} font-bold`}>
                          {symbol}
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
            <p className="text-xs text-gray-500 mt-2">
              ○: シフト割り当てあり、×: シフト割り当てなし
            </p>
          </div>
        ) : (
          <div className="text-center py-12 text-gray-500">
            本出席データがありません
          </div>
        )}
      </div>
    </div>
  );
}

// 出欠確認モーダル
function AttendanceConfirmationModal({
  data,
  loading,
  onClose,
}: {
  data: RecentAttendanceResponse | null;
  loading: boolean;
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-6xl w-full p-6 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h3 className="text-xl font-bold text-gray-900">出欠確認（予定）</h3>
            <p className="text-sm text-gray-600 mt-1">
              メンバーが回答した出欠予定データです。○: 参加予定、×: 不参加、-: 未回答
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-2xl"
          >
            ×
          </button>
        </div>

        {loading ? (
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">読み込み中...</p>
          </div>
        ) : data && data.target_dates && data.target_dates.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full text-xs border-collapse border border-gray-300">
              <thead>
                <tr className="bg-gray-100">
                  <th className="border border-gray-300 px-2 py-1 text-left font-semibold sticky left-0 bg-gray-100 z-10">
                    メンバー
                  </th>
                  {(data.target_dates || []).map((td) => (
                    <th key={td.target_date_id} className="border border-gray-300 px-2 py-1 text-center font-semibold whitespace-nowrap">
                      {new Date(td.target_date).toLocaleDateString('ja-JP', {
                        month: 'numeric',
                        day: 'numeric',
                      })}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(data.member_attendances || []).map((memberAtt) => (
                  <tr key={memberAtt.member_id} className="hover:bg-gray-50">
                    <td className="border border-gray-300 px-2 py-1 font-medium sticky left-0 bg-white z-10">
                      {memberAtt.member_name}
                    </td>
                    {(data.target_dates || []).map((td) => {
                      const status = memberAtt.attendance_map[td.target_date_id] || '';
                      let symbol = '-';
                      let color = 'text-gray-400';
                      if (status === 'attending') {
                        symbol = '○';
                        color = 'text-green-600';
                      } else if (status === 'absent') {
                        symbol = '×';
                        color = 'text-red-600';
                      }
                      return (
                        <td key={td.target_date_id} className={`border border-gray-300 px-2 py-1 text-center ${color} font-bold`}>
                          {symbol}
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
            <p className="text-xs text-gray-500 mt-2">
              ○: 参加予定、×: 不参加、-: 未回答
            </p>
          </div>
        ) : (
          <div className="text-center py-12 text-gray-500">
            出欠確認データがありません
          </div>
        )}
      </div>
    </div>
  );
}

// 一括登録モーダル
function BulkImportModal({
  text,
  setText,
  submitting,
  result,
  onSubmit,
  onClose,
}: {
  text: string;
  setText: (v: string) => void;
  submitting: boolean;
  result: BulkImportResponse | null;
  onSubmit: () => void;
  onClose: () => void;
}) {
  const lineCount = text
    .split('\n')
    .filter((line) => line.trim().length > 0).length;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-lg w-full p-6 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-xl font-bold text-gray-900">メンバー一括登録</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-2xl"
          >
            ×
          </button>
        </div>

        {result ? (
          // 結果表示
          <div>
            <div className="mb-4 p-4 rounded-lg bg-gray-50">
              <div className="grid grid-cols-3 gap-4 text-center">
                <div>
                  <div className="text-2xl font-bold text-gray-900">{result.total_count}</div>
                  <div className="text-sm text-gray-600">合計</div>
                </div>
                <div>
                  <div className="text-2xl font-bold text-green-600">{result.success_count}</div>
                  <div className="text-sm text-gray-600">成功</div>
                </div>
                <div>
                  <div className="text-2xl font-bold text-red-600">{result.failed_count}</div>
                  <div className="text-sm text-gray-600">失敗</div>
                </div>
              </div>
            </div>

            {result.failed_count > 0 && (
              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-900 mb-2">エラー詳細:</h4>
                <div className="max-h-40 overflow-y-auto border border-gray-200 rounded p-2">
                  {result.results
                    .filter((r) => !r.success)
                    .map((r, i) => (
                      <div key={i} className="text-sm text-red-600 mb-1">
                        「{r.display_name}」: {r.error}
                      </div>
                    ))}
                </div>
              </div>
            )}

            <button
              onClick={onClose}
              className="w-full btn-primary"
            >
              閉じる
            </button>
          </div>
        ) : (
          // 入力フォーム
          <div>
            <p className="text-sm text-gray-600 mb-4">
              メンバー名を1行に1名ずつ入力してください。最大100名まで一度に登録できます。
            </p>

            <div className="mb-4">
              <textarea
                value={text}
                onChange={(e) => setText(e.target.value)}
                className="w-full h-48 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
                placeholder="例:
山田太郎
佐藤花子
鈴木一郎"
                disabled={submitting}
                autoFocus
              />
              <div className="text-right text-sm text-gray-500 mt-1">
                {lineCount} 名
              </div>
            </div>

            <div className="flex space-x-3">
              <button
                onClick={onClose}
                className="flex-1 btn-secondary"
                disabled={submitting}
              >
                キャンセル
              </button>
              <button
                onClick={onSubmit}
                className="flex-1 btn-primary"
                disabled={submitting || lineCount === 0}
              >
                {submitting ? '登録中...' : `${lineCount}名を登録`}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
