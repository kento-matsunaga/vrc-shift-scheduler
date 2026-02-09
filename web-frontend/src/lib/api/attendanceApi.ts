import type { ApiResponse } from '../../types/api';
import { apiClient } from '../apiClient';

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
 * 出欠確認更新リクエスト
 */
export interface UpdateAttendanceCollectionRequest {
  title: string;
  description: string;
  deadline?: string; // ISO 8601 format
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
 * 出欠確認一覧レスポンス
 */
interface ListAttendanceCollectionsResponse {
  collections: AttendanceCollection[];
}

/**
 * 出欠確認一覧を取得
 */
export async function listAttendanceCollections(): Promise<AttendanceCollection[]> {
  const result = await apiClient.get<ApiResponse<ListAttendanceCollectionsResponse>>(
    '/api/v1/attendance/collections'
  );
  return result.data.collections || [];
}

/**
 * 出欠確認を作成
 */
export async function createAttendanceCollection(
  data: CreateAttendanceRequest
): Promise<AttendanceCollection> {
  const result = await apiClient.post<ApiResponse<AttendanceCollection>>(
    '/api/v1/attendance/collections',
    data
  );
  return result.data;
}

/**
 * 出欠確認を更新
 */
export async function updateAttendanceCollection(
  collectionId: string,
  data: UpdateAttendanceCollectionRequest
): Promise<AttendanceCollection> {
  const result = await apiClient.put<ApiResponse<AttendanceCollection>>(
    `/api/v1/attendance/collections/${collectionId}`,
    data
  );
  return result.data;
}

/**
 * 出欠確認を取得
 */
export async function getAttendanceCollection(collectionId: string): Promise<AttendanceCollection> {
  const result = await apiClient.get<ApiResponse<AttendanceCollection>>(
    `/api/v1/attendance/collections/${collectionId}`
  );
  return result.data;
}

/**
 * 出欠確認を締め切る
 */
export async function closeAttendanceCollection(collectionId: string): Promise<void> {
  await apiClient.post(`/api/v1/attendance/collections/${collectionId}/close`, {});
}

/**
 * 出欠確認を削除
 * 成功時: 204 No Content（レスポンスボディなし）
 */
export async function deleteAttendanceCollection(collectionId: string): Promise<void> {
  await apiClient.delete(`/api/v1/attendance/collections/${collectionId}`);
}

/**
 * 出欠回答一覧レスポンス
 */
interface GetAttendanceResponsesResult {
  collection_id: string;
  responses: AttendanceResponse[];
}

/**
 * 出欠回答一覧を取得
 */
export async function getAttendanceResponses(
  collectionId: string
): Promise<AttendanceResponse[]> {
  const result = await apiClient.get<ApiResponse<GetAttendanceResponsesResult>>(
    `/api/v1/attendance/collections/${collectionId}/responses`
  );
  return result.data.responses;
}

/**
 * 出欠回答更新リクエスト（管理者用）
 */
export interface UpdateAttendanceResponseRequest {
  member_id: string;
  target_date_id: string;
  response: 'attending' | 'absent' | 'undecided';
  note?: string;
  available_from?: string; // HH:MM format
  available_to?: string;   // HH:MM format
}

/**
 * 出欠回答更新レスポンス（管理者用）
 */
export interface UpdateAttendanceResponseResult {
  response_id: string;
  collection_id: string;
  member_id: string;
  target_date_id: string;
  response: 'attending' | 'absent' | 'undecided';
  note: string;
  available_from?: string;
  available_to?: string;
  responded_at: string;
}

/**
 * 出欠回答を更新（管理者用・締め切り後も可能）
 */
export async function updateAttendanceResponse(
  collectionId: string,
  data: UpdateAttendanceResponseRequest
): Promise<UpdateAttendanceResponseResult> {
  const result = await apiClient.put<ApiResponse<UpdateAttendanceResponseResult>>(
    `/api/v1/attendance/collections/${collectionId}/responses`,
    data
  );
  return result.data;
}
