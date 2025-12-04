import type { ApiResponse } from '../../types/api';

/**
 * ログインリクエスト
 */
export interface LoginRequest {
  tenant_id: string;
  display_name: string;
  password: string;
}

/**
 * ログインレスポンス
 */
export interface LoginResponse {
  token: string;
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
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  
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
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  
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
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  
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

