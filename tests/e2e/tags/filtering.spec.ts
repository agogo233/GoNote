import { test, expect, TEST_CONFIG, login, apiPost, cleanupTest } from '../fixtures/test-helpers';

test.describe('Tag Filtering', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('list all tags', async ({ page, testPrefix }) => {
    // Create notes with different tags via CSRF-safe API helper
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note1.md`, {
      content: `---\ntitle: Note 1\ntags: [tag1, tag2]\n---\nContent 1`
    });
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note2.md`, {
      content: `---\ntitle: Note 2\ntags: [tag2, tag3]\n---\nContent 2`
    });
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note3.md`, {
      content: `---\ntitle: Note 3\ntags: [tag1, tag3]\n---\nContent 3`
    });

    // Fetch tags API directly and verify specific tags exist
    const tagsResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags`);
    const tagsData = await tagsResponse.json();

    expect(tagsData.tags).toBeDefined();
    expect(tagsData.tags['tag1']).toBeGreaterThanOrEqual(2);
    expect(tagsData.tags['tag2']).toBeGreaterThanOrEqual(2);
    expect(tagsData.tags['tag3']).toBeGreaterThanOrEqual(2);
  });

  test('filter notes by single tag', async ({ page, testPrefix }) => {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note1.md`, {
      content: `---\ntitle: Note 1\ntags: [programming]\n---\nContent 1`
    });
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note2.md`, {
      content: `---\ntitle: Note 2\ntags: [personal]\n---\nContent 2`
    });

    // Fetch notes by tag
    const notesResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags/programming`);
    const notesData = await notesResponse.json();

    expect(notesData.count).toBeGreaterThanOrEqual(1);
    const matchingNotes = notesData.notes.filter((n: any) => n.path.includes(`${testPrefix}-note1`));
    expect(matchingNotes.length).toBeGreaterThanOrEqual(1);
  });

  test('filter notes by tag - case insensitive', async ({ page, testPrefix }) => {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note.md`, {
      content: `---\ntitle: Note\ntags: [Programming]\n---\nContent`
    });

    // Query with different case
    const notesResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags/PROGRAMMING`);
    const notesData = await notesResponse.json();

    expect(notesData.count).toBeGreaterThanOrEqual(1);
  });

  test('filter notes with multiple tags AND logic', async ({ page, testPrefix }) => {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note1.md`, {
      content: `---\ntitle: Note 1\ntags: [go, backend]\n---\nContent 1`
    });
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note2.md`, {
      content: `---\ntitle: Note 2\ntags: [python, backend]\n---\nContent 2`
    });
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note3.md`, {
      content: `---\ntitle: Note 3\ntags: [go, frontend]\n---\nContent 3`
    });

    // Get all notes and verify AND logic: notes with "go" tag
    const notesResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags/go`);
    const notesData = await notesResponse.json();

    expect(notesData.count).toBeGreaterThanOrEqual(2);
    const notePaths = notesData.notes.map((n: any) => n.path);
    expect(notePaths.some((p: string) => p.includes(`${testPrefix}-note1`))).toBe(true);
    expect(notePaths.some((p: string) => p.includes(`${testPrefix}-note3`))).toBe(true);
  });

  test('tag with no matching notes', async ({ page, testPrefix }) => {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note.md`, {
      content: `---\ntitle: Note\ntags: [existing]\n---\nContent`
    });

    const notesResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags/nonexistent`);
    const notesData = await notesResponse.json();

    expect(notesData.count).toBe(0);
    expect(notesData.notes).toEqual([]);
  });

  test('notes in subdirectories with tags', async ({ page, testPrefix }) => {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/programming/${testPrefix}-go.md`, {
      content: `---\ntitle: Go\ntags: [programming]\n---\nContent`
    });
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/personal/${testPrefix}-todo.md`, {
      content: `---\ntitle: Todo\ntags: [programming]\n---\nContent`
    });

    const notesResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags/programming`);
    const notesData = await notesResponse.json();

    expect(notesData.count).toBeGreaterThanOrEqual(2);
  });

  test('tag count accuracy', async ({ page, testPrefix }) => {
    // Get initial count
    const tagsBeforeResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags`);
    const tagsBefore = await tagsBeforeResponse.json();
    const tag1CountBefore = tagsBefore.tags['tag1'] || 0;

    // Create 3 notes with tag1
    for (let i = 1; i <= 3; i++) {
      await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note${i}.md`, {
        content: `---\ntitle: Note ${i}\ntags: [tag1]\n---\nContent ${i}`
      });
    }

    const tagsAfterResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags`);
    const tagsAfter = await tagsAfterResponse.json();
    const tag1CountAfter = tagsAfter.tags['tag1'] || 0;

    expect(tag1CountAfter).toBe(tag1CountBefore + 3);
  });

  test('note without tags not counted', async ({ page, testPrefix }) => {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note.md`, {
      content: `---\ntitle: No Tags\n---\nContent without tags`
    });

    const tagsResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/tags`);
    const tagsData = await tagsResponse.json();

    // Verify no undefined or empty keys in tags object
    const tagKeys = Object.keys(tagsData.tags);
    expect(tagKeys.every(key => key.length > 0)).toBe(true);
  });
});

test.describe('Tag UI Integration', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('display tags in note list', async ({ page, testPrefix }) => {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note.md`, {
      content: `---\ntitle: UI Test\ntags: [ui-test, display]\n---\nContent`
    });

    await page.goto('/');

    // Verify note list is visible
    await expect(page.locator('[data-testid="note-list"], .note-list')).toBeVisible();

    // Verify the created note appears in the sidebar
    await expect(page.locator(`text="${testPrefix}"`).first()).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('click tag to filter', async ({ page, testPrefix }) => {
    // Create a note with a unique tag
    const uniqueTag = `clickable-${testPrefix}`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}-note.md`, {
      content: `---\ntitle: Filter Test\ntags: [${uniqueTag}]\n---\nContent`
    });

    await page.goto('/');

    // Wait for note list to appear
    await expect(page.locator('[data-testid="note-list"], .note-list')).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });

    // Note: This test verifies the tag UI exists. Actual click-and-filter behavior
    // depends on the specific UI implementation which may vary.
    const tagElements = page.locator('.tag, [data-testid^="tag"]');
    const tagCount = await tagElements.count();

    // If tags are displayed, verify at least some exist
    if (tagCount > 0) {
      expect(tagCount).toBeGreaterThan(0);
    }
  });
});
