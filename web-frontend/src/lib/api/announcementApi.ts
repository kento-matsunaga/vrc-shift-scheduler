import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

export interface Announcement {
  id: string;
  tenant_id: string | null;
  title: string;
  body: string;
  published_at: string;
  created_at: string;
  is_read: boolean;
}

export interface AnnouncementListResponse {
  announcements: Announcement[];
}

export interface UnreadCountResponse {
  count: number;
}

/**
 * お知らせ一覧取得
 */
export async function getAnnouncements(): Promise<Announcement[]> {
  const res = await apiClient.get<ApiResponse<AnnouncementListResponse>>('/api/v1/announcements');
  return res.data.announcements;
}

/**
 * 未読件数取得
 */
export async function getUnreadCount(): Promise<number> {
  const res = await apiClient.get<ApiResponse<UnreadCountResponse>>('/api/v1/announcements/unread-count');
  return res.data.count;
}

/**
 * お知らせを既読にする
 */
export async function markAsRead(id: string): Promise<void> {
  await apiClient.post(`/api/v1/announcements/${id}/read`, {});
}

/**
 * すべてのお知らせを既読にする
 */
export async function markAllAsRead(): Promise<void> {
  await apiClient.post('/api/v1/announcements/read-all', {});
}
