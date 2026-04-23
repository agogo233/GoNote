import { test, expect, TEST_CONFIG, login, waitForAutosave } from '../fixtures/test-helpers';

test.describe('Checkbox Interaction', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('render checkboxes in preview mode', async ({ page }) => {
    const testPrefix = `cb_render_${Date.now()}`;

    // Create a note with task items via API
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}.md`,
          content: `# Task Test\n\n- [ ] Pending task\n- [x] Completed task`
        })
      });
    }, { testPrefix });

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${testPrefix}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Switch to Preview mode using the view mode selector buttons
    // Buttons use i18n text: "Preview" (en) or "预览" (zh)
    const previewButton = page.locator('button').filter({ hasText: /^(Preview|预览)$/ }).first();
    await previewButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await previewButton.click();
    await page.waitForTimeout(500);

    // Wait for preview container to be visible
    const previewContainer = page.locator('.markdown-preview');
    await previewContainer.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Verify 2 checkboxes rendered
    const checkboxes = page.locator('input[data-interactive-checkbox]');
    await expect(checkboxes).toHaveCount(2, { timeout: TEST_CONFIG.defaultTimeout });

    // First checkbox should be unchecked
    const firstCheckbox = checkboxes.nth(0);
    await expect(firstCheckbox).not.toBeChecked();

    // Second checkbox should be checked
    const secondCheckbox = checkboxes.nth(1);
    await expect(secondCheckbox).toBeChecked();
  });

  test('click checkbox toggles state and updates editor content', async ({ page }) => {
    const testPrefix = `cb_toggle_${Date.now()}`;

    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}.md`,
          content: `- [ ] Click me`
        })
      });
    }, { testPrefix });

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${testPrefix}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Switch to Preview mode
    const previewButton = page.locator('button').filter({ hasText: /^(Preview|预览)$/ }).first();
    await previewButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await previewButton.click();
    await page.waitForTimeout(500);

    // Wait for preview container
    await page.locator('.markdown-preview').waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Click the checkbox
    const checkbox = page.locator('input[data-interactive-checkbox]').first();
    await expect(checkbox).not.toBeChecked();
    await checkbox.click();

    // Verify checkbox is now checked
    await expect(checkbox).toBeChecked({ timeout: TEST_CONFIG.defaultTimeout });

    // Switch to Edit mode to verify content changed
    const editButton = page.locator('button').filter({ hasText: /^(Edit|编辑)$/ }).first();
    await editButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await editButton.click();
    await page.waitForTimeout(500);

    // Editor should contain "- [x] Click me"
    const editor = page.locator('textarea#note-editor');
    await expect(editor).toContainText('- [x] Click me', { timeout: TEST_CONFIG.defaultTimeout });
  });

  test('auto-save after checkbox toggle persists on reload', async ({ page }) => {
    const testPrefix = `cb_autosave_${Date.now()}`;

    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}.md`,
          content: `- [ ] Persist me`
        })
      });
    }, { testPrefix });

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${testPrefix}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Switch to Preview mode
    const previewButton = page.locator('button').filter({ hasText: /^(Preview|预览)$/ }).first();
    await previewButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await previewButton.click();
    await page.waitForTimeout(500);

    // Wait for preview container
    await page.locator('.markdown-preview').waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Click checkbox and wait for autosave
    const checkbox = page.locator('input[data-interactive-checkbox]').first();
    await checkbox.click();
    await expect(checkbox).toBeChecked({ timeout: TEST_CONFIG.defaultTimeout });

    // Wait for autosave API call
    await waitForAutosave(page);

    // Reload page
    await page.reload();

    // Reopen the note
    const noteItemAfter = page.locator(`text="${testPrefix}"`).first();
    await noteItemAfter.click();
    await page.waitForTimeout(200);

    // Switch to Edit mode and verify content
    const editButton = page.locator('button').filter({ hasText: /^(Edit|编辑)$/ }).first();
    await editButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await editButton.click();
    await page.waitForTimeout(500);

    const editor = page.locator('textarea#note-editor');
    await expect(editor).toContainText('- [x] Persist me', { timeout: TEST_CONFIG.defaultTimeout });
  });

  test('code block content does not render as checkboxes', async ({ page }) => {
    const testPrefix = `cb_codeblock_${Date.now()}`;
    const noteContent = `# Code Block Test

- [ ] Real task 1

\`\`\`
- [ ] Fake task in code block
\`\`\`

- [ ] Real task 2`;

    await page.evaluate(async ({ prefix, content }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}.md`,
          content
        })
      });
    }, { testPrefix, content: noteContent });

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${testPrefix}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Switch to Preview mode
    const previewButton = page.locator('button').filter({ hasText: /^(Preview|预览)$/ }).first();
    await previewButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await previewButton.click();
    await page.waitForTimeout(500);

    // Wait for preview container
    await page.locator('.markdown-preview').waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Only 2 checkboxes should render (not the one in code block)
    const checkboxes = page.locator('input[data-interactive-checkbox]');
    await expect(checkboxes).toHaveCount(2, { timeout: TEST_CONFIG.defaultTimeout });

    // Click the second checkbox and verify it updates the correct line
    const secondCheckbox = checkboxes.nth(1);
    await secondCheckbox.click();
    await expect(secondCheckbox).toBeChecked({ timeout: TEST_CONFIG.defaultTimeout });

    // Switch to Edit mode
    const editButton = page.locator('button').filter({ hasText: /^(Edit|编辑)$/ }).first();
    await editButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await editButton.click();
    await page.waitForTimeout(500);

    const editor = page.locator('textarea#note-editor');
    const editorContent = await editor.inputValue();

    // Real task 2 should be toggled
    expect(editorContent).toContain('- [x] Real task 2');
    // Code block content should NOT be modified
    expect(editorContent).toContain('- [ ] Fake task in code block');
  });

  test('uppercase [X] and ordered list tasks work correctly', async ({ page }) => {
    const testPrefix = `cb_upper_${Date.now()}`;
    const noteContent = `- [X] Already done
- [ ] Pending uppercase
1. [ ] Ordered task`;

    await page.evaluate(async ({ prefix, content }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}.md`,
          content
        })
      });
    }, { testPrefix, content: noteContent });

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${testPrefix}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Switch to Preview mode
    const previewButton = page.locator('button').filter({ hasText: /^(Preview|预览)$/ }).first();
    await previewButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await previewButton.click();
    await page.waitForTimeout(500);

    // Wait for preview container
    await page.locator('.markdown-preview').waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Verify 3 checkboxes
    const checkboxes = page.locator('input[data-interactive-checkbox]');
    await expect(checkboxes).toHaveCount(3, { timeout: TEST_CONFIG.defaultTimeout });

    // First should be checked ([X])
    await expect(checkboxes.nth(0)).toBeChecked();

    // Toggle the second checkbox
    const secondCheckbox = checkboxes.nth(1);
    await secondCheckbox.click();
    await expect(secondCheckbox).toBeChecked({ timeout: TEST_CONFIG.defaultTimeout });

    // Toggle the ordered list checkbox
    const thirdCheckbox = checkboxes.nth(2);
    await thirdCheckbox.click();
    await expect(thirdCheckbox).toBeChecked({ timeout: TEST_CONFIG.defaultTimeout });

    // Verify content in editor
    const editButton = page.locator('button').filter({ hasText: /^(Edit|编辑)$/ }).first();
    await editButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await editButton.click();
    await page.waitForTimeout(500);

    const editor = page.locator('textarea#note-editor');
    const editorContent = await editor.inputValue();
    expect(editorContent).toContain('- [X] Already done');
    expect(editorContent).toContain('- [x] Pending uppercase');
    expect(editorContent).toContain('1. [x] Ordered task');
  });

  test('nested list task checkbox mapping is correct', async ({ page }) => {
    const testPrefix = `cb_nested_${Date.now()}`;
    const noteContent = `- [ ] Outer 1
    - [ ] Nested 1
    - [ ] Nested 2
- [ ] Outer 2`;

    await page.evaluate(async ({ prefix, content }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}.md`,
          content
        })
      });
    }, { testPrefix, content: noteContent });

    await page.reload();

    // Open the note
    const noteItem = page.locator(`text="${testPrefix}"`).first();
    await noteItem.click();
    await page.waitForTimeout(200);

    // Switch to Preview mode
    const previewButton = page.locator('button').filter({ hasText: /^(Preview|预览)$/ }).first();
    await previewButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await previewButton.click();
    await page.waitForTimeout(500);

    // Wait for preview container
    await page.locator('.markdown-preview').waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });

    // Verify 4 checkboxes
    const checkboxes = page.locator('input[data-interactive-checkbox]');
    await expect(checkboxes).toHaveCount(4, { timeout: TEST_CONFIG.defaultTimeout });

    // Click the 3rd checkbox (Nested 2)
    const thirdCheckbox = checkboxes.nth(2);
    await thirdCheckbox.click();
    await expect(thirdCheckbox).toBeChecked({ timeout: TEST_CONFIG.defaultTimeout });

    // Switch to Edit mode
    const editButton = page.locator('button').filter({ hasText: /^(Edit|编辑)$/ }).first();
    await editButton.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout });
    await editButton.click();
    await page.waitForTimeout(500);

    const editor = page.locator('textarea#note-editor');
    const editorContent = await editor.inputValue();

    // Only Nested 2 should be toggled
    expect(editorContent).toContain('- [ ] Outer 1');
    expect(editorContent).toContain('- [ ] Nested 1');
    expect(editorContent).toContain('    - [x] Nested 2');
    expect(editorContent).toContain('- [ ] Outer 2');
  });
});
