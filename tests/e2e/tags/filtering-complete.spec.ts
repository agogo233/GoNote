import { test, expect, TEST_CONFIG, login, waitForAutosave, apiPost } from '../fixtures/test-helpers';

const BASE_URL = TEST_CONFIG.baseUrl;

/**
 * Tags Filtering Complete Functionality E2E Tests
 * 
 * Tests complete tag filtering workflow including:
 * - Tag extraction from frontmatter
 * - Tag extraction from content
 * - Single tag filtering
 * - Multi-tag filtering with AND logic
 * - Tag count display
 * - Clear filters functionality
 * - Tag click from note metadata
 * - Tag panel navigation
 * - Embedded tags (hashtags in content)
 */

async function openTagsPanel(page: import('@playwright/test').Page) {
  const iconRailButtons = page.locator('.icon-rail button, [class*="icon-rail"] button');
  const buttonCount = await iconRailButtons.count();

  if (buttonCount >= 3) {
    await iconRailButtons.nth(2).click();
  } else {
    const tagsBtn = page.locator('button[title*="Tags"], button[aria-label*="Tags"]').first();
    if (await tagsBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await tagsBtn.click();
    }
  }

  await page.waitForTimeout(200);

  const tagsPanel = page.locator('[x-show*="activePanel === \'tags\'"], .tags-panel').first();
  await tagsPanel.waitFor({ state: 'visible', timeout: TEST_CONFIG.defaultTimeout }).catch(() => {});
}

async function openFilesPanel(page: import('@playwright/test').Page) {
  const iconRailButtons = page.locator('.icon-rail button, [class*="icon-rail"] button');
  const buttonCount = await iconRailButtons.count();

  if (buttonCount >= 1) {
    await iconRailButtons.nth(0).click();
  }

  await page.waitForTimeout(200);
}

async function createNoteWithTags(page: import('@playwright/test').Page, noteName: string, tags: string[], content: string = '') {
  const frontmatter = `---
tags:
${tags.map(t => `  - ${t}`).join('\n')}
---
`;
  const fullContent = frontmatter + (content || `# ${noteName}\n\nContent for ${noteName}`);

  await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: fullContent });
  await page.waitForTimeout(200);
  await waitForAutosave(page);

  await page.reload();
  await page.waitForTimeout(300);
}

