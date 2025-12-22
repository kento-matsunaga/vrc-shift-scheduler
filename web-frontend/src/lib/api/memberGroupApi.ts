import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

/**
 * メンバーグループ
 */
export interface MemberGroup {
  group_id: string;
  tenant_id: string;
  name: string;
  description: string;
  color: string;
  display_order: number;
  member_ids?: string[];
  created_at: string;
  updated_at: string;
}

/**
 * グループ一覧レスポンス
 */
export interface MemberGroupListResponse {
  groups: MemberGroup[];
}

/**
 * グループ作成
 */
export async function createMemberGroup(data: {
  name: string;
  description?: string;
  color?: string;
  display_order?: number;
}): Promise<MemberGroup> {
  const res = await apiClient.post<ApiResponse<MemberGroup>>('/api/v1/member-groups', data);
  return res.data;
}

/**
 * グループ一覧取得
 */
export async function getMemberGroups(): Promise<MemberGroupListResponse> {
  const res = await apiClient.get<ApiResponse<MemberGroupListResponse>>('/api/v1/member-groups');
  return res.data;
}

/**
 * グループ詳細取得
 */
export async function getMemberGroupDetail(groupId: string): Promise<MemberGroup> {
  const res = await apiClient.get<ApiResponse<MemberGroup>>(`/api/v1/member-groups/${groupId}`);
  return res.data;
}

/**
 * グループ更新
 */
export async function updateMemberGroup(groupId: string, data: {
  name: string;
  description?: string;
  color?: string;
  display_order?: number;
}): Promise<MemberGroup> {
  const res = await apiClient.put<ApiResponse<MemberGroup>>(`/api/v1/member-groups/${groupId}`, data);
  return res.data;
}

/**
 * グループ削除
 */
export async function deleteMemberGroup(groupId: string): Promise<void> {
  await apiClient.delete(`/api/v1/member-groups/${groupId}`);
}

/**
 * グループにメンバーを割り当て
 */
export async function assignMembersToGroup(groupId: string, memberIds: string[]): Promise<{ group_id: string; member_ids: string[] }> {
  const res = await apiClient.put<ApiResponse<{ group_id: string; member_ids: string[] }>>(`/api/v1/member-groups/${groupId}/members`, { member_ids: memberIds });
  return res.data;
}
