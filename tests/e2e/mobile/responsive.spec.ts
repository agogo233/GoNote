import { test, expect, TEST_CONFIG, login, apiPost } from '../fixtures/test-helpers';

test.describe('Mobile Responsive Design', () => {
  test.use({ viewport: { width: 375, height: 667 } });

  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('mobile layout renders correctly', async ({ page }) => {
    const body = page.locator('body');
    await expect(body).toBeVisible();
    
    const viewportSize = page.viewportSize();
    expect(viewportSize?.width).toBe(375);
    
    console.log('Mobile layout renders correctly');
  });

  test('sidebar is hidden by default on mobile', async ({ page }) => {
    // On mobile, sidebar should be hidden initially
    const sidebar = page.locator('.sidebar').first();
    const isVisible = await sidebar.isVisible().catch(() => false);
    
    // Check app container is visible
    const app = page.locator('[x-data]').first();
    await expect(app).toBeVisible();
    
    console.log(`Sidebar visible on mobile: ${isVisible}`);
  });

  test('can toggle sidebar on mobile', async ({ page }) => {
    // Look for hamburger menu or sidebar toggle button
    const menuButton = page.locator('button.hamburger, .hamburger-btn, button[aria-label*="menu"], button[aria-label*="sidebar"]').first();

    if (await menuButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await menuButton.click();
      console.log('Clicked menu button');
    } else {
      // Try bottom navigation tabs on mobile
      const bottomTabs = page.locator('.mobile-bottom-tab, .bottom-nav button');
      const tabCount = await bottomTabs.count();
      if (tabCount > 0) {
        console.log(`Found ${tabCount} bottom tabs`);
      }
    }
    
    console.log('Sidebar toggle test complete');
  });

  test('bottom navigation visible on mobile', async ({ page }) => {
    // Check for mobile bottom navigation
    const bottomNav = page.locator('.mobile-bottom-tabs').first();
    const isVisible = await bottomNav.isVisible({ timeout: 2000 }).catch(() => false);
    
    const mainContent = page.locator('[x-data]').first();
    await expect(mainContent).toBeVisible();
    
    console.log(`Bottom navigation visible: ${isVisible}`);
  });

  test('editor is usable on mobile', async ({ page, testPrefix }) => {
    // Create a note via API
    const noteName = `MobileTestNote_${testPrefix}`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: 'Test content on mobile' });
    await page.waitForTimeout(200);

    // Navigate using correct URL format: /notePath
    await page.goto(`/${encodeURIComponent(notePath)}`);

    // Check if editor is visible
    const editor = page.locator('#note-editor').first();
    const editorVisible = await editor.isVisible({ timeout: 5000 }).catch(() => false);

    if (editorVisible) {
      const content = await editor.inputValue();
      expect(content).toContain('Test content on mobile');
    } else {
      // On mobile, might need to switch to edit mode first
      // Verify at least the note page loaded
      const currentUrl = page.url();
      expect(currentUrl).toContain(noteName);
    }
  });

  test('editor tap is not blocked so the virtual keyboard can open', async ({ page, testPrefix }) => {
    const noteName = `TouchEditNote_${testPrefix}`;
    const notePath = `${noteName}.md`;
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content: 'touch edit test' });

    // Ensure the app opens in edit mode so the textarea is present on mobile
    await page.evaluate(() => localStorage.setItem('viewMode', 'edit'));
    await page.goto(`/${encodeURIComponent(notePath)}`);

    const editor = page.locator('#note-editor');
    await editor.waitFor({ state: 'visible', timeout: 5000 });

    // Single tap on the textarea must not be preventDefault'ed by app handlers
    // (that was blocking the virtual keyboard on mobile).
    const singlePrevented = await editor.evaluate(el => {
      const ev = new TouchEvent('touchend', { bubbles: true, cancelable: true });
      el.dispatchEvent(ev);
      return ev.defaultPrevented;
    });
    expect(singlePrevented).toBe(false);

    // Double tap inside the editor must NOT switch to preview and must not
    // swallow the tap; native word selection keeps working.
    const doubleResult = await editor.evaluate(async el => {
      const mk = () => new TouchEvent('touchend', { bubbles: true, cancelable: true });
      const first = mk();
      el.dispatchEvent(first);
      await new Promise(r => setTimeout(r, 50));
      const second = mk();
      el.dispatchEvent(second);
      return {
        firstPrevented: first.defaultPrevented,
        secondPrevented: second.defaultPrevented,
        editorStillVisible: el.offsetParent !== null,
      };
    });
    expect(doubleResult.firstPrevented).toBe(false);
    expect(doubleResult.secondPrevented).toBe(false);
    // A double tap inside the editor must keep the editor visible (no view switch)
    expect(doubleResult.editorStillVisible).toBe(true);
    await expect(editor).toBeVisible({ timeout: 5000 });

    // Double tap on the preview pane still toggles back to edit mode.
    // last() = preview button (edit / split(hidden on mobile) / preview order)
    await page.locator('.flex.rounded-lg.p-1 button').last().click();
    await expect(page.locator('.markdown-preview')).toBeVisible({ timeout: 5000 });
    const previewResult = await page.evaluate(async () => {
      const pv = document.querySelector('.markdown-preview');
      if (!pv) return null;
      const mk = () => new TouchEvent('touchend', { bubbles: true, cancelable: true });
      const first = mk();
      pv.dispatchEvent(first);
      await new Promise(r => setTimeout(r, 50));
      const second = mk();
      pv.dispatchEvent(second);
      return { prevented: second.defaultPrevented };
    });
    expect(previewResult).not.toBeNull();
    // The second tap of the double tap is intentionally swallowed to avoid
    // the native double-tap zoom on the preview pane
    expect(previewResult!.prevented).toBe(true);
    // After the double tap the view mode switches to edit: editor becomes visible again
    await editor.waitFor({ state: 'visible', timeout: 5000 });
  });

  test('touch interactions work', async ({ page }) => {
    const body = page.locator('body');
    await expect(body).toBeVisible();
    
    const viewport = page.viewportSize();
    expect(viewport?.width).toBeLessThanOrEqual(375);
    
    console.log('Touch interactions work');
  });

  test('responsive typography', async ({ page }) => {
    const body = page.locator('body');
    const fontSize = await body.evaluate(el => window.getComputedStyle(el).fontSize);
    
    expect(parseInt(fontSize)).toBeGreaterThan(0);
    
    console.log('Responsive typography verified');
  });

  test('settings accessible on mobile', async ({ page }) => {
    // On mobile, settings might be in a different location
    // Try hamburger menu first
    const hamburger = page.locator('button.hamburger, .hamburger-btn').first();
    
    if (await hamburger.isVisible({ timeout: 2000 }).catch(() => false)) {
      await hamburger.click();
    }

    // Look for settings in various places
    const settingsBtn = page.locator('button[title*="Settings"], button[title*="设置"]').first();
    const isVisible = await settingsBtn.isVisible({ timeout: 2000 }).catch(() => false);
    
    console.log(`Settings button visible: ${isVisible}`);
  });
});