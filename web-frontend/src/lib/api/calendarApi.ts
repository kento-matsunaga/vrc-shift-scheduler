import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

/**
 * Calendar の型定義
 */
export interface Calendar {
  id: string;
  tenant_id: string;
  title: string;
  description: string;
  is_public: boolean;
  public_token?: string;
  event_ids: string[];
  created_at: string;
  updated_at: string;
}

/**
 * Calendar 作成リクエストの型
 */
export interface CreateCalendarRequest {
  title: string;
  description: string;
  event_ids: string[];
}

/**
 * Calendar 更新リクエストの型
 */
export interface UpdateCalendarRequest {
  title: string;
  description: string;
  event_ids: string[];
  is_public: boolean;
}

/**
 * Calendar 一覧レスポンスの型
 */
export interface CalendarListResponse {
  calendars: Calendar[];
}

/**
 * Calendar 一覧取得
 */
export async function getCalendars(): Promise<CalendarListResponse> {
  const res = await apiClient.get<ApiResponse<CalendarListResponse>>('/api/v1/calendars');
  return res.data;
}

/**
 * Calendar 詳細取得
 */
export async function getCalendarById(id: string): Promise<Calendar> {
  const res = await apiClient.get<ApiResponse<Calendar>>(`/api/v1/calendars/${id}`);
  return res.data;
}

/**
 * Calendar 作成
 */
export async function createCalendar(data: CreateCalendarRequest): Promise<Calendar> {
  const res = await apiClient.post<ApiResponse<Calendar>>('/api/v1/calendars', data);
  return res.data;
}

/**
 * Calendar 更新
 */
export async function updateCalendar(id: string, data: UpdateCalendarRequest): Promise<Calendar> {
  const res = await apiClient.put<ApiResponse<Calendar>>(`/api/v1/calendars/${id}`, data);
  return res.data;
}

/**
 * Calendar 削除
 */
export async function deleteCalendar(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/calendars/${id}`);
}

/**
 * 公開カレンダーのURLを生成
 */
export function getPublicCalendarUrl(publicToken: string): string {
  return `${window.location.origin}/p/calendar/${publicToken}`;
}
