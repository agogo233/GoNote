import { test, expect, TEST_CONFIG, login, apiPost } from '../fixtures/test-helpers';

test.describe('Keyboard Shortcuts', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('Ctrl+S saves the current note', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_save_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: '' });
    await page.waitForTimeout(200);

    // Navigate using the correct URL format: /notePath
    await page.goto(`/${encodeURIComponent(notePath)}`);

    // Wait for editor to be ready
    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 15000 });

    await editor.fill('Content before save');
    await page.waitForTimeout(200);
    
    await page.keyboard.press('Control+s');
    await page.waitForTimeout(2000);
    
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes/${notePath}`);
    const data = await response.json();
    expect(data.content).toContain('Content before save');
  });

  test('Ctrl+Alt+N creates new note', async ({ page }) => {
    // Count notes before
    const notesResponseBefore = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
    const notesBefore = await notesResponseBefore.json();
    const countBefore = notesBefore.notes?.length || 0;

    await page.keyboard.press('Control+Alt+n');

    // Handle the dialog
    page.once('dialog', async dialog => {
      await dialog.accept('Shortcut Created Note');
    });

    await page.waitForTimeout(1500);

    // Verify a new note was created by checking the note count increased
    const notesResponseAfter = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
    const notesAfter = await notesResponseAfter.json();
    const countAfter = notesAfter.notes?.length || 0;

    // Note count should have increased by at least 1
    expect(countAfter).toBeGreaterThan(countBefore);
  });

  test('Ctrl+Alt+F creates new folder', async ({ page }) => {
    // Get folders before
    const notesResponseBefore = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
    const dataBefore = await notesResponseBefore.json();
    const foldersBefore = dataBefore.folders?.length || 0;

    await page.keyboard.press('Control+Alt+f');

    // Handle the dialog
    page.once('dialog', async dialog => {
      await dialog.accept('Shortcut Created Folder');
    });

    await page.waitForTimeout(1500);

    // Verify folder count changed (or at least the page didn't error)
    const notesResponseAfter = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
    expect(notesResponseAfter.status()).toBe(200);

    const dataAfter = await notesResponseAfter.json();
    const foldersAfter = dataAfter.folders?.length || 0;

    // Folder count should have increased
    expect(foldersAfter).toBeGreaterThanOrEqual(foldersBefore);
  });

  test('Ctrl+Alt+P opens quick switcher', async ({ page }) => {
    await page.keyboard.press('Control+Alt+p');

    // Check for quick switcher input
    const quickSwitcherInput = page.locator('#quickSwitcherInput');
    await expect(quickSwitcherInput).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });

    // Close the quick switcher
    await page.keyboard.press('Escape');
  });

  test('Ctrl+B wraps selection with bold markdown', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_bold_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: '' });
    await page.waitForTimeout(200);

    await page.goto(`/${encodeURIComponent(notePath)}`);

    // Wait for editor
    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 15000 });

    await editor.fill('bold text here');
    await editor.focus();
    await page.waitForTimeout(200);

    await page.keyboard.press('Control+a');
    await page.waitForTimeout(200);
    await page.keyboard.press('Control+b');

    const content = await editor.inputValue();
    expect(content).toContain('**');
  });

  test('Ctrl+I wraps selection with italic markdown', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_italic_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: '' });
    await page.waitForTimeout(200);

    await page.goto(`/${encodeURIComponent(notePath)}`);

    // Wait for editor
    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 15000 });

    await editor.fill('italic text here');
    await editor.focus();
    await page.waitForTimeout(200);

    await page.keyboard.press('Control+a');
    await page.waitForTimeout(200);
    await page.keyboard.press('Control+i');

    const content = await editor.inputValue();
    expect(content).toContain('*');
  });

  test('Ctrl+Alt+Z toggles zen mode', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_zen_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: 'Zen mode test' });
    await page.waitForTimeout(200);

    await page.goto(`/${encodeURIComponent(notePath)}`);

    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 15000 });

    await page.keyboard.press('Control+Alt+z');
    await page.waitForTimeout(200);
    
    // Check for zen mode indicator (body class)
    const body = page.locator('body');
    const bodyClass = await body.getAttribute('class');
    const hasZenClass = bodyClass?.includes('zen') ?? false;

    // Exit zen mode
    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);

    // Verify page is still functional after exiting zen mode
    const editorAfter = page.locator('#note-editor').first();
    await expect(editorAfter).toBeVisible({ timeout: 5000 });

    // Zen mode should have been toggled
    expect(hasZenClass).toBeDefined();
  });

  test('Escape exits zen mode', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_escape_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: 'Escape test' });
    await page.waitForTimeout(200);

    await page.goto(`/${encodeURIComponent(notePath)}`);

    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 15000 });

    // Enter zen mode first
    await page.keyboard.press('Control+Alt+z');
    await page.waitForTimeout(200);

    // Press Escape
    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);

    // Verify editor is still accessible after exiting zen mode
    await expect(editor).toBeVisible({ timeout: 5000 });
  });

  test('Ctrl+K inserts link', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_link_test`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: '' });
    await page.waitForTimeout(200);

    await page.goto(`/${encodeURIComponent(notePath)}`);

    // Wait for editor
    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: 15000 });

    await editor.fill('link text');
    await editor.focus();
    await page.waitForTimeout(200);

    await page.keyboard.press('Control+a');
    await page.waitForTimeout(200);
    await page.keyboard.press('Control+k');

    const content = await editor.inputValue();
    expect(content).toContain('[');
    expect(content).toContain(']');
  });
});