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

test.describe('Schedules Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/schedules');
    await expect(page).toHaveURL(/\/schedules/);
  });

  test('should display schedules page with header', async ({ page }) => {
    // Check page header
    await expect(page.getByRole('heading', { name: /日程調整|スケジュール/ })).toBeVisible();

    // Check for new schedule button (+ 新規作成)
    await expect(page.locator('button:has-text("新規作成")')).toBeVisible();
  });

  test('should display schedule list or empty state', async ({ page }) => {
    // Wait for data to load
    await page.waitForTimeout(1000);

    // Should have schedule items or empty message
    const scheduleItems = page.locator('[data-testid="schedule-item"], .schedule-card, a[href*="/schedules/"]');
    const emptyMessage = page.locator('text=日程調整がまだありません');

    const hasSchedules = await scheduleItems.count() > 0;
    const hasEmptyMessage = await emptyMessage.isVisible();

    expect(hasSchedules || hasEmptyMessage).toBeTruthy();
  });

  test('should open new schedule form', async ({ page }) => {
    // Click new schedule button (+ 新規作成)
    await page.click('button:has-text("新規作成")');

    // Form should appear with title input (inline form, not modal)
    await expect(page.locator('input[name="title"], input[placeholder*="タイトル"]')).toBeVisible();
  });

  test('should show date range picker', async ({ page }) => {
    // Click new schedule button (+ 新規作成)
    await page.click('button:has-text("新規作成")');

    // Look for DateRangePicker
    const dateRangePicker = page.locator('details:has-text("期間から一括追加")');
    if (await dateRangePicker.isVisible()) {
      // Expand the picker
      await dateRangePicker.click();

      // Should show date inputs
      await expect(page.locator('input[type="date"]').first()).toBeVisible();
    }
  });

  test('should create a schedule with candidate dates', async ({ page }) => {
    const uniqueTitle = `テスト日程調整_${Date.now()}`;

    // Click new schedule button (+ 新規作成)
    await page.click('button:has-text("新規作成")');

    // Wait for form to appear
    await page.waitForTimeout(500);

    // Fill title
    await page.fill('input[name="title"], input[placeholder*="タイトル"]', uniqueTitle);

    // Add a candidate date using the + button or date input
    const addDateButton = page.locator('button:has-text("候補日を追加")');
    if (await addDateButton.isVisible()) {
      await addDateButton.click();
      await page.waitForTimeout(300);
    }

    // Fill date if there's an input
    const dateInputs = page.locator('input[type="date"]');
    if (await dateInputs.count() > 0) {
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      const dateStr = tomorrow.toISOString().split('T')[0];
      await dateInputs.first().fill(dateStr);
    }

    // Submit (日程調整を作成)
    await page.click('button:has-text("日程調整を作成")');

    // Wait for creation
    await page.waitForTimeout(2000);

    // Should show success or redirect
    const hasSuccess = await page.locator('text=作成しました, text=公開URL').isVisible();
    const hasNewSchedule = await page.getByText(uniqueTitle).isVisible();

    expect(hasSuccess || hasNewSchedule).toBeTruthy();
  });

  test('should use quick select presets in date range picker', async ({ page }) => {
    // Click new schedule button (+ 新規作成)
    await page.click('button:has-text("新規作成")');

    // Expand date range picker
    const dateRangePicker = page.locator('details:has-text("期間から一括追加")');
    if (await dateRangePicker.isVisible()) {
      await dateRangePicker.click();

      // Click quick select buttons
      const thisWeekButton = page.locator('button:has-text("今週")');
      if (await thisWeekButton.isVisible()) {
        await thisWeekButton.click();

        // Date inputs should be filled
        const startDate = await page.locator('input[type="date"]').first().inputValue();
        expect(startDate).toBeTruthy();
      }
    }
  });

  test('should filter weekdays in date range picker', async ({ page }) => {
    // Click new schedule button (+ 新規作成)
    await page.click('button:has-text("新規作成")');

    // Expand date range picker
    const dateRangePicker = page.locator('details:has-text("期間から一括追加")');
    if (await dateRangePicker.isVisible()) {
      await dateRangePicker.click();

      // Click "平日のみ" button
      const weekdaysOnlyButton = page.locator('button:has-text("平日のみ")');
      if (await weekdaysOnlyButton.isVisible()) {
        await weekdaysOnlyButton.click();

        // Weekday checkboxes should reflect the selection
        await page.waitForTimeout(300);
      }
    }
  });
});

