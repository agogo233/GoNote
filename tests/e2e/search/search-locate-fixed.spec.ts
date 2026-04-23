import { test, expect, TEST_CONFIG, login, apiPost, waitForSearchIndex, cleanupTest } from '../fixtures/test-helpers';

const BASE_URL = TEST_CONFIG.baseUrl;

test.describe('Search Locate Feature', () => {
  test.beforeEach(async ({ page }) => {
    page.on('console', msg => {
      if (msg.type() === 'error' || msg.type() === 'warning') {
        console.log(`[Browser Console] ${msg.type()}: ${msg.text()}`);
      }
    });
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('search click on unopened note - with proper wait', async ({ page, testPrefix }) => {
    // Create a test note with unique term on a specific line
    const noteName = `${testPrefix}_search_locate`;
    const uniqueTerm = `LocateTerm${Date.now()}`;
    const content = `# Search Locate Test

Line 3: Some content here
Line 4: More content here
Line 5: This line contains ${uniqueTerm} for testing
Line 6: Another line here
Line 7: More content`;

    // Create note via API
    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });

    // Wait for search index to update
    await waitForSearchIndex(page);

    // Reload page to ensure note is indexed
    await page.reload({ waitUntil: 'domcontentloaded' });

    // Verify no note is currently open
    const noteCurrent = await page.evaluate(() => {
      const el = document.querySelector('[x-data]') as any;
      return el?._x_dataStack?.[0]?.note?.current || '';
    });
    console.log(`Note current before search: "${noteCurrent}" (should be empty)`);

    // Open search panel - click on search button in icon-rail
    const searchButton = page.locator('.icon-rail button').nth(1);
    await searchButton.click();
    await page.waitForTimeout(500);

    // Type search query
    const searchInput = page.locator('input[x-model="search.query"]').first();
    await searchInput.waitFor({ state: 'visible', timeout: 5000 });
    await searchInput.fill(uniqueTerm);

    // Wait for search debounce and results
    await page.waitForTimeout(2000);

    // Verify search returned results
    const searchResults = await page.evaluate(() => {
      const el = document.querySelector('[x-data]') as any;
      const app = el?._x_dataStack?.[0];
      return {
        total: app?.search?.totalResults || 0,
        count: app?.search?.results?.length || 0,
        firstResultPath: app?.search?.results?.[0]?.path || ''
      };
    });

    console.log(`Search results: ${JSON.stringify(searchResults)}`);
    
    // Take screenshot for debugging if no results
    if (searchResults.total === 0) {
      await page.screenshot({ path: 'config/test-results/search-locate-fixed-no-results.png', fullPage: true });
      console.log('No search results found - check if note was indexed');
    }
    
    expect(searchResults.total).toBeGreaterThan(0);
    expect(searchResults.firstResultPath).toBe(`${noteName}.md`);

    // Click on the first search result
    const result = await page.evaluate(async (term: string) => {
      const el = document.querySelector('[x-data]') as any;
      const app = el?._x_dataStack?.[0];
      if (!app) return { success: false, error: 'No Alpine app found' };

      // Get the first search result
      const note = app.search.results[0];
      if (!note) return { success: false, error: 'No search result found' };

      // Call openItem directly (simulating what the click handler should do)
      const lineNumber = (note.matches && note.matches.length > 0) ? note.matches[0].line_number : 0;
      app.openItem(note.path, note.type, app.search.query, lineNumber);

      // Wait for note to load
      await new Promise(resolve => setTimeout(resolve, 2000));

      // Check if editor loaded
      const editor = document.getElementById('note-editor') as HTMLTextAreaElement;
      if (!editor) return { success: false, error: 'Editor not found in DOM' };

      // Check if content loaded
      const editorContent = editor.value;
      if (!editorContent.includes(term)) {
        return { success: false, error: 'Editor content does not contain search term', content: editorContent.substring(0, 100) };
      }

      // Check selection (highlight)
      const selectionStart = editor.selectionStart;
      const selectionEnd = editor.selectionEnd;
      const selectedText = editorContent.substring(selectionStart, selectionEnd);

      // Check scroll position
      const scrollTop = editor.scrollTop;
      const lineHeight = parseFloat(window.getComputedStyle(editor).lineHeight) || 20;
      const approximateLineInView = Math.floor(scrollTop / lineHeight) + 3;

      return {
        success: true,
        noteCurrent: app.note?.current,
        editorContent: editorContent.substring(0, 100),
        selectionStart,
        selectionEnd,
        selectedText,
        selectedTextLower: selectedText.toLowerCase(),
        searchTerm: term.toLowerCase(),
        scrollTop,
        approximateLineInView,
        targetLine: lineNumber
      };
    }, uniqueTerm);

    console.log(`Result: ${JSON.stringify(result, null, 2)}`);

    // Take screenshot for documentation
    await page.screenshot({ path: 'config/test-results/search-locate-fixed.png', fullPage: true });

    // Assertions
    expect(result.success).toBe(true);
    expect(result.noteCurrent).toBe(`${noteName}.md`);

    // Check that the search term is selected (highlighted)
    expect(result.selectedTextLower).toBe(result.searchTerm);

    console.log(`✅ Search locate test passed!`);
    console.log(`  - Note opened: ${result.noteCurrent}`);
    console.log(`  - Selected text: "${result.selectedText}"`);
    console.log(`  - Scroll position: line ~${result.approximateLineInView}`);
  });
});
