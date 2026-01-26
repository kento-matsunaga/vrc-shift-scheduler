import type { ApiResponse } from '../../types/api';

/**
 * ログインリクエスト（管理者認証）
 */
export interface LoginRequest {
  // tenant_id削除: email + password のみ
  email: string;
  password: string;
}

/**
 * ログインレスポンス
 */
export interface LoginResponse {
  token: string;
  admin_id: string;
  tenant_id: string;
  role: string;
  expires_at: string;
}

/**
 * セットアップリクエスト
 */
export interface SetupRequest {
  organization_name: string;
  admin_name: string;
  password: string;
  timezone?: string;
}

/**
 * セットアップレスポンス
 */
export interface SetupResponse {
  tenant_id: string;
  member_id: string;
  token: string;
  message: string;
  invite_url: string;
}

/**
 * 招待URL経由でのメンバー登録リクエスト
 */
export interface RegisterByInviteRequest {
  invite_token: string;
  display_name: string;
  password: string;
}

/**
 * 招待URL経由でのメンバー登録レスポンス
 */
export interface RegisterByInviteResponse {
  tenant_id: string;
  member_id: string;
  token: string;
}

/**
 * ログイン
 */
export async function login(data: LoginRequest): Promise<LoginResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  
  const response = await fetch(`${baseURL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    // エラーレスポンスを安全にパース
    let errorData: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch (e) {
        // JSONパースに失敗した場合は、テキストとして読み取る
        const text = await response.text();
        throw new Error(`ログインに失敗しました: ${text || response.statusText}`);
      }
    } else {
      // JSONでない場合はテキストとして読み取る
      const text = await response.text();
      throw new Error(`ログインに失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to login');
  }

  // 成功レスポンスをパース
  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<LoginResponse> = await response.json();
  return result.data;
}

/**
 * 初回セットアップ（テナントと管理者を作成）
 */
export async function setup(data: SetupRequest): Promise<SetupResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  
  const response = await fetch(`${baseURL}/api/v1/setup`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    // エラーレスポンスを安全にパース
    let errorData: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch (e) {
        // JSONパースに失敗した場合は、テキストとして読み取る
        const text = await response.text();
        throw new Error(`セットアップに失敗しました: ${text || response.statusText}`);
      }
    } else {
      // JSONでない場合はテキストとして読み取る
      const text = await response.text();
      throw new Error(`セットアップに失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to setup');
  }

  // 成功レスポンスをパース
  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<SetupResponse> = await response.json();
  return result.data;
}

/**
 * 招待URL経由でのメンバー登録
 */
export async function registerByInvite(data: RegisterByInviteRequest): Promise<RegisterByInviteResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';

  const response = await fetch(`${baseURL}/api/v1/auth/register-by-invite`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    // エラーレスポンスを安全にパース
    let errorData: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch (e) {
        // JSONパースに失敗した場合は、テキストとして読み取る
        const text = await response.text();
        throw new Error(`登録に失敗しました: ${text || response.statusText}`);
      }
    } else {
      // JSONでない場合はテキストとして読み取る
      const text = await response.text();
      throw new Error(`登録に失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to register by invite');
  }

  // 成功レスポンスをパース
  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<RegisterByInviteResponse> = await response.json();
  return result.data;
}

// ============================================================
// パスワードリセット関連
// ============================================================

/**
 * パスワードリセット状態確認レスポンス
 */
export interface PasswordResetStatusResponse {
  allowed: boolean;
  expires_at?: string;
  tenant_id?: string;
}

/**
 * パスワードリセットリクエスト
 */
export interface ResetPasswordRequest {
  email: string;
  license_key: string;
  new_password: string;
  confirm_new_password: string;
}

/**
 * パスワードリセットレスポンス
 */
export interface ResetPasswordResponse {
  success: boolean;
  message: string;
}

/**
 * パスワードリセット状態を確認
 * (認証不要)
 */
export async function checkPasswordResetStatus(email: string): Promise<PasswordResetStatusResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';

  const response = await fetch(`${baseURL}/api/v1/auth/password-reset-status?email=${encodeURIComponent(email)}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    let errorData: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch (e) {
        const text = await response.text();
        throw new Error(`確認に失敗しました: ${text || response.statusText}`);
      }
    } else {
      const text = await response.text();
      throw new Error(`確認に失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to check password reset status');
  }

  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<PasswordResetStatusResponse> = await response.json();
  return result.data;
}

/**
 * パスワードをリセット
 * (認証不要、ライセンスキーで本人確認)
 */
export async function resetPassword(data: ResetPasswordRequest): Promise<ResetPasswordResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';

  const response = await fetch(`${baseURL}/api/v1/auth/reset-password`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    let errorData: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch (e) {
        const text = await response.text();
        throw new Error(`リセットに失敗しました: ${text || response.statusText}`);
      }
    } else {
      const text = await response.text();
      throw new Error(`リセットに失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to reset password');
  }

  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<ResetPasswordResponse> = await response.json();
  return result.data;
}

// ============================================================
// メールベースのパスワードリセット関連
// ============================================================

/**
 * パスワードリセットリクエスト（メール送信）
 */
export interface ForgotPasswordRequest {
  email: string;
}

/**
 * パスワードリセットリクエストレスポンス
 */
export interface ForgotPasswordResponse {
  success: boolean;
  message: string;
}

/**
 * パスワードリセット用メールを送信
 * (認証不要)
 */
export async function forgotPassword(data: ForgotPasswordRequest): Promise<ForgotPasswordResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';

  const response = await fetch(`${baseURL}/api/v1/auth/forgot-password`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    let errorData: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch (e) {
        const text = await response.text();
        throw new Error(`リクエストに失敗しました: ${text || response.statusText}`);
      }
    } else {
      const text = await response.text();
      throw new Error(`リクエストに失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to request password reset');
  }

  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<ForgotPasswordResponse> = await response.json();
  return result.data;
}

/**
 * トークンによるパスワードリセットリクエスト
 */
export interface ResetPasswordWithTokenRequest {
  token: string;
  new_password: string;
  confirm_new_password: string;
}

/**
 * トークンによるパスワードリセットレスポンス
 */
export interface ResetPasswordWithTokenResponse {
  success: boolean;
  message: string;
}

/**
 * トークンを使用してパスワードをリセット
 * (認証不要)
 */
export async function resetPasswordWithToken(data: ResetPasswordWithTokenRequest): Promise<ResetPasswordWithTokenResponse> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';

  const response = await fetch(`${baseURL}/api/v1/auth/reset-password-with-token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    let errorData: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        errorData = await response.json();
      } catch (e) {
        const text = await response.text();
        throw new Error(`リセットに失敗しました: ${text || response.statusText}`);
      }
    } else {
      const text = await response.text();
      throw new Error(`リセットに失敗しました: ${text || response.statusText}`);
    }
    throw new Error(errorData.error?.message || 'Failed to reset password with token');
  }

  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    const text = await response.text();
    throw new Error(`予期しないレスポンス形式: ${text}`);
  }

  const result: ApiResponse<ResetPasswordWithTokenResponse> = await response.json();
  return result.data;
}

