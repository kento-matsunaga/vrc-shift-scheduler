import { apiClient } from '../apiClient';
import type { ApiResponse, Event, EventListResponse, GenerateBusinessDaysResponse } from '../../types/api';

/**
 * Event 作成リクエストの型
 */
export interface CreateEventRequest {
  event_name: string;
  event_type: 'normal' | 'special';
  description: string;
  recurrence_type?: 'none' | 'weekly' | 'biweekly';
  recurrence_start_date?: string; // YYYY-MM-DD
  recurrence_day_of_week?: number; // 0-6: 日曜=0, 土曜=6
  default_start_time?: string; // HH:MM:SS
  default_end_time?: string; // HH:MM:SS
}

/**
 * Event 作成
 */
export async function createEvent(data: CreateEventRequest): Promise<Event> {
  const res = await apiClient.post<ApiResponse<Event>>('/api/v1/events', data);
  return res.data;
}

/**
 * Event 一覧取得
 */
export async function getEvents(params?: {
  is_active?: boolean;
}): Promise<EventListResponse> {
  const res = await apiClient.get<ApiResponse<EventListResponse>>('/api/v1/events', params);
  return res.data;
}

/**
 * Event 詳細取得
 */
export async function getEventDetail(eventId: string): Promise<Event> {
  const res = await apiClient.get<ApiResponse<Event>>(`/api/v1/events/${eventId}`);
  return res.data;
}

/**
 * Event 更新（v1.1）
 */
export async function updateEvent(
  eventId: string,
  data: {
    event_name?: string;
    description?: string;
  }
): Promise<Event> {
  const res = await apiClient.put<ApiResponse<Event>>(`/api/v1/events/${eventId}`, data);
  return res.data;
}

/**
 * Event 削除（v1.1）
 */
export async function deleteEvent(eventId: string): Promise<void> {
  await apiClient.delete(`/api/v1/events/${eventId}`);
}

/**
 * 営業日を自動生成（定期イベント用）
 * 今月〜来月末までの営業日を生成する
 */
export async function generateBusinessDays(eventId: string): Promise<GenerateBusinessDaysResponse> {
  const res = await apiClient.post<ApiResponse<GenerateBusinessDaysResponse>>(
    `/api/v1/events/${eventId}/generate-business-days`,
    {}
  );
  return res.data;
}

