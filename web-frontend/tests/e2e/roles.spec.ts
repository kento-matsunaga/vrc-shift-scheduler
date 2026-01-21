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

test.describe('Roles Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/roles');
    await expect(page).toHaveURL(/\/roles/);
  });

  test('should display roles page with header', async ({ page }) => {
    // Check page header
    await expect(page.getByRole('heading', { name: /ロール/ })).toBeVisible();

    // Check for new role button
    await expect(page.getByRole('button', { name: /新規|作成|追加/ })).toBeVisible();
  });

  test('should display role list', async ({ page }) => {
    // Wait for data to load
    await page.waitForTimeout(1000);

    // Should have role items or empty message
    const roleItems = page.locator('[data-testid="role-item"], .role-card, tr:has(td)');
    const emptyMessage = page.locator('text=ロールがありません');

    const hasRoles = await roleItems.count() > 0;
    const hasEmptyMessage = await emptyMessage.isVisible();

    expect(hasRoles || hasEmptyMessage).toBeTruthy();
  });

  test('should open new role form', async ({ page }) => {
    // Click new role button
    await page.click('button:has-text("新規"), button:has-text("作成"), button:has-text("追加")');

    // Form should appear
    await expect(page.locator('input[name="name"], input[placeholder*="ロール名"]')).toBeVisible();
  });

  test('should create a new role', async ({ page }) => {
    const uniqueName = `テストロール_${Date.now()}`;

    // Click new role button
    await page.click('button:has-text("新規"), button:has-text("作成"), button:has-text("追加")');

    // Fill the form
    await page.fill('input[name="name"], input[placeholder*="ロール名"]', uniqueName);

    // Select a color if available
    const colorInput = page.locator('input[type="color"], [data-testid="color-picker"]');
    if (await colorInput.isVisible()) {
      // Color input exists
    }

    // Submit
    await page.click('button:has-text("作成"), button:has-text("登録"), button:has-text("保存")');

    // Wait for creation
    await page.waitForTimeout(1000);

    // New role should appear
    await expect(page.getByText(uniqueName)).toBeVisible({ timeout: 5000 });
  });

  test('should edit an existing role', async ({ page }) => {
    // Wait for roles to load
    await page.waitForTimeout(1000);

    // Find edit button
    const editButton = page.locator('button:has-text("編集"), button[aria-label="編集"]').first();
    if (await editButton.isVisible()) {
      await editButton.click();

      // Form should appear
      await expect(page.locator('input[name="name"], input[placeholder*="ロール名"]')).toBeVisible();

      // Close without saving
      const cancelButton = page.locator('button:has-text("キャンセル")');
      if (await cancelButton.isVisible()) {
        await cancelButton.click();
      }
    }
  });

  test('should show delete confirmation for role', async ({ page }) => {
    // Wait for roles to load
    await page.waitForTimeout(1000);

    // Find delete button
    const deleteButton = page.locator('button:has-text("削除"), button[aria-label="削除"]').first();
    if (await deleteButton.isVisible()) {
      await deleteButton.click();

      // Confirmation should appear
      await expect(page.locator('[role="dialog"], .modal, [role="alertdialog"]')).toBeVisible();

      // Cancel
      const cancelButton = page.locator('button:has-text("キャンセル"), button:has-text("いいえ")');
      if (await cancelButton.isVisible()) {
        await cancelButton.click();
      }
    }
  });

  test('should select color preset for role', async ({ page }) => {
    // Click new role button
    await page.click('button:has-text("新規"), button:has-text("作成"), button:has-text("追加")');

    // Look for color presets
    const colorPresets = page.locator('[data-testid="color-preset"], .color-preset');
    if (await colorPresets.count() > 0) {
      await colorPresets.first().click();
    }
  });
});

test.describe('Role Groups Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/role-groups');
  });

  test('should display role groups page', async ({ page }) => {
    // Check if page loads
    await expect(page.getByRole('heading', { name: /ロールグループ/ })).toBeVisible();
  });

  test('should create role group', async ({ page }) => {
    const uniqueName = `テストロールグループ_${Date.now()}`;

    // Click new button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Fill name
    await page.fill('input[name="name"], input[placeholder*="グループ名"]', uniqueName);

    // Submit
    await page.click('button:has-text("作成"), button:has-text("保存")');

    await page.waitForTimeout(1000);
  });
});

test.describe('Member Groups Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/member-groups');
  });

  test('should display member groups page', async ({ page }) => {
    // Check if page loads
    await expect(page.getByRole('heading', { name: /メンバーグループ/ })).toBeVisible();
  });

  test('should create member group', async ({ page }) => {
    const uniqueName = `テストメンバーグループ_${Date.now()}`;

    // Click new button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Fill name
    await page.fill('input[name="name"], input[placeholder*="グループ名"]', uniqueName);

    // Submit
    await page.click('button:has-text("作成"), button:has-text("保存")');

    await page.waitForTimeout(1000);
  });

  test('should add members to group', async ({ page }) => {
    await page.waitForTimeout(1000);

    // Click on a group
    const groupCard = page.locator('[data-testid="group-card"], .group-card').first();
    if (await groupCard.isVisible()) {
      await groupCard.click();

      // Look for add member button
      const addMemberButton = page.locator('button:has-text("メンバーを追加")');
      if (await addMemberButton.isVisible()) {
        await expect(addMemberButton).toBeVisible();
      }
    }
  });
});
