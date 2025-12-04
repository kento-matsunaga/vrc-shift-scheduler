import { apiClient } from '../apiClient';
import type { ApiResponse, ShiftAssignment, ShiftAssignmentListResponse } from '../../types/api';

/**
 * シフト確定
 */
export async function confirmAssignment(data: {
  slot_id: string;
  member_id: string;
  note?: string;
}): Promise<ShiftAssignment> {
  const res = await apiClient.post<ApiResponse<ShiftAssignment>>('/api/v1/shift-assignments', data);
  return res.data;
}

/**
 * シフト割り当て一覧取得
 */
export async function getAssignments(params?: {
  event_id?: string;
  member_id?: string;
  business_day_id?: string;
  slot_id?: string;
  assignment_status?: 'confirmed' | 'cancelled';
  start_date?: string; // YYYY-MM-DD
  end_date?: string; // YYYY-MM-DD
}): Promise<ShiftAssignmentListResponse> {
  const res = await apiClient.get<ApiResponse<ShiftAssignmentListResponse>>(
    '/api/v1/shift-assignments',
    params
  );
  return res.data;
}

/**
 * シフト割り当て詳細取得
 */
export async function getAssignmentDetail(assignmentId: string): Promise<ShiftAssignment> {
  const res = await apiClient.get<ApiResponse<ShiftAssignment>>(
    `/api/v1/shift-assignments/${assignmentId}`
  );
  return res.data;
}

/**
 * シフト割り当てステータス変更（v1.1）
 */
export async function updateAssignmentStatus(
  assignmentId: string,
  data: {
    status: 'confirmed' | 'cancelled';
    reason?: string;
  }
): Promise<ShiftAssignment> {
  const res = await apiClient.patch<ApiResponse<ShiftAssignment>>(
    `/api/v1/shift-assignments/${assignmentId}/status`,
    data
  );
  return res.data;
}

/**
 * シフト割り当てキャンセル（v1.1）
 */
export async function cancelAssignment(assignmentId: string, _reason?: string): Promise<void> {
  await apiClient.delete(`/api/v1/shift-assignments/${assignmentId}`);
}

