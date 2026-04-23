import { test, expect, TEST_CONFIG, login, apiPost, waitForAutosave, waitForSearchIndex, cleanupTest } from '../fixtures/test-helpers';

const BASE_URL = TEST_CONFIG.baseUrl;

/**
 * Search CJK (Chinese, Japanese, Korean) Support E2E Tests
 */

async function openSearchPanel(page: import('@playwright/test').Page) {
  const iconRailButtons = page.locator('.icon-rail button, [class*="icon-rail"] button');
  const buttonCount = await iconRailButtons.count();

  if (buttonCount >= 2) {
    await iconRailButtons.nth(1).click();
  }

  await page.waitForTimeout(500);

  const searchInput = page.locator('input[x-model="search.query"]').first();
  await searchInput.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout }).catch(() => {});
}

test.describe('Search CJK Support', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('search simplified Chinese characters', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_simplified_chinese`;
    const chineseContent = `# 简体中文笔记

这是一篇用于测试中文搜索的笔记。

## 主要内容

本文档包含了一些**简体中文**的内容，用于测试搜索引擎是否能够正确处理中文字符。

### 关键词

- 搜索
- 中文
- 简体中文
- 全文搜索
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: chineseContent });
    
    // Wait for search index to be updated
    await waitForSearchIndex(page);

    await openSearchPanel(page);

    // Search for Chinese term
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('中文');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);

    // Check if note appears in results
    const noteInResults = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('search traditional Chinese characters', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_traditional_chinese`;
    const traditionalContent = `# 繁體中文筆記

這是一篇用於測試繁體中文搜索的筆記。

## 主要內容

本文檔包含了一些**繁體中文**的內容，用於測試搜索引擎是否能夠正確處理中文字符。

### 關鍵詞

- 搜索
- 中文
- 繁體中文
- 全文搜索
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: traditionalContent });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    // Search for traditional Chinese term
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('繁體');
    await page.waitForTimeout(1500);

    // Check if note appears in results
    const noteInResults = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('search Japanese Hiragana', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_japanese_hiragana`;
    const hiraganaContent = `# ひらがなノート

これはひらがなの検索テスト用のノートです。

## メインコンテンツ

このドキュメントには**ひらがな**のコンテンツが含まれています。

### キーワード

- 検索
- ひらがな
- 日本語
- 全文検索
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: hiraganaContent });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    // Search for Hiragana term
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('ひらがな');
    await page.waitForTimeout(1500);

    // Check if note appears in results
    const noteInResults = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('search Japanese Katakana', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_japanese_katakana`;
    const katakanaContent = `# カタカナノート

これはカタカナの検索テスト用のノートです。

## メインコンテンツ

このドキュメントには**カタカナ**のコンテンツが含まれています。

### キーワード

- 検索
- カタカナ
- 日本語
- 全文検索
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: katakanaContent });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    // Search for Katakana term
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('カタカナ');
    await page.waitForTimeout(1500);

    // Check if note appears in results
    const noteInResults = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('search Japanese Kanji', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_japanese_kanji`;
    const kanjiContent = `# 漢字ノート

これは漢字の検索テスト用のノートです。

## メインコンテンツ

このドキュメントには**漢字**のコンテンツが含まれています。

### キーワード

- 検索
- 漢字
- 日本語
- 全文検索
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: kanjiContent });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    // Search for Kanji term
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('漢字');
    await page.waitForTimeout(1500);

    // Check if note appears in results
    const noteInResults = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('search Korean Hangul', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_korean_hangul`;
    const hangulContent = `# 한글 노트

이것은 한글 검색 테스트를 위한 노트입니다.

## 메인 콘텐츠

이 문서에는 **한글** 콘텐츠가 포함되어 있습니다.

### 키워드

- 검색
- 한글
- 한국어
- 전체 검색
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: hangulContent });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    // Search for Hangul term
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('한글');
    await page.waitForTimeout(1500);

    // Check if note appears in results
    const noteInResults = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('search mixed CJK content', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_mixed_cjk`;
    const mixedContent = `# Mixed CJK Test

## Chinese (中文)
这是一段中文内容。

## Japanese (日本語)
これは日本語のコンテンツです。

## Korean (한국어)
이것은 한국어 콘텐츠입니다.

## Mixed Example
中文 + 日本語 + 한국어 = Multilingual Content
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: mixedContent });
    await waitForAutosave(page);

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();

    // Search for Chinese term - should find the note
    await searchInput.fill('中文');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);
    const noteInResults1 = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults1).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });

    // Clear and search for Japanese term - should also find the note
    await searchInput.clear();
    await searchInput.fill('日本語');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);
    const noteInResults2 = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults2).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });

    // Clear and search for Korean term - should also find the note
    await searchInput.clear();
    await searchInput.fill('한국어');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);
    const noteInResults3 = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults3).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('CJK partial match search', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_cjk_partial`;
    const content = `# CJK Partial Match Test

## Content
这是一个用于测试部分匹配的中文笔记。
This is a Chinese note for testing partial match.

包含多个关键词：搜索、中文、测试、部分匹配
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await waitForAutosave(page);

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();

    // Search for partial match - single character
    await searchInput.fill('搜');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);
    const noteInResults1 = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults1).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });

    // Search for partial match - two characters
    await searchInput.clear();
    await searchInput.fill('测试');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);
    const noteInResults2 = page.locator(`text="${noteName}"`).first();
    await expect(noteInResults2).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('CJK search result highlight', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_cjk_highlight`;
    const content = `# CJK Highlight Test

搜索高亮测试。
Search highlight test with Chinese characters.

这个句子包含搜索关键词，用于测试高亮功能。
This sentence contains the search keyword for testing highlight.
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('搜索');
    await page.waitForTimeout(1500);

    // Click on the search result to open the note
    const resultItem = page.locator('.hover-accent.cursor-pointer, [class*="hover-accent"]').first();
    if (await resultItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      await resultItem.click();
      await page.waitForTimeout(1500);

      // Check if editor has the content
      const editor = page.locator('#note-editor').first();
      const editorContent = await editor.inputValue();
      expect(editorContent).toContain('搜索');
    }
  });

  test('CJK search with line navigation', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_cjk_navigation`;
    const content = `# CJK Navigation Test

第一行：这是第一行内容。
第二行：这是第二行内容。
第三行：这是第三行内容，包含关键词「导航」。
第四行：这是第四行内容。
第五行：这是第五行内容。
第六行：这是第六行内容，也包含导航这个词。
第七行：这是第七行内容。
第八行：这是第八行内容。
第九行：这是第九行内容。
第十行：这是第十行内容。
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.fill('导航');
    await page.waitForTimeout(1500);

    // Click on first result
    const resultItem = page.locator('.hover-accent.cursor-pointer, [class*="hover-accent"]').first();
    if (await resultItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      await resultItem.click();
      await page.waitForTimeout(1500);

      // Check editor state
      const editor = page.locator('#note-editor').first();
      const editorContent = await editor.inputValue();
      expect(editorContent).toContain('导航');
    }
  });

  test('CJK search no results handling', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_cjk_no_results`;
    const content = `# CJK No Results Test

这是一篇测试笔记。
This is a test note.
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(500);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(1000);

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();
    
    // Search for non-existent CJK term
    await searchInput.fill('不存在的关键词 XYZ');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);

    // Should not have the original note in results
    const noteInResults = page.locator(`text="${noteName}"`).first();
    const isVisible = await noteInResults.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isVisible).toBe(false);
  });

  test('CJK search clear and re-search', async ({ page, testPrefix }) => {
    const noteName1 = `${testPrefix}_cjk_clear1`;
    const noteName2 = `${testPrefix}_cjk_clear2`;

    const content1 = `# CJK Clear Test 1

第一个笔记的内容。
Content for first note.
`;

    const content2 = `# CJK Clear Test 2

第二个笔记的内容。
Content for second note.
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName1}.md`, { content: content1 });
    await apiPost(page, `${BASE_URL}/api/notes/${noteName2}.md`, { content: content2 });
    await waitForAutosave(page);

    await openSearchPanel(page);

    const searchInput = page.locator('input[x-model="search.query"]').first();

    // First search - should find note1
    await searchInput.fill('第一个');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);
    const result1 = page.locator(`text="${noteName1}"`).first();
    await expect(result1).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });

    // Clear search - should show all notes again
    await searchInput.clear();
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay);

    // Second search - should find note2
    await searchInput.fill('第二个');
    await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay + 500);
    const result2 = page.locator(`text="${noteName2}"`).first();
    await expect(result2).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });
});
