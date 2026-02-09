import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for API integration tests and E2E tests
 * See https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './tests',

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
    baseURL: process.env.API_BASE_URL || 'http://backend:8080',

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
        baseURL: process.env.FRONTEND_URL || 'http://localhost:5173',
        // Docker Alpine環境ではシステムのChromiumを使用
        ...(process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH && {
          launchOptions: {
            executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH,
          },
        }),
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
