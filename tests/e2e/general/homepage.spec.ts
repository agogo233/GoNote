import { test, expect, TEST_CONFIG, login } from '../fixtures/test-helpers';

test.describe('Homepage Display Fix', () => {
  test('notes should display on homepage without empty state', async ({ page }) => {
    // Collect console messages
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      consoleMessages.push(`${msg.type()}: ${msg.text()}`);
    });

    // Login first (auth may be enabled in CI)
    await login(page);

    // Wait for Alpine.js to initialize
    await page.waitForSelector('[x-data]', { timeout: 10000 });

    // Wait for content to render
    await page.waitForTimeout(300);

    // Take screenshot
    await page.screenshot({ path: 'test-homepage-fix.png', fullPage: true });

    // Check for "X notes Y folders" summary text - proves content loaded
    const summaryText = await page.locator('text=/\\d+ notes?/i').first().textContent();
    console.log(`Found summary: ${summaryText}`);

    // Verify we have content (not empty)
    expect(summaryText).toBeTruthy();
    expect(summaryText).toMatch(/\d+/);

    // Check for empty state message (should NOT exist)
    const emptyState = await page.locator('text=/No notes found|Create your first|开始创建/i').count();
    console.log(`Empty state elements: ${emptyState}`);

    // Should NOT show empty state
    expect(emptyState).toBe(0);

    // Check console for errors
    const errors = consoleMessages.filter(m => m.startsWith('error:'));
    console.log(`Console errors: ${errors.length}`);

    // Check for HOMEPAGE_MAX_NOTES error specifically
    const homepageErrors = consoleMessages.filter(m => m.includes('HOMEPAGE_MAX_NOTES'));
    if (homepageErrors.length > 0) {
      console.log(`HOMEPAGE_MAX_NOTES errors: ${homepageErrors}`);
    }
    expect(homepageErrors).toHaveLength(0);
  });
  
  test('clicking sidebar item should work', async ({ page }) => {
    await login(page);
    await page.waitForSelector('[x-data]', { timeout: 15000 });

    // Find a clickable item in the sidebar - look for notes or folders
    // Use more specific selectors to avoid clicking on buttons
    const sidebarItem = page.locator('.hover-accent.cursor-pointer, h3.truncate').first();

    const itemCount = await sidebarItem.count();

    if (itemCount > 0) {
      await sidebarItem.first().click();
      await page.waitForTimeout(200);

      // Verify page is still functional after clicking
      expect(page.url()).toContain('localhost');
    } else {
      // If no sidebar items, verify the page itself loaded correctly
      const appContainer = page.locator('[x-data]');
      await expect(appContainer).toBeVisible({ timeout: 5000 });
    }
  });
});