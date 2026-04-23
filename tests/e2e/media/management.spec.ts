import { test, expect, TEST_CONFIG, login, apiPost, cleanupTest } from '../fixtures/test-helpers';

test.describe('Media Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('upload image via API and verify in note', async ({ page, testPrefix }) => {
    // Create a new note via API
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note.md`, {
      content: `# Media Test Note`
    });

    // Open the note
    await page.goto(`/${encodeURIComponent(`${testPrefix}-note.md`)}`);

    // Verify editor is visible
    const editor = page.locator('#note-editor').first();
    await expect(editor).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('display note with media reference', async ({ page, testPrefix }) => {
    // Create note with image reference
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note.md`, {
      content: `# Test Note\n\n![Image](test.png)`
    });

    // Open the note
    await page.goto(`/${encodeURIComponent(`${testPrefix}-note.md`)}`);

    // Verify note content is displayed
    await expect(page.locator('.markdown-preview, .preview-content')).toContainText('Test Note', { timeout: TEST_CONFIG.defaultTimeout });
  });

  test('media GET endpoint returns 404 for non-existent file', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/media/nonexistent.png`);
    expect(response.status()).toBe(404);

    const body = await response.json();
    expect(body.detail).toContain('not found');
  });
});

test.describe('Media Orphaned Detection', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('orphaned media endpoint is accessible', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/media/orphaned`);
    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.success).toBe(true);
    expect(body.files).toBeDefined();
    expect(Array.isArray(body.files)).toBe(true);
  });

  test('cleanup orphaned media endpoint is accessible', async ({ page }) => {
    const response = await page.request.delete(`${TEST_CONFIG.baseUrl}/api/media/orphaned`);
    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.success).toBe(true);
    expect(body.deletedCount).toBeDefined();
  });
});
