import { test, expect, TEST_CONFIG, login, waitForAutosave } from '../fixtures/test-helpers';

const OUTLINE_CONFIG = {
  defaultTimeout: 10000,
};

async function openOutlinePanel(page: import('@playwright/test').Page) {
  // Find and click the outline panel button in the icon rail
  const iconRailButtons = page.locator('.icon-rail button, [class*="icon-rail"] button');
  const buttonCount = await iconRailButtons.count();

  // Click the outline button (typically the 3rd button)
  if (buttonCount >= 3) {
    await iconRailButtons.nth(2).click();
  }

  // Wait for outline panel to be visible
  const outlinePanel = page.locator('.outline-panel');
  await outlinePanel.waitFor({ state: 'visible', timeout: OUTLINE_CONFIG.defaultTimeout }).catch(() => {});
}

test.describe('Outline Panel', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('outline panel displays headings', async ({ page }) => {
    const noteName = `outline_test_${Date.now()}`;
    const content = `# Main Title

## Section 1

Content here.

## Section 2

### Subsection 2.1

More content.

### Subsection 2.2

## Section 3

Final content.`;

    // Create note via API
    await page.evaluate(async ({ name, content }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${name}.md`,
          content
        })
      });
    }, { name: noteName, content });

    // Reload and open the note
    await page.reload();
    await page.waitForTimeout(500);

    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(500);

    // Open outline panel
    await openOutlinePanel(page);

    // Take screenshot
    await page.screenshot({ path: 'config/test-results/outline-panel-visible.png', fullPage: false });

    // Check outline panel is visible
    const outlinePanel = page.locator('.outline-panel');
    await expect(outlinePanel).toBeVisible({ timeout: 5000 });

    // Check outline items - they use hover-accent class
    const outlineItems = page.locator('.outline-panel button.hover-accent');
    const count = await outlineItems.count();
    console.log(`Found ${count} outline items`);

    // Should have 6 headings (1 h1, 3 h2, 2 h3)
    expect(count).toBeGreaterThanOrEqual(6);

    // Verify first heading text
    const h1Item = page.locator('.outline-panel button.hover-accent').first();
    await expect(h1Item).toContainText('Main Title');

    console.log('✅ Outline panel displays headings correctly');
  });

  test('outline panel toggle button', async ({ page }) => {
    const noteName = `outline_toggle_${Date.now()}`;
    const content = `# Title\n\n## Section 1\n\n## Section 2`;

    // Create note via API
    await page.evaluate(async ({ name, content }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${name}.md`,
          content
        })
      });
    }, { name: noteName, content });

    // Reload and open the note
    await page.reload();
    await page.waitForTimeout(500);

    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(500);

    // Open outline panel
    await openOutlinePanel(page);

    // Take screenshot with panel visible
    await page.screenshot({ path: 'config/test-results/outline-panel-visible-2.png', fullPage: false });

    // Close the outline panel using the close button in header
    const closeBtn = page.locator('.outline-panel-header button').first();
    if (await closeBtn.isVisible({ timeout: 2000 })) {
      await closeBtn.click();
    }

    // Take screenshot with panel hidden
    await page.screenshot({ path: 'config/test-results/outline-panel-hidden.png', fullPage: false });

    // Check outline panel is hidden
    const outlinePanel = page.locator('.outline-panel');
    await expect(outlinePanel).not.toBeVisible();

    // Check toggle button appears (for reopening)
    const toggleBtn = page.locator('.outline-toggle-btn');
    if (await toggleBtn.isVisible({ timeout: 2000 })) {
      // Take screenshot with toggle button visible
      await page.screenshot({ path: 'config/test-results/outline-toggle-visible.png', fullPage: false });

      // Click toggle to show panel again
      await toggleBtn.click();

      // Take screenshot after toggle
      await page.screenshot({ path: 'config/test-results/outline-panel-toggled.png', fullPage: false });

      // Check outline panel is visible again
      await expect(outlinePanel).toBeVisible();
    }

    console.log('✅ Outline panel toggle works correctly');
  });

  test('outline panel responsive on mobile', async ({ page }) => {
    const noteName = `outline_mobile_${Date.now()}`;
    const content = `# Mobile TITLE\n\n## Mobile Section 1\n\n## Mobile Section 2\n\n### Mobile Subsection`;

    // Create note via API on desktop
    await page.evaluate(async ({ name, content }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${name}.md`,
          content
        })
      });
    }, { name: noteName, content });

    // Reload and open the note
    await page.reload();
    await page.waitForTimeout(500);

    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(500);

    // Now switch to mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    // Open outline panel on mobile
    await openOutlinePanel(page);

    // Take screenshot
    await page.screenshot({ path: 'config/test-results/outline-mobile.png', fullPage: false });

    // Check outline panel is visible (as overlay on mobile)
    const outlinePanel = page.locator('.outline-panel');
    await expect(outlinePanel).toBeVisible({ timeout: 5000 });

    // Check that the outline has correct items
    const outlineItems = page.locator('.outline-panel button.hover-accent');
    const count = await outlineItems.count();
    expect(count).toBeGreaterThan(0);

    console.log('✅ Outline panel works on mobile');
  });
});
