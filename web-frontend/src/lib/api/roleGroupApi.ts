import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

/**
 * ロールグループ
 */
export interface RoleGroup {
  group_id: string;
  tenant_id: string;
  name: string;
  description: string;
  color: string;
  display_order: number;
  role_ids?: string[];
  created_at: string;
  updated_at: string;
}

/**
 * グループ一覧レスポンス
 */
export interface RoleGroupListResponse {
  groups: RoleGroup[];
}

/**
 * グループ作成
 */
export async function createRoleGroup(data: {
  name: string;
  description?: string;
  color?: string;
  display_order?: number;
}): Promise<RoleGroup> {
  const res = await apiClient.post<ApiResponse<RoleGroup>>('/api/v1/role-groups', data);
  return res.data;
}

/**
 * グループ一覧取得
 */
export async function getRoleGroups(): Promise<RoleGroupListResponse> {
  const res = await apiClient.get<ApiResponse<RoleGroupListResponse>>('/api/v1/role-groups');
  return res.data;
}

/**
 * グループ詳細取得
 */
export async function getRoleGroupDetail(groupId: string): Promise<RoleGroup> {
  const res = await apiClient.get<ApiResponse<RoleGroup>>(`/api/v1/role-groups/${groupId}`);
  return res.data;
}

/**
 * グループ更新
 */
export async function updateRoleGroup(groupId: string, data: {
  name: string;
  description?: string;
  color?: string;
  display_order?: number;
}): Promise<RoleGroup> {
  const res = await apiClient.put<ApiResponse<RoleGroup>>(`/api/v1/role-groups/${groupId}`, data);
  return res.data;
}

/**
 * グループ削除
 */
export async function deleteRoleGroup(groupId: string): Promise<void> {
  await apiClient.delete(`/api/v1/role-groups/${groupId}`);
}

/**
 * グループにロールを割り当て
 */
export async function assignRolesToGroup(groupId: string, roleIds: string[]): Promise<{ group_id: string; role_ids: string[] }> {
  const res = await apiClient.put<ApiResponse<{ group_id: string; role_ids: string[] }>>(`/api/v1/role-groups/${groupId}/roles`, { role_ids: roleIds });
  return res.data;
}