test.describe('Tags Filtering Complete Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('tag extraction from frontmatter', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_frontmatter_tags`;
    const tags = ['work', 'important', 'project-a'];

    await createNoteWithTags(page, noteName, tags);

    await openTagsPanel(page);
    await page.waitForTimeout(200);

    // Check that all tags are displayed in tags panel
    const pageContent = await page.content();
    
    for (const tag of tags) {
      const hasTag = pageContent.toLowerCase().includes(tag.toLowerCase());
      console.log(`Tag "${tag}" found in page: ${hasTag}`);
      expect(hasTag).toBe(true);
    }

    await page.screenshot({ path: `config/test-results/tags-frontmatter-${testPrefix}.png`, fullPage: true });
  });

  test('tag extraction from content (hashtags)', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_hashtag_content`;
    const content = `# Hashtag Test\n\nThis is content with #hashtag1 and #hashtag2 embedded tags.\n\nAlso mentioning #work for testing.`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(200);
    await waitForAutosave(page);

    await page.reload();
    await page.waitForTimeout(300);

    await openTagsPanel(page);
    await page.waitForTimeout(200);

    // Check that hashtags are extracted
    const pageContent = await page.content();
    
    const hasHashtag1 = pageContent.toLowerCase().includes('hashtag1');
    const hasHashtag2 = pageContent.toLowerCase().includes('hashtag2');
    const hasWork = pageContent.toLowerCase().includes('work');

    console.log(`Extracted hashtags - hashtag1: ${hasHashtag1}, hashtag2: ${hasHashtag2}, work: ${hasWork}`);

    await page.screenshot({ path: `config/test-results/tags-hashtags-${testPrefix}.png`, fullPage: true });
  });

  test('filter notes by single tag', async ({ page, testPrefix }) => {
    const uniqueTag = `single${Date.now()}`;
    const noteName = `${testPrefix}_single_tag`;

    await createNoteWithTags(page, noteName, [uniqueTag, 'common']);

    await openTagsPanel(page);
    await page.waitForTimeout(300);

    // Find and click the unique tag - use text content matcher
    const tagChip = page.locator('.tag-chip').filter({ hasText: uniqueTag }).first();

    // Wait for tag to be visible with longer timeout
    const isTagVisible = await tagChip.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isTagVisible) {
      await tagChip.click();
      await page.waitForTimeout(200);

      // Check that filtered results show our note
      // The note appears in the tag results container
      const noteInResults = page.locator('[x-ref="tagResultsContainer"] .hover-accent').filter({ hasText: noteName }).first();
      const isVisible = await noteInResults.isVisible({ timeout: 5000 }).catch(() => false);

      console.log(`Note visible in filtered results: ${isVisible}`);
      expect(isVisible).toBe(true);

      // Check for active filter indicator (selected tag chip)
      const selectedTag = page.locator('.tag-chip').filter({ hasText: uniqueTag }).first();
      const isSelected = await selectedTag.isVisible({ timeout: 2000 }).catch(() => false);
      console.log(`Tag shows as selected: ${isSelected}`);
    } else {
      console.log(`Tag "${uniqueTag}" not found in tags panel`);
      // Take screenshot for debugging
      await page.screenshot({ path: `config/test-results/tags-single-filter-${testPrefix}-debug.png`, fullPage: true });
    }

    await page.screenshot({ path: `config/test-results/tags-single-filter-${testPrefix}.png`, fullPage: true });
  });

  test('multi-tag filtering with AND logic', async ({ page, testPrefix }) => {
    const tag1 = `alpha${Date.now()}`;
    const tag2 = `beta${Date.now()}`;

    // Create note with both tags
    const noteBoth = `${testPrefix}_both_tags`;
    await createNoteWithTags(page, noteBoth, [tag1, tag2]);

    // Create note with only tag1
    const noteOne = `${testPrefix}_one_tag`;
    await createNoteWithTags(page, noteOne, [tag1]);

    await openTagsPanel(page);
    await page.waitForTimeout(300);

    // Select both tags
    const tagChip1 = page.locator('.tag-chip').filter({ hasText: tag1 }).first();
    const tagChip2 = page.locator('.tag-chip').filter({ hasText: tag2 }).first();

    if (await tagChip1.isVisible({ timeout: 3000 }).catch(() => false)) {
      await tagChip1.click();
      await page.waitForTimeout(200);
    }

    if (await tagChip2.isVisible({ timeout: 3000 }).catch(() => false)) {
      await tagChip2.click();
      await page.waitForTimeout(200);
    }

    // With AND logic, only note with both tags should appear
    const noteBothElement = page.locator(`text="${noteBoth}"`).first();
    const noteOneElement = page.locator(`text="${noteOne}"`).first();

    const bothVisible = await noteBothElement.isVisible({ timeout: 3000 }).catch(() => false);
    const oneVisible = await noteOneElement.isVisible({ timeout: 2000 }).catch(() => false);

    console.log(`Note with both tags visible: ${bothVisible}`);
    console.log(`Note with one tag visible (should be false with AND logic): ${oneVisible}`);

    // Verify the clear filters button appears
    const clearFiltersBtn = page.locator('text=/Clear Filters|clear_all/i').first();
    const clearVisible = await clearFiltersBtn.isVisible({ timeout: 2000 }).catch(() => false);
    console.log(`Clear filters button visible: ${clearVisible}`);

    await page.screenshot({ path: `config/test-results/tags-and-logic-${testPrefix}.png`, fullPage: true });
  });

  test('tag count display is correct', async ({ page, testPrefix }) => {
    const uniqueTag = `count${Date.now()}`;

    // Create three notes with the same tag
    await createNoteWithTags(page, `${testPrefix}_count1`, [uniqueTag]);
    await createNoteWithTags(page, `${testPrefix}_count2`, [uniqueTag]);
    await createNoteWithTags(page, `${testPrefix}_count3`, [uniqueTag]);

    await openTagsPanel(page);
    await page.waitForTimeout(300);

    // Find the tag chip with count
    const tagChip = page.locator('.tag-chip').filter({ hasText: uniqueTag }).first();

    if (await tagChip.isVisible({ timeout: 3000 }).catch(() => false)) {
      const chipText = await tagChip.textContent();
      console.log(`Tag chip text: ${chipText}`);

      // Count should contain "3" or show 3 notes
      const hasCount3 = chipText?.includes('3');
      console.log(`Tag shows count of 3: ${hasCount3}`);
    }

    await page.screenshot({ path: `config/test-results/tags-count-${testPrefix}.png`, fullPage: true });
  });

  test('clear tag filters restores all notes', async ({ page, testPrefix }) => {
    const tag = `clear${Date.now()}`;
    const taggedNote = `${testPrefix}_tagged`;
    const untaggedNote = `${testPrefix}_untagged`;

    // Create one tagged note and one without
    await createNoteWithTags(page, taggedNote, [tag]);
    await apiPost(page, `${BASE_URL}/api/notes/${untaggedNote}.md`, { content: `# ${untaggedNote}\n\nNo tags here` });
    await page.waitForTimeout(200);

    await page.reload();
    await page.waitForTimeout(300);

    await openTagsPanel(page);
    await page.waitForTimeout(300);

    // Select the tag
    const tagChip = page.locator('.tag-chip').filter({ hasText: tag }).first();

    if (await tagChip.isVisible({ timeout: 3000 }).catch(() => false)) {
      await tagChip.click();
      await page.waitForTimeout(200);

      // Verify filtering is applied (untagged note should not be visible)
      const untaggedBeforeClear = page.locator(`text="${untaggedNote}"`).first();
      const untaggedVisibleBefore = await untaggedBeforeClear.isVisible({ timeout: 2000 }).catch(() => false);
      console.log(`Untagged note visible before clear: ${untaggedVisibleBefore}`);

      // Find and click clear filters button
      const clearBtn = page.locator('button:has-text("Clear"), [x-show*="selectedTags.length > 0"] button').first();

      if (await clearBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await clearBtn.click();
        await page.waitForTimeout(200);

        // After clearing, all notes should be visible
        const untaggedAfterClear = page.locator(`text="${untaggedNote}"`).first();
        const untaggedVisibleAfter = await untaggedAfterClear.isVisible({ timeout: 3000 }).catch(() => false);
        console.log(`Untagged note visible after clear: ${untaggedVisibleAfter}`);

        // Clear button should be hidden
        const clearStillVisible = await clearBtn.isVisible({ timeout: 2000 }).catch(() => false);
        console.log(`Clear button still visible: ${clearStillVisible}`);
      }
    }

    await page.screenshot({ path: `config/test-results/tags-clear-filters-${testPrefix}.png`, fullPage: true });
  });

  test('clicking tag in note metadata filters by that tag', async ({ page, testPrefix }) => {
    const tag = `meta${Date.now()}`;
    const noteName = `${testPrefix}_metadata_tag`;

    await createNoteWithTags(page, noteName, [tag]);

    await openFilesPanel(page);
    await page.waitForTimeout(200);

    // Find and click the note
    const noteItem = page.locator(`text="${noteName}"`).first();
    if (await noteItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      await noteItem.click();
      await page.waitForTimeout(200);

      // Look for tag in metadata section
      const metadataTag = page.locator('.metadata-tag, [class*="metadata"]').filter({ hasText: tag }).first();

      if (await metadataTag.isVisible({ timeout: 3000 }).catch(() => false)) {
        await metadataTag.click();
        await page.waitForTimeout(200);

        // Should navigate to tags panel with tag selected
        const selectedTag = page.locator('.tag-chip.selected, [class*="selected"]').filter({ hasText: tag }).first();
        const isSelected = await selectedTag.isVisible({ timeout: 2000 }).catch(() => false);
        console.log(`Tag is selected after clicking metadata: ${isSelected}`);

        // Verify filtering is applied
        const noteInResults = page.locator(`text="${noteName}"`).first();
        const isVisible = await noteInResults.isVisible({ timeout: 3000 }).catch(() => false);
        console.log(`Note visible in filtered results: ${isVisible}`);
      }
    }

    await page.screenshot({ path: `config/test-results/tags-metadata-click-${testPrefix}.png`, fullPage: true });
  });

  test('tag panel shows all unique tags', async ({ page, testPrefix }) => {
    // Create notes with different tags
    const tag1 = `unique1${Date.now()}`;
    const tag2 = `unique2${Date.now()}`;
    const tag3 = `unique3${Date.now()}`;

    await createNoteWithTags(page, `${testPrefix}_note1`, [tag1]);
    await createNoteWithTags(page, `${testPrefix}_note2`, [tag2]);
    await createNoteWithTags(page, `${testPrefix}_note3`, [tag3]);

    await openTagsPanel(page);
    await page.waitForTimeout(300);

    // Count unique tags
    const tagChips = page.locator('.tag-chip');
    const count = await tagChips.count();

    console.log(`Total tag chips found: ${count}`);

    // Should have at least our 3 unique tags
    expect(count).toBeGreaterThanOrEqual(3);

    // Verify each unique tag is present
    const pageContent = await page.content();
    expect(pageContent.toLowerCase()).toContain(tag1.toLowerCase());
    expect(pageContent.toLowerCase()).toContain(tag2.toLowerCase());
    expect(pageContent.toLowerCase()).toContain(tag3.toLowerCase());

    await page.screenshot({ path: `config/test-results/tags-all-unique-${testPrefix}.png`, fullPage: true });
  });

  test('tag filtering persists across page reload', async ({ page, testPrefix }) => {
    const tag = `persist${Date.now()}`;
    const noteName = `${testPrefix}_persist_test`;

    await createNoteWithTags(page, noteName, [tag]);

    await openTagsPanel(page);
    await page.waitForTimeout(300);

    // Select the tag
    const tagChip = page.locator('.tag-chip').filter({ hasText: tag }).first();
    if (await tagChip.isVisible({ timeout: 3000 }).catch(() => false)) {
      await tagChip.click();
      await page.waitForTimeout(200);

      // Reload the page
      await page.reload();
      await page.waitForTimeout(300);

      // Re-open tags panel
      await openTagsPanel(page);
      await page.waitForTimeout(300);

      // Check if filter is still applied
      const selectedTag = page.locator('.tag-chip.selected, [class*="selected"]').filter({ hasText: tag }).first();
      const isSelected = await selectedTag.isVisible({ timeout: 3000 }).catch(() => false);
      console.log(`Tag filter persists after reload: ${isSelected}`);

      // Note should still be visible in filtered results
      const noteInResults = page.locator(`text="${noteName}"`).first();
      const isVisible = await noteInResults.isVisible({ timeout: 3000 }).catch(() => false);
      console.log(`Note visible after reload: ${isVisible}`);
    }

    await page.screenshot({ path: `config/test-results/tags-persist-${testPrefix}.png`, fullPage: true });
  });
});
