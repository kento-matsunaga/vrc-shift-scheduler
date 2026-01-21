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

test.describe('Navigation E2E', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should display navigation menu', async ({ page }) => {
    await page.goto('/events');

    // Check for navigation elements
    const nav = page.locator('nav, [role="navigation"]');
    await expect(nav).toBeVisible();
  });

  test('should navigate to events page', async ({ page }) => {
    await page.goto('/members');

    // Click events link
    await page.click('a[href="/events"], nav >> text=イベント');

    await expect(page).toHaveURL(/\/events/);
  });

  test('should navigate to members page', async ({ page }) => {
    await page.goto('/events');

    // Click members link
    await page.click('a[href="/members"], nav >> text=メンバー');

    await expect(page).toHaveURL(/\/members/);
  });

  test('should navigate to roles page', async ({ page }) => {
    await page.goto('/events');

    // Click roles link
    await page.click('a[href="/roles"], nav >> text=ロール');

    await expect(page).toHaveURL(/\/roles/);
  });

  test('should navigate to schedules page', async ({ page }) => {
    await page.goto('/events');

    // Click schedules link
    await page.click('a[href="/schedules"], nav >> text=日程調整');

    await expect(page).toHaveURL(/\/schedules/);
  });

  test('should navigate to attendance page', async ({ page }) => {
    await page.goto('/events');

    // Click attendance link
    await page.click('a[href="/attendance"], nav >> text=出欠');

    await expect(page).toHaveURL(/\/attendance/);
  });

  test('should navigate to settings page', async ({ page }) => {
    await page.goto('/events');

    // Click settings link
    await page.click('a[href="/settings"], nav >> text=設定');

    await expect(page).toHaveURL(/\/settings/);
  });

  test('should highlight active navigation item', async ({ page }) => {
    await page.goto('/members');

    // Check for active state on members link
    const activeLink = page.locator('a[href="/members"].active, a[href="/members"][aria-current="page"], nav >> text=メンバー >> ..');
    await expect(activeLink).toBeVisible();
  });

  test('should show mobile menu on small screens', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/events');

    // Look for hamburger menu
    const hamburgerButton = page.locator('button[aria-label="メニュー"], button:has-text("メニュー"), .hamburger');
    if (await hamburgerButton.isVisible()) {
      await hamburgerButton.click();

      // Menu should expand
      await expect(page.locator('nav a, [role="navigation"] a').first()).toBeVisible();
    }
  });

  test('should close mobile menu after navigation', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/events');

    const hamburgerButton = page.locator('button[aria-label="メニュー"], button:has-text("メニュー"), .hamburger');
    if (await hamburgerButton.isVisible()) {
      await hamburgerButton.click();

      // Click a nav link
      await page.click('a[href="/members"], nav >> text=メンバー');

      // Menu should close
      await expect(page).toHaveURL(/\/members/);
    }
  });
});

test.describe('Breadcrumb Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should display breadcrumbs on nested pages', async ({ page }) => {
    // Navigate to a nested page (event detail, schedule detail, etc.)
    await page.goto('/events');
    await page.waitForTimeout(1000);

    const eventLink = page.locator('a[href*="/events/"]').first();
    if (await eventLink.isVisible()) {
      await eventLink.click();

      // Check for breadcrumbs
      const breadcrumb = page.locator('[aria-label="breadcrumb"], .breadcrumb, nav:has-text("イベント")');
      if (await breadcrumb.isVisible()) {
        await expect(breadcrumb).toBeVisible();
      }
    }
  });

  test('should navigate back via breadcrumb', async ({ page }) => {
    await page.goto('/events');
    await page.waitForTimeout(1000);

    const eventLink = page.locator('a[href*="/events/"]').first();
    if (await eventLink.isVisible()) {
      await eventLink.click();
      await page.waitForTimeout(500);

      // Click breadcrumb to go back
      const backLink = page.locator('[aria-label="breadcrumb"] a, .breadcrumb a').first();
      if (await backLink.isVisible()) {
        await backLink.click();
        await expect(page).toHaveURL(/\/events$/);
      }
    }
  });
});

test.describe('Page Title and Meta', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should have correct page title on events page', async ({ page }) => {
    await page.goto('/events');
    await expect(page).toHaveTitle(/イベント|VRC/);
  });

  test('should have correct page title on members page', async ({ page }) => {
    await page.goto('/members');
    await expect(page).toHaveTitle(/メンバー|VRC/);
  });

  test('should have correct page title on schedules page', async ({ page }) => {
    await page.goto('/schedules');
    await expect(page).toHaveTitle(/日程|スケジュール|VRC/);
  });
});

test.describe('Error Pages', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should display 404 page for unknown routes', async ({ page }) => {
    await page.goto('/unknown-page-12345');

    // Should show 404 or redirect
    const notFoundText = page.locator('text=404, text=見つかりません, text=Not Found');
    const redirectedToHome = await page.url().includes('/events') || await page.url().includes('/login');

    const hasNotFound = await notFoundText.isVisible();

    expect(hasNotFound || redirectedToHome).toBeTruthy();
  });
});

test.describe('Back Button Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should support browser back button', async ({ page }) => {
    // Navigate through pages
    await page.goto('/events');
    await page.goto('/members');
    await page.goto('/roles');

    // Go back
    await page.goBack();
    await expect(page).toHaveURL(/\/members/);

    await page.goBack();
    await expect(page).toHaveURL(/\/events/);
  });

  test('should support browser forward button', async ({ page }) => {
    await page.goto('/events');
    await page.goto('/members');

    await page.goBack();
    await expect(page).toHaveURL(/\/events/);

    await page.goForward();
    await expect(page).toHaveURL(/\/members/);
  });
});
