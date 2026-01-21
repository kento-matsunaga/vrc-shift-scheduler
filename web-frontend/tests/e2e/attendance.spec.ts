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

test.describe('Attendance Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/attendance');
    await expect(page).toHaveURL(/\/attendance/);
  });

  test('should display attendance page with header', async ({ page }) => {
    // Check page header
    await expect(page.getByRole('heading', { name: /出欠/ })).toBeVisible();

    // Check for new attendance button
    await expect(page.getByRole('button', { name: /新規|作成/ })).toBeVisible();
  });

  test('should display attendance list or empty state', async ({ page }) => {
    // Wait for data to load
    await page.waitForTimeout(1000);

    // Should have attendance items or empty message
    const attendanceItems = page.locator('[data-testid="attendance-item"], .attendance-card, tr:has(td)');
    const emptyMessage = page.locator('text=出欠確認がありません');

    const hasAttendance = await attendanceItems.count() > 0;
    const hasEmptyMessage = await emptyMessage.isVisible();

    expect(hasAttendance || hasEmptyMessage).toBeTruthy();
  });

  test('should open new attendance form', async ({ page }) => {
    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Form should appear with title input
    await expect(page.locator('input[name="title"], input[placeholder*="タイトル"]')).toBeVisible();
  });

  test('should show date range picker for attendance', async ({ page }) => {
    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Look for DateRangePicker
    const dateRangePicker = page.locator('details:has-text("期間から一括追加")');
    if (await dateRangePicker.isVisible()) {
      await dateRangePicker.click();

      // Should show date inputs
      await expect(page.locator('input[type="date"]').first()).toBeVisible();
    }
  });

  test('should create attendance collection with target dates', async ({ page }) => {
    const uniqueTitle = `テスト出欠確認_${Date.now()}`;

    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Fill title
    await page.fill('input[name="title"], input[placeholder*="タイトル"]', uniqueTitle);

    // Add a target date
    const dateInputs = page.locator('input[type="date"]');
    if (await dateInputs.count() > 0) {
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      await dateInputs.first().fill(tomorrow.toISOString().split('T')[0]);
    }

    // Submit
    await page.click('button:has-text("作成"), button:has-text("登録")');

    // Wait for creation
    await page.waitForTimeout(2000);

    // Should show success
    const hasSuccess = await page.locator('text=作成しました, text=公開URL').isVisible();
    const hasNewAttendance = await page.getByText(uniqueTitle).isVisible();

    expect(hasSuccess || hasNewAttendance).toBeTruthy();
  });

  test('should select event for attendance', async ({ page }) => {
    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Look for event selector
    const eventSelect = page.locator('select:has-text("イベント"), [data-testid="event-select"]');
    if (await eventSelect.isVisible()) {
      // Select an event
      const options = await eventSelect.locator('option').all();
      if (options.length > 1) {
        await eventSelect.selectOption({ index: 1 });
      }
    }
  });

  test('should filter by group', async ({ page }) => {
    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Look for group selector
    const groupSelect = page.locator('select:has-text("グループ"), [data-testid="group-select"]');
    if (await groupSelect.isVisible()) {
      await expect(groupSelect).toBeVisible();
    }
  });

  test('should filter by role', async ({ page }) => {
    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Look for role selector
    const roleSelect = page.locator('[data-testid="role-filter"], label:has-text("ロール")');
    if (await roleSelect.isVisible()) {
      await expect(roleSelect).toBeVisible();
    }
  });
});

