import { defineConfig, devices } from '@playwright/test';

// ============================================================
// 環境変数のバリデーション
// ============================================================
const API_BASE_URL = process.env.API_BASE_URL || 'http://backend:8080';
const FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:5173';

// テスト実行前に環境変数を検証（警告のみ、テストは続行）
if (!process.env.API_BASE_URL) {
  console.warn('⚠️  API_BASE_URL is not set. Using default:', API_BASE_URL);
}
if (!process.env.FRONTEND_URL && process.env.npm_lifecycle_script?.includes('e2e')) {
  console.warn('⚠️  FRONTEND_URL is not set. Using default:', FRONTEND_URL);
}

/**
 * Playwright configuration for API integration tests and E2E tests
 * See https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './tests',

  // ============================================================
  // グローバルタイムアウト設定
  // ============================================================
  // 各テストのタイムアウト（30秒）
  timeout: 30000,

  // アサーション（expect）のタイムアウト（5秒）
  expect: {
    timeout: 5000,
  },

  // Run tests in files in parallel
  fullyParallel: true,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Opt out of parallel tests on CI
  workers: process.env.CI ? 1 : undefined,

  // Reporter to use
  reporter: [
    ['list'],
    ['html', { open: 'never' }],
  ],

  // Shared settings for all projects
  use: {
    // Base URL for API tests
    // Docker環境内では 'backend' コンテナ名を使用、ローカルでは localhost
    baseURL: API_BASE_URL,

    // Collect trace when retrying the failed test
    trace: 'on-first-retry',

    // Extra HTTP headers for API tests
    extraHTTPHeaders: {
      'Content-Type': 'application/json',
    },
  },

  // Configure projects for API and E2E tests
  projects: [
    // API Integration Tests - no browser needed
    {
      name: 'api',
      testDir: './tests/api',
      use: {
        // No browser for API tests
      },
    },

    // E2E Tests - browser-based
    {
      name: 'e2e',
      testDir: './tests/e2e',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: FRONTEND_URL,
      },
    },
  ],

  // Run backend and frontend before starting the tests (optional)
  // Uncomment if you want Playwright to start the servers automatically
  // webServer: [
  //   {
  //     command: 'cd ../backend && go run ./cmd/server',
  //     url: 'http://localhost:8080/health',
  //     reuseExistingServer: !process.env.CI,
  //     timeout: 120 * 1000,
  //   },
  //   {
  //     command: 'npm run dev',
  //     url: 'http://localhost:5173',
  //     reuseExistingServer: !process.env.CI,
  //     timeout: 60 * 1000,
  //   },
  // ],
});
