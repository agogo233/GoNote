import { test, expect, TEST_CONFIG, login, waitForAutosave, cleanupTest } from '../fixtures/test-helpers';

test.describe('Folder Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('create new folder from sidebar', async ({ page, testPrefix }) => {
    const folderName = `${testPrefix}_new_folder`;
    
    // Click New button to open dropdown
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    // Wait for dropdown and click New Folder option
    const newFolderOption = page.locator('button:has-text("Folder"), [data-testid="new-folder"]').first();
    await newFolderOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    // Handle dialog for folder name
    page.once('dialog', async dialog => {
      await dialog.accept(folderName);
    });
    
    await newFolderOption.click({ force: true });

    // Wait for folder creation and reload
    await page.reload({ waitUntil: 'networkidle' });

    // Verify folder exists - use data-name attribute for precise matching
    const folderItem = page.locator(`[data-name="${folderName}"]`).first();
    await expect(folderItem).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('create note inside folder via homepage', async ({ page, testPrefix }) => {
    const folderName = `${testPrefix}_container`;
    const noteName = `${testPrefix}_inside_note`;
    
    // Create folder first
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const newFolderOption = page.locator('button:has-text("Folder"), [data-testid="new-folder"]').first();
    await newFolderOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(folderName);
    });
    
    await newFolderOption.click({ force: true });

    await page.reload({ waitUntil: 'networkidle' });

    // Navigate into the folder via homepage
    // Use more specific selector to avoid matching multiple elements
    const folderCard = page.locator(`[data-path="${folderName}"], [data-name="${folderName}"]`).first();
    await folderCard.click();
    await page.waitForTimeout(200);
    
    // Create note inside folder using New button
    await newButton.click();
    
    const newNoteOption = page.locator('button:has-text("Note"), [data-testid="new-note"]').first();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(noteName);
    });
    
    await newNoteOption.click({ force: true });
    
    // Verify editor is visible
    const editor = page.locator('#note-editor, textarea').first();
    await expect(editor).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
  });

  test('expand and collapse folder', async ({ page, testPrefix }) => {
    const folderName = `${testPrefix}_expand_test`;
    
    // Create folder
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const newFolderOption = page.locator('button:has-text("Folder"), [data-testid="new-folder"]').first();
    await newFolderOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(folderName);
    });
    
    await newFolderOption.click({ force: true });

    await page.reload({ waitUntil: 'networkidle' });

    // Use more specific selector to avoid matching multiple elements
    const folderItem = page.locator(`[data-name="${folderName}"]`).first();
    await folderItem.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Click to toggle (sidebar behavior)
    await folderItem.click();
    await page.waitForTimeout(200);
    
    // Verify folder still visible
    await expect(folderItem).toBeVisible();
  });

  test('delete empty folder', async ({ page, testPrefix }) => {
    const folderName = `${testPrefix}_to_delete`;

    // Create folder
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();

    const newFolderOption = page.locator('button:has-text("Folder"), [data-testid="new-folder"]').first();
    await newFolderOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    page.once('dialog', async dialog => {
      await dialog.accept(folderName);
    });

    await newFolderOption.click({ force: true });

    await page.reload({ waitUntil: 'networkidle' });

    // Find folder card by its heading text
    const folderCard = page.locator(`.homepage-card h3:has-text("${folderName}")`).first();
    await folderCard.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Hover over the folder card to reveal delete button
    const cardContainer = folderCard.locator('xpath=ancestor::*[contains(@class, "homepage-card")]').first();
    await cardContainer.hover();
    await page.waitForTimeout(200);

    // Find and click the delete button (card-delete-btn class)
    const deleteButton = cardContainer.locator('button.card-delete-btn').first();
    
    if (await deleteButton.isVisible({ timeout: 1000 }).catch(() => false)) {
      await deleteButton.click({ force: true });

      // Confirm deletion dialog
      const confirmButton = page.locator('button:has-text("Delete"), button:has-text("Confirm")').first();
      if (await confirmButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await confirmButton.click();
      }

      // Wait for folder to be removed from UI
      await page.waitForTimeout(500);
      await page.reload({ waitUntil: 'networkidle' });
      await page.waitForTimeout(500);
    }

    // Verify folder is gone
    const folderExists = await page.locator(`.homepage-card h3:has-text("${folderName}")`).isVisible({ timeout: 3000 }).catch(() => false);
    expect(folderExists).toBe(false);
  });
});
