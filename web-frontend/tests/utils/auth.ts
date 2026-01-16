import { APIRequestContext } from '@playwright/test';
import { ApiClient, ENDPOINTS, TEST_CREDENTIALS, ApiResponse } from './api-client';

/**
 * Login response from API
 */
export interface LoginResponse {
  token: string;
  admin_id: string;
  tenant_id: string;
  role: string;
  expires_at: string;
}

/**
 * Setup request
 */
export interface SetupRequest {
  organization_name: string;
  admin_name: string;
  password: string;
  timezone?: string;
}

/**
 * Setup response
 */
export interface SetupResponse {
  tenant_id: string;
  member_id: string;
  token: string;
  message: string;
  invite_url: string;
}

/**
 * Login with test credentials and return authenticated API client
 * @throws Error if login fails
 */
export async function loginAsAdmin(request: APIRequestContext): Promise<{
  client: ApiClient;
  loginData: LoginResponse;
}> {
  const client = new ApiClient(request);

  const response = await client.raw('POST', ENDPOINTS.login, TEST_CREDENTIALS);

  if (response.status() !== 200) {
    throw new Error(
      `Failed to login as admin. Status: ${response.status()}. ` +
        'The password may have been changed by a previous test. ' +
        'Try running: docker compose exec backend /app/seed'
    );
  }

  const result = (await response.json()) as ApiResponse<LoginResponse>;

  if (!result.data?.token) {
    throw new Error('Login response did not contain a token');
  }

  client.setToken(result.data.token);

  return {
    client,
    loginData: result.data,
  };
}

/**
 * Login with custom credentials
 */
export async function loginWithCredentials(
  request: APIRequestContext,
  email: string,
  password: string
): Promise<{
  client: ApiClient;
  loginData: LoginResponse;
}> {
  const client = new ApiClient(request);

  const response = await client.raw('POST', ENDPOINTS.login, { email, password });
  const result = (await response.json()) as ApiResponse<LoginResponse>;

  client.setToken(result.data.token);

  return {
    client,
    loginData: result.data,
  };
}

/**
 * Create a new tenant via setup endpoint (only works when no tenant exists)
 */
export async function setupNewTenant(
  request: APIRequestContext,
  setupData: SetupRequest
): Promise<{
  client: ApiClient;
  setupResponse: SetupResponse;
}> {
  const client = new ApiClient(request);

  const response = await client.raw('POST', ENDPOINTS.setup, setupData);
  const result = (await response.json()) as ApiResponse<SetupResponse>;

  client.setToken(result.data.token);

  return {
    client,
    setupResponse: result.data,
  };
}

/**
 * Get unauthenticated API client
 */
export function getUnauthenticatedClient(request: APIRequestContext): ApiClient {
  return new ApiClient(request);
}
