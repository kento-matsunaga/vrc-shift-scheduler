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