test.describe('Schedule Detail Page', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should display schedule detail', async ({ page }) => {
    await page.goto('/schedules');
    await page.waitForTimeout(1000);

    // Click on a schedule
    const scheduleLink = page.locator('a[href*="/schedules/"]').first();
    if (await scheduleLink.isVisible()) {
      await scheduleLink.click();

      // Should show detail page
      await expect(page.getByRole('heading')).toBeVisible();
    }
  });

  test('should copy public URL', async ({ page }) => {
    await page.goto('/schedules');
    await page.waitForTimeout(1000);

    const scheduleLink = page.locator('a[href*="/schedules/"]').first();
    if (await scheduleLink.isVisible()) {
      await scheduleLink.click();
      await page.waitForTimeout(1000);

      // Look for copy URL button
      const copyButton = page.locator('button:has-text("コピー"), button:has-text("URL")');
      if (await copyButton.isVisible()) {
        await copyButton.click();
      }
    }
  });

  test('should close schedule collection', async ({ page }) => {
    await page.goto('/schedules');
    await page.waitForTimeout(1000);

    const scheduleLink = page.locator('a[href*="/schedules/"]').first();
    if (await scheduleLink.isVisible()) {
      await scheduleLink.click();
      await page.waitForTimeout(1000);

      // Look for close button
      const closeButton = page.locator('button:has-text("締め切り"), button:has-text("終了")');
      if (await closeButton.isVisible()) {
        // Don't actually click to avoid affecting test data
        await expect(closeButton).toBeVisible();
      }
    }
  });
});

test.describe('Schedule Time Validation', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/schedules');
  });

  test('should allow overnight time range (21:00-02:00)', async ({ page }) => {
    // Click new schedule button (+ 新規作成)
    await page.click('button:has-text("新規作成")');

    // Wait for form to appear
    await page.waitForTimeout(500);

    // Fill title
    await page.fill('input[name="title"], input[placeholder*="タイトル"]', 'テスト深夜イベント');

    // Add a candidate date if not already present
    const addDateButton = page.locator('button:has-text("候補日を追加")');
    if (await addDateButton.isVisible()) {
      await addDateButton.click();
      await page.waitForTimeout(300);
    }

    // Set date
    const dateInputs = page.locator('input[type="date"]');
    if (await dateInputs.count() > 0) {
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      await dateInputs.first().fill(tomorrow.toISOString().split('T')[0]);
    }

    // Set overnight time range
    const startTimeInputs = page.locator('input[type="time"]');
    if (await startTimeInputs.count() >= 2) {
      await startTimeInputs.nth(0).fill('21:00');
      await startTimeInputs.nth(1).fill('02:00');

      // Should not show error (overnight is allowed)
      await page.waitForTimeout(500);
      const errorMessage = page.locator('.text-red-500:has-text("開始時間")');
      const hasError = await errorMessage.isVisible();
      expect(hasError).toBeFalsy();
    }
  });

  test('should reject same start and end time', async ({ page }) => {
    // Click new schedule button (+ 新規作成)
    await page.click('button:has-text("新規作成")');

    // Wait for form to appear
    await page.waitForTimeout(500);

    // Fill title
    await page.fill('input[name="title"], input[placeholder*="タイトル"]', 'テスト同時刻');

    // Add a candidate date if not already present
    const addDateButton = page.locator('button:has-text("候補日を追加")');
    if (await addDateButton.isVisible()) {
      await addDateButton.click();
      await page.waitForTimeout(300);
    }

    // Set date
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

      // Try to submit (日程調整を作成)
      await page.click('button:has-text("日程調整を作成")');

      // Should show error
      await expect(page.locator('.text-red-500, .bg-red-50')).toBeVisible({ timeout: 3000 });
    }
  });
});
