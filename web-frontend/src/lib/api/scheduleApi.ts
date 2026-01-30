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
 * 日程調整更新リクエスト
 */
export interface UpdateScheduleRequest {
  title: string;
  description: string;
  candidates?: CandidateDate[];
  deadline?: string; // ISO 8601 format
  force_delete_candidate_responses?: boolean;
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
  const result = await apiClient.post<ApiResponse<Schedule>>(
    '/api/v1/schedules',
    data
  );
  return result.data;
}

/**
 * 日程調整を更新
 */
export async function updateSchedule(
  scheduleId: string,
  data: UpdateScheduleRequest
): Promise<Schedule> {
  const result = await apiClient.put<ApiResponse<Schedule>>(
    `/api/v1/schedules/${scheduleId}`,
    data
  );
  return result.data;
}

/**
 * 日程調整一覧レスポンス
 */
interface ListSchedulesResponse {
  schedules: Schedule[];
}

/**
 * 日程調整一覧を取得
 */
export async function listSchedules(): Promise<Schedule[]> {
  const result = await apiClient.get<ListSchedulesResponse>('/api/v1/schedules');
  return result.schedules || [];
}

/**
 * 日程調整を取得
 * NOTE: このAPIはレスポンス直下にScheduleオブジェクトを返す
 */
export async function getSchedule(scheduleId: string): Promise<Schedule> {
  return apiClient.get<Schedule>(`/api/v1/schedules/${scheduleId}`);
}

/**
 * 日程を決定
 */
export async function decideSchedule(scheduleId: string, decidedDate: string): Promise<void> {
  await apiClient.post(`/api/v1/schedules/${scheduleId}/decide`, { decided_date: decidedDate });
}

/**
 * 日程調整を締め切る
 */
export async function closeSchedule(scheduleId: string): Promise<void> {
  await apiClient.post(`/api/v1/schedules/${scheduleId}/close`, {});
}

/**
 * 日程調整を削除
 * 成功時: 204 No Content（レスポンスボディなし）
 */
export async function deleteSchedule(scheduleId: string): Promise<void> {
  await apiClient.delete(`/api/v1/schedules/${scheduleId}`);
}

/**
 * 日程回答一覧レスポンス
 */
interface GetScheduleResponsesResult {
  responses: ScheduleResponse[];
}

/**
 * 日程回答一覧を取得
 */
export async function getScheduleResponses(scheduleId: string): Promise<ScheduleResponse[]> {
  const result = await apiClient.get<GetScheduleResponsesResult>(
    `/api/v1/schedules/${scheduleId}/responses`
  );
  return result.responses || [];
}

/**
 * 出欠確認変換リクエスト
 */
export interface ConvertToAttendanceRequest {
  candidate_ids: string[];
  title?: string; // 省略時は元のタイトル
}

/**
 * 出欠確認変換レスポンス
 */
export interface ConvertToAttendanceResponse {
  collection_id: string;
  public_token: string;
  title: string;
}

/**
 * 日程調整を出欠確認に変換
 */
export async function convertToAttendance(
  scheduleId: string,
  data: ConvertToAttendanceRequest
): Promise<ConvertToAttendanceResponse> {
  const result = await apiClient.post<ApiResponse<ConvertToAttendanceResponse>>(
    `/api/v1/schedules/${scheduleId}/convert-to-attendance`,
    data
  );
  return result.data;
}
