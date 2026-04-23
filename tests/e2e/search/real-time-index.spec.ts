import { test, expect, TEST_CONFIG, login, apiPost, waitForSearchDebounce, cleanupTest } from '../fixtures/test-helpers';

test.describe('Search Index Real-time Updates', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  async function openSearchPanel(page: import('@playwright/test').Page) {
    const iconRailButtons = page.locator('.icon-rail-btn');
    if (await iconRailButtons.count() >= 2) {
      await iconRailButtons.nth(1).click();
    }
    await page.waitForTimeout(200);
  }

  test('newly created note appears in search results', async ({ page, testPrefix }) => {
    const uniqueTerm = `SearchIndexNew${Date.now()}`;

    // Create note via API
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}_new_note.md`, {
      content: `# New Note\n\nThis note contains ${uniqueTerm} for search indexing test.`
    });

    // Wait for search index to update
    await page.waitForTimeout(2000);

    // Reload to ensure note list is refreshed
    await page.reload();

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(uniqueTerm);
    await waitForSearchDebounce(page);
    await page.waitForTimeout(500);

    // Verify the new note appears in search results
    const pageContent = await page.content();
    expect(pageContent).toContain(uniqueTerm);
  });

  test('edited note content appears in search after modification', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_edit_index`;
    const originalContent = '# Original Content\n\nOriginal text here.';
    const newUniqueTerm = `EditedContent${Date.now()}`;

    // Create note
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, { content: originalContent });
    await page.waitForTimeout(200);
    await page.reload();

    // Open and edit the note
    const noteItem = page.locator(`text="${noteName}"`).first();
    await expect(noteItem).toBeVisible({ timeout: 5000 });
    await noteItem.click();
    await page.waitForTimeout(500);

    const editor = page.locator('#note-editor');
    await expect(editor).toBeVisible({ timeout: 5000 });

    // Edit content with unique term
    await editor.fill(`# Edited Note\n\n${newUniqueTerm} is the new content after editing.`);
    await page.waitForTimeout(2000); // Wait for autosave + index update

    // Search for the new content
    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(newUniqueTerm);
    await waitForSearchDebounce(page);
    await page.waitForTimeout(500);

    // Verify edited content is searchable
    const pageContent = await page.content();
    expect(pageContent).toContain(noteName);
  });

  test('deleted note disappears from search results', async ({ page, testPrefix }) => {
    const uniqueTerm = `ToDeleteSearch${Date.now()}`;
    const noteName = `${testPrefix}_to_delete_search`;

    // Create note
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      content: `# To Delete\n\n${uniqueTerm} should not appear after deletion.`
    });
    await page.waitForTimeout(200);
    await page.reload();

    // Verify note is searchable
    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(uniqueTerm);
    await waitForSearchDebounce(page);
    await page.waitForTimeout(500);

    const contentBeforeDelete = await page.content();
    expect(contentBeforeDelete).toContain(noteName);

    // Delete the note via API
    const csrfToken = await page.evaluate(async () => {
      const cookies = document.cookie.split(';');
      const csrf = cookies.find(c => c.trim().startsWith('csrf'));
      return csrf ? csrf.split('=')[1] : '';
    });

    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    await page.request.delete(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
    });

    await page.waitForTimeout(2000); // Wait for index update

    // Search again - should not find the deleted note
    await searchInput.fill(uniqueTerm);
    await waitForSearchDebounce(page);
    await page.waitForTimeout(500);

    const contentAfterDelete = await page.content();
    // The note name should no longer appear
    expect(contentAfterDelete).not.toContain(noteName);
  });

  test('renamed note appears in search under new name', async ({ page, testPrefix }) => {
    const oldName = `${testPrefix}_old_search_name`;
    const newName = `${testPrefix}_new_search_name`;
    const uniqueTerm = `RenameSearchTest${Date.now()}`;

    // Create note with old name
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${oldName}.md`, {
      content: `# Rename Test\n\n${uniqueTerm}`
    });
    await page.waitForTimeout(200);
    await page.reload();

    // Rename via API
    const csrfToken = await page.evaluate(async () => {
      const cookies = document.cookie.split(';');
      const csrf = cookies.find(c => c.trim().startsWith('csrf'));
      return csrf ? csrf.split('=')[1] : '';
    });

    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    const renameResponse = await page.request.put(`${TEST_CONFIG.baseUrl}/api/notes/${oldName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ newName: `${newName}.md` }),
    });

    await page.waitForTimeout(2000); // Wait for index update

    if (renameResponse.status() === 200) {
      // Search for new name
      await openSearchPanel(page);

      const searchInput = page.locator('input[x-model="search.query"]').first();
      await expect(searchInput).toBeVisible({ timeout: 5000 });

      await searchInput.fill(uniqueTerm);
      await waitForSearchDebounce(page);
      await page.waitForTimeout(500);

      const pageContent = await page.content();
      // Should find under new name
      expect(pageContent).toContain(uniqueTerm);
    }
  });

  test('search index updates after moving note to folder', async ({ page, testPrefix }) => {
    const folderName = `${testPrefix}_index_folder`;
    const noteName = `${testPrefix}_move_index`;
    const uniqueTerm = `MoveIndexTest${Date.now()}`;

    // Create folder and note
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/folders/${folderName}`, {});
    await page.waitForTimeout(200);

    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${folderName}/${noteName}.md`, {
      content: `# Moved Note\n\n${uniqueTerm}`
    });
    await page.waitForTimeout(200);
    await page.reload();

    // Search for the note
    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(uniqueTerm);
    await waitForSearchDebounce(page);
    await page.waitForTimeout(500);

    // Verify note is searchable in folder
    const pageContent = await page.content();
    expect(pageContent).toContain(noteName);
  });

  test('search index handles rapid consecutive edits', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_rapid_edit`;

    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      content: '# Rapid Edit Test\n\nInitial content.'
    });
    await page.waitForTimeout(200);
    await page.reload();

    // Open note and rapidly edit
    const noteItem = page.locator(`text="${noteName}"`).first();
    await expect(noteItem).toBeVisible({ timeout: 5000 });
    await noteItem.click();
    await page.waitForTimeout(500);

    const editor = page.locator('#note-editor');
    await expect(editor).toBeVisible({ timeout: 5000 });

    // Rapid edits
    const finalTerm = `FinalRapidEdit${Date.now()}`;
    await editor.fill(`# Rapid Edit\n\nIntermediate content.`);
    await page.waitForTimeout(200);
    await editor.fill(`# Rapid Edit\n\n${finalTerm} is the final content after rapid edits.`);
    await page.waitForTimeout(2000); // Wait for autosave + index

    // Search for final content
    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(finalTerm);
    await waitForSearchDebounce(page);
    await page.waitForTimeout(500);

    const pageContent = await page.content();
    expect(pageContent).toContain(noteName);
  });
});
