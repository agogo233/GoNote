import { test, expect, TEST_CONFIG, login } from '../fixtures/test-helpers';

test.describe('Security Boundary Tests', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // Path traversal with URL encoding should be blocked by server-side validation
  test('path traversal with encoded slashes is blocked', async ({ page }) => {
    const response = await page.request.get('/api/notes/..%2F..%2Fetc%2Fpasswd');
    // Server should return 400 for path traversal attempts
    expect(response.status()).toBe(400);
    const data = await response.json();
    expect(data.detail).toContain('Invalid');
  });

  test('path traversal with double encoding does not leak system files', async ({ page }) => {
    const response = await page.request.get('/api/notes/..%252F..%252Fetc%252Fpasswd');
    // Server may return 200 but should NOT leak actual /etc/passwd content
    const text = await response.text();
    // Should not contain typical /etc/passwd content
    expect(text).not.toContain('root:x:0:0');
    expect(text).not.toContain('/bin/bash');
    expect(text).not.toContain('nobody:x:');
  });

  test('accessing files outside notes directory does not leak content', async ({ page }) => {
    const response = await page.request.get('/api/notes/..%2Fconfig.yaml');
    const text = await response.text();
    // Should not contain actual config content
    expect(text).not.toContain('secret_key');
    expect(text).not.toContain('password');
  });

  test('folder creation with path traversal is blocked', async ({ page }) => {
    const response = await page.request.post('/api/folders', {
      data: { path: '../../../tmp/malicious' }
    });
    const status = response.status();
    expect([400, 403, 404]).toContain(status);
  });

  test('folder deletion with path traversal is blocked', async ({ page }) => {
    const response = await page.request.delete('/api/folders/..%2F..%2F..%2Ftmp');
    const status = response.status();
    expect([400, 403, 404]).toContain(status);
  });

  test('note creation with path traversal in filename is blocked', async ({ page }) => {
    const response = await page.request.post('/api/notes/..%2F..%2F..%2Ftmp%2Fmalicious.md', {
      data: { 
        path: '../../../tmp/malicious.md',
        content: 'malicious content'
      }
    });
    const status = response.status();
    expect([400, 403, 404]).toContain(status);
  });

  test('note deletion with path traversal is blocked', async ({ page }) => {
    const response = await page.request.delete('/api/notes/..%2F..%2F..%2Fetc%2Fpasswd');
    const status = response.status();
    expect([400, 403, 404]).toContain(status);
  });

  test('media access with path traversal is blocked', async ({ page }) => {
    const response = await page.request.get('/api/media/..%2F..%2F..%2Fetc%2Fpasswd');
    const status = response.status();
    expect([400, 403, 404]).toContain(status);
  });

  test('null byte injection is handled safely', async ({ page }) => {
    const response = await page.request.get('/api/notes/test%00.md');
    // Should not cause server error (500)
    expect(response.status()).toBeLessThan(500);
  });

  test('unicode path traversal attempts do not leak system files', async ({ page }) => {
    const response = await page.request.get('/api/notes/..%c0%af..%c0%afetc/passwd');
    // Should not cause server error and should not leak files
    expect(response.status()).toBeLessThan(500);
    const text = await response.text();
    expect(text).not.toContain('root:x:0:0');
  });

  test('API returns proper error for invalid paths', async ({ page }) => {
    const response = await page.request.get('/api/notes/..%2F..%2Finvalid');
    const status = response.status();
    expect([400, 403, 404]).toContain(status);
  });

  // When auth is disabled, API should still be accessible
  test('API is accessible when auth is disabled', async ({ page }) => {
    const response = await page.request.get('/api/notes');
    const status = response.status();
    // If auth is disabled, should return 200; if enabled, 401
    expect([200, 401]).toContain(status);
  });

  // Test that path traversal doesn't leak file contents
  test('path traversal does not expose system files', async ({ page }) => {
    const response = await page.request.get('/api/notes/..%2F..%2F..%2Fetc%2Fpasswd');
    const text = await response.text();
    // Should not contain typical /etc/passwd content
    expect(text).not.toContain('root:x:0:0');
    expect(text).not.toContain('/bin/bash');
  });
});
