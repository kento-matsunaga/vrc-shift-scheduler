import { apiClient } from '../apiClient';
import type { ApiResponse } from '../../types/api';

export interface Instance {
  instance_id: string;
  tenant_id: string;
  event_id: string;
  name: string;
  display_order: number;
  max_members: number | null;
  created_at: string;
  updated_at: string;
}

export interface CreateInstanceInput {
  name: string;
  display_order?: number;
  max_members?: number | null;
}

export interface UpdateInstanceInput {
  name?: string;
  display_order?: number;
  max_members?: number | null;
}

export interface InstanceListResponse {
  instances: Instance[];
  count: number;
}

/**
 * イベントのインスタンス一覧を取得
 */
export async function listInstances(eventId: string): Promise<Instance[]> {
  const response = await apiClient.get<ApiResponse<InstanceListResponse>>(`/api/v1/events/${eventId}/instances`);
  return response.data.instances;
}

/**
 * インスタンスを作成
 */
export async function createInstance(eventId: string, input: CreateInstanceInput): Promise<Instance> {
  const response = await apiClient.post<ApiResponse<Instance>>(`/api/v1/events/${eventId}/instances`, input);
  return response.data;
}

/**
 * インスタンスを取得
 */
export async function getInstance(instanceId: string): Promise<Instance> {
  const response = await apiClient.get<ApiResponse<Instance>>(`/api/v1/instances/${instanceId}`);
  return response.data;
}

/**
 * インスタンスを更新
 */
export async function updateInstance(instanceId: string, input: UpdateInstanceInput): Promise<Instance> {
  const response = await apiClient.put<ApiResponse<Instance>>(`/api/v1/instances/${instanceId}`, input);
  return response.data;
}

// ===========================================
// インスタンス管理用API（イベントレベル）
// ===========================================

export interface CheckInstanceDeletableResponse {
  can_delete: boolean;
  slot_count: number;
  assigned_slots: number;
  blocking_reason?: string;
}

/**
 * インスタンスが削除可能か確認（全営業日のシフト枠も含めてチェック）
 */
export async function checkInstanceDeletable(instanceId: string): Promise<CheckInstanceDeletableResponse> {
  const response = await apiClient.get<ApiResponse<CheckInstanceDeletableResponse>>(`/api/v1/instances/${instanceId}/deletable`);
  return response.data;
}

/**
 * インスタンスを削除（インスタンス自体と全営業日の紐づくシフト枠も削除）
 */
export async function deleteInstance(instanceId: string): Promise<void> {
  await apiClient.delete(`/api/v1/instances/${instanceId}`);
}

// ===========================================
// 営業日のシフト枠一括削除API（営業日レベル）
// ===========================================

export interface CheckSlotsDeletableResponse {
  can_delete: boolean;
  slot_count: number;
  assigned_slots: number;
  blocking_reason?: string;
}

/**
 * 営業日+インスタンスに紐づくシフト枠が削除可能か確認
 */
export async function checkSlotsByInstanceDeletable(businessDayId: string, instanceId: string): Promise<CheckSlotsDeletableResponse> {
  const response = await apiClient.get<ApiResponse<CheckSlotsDeletableResponse>>(`/api/v1/business-days/${businessDayId}/instances/${instanceId}/slots/deletable`);
  return response.data;
}

/**
 * 営業日+インスタンスに紐づくシフト枠を一括削除（インスタンス自体は削除されない）
 */
export async function deleteSlotsByInstance(businessDayId: string, instanceId: string): Promise<void> {
  await apiClient.delete(`/api/v1/business-days/${businessDayId}/instances/${instanceId}/slots`);
}
