// 管理API クライアント
// NOTE: Cloudflare Access による認証が必要
// ローカル開発時は CF_ACCESS_TEAM_DOMAIN が未設定の場合、認証をスキップします

const API_BASE = '/api/v1/admin';

interface ApiResponse<T> {
  data: T;
}

class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  // ローカル開発用: X-Admin-Email ヘッダーを追加
  // NOTE: 本番環境では Cloudflare Access が CF-Access-JWT-Assertion を付与
  if (import.meta.env.DEV) {
    headers['X-Admin-Email'] = 'admin@example.com';
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new ApiError(
      response.status,
      error.error?.code || 'UNKNOWN',
      error.error?.message || 'Unknown error'
    );
  }

  // Backend already wraps response in { data: ... }
  return await response.json();
}

// ============================================================
// ライセンスキー API
// ============================================================

export interface LicenseKey {
  key_id: string;
  status: 'unused' | 'used' | 'revoked';
  expires_at: string | null;
  claimed_at: string | null;
  claimed_by: string | null;
  memo: string;
  created_at: string;
}

export interface GeneratedKey {
  key_id: string;
  key: string;
  expires_at: string | null;
  created_at: string;
}

export interface GenerateLicenseKeysInput {
  count: number;
  memo?: string;
  expires_at?: string;
}

export interface GenerateLicenseKeysOutput {
  keys: GeneratedKey[];
}

export interface ListLicenseKeysOutput {
  keys: LicenseKey[];
  total_count: number;
}

export async function generateLicenseKeys(
  input: GenerateLicenseKeysInput
): Promise<ApiResponse<GenerateLicenseKeysOutput>> {
  return request('POST', '/license-keys', input);
}

export async function listLicenseKeys(params?: {
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<ApiResponse<ListLicenseKeysOutput>> {
  const query = new URLSearchParams();
  if (params?.status) query.set('status', params.status);
  if (params?.limit) query.set('limit', String(params.limit));
  if (params?.offset) query.set('offset', String(params.offset));
  const queryString = query.toString();
  return request('GET', `/license-keys${queryString ? '?' + queryString : ''}`);
}

export async function revokeLicenseKey(keyId: string): Promise<void> {
  await request('PATCH', `/license-keys/${keyId}`, { action: 'revoke' });
}

// ============================================================
// テナント API
// ============================================================

export interface TenantListItem {
  tenant_id: string;
  tenant_name: string;
  status: 'active' | 'grace' | 'suspended';
  grace_until: string | null;
  created_at: string;
}

export interface TenantDetail {
  tenant_id: string;
  tenant_name: string;
  status: 'active' | 'grace' | 'suspended';
  grace_until: string | null;
  created_at: string;
  entitlements: Entitlement[];
  admins: Admin[];
}

export interface Entitlement {
  entitlement_id: string;
  plan_code: string;
  started_at: string;
  expires_at: string | null;
  revoked_at: string | null;
}

export interface Admin {
  admin_id: string;
  email: string;
  display_name: string;
  role: string;
}

export interface ListTenantsOutput {
  tenants: TenantListItem[];
  total_count: number;
}

export async function listTenants(params?: {
  status?: string;
  search?: string;
  limit?: number;
  offset?: number;
}): Promise<ApiResponse<ListTenantsOutput>> {
  const query = new URLSearchParams();
  if (params?.status) query.set('status', params.status);
  if (params?.search) query.set('search', params.search);
  if (params?.limit) query.set('limit', String(params.limit));
  if (params?.offset) query.set('offset', String(params.offset));
  const queryString = query.toString();
  return request('GET', `/tenants${queryString ? '?' + queryString : ''}`);
}

export async function getTenantDetail(
  tenantId: string
): Promise<ApiResponse<TenantDetail>> {
  return request('GET', `/tenants/${tenantId}`);
}

export async function updateTenantStatus(
  tenantId: string,
  input: { status: string; grace_until?: string }
): Promise<void> {
  await request('PATCH', `/tenants/${tenantId}/status`, input);
}

// ============================================================
// 監査ログ API
// ============================================================

export interface AuditLogItem {
  log_id: string;
  actor_type: string;
  actor_id: string | null;
  action: string;
  target_type: string | null;
  target_id: string | null;
  before_json: string | null;
  after_json: string | null;
  created_at: string;
}

export interface ListAuditLogsOutput {
  logs: AuditLogItem[];
  total_count: number;
}

export async function listAuditLogs(params?: {
  actor_type?: string;
  action?: string;
  limit?: number;
  offset?: number;
}): Promise<ApiResponse<ListAuditLogsOutput>> {
  const query = new URLSearchParams();
  if (params?.actor_type) query.set('actor_type', params.actor_type);
  if (params?.action) query.set('action', params.action);
  if (params?.limit) query.set('limit', String(params.limit));
  if (params?.offset) query.set('offset', String(params.offset));
  const queryString = query.toString();
  return request('GET', `/audit-logs${queryString ? '?' + queryString : ''}`);
}

// ============================================================
// パスワードリセット許可 API
// ============================================================

export interface AllowPasswordResetOutput {
  target_admin_id: string;
  target_email: string;
  tenant_id: string;
  allowed_at: string;
  expires_at: string;
  message: string;
}

export async function allowPasswordReset(
  adminId: string
): Promise<ApiResponse<AllowPasswordResetOutput>> {
  return request('POST', `/admins/${adminId}/allow-password-reset`);
}
