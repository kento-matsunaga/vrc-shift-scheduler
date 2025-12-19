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
