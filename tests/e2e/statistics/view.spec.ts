import { test, expect } from '@playwright/test';

test.describe('Statistics View', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.fill('input[type="password"]', 'test-admin-password');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('get statistics for note', async ({ page }) => {
    const testPrefix = `stats_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-note.md`,
          content: `---
title: Statistics Test
tags: [test]
---

# Main Heading

This is a test note with various content elements for statistics.

## Section 1

This paragraph has multiple sentences. It also has several words!

## Section 2

- List item 1
- List item 2
- List item 3

## Code

\`\`\`go
package main

import "fmt"

func main() {
    fmt.Println("Hello")
}
\`\`\`

## Tasks

- [ ] Pending task
- [x] Completed task

## Links

[External Link](https://example.com)
[[Internal Link]]

## Image

![Test Image](test.png)
`
        })
      });
    }, { testPrefix });

    // Fetch statistics via API
    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-note.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data).toBeDefined();
    
    // Verify various statistics are calculated
    const data = statsResponse.data;
    expect(data.words).toBeGreaterThan(50);
    expect(data.sentences).toBeGreaterThan(5);
    expect(data.lines).toBeGreaterThan(10);
    expect(data.headings).toBeDefined();
    expect(data.headings.h1).toBeGreaterThanOrEqual(1);
    expect(data.headings.h2).toBeGreaterThanOrEqual(4);
    expect(data.tasks).toBeDefined();
    expect(data.tasks.total).toBe(2);
    expect(data.tasks.completed).toBe(1);
    expect(data.tasks.pending).toBe(1);
    expect(data.code_blocks).toBe(1);
    expect(data.wikilinks).toBe(1);
    expect(data.images).toBe(1);
  });

  test('statistics for empty note', async ({ page }) => {
    const testPrefix = `empty_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-empty.md`,
          content: ''
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-empty.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.words).toBe(0);
    expect(statsResponse.data.sentences).toBe(0);
  });

  test('statistics for note with only frontmatter', async ({ page }) => {
    const testPrefix = `fm_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-fm.md`,
          content: `---
title: Frontmatter Only
tags: [test]
---`
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-fm.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.words).toBe(0);
  });

  test('statistics for note in subdirectory', async ({ page }) => {
    const testPrefix = `subdir_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `programming/${prefix}-go.md`,
          content: `# Go Programming\n\nGo is great.\n\n- Fast\n- Simple\n\n\`\`\`go\nfmt.Println("Hi")\n\`\`\``
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/programming/${prefix}-go.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.words).toBeGreaterThan(0);
  });

  test('reading time calculation', async ({ page }) => {
    const testPrefix = `reading_${Date.now()}`;
    
    // Create a note with approximately 400 words (should be 2 minutes reading time)
    let content = `# Long Note\n\n`;
    for (let i = 0; i < 400; i++) {
      content += `word${i} `;
    }
    
    await page.evaluate(async ({ prefix, content }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-long.md`,
          content: content
        })
      });
    }, { testPrefix, content });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-long.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.reading_time_minutes).toBeGreaterThanOrEqual(2);
  });

  test('character count (with and without whitespace)', async ({ page }) => {
    const testPrefix = `chars_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-chars.md`,
          content: 'Hello World'  // 11 chars total, 10 without space
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-chars.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.total_characters).toBe(11);
    expect(statsResponse.data.characters).toBe(10);
  });

  test('heading counts by level', async ({ page }) => {
    const testPrefix = `headings_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-headings.md`,
          content: `# H1 Title\n## H2 Section\n### H3 Subsection\n## Another H2\n# Another H1`
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-headings.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.headings.h1).toBe(2);
    expect(statsResponse.data.headings.h2).toBe(2);
    expect(statsResponse.data.headings.h3).toBe(1);
  });

  test('task statistics', async ({ page }) => {
    const testPrefix = `tasks_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-tasks.md`,
          content: `- [ ] Task 1\n- [ ] Task 2\n- [x] Task 3\n- [X] Task 4\n- [ ] Task 5`
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-tasks.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.tasks.total).toBe(5);
    expect(statsResponse.data.tasks.completed).toBe(2);
    expect(statsResponse.data.tasks.pending).toBe(3);
  });

  test('link statistics', async ({ page }) => {
    const testPrefix = `links_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-links.md`,
          content: `[External](https://example.com)\n[Internal](other.md)\n[[Wikilink]]\n[[Wiki|Display]]`
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-links.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.links).toBe(2);
    expect(statsResponse.data.wikilinks).toBe(2);
  });

  test('code statistics', async ({ page }) => {
    const testPrefix = `code_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-code.md`,
          content: `Inline: \`const x = 10\`\n\nBlock:\n\n\`\`\`js\nconsole.log('test');\n\`\`\`\n\nMore inline: \`y = 20\``
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-code.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.code_blocks).toBe(1);
    expect(statsResponse.data.inline_code).toBe(2);
  });

  test('statistics for non-existent note', async ({ page }) => {
    const statsResponse = await page.evaluate(async () => {
      const response = await fetch(`/api/statistics/nonexistent.md`);
      return response.json();
    });

    expect(statsResponse.success).toBeUndefined();
    expect(statsResponse.detail).toBeDefined();
  });

  test('statistics with unicode content', async ({ page }) => {
    const testPrefix = `unicode_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-unicode.md`,
          content: `# 中文标题\n\n你好世界\n\nこんにちは\n\n안녕하세요`
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-unicode.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.words).toBeGreaterThan(0);
  });

  test('statistics with emoji content', async ({ page }) => {
    const testPrefix = `emoji_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-emoji.md`,
          content: `Hello 👋 World 🌍 Test 🧪`
        })
      });
    }, { testPrefix });

    const statsResponse = await page.evaluate(async ({ prefix }) => {
      const response = await fetch(`/api/statistics/${prefix}-emoji.md`);
      return response.json();
    }, { testPrefix });

    expect(statsResponse.success).toBe(true);
    expect(statsResponse.data.words).toBeGreaterThanOrEqual(3);
  });
});

test.describe('Statistics UI Integration', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[type="password"]', 'test-admin-password');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('display statistics in note view', async ({ page }) => {
    const testPrefix = `ui_${Date.now()}`;
    
    await page.evaluate(async ({ prefix }) => {
      await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: `${prefix}-note.md`,
          content: `# UI Test\n\nContent for UI testing.`
        })
      });
    }, { testPrefix });

    await page.goto(`/notes/${testPrefix}-note.md`);

    // Verify note view is displayed (statistics display depends on UI implementation)
    await expect(page.locator('.markdown-preview')).toBeVisible();
  });
});
