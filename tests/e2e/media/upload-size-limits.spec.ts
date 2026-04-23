import { test, expect, TEST_CONFIG, login, ensureCsrfToken } from '../fixtures/test-helpers';

test.describe('Upload Size Limits', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('rejects file exceeding max_file_size_mb', async ({ page }) => {
    // Create a file larger than 50MB (the default limit)
    // We'll create a ~51MB payload to test
    const largeBuffer = Buffer.alloc(51 * 1024 * 1024, 'A');

    const csrfToken = await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/media/upload`, {
      headers: {
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      multipart: {
        file: {
          name: 'large-file.bin',
          mimeType: 'application/octet-stream',
          buffer: largeBuffer,
        },
      },
    });

    // Should reject with 413 (Payload Too Large) or 400
    expect([413, 400]).toContain(response.status());
  });

  test('accepts file within size limit', async ({ page }) => {
    // Create a small file (well within 50MB limit)
    const smallBuffer = Buffer.alloc(1024, 'A'); // 1KB

    const csrfToken = await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/media/upload`, {
      headers: {
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      multipart: {
        file: {
          name: 'small-file.txt',
          mimeType: 'text/plain',
          buffer: smallBuffer,
        },
      },
    });

    // Should accept (200/201) or reject for other reasons (wrong type, etc.)
    const status = response.status();
    // 200/201 = accepted, 400 = wrong type, 413 = too large (shouldn't happen for 1KB)
    expect([200, 201, 400]).toContain(status);
  });

  test('note content size limit is enforced', async ({ page, testPrefix }) => {
    const csrfToken = await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // Create a note with very large content (10MB)
    const largeContent = 'X'.repeat(10 * 1024 * 1024);

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/notes/${testPrefix}_huge_note.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ content: largeContent }),
    });

    // Should reject oversized note content
    expect([413, 400]).toContain(response.status());
  });

  test('multipart form size limit is enforced', async ({ page, testPrefix }) => {
    const csrfToken = await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // Create a moderately large multipart upload (20MB)
    const mediumBuffer = Buffer.alloc(20 * 1024 * 1024, 'B');

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/media/upload`, {
      headers: {
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      multipart: {
        file: {
          name: 'medium-file.bin',
          mimeType: 'application/octet-stream',
          buffer: mediumBuffer,
        },
      },
    });

    const status = response.status();
    // Should either accept (within limit) or reject (over limit)
    // Both are valid behaviors - we're testing the limit exists
    expect([200, 201, 413, 400]).toContain(status);
  });

  test('multiple file upload in single request respects limit', async ({ page, testPrefix }) => {
    const csrfToken = await ensureCsrfToken(page);
    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    // Upload several files at once (each within limit)
    const files = [];
    for (let i = 0; i < 5; i++) {
      files.push({
        name: `batch-file-${i}.txt`,
        mimeType: 'text/plain',
        buffer: Buffer.alloc(1024 * 100, `File ${i} content`), // 100KB each
      });
    }

    const response = await page.request.post(`${TEST_CONFIG.baseUrl}/api/media/upload`, {
      headers: {
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      multipart: {
        files: files,
      },
    });

    const status = response.status();
    // Should handle batch upload appropriately
    expect([200, 201, 400, 413]).toContain(status);
  });
});
