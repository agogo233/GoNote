import { defineConfig, devices } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Global teardown: clean up test data directory after all tests complete
 */
async function globalTeardown() {
  const testDataDir = path.join(__dirname, 'e2e', 'fixtures', 'test-data');
  try {
    if (fs.existsSync(testDataDir)) {
      fs.rmSync(testDataDir, { recursive: true, force: true });
      console.log(`Cleaned up test data directory: ${testDataDir}`);
    }
  } catch (error) {
    console.error('Failed to cleanup test data directory:', error);
  }
}

/**
 * See https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './e2e',
  /* Global teardown to clean up test data */
  globalTeardown: './e2e/fixtures/global-teardown.ts',
  /* Global timeout for entire test suite (30 min) */
  globalTimeout: 30 * 60 * 1000,
  /* Default timeout for each test action - reduced from 45s to 30s */
  timeout: 30 * 1000,
  /* Run tests in files in parallel */
  fullyParallel: true, // Enabled for faster CI execution
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only - reduced from 2 to 1 to save time */
  retries: process.env.CI ? 1 : 0,
  /* Opt out of parallel tests on CI - use multiple workers instead */
  workers: process.env.CI ? 4 : undefined,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: process.env.CI ? [['github'], ['list']] : 'list',
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions */
  use: {
    /* Base URL to use in actions like `await page.goto('/') */
    baseURL: 'http://localhost:9000',

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',

    /* Capture screenshot only on failure to reduce artifact size */
    screenshot: 'only-on-failure',

    /* Capture video only on failure to reduce artifact size */
    video: 'retain-on-failure',
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: {
          args: ['--no-sandbox', '--disable-setuid-sandbox']
        }
      },
      // Test execution order: auth -> CRUD -> folders -> search -> share -> advanced features
      dependencies: [],
    },
  ],

  /* Run tests in specific order for dependency management */
  // Order controlled by test file naming and fullyParallel: false
  // For CI, consider: npx playwright test --shard=1/3 auth/ notes/ search/

  /* Folder for test artifacts such as screenshots, videos, traces, etc. */
  outputDir: 'test-results/',
});
