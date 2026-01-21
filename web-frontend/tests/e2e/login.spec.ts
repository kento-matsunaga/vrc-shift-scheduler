import { test, expect } from '@playwright/test';

test.describe('Login Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    // Clear localStorage before each test
    await page.goto('/login');
    await page.evaluate(() => localStorage.clear());
  });

  test('should display login page with all elements', async ({ page }) => {
    await page.goto('/login');

    // Check page title/header
    await expect(page.locator('h1')).toContainText('VRC Shift Scheduler');
    await expect(page.locator('h2')).toContainText('ログイン');

    // Check for login form elements
    await expect(page.locator('input#email')).toBeVisible();
    await expect(page.locator('input#password')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toContainText('ログイン');

    // Check password reset link
    await expect(page.getByText('パスワードを忘れた場合')).toBeVisible();
  });

  test('should show validation error when email is empty', async ({ page }) => {
    await page.goto('/login');

    // Fill only password
    await page.fill('input#password', 'password123');

    // Submit button should be disabled when email is empty
    await expect(page.locator('button[type="submit"]')).toBeDisabled();
  });

  test('should show validation error when password is empty', async ({ page }) => {
    await page.goto('/login');

    // Fill only email
    await page.fill('input#email', 'admin1@example.com');

    // Submit button should be disabled when password is empty
    await expect(page.locator('button[type="submit"]')).toBeDisabled();
  });

  test('should show error on invalid credentials', async ({ page }) => {
    await page.goto('/login');

    // Fill in invalid credentials
    await page.fill('input#email', 'invalid@example.com');
    await page.fill('input#password', 'wrongpassword');

    // Submit the form
    await page.click('button[type="submit"]');

    // Wait for error message
    await expect(page.locator('.bg-red-50, [role="alert"]')).toBeVisible({
      timeout: 5000,
    });
    // Check Japanese error message
    await expect(page.locator('.text-red-600')).toContainText('メールアドレスまたはパスワードが正しくありません');
  });

  test('should show loading state during login', async ({ page }) => {
    await page.goto('/login');

    // Fill in credentials
    await page.fill('input#email', 'admin1@example.com');
    await page.fill('input#password', 'password123');

    // Click submit and check loading state
    await page.click('button[type="submit"]');

    // The button should show loading text (this might be brief)
    // We check that the button becomes disabled during loading
    await expect(page.locator('button[type="submit"]')).toBeDisabled();
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    await page.goto('/login');

    // Fill in valid credentials (seeded admin user)
    await page.fill('input#email', 'admin1@example.com');
    await page.fill('input#password', 'password123');

    // Submit the form
    await page.click('button[type="submit"]');

    // Should redirect to events page
    await expect(page).toHaveURL(/\/events/, { timeout: 10000 });
  });

  test('should navigate to password reset page', async ({ page }) => {
    await page.goto('/login');

    // Click password reset link
    await page.click('text=パスワードを忘れた場合');

    // Should navigate to password reset page
    await expect(page).toHaveURL(/\/reset-password/);
  });

  test('should persist login state after successful login', async ({ page }) => {
    await page.goto('/login');

    // Login
    await page.fill('input#email', 'admin1@example.com');
    await page.fill('input#password', 'password123');
    await page.click('button[type="submit"]');

    // Wait for redirect
    await expect(page).toHaveURL(/\/events/, { timeout: 10000 });

    // Verify localStorage has auth token
    const authToken = await page.evaluate(() => localStorage.getItem('auth_token'));
    expect(authToken).toBeTruthy();
  });
});

test.describe('Authentication Guard', () => {
  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    // Clear any existing auth
    await page.goto('/login');
    await page.evaluate(() => localStorage.clear());

    // Try to access protected route
    await page.goto('/events');

    // Should redirect to login
    await expect(page).toHaveURL(/\/login/, { timeout: 5000 });
  });

  test('should allow access to protected route with valid auth', async ({ page }) => {
    // First login
    await page.goto('/login');
    await page.fill('input#email', 'admin1@example.com');
    await page.fill('input#password', 'password123');
    await page.click('button[type="submit"]');

    // Wait for redirect to events
    await expect(page).toHaveURL(/\/events/, { timeout: 10000 });

    // Navigate to another protected route
    await page.goto('/members');

    // Should be able to access
    await expect(page).toHaveURL(/\/members/);
  });
});
