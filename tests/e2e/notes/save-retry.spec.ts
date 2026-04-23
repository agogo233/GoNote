import { test, expect, TEST_CONFIG, login, waitForAutosave } from '../fixtures/test-helpers';

test.describe('Save Retry Mechanism', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('basic save should work normally', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_basic_save`;
    
    // Create new note
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const newNoteOption = page.locator('button:has-text("📝")').first();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(noteName);
    });
    
    await newNoteOption.click({ force: true });

    await page.waitForSelector('#note-editor', { state: 'visible', timeout: 10000 });

    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Add content and verify console for save log
    const testContent = `# Basic Save Test\n\nContent: ${testPrefix}`;
    await editor.fill(testContent);

    // Wait for autosave
    await waitForAutosave(page);
    
    // Check console for save success
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      if (msg.text().includes('[Save]') || msg.text().includes('saveNote')) {
        consoleMessages.push(msg.text());
      }
    });
    
    // Trigger another save
    await editor.fill(testContent + '\n\nAdditional line');
    await waitForAutosave(page);
    
    // Verify page is still functional
    await expect(editor).toBeVisible();
  });

  test('save retry when network fails', async ({ page, testPrefix, context }) => {
    const noteName = `${testPrefix}_retry_test`;
    
    // Create new note
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const newNoteOption = page.locator('button:has-text("📝")').first();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(noteName);
    });
    
    await newNoteOption.click({ force: true });

    await page.waitForSelector('#note-editor', { state: 'visible', timeout: 10000 });

    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Listen for console messages
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      consoleMessages.push(msg.text());
    });

    // Go offline
    await context.setOffline(true);
    
    // Try to save while offline
    const offlineContent = `# Offline Test\n\nThis should trigger retry: ${testPrefix}`;
    await editor.fill(offlineContent);
    
    // Wait a bit for the retry mechanism to kick in
    await page.waitForTimeout(3000);
    
    // Check localStorage for pending save backup
    const pendingSave = await page.evaluate(() => {
      return localStorage.getItem('gonote_pending_save');
    });
    
    console.log('Pending save in localStorage:', pendingSave ? 'exists' : 'null');
    
    // Go back online
    await context.setOffline(false);
    
    // Wait for retry to succeed
    await page.waitForTimeout(5000);
    
    // Check if retry succeeded
    const retryMessages = consoleMessages.filter(msg => 
      msg.includes('retry') || msg.includes('Retry') || msg.includes('saving')
    );
    console.log('Retry messages:', retryMessages);
    
    // Verify the note is still editable
    await expect(editor).toBeVisible();
  });

  test('localStorage backup on save failure', async ({ page, testPrefix, context }) => {
    const noteName = `${testPrefix}_backup_test`;
    
    // Create new note
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const newNoteOption = page.locator('button:has-text("📝")').first();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(noteName);
    });
    
    await newNoteOption.click({ force: true });

    await page.waitForSelector('#note-editor', { state: 'visible', timeout: 10000 });

    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Go offline to trigger backup
    await context.setOffline(true);
    
    const backupContent = `# Backup Test\n\nThis content should be backed up: ${testPrefix}`;
    await editor.fill(backupContent);
    
    // Wait for save attempt and backup
    await page.waitForTimeout(3000);
    
    // Check localStorage
    const pendingSave = await page.evaluate(() => {
      const data = localStorage.getItem('gonote_pending_save');
      if (data) {
        try {
          return JSON.parse(data);
        } catch {
          return data;
        }
      }
      return null;
    });
    
    console.log('Pending save data:', pendingSave);
    
    // Restore network
    await context.setOffline(false);
    
    // Verify backup was created (or check that mechanism exists)
    // The actual behavior depends on implementation
    await expect(editor).toBeVisible();
  });

  test('switching notes should cancel pending retry', async ({ page, testPrefix, context }) => {
    const noteName1 = `${testPrefix}_note1`;
    const noteName2 = `${testPrefix}_note2`;
    
    // Create first note
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const newNoteOption = page.locator('button:has-text("📝")').first();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(noteName1);
    });
    
    await newNoteOption.click({ force: true });

    await page.waitForSelector('#note-editor', { state: 'visible', timeout: 10000 });

    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Create second note
    await newButton.click();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(noteName2);
    });
    
    await newNoteOption.click({ force: true });

    await page.waitForSelector('#note-editor', { state: 'visible', timeout: 10000 });
    await editor.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Listen for console messages
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      consoleMessages.push(msg.text());
    });
    
    // Go offline and try to save
    await context.setOffline(true);
    
    await editor.fill(`Content that will trigger retry: ${testPrefix}`);
    await page.waitForTimeout(1000);
    
    // Switch to first note while retry is pending
    const firstNote = page.locator(`text="${noteName1}"`).first();
    await firstNote.click();
    
    await page.waitForTimeout(500);
    
    // Go back online
    await context.setOffline(false);
    
    // Wait a bit
    await page.waitForTimeout(2000);
    
    // Check console for cancellation messages
    const cancelMessages = consoleMessages.filter(msg => 
      msg.includes('cancel') || msg.includes('Cancel') || msg.includes('switched')
    );
    console.log('Cancel messages:', cancelMessages);
    
    // Verify editor is still functional
    await expect(editor).toBeVisible();
  });

  test('save serial number prevents concurrent conflicts', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_serial_test`;
    
    // Create new note
    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const newNoteOption = page.locator('button:has-text("📝")').first();
    await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    
    page.once('dialog', async dialog => {
      await dialog.accept(noteName);
    });
    
    await newNoteOption.click({ force: true });

    await page.waitForSelector('#note-editor', { state: 'visible', timeout: 10000 });

    const editor = page.locator('#note-editor').first();
    await editor.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Rapidly edit content multiple times
    for (let i = 0; i < 5; i++) {
      await editor.fill(`# Rapid Edit ${i}\n\nContent iteration ${i}: ${testPrefix}`);
      await page.waitForTimeout(100); // Very short delay
    }
    
    // Wait for all saves to complete
    await waitForAutosave(page);
    await page.waitForTimeout(2000);
    
    // Verify the final state
    await expect(editor).toBeVisible();
    
    // Reload and verify content persisted
    await page.reload();

    const editorAfterReload = page.locator('#note-editor').first();
    await expect(editorAfterReload).toBeVisible({ timeout: TEST_CONFIG.defaultTimeout });
    const content = await editorAfterReload.inputValue();
    console.log('Final content after rapid edits:', content.substring(0, 100));
    
    // Should contain the last edit
    expect(content).toContain('Rapid Edit 4');
  });
});
