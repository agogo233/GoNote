import { test, expect, TEST_CONFIG, login } from '../fixtures/test-helpers';

test.describe('Error Sanitization', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('error responses do not leak file system paths', async ({ page }) => {
    const response = await page.request.get('/api/notes/nonexistent_note_xyz.md');

    // Should return 404
    expect(response.status()).toBe(404);

    const body = await response.json();
    const responseBody = JSON.stringify(body);

    // Should NOT contain absolute file paths
    expect(responseBody).not.toMatch(/\/[a-zA-Z]+\/[a-zA-Z]+\/.*\//);
    expect(responseBody).not.toMatch(/[A-Z]:\\/);
    expect(responseBody).not.toContain(process.cwd());
  });

  test('error responses do not leak stack traces', async ({ page }) => {
    const response = await page.request.get('/api/notes/nonexistent_note_xyz.md');

    expect(response.status()).toBe(404);

    const body = await response.json();
    const responseBody = JSON.stringify(body);

    // Should NOT contain stack trace indicators
    expect(responseBody).not.toContain('stack');
    expect(responseBody).not.toContain('Stack');
    expect(responseBody).not.toContain('goroutine');
    expect(responseBody).not.toContain('at line');
    expect(responseBody).not.toContain(':'); // avoid line numbers in stack traces
  });

  test('error responses do not leak database queries', async ({ page }) => {
    const response = await page.request.get('/api/notes/nonexistent_note_xyz.md');

    expect(response.status()).toBe(404);

    const body = await response.json();
    const responseBody = JSON.stringify(body);

    // Should NOT contain SQL or query patterns
    expect(responseBody).not.toContain('SELECT');
    expect(responseBody).not.toContain('INSERT');
    expect(responseBody).not.toContain('WHERE');
    expect(responseBody).not.toContain('table');
  });

  test('error responses do not leak internal configuration', async ({ page }) => {
    const response = await page.request.get('/api/notes/nonexistent_note_xyz.md');

    expect(response.status()).toBe(404);

    const body = await response.json();
    const responseBody = JSON.stringify(body);

    // Should NOT contain sensitive config keys
    expect(responseBody).not.toContain('secret_key');
    expect(responseBody).not.toContain('password');
    expect(responseBody).not.toContain('auth_token');
    expect(responseBody).not.toContain('api_key');
  });

  test('invalid JSON body returns generic error', async ({ page, testPrefix }) => {
    const csrfToken = await page.evaluate(async () => {
      const cookies = document.cookie.split(';');
      const csrf = cookies.find(c => c.trim().startsWith('csrf'));
      return csrf ? csrf.split('=')[1] : '';
    });

    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/test.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: 'not valid json{{{',
    });

    // Should return 400 for invalid JSON
    expect(response.status()).toBe(400);

    const body = await response.json();
    const responseBody = JSON.stringify(body);

    // Error message should be generic, not contain parser internals
    expect(responseBody).not.toContain('stack');
    expect(responseBody).not.toContain('goroutine');
  });

  test('oversized request body returns appropriate error', async ({ page, testPrefix }) => {
    const csrfToken = await page.evaluate(async () => {
      const cookies = document.cookie.split(';');
      const csrf = cookies.find(c => c.trim().startsWith('csrf'));
      return csrf ? csrf.split('=')[1] : '';
    });

    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // Create a very large content (1MB+)
    const largeContent = 'A'.repeat(2 * 1024 * 1024);

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/large_${testPrefix}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ content: largeContent }),
    });

    // Should reject oversized requests
    const status = response.status();
    expect([400, 413, 500]).toContain(status);

    // Even on error, should not leak internals
    if (status >= 400) {
      const body = await response.json().catch(() => ({}));
      const responseBody = JSON.stringify(body);
      expect(responseBody).not.toContain('stack');
      expect(responseBody).not.toContain('goroutine');
    }
  });

  test('malformed note path returns safe error', async ({ page }) => {
    const response = await page.request.get('/api/notes/%%invalid%%path%%.md');

    const status = response.status();
    expect([400, 404]).toContain(status);

    const body = await response.json().catch(() => ({}));
    if (body) {
      const responseBody = JSON.stringify(body);
      expect(responseBody).not.toMatch(/\/[a-zA-Z]+\/[a-zA-Z]+\/.*\//);
    }
  });
});
