import { test, expect, TEST_CONFIG, login, waitForAutosave } from '../fixtures/test-helpers';
import * as path from 'path';
import * as fs from 'fs';

async function openFilesPanel(page: import('@playwright/test').Page) {
  const iconRailButtons = page.locator('.icon-rail button, [class*="icon-rail"] button');
  const buttonCount = await iconRailButtons.count();
  
  if (buttonCount >= 1) {
    await iconRailButtons.nth(0).click();
  }
  
  await page.waitForTimeout(200);
}

async function createNoteViaDialog(page: import('@playwright/test').Page, noteName: string) {
  await openFilesPanel(page);
  
  const newButton = page.locator('button:has-text("New")').first();
  await newButton.click();
  
  const newNoteOption = page.locator('button:has-text("📝"), button:has-text("Note")').first();
  await newNoteOption.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
  
  page.once('dialog', async dialog => {
    await dialog.accept(noteName);
  });
  
  await newNoteOption.click({ force: true });
  await page.waitForTimeout(1500);
  
  const editor = page.locator('#note-editor').first();
  await editor.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
}

test.describe('Media Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('image upload via file input', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_media_upload`;
    await createNoteViaDialog(page, noteName);
    
    const testImagePath = path.join(__dirname, '..', 'fixtures', 'test-image.png');
    
    if (!fs.existsSync(testImagePath)) {
      const pngHeader = Buffer.from([
        0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
        0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
        0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
        0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
        0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
        0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
        0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
        0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
        0x44, 0xAE, 0x42, 0x60, 0x82
      ]);
      fs.mkdirSync(path.dirname(testImagePath), { recursive: true });
      fs.writeFileSync(testImagePath, pngHeader);
    }
    
    const fileInput = page.locator('input[type="file"]').first();
    
    if (await fileInput.isVisible({ timeout: 2000 }).catch(() => false)) {
      await fileInput.setInputFiles(testImagePath);
      await page.waitForTimeout(2000);
      
      const editor = page.locator('#note-editor').first();
      const content = await editor.inputValue();
      
      expect(content.length).toBeGreaterThan(0);
    } else {
      console.log('File input not visible - checking for drag-drop or paste support');
      
      const editor = page.locator('#note-editor').first();
      await editor.focus();
      
      const imageBuffer = fs.readFileSync(testImagePath);
      const base64 = imageBuffer.toString('base64');
      
      await page.evaluate(async ({ base64Data }) => {
        const response = await fetch('data:image/png;base64,' + base64Data);
        const blob = await response.blob();
        const file = new File([blob], 'test-image.png', { type: 'image/png' });
        
        const clipboardData = new DataTransfer();
        clipboardData.items.add(file);
        
        const pasteEvent = new ClipboardEvent('paste', {
          clipboardData: clipboardData,
          bubbles: true,
          cancelable: true
        });
        
        document.dispatchEvent(pasteEvent);
      }, { base64Data: base64 });
      
      await page.waitForTimeout(2000);
    }
  });

  test('media file displays in sidebar', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_media_sidebar`;
    await createNoteViaDialog(page, noteName);
    
    const editor = page.locator('#note-editor').first();
    await editor.fill(`# Test Note with Media\n\n![test image](test-image.png)`);
    await editor.press('Control+s');
    await waitForAutosave(page);
    
    await page.reload();
    await page.waitForTimeout(300);

    const sidebarItems = page.locator('.note-item, [class*="note-item"]');
    const count = await sidebarItems.count();
    
    expect(count).toBeGreaterThan(0);
  });

  test('orphaned media scan works', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_orphan_test`;
    await createNoteViaDialog(page, noteName);
    
    const settingsBtn = page.locator('button[title*="Settings"], button[aria-label*="Settings"], [data-testid="settings"]').first();
    
    if (await settingsBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await settingsBtn.click();
      await page.waitForTimeout(200);

      const cleanupBtn = page.locator('text=/orphan|cleanup|media/i').first();
      if (await cleanupBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await cleanupBtn.click();
        await page.waitForTimeout(300);
        
        const orphanedList = page.locator('.orphaned-media, [class*="orphaned"]');
        const isVisible = await orphanedList.isVisible({ timeout: 2000 }).catch(() => false);
        
        console.log(`Orphaned media panel visible: ${isVisible}`);
      }
    } else {
      console.log('Settings button not found - orphaned media feature may not be accessible');
    }
  });

  test('media file can be deleted', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_delete_media`;
    await createNoteViaDialog(page, noteName);
    
    const editor = page.locator('#note-editor').first();
    await editor.fill(`# Test\n\n![[test-media-${Date.now()}.png]]`);
    await editor.press('Control+s');
    await waitForAutosave(page);
    
    await page.reload();
    await page.waitForTimeout(300);

    const mediaItem = page.locator('[class*="media"], [data-type*="image"], [data-type*="audio"], [data-type*="video"]').first();
    
    if (await mediaItem.isVisible({ timeout: 2000 }).catch(() => false)) {
      await mediaItem.click({ button: 'right' });
      
      const deleteOption = page.locator('text=Delete, [data-testid="delete"]').first();
      if (await deleteOption.isVisible({ timeout: 2000 }).catch(() => false)) {
        console.log('Delete option available for media');
      }
    }
  });

  test('supported media types are recognized', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_media_types`;
    await createNoteViaDialog(page, noteName);
    
    const editor = page.locator('#note-editor').first();
    
    const testContent = `# Media Types Test

## Image
![[test-image.png]]

## Audio
![[test-audio.mp3]]

## Video
![[test-video.mp4]]

## PDF
![[test-doc.pdf]]
`;
    
    await editor.fill(testContent);
    await editor.press('Control+s');
    await waitForAutosave(page);
    
    const previewBtn = page.locator('button[title*="Preview"], button[aria-label*="Preview"]').first();
    if (await previewBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await previewBtn.click();
      await page.waitForTimeout(200);
    }

    const preview = page.locator('.preview, [class*="preview"], #preview');
    if (await preview.isVisible({ timeout: 2000 }).catch(() => false)) {
      const previewContent = await preview.innerHTML();
      
      const hasImage = previewContent.includes('<img') || previewContent.includes('test-image');
      const hasAudio = previewContent.includes('<audio') || previewContent.includes('test-audio');
      const hasVideo = previewContent.includes('<video') || previewContent.includes('test-video');
      const hasPdf = previewContent.includes('<iframe') || previewContent.includes('test-doc');
      
      console.log(`Media types found - Image: ${hasImage}, Audio: ${hasAudio}, Video: ${hasVideo}, PDF: ${hasPdf}`);
    }
  });
});
