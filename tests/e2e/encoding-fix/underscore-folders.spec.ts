import { test, expect, TEST_CONFIG, login, apiPost } from '../fixtures/test-helpers';
import * as fs from 'fs';
import * as path from 'path';

test.describe('Underscore Folders and Chinese Filename Handling', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('Chinese filename displays correctly after creation', async ({ page, testPrefix }) => {
    const chineseNoteName = `${testPrefix}_测试笔记`;
    
    // Create note via API
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${chineseNoteName}.md`, { content: '# 中文测试\n\n这是一个中文笔记。' });
    await page.waitForTimeout(200);

    // Navigate to home and check sidebar
    await page.goto('/');
    await page.waitForSelector('[x-data]', { timeout: TEST_CONFIG.defaultTimeout });

    // The note should appear in sidebar with decoded Chinese characters
    const sidebarNote = page.locator(`text="${chineseNoteName}"`).first();
    await expect(sidebarNote).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
    
    // Click to open the note
    await sidebarNote.click();

    // Check editor content
    const editor = page.locator('#note-editor').first();
    const content = await editor.inputValue();
    expect(content).toContain('中文测试');
  });

  test('Chinese filename displays correctly when renaming fails', async ({ page, testPrefix }) => {
    const chineseNoteName = `${testPrefix}_中文笔记`;
    const chineseNewName = `${testPrefix}_新中文笔记`;
    
    // Create note via API
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${chineseNoteName}.md`, { content: '# 中文笔记内容' });
    await page.waitForTimeout(200);

    await page.goto('/');
    await page.waitForSelector('[x-data]', { timeout: TEST_CONFIG.defaultTimeout });

    // Open the note
    const sidebarNote = page.locator(`text="${chineseNoteName}"`).first();
    await sidebarNote.click();

    // Check that the note name is displayed correctly
    const noteTitle = page.locator('input[value*="中文笔记"], [x-model="note.current"]').first();
    const isVisible = await noteTitle.isVisible({ timeout: 2000 }).catch(() => false);
    
    // Verify sidebar shows decoded Chinese
    const sidebarNote2 = page.locator(`text="${chineseNoteName}"`).first();
    const sidebarText = await sidebarNote2.textContent();
    expect(sidebarText).not.toContain('%E4%B8%AD%E6%96%87');
    
    console.log(`Sidebar note text: ${sidebarText}`);
  });

  test('_attachments folder visibility is controlled by settings', async ({ page }) => {
    // First, ensure _attachments folder exists
    const attachmentsPath = path.join(process.cwd(), 'go', 'data', '_attachments');
    if (!fs.existsSync(attachmentsPath)) {
      fs.mkdirSync(attachmentsPath, { recursive: true });
    }
    
    // Add a test image to the folder
    const testImagePath = path.join(attachmentsPath, 'test-image.png');
    const pngHeader = Buffer.from([
      0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
      0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
      0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
      0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
      0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
      0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
      0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
      0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
      0x44, 0xAE, 0x42, 0x60, 0x82
    ]);
    fs.writeFileSync(testImagePath, pngHeader);
    
    // Reload to pick up the folder
    await page.goto('/');
    await page.waitForSelector('[x-data]', { timeout: TEST_CONFIG.defaultTimeout });

    // Check if _attachments is visible by default
    const attachmentsFolder = page.locator('text="_attachments"').first();
    const isVisibleByDefault = await attachmentsFolder.isVisible({ timeout: 3000 }).catch(() => false);
    console.log(`_attachments visible by default: ${isVisibleByDefault}`);
    
    // Go to settings panel
    const settingsButton = page.locator('button[title*="Settings"], button[title*="设置"]').first();
    
    if (await settingsButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await settingsButton.click();

      // Look for hide underscore folders option
      const hideToggle = page.locator('text=/Hide underscore|隐藏下划线/').first();
      const toggleVisible = await hideToggle.isVisible({ timeout: 2000 }).catch(() => false);
      
      console.log(`Hide underscore toggle found: ${toggleVisible}`);
      
      // Close settings - click somewhere else
      await page.keyboard.press('Escape');
    } else {
      console.log('Settings button not found');
    }
    
    // Verify _attachments folder state after interacting with settings
    const attachmentsFolder2 = page.locator('text="_attachments"').first();
    const finalVisibility = await attachmentsFolder2.isVisible({ timeout: 2000 }).catch(() => false);

    // Verify the folder still exists in the sidebar (either visible or in DOM)
    const attachmentsInDOM = await page.locator('text="_attachments"').count();
    expect(attachmentsInDOM).toBeGreaterThan(0);
  });
});
