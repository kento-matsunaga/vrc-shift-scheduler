import { apiClient } from '../apiClient';
import type { ApiResponse, BusinessDay, BusinessDayListResponse } from '../../types/api';

/**
 * 営業日作成（手動）
 */
export async function createBusinessDay(
  eventId: string,
  data: {
    target_date: string; // YYYY-MM-DD
    start_time: string; // HH:MM
    end_time: string; // HH:MM
    occurrence_type: 'recurring' | 'special';
    template_id?: string; // optional: テンプレートから自動的にシフト枠を作成
  }
): Promise<BusinessDay> {
  const res = await apiClient.post<ApiResponse<BusinessDay>>(
    `/api/v1/events/${eventId}/business-days`,
    data
  );
  return res.data;
}

/**
 * 営業日一覧取得
 */
export async function getBusinessDays(
  eventId: string,
  params?: {
    start_date?: string; // YYYY-MM-DD
    end_date?: string; // YYYY-MM-DD
  }
): Promise<BusinessDayListResponse> {
  const res = await apiClient.get<ApiResponse<BusinessDayListResponse>>(
    `/api/v1/events/${eventId}/business-days`,
    params
  );
  return res.data;
}

/**
 * 営業日詳細取得
 */
export async function getBusinessDayDetail(businessDayId: string): Promise<BusinessDay> {
  const res = await apiClient.get<ApiResponse<BusinessDay>>(`/api/v1/business-days/${businessDayId}`);
  return res.data;
}

/**
 * 営業日のアクティブ状態変更（v1.1）
 */
export async function updateBusinessDayStatus(
  businessDayId: string,
  isActive: boolean
): Promise<BusinessDay> {
  const res = await apiClient.patch<ApiResponse<BusinessDay>>(
    `/api/v1/business-days/${businessDayId}`,
    { is_active: isActive }
  );
  return res.data;
}

/**
 * 既存営業日にテンプレートを適用してシフト枠を作成
 */
export async function applyTemplateToBusinessDay(
  businessDayId: string,
  templateId: string
): Promise<void> {
  await apiClient.post<ApiResponse<{ message: string; business_day_id: string; template_id: string; items_count: number }>>(
    `/api/v1/business-days/${businessDayId}/apply-template`,
    { template_id: templateId }
  );
}

