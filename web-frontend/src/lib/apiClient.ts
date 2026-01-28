import type { ApiError } from '../types/api';

/**
 * API クライアントクラス
 * 全ての API リクエストに共通のヘッダー・エラーハンドリングを適用
 */
export class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = import.meta.env.VITE_API_BASE_URL || '') {
    this.baseURL = baseURL;
  }

  /**
   * localStorage から認証情報を取得してヘッダーに追加
   * JWT優先、フォールバックでX-Tenant-IDヘッダー
   */
  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    // JWT トークンがあれば Authorization ヘッダーを付与（優先）
    const authToken = localStorage.getItem('auth_token');
    if (authToken) {
      headers['Authorization'] = `Bearer ${authToken}`;
      return headers;
    }

    // JWT がなければ従来の X-Tenant-ID ヘッダー（フォールバック）
    const tenantId = localStorage.getItem('tenant_id') || import.meta.env.VITE_TENANT_ID;
    const memberId = localStorage.getItem('member_id');

    if (tenantId) {
      headers['X-Tenant-ID'] = tenantId;
    }

    if (memberId) {
      headers['X-Member-ID'] = memberId;
    }

    return headers;
  }

  /**
   * 汎用 HTTP リクエストメソッド
   */
  async request<T>(
    method: string,
    path: string,
    body?: unknown,
    queryParams?: Record<string, string | number | boolean | undefined>
  ): Promise<T> {
    // クエリパラメータをURLに追加
    let url = `${this.baseURL}${path}`;
    if (queryParams) {
      const params = new URLSearchParams();
      Object.entries(queryParams).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          params.append(key, String(value));
        }
      });
      const paramString = params.toString();
      if (paramString) {
        url += `?${paramString}`;
      }
    }

    try {
      const res = await fetch(url, {
        method,
        headers: this.getHeaders(),
        body: body ? JSON.stringify(body) : undefined,
      });

      if (!res.ok) {
        // エラーレスポンスを解析
        const errorData: ApiError = await res.json().catch(() => ({
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

      // 204 No Content の場合は null を返す
      if (res.status === 204) {
        return null as T;
      }

      return await res.json();
    } catch (error) {
      // ApiClientError はそのまま throw
      if (error instanceof ApiClientError) {
        throw error;
      }

      // ネットワークエラーなどは ApiClientError でラップ
      throw new ApiClientError(
        error instanceof Error ? error.message : 'Network error',
        0,
        'ERR_NETWORK'
      );
    }
  }

  /**
   * GET リクエスト
   */
  async get<T>(path: string, queryParams?: Record<string, string | number | boolean | undefined>): Promise<T> {
    return this.request<T>('GET', path, undefined, queryParams);
  }

  /**
   * POST リクエスト
   */
  async post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('POST', path, body);
  }

  /**
   * PUT リクエスト
   */
  async put<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('PUT', path, body);
  }

  /**
   * PATCH リクエスト
   */
  async patch<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('PATCH', path, body);
  }

  /**
   * DELETE リクエスト
   */
  async delete<T>(path: string): Promise<T> {
    return this.request<T>('DELETE', path);
  }
}

/**
 * API エラークラス
 */
export class ApiClientError extends Error {
  public statusCode: number;
  public errorCode: string;
  public details?: Record<string, unknown>;

  constructor(
    message: string,
    statusCode: number,
    errorCode: string,
    details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'ApiClientError';
    this.statusCode = statusCode;
    this.errorCode = errorCode;
    this.details = details;
  }

  /**
   * エラーがバリデーションエラーかどうか
   */
  isValidationError(): boolean {
    return this.errorCode === 'ERR_INVALID_REQUEST';
  }

  /**
   * エラーが Not Found かどうか
   */
  isNotFoundError(): boolean {
    return this.errorCode === 'ERR_NOT_FOUND';
  }

  /**
   * エラーが競合エラーかどうか
   */
  isConflictError(): boolean {
    return this.errorCode === 'ERR_CONFLICT';
  }

  /**
   * エラーが権限エラーかどうか
   */
  isForbiddenError(): boolean {
    return this.errorCode === 'ERR_FORBIDDEN';
  }

  /**
   * ユーザーに表示するメッセージを取得
   */
  getUserMessage(): string {
    // ERR_INVALID_REQUEST またはERR_CONFLICT の場合はAPIからの具体的なメッセージを優先表示
    // （日本語メッセージが返されている場合があるため）
    if ( (this.errorCode === 'ERR_INVALID_REQUEST' || this.errorCode === 'ERR_CONFLICT') && this.message) {
      // APIから返されたメッセージが英語の場合のみ汎用メッセージを使う
      const isEnglishOnly = /^[a-zA-Z0-9\s_\-().,:]+$/.test(this.message);
      if (!isEnglishOnly) {
        return this.message;
      }
    }

    // 日本語メッセージマッピング
    const messageMap: Record<string, string> = {
      ERR_INVALID_REQUEST: '入力内容に誤りがあります',
      ERR_NOT_FOUND: '指定されたデータが見つかりません',
      ERR_CONFLICT: '競合が発生しました。再度お試しください',
      ERR_FORBIDDEN: 'この操作を実行する権限がありません',
      ERR_INTERNAL: 'サーバーエラーが発生しました',
      ERR_NETWORK: 'ネットワークエラーが発生しました',
      ERR_SLOT_FULL: 'このシフト枠は既に満員です',
    };

    return messageMap[this.errorCode] || this.message;
  }
}

// シングルトンインスタンス
export const apiClient = new ApiClient();

