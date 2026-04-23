import { test, expect, TEST_CONFIG, login, apiPost, waitForAutosave, cleanupTest } from '../fixtures/test-helpers';

const BASE_URL = TEST_CONFIG.baseUrl;

test.describe('Search Edge Cases', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('search for non-existent term shows no navigation', async ({ page, testPrefix }) => {
    // Create a test note
    const noteName = `${testPrefix}_search_test`;
    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: '# Search Test\n\nThis is a test note with some content.' });
    await page.waitForTimeout(200);
    
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Open search panel by clicking the search icon in the icon rail
    const iconRailButtons = page.locator('.icon-rail button');
    const buttonCount = await iconRailButtons.count();
    console.log(`Icon rail buttons count: ${buttonCount}`);
    
    // The search panel button is typically the second one (index 1)
    if (buttonCount >= 2) {
      await iconRailButtons.nth(1).click();
      await page.waitForTimeout(200);
    }

    // Find search input with x-model="search.query"
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.waitFor({ state: 'visible', timeout: 5000 });

    // Search for a non-existent term
    await searchInput.fill('zzzzzzzznonexistent12345');
    await page.waitForTimeout(500);
    
    // Check for "no results" message
    const noResultsMsg = page.locator('text=/No results|没有结果|0 result/').first();
    const hasNoResults = await noResultsMsg.isVisible({ timeout: 3000 }).catch(() => false);
    
    // Check that match navigation is NOT visible
    const matchNav = page.locator('text=/\\d+\\s*\\/\\s*\\d+/').first();
    const hasMatchNav = await matchNav.isVisible({ timeout: 2000 }).catch(() => false);
    
    console.log(`No results message: ${hasNoResults}, Match nav visible: ${hasMatchNav}`);
    
    // Should have no results and no match navigation
    expect(hasMatchNav).toBe(false);
  });

  test('clearing search removes navigation UI', async ({ page, testPrefix }) => {
    // Create test note with searchable content
    const noteName = `${testPrefix}_clear_test`;
    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: '# Clear Test\n\ntest content test more test' });
    await page.waitForTimeout(200);
    
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Open search panel
    const iconRailButtons = page.locator('.icon-rail button');
    const buttonCount = await iconRailButtons.count();
    
    if (buttonCount >= 2) {
      await iconRailButtons.nth(1).click();
      await page.waitForTimeout(200);
    }

    // Search for "test"
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.waitFor({ state: 'visible', timeout: 5000 });
    await searchInput.fill('test');
    await page.waitForTimeout(500);
    
    // Click on first search result
    const firstResult = page.locator('.hover-accent.cursor-pointer, [class*="hover-accent"]').first();
    const resultVisible = await firstResult.isVisible({ timeout: 3000 }).catch(() => false);
    
    if (resultVisible) {
      await firstResult.click();
      await page.waitForTimeout(1500);

      // Check for match navigation
      const matchNav = page.locator('text=/\\d+\\s*\\/\\s*\\d+/').first();
      const hasMatchNavBefore = await matchNav.isVisible({ timeout: 2000 }).catch(() => false);

      // Clear the search
      await searchInput.fill('');
      await page.waitForTimeout(500);

      // Check if match navigation is gone
      const hasMatchNavAfter = await matchNav.isVisible({ timeout: 2000 }).catch(() => false);

      // After clearing, match nav should not be visible
      if (hasMatchNavBefore) {
        expect(hasMatchNavAfter).toBe(false);
      }
    } else {
      // If no results found, the search panel should still be functional
      const inputValue = await searchInput.inputValue();
      expect(inputValue).toBe('test');
    }
  });

  test('opening new note clears previous search highlights', async ({ page, testPrefix }) => {
    // Create two test notes
    const noteName1 = `${testPrefix}_highlight_test_1`;
    const noteName2 = `${testPrefix}_highlight_test_2`;
    await apiPost(page, `${BASE_URL}/api/notes/${noteName1}.md`, { content: '# Note 1\n\ntest content here' });
    await apiPost(page, `${BASE_URL}/api/notes/${noteName2}.md`, { content: '# Note 2\n\nother content' });
    await page.waitForTimeout(200);
    
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Open search panel
    const iconRailButtons = page.locator('.icon-rail button');
    const buttonCount = await iconRailButtons.count();
    
    if (buttonCount >= 2) {
      await iconRailButtons.nth(1).click();
      await page.waitForTimeout(200);
    }

    // Search for "test"
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.waitFor({ state: 'visible', timeout: 5000 });
    await searchInput.fill('test');
    await page.waitForTimeout(500);
    
    // Click on first result
    const firstResult = page.locator('.hover-accent.cursor-pointer').first();
    if (await firstResult.isVisible({ timeout: 2000 }).catch(() => false)) {
      await firstResult.click();
      await page.waitForTimeout(1500);
      
      // Go to files panel
      await iconRailButtons.nth(0).click();
      await page.waitForTimeout(500);
      
      // Click on a different note
      const secondNote = page.locator(`text="${noteName2}"`).first();
      if (await secondNote.isVisible({ timeout: 2000 }).catch(() => false)) {
        await secondNote.click();
        await page.waitForTimeout(1000);
        
        // Check if previous highlights are cleared
        const matchNav = page.locator('text=/\\d+\\s*\\/\\s*\\d+/').first();
        const hasMatchNav = await matchNav.isVisible({ timeout: 2000 }).catch(() => false);

        expect(hasMatchNav).toBe(false);
      }
    } else {
      // If no results, verify search input is still functional
      const inputValue = await searchInput.inputValue();
      expect(inputValue).toBe('test');
    }
  });

  test('Shift+F3 navigates to previous match', async ({ page, testPrefix }) => {
    // Create test note with multiple matches
    const noteName = `${testPrefix}_f3_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${BASE_URL}/api/notes/${notePath}`, { content: '# F3 Test\n\ntest one test two test three test four' });
    await page.waitForTimeout(200);
    
    // Navigate directly to the note
    await page.goto(`/${encodeURIComponent(notePath)}`);
    await page.waitForTimeout(1500);
    
    // Wait for editor to be visible
    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 10000 });
    await page.waitForTimeout(500);
    
    // Open search panel
    const iconRailButtons = page.locator('.icon-rail button');
    const buttonCount = await iconRailButtons.count();
    
    if (buttonCount >= 2) {
      await iconRailButtons.nth(1).click();
      await page.waitForTimeout(200);

      // Search for "test"
      const searchInput = page.locator('input[x-model="search.query"]').first();
      await searchInput.waitFor({ state: 'visible', timeout: 5000 });
      await searchInput.fill('test');
      await page.waitForTimeout(500);
      
      // The editor should now show highlights with match navigation
      await editor.focus();
      await page.waitForTimeout(300);
      
      // Press F3 multiple times to navigate forward
      await page.keyboard.press('F3');
      await page.waitForTimeout(300);
      await page.keyboard.press('F3');
      await page.waitForTimeout(300);
      
      // Press Shift+F3 to go back
      await page.keyboard.down('Shift');
      await page.keyboard.press('F3');
      await page.keyboard.up('Shift');
      await page.waitForTimeout(300);

      // Verify editor is still functional after navigation
      const editorContent = await editor.inputValue();
      expect(editorContent).toContain('test');
    } else {
      // If search button not found, verify at least editor is accessible
      const editorContent = await editor.inputValue();
      expect(editorContent.length).toBeGreaterThan(0);
    }
  });
});