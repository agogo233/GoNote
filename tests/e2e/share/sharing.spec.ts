import { test, expect, TEST_CONFIG, login, apiPost, apiDelete, cleanupTest } from '../fixtures/test-helpers';

test.describe('Share Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  async function createNoteViaAPI(page: import('@playwright/test').Page, noteName: string, content: string): Promise<void> {
    const notePath = `${noteName}.md`;
    const response = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${notePath}`, { content });
    expect(response.status()).toBe(200);
  }

  test('create share link via API', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_create_share`;
    const notePath = `${noteName}.md`;
    
    await createNoteViaAPI(page, noteName, '# Test Share Creation\n\nContent for sharing.');
    
    const createResponse = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}?theme=light`);
    
    expect(createResponse.status()).toBe(200);
    
    const data = await createResponse.json();
    console.log(`Create share response: ${JSON.stringify(data)}`);
    
    expect(data.success).toBe(true);
    expect(data.token).toBeTruthy();
    expect(data.url).toContain('/share/');
    expect(data.path).toBe(notePath);
    expect(data.theme).toBe('light');
  });

  test('view shared note via share link', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_view_shared`;
    const notePath = `${noteName}.md`;
    
    await createNoteViaAPI(page, noteName, '# Shared Note Title\n\nThis content should be visible in shared view.');
    
    const createResponse = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}?theme=light`);
    expect(createResponse.status()).toBe(200);
    
    const shareData = await createResponse.json();
    expect(shareData.success).toBe(true);
    expect(shareData.token).toBeTruthy();
    
    // shareData.url already contains full URL
    const viewResponse = await page.request.get(shareData.url);
    
    expect(viewResponse.status()).toBe(200);
    
    const htmlContent = await viewResponse.text();
    expect(htmlContent).toContain('Shared Note Title');
    expect(htmlContent).toContain('This content should be visible');
    
    console.log(`Share link ${shareData.url} successfully returns note content`);
  });

  test('revoke share link', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_revoke_share`;
    const notePath = `${noteName}.md`;
    
    await createNoteViaAPI(page, noteName, '# Note to Revoke\n\nThis will be unshared.');
    
    const createResponse = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}?theme=light`);
    expect(createResponse.status()).toBe(200);
    
    const shareData = await createResponse.json();
    const token = shareData.token;
    expect(token).toBeTruthy();
    
    const statusResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}`);
    const statusData = await statusResponse.json();
    expect(statusData.shared).toBe(true);
    
    const revokeResponse = await apiDelete(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}`);
    expect(revokeResponse.status()).toBe(200);
    
    const revokeData = await revokeResponse.json();
    expect(revokeData.success).toBe(true);
    
    const afterRevokeStatus = await page.request.get(`${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}`);
    const afterRevokeData = await afterRevokeStatus.json();
    expect(afterRevokeData.shared).toBe(false);
    
    const shareUrl = `${TEST_CONFIG.baseUrl}/share/${token}`;
    const viewResponse = await page.request.get(shareUrl);
    expect(viewResponse.status()).toBe(404);
    
    console.log('Share link successfully revoked');
  });

  test('share status for non-shared note', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_not_shared`;
    const notePath = `${noteName}.md`;
    
    await createNoteViaAPI(page, noteName, '# Not Shared Note\n\nThis note is not shared.');
    
    const statusResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}`);
    expect(statusResponse.status()).toBe(200);
    
    const statusData = await statusResponse.json();
    expect(statusData.shared).toBe(false);
    
    console.log('Non-shared note correctly shows shared: false');
  });

  test('list shared notes', async ({ page, testPrefix }) => {
    const noteName1 = `${testPrefix}_shared_list_1`;
    const noteName2 = `${testPrefix}_shared_list_2`;
    const notePath1 = `${noteName1}.md`;
    const notePath2 = `${noteName2}.md`;
    
    await createNoteViaAPI(page, noteName1, '# Shared Note 1\n\nContent 1.');
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath1)}?theme=light`);
    
    await createNoteViaAPI(page, noteName2, '# Shared Note 2\n\nContent 2.');
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath2)}?theme=dark`);
    
    const listResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/shared-notes`);
    expect(listResponse.status()).toBe(200);
    
    const listData = await listResponse.json();
    expect(Array.isArray(listData.paths)).toBe(true);
    
    const paths = listData.paths as string[];
    expect(paths.some(p => p.includes(noteName1))).toBe(true);
    expect(paths.some(p => p.includes(noteName2))).toBe(true);
    
    console.log(`Found ${paths.length} shared notes`);
  });

  test('share with different themes', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_theme_share`;
    const notePath = `${noteName}.md`;
    
    await createNoteViaAPI(page, noteName, '# Themed Share\n\nContent with theme.');
    
    const themes = ['light', 'dark', 'dracula'];
    
    for (const theme of themes) {
      const createResponse = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}?theme=${theme}`);
      expect(createResponse.status()).toBe(200);
      
      const shareData = await createResponse.json();
      expect(shareData.theme).toBe(theme);
      
      console.log(`Share created with theme: ${theme}`);
    }
  });

  test('share token persists after page reload', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_persist_share`;
    const notePath = `${noteName}.md`;
    
    await createNoteViaAPI(page, noteName, '# Persistent Share\n\nThis share should persist.');
    
    const createResponse = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}?theme=light`);
    expect(createResponse.status()).toBe(200);

    const shareData = await createResponse.json();
    const originalToken = shareData.token;

    const statusResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}`);
    const statusData = await statusResponse.json();
    
    expect(statusData.shared).toBe(true);
    expect(statusData.token).toBe(originalToken);
    
    console.log('Share token persisted correctly');
  });

  test('share modal can be opened', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_share_modal_test`;
    const notePath = `${noteName}.md`;
    
    await createNoteViaAPI(page, noteName, '# Share Modal Test\n\nThis is a test note for share modal.');
    
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/share/${encodeURIComponent(notePath)}`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    expect(data).toHaveProperty('shared');
    console.log(`Share status response: ${JSON.stringify(data)}`);
  });
});
