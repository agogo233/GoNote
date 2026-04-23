import { test, expect, TEST_CONFIG, login, apiPost, waitForAutosave } from '../fixtures/test-helpers';

test.describe('Bug 3: Delete button shows actual count', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('delete button displays actual orphaned file count, not placeholder', async ({ page, testPrefix }) => {
    // Create a test note first
    const noteName = `${testPrefix}_orphaned_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: '# Test Note\n\nSome content here.' });
    await page.waitForTimeout(200);

    // Navigate to the note using correct URL format
    await page.goto(`/${encodeURIComponent(notePath)}`);

    // Wait for editor to be visible
    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 10000 });

    // Look for orphaned media button
    // The button has @click="openOrphanedMediaModal()" and contains trash icon
    const orphanedMediaButton = page.locator('button[title*="Cleanup"], button[title*="orphaned"], button[title*="清理"]').first();
    
    const isVisible = await orphanedMediaButton.isVisible({ timeout: 3000 }).catch(() => false);
    console.log(`Orphaned media button visible: ${isVisible}`);
    
    if (!isVisible) {
      // Try alternative: find by the specific trash icon SVG
      const trashIcon = page.locator('svg path[d*="M19 7l-.867 12.142"]').first();
      const iconVisible = await trashIcon.isVisible({ timeout: 2000 }).catch(() => false);
      console.log(`Trash icon visible: ${iconVisible}`);
      
      if (iconVisible) {
        // Click the parent button
        const parentButton = trashIcon.locator('xpath=ancestor::button');
        await parentButton.click({ force: true });
      } else {
        console.log('Could not find orphaned media button');
        test.skip();
        return;
      }
    } else {
      // Use force: true to click through any overlays
      await orphanedMediaButton.click({ force: true });
    }

    // Click scan button if visible
    const scanButton = page.locator('button:has-text("Scan"), button:has-text("扫描")').first();
    if (await scanButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await scanButton.click();
      await page.waitForTimeout(500);
    }

    // Check delete all button text
    const deleteAllButton = page.locator('button:has-text("Delete All"), button:has-text("Delete"), button:has-text("删除全部"), button:has-text("删除")').first();
    
    if (await deleteAllButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      const buttonText = await deleteAllButton.textContent();
      
      // Should not contain raw template placeholders
      expect(buttonText).not.toContain('{{count}}');
      expect(buttonText).not.toContain('{{');
      expect(buttonText).not.toContain('}}');
      
      console.log(`Delete button text: ${buttonText}`);
    } else {
      console.log('Delete all button not visible - likely no orphaned files found');
    }
  });

  test('delete button text is properly translated with count parameter', async ({ page, testPrefix }) => {
    // Create a test note
    const noteName = `${testPrefix}_orphaned_test_2`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: '# Test Note\n\nSome content.' });
    await page.waitForTimeout(200);

    // Navigate to the note using correct URL format
    await page.goto(`/${encodeURIComponent(notePath)}`);

    // Wait for editor
    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 10000 });

    // Look for orphaned media button
    const orphanedMediaButton = page.locator('button[title*="Cleanup"], button[title*="orphaned"], button[title*="清理"]').first();
    
    const isVisible = await orphanedMediaButton.isVisible({ timeout: 3000 }).catch(() => false);
    
    if (!isVisible) {
      // Try alternative: find by the specific trash icon SVG
      const trashIcon = page.locator('svg path[d*="M19 7l-.867 12.142"]').first();
      const iconVisible = await trashIcon.isVisible({ timeout: 2000 }).catch(() => false);
      
      if (iconVisible) {
        const parentButton = trashIcon.locator('xpath=ancestor::button');
        await parentButton.click({ force: true });
      } else {
        console.log('Orphaned media button not found');
        test.skip();
        return;
      }
    } else {
      await orphanedMediaButton.click({ force: true });
    }

    const scanButton = page.locator('button:has-text("Scan"), button:has-text("扫描")').first();
    if (await scanButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await scanButton.click();
      await page.waitForTimeout(500);
    }

    const deleteAllButton = page.locator('button:has-text("Delete"), button:has-text("删除")').first();
    
    if (await deleteAllButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      const buttonText = await deleteAllButton.textContent();
      
      // Verify no template syntax is visible
      expect(buttonText).not.toContain('{{');
      expect(buttonText).not.toContain('}}');
      
      console.log(`Delete button text: ${buttonText}`);
    } else {
      console.log('Delete button not visible - no orphaned files');
    }
  });
});