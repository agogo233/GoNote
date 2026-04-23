import { test, expect, TEST_CONFIG, login } from '../fixtures/test-helpers';

test.describe('Theme Switching Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('theme list API returns available themes', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/themes`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    expect(data).toHaveProperty('themes');
    expect(Array.isArray(data.themes)).toBe(true);
    expect(data.themes.length).toBeGreaterThan(0);
    
    console.log(`Found ${data.themes.length} themes`);
  });

  test('theme list includes default themes', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/themes`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    const themeIds = data.themes.map((t: { id: string }) => t.id);
    
    expect(themeIds).toContain('light');
    expect(themeIds).toContain('dark');
    
    console.log(`Available themes: ${themeIds.join(', ')}`);
  });

  test('get light theme CSS', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/themes/light`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    expect(data).toHaveProperty('css');
    expect(data.theme_id || data.themeId).toBe('light');
    expect(data.css).toBeTruthy();
    expect(data.css.length).toBeGreaterThan(0);
    
    console.log(`Light theme CSS length: ${data.css.length} characters`);
  });

  test('get dark theme CSS', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/themes/dark`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    expect(data).toHaveProperty('css');
    expect(data.theme_id || data.themeId).toBe('dark');
    expect(data.css).toBeTruthy();
    
    console.log(`Dark theme CSS length: ${data.css.length} characters`);
  });

  test('theme CSS contains expected variables', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/themes/light`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    const css = data.css;
    
    expect(css).toContain('--bg-primary');
    expect(css).toContain('--text-primary');
    
    console.log('Theme CSS contains expected CSS variables');
  });

  test('invalid theme returns error or empty CSS', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/themes/nonexistent_theme_xyz`);
    
    const status = response.status();
    expect([200, 404, 500]).toContain(status);
    
    if (status === 200) {
      const data = await response.json();
      expect(data.css === '' || data.css === null).toBe(true);
    }
    
    console.log(`Invalid theme request returned status: ${status}`);
  });

  test('theme objects have required properties', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/themes`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    const themes = data.themes;
    
    for (const theme of themes) {
      expect(theme).toHaveProperty('id');
      expect(theme).toHaveProperty('name');
      expect(typeof theme.id).toBe('string');
      expect(typeof theme.name).toBe('string');
    }
    
    console.log(`All ${themes.length} themes have required properties`);
  });
});
