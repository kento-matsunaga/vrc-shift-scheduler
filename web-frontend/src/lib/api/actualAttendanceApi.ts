import { apiClient } from '../apiClient';
import type { ApiResponse, RecentAttendanceResponse } from '../../types/api';

/**
 * 本出席データ取得（実際のシフト割り当て実績）
 *
 * これは出欠確認（予定）ではなく、実際にシフトに割り当てられた実績データ
 * シフト割り当てあり → "attended" (○)
 * シフト割り当てなし → "absent" (×)
 */
export async function getActualAttendance(params?: {
  limit?: number;
}): Promise<RecentAttendanceResponse> {
  const res = await apiClient.get<ApiResponse<RecentAttendanceResponse>>('/api/v1/actual-attendance', params);
  return res.data;
}
