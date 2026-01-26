import type { ApiError } from '../types/api';

/**
 * Error code to user-friendly message mapping
 * Backend error codes are mapped to localized messages with additional context
 */
const ERROR_MESSAGES: Record<string, string> = {
  // Authentication / Authorization errors
  ERR_UNAUTHORIZED: '認証が必要です。再度ログインしてください。',
  ERR_FORBIDDEN: 'この操作を行う権限がありません。',
  ERR_TOKEN_EXPIRED: 'セッションが期限切れです。再度ログインしてください。',

  // Validation errors
  ERR_VALIDATION: '入力内容に誤りがあります。',
  ERR_EMAIL_EXISTS: 'このメールアドレスは既に登録されています。',
  ERR_INVALID_EMAIL: 'メールアドレスの形式が正しくありません。',
  ERR_WEAK_PASSWORD: 'パスワードが弱すぎます。8文字以上で、大文字・小文字・数字を含めてください。',

  // Rate limiting
  ERR_RATE_LIMITED: 'リクエストが多すぎます。しばらく待ってから再度お試しください。',

  // Not found
  ERR_NOT_FOUND: '指定されたリソースが見つかりません。',

  // Payment / Stripe errors
  ERR_STRIPE: '決済処理中にエラーが発生しました。',
  ERR_STRIPE_CHECKOUT: 'チェックアウトセッションの作成に失敗しました。',
  ERR_STRIPE_CUSTOMER: '顧客情報の作成に失敗しました。',
  ERR_PAYMENT_FAILED: '決済に失敗しました。カード情報を確認してください。',

  // Subscription errors
  ERR_SUBSCRIPTION_NOT_FOUND: 'サブスクリプションが見つかりません。',
  ERR_SUBSCRIPTION_EXPIRED: 'サブスクリプションが期限切れです。',

  // Server errors
  ERR_INTERNAL: 'サーバーエラーが発生しました。しばらく待ってから再度お試しください。',
  ERR_SERVICE_UNAVAILABLE: 'サービスが一時的に利用できません。',

  // Signature / Webhook errors
  ERR_MISSING_SIGNATURE: '署名が見つかりません。',
  ERR_INVALID_SIGNATURE: '署名が無効です。',
};

/**
 * Checks if the response is an API error
 */
export function isApiError(data: unknown): data is ApiError {
  return (
    typeof data === 'object' &&
    data !== null &&
    'error' in data &&
    typeof (data as ApiError).error === 'object' &&
    'code' in (data as ApiError).error
  );
}

/**
 * Gets a user-friendly error message from an API error response
 *
 * @param error - The API error response or unknown error
 * @param fallback - Fallback message if no specific message is found
 * @returns User-friendly error message
 */
export function getErrorMessage(error: unknown, fallback = 'エラーが発生しました。'): string {
  // Handle ApiError type
  if (isApiError(error)) {
    const code = error.error.code;
    // First, check if we have a mapped message for this code
    if (ERROR_MESSAGES[code]) {
      return ERROR_MESSAGES[code];
    }
    // Otherwise, use the message from the API (which is already in Japanese)
    if (error.error.message) {
      return error.error.message;
    }
  }

  // Handle standard Error objects
  if (error instanceof Error) {
    return error.message;
  }

  // Handle string errors
  if (typeof error === 'string') {
    return error;
  }

  return fallback;
}

/**
 * Extracts the error code from an API error response
 */
export function getErrorCode(error: unknown): string | null {
  if (isApiError(error)) {
    return error.error.code;
  }
  return null;
}

/**
 * Checks if the error is a specific error code
 */
export function isErrorCode(error: unknown, code: string): boolean {
  return getErrorCode(error) === code;
}

/**
 * Checks if the error is a rate limit error
 */
export function isRateLimitError(error: unknown): boolean {
  return isErrorCode(error, 'ERR_RATE_LIMITED');
}

/**
 * Checks if the error is an authentication error
 */
export function isAuthError(error: unknown): boolean {
  const code = getErrorCode(error);
  return code === 'ERR_UNAUTHORIZED' || code === 'ERR_TOKEN_EXPIRED';
}

/**
 * Checks if the error is a validation error
 */
export function isValidationError(error: unknown): boolean {
  const code = getErrorCode(error);
  return code?.startsWith('ERR_VALIDATION') || code === 'ERR_EMAIL_EXISTS' || code === 'ERR_INVALID_EMAIL';
}