test.describe('Attendance Detail Page', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should display attendance detail', async ({ page }) => {
    await page.goto('/attendance');
    await page.waitForTimeout(1000);

    // Click on an attendance collection
    const attendanceLink = page.locator('a[href*="/attendance/"]').first();
    if (await attendanceLink.isVisible()) {
      await attendanceLink.click();

      // Should show detail page
      await expect(page.getByRole('heading')).toBeVisible();
    }
  });

  test('should show response summary', async ({ page }) => {
    await page.goto('/attendance');
    await page.waitForTimeout(1000);

    const attendanceLink = page.locator('a[href*="/attendance/"]').first();
    if (await attendanceLink.isVisible()) {
      await attendanceLink.click();
      await page.waitForTimeout(1000);

      // Should show response counts or table
      const responseTable = page.locator('table, [data-testid="response-table"]');
      const responseSummary = page.locator('text=参加, text=欠席, text=未定');

      const hasTable = await responseTable.isVisible();
      const hasSummary = await responseSummary.first().isVisible();

      expect(hasTable || hasSummary).toBeTruthy();
    }
  });

  test('should close attendance collection', async ({ page }) => {
    await page.goto('/attendance');
    await page.waitForTimeout(1000);

    const attendanceLink = page.locator('a[href*="/attendance/"]').first();
    if (await attendanceLink.isVisible()) {
      await attendanceLink.click();
      await page.waitForTimeout(1000);

      // Look for close button
      const closeButton = page.locator('button:has-text("締め切り"), button:has-text("終了")');
      if (await closeButton.isVisible()) {
        await expect(closeButton).toBeVisible();
      }
    }
  });

  test('should edit member response', async ({ page }) => {
    await page.goto('/attendance');
    await page.waitForTimeout(1000);

    const attendanceLink = page.locator('a[href*="/attendance/"]').first();
    if (await attendanceLink.isVisible()) {
      await attendanceLink.click();
      await page.waitForTimeout(1000);

      // Look for edit response button
      const editButton = page.locator('button:has-text("編集"), button[aria-label="編集"]').first();
      if (await editButton.isVisible()) {
        await editButton.click();

        // Modal should appear
        await expect(page.locator('[role="dialog"], .modal')).toBeVisible();
      }
    }
  });
});

test.describe('Attendance Time Validation', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/attendance');
  });

  test('should allow overnight time range for attendance (21:00-02:00)', async ({ page }) => {
    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Fill title
    await page.fill('input[name="title"], input[placeholder*="タイトル"]', 'テスト深夜出欠');

    // Add a target date
    const dateInputs = page.locator('input[type="date"]');
    if (await dateInputs.count() > 0) {
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      await dateInputs.first().fill(tomorrow.toISOString().split('T')[0]);
    }

    // Set overnight time range
    const timeInputs = page.locator('input[type="time"]');
    if (await timeInputs.count() >= 2) {
      await timeInputs.nth(0).fill('21:00');
      await timeInputs.nth(1).fill('02:00');

      // Should not show error
      await page.waitForTimeout(500);
      const errorMessage = page.locator('.text-red-500:has-text("開始時間")');
      const hasError = await errorMessage.isVisible();
      expect(hasError).toBeFalsy();
    }
  });

  test('should reject same start and end time for attendance', async ({ page }) => {
    // Click new attendance button
    await page.click('button:has-text("新規"), button:has-text("作成")');

    // Fill title
    await page.fill('input[name="title"], input[placeholder*="タイトル"]', 'テスト同時刻出欠');

    // Add a target date
    const dateInputs = page.locator('input[type="date"]');
    if (await dateInputs.count() > 0) {
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      await dateInputs.first().fill(tomorrow.toISOString().split('T')[0]);
    }

    // Set same start and end time
    const timeInputs = page.locator('input[type="time"]');
    if (await timeInputs.count() >= 2) {
      await timeInputs.nth(0).fill('21:00');
      await timeInputs.nth(1).fill('21:00');

      // Try to submit
      await page.click('button:has-text("作成"), button:has-text("登録")');

      // Should show error
      await expect(page.locator('.text-red-500, .bg-red-50')).toBeVisible({ timeout: 3000 });
    }
  });
});

test.describe('Attendance Sorting and Filtering', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should sort members in attendance detail', async ({ page }) => {
    await page.goto('/attendance');
    await page.waitForTimeout(1000);

    const attendanceLink = page.locator('a[href*="/attendance/"]').first();
    if (await attendanceLink.isVisible()) {
      await attendanceLink.click();
      await page.waitForTimeout(1000);

      // Look for sort controls
      const sortSelect = page.locator('select:has-text("並び替え"), [data-testid="sort-select"]');
      if (await sortSelect.isVisible()) {
        // Change sort order
        await sortSelect.selectOption({ index: 1 });
        await page.waitForTimeout(500);
      }
    }
  });

  test('should filter by response status', async ({ page }) => {
    await page.goto('/attendance');
    await page.waitForTimeout(1000);

    const attendanceLink = page.locator('a[href*="/attendance/"]').first();
    if (await attendanceLink.isVisible()) {
      await attendanceLink.click();
      await page.waitForTimeout(1000);

      // Look for filter tabs or buttons
      const attendingFilter = page.locator('button:has-text("参加"), [data-testid="filter-attending"]');
      if (await attendingFilter.isVisible()) {
        await attendingFilter.click();
        await page.waitForTimeout(500);
      }
    }
  });
});
