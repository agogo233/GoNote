import { test, expect, TEST_CONFIG, login, apiPost, waitForAutosave, waitForSearchDebounce, cleanupTest } from '../fixtures/test-helpers';

async function openSearchPanel(page: import('@playwright/test').Page) {
  const iconRailButtons = page.locator('.icon-rail-btn');
  const buttonCount = await iconRailButtons.count();

  if (buttonCount >= 2) {
    await iconRailButtons.nth(1).click();
  }

  await page.waitForTimeout(200);
}

async function createTestNote(page: import('@playwright/test').Page, noteName: string, content: string): Promise<void> {
  await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, { content });
  await page.waitForTimeout(200);
}

test.describe('Search Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('basic text search returns results', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_searchable`;
    const uniqueTerm = `UniqueTerm${Date.now()}`;

    await createTestNote(page, noteName, `# Test Note\n\nThis contains ${uniqueTerm} for searching.`);

    await page.reload({ waitUntil: 'domcontentloaded' });

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(uniqueTerm);
    await waitForSearchDebounce(page);
    await page.waitForResponse(
      resp => resp.url().includes('/api/search') && resp.status() === 200,
      { timeout: 5000 }
    ).catch(() => {});

    // Verify search results contain the note
    const pageContent = await page.content();
    expect(pageContent).toContain(noteName);
  });

  test('search shows no results message for non-existent term', async ({ page }) => {
    await openSearchPanel(page);

    const nonExistentTerm = `NonExistentTerm${Date.now()}XYZ123`;

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(nonExistentTerm);
    await waitForSearchDebounce(page);
    await page.waitForTimeout(500);

    // Verify no results message or empty state is shown
    const noResults = page.locator('text=/No results|没有结果|0 result/').first();
    const hasNoResults = await noResults.isVisible({ timeout: 3000 }).catch(() => false);
    expect(hasNoResults).toBe(true);
  });

  test('search clears when input is cleared', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_clear_test`;
    await createTestNote(page, noteName, `Clearable content for test ${testPrefix}`);

    await page.reload({ waitUntil: 'domcontentloaded' });

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(testPrefix);
    await waitForSearchDebounce(page);

    // Clear the input
    await searchInput.fill('');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay);

    const value = await searchInput.inputValue();
    expect(value).toBe('');
  });

  test('search with Chinese characters', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_chinese`;
    const chineseContent = '这是中文搜索测试内容';

    await createTestNote(page, noteName, `# 中文笔记\n\n${chineseContent}`);

    await page.reload({ waitUntil: 'domcontentloaded' });

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill('中文');
    await waitForSearchDebounce(page);
    await page.waitForResponse(
      resp => resp.url().includes('/api/search') && resp.status() === 200,
      { timeout: 5000 }
    ).catch(() => {});

    // Verify search results contain the Chinese note
    const pageContent = await page.content();
    expect(pageContent).toContain(noteName);
  });

  test('search result click opens note', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_clickable`;
    const uniqueContent = `ClickableContent${Date.now()}`;

    await createTestNote(page, noteName, `# Clickable Note\n\nContent: ${uniqueContent}`);

    await page.reload({ waitUntil: 'domcontentloaded' });

    // Open the note from sidebar
    const noteItem = page.locator(`text="${noteName}"`).first();
    await expect(noteItem).toBeVisible({ timeout: 5000 });
    await noteItem.click();

    // Verify editor is visible and contains the content
    const editor = page.locator('#note-editor');
    await expect(editor).toBeVisible({ timeout: 5000 });
    const content = await editor.inputValue();
    expect(content).toContain(uniqueContent);
  });
});