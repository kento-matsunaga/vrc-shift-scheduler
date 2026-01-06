import { apiClient } from '../apiClient';
import type { ApiResponse, ShiftSlot, ShiftSlotListResponse } from '../../types/api';

/**
 * シフト枠作成
 */
export async function createShiftSlot(
  businessDayId: string,
  data: {
    slot_name: string;
    instance_name: string;
    start_time: string; // HH:MM
    end_time: string; // HH:MM
    required_count: number;
    priority?: number;
  }
): Promise<ShiftSlot> {
  const res = await apiClient.post<ApiResponse<ShiftSlot>>(
    `/api/v1/business-days/${businessDayId}/shift-slots`,
    data
  );
  return res.data;
}

/**
 * シフト枠一覧取得（営業日ごと）
 */
export async function getShiftSlots(businessDayId: string): Promise<ShiftSlotListResponse> {
  const res = await apiClient.get<ApiResponse<ShiftSlotListResponse>>(
    `/api/v1/business-days/${businessDayId}/shift-slots`
  );
  return res.data;
}

/**
 * シフト枠詳細取得
 */
export async function getShiftSlotDetail(slotId: string): Promise<ShiftSlot> {
  const res = await apiClient.get<ApiResponse<ShiftSlot>>(`/api/v1/shift-slots/${slotId}`);
  return res.data;
}

/**
 * シフト枠更新（v1.1）
 */
export async function updateShiftSlot(
  slotId: string,
  data: {
    slot_name?: string;
    start_time?: string;
    end_time?: string;
    required_count?: number;
    priority?: number;
  }
): Promise<ShiftSlot> {
  const res = await apiClient.put<ApiResponse<ShiftSlot>>(`/api/v1/shift-slots/${slotId}`, data);
  return res.data;
}

/**
 * シフト枠削除（v1.1）
 */
export async function deleteShiftSlot(slotId: string): Promise<void> {
  await apiClient.delete(`/api/v1/shift-slots/${slotId}`);
}

