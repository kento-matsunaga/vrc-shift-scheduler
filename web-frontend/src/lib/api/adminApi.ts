import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

/**
 * Change password request type
 */
export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  confirm_new_password: string;
}

/**
 * Change password response type
 */
export interface ChangePasswordResponse {
  message: string;
}

/**
 * Change current admin's password
 */
export async function changePassword(data: ChangePasswordRequest): Promise<ChangePasswordResponse> {
  const res = await apiClient.post<ApiResponse<ChangePasswordResponse>>(
    '/api/v1/admins/me/change-password',
    data
  );
  return res.data;
}

/**
 * Allow password reset response type
 */
export interface AllowPasswordResetResponse {
  target_admin_id: string;
  target_email: string;
  allowed_at: string;
  expires_at: string;
  allowed_by_name: string;
  message: string;
}

/**
 * Allow password reset for another admin (Owner only)
 */
export async function allowPasswordReset(adminId: string): Promise<AllowPasswordResetResponse> {
  const res = await apiClient.post<ApiResponse<AllowPasswordResetResponse>>(
    `/api/v1/admins/${adminId}/allow-password-reset`,
    {}
  );
  return res.data;
}
