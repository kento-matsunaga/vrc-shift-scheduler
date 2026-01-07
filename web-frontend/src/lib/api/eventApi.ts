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

/**
 * イベントグループ割り当ての型
 */
export interface EventGroupAssignments {
  member_group_ids: string[];
  role_group_ids: string[];
}

/**
 * イベントのグループ割り当て取得
 */
export async function getEventGroupAssignments(eventId: string): Promise<EventGroupAssignments> {
  const res = await apiClient.get<ApiResponse<EventGroupAssignments>>(`/api/v1/events/${eventId}/groups`);
  return res.data;
}

/**
 * イベントのグループ割り当て更新
 */
export async function updateEventGroupAssignments(
  eventId: string,
  data: EventGroupAssignments
): Promise<EventGroupAssignments> {
  const res = await apiClient.put<ApiResponse<EventGroupAssignments>>(`/api/v1/events/${eventId}/groups`, data);
  return res.data;
}

/**
 * 営業日の型
 */
export interface BusinessDay {
  business_day_id: string;
  tenant_id: string;
  event_id: string;
  target_date: string; // YYYY-MM-DD
  start_time: string; // HH:MM:SS
  end_time: string; // HH:MM:SS
  occurrence_type: 'recurring' | 'special';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * イベントの営業日一覧を取得
 */
export async function getEventBusinessDays(
  eventId: string,
  params?: {
    start_date?: string;
    end_date?: string;
    is_active?: boolean;
  }
): Promise<BusinessDay[]> {
  const res = await apiClient.get<ApiResponse<{ business_days: BusinessDay[]; count: number }>>(
    `/api/v1/events/${eventId}/business-days`,
    params
  );
  return res.data.business_days || [];
}

