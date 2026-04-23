import { test, expect, TEST_CONFIG, login, waitForAutosave, cleanupTest } from '../fixtures/test-helpers';

/**
 * Export tests - focused on HTML rendering quality
 * Share link functionality is tested in share/sharing.spec.ts
 */
test.describe('Export HTML Rendering', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  /**
   * Helper: create note via UI and return editor locator
   */
  async function createNoteViaUI(page: any, noteName: string, content: string) {
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();

    const newNoteOption = page.locator('button:has-text("📝")').first();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    page.once('dialog', async dialog => {
      await dialog.accept(noteName);
    });

    await newNoteOption.click({ force: true });

    // Wait for note creation API response
    await page.waitForResponse(
      resp => resp.url().includes('/api/notes/') && [200, 201].includes(resp.status()),
      { timeout: 10000 }
    ).catch(() => {});

    const editor = page.locator('#note-editor').first();
    await editor.fill(content);
    await waitForAutosave(page);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    return await shareResponse.json();
  }

  test('shared note renders as valid HTML document', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_html_render`;
    const shareData = await createNoteViaUI(page, noteName, '# Shared Content\n\nThis is a shared paragraph.');

    const viewResponse = await page.request.get(shareData.url);
    expect(viewResponse.status()).toBe(200);

    const html = await viewResponse.text();
    expect(html).toContain('<!DOCTYPE html>');
    expect(html).toContain('Shared Content');
    expect(html).toContain('This is a shared paragraph');
  });

  test('shared note includes code highlighting', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_code_share`;
    const shareData = await createNoteViaUI(page, noteName, '# Code Test\n\n```python\ndef hello():\n    print("world")\n```');

    const viewResponse = await page.request.get(shareData.url);
    expect(viewResponse.status()).toBe(200);

    const html = await viewResponse.text();
    expect(html).toContain('highlight.js');
  });

  test('shared note includes MathJax for math rendering', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_math_share`;
    const shareData = await createNoteViaUI(page, noteName, '# Math Test\n\n$$E = mc^2$$');

    const viewResponse = await page.request.get(shareData.url);
    expect(viewResponse.status()).toBe(200);

    const html = await viewResponse.text();
    expect(html).toContain('MathJax');
  });

  test('shared note with theme parameter renders correctly', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_theme_share`;
    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    const shareData = await createNoteViaUI(page, noteName, '# Themed Content\n\nContent with theme.');

    // Test dark theme
    const darkUrl = `${TEST_CONFIG.baseUrl}/share/${shareData.token}?theme=dark`;
    const darkResponse = await page.request.get(darkUrl);
    expect(darkResponse.status()).toBe(200);

    const darkHtml = await darkResponse.text();
    expect(darkHtml).toContain('Themed Content');
  });
});
