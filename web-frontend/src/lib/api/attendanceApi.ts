import type { ApiResponse } from '../../types/api';

/**
 * 対象日入力（リクエスト用）
 */
export interface TargetDateInput {
  target_date: string; // ISO 8601 format
  start_time?: string; // HH:MM format (optional)
  end_time?: string;   // HH:MM format (optional)
}

/**
 * 出欠確認作成リクエスト
 */
export interface CreateAttendanceRequest {
  title: string;
  description: string;
  target_type: 'event' | 'business_day';
  target_id?: string;
  target_dates?: TargetDateInput[]; // ISO 8601 format array with optional start/end time
  deadline?: string; // ISO 8601 format
  group_ids?: string[]; // optional: target member group IDs
  role_ids?: string[]; // optional: target role IDs
}

/**
 * 対象日（レスポンス用）
 */
export interface TargetDate {
  target_date_id: string;
  target_date: string;
  start_time?: string; // HH:MM format (optional)
  end_time?: string;   // HH:MM format (optional)
  display_order: number;
}

/**
 * 出欠確認レスポンス
 */
export interface AttendanceCollection {
  collection_id: string;
  tenant_id: string;
  title: string;
  description: string;
  target_type: string;
  target_id: string;
  target_dates?: TargetDate[];
  public_token: string;
  status: 'open' | 'closed';
  deadline?: string;
  target_date_count?: number;
  response_count?: number;
  group_ids?: string[]; // 対象グループIDs
  role_ids?: string[]; // 対象ロールIDs
  created_at: string;
  updated_at: string;
}

/**
 * 出欠回答
 */
export interface AttendanceResponse {
  response_id: string;
  member_id: string;
  member_name: string; // メンバー表示名
  target_date_id: string; // 対象日ID
  target_date: string; // 対象日（ISO 8601）
  response: 'attending' | 'absent' | 'undecided';
  note: string;
  available_from?: string; // 参加可能開始時間 (HH:MM)
  available_to?: string;   // 参加可能終了時間 (HH:MM)
  responded_at: string;
}

/**
 * 出欠確認一覧を取得
 */
export async function listAttendanceCollections(): Promise<AttendanceCollection[]> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/attendance/collections`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`出欠確認一覧の取得に失敗しました: ${text || response.statusText}`);
  }

  const result = await response.json();
  return result.data.collections || [];
}

/**
 * 出欠確認を作成
 */
export async function createAttendanceCollection(
  data: CreateAttendanceRequest
): Promise<AttendanceCollection> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/attendance/collections`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`出欠確認の作成に失敗しました: ${text || response.statusText}`);
  }

  const result: ApiResponse<AttendanceCollection> = await response.json();
  return result.data;
}

/**
 * 出欠確認を取得
 */
export async function getAttendanceCollection(collectionId: string): Promise<AttendanceCollection> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(
    `${baseURL}/api/v1/attendance/collections/${collectionId}`,
    {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    }
  );

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`出欠確認の取得に失敗しました: ${text || response.statusText}`);
  }

  const result: ApiResponse<AttendanceCollection> = await response.json();
  return result.data;
}

/**
 * 出欠確認を締め切る
 */
export async function closeAttendanceCollection(collectionId: string): Promise<void> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(
    `${baseURL}/api/v1/attendance/collections/${collectionId}/close`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    }
  );

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`出欠確認の締切に失敗しました: ${text || response.statusText}`);
  }
}

/**
 * 出欠確認を削除
 * 成功時: 204 No Content（レスポンスボディなし）
 */
export async function deleteAttendanceCollection(collectionId: string): Promise<void> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(
    `${baseURL}/api/v1/attendance/collections/${collectionId}`,
    {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    }
  );

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`出欠確認の削除に失敗しました: ${text || response.statusText}`);
  }
}


/**
 * 出欠回答一覧を取得
 */
export async function getAttendanceResponses(
  collectionId: string
): Promise<AttendanceResponse[]> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(
    `${baseURL}/api/v1/attendance/collections/${collectionId}/responses`,
    {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    }
  );

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`回答一覧の取得に失敗しました: ${text || response.statusText}`);
  }

  const result: ApiResponse<{ collection_id: string; responses: AttendanceResponse[] }> =
    await response.json();
  return result.data.responses;
}
