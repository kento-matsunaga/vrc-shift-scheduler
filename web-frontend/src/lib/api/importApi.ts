/**
 * インポートAPI クライアント
 * CSVファイルのアップロードとインポートジョブ管理
 */

import { ApiClientError } from '../apiClient';

// ========================
// Types
// ========================

export type ImportStatus = 'pending' | 'processing' | 'completed' | 'failed';
export type ImportType = 'members';

export interface ImportError {
  row: number;
  message: string;
}

export interface ImportMembersOptions {
  skipExisting?: boolean;
  updateExisting?: boolean;
  fuzzyMatch?: boolean;
}

export interface ImportMembersResponse {
  import_job_id: string;
  status: ImportStatus;
  total_rows: number;
  success_count: number;
  error_count: number;
  errors?: ImportError[];
}

export interface ImportStatusResponse {
  import_job_id: string;
  status: ImportStatus;
  import_type: ImportType;
  file_name: string;
  total_rows: number;
  processed_rows: number;
  success_count: number;
  error_count: number;
  progress: number;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface ImportResultResponse {
  import_job_id: string;
  status: ImportStatus;
  total_rows: number;
  success_count: number;
  error_count: number;
  skipped_count: number;
  errors?: ImportError[];
}

export interface ImportJobListResponse {
  jobs: ImportStatusResponse[];
  total_count: number;
}

// ========================
// Helper Functions
// ========================

/**
 * 認証ヘッダーを取得
 */
function getAuthHeaders(): HeadersInit {
  const headers: HeadersInit = {};

  const authToken = localStorage.getItem('auth_token');
  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`;
  }

  return headers;
}

/**
 * APIベースURLを取得
 */
function getBaseURL(): string {
  return import.meta.env.VITE_API_BASE_URL || '';
}

/**
 * レスポンスのエラーハンドリング
 */
async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({
      error: {
        code: 'ERR_UNKNOWN',
        message: `HTTP ${res.status}: ${res.statusText}`,
      },
    }));

    throw new ApiClientError(
      errorData.error.message,
      res.status,
      errorData.error.code,
      errorData.error.details
    );
  }

  if (res.status === 204) {
    return null as T;
  }

  const json = await res.json();
  return json.data !== undefined ? json.data : json;
}

// ========================
// API Functions
// ========================

/**
 * メンバーをCSVからインポート
 * @param file CSVファイル
 * @param options インポートオプション
 */
export async function importMembersFromCSV(
  file: File,
  options: ImportMembersOptions = {}
): Promise<ImportMembersResponse> {
  const formData = new FormData();
  formData.append('file', file);

  if (options.skipExisting) {
    formData.append('skip_existing', 'true');
  }
  if (options.updateExisting) {
    formData.append('update_existing', 'true');
  }
  if (options.fuzzyMatch) {
    formData.append('fuzzy_match', 'true');
  }

  const res = await fetch(`${getBaseURL()}/api/v1/imports/members`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: formData,
  });

  return handleResponse<ImportMembersResponse>(res);
}

/**
 * インポートジョブ一覧を取得
 * @param limit 取得件数（デフォルト: 20）
 * @param offset オフセット（デフォルト: 0）
 */
export async function getImportJobs(
  limit: number = 20,
  offset: number = 0
): Promise<ImportJobListResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });

  const res = await fetch(`${getBaseURL()}/api/v1/imports?${params}`, {
    method: 'GET',
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'application/json',
    },
  });

  return handleResponse<ImportJobListResponse>(res);
}

/**
 * インポートジョブのステータスを取得
 * @param importJobId インポートジョブID
 */
export async function getImportStatus(
  importJobId: string
): Promise<ImportStatusResponse> {
  const res = await fetch(`${getBaseURL()}/api/v1/imports/${importJobId}/status`, {
    method: 'GET',
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'application/json',
    },
  });

  return handleResponse<ImportStatusResponse>(res);
}

/**
 * インポートジョブの結果詳細を取得
 * @param importJobId インポートジョブID
 */
export async function getImportResult(
  importJobId: string
): Promise<ImportResultResponse> {
  const res = await fetch(`${getBaseURL()}/api/v1/imports/${importJobId}/result`, {
    method: 'GET',
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'application/json',
    },
  });

  return handleResponse<ImportResultResponse>(res);
}

/**
 * CSVテンプレートをダウンロード
 */
export function downloadCSVTemplate(): void {
  const csvContent = 'name,display_name,note\n佐藤太郎,たろう,一期生\n鈴木花子,はなこ,二期生\n';
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = 'members_template.csv';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  URL.revokeObjectURL(url);
}
