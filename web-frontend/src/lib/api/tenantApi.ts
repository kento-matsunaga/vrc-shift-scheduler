import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

/**
 * Tenant response type
 */
export interface Tenant {
  tenant_id: string;
  tenant_name: string;
  timezone: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * Get current tenant info
 */
export async function getCurrentTenant(): Promise<Tenant> {
  const res = await apiClient.get<ApiResponse<Tenant>>('/api/v1/tenants/me');
  return res.data;
}

/**
 * Update current tenant
 */
export async function updateTenant(data: { tenant_name: string }): Promise<Tenant> {
  const res = await apiClient.put<ApiResponse<Tenant>>('/api/v1/tenants/me', data);
  return res.data;
}

/**
 * Manager permissions response type
 */
export interface ManagerPermissions {
  can_add_member: boolean;
  can_edit_member: boolean;
  can_delete_member: boolean;
  can_create_event: boolean;
  can_edit_event: boolean;
  can_delete_event: boolean;
  can_assign_shift: boolean;
  can_edit_shift: boolean;
  can_create_attendance: boolean;
  can_create_schedule: boolean;
  can_manage_roles: boolean;
  can_manage_positions: boolean;
  can_manage_groups: boolean;
  can_invite_manager: boolean;
}

/**
 * Get manager permissions
 */
export async function getManagerPermissions(): Promise<ManagerPermissions> {
  const res = await apiClient.get<ApiResponse<ManagerPermissions>>('/api/v1/settings/manager-permissions');
  return res.data;
}

/**
 * Update manager permissions (owner only)
 */
export async function updateManagerPermissions(data: ManagerPermissions): Promise<ManagerPermissions> {
  const res = await apiClient.put<ApiResponse<ManagerPermissions>>('/api/v1/settings/manager-permissions', data);
  return res.data;
}
