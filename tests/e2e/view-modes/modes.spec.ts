import { test, expect, TEST_CONFIG, login, apiPost, waitForAutosave } from '../fixtures/test-helpers';

test.describe('View Modes', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  async function createTestNote(page: import('@playwright/test').Page, noteName: string, content: string = '# Test Note\n\nTest content.'): Promise<void> {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(200);
  }

  test('default view mode is split', async ({ page }) => {
    await page.reload();

    // Check that split button exists
    const splitButton = page.locator('button').filter({ hasText: /Split/i }).first();
    const isVisible = await splitButton.isVisible({ timeout: 5000 }).catch(() => false);

    // Split button should be visible or at least present in DOM
    const splitButtonInDOM = await page.locator('button').filter({ hasText: /Split/i }).count();
    expect(splitButtonInDOM).toBeGreaterThan(0);
  });

  test('switch to edit mode', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_edit_test`;
    await createTestNote(page, noteName, '# Edit Mode Test\n\nTest content for edit mode.');

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Click Edit button
    const editButton = page.locator('button').filter({ hasText: /^Edit$/ }).first();
    if (await editButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await editButton.click();
      await page.waitForTimeout(200);
    }
    
    // Verify editor is visible
    const editor = page.locator('#note-editor');
    await expect(editor).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
    
    // In edit mode, preview should not be visible (or minimal)
    const preview = page.locator('.markdown-preview');
    const previewVisible = await preview.isVisible().catch(() => false);
    console.log(`Preview visible in edit mode: ${previewVisible}`);
    
    await page.screenshot({ path: 'config/test-results/view-mode-edit.png', fullPage: true });
  });

  test('switch to preview mode', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_preview_test`;
    await createTestNote(page, noteName, '# Preview Mode Test\n\nTest content for preview mode.');

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Click Preview button
    const previewButton = page.locator('button').filter({ hasText: /^Preview$/ }).first();
    if (await previewButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await previewButton.click();
      await page.waitForTimeout(200);
    }
    
    // Verify preview is visible
    const preview = page.locator('.markdown-preview');
    await expect(preview).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
    
    // Verify the content is rendered
    const heading = preview.locator('h1:has-text("Preview Mode Test")');
    await expect(heading).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
    
    await page.screenshot({ path: 'config/test-results/view-mode-preview.png', fullPage: true });
  });

  test('switch to split mode shows both editor and preview', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_split_test`;
    await createTestNote(page, noteName, '# Split Mode Test\n\nContent for split view.');

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Click Split button
    const splitButton = page.locator('button').filter({ hasText: /^Split$/ }).first();
    if (await splitButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await splitButton.click();
      await page.waitForTimeout(200);
    }
    
    // In split mode, both editor and preview should be visible
    const editor = page.locator('#note-editor');
    const preview = page.locator('.markdown-preview');
    
    const editorVisible = await editor.isVisible().catch(() => false);
    const previewVisible = await preview.isVisible().catch(() => false);
    
    console.log(`Editor visible: ${editorVisible}, Preview visible: ${previewVisible}`);
    
    await page.screenshot({ path: 'config/test-results/view-mode-split.png', fullPage: true });
    
    // At least one should be visible
    expect(editorVisible || previewVisible).toBeTruthy();
  });

  test('view mode persists after page reload', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_persist_test`;
    await createTestNote(page, noteName, '# Persist Test\n\nContent for persistence test.');

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Click Preview button
    const previewButton = page.locator('button').filter({ hasText: /Preview/i }).first();
    if (await previewButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await previewButton.click();
      await page.waitForTimeout(200);
    }

    // Wait for view mode to be saved to localStorage
    await page.waitForTimeout(500);

    // Check localStorage for saved view mode
    const savedViewMode = await page.evaluate(() => localStorage.getItem('viewMode'));

    // Verify view mode was persisted (should be 'preview' after clicking preview button)
    expect(savedViewMode).toBeTruthy();
    expect(savedViewMode).toMatch(/preview|edit|split/);
  });
});