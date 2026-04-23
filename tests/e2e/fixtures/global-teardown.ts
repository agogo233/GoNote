/**
 * Global teardown: runs once after all tests complete
 * Cleans up test artifacts and resets state
 */
import * as fs from 'fs';
import * as path from 'path';

export default async function globalTeardown() {
  const dataDir = path.join(process.cwd(), 'go', 'data');
  const testPattern = 'test_';

  try {
    if (!fs.existsSync(dataDir)) {
      return;
    }

    let cleanedCount = 0;

    // Clean up test notes (files starting with test_)
    function cleanupDirRecursive(dir: string) {
      if (!fs.existsSync(dir)) return;

      const entries = fs.readdirSync(dir, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);

        if (entry.isDirectory()) {
          // Check if folder name starts with test pattern
          if (entry.name.startsWith(testPattern)) {
            fs.rmSync(fullPath, { recursive: true, force: true });
            cleanedCount++;
          } else {
            // Recurse into non-test folders
            cleanupDirRecursive(fullPath);
          }
        } else if (entry.name.startsWith(testPattern)) {
          fs.unlinkSync(fullPath);
          cleanedCount++;
        }
      }
    }

    cleanupDirRecursive(dataDir);
    console.log(`\n Cleanup complete: removed ${cleanedCount} test artifact(s)`);
  } catch (error) {
    console.error(' Warning: Failed to cleanup test artifacts:', error);
  }
}
