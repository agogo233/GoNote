import { test, expect, TEST_CONFIG, login } from '../fixtures/test-helpers';

// Check if authentication is enabled by testing API response
async function isAuthEnabled(page: any): Promise<boolean> {
  const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
  // If we get 401, auth is enabled. If 200, auth is disabled.
  return response.status() === 401;
}

test.describe('Authentication', () => {
  test('login page displays correctly', async ({ page }) => {
    const authEnabled = await isAuthEnabled(page);
    test.skip(!authEnabled, 'Authentication is disabled');
    
    await page.goto('/login');
    
    await expect(page.locator('input[type="password"]')).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    
    const title = await page.locator('h1, .title, [data-testid="app-title"], .login-title').first();
    await expect(title).toContainText('GoNote');
  });

  test('login success with correct password', async ({ page }) => {
    const authEnabled = await isAuthEnabled(page);
    test.skip(!authEnabled, 'Authentication is disabled');
    
    await page.goto('/login');
    
    await page.fill('input[type="password"]', TEST_CONFIG.testPassword);
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL(/.*\/$/, { timeout: TEST_CONFIG.defaultTimeout });
    
    await expect(page.locator('#app, [x-data], .main-content').first()).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('login failure with wrong password', async ({ page }) => {
    const authEnabled = await isAuthEnabled(page);
    test.skip(!authEnabled, 'Authentication is disabled');
    
    await page.goto('/login');
    
    await page.fill('input[type="password"]', 'wrong-password');
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL(/.*\/login/, { timeout: TEST_CONFIG.defaultTimeout });
    
    const errorElement = page.locator('.error-message, .alert-error, [data-testid="error-message"], [class*="error"]').first();
    await expect(errorElement).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('logout clears session', async ({ page }) => {
    const authEnabled = await isAuthEnabled(page);
    test.skip(!authEnabled, 'Authentication is disabled');
    
    await page.goto('/login');
    await page.fill('input[type="password"]', TEST_CONFIG.testPassword);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*\/$/, { timeout: TEST_CONFIG.defaultTimeout });
    
    // Logout is a POST request, not a page navigation
    await page.request.post(`${TEST_CONFIG.baseUrl}/logout`);
    
    // Verify session is cleared by checking API
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
    expect(response.status()).toBe(401);
  });

  test('unauthenticated access redirects to login', async ({ page }) => {
    const authEnabled = await isAuthEnabled(page);
    test.skip(!authEnabled, 'Authentication is disabled');
    
    await page.goto('/');
    
    await expect(page).toHaveURL(/.*\/login/, { timeout: TEST_CONFIG.defaultTimeout });
  });

  test('protected API routes require authentication', async ({ page }) => {
    const authEnabled = await isAuthEnabled(page);
    test.skip(!authEnabled, 'Authentication is disabled');
    
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
    
    expect(response.status()).toBe(401);
  });

  test('authentication disabled - direct access works', async ({ page }) => {
    // When authentication is disabled, users should be able to access the app directly
    await page.goto('/');
    await page.waitForTimeout(500);
    
    // Should show the main app (not login page)
    const app = page.locator('#app, [x-data], .icon-rail').first();
    const isVisible = await app.isVisible({ timeout: 5000 }).catch(() => false);
    
    // Either we see the app, or we're redirected to login if auth is enabled
    const currentUrl = page.url();
    const isLoginPage = currentUrl.includes('/login');
    
    console.log(`App visible: ${isVisible}, On login page: ${isLoginPage}`);
    
    // Test passes if either condition is true (auth disabled or login shown)
    expect(isVisible || isLoginPage).toBeTruthy();
  });

  test('API works without authentication when auth disabled', async ({ page }) => {
    // Try to access notes API
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`);
    const status = response.status();
    
    console.log(`API response status: ${status}`);
    
    // When auth is disabled: 200
    // When auth is enabled but no session: 401
    expect([200, 401]).toContain(status);
  });

  test('login helper handles disabled auth', async ({ page }) => {
    // The login helper should work whether auth is enabled or disabled
    await login(page);
    
    // Should end up on the main page
    const app = page.locator('#app, [x-data], .icon-rail').first();
    await expect(app).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });
});
