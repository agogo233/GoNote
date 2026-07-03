import { test, expect, TEST_CONFIG, login, apiPost, cleanupTest } from '../fixtures/test-helpers';

test.describe('XSS Sanitization (S-11)', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  async function createNoteAndOpenPreview(page: import('@playwright/test').Page, noteName: string, content: string) {
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, { content });
    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${noteName}"`).first();
    await noteItem.click();
    await page.waitForTimeout(300);

    // Switch to preview mode
    const previewButton = page.locator('button').filter({ hasText: /^Preview$/ }).first();
    if (await previewButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await previewButton.click();
      await page.waitForTimeout(300);
    }

    const preview = page.locator('.markdown-preview');
    await expect(preview).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
    // Allow marked + DOMPurify + debounce to settle
    await page.waitForTimeout(500);
    return preview;
  }

  test('script tag is stripped from note content', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_xss_script`;
    const payload = '# Safe Title\n\n<script>window.__xss_fired=true;</script>\n\nNormal text.';

    let dialogFired = false;
    page.on('dialog', () => { dialogFired = true; });

    await createNoteAndOpenPreview(page, noteName, payload);

    expect(dialogFired).toBe(false);

    // No script element should remain in the preview
    const scriptCount = await page.locator('.markdown-preview script').count();
    expect(scriptCount).toBe(0);

    // Payload JS must not have leaked into window
    const leaked = await page.evaluate(() => (window as any).__xss_fired === true);
    expect(leaked).toBe(false);
  });

  test('img onerror handler is neutralized', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_xss_img_onerror`;
    const payload = '# Img Onerror\n\n<img src="x" onerror="window.__xss_img_fired=true">';

    let dialogFired = false;
    page.on('dialog', () => { dialogFired = true; });

    await createNoteAndOpenPreview(page, noteName, payload);

    expect(dialogFired).toBe(false);

    const leaked = await page.evaluate(() => (window as any).__xss_img_fired === true);
    expect(leaked).toBe(false);

    // onerror attribute must have been stripped from any img surviving
    const badImg = await page.locator('.markdown-preview img[onerror]').count();
    expect(badImg).toBe(0);
  });

  test('iframe with javascript: scheme is stripped', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_xss_iframe_js`;
    const payload = '# Iframe JS\n\n<iframe src="javascript:window.__xss_iframe_fired=true"></iframe>';

    let dialogFired = false;
    page.on('dialog', () => { dialogFired = true; });

    await createNoteAndOpenPreview(page, noteName, payload);

    expect(dialogFired).toBe(false);

    const leaked = await page.evaluate(() => (window as any).__xss_iframe_fired === true);
    expect(leaked).toBe(false);

    // No iframe carrying a javascript: src should survive sanitization
    const jsIframe = await page.locator('.markdown-preview iframe[src^="javascript:"]').count();
    expect(jsIframe).toBe(0);
  });
});
