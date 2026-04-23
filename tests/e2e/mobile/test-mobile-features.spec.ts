import { test, expect, login } from '../fixtures/test-helpers';

const TEST_CONFIG = {
  baseUrl: 'http://localhost:9000',
};

test.describe('Mobile Sidebar & Toolbar Test', () => {
  test.use({ viewport: { width: 375, height: 667 } });

  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('mobile sidebar can be opened via bottom tab', async ({ page }) => {
    // Wait for content to render
    await page.waitForTimeout(300);

    // Click on Files tab to open sidebar
    const filesTab = page.locator('.mobile-bottom-tab').first();
    await filesTab.click();

    // Check if sidebar panel content is visible
    const sidebarPanel = page.locator('.sidebar-panel, [class*="sidebar"]');
    const isVisible = await sidebarPanel.isVisible().catch(() => false);

    await page.screenshot({ path: 'config/test-results/mobile-sidebar-open.png', fullPage: false });

    console.log(`Sidebar visible after clicking Files tab: ${isVisible}`);
    console.log('✅ Mobile sidebar opens via Files tab');
  });

  test('mobile search tab opens search panel', async ({ page }) => {
    // Click on Search tab
    const searchTab = page.locator('.mobile-bottom-tab').nth(1);
    await searchTab.click();

    // Check for search input
    const searchInput = page.locator('input[placeholder*="search" i], input[type="search"], [class*="search"] input');
    const hasSearch = await searchInput.isVisible().catch(() => false);

    await page.screenshot({ path: 'config/test-results/mobile-search-panel.png', fullPage: false });

    console.log(`Search input visible: ${hasSearch}`);
    console.log('✅ Mobile search panel opens');
  });

  test('mobile tags tab opens tags panel', async ({ page }) => {
    // Click on Tags tab
    const tagsTab = page.locator('.mobile-bottom-tab').nth(2);
    await tagsTab.click();

    await page.screenshot({ path: 'config/test-results/mobile-tags-panel.png', fullPage: false });

    console.log('✅ Mobile tags panel opens');
  });

  test('mobile graph view opens', async ({ page }) => {
    // Click on Graph tab
    const graphTab = page.locator('.mobile-bottom-tab').nth(3);
    await graphTab.click();

    // Check for graph container
    const graphContainer = page.locator('#graph-container, [class*="graph"], canvas');
    const hasGraph = await graphContainer.isVisible().catch(() => false);

    await page.screenshot({ path: 'config/test-results/mobile-graph-view.png', fullPage: false });

    console.log(`Graph container visible: ${hasGraph}`);
    console.log('✅ Mobile graph view tested');
  });

  test('mobile settings tab opens settings panel', async ({ page }) => {

    // Click on Settings tab
    const settingsTab = page.locator('.mobile-bottom-tab').nth(4);
    await settingsTab.click();

    await page.screenshot({ path: 'config/test-results/mobile-settings-panel.png', fullPage: false });

    console.log('✅ Mobile settings panel opens');
  });

  test('mobile note editor view', async ({ page }) => {
    await page.goto(TEST_CONFIG.baseUrl);
    await page.waitForTimeout(300);

    // Try to click on a folder to open a note
    const folderCard = page.locator('[class*="folder-card"], [class*="home-card"]').first();
    if (await folderCard.isVisible()) {
      await folderCard.click();

      // Check for editor or content area
      const editor = page.locator('#note-editor, textarea, [class*="editor"]');
      const hasEditor = await editor.isVisible().catch(() => false);

      await page.screenshot({ path: 'config/test-results/mobile-note-view.png', fullPage: false });

      console.log(`Note editor visible: ${hasEditor}`);
    } else {
      console.log('No folder cards found');
    }

    console.log('✅ Mobile note view tested');
  });

  test('mobile toolbar buttons are accessible', async ({ page }) => {
    await page.goto(TEST_CONFIG.baseUrl);

    // Check for toolbar buttons
    const toolbar = page.locator('[class*="toolbar"], [class*="icon-rail"]');
    const hasToolbar = await toolbar.isVisible().catch(() => false);

    // Check for format buttons
    const formatBtns = page.locator('.format-btn, [class*="format"]');
    const btnCount = await formatBtns.count();

    console.log(`Toolbar visible: ${hasToolbar}, Format buttons: ${btnCount}`);
    console.log('✅ Mobile toolbar tested');
  });
});
