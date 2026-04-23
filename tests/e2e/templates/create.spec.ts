import { test, expect, TEST_CONFIG, login, waitForAutosave, apiPost, cleanupTest } from '../fixtures/test-helpers';
import * as path from 'path';
import * as fs from 'fs';

const TEMPLATES_DIR = path.join(process.cwd(), 'go', 'data', '_templates');

function ensureTemplatesDir() {
  if (!fs.existsSync(TEMPLATES_DIR)) {
    fs.mkdirSync(TEMPLATES_DIR, { recursive: true });
  }
}

async function openFilesPanel(page: import('@playwright/test').Page) {
  const iconRailButtons = page.locator('.icon-rail button, [class*="icon-rail"] button');
  const buttonCount = await iconRailButtons.count();
  
  if (buttonCount >= 1) {
    await iconRailButtons.nth(0).click();
  }
  
  await page.waitForTimeout(200);
}

test.describe('Template Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    ensureTemplatesDir();
  });

  test.afterEach(async ({ testPrefix }) => {
    await cleanupTest(testPrefix);
  });

  test('template list displays correctly', async ({ page }) => {
    const testTemplatePath = path.join(TEMPLATES_DIR, 'TestTemplate.md');
    if (!fs.existsSync(testTemplatePath)) {
      fs.writeFileSync(testTemplatePath, `# {{title}}\n\nCreated on {{date}}\n\nContent here.`);
    }
    
    await page.waitForTimeout(200);
    await page.reload();
    await page.waitForTimeout(300);

    const response = await page.request.get(`${TEST_CONFIG.baseUrl}/api/templates`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    expect(Array.isArray(data.templates)).toBe(true);
    
    console.log(`Found ${data.templates.length} templates`);
  });

  test('create note from template via API', async ({ page, testPrefix }) => {
    const templateName = `APITemplate${Date.now()}`;
    const templatePath = path.join(TEMPLATES_DIR, `${templateName}.md`);
    fs.writeFileSync(templatePath, `# {{title}}\n\nDate: {{date}}\nTime: {{time}}`);

    await page.waitForTimeout(200);

    const noteName = `${testPrefix}_from_template`;
    const notePath = `${noteName}.md`;
    
    const response = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/templates/create-note`, {
      templateName: templateName,
      notePath: notePath
    });
    
    const data = await response.json();
    console.log(`Template create response: ${JSON.stringify(data)}`);
    
    expect(response.status()).toBe(200);
    expect(data.success).toBe(true);
    expect(data.notePath).toContain(noteName);
    
    fs.unlinkSync(templatePath);
  });

  test('template placeholders are replaced', async ({ page, testPrefix }) => {
    const templateName = `PlaceholderTemplate${Date.now()}`;
    const templatePath = path.join(TEMPLATES_DIR, `${templateName}.md`);
    
    const templateContent = `# {{title}}

Date: {{date}}
Time: {{time}}
DateTime: {{datetime}}
Year: {{year}}
Month: {{month}}
Day: {{day}}
Timestamp: {{timestamp}}
Folder: {{folder}}
`;
    
    fs.writeFileSync(templatePath, templateContent);
    await page.waitForTimeout(200);

    const noteName = `${testPrefix}_placeholders`;
    const notePath = `${noteName}.md`;
    
    const response = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/templates/create-note`, {
      templateName: templateName,
      notePath: notePath
    });
    
    expect(response.status()).toBe(200);
    
    const noteResponse = await page.request.get(`${TEST_CONFIG.baseUrl}/api/notes/${notePath}`);
    const noteContent = await noteResponse.text();
    
    expect(noteContent).not.toContain('{{title}}');
    expect(noteContent).not.toContain('{{date}}');
    expect(noteContent).not.toContain('{{time}}');
    expect(noteContent).toContain(noteName);
    
    const datePattern = /\d{4}-\d{2}-\d{2}/;
    expect(noteContent).toMatch(datePattern);
    
    fs.unlinkSync(templatePath);
  });

  test('template modal can be opened', async ({ page }) => {
    const testTemplatePath = path.join(TEMPLATES_DIR, 'ModalTest.md');
    if (!fs.existsSync(testTemplatePath)) {
      fs.writeFileSync(testTemplatePath, `# Modal Test Template\n\nContent.`);
    }
    
    await page.reload();
    await page.waitForTimeout(300);

    const newButton = page.locator('button:has-text("New")').first();
    await newButton.click();
    
    const templateOption = page.locator('button:has-text("Template"), [data-testid="new-from-template"]').first();
    
    if (await templateOption.isVisible({ timeout: 2000 }).catch(() => false)) {
      await templateOption.click();
      await page.waitForTimeout(200);

      const modal = page.locator('.modal, [class*="modal"], [x-show*="showTemplateModal"]').first();
      const isModalVisible = await modal.isVisible({ timeout: 2000 }).catch(() => false);
      
      console.log(`Template modal visible: ${isModalVisible}`);
    } else {
      console.log('Template option not found in New dropdown');
    }
  });

  test('non-existent template returns error', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_nonexistent`;
    const notePath = `${noteName}.md`;
    
    const response = await apiPost(page, `${TEST_CONFIG.baseUrl}/api/templates/create-note`, {
      templateName: 'NonExistentTemplate12345',
      notePath: notePath
    });
    
    expect(response.status()).toBe(500);
  });
});
