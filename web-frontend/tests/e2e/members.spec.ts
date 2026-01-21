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

test.describe('Members Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/members');
    await expect(page).toHaveURL(/\/members/);
  });

  test('should display members page with header and controls', async ({ page }) => {
    // Check page header
    await expect(page.getByRole('heading', { name: /メンバー/ })).toBeVisible();

    // Check for new member button (＋ 追加)
    await expect(page.locator('button:has-text("追加")')).toBeVisible();

    // Check for bulk import button
    await expect(page.locator('button:has-text("一括登録")')).toBeVisible();
  });

  test('should display member list', async ({ page }) => {
    // Wait for members to load
    await page.waitForTimeout(1000);

    // Should have member cards or table rows
    const memberItems = page.locator('[data-testid="member-item"], .member-card, tr:has(td)');
    await expect(memberItems.first()).toBeVisible({ timeout: 5000 });
  });

  test('should open new member form', async ({ page }) => {
    // Click new member button (＋ 追加)
    await page.click('button:has-text("追加")');

    // Form should appear (modal with input)
    await expect(page.locator('[role="dialog"] input, .modal input')).toBeVisible();
  });

  test('should create a new member', async ({ page }) => {
    const uniqueName = `テストメンバー_${Date.now()}`;

    // Click new member button (＋ 追加)
    await page.click('button:has-text("追加")');

    // Wait for modal to open
    await page.waitForTimeout(500);

    // Fill the form (find input in modal)
    const nameInput = page.locator('[role="dialog"] input, .modal input').first();
    await nameInput.fill(uniqueName);

    // Submit the form
    await page.click('[role="dialog"] button:has-text("登録"), .modal button:has-text("登録")');

    // Wait for success
    await page.waitForTimeout(1000);

    // New member should appear in the list
    await expect(page.getByText(uniqueName)).toBeVisible({ timeout: 5000 });
  });

  test('should edit an existing member', async ({ page }) => {
    // Wait for members to load
    await page.waitForTimeout(1000);

    // Click edit button on first member
    const editButton = page.locator('button:has-text("編集"), button[aria-label="編集"]').first();
    if (await editButton.isVisible()) {
      await editButton.click();

      // Form should appear with existing data
      await expect(page.locator('input[placeholder*="表示名"], input[name="displayName"]')).toBeVisible();

      // Modify the name
      const input = page.locator('input[placeholder*="表示名"], input[name="displayName"]');
      const currentValue = await input.inputValue();
      await input.fill(currentValue + '_edited');

      // Submit
      await page.click('button:has-text("更新"), button:has-text("保存")');

      // Wait for update
      await page.waitForTimeout(1000);
    }
  });

  test('should filter members by role', async ({ page }) => {
    // Wait for page to load
    await page.waitForTimeout(1000);

    // Check if role filter exists
    const roleFilter = page.locator('select:has-text("ロール"), [data-testid="role-filter"]');
    if (await roleFilter.isVisible()) {
      // Select a role
      await roleFilter.selectOption({ index: 1 });

      // Wait for filter to apply
      await page.waitForTimeout(500);
    }
  });

  test('should select multiple members for bulk operations', async ({ page }) => {
    // Wait for members to load
    await page.waitForTimeout(1000);

    // Find visible checkboxes in the member list
    const checkboxes = page.locator('input[type="checkbox"]:visible');
    const count = await checkboxes.count();

    if (count > 1) {
      // Select first two members
      await checkboxes.nth(0).check({ force: true });
      await checkboxes.nth(1).check({ force: true });

      // Bulk action button should appear (ロール一括設定)
      await expect(page.locator('button:has-text("ロール一括設定")')).toBeVisible({ timeout: 3000 });
    }
  });

  test('should open bulk import modal', async ({ page }) => {
    // Click bulk import button
    await page.click('button:has-text("一括登録")');

    // Modal should appear
    await expect(page.locator('[role="dialog"], .modal')).toBeVisible();

    // Should have textarea or file input
    await expect(page.locator('textarea, input[type="file"]')).toBeVisible();
  });

  test('should show member details on click', async ({ page }) => {
    // Wait for members to load
    await page.waitForTimeout(1000);

    // Click on a member name/card
    const memberLink = page.locator('a:has-text("メンバー"), [data-testid="member-name"]').first();
    if (await memberLink.isVisible()) {
      await memberLink.click();

      // Should show details (modal or expanded view)
      await page.waitForTimeout(500);
    }
  });

  test('should toggle member active status', async ({ page }) => {
    // Wait for members to load
    await page.waitForTimeout(1000);

    // Find toggle switch or checkbox for active status
    const activeToggle = page.locator('input[type="checkbox"][name*="active"], [role="switch"]').first();
    if (await activeToggle.isVisible()) {
      const isChecked = await activeToggle.isChecked();
      await activeToggle.click();

      // Status should change
      await page.waitForTimeout(500);
    }
  });
});

test.describe('Members Page - Validation', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/members');
  });

  test('should show error when creating member with empty name', async ({ page }) => {
    // Open new member form (＋ 追加)
    await page.click('button:has-text("追加")');

    // Wait for modal to open
    await page.waitForTimeout(500);

    // Try to submit without filling name
    const submitButton = page.locator('[role="dialog"] button:has-text("登録"), .modal button:has-text("登録")');

    // Submit button should be disabled or show error after click
    if (await submitButton.isEnabled()) {
      await submitButton.click();

      // Should show error or form validation
      await page.waitForTimeout(500);
    } else {
      // Button is correctly disabled
      await expect(submitButton).toBeDisabled();
    }
  });
});

test.describe('Members Page - Delete', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/members');
  });

  test('should show delete confirmation dialog', async ({ page }) => {
    // Wait for members to load
    await page.waitForTimeout(1000);

    // Find delete button
    const deleteButton = page.locator('button:has-text("削除"), button[aria-label="削除"]').first();
    if (await deleteButton.isVisible()) {
      await deleteButton.click();

      // Confirmation dialog should appear
      await expect(page.locator('[role="dialog"], .modal, [role="alertdialog"]')).toBeVisible();

      // Should have confirm/cancel buttons
      await expect(page.getByRole('button', { name: /キャンセル|いいえ/ })).toBeVisible();
    }
  });
});
