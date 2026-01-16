import { test, expect } from '@playwright/test';

test.describe('Login Page E2E', () => {
  test('should display login page', async ({ page }) => {
    await page.goto('/login');

    // Check for login form elements
    await expect(page.locator('input[type="email"], input[name="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
  });

  test('should show error on invalid login', async ({ page }) => {
    await page.goto('/login');

    // Fill in invalid credentials
    await page.fill('input[type="email"], input[name="email"]', 'invalid@example.com');
    await page.fill('input[type="password"]', 'wrongpassword');

    // Submit the form
    await page.click('button[type="submit"]');

    // Wait for error message (implementation-specific)
    // This is a placeholder - adjust based on actual UI
    await expect(page.locator('[role="alert"], .error-message, .text-red-500')).toBeVisible({
      timeout: 5000,
    });
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    await page.goto('/login');

    // Fill in valid credentials (seeded admin user)
    await page.fill('input[type="email"], input[name="email"]', 'admin1@example.com');
    await page.fill('input[type="password"]', 'password123');

    // Submit the form
    await page.click('button[type="submit"]');

    // Should redirect to dashboard or home page
    await expect(page).not.toHaveURL('/login', { timeout: 10000 });
  });
});
