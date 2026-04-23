import { test as base, expect, Page } from '@playwright/test';

const TEST_CONFIG = {
  baseUrl: 'http://localhost:9000',
  testPassword: 'test-admin-password',
  autosaveDelay: 800,
  searchDebounceDelay: 400,
  cacheTtl: 3000,
  defaultTimeout: 10000,
  shortTimeout: 3000,
};

function generateUniqueTestPrefix(): string {
  const timestamp = Date.now();
  const random = Math.random().toString(36).substring(2, 8);
  return `test_${timestamp}_${random}`;
}

// --- CSRF Helpers ---

async function getCsrfToken(page: Page): Promise<string | null> {
  const context = page.context();
  const cookies = await context.cookies();
  const csrfCookie = cookies.find(c => c.name === 'csrf_');
  return csrfCookie?.value || null;
}

async function ensureCsrfToken(page: Page): Promise<string> {
  let csrfToken = await getCsrfToken(page);
  if (!csrfToken) {
    await page.goto('/');
    await page.waitForTimeout(200);
    csrfToken = await getCsrfToken(page);
  }
  return csrfToken || '';
}

async function apiPost(page: Page, url: string, data?: any): Promise<any> {
  const context = page.context();
  const csrfToken = await ensureCsrfToken(page);
  const cookies = await context.cookies();
  const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Cookie': cookieHeader,
  };
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }
  return context.request.post(url, {
    headers,
    data: data ? JSON.stringify(data) : undefined,
  });
}

async function apiDelete(page: Page, url: string): Promise<any> {
  const context = page.context();
  const csrfToken = await ensureCsrfToken(page);
  const cookies = await context.cookies();
  const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');
  const headers: Record<string, string> = {
    'Cookie': cookieHeader,
  };
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }
  return context.request.delete(url, { headers });
}

// --- Event-Driven Wait Functions ---

/**
 * Wait for note autosave to complete by intercepting the API response.
 * Falls back to a short timeout if no API call is detected.
 */
async function waitForAutosave(page: Page): Promise<void> {
  try {
    await page.waitForResponse(
      resp => resp.url().includes('/api/notes/') && (resp.status() === 200 || resp.status() === 201),
      { timeout: 3000 }
    );
  } catch {
    // Fallback: no API call detected, wait briefly
    await page.waitForTimeout(300);
  }
}

/**
 * Wait for search debounce and results to load.
 */
async function waitForSearchDebounce(page: Page): Promise<void> {
  await page.waitForTimeout(TEST_CONFIG.searchDebounceDelay);
}

/**
 * Wait for search index to update after creating/editing a note.
 * Uses a page reload with dom ready as the index trigger.
 */
async function waitForSearchIndex(page: Page, timeout?: number): Promise<void> {
  try {
    // Wait a bit for the backend to process the note
    await page.waitForTimeout(500);
    // Reload page to trigger search index rebuild
    await page.reload({ waitUntil: 'dom', timeout: timeout || 5000 });
    // Wait additional time for index to be ready
    await page.waitForTimeout(500);
  } catch {
    await page.waitForTimeout(500);
  }
}

// --- Navigation Helpers ---

async function login(page: Page, password: string = TEST_CONFIG.testPassword): Promise<void> {
  await page.goto('/login');
  const passwordInput = page.locator('input[type="password"]');
  const isAuthEnabled = await passwordInput.isVisible({ timeout: 3000 }).catch(() => false);

  if (!isAuthEnabled) {
    await page.goto('/');
    await page.waitForSelector('#app, [x-data]', { timeout: TEST_CONFIG.defaultTimeout });
    return;
  }

  await passwordInput.fill(password);
  await page.click('button[type="submit"]');

  await Promise.race([
    page.waitForURL('**/', { timeout: 30000 }),
    page.waitForSelector('#app', { timeout: 30000 }),
    page.waitForSelector('[x-data]', { timeout: 30000 })
  ]);

  await page.waitForTimeout(200);
}

async function logout(page: Page): Promise<void> {
  await page.goto('/logout');
  await Promise.race([
    page.waitForURL('**/login', { timeout: 5000 }).catch(() => {}),
    page.waitForURL('**/', { timeout: 5000 }).catch(() => {})
  ]);
}

// --- Test Fixtures ---

type TestFixtures = {
  testPrefix: string;
};

export const test = base.extend<TestFixtures>({
  testPrefix: async ({}, use) => {
    const prefix = generateUniqueTestPrefix();
    await use(prefix);
  },
});

// --- Per-test cleanup hook ---

async function cleanupTestData(baseUrl: string, testPrefix: string): Promise<void> {
  try {
    const timeout = setTimeout(() => {
      console.warn(' Cleanup timeout reached');
    }, 10000);

    // Cleanup via API: list all notes and delete those matching test prefix
    // This is best-effort and non-blocking
    const response = await fetch(`${baseUrl}/api/notes`);
    if (response.ok) {
      const notes: Array<{ path: string }> = await response.json();
      for (const note of notes) {
        if (note.path && note.path.includes(testPrefix)) {
          try {
            await fetch(`${baseUrl}/api/notes/${encodeURIComponent(note.path)}`, {
              method: 'DELETE',
            });
          } catch {
            // Ignore individual delete failures
          }
        }
      }
    }

    clearTimeout(timeout);
  } catch {
    // Ignore cleanup failures to not break test reporting
  }
}

/**
 * Cleanup test data matching the given prefix.
 * Usage in test files: afterEach(async ({ testPrefix }) => { await cleanupTest(testPrefix); });
 */
async function cleanupTest(testPrefix: string): Promise<void> {
  await cleanupTestData(TEST_CONFIG.baseUrl, testPrefix);
}

export {
  expect,
  TEST_CONFIG,
  waitForAutosave,
  waitForSearchDebounce,
  waitForSearchIndex,
  login,
  logout,
  getCsrfToken,
  ensureCsrfToken,
  apiPost,
  apiDelete,
  cleanupTest,
};
