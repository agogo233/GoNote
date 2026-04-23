import { test, expect } from '@playwright/test';
import { TEST_CONFIG, login } from '../fixtures/test-helpers';

test.describe('Internationalization Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('locale list API returns available locales', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/locales`);
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    expect(data).toHaveProperty('locales');
    expect(Array.isArray(data.locales)).toBeTruthy();
    expect(data.locales.length).toBeGreaterThan(0);
    
    console.log(`Found ${data.locales.length} locales`);
  });

  test('locale list includes default locales', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/locales`);
    const data = await response.json();
    
    const localeCodes = data.locales.map((l: any) => l.code);
    console.log('Available locales:', localeCodes.join(', '));
    
    expect(localeCodes).toContain('en-US');
    expect(localeCodes).toContain('zh-CN');
  });

  test('get English locale content', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/locales/en-US`);
    expect(response.ok()).toBeTruthy();
    
    const content = await response.json();
    
    expect(content).toHaveProperty('_meta');
    expect(content._meta).toHaveProperty('code', 'en-US');
    expect(content._meta).toHaveProperty('name', 'English');
    
    expect(content).toHaveProperty('common');
    expect(content.common).toHaveProperty('save');
    expect(content.common).toHaveProperty('cancel');
    expect(content.common).toHaveProperty('delete');
    
    console.log('English locale loaded successfully');
  });

  test('get Chinese locale content', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/locales/zh-CN`);
    expect(response.ok()).toBeTruthy();
    
    const content = await response.json();
    
    expect(content).toHaveProperty('_meta');
    expect(content._meta).toHaveProperty('code', 'zh-CN');
    
    expect(content).toHaveProperty('common');
    expect(content.common.save).not.toBe('Save');
    
    console.log('Chinese locale loaded successfully');
  });

  test('invalid locale returns error', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/locales/nonexistent_XY`);
    
    expect(response.status()).toBe(404);
    
    console.log('Invalid locale correctly returns 404');
  });

  test('locale objects have required properties', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/locales`);
    const data = await response.json();
    
    for (const locale of data.locales) {
      expect(locale).toHaveProperty('code');
      expect(locale).toHaveProperty('name');
      expect(locale).toHaveProperty('flag');
      expect(typeof locale.code).toBe('string');
      expect(typeof locale.name).toBe('string');
    }
    
    console.log(`All ${data.locales.length} locales have required properties`);
  });

  test('locale content has all required sections', async ({ page }) => {
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/locales/en-US`);
    const content = await response.json();
    
    const expectedSections = [
      'app', 'common', 'sidebar', 'editor', 'notes', 'folders',
      'toolbar', 'zen_mode', 'share', 'quick_switcher', 'tags',
      'outline', 'stats', 'graph', 'templates', 'export', 'login',
      'media', 'move', 'search', 'theme', 'language', 'settings',
      'homepage', 'format', 'validation'
    ];
    
    for (const section of expectedSections) {
      expect(content).toHaveProperty(section);
    }
    
    console.log('All required sections present in locale');
  });
});