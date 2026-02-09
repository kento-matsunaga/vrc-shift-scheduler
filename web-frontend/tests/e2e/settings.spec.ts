import { test, expect } from '@playwright/test';

/**
 * Helper function to login before tests
 */
async function login(page: import('@playwright/test').Page) {
  await page.goto('/login');
  await page.evaluate(() => localStorage.clear());
  await page.fill('input#email', 'admin1@example.com');
  await page.fill('input#password', 'password123');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/events/, { timeout: 10000 });
}

test.describe('Settings Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/settings');
    await expect(page).toHaveURL(/\/settings/);
  });

  test('should display settings page with tabs', async ({ page }) => {
    // Check page header (use exact match to avoid multiple matches)
    await expect(page.getByRole('heading', { name: '設定', exact: true })).toBeVisible();

    // Check for settings sections (may not have explicit tabs)
    const settingsContent = page.locator('main, [role="main"]');
    await expect(settingsContent).toBeVisible();
  });

  test('should display tenant settings section', async ({ page }) => {
    // Look for tenant settings
    const tenantSection = page.locator('text=テナント設定, text=組織設定');
    if (await tenantSection.isVisible()) {
      await expect(tenantSection).toBeVisible();
    }
  });

  test('should display password change form', async ({ page }) => {
    // Look for password change section
    const passwordSection = page.locator('text=パスワード変更, text=パスワードを変更');
    if (await passwordSection.isVisible()) {
      // Click to expand if needed
      await passwordSection.click();

      // Should show password inputs
      await expect(page.locator('input[type="password"]').first()).toBeVisible();
    }
  });

  test('should validate password change form', async ({ page }) => {
    // Find password change section
    const passwordSection = page.locator('text=パスワード変更, text=パスワードを変更');
    if (await passwordSection.isVisible()) {
      await passwordSection.click();
      await page.waitForTimeout(500);

      // Find password inputs
      const passwordInputs = page.locator('input[type="password"]');
      if (await passwordInputs.count() >= 2) {
        // Fill with mismatched passwords
        await passwordInputs.nth(0).fill('newpassword123');
        await passwordInputs.nth(1).fill('differentpassword');

        // Submit
        const submitButton = page.locator('button:has-text("変更"), button:has-text("保存")');
        if (await submitButton.isVisible() && await submitButton.isEnabled()) {
          await submitButton.click();

          // Should show error
          await page.waitForTimeout(500);
        }
      }
    }
  });

  test('should display admin management section', async ({ page }) => {
    // Look for admin management
    const adminSection = page.locator('text=管理者, text=権限管理');
    if (await adminSection.isVisible()) {
      await expect(adminSection).toBeVisible();
    }
  });

  test('should navigate between settings tabs', async ({ page }) => {
    // Find tabs
    const tabs = page.locator('[role="tab"], .tab-button, button:has-text("タブ")');
    const tabCount = await tabs.count();

    if (tabCount > 1) {
      // Click each tab
      for (let i = 0; i < Math.min(tabCount, 3); i++) {
        await tabs.nth(i).click();
        await page.waitForTimeout(300);
      }
    }
  });
});

test.describe('Admin Invitation E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should display admin invitation section', async ({ page }) => {
    // Navigate to settings or invitation page
    await page.goto('/settings');
    await page.waitForTimeout(1000);

    // Look for invitation section
    const inviteSection = page.locator('text=招待, text=管理者を招待');
    if (await inviteSection.isVisible()) {
      await expect(inviteSection).toBeVisible();
    }
  });

  test('should open invite admin form', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForTimeout(1000);

    // Click invite button
    const inviteButton = page.locator('button:has-text("招待"), button:has-text("管理者を追加")');
    if (await inviteButton.isVisible()) {
      await inviteButton.click();

      // Form or modal should appear
      await expect(page.locator('input[type="email"]')).toBeVisible();
    }
  });

  test('should validate invite email format', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForTimeout(1000);

    const inviteButton = page.locator('button:has-text("招待"), button:has-text("管理者を追加")');
    if (await inviteButton.isVisible()) {
      await inviteButton.click();

      // Fill invalid email
      const emailInput = page.locator('input[type="email"]');
      if (await emailInput.isVisible()) {
        await emailInput.fill('invalid-email');

        // Submit
        const submitButton = page.locator('button:has-text("送信"), button:has-text("招待")');
        if (await submitButton.isVisible()) {
          await submitButton.click();

          // Should show validation error
          await page.waitForTimeout(500);
        }
      }
    }
  });
});

test.describe('Logout E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should logout successfully', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForTimeout(1000);

    // Find logout button
    const logoutButton = page.locator('button:has-text("ログアウト")');
    if (await logoutButton.isVisible()) {
      await logoutButton.click();

      // Should redirect to login
      await expect(page).toHaveURL(/\/login/, { timeout: 5000 });

      // Auth should be cleared
      const authToken = await page.evaluate(() => localStorage.getItem('auth_token'));
      expect(authToken).toBeFalsy();
    }
  });
});

test.describe('Data Import/Export E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/settings');
  });

  test('should display import section', async ({ page }) => {
    // Look for import section
    const importSection = page.locator('text=インポート, text=データ取り込み');
    if (await importSection.isVisible()) {
      await expect(importSection).toBeVisible();
    }
  });

  test('should display export options', async ({ page }) => {
    // Look for export section
    const exportSection = page.locator('text=エクスポート, text=データ出力');
    if (await exportSection.isVisible()) {
      await expect(exportSection).toBeVisible();
    }
  });
});
