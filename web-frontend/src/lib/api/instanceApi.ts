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

/**
 * インスタンスを削除
 */
export async function deleteInstance(instanceId: string): Promise<void> {
  await apiClient.delete(`/api/v1/instances/${instanceId}`);
}
