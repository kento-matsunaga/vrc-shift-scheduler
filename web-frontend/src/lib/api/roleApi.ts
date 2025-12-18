import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

export interface Role {
  role_id: string;
  tenant_id: string;
  name: string;
  description: string;
  color: string;
  display_order: number;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleInput {
  name: string;
  description?: string;
  color?: string;
  display_order?: number;
}

export interface UpdateRoleInput {
  name: string;
  description?: string;
  color?: string;
  display_order?: number;
}

export interface RoleListResponse {
  roles: Role[];
}

/**
 * ロール一覧を取得
 */
export async function listRoles(): Promise<Role[]> {
  const response = await apiClient.get<ApiResponse<RoleListResponse>>('/api/v1/roles');
  return response.data.roles;
}

/**
 * ロールを作成
 */
export async function createRole(input: CreateRoleInput): Promise<Role> {
  const response = await apiClient.post<ApiResponse<Role>>('/api/v1/roles', input);
  return response.data;
}

/**
 * ロールを更新
 */
export async function updateRole(roleId: string, input: UpdateRoleInput): Promise<Role> {
  const response = await apiClient.put<ApiResponse<Role>>(`/api/v1/roles/${roleId}`, input);
  return response.data;
}

/**
 * ロールを削除
 */
export async function deleteRole(roleId: string): Promise<void> {
  await apiClient.delete(`/api/v1/roles/${roleId}`);
}

/**
 * ロールを取得
 */
export async function getRole(roleId: string): Promise<Role> {
  const response = await apiClient.get<ApiResponse<Role>>(`/api/v1/roles/${roleId}`);
  return response.data;
}
