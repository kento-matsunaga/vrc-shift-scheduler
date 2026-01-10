import type { ApiResponse } from '../../types/api';
import { apiClient } from '../apiClient';

/**
 * 候補日
 */
export interface CandidateDate {
  candidate_id?: string; // サーバーから取得時に含まれる
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
  group_ids?: string[]; // optional: target member group IDs
}

/**
 * 日程調整レスポンス
 */
export interface Schedule {
  schedule_id: string;
  tenant_id: string;
  title: string;
  description?: string;
  event_id?: string;
  public_token: string;
  status: 'open' | 'decided' | 'closed';
  deadline?: string;
  decided_candidate_id?: string;
  candidate_count?: number;
  response_count?: number;
  candidates?: CandidateDate[];
  group_ids?: string[]; // 対象グループIDs
  created_at: string;
  updated_at?: string;
}

/**
 * 日程回答
 */
export interface ScheduleResponse {
  response_id: string;
  member_id: string;
  candidate_id: string;
  availability: 'available' | 'maybe' | 'unavailable';
  note: string;
  responded_at: string;
}

/**
 * 日程調整を作成
 */
export async function createSchedule(data: CreateScheduleRequest): Promise<Schedule> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
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
 * 日程調整一覧を取得
 */
export async function listSchedules(): Promise<Schedule[]> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  const token = localStorage.getItem('auth_token');

  if (!token) {
    throw new Error('認証が必要です。ログインしてください。');
  }

  const response = await fetch(`${baseURL}/api/v1/schedules`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`日程調整一覧の取得に失敗しました: ${text || response.statusText}`);
  }

  const result = await response.json();
  return result.schedules || [];
}

/**
 * 日程調整を取得
 */
export async function getSchedule(scheduleId: string): Promise<Schedule> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
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

  const result = await response.json();
  return result;
}

/**
 * 日程を決定
 */
export async function decideSchedule(scheduleId: string, decidedDate: string): Promise<void> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
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
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
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
 * 日程調整を削除
 * 成功時: 204 No Content（レスポンスボディなし）
 */
export async function deleteSchedule(scheduleId: string): Promise<void> {
  await apiClient.delete(`/api/v1/schedules/${scheduleId}`);
}

/**
 * 日程回答一覧を取得
 */
export async function getScheduleResponses(scheduleId: string): Promise<ScheduleResponse[]> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
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

  const result = await response.json();
  return result.responses || [];
}
