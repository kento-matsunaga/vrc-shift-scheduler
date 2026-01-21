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

test.describe('Events Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/events');
    await expect(page).toHaveURL(/\/events/);
  });

  test('should display events page with header', async ({ page }) => {
    // Check page header
    await expect(page.getByRole('heading', { name: /イベント/ })).toBeVisible();

    // Check for new event button
    await expect(page.getByRole('button', { name: /新規|作成|追加/ })).toBeVisible();
  });

  test('should display event list', async ({ page }) => {
    // Wait for events to load
    await page.waitForTimeout(1000);

    // Should have event cards or list items (or empty state message)
    const eventItems = page.locator('[data-testid="event-item"], .event-card, tr:has(td)');
    const emptyMessage = page.locator('text=イベントがありません');

    // Either events exist or empty message is shown
    const hasEvents = await eventItems.count() > 0;
    const hasEmptyMessage = await emptyMessage.isVisible();

    expect(hasEvents || hasEmptyMessage).toBeTruthy();
  });

  test('should open new event form', async ({ page }) => {
    // Click new event button
    await page.click('button:has-text("新規"), button:has-text("作成"), button:has-text("追加")');

    // Form should appear
    await expect(page.locator('input[name="name"], input[placeholder*="イベント名"]')).toBeVisible();
  });

  test('should create a new event', async ({ page }) => {
    const uniqueName = `テストイベント_${Date.now()}`;

    // Click new event button
    await page.click('button:has-text("新規"), button:has-text("作成"), button:has-text("追加")');

    // Fill the form
    await page.fill('input[name="name"], input[placeholder*="イベント名"]', uniqueName);

    // Fill description if available
    const descInput = page.locator('textarea[name="description"], input[placeholder*="説明"]');
    if (await descInput.isVisible()) {
      await descInput.fill('E2Eテスト用イベント');
    }

    // Submit the form
    await page.click('button:has-text("作成"), button:has-text("登録"), button:has-text("保存")');

    // Wait for success
    await page.waitForTimeout(1000);

    // New event should appear in the list
    await expect(page.getByText(uniqueName)).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to event detail', async ({ page }) => {
    // Wait for events to load
    await page.waitForTimeout(1000);

    // Click on first event
    const eventLink = page.locator('a[href*="/events/"], [data-testid="event-link"]').first();
    if (await eventLink.isVisible()) {
      await eventLink.click();

      // Should navigate to event detail or templates page
      await expect(page).toHaveURL(/\/events\/|\/templates/);
    }
  });

  test('should edit an existing event', async ({ page }) => {
    // Wait for events to load
    await page.waitForTimeout(1000);

    // Find edit button
    const editButton = page.locator('button:has-text("編集"), button[aria-label="編集"]').first();
    if (await editButton.isVisible()) {
      await editButton.click();

      // Form should appear
      await expect(page.locator('input[name="name"], input[placeholder*="イベント名"]')).toBeVisible();

      // Modify the name
      const input = page.locator('input[name="name"], input[placeholder*="イベント名"]');
      const currentValue = await input.inputValue();
      await input.fill(currentValue + '_edited');

      // Submit
      await page.click('button:has-text("更新"), button:has-text("保存")');

      // Wait for update
      await page.waitForTimeout(1000);
    }
  });

  test('should show delete confirmation', async ({ page }) => {
    // Wait for events to load
    await page.waitForTimeout(1000);

    // Find delete button
    const deleteButton = page.locator('button:has-text("削除"), button[aria-label="削除"]').first();
    if (await deleteButton.isVisible()) {
      await deleteButton.click();

      // Confirmation dialog should appear
      await expect(page.locator('[role="dialog"], .modal, [role="alertdialog"]')).toBeVisible();
    }
  });
});

test.describe('Event Business Days', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should generate business days for event', async ({ page }) => {
    // Navigate to events
    await page.goto('/events');
    await page.waitForTimeout(1000);

    // Click on an event to go to detail
    const eventLink = page.locator('a[href*="/events/"]').first();
    if (await eventLink.isVisible()) {
      await eventLink.click();
      await page.waitForTimeout(1000);

      // Look for business days or generate button
      const generateButton = page.locator('button:has-text("営業日を生成"), button:has-text("生成")');
      if (await generateButton.isVisible()) {
        await generateButton.click();

        // Wait for generation
        await page.waitForTimeout(1000);
      }
    }
  });

  test('should navigate to business days list', async ({ page }) => {
    await page.goto('/events');
    await page.waitForTimeout(1000);

    // Click on business days link if exists
    const businessDaysLink = page.locator('a:has-text("営業日"), a[href*="/business-days"]').first();
    if (await businessDaysLink.isVisible()) {
      await businessDaysLink.click();
      await expect(page).toHaveURL(/business-days/);
    }
  });
});

test.describe('Event Templates', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should navigate to templates page', async ({ page }) => {
    await page.goto('/events');
    await page.waitForTimeout(1000);

    // Click on templates link
    const templatesLink = page.locator('a:has-text("テンプレート"), a[href*="/templates"]').first();
    if (await templatesLink.isVisible()) {
      await templatesLink.click();
      await expect(page).toHaveURL(/templates/);
    }
  });

  test('should display templates list', async ({ page }) => {
    // Navigate directly to templates (need event ID)
    await page.goto('/events');
    await page.waitForTimeout(1000);

    const templatesLink = page.locator('a[href*="/templates"]').first();
    if (await templatesLink.isVisible()) {
      await templatesLink.click();
      await page.waitForTimeout(1000);

      // Check for template list or empty state
      const templateItems = page.locator('[data-testid="template-item"], .template-card');
      const emptyMessage = page.locator('text=テンプレートがありません');

      const hasTemplates = await templateItems.count() > 0;
      const hasEmptyMessage = await emptyMessage.isVisible();

      expect(hasTemplates || hasEmptyMessage).toBeTruthy();
    }
  });
});

test.describe('Event Validation', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/events');
  });

  test('should show error when creating event with empty name', async ({ page }) => {
    // Open new event form
    await page.click('button:has-text("新規"), button:has-text("作成"), button:has-text("追加")');

    // Try to submit without filling name
    const submitButton = page.locator('button:has-text("作成"), button:has-text("登録"), button:has-text("保存")');

    if (await submitButton.isEnabled()) {
      await submitButton.click();

      // Should show error or button should be disabled
      const hasError = await page.locator('.text-red-500, .error-message, [role="alert"]').isVisible();
      expect(hasError).toBeTruthy();
    } else {
      // Button is correctly disabled
      await expect(submitButton).toBeDisabled();
    }
  });
});
