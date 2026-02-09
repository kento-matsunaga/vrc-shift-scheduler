import type { ApiResponse } from '../../types/api';

/**
 * 管理者招待リクエスト
 */
export interface InviteAdminRequest {
  email: string;
  role: string;
}

/**
 * 管理者招待レスポンス
 */
export interface InviteAdminResponse {
  invitation_id: string;
  email: string;
  role: string;
  token: string;
  expires_at: string;
}

/**
 * 招待受理リクエスト
 */
export interface AcceptInvitationRequest {
  display_name: string;
  password: string;
}

/**
 * 招待受理レスポンス
 */
export interface AcceptInvitationResponse {
  admin_id: string;
  tenant_id: string;
  email: string;
  role: string;
}

/**
 * 管理者を招待
 */
export async function inviteAdmin(data: InviteAdminRequest): Promise<InviteAdminResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/invitations`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    let errorData: { error?: { message?: string } } | undefined;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch {
        const text = await response.text();
        throw new Error(`招待に失敗しました: ${text || response.statusText}`);
      }
    } else {
      const text = await response.text();
      throw new Error(`招待に失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to invite admin');
  }

  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<InviteAdminResponse> = await response.json();
  return result.data;
}

/**
 * 招待を受理
 */
export async function acceptInvitation(
  token: string,
  data: AcceptInvitationRequest
): Promise<AcceptInvitationResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';

  const response = await fetch(`${baseURL}/api/v1/invitations/accept/${token}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    let errorData: { error?: { message?: string } } | undefined;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch {
        const text = await response.text();
        throw new Error(`登録に失敗しました: ${text || response.statusText}`);
      }
    } else {
      const text = await response.text();
      throw new Error(`登録に失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to accept invitation');
  }

  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<AcceptInvitationResponse> = await response.json();
  return result.data;
}
