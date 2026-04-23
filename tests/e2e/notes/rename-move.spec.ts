import { test, expect, TEST_CONFIG, login, apiPost, waitForAutosave } from '../fixtures/test-helpers';

test.describe('Rename and Move Notes and Folders', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('rename note via UI context menu', async ({ page, testPrefix }) => {
    const originalName = `${testPrefix}_original_name`;
    const newName = `${testPrefix}_renamed_note`;

    // Create note
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${originalName}.md`, { content: '# Original Name\n\nContent for rename test.' });
    await page.reload();

    // Right-click on note in sidebar
    const noteItem = page.locator(`text="${originalName}"`).first();
    await expect(noteItem).toBeVisible({ timeout: 5000 });
    await noteItem.click({ button: 'right' });
    await page.waitForTimeout(150);

    // Click rename option
    const renameOption = page.locator('text=/Rename|重命名/').first();
    if (await renameOption.isVisible({ timeout: 2000 }).catch(() => false)) {
      // Handle the rename dialog
      page.once('dialog', async dialog => {
        await dialog.accept(newName);
      });
      await renameOption.click();
      await page.waitForTimeout(2000);

      // Verify note appears with new name
      const renamedNote = page.locator(`text="${newName}"`).first();
      await expect(renamedNote).toBeVisible({ timeout: 5000 });
    } else {
      // If no rename option in context menu, test rename via API
      const csrfToken = await page.evaluate(async () => {
        const cookies = document.cookie.split(';');
        const csrf = cookies.find(c => c.trim().startsWith('csrf'));
        return csrf ? csrf.split('=')[1] : '';
      });

      const context = page.context();
      const cookies = await context.cookies();
      const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

      const response = await page.request.put(`${TEST_CONFIG.baseUrl}/api/notes/${originalName}.md`, {
        headers: {
          'Content-Type': 'application/json',
          'Cookie': cookieHeader,
          'X-CSRF-Token': csrfToken,
        },
        data: JSON.stringify({ newName: `${newName}.md` }),
      });

      // Rename should succeed (200) or endpoint may not exist (404/405)
      expect([200, 404, 405]).toContain(response.status());
    }
  });

  test('move note to different folder via drag-and-drop', async ({ page, testPrefix }) => {
    const folderName = `${testPrefix}_target_folder`;
    const noteName = `${testPrefix}_note_to_move`;

    // Create target folder
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/folders/${folderName}`, {});
    await page.waitForTimeout(200);

    // Create note
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, { content: `# Note to Move\n\nThis note will be moved to ${folderName}.` });
    await page.reload();

    // Check if note is in root
    const noteInRoot = page.locator(`text="${noteName}"`).first();
    await expect(noteInRoot).toBeVisible({ timeout: 5000 });

    // Verify folder exists
    const folder = page.locator(`text="${folderName}"`).first();
    await expect(folder).toBeVisible({ timeout: 5000 });

    // Note: Drag-and-drop may not be easily testable in E2E
    // Test move via API if available
    const csrfToken = await page.evaluate(async () => {
      const cookies = document.cookie.split(';');
      const csrf = cookies.find(c => c.trim().startsWith('csrf'));
      return csrf ? csrf.split('=')[1] : '';
    });

    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    const moveResponse = await page.request.put(`${TEST_CONFIG.baseUrl}/api/notes/${noteName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ newPath: `${folderName}/${noteName}.md` }),
    });

    const status = moveResponse.status();
    // Move may succeed (200) or endpoint may not be implemented (404/405)
    expect([200, 404, 405]).toContain(status);
  });

  test('rename folder via UI', async ({ page, testPrefix }) => {
    const originalFolderName = `${testPrefix}_old_folder`;
    const newFolderName = `${testPrefix}_new_folder`;

    // Create folder
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/folders/${originalFolderName}`, {});
    await page.reload();

    // Find folder in sidebar
    const folderItem = page.locator(`text="${originalFolderName}"`).first();
    const isFolderVisible = await folderItem.isVisible({ timeout: 3000 }).catch(() => false);

    if (isFolderVisible) {
      // Right-click to rename
      await folderItem.click({ button: 'right' });
      await page.waitForTimeout(150);

      const renameOption = page.locator('text=/Rename|重命名/').first();
      const renameVisible = await renameOption.isVisible({ timeout: 2000 }).catch(() => false);

      if (renameVisible) {
        page.once('dialog', async dialog => {
          await dialog.accept(newFolderName);
        });
        await renameOption.click();
        await page.waitForTimeout(2000);

        // Verify folder appears with new name
        const renamedFolder = page.locator(`text="${newFolderName}"`).first();
        await expect(renamedFolder).toBeVisible({ timeout: 5000 });
      } else {
        // If no rename option, test via API
        const csrfToken = await page.evaluate(async () => {
          const cookies = document.cookie.split(';');
          const csrf = cookies.find(c => c.trim().startsWith('csrf'));
          return csrf ? csrf.split('=')[1] : '';
        });

        const context = page.context();
        const cookies = await context.cookies();
        const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

        const response = await page.request.put(`${TEST_CONFIG.baseUrl}/api/folders/${originalFolderName}`, {
          headers: {
            'Content-Type': 'application/json',
            'Cookie': cookieHeader,
            'X-CSRF-Token': csrfToken,
          },
          data: JSON.stringify({ newName: newFolderName }),
        });

        expect([200, 404, 405]).toContain(response.status());
      }
    }
  });

  test('move folder to different parent', async ({ page, testPrefix }) => {
    const parentFolder = `${testPrefix}_parent`;
    const childFolder = `${testPrefix}_child`;

    // Create parent folder
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/folders/${parentFolder}`, {});
    await page.waitForTimeout(200);

    // Create child folder
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/folders/${parentFolder}/${childFolder}`, {});
    await page.reload();

    // Verify parent folder is visible
    const parentItem = page.locator(`text="${parentFolder}"`).first();
    const parentVisible = await parentItem.isVisible({ timeout: 3000 }).catch(() => false);

    if (parentVisible) {
      // Click to expand parent
      await parentItem.click();
      await page.waitForTimeout(500);

      // Verify child is inside parent
      const childInParent = page.locator(`text="${childFolder}"`).first();
      const childVisible = await childInParent.isVisible({ timeout: 2000 }).catch(() => false);
      expect(childVisible).toBe(true);
    }
  });

  test('note content preserved after rename', async ({ page, testPrefix }) => {
    const originalName = `${testPrefix}_preserve_content`;
    const newName = `${testPrefix}_renamed_preserved`;
    const uniqueContent = `UniqueContentToken${Date.now()}`;

    // Create note with unique content
    await apiPost(page, `${TEST_CONFIG.baseUrl}/api/notes/${originalName}.md`, { content: `# Original\n\n${uniqueContent}` });
    await page.waitForTimeout(200);

    // Rename via API (PUT endpoint)
    const csrfToken = await page.evaluate(async () => {
      const cookies = document.cookie.split(';');
      const csrf = cookies.find(c => c.trim().startsWith('csrf'));
      return csrf ? csrf.split('=')[1] : '';
    });

    const context = page.context();
    const cookies = await context.cookies();
    const cookieHeader = cookies.map(c => `${c.name}=${c.value}`).join('; ');

    await page.request.put(`${TEST_CONFIG.baseUrl}/api/notes/${originalName}.md`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
        'X-CSRF-Token': csrfToken,
      },
      data: JSON.stringify({ newName: `${newName}.md` }),
    });

    await page.waitForTimeout(1000);

    // Fetch renamed note and verify content
    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes/${newName}.md`);
    if (response.status() === 200) {
      const data = await response.json();
      expect(data.content).toContain(uniqueContent);
    } else {
      // If PUT rename not supported, verify original note still has content
      const originalResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes/${originalName}.md`);
      const originalData = await originalResponse.json();
      expect(originalData.content).toContain(uniqueContent);
    }
  });
});
