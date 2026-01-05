import { apiClient } from '../apiClient';
import type { ApiResponse, Member, MemberListResponse, RecentAttendanceResponse } from '../../types/api';

/**
 * メンバー作成
 */
export async function createMember(data: {
  display_name: string;
  discord_user_id?: string;
  email?: string;
  role_ids?: string[];
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

/**
 * メンバー削除（ソフトデリート）
 */
export async function deleteMember(memberId: string): Promise<void> {
  await apiClient.delete(`/api/v1/members/${memberId}`);
}

/**
 * 一括登録の結果
 */
export interface BulkImportResult {
  display_name: string;
  success: boolean;
  member_id?: string;
  error?: string;
}

/**
 * 一括登録のレスポンス
 */
export interface BulkImportResponse {
  total_count: number;
  success_count: number;
  failed_count: number;
  results: BulkImportResult[];
}

/**
 * 一括登録のメンバー入力
 */
export interface BulkImportMemberInput {
  display_name: string;
  role_ids?: string[];
}

/**
 * メンバー一括登録
 */
export async function bulkImportMembers(members: BulkImportMemberInput[]): Promise<BulkImportResponse> {
  const res = await apiClient.post<ApiResponse<BulkImportResponse>>('/api/v1/members/bulk-import', { members });
  return res.data;
}

/**
 * 失敗詳細
 */
export interface FailureDetail {
  member_id: string;
  reason: string;
}

/**
 * ロール一括更新のレスポンス
 */
export interface BulkUpdateRolesResponse {
  total_count: number;
  success_count: number;
  failed_count: number;
  failures?: FailureDetail[];
}

/**
 * メンバーのロール一括更新
 */
export async function bulkUpdateRoles(data: {
  member_ids: string[];
  add_role_ids?: string[];
  remove_role_ids?: string[];
}): Promise<BulkUpdateRolesResponse> {
  const res = await apiClient.post<ApiResponse<BulkUpdateRolesResponse>>('/api/v1/members/bulk-update-roles', data);
  return res.data;
}

