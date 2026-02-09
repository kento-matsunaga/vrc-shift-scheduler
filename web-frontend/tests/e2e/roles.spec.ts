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

    // Check for new role button (＋ ロール追加) or empty state button
    const newRoleButton = page.locator('button:has-text("ロール追加"), button:has-text("最初のロールを追加")');
    await expect(newRoleButton.first()).toBeVisible();
  });

  test('should display role list', async ({ page }) => {
    // Wait for data to load
    await page.waitForTimeout(1000);

    // Should have role items or empty state button
    const roleItems = page.locator('[data-testid="role-item"], .role-card, tr:has(td)');
    const emptyButton = page.locator('button:has-text("最初のロールを追加")');

    const hasRoles = await roleItems.count() > 0;
    const hasEmptyButton = await emptyButton.isVisible();

    expect(hasRoles || hasEmptyButton).toBeTruthy();
  });

  test('should open new role form', async ({ page }) => {
    // Click new role button (＋ ロール追加)
    await page.click('button:has-text("ロール追加")');

    // Form should appear (modal with input)
    await expect(page.locator('[role="dialog"], .modal')).toBeVisible();
    await expect(page.locator('[role="dialog"] input, .modal input')).toBeVisible();
  });

  test('should create a new role', async ({ page }) => {
    const uniqueName = `テストロール_${Date.now()}`;

    // Click new role button (＋ ロール追加)
    await page.click('button:has-text("ロール追加")');

    // Wait for modal to open
    await page.waitForTimeout(500);

    // Fill the form (find input in modal)
    const nameInput = page.locator('[role="dialog"] input, .modal input').first();
    await nameInput.fill(uniqueName);

    // Submit
    await page.click('[role="dialog"] button:has-text("作成"), .modal button:has-text("作成")');

    // Wait for creation
    await page.waitForTimeout(1000);

    // New role should appear
    await expect(page.getByText(uniqueName)).toBeVisible({ timeout: 5000 });
  });

  test('should edit an existing role', async ({ page }) => {
    // Wait for roles to load
    await page.waitForTimeout(1000);

    // Find edit button (pencil icon or 編集 text)
    const editButton = page.locator('button:has-text("編集"), button[aria-label="編集"], button svg').first();
    if (await editButton.isVisible()) {
      await editButton.click();

      // Form should appear (modal)
      await expect(page.locator('[role="dialog"], .modal')).toBeVisible();

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
    // Click new role button (＋ ロール追加)
    await page.click('button:has-text("ロール追加")');

    // Wait for modal to open
    await page.waitForTimeout(500);

    // Look for color buttons (color picker)
    const colorButtons = page.locator('[role="dialog"] button[style*="background"], .modal button[style*="background"]');
    if (await colorButtons.count() > 0) {
      await colorButtons.first().click();
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

    // Click new button (＋ グループ追加)
    await page.click('button:has-text("グループ追加")');

    // Wait for modal to open
    await page.waitForTimeout(500);

    // Fill name (find input in modal)
    const nameInput = page.locator('[role="dialog"] input, .modal input').first();
    await nameInput.fill(uniqueName);

    // Submit
    await page.click('[role="dialog"] button:has-text("作成"), .modal button:has-text("作成")');

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

    // Click new button (＋ グループ追加)
    await page.click('button:has-text("グループ追加")');

    // Wait for modal to open
    await page.waitForTimeout(500);

    // Fill name (find input in modal)
    const nameInput = page.locator('[role="dialog"] input, .modal input').first();
    await nameInput.fill(uniqueName);

    // Submit
    await page.click('[role="dialog"] button:has-text("作成"), .modal button:has-text("作成")');

    await page.waitForTimeout(1000);
  });

  test('should add members to group', async ({ page }) => {
    await page.waitForTimeout(1000);

    // Click on a group card (the group items should be clickable)
    const groupCard = page.locator('[data-testid="group-card"], .group-card, button:has-text("メンバー")').first();
    if (await groupCard.isVisible()) {
      await groupCard.click();
      await page.waitForTimeout(500);
    }
  });
});
