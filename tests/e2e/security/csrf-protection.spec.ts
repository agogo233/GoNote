import { test, expect, TEST_CONFIG, login, getCsrfToken, ensureCsrfToken } from '../fixtures/test-helpers';

test.describe('CSRF Protection', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('CSRF token is set as cookie after login', async ({ page }) => {
    const cookies = await page.context().cookies();
    const csrfCookie = cookies.find(c => c.name.startsWith('csrf'));
    expect(csrfCookie).toBeDefined();
    expect(csrfCookie?.value).toBeTruthy();
  });

  test('POST request without CSRF token is rejected', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_no_csrf`;

    // Get cookies but deliberately omit CSRF header
    const cookies = await page.context().cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        // Deliberately NOT sending X-CSRF-Token
      },
      data: JSON.stringify({ content: 'This should fail without CSRF token' }),
    });

    // Should be rejected (403) or accepted (200/201) depending on CSRF config
    // In strict CSRF mode, this should be 403
    const status = response.status();
    // If 200/201, CSRF might be disabled or using double-submit cookie pattern
    // where cookie + header match is required. Without the header it should fail.
    expect([403, 400]).toContain(status);
  });

  test('POST request with valid CSRF token succeeds', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_with_csrf`;

    const csrfToken = await ensureCsrfToken(page);
    expect(csrfToken).toBeTruthy();

    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ content: 'This should succeed with valid CSRF token' }),
    });

    expect([200, 201]).toContain(response.status());
  });

  test('DELETE request without CSRF token is rejected', async ({ page, testPrefix }) => {
    // First create a note to delete
    const noteName = `${testPrefix}_delete_no_csrf`;
    const csrfToken = await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // Create note with CSRF
    await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ content: 'Note to delete without CSRF' }),
    });

    // Try to delete WITHOUT CSRF token
    const deleteResponse = await page.request.delete(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Cookie': cookieHeader,
        // No X-CSRF-Token
      },
    });

    expect([403, 400]).toContain(deleteResponse.status());
  });

  test('DELETE request with valid CSRF token succeeds', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_delete_with_csrf`;
    const csrfToken = await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // Create note
    await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ content: 'Note to delete with CSRF' }),
    });

    // Delete with CSRF
    const deleteResponse = await page.request.delete(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
    });

    expect([200, 204]).toContain(deleteResponse.status());
  });

  test('CSRF token mismatch is rejected', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_mismatch_csrf`;

    await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // Send a deliberately invalid CSRF token
    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': 'invalid-token-12345',
      },
      data: JSON.stringify({ content: 'Should fail with invalid CSRF' }),
    });

    expect([403, 400]).toContain(response.status());
  });

  test('GET requests do not require CSRF token', async ({ page }) => {
    const cookies = await page.context().cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // GET requests should work with just cookies
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes`, {
      headers: {
        'Cookie': cookieHeader,
      },
    });

    expect([200, 401]).toContain(response.status());
  });

  test('CSRF cookie is httponly or samesite configured', async ({ page }) => {
    const cookies = await page.context().cookies();
    const csrfCookie = cookies.find(c => c.name.startsWith('csrf'));

    expect(csrfCookie).toBeDefined();
    // Verify SameSite is set (Lax or Strict)
    expect(['Lax', 'Strict', 'lax', 'strict']).toContain(csrfCookie?.sameSite);
  });
});
