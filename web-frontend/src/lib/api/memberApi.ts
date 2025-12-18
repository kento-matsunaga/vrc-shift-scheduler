import { apiClient } from '../apiClient';
import type { ApiResponse, Member, MemberListResponse, RecentAttendanceResponse } from '../../types/api';

/**
 * メンバー作成
 */
export async function createMember(data: {
  display_name: string;
  discord_user_id?: string;
  email?: string;
}): Promise<Member> {
  const res = await apiClient.post<ApiResponse<Member>>('/api/v1/members', data);
  return res.data;
}

/**
 * メンバー一覧取得
 */
export async function getMembers(params?: {
  is_active?: boolean;
}): Promise<MemberListResponse> {
  const res = await apiClient.get<ApiResponse<MemberListResponse>>('/api/v1/members', params);
  return res.data;
}

/**
 * メンバー詳細取得
 */
export async function getMemberDetail(memberId: string): Promise<Member> {
  const res = await apiClient.get<ApiResponse<Member>>(`/api/v1/members/${memberId}`);
  return res.data;
}

/**
 * メンバー更新
 */
export async function updateMember(memberId: string, data: {
  display_name: string;
  discord_user_id?: string;
  email?: string;
  is_active: boolean;
  role_ids?: string[];
}): Promise<Member> {
  const res = await apiClient.put<ApiResponse<Member>>(`/api/v1/members/${memberId}`, data);
  return res.data;
}

/**
 * 直近の出欠状況を取得
 */
export async function getRecentAttendance(params?: {
  limit?: number;
}): Promise<RecentAttendanceResponse> {
  const res = await apiClient.get<ApiResponse<RecentAttendanceResponse>>('/api/v1/members/recent-attendance', params);
  return res.data;
}

