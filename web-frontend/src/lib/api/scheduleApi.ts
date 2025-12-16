import type { ApiResponse } from '../../types/api';

/**
 * 候補日
 */
export interface CandidateDate {
  date: string; // ISO 8601 format
  start_time?: string; // ISO 8601 format (optional)
  end_time?: string; // ISO 8601 format (optional)
}

/**
 * 日程調整作成リクエスト
 */
export interface CreateScheduleRequest {
  title: string;
  description: string;
  candidates: CandidateDate[];
  deadline?: string; // ISO 8601 format
}

/**
 * 日程調整レスポンス
 */
export interface Schedule {
  schedule_id: string;
  tenant_id: string;
  title: string;
  description?: string;
  public_token: string;
  status: 'open' | 'decided' | 'closed';
  deadline?: string;
  created_at: string;
  updated_at?: string;
}

/**
 * 日程回答
 */
export interface ScheduleResponse {
  response_id: string;
  member_id: string;
  available_dates: string[];
  note: string;
  responded_at: string;
}

/**
 * 日程調整を作成
 */
export async function createSchedule(data: CreateScheduleRequest): Promise<Schedule> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/schedules`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`日程調整の作成に失敗しました: ${text || response.statusText}`);
  }

  const result: ApiResponse<Schedule> = await response.json();
  return result.data;
}

/**
 * 日程調整を取得
 */
export async function getSchedule(scheduleId: string): Promise<Schedule> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/schedules/${scheduleId}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`日程調整の取得に失敗しました: ${text || response.statusText}`);
  }

  const result: ApiResponse<Schedule> = await response.json();
  return result.data;
}

/**
 * 日程を決定
 */
export async function decideSchedule(scheduleId: string, decidedDate: string): Promise<void> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/schedules/${scheduleId}/decide`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ decided_date: decidedDate }),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`日程決定に失敗しました: ${text || response.statusText}`);
  }
}

/**
 * 日程調整を締め切る
 */
export async function closeSchedule(scheduleId: string): Promise<void> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/schedules/${scheduleId}/close`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`日程調整の締切に失敗しました: ${text || response.statusText}`);
  }
}

/**
 * 日程回答一覧を取得
 */
export async function getScheduleResponses(scheduleId: string): Promise<ScheduleResponse[]> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/schedules/${scheduleId}/responses`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`回答一覧の取得に失敗しました: ${text || response.statusText}`);
  }

  const result: ApiResponse<{ schedule_id: string; responses: ScheduleResponse[] }> =
    await response.json();
  return result.data.responses;
}
