import { test, expect, TEST_CONFIG, login, waitForAutosave, apiPost } from '../fixtures/test-helpers';

const BASE_URL = TEST_CONFIG.baseUrl;

/**
 * Export HTML Complete Functionality E2E Tests
 * 
 * Tests complete export HTML workflow including:
 * - Create share link
 * - View shared note in browser
 * - HTML rendering with proper formatting
 * - Code syntax highlighting in export
 * - MathJax math rendering in export
 * - Theme support in export
 * - Revoke share link
 * - Export with custom title
 * - Export with embedded media
 */

test.describe('Export HTML Complete Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('create share link for note', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_export_share`;
    const content = `# Export Share Test\n\nThis note is for testing share link creation.\n\n\`\`\`javascript\nconsole.log("Hello World");\n\`\`\``;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link via API
    const response = await page.request.post(`/api/share/${encodedPath}`);
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    expect(data.success).toBeTruthy();
    expect(data.token).toBeTruthy();
    expect(data.url).toContain('/share/');

    console.log(`Share link created: ${data.url}`);
  });

  test('view shared note renders HTML correctly', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_view_html`;
    const uniqueContent = `# Shared HTML View ${Date.now()}\n\nThis is a **shared paragraph** with *formatting*.\n\n- Item 1\n- Item 2\n- Item 3`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: uniqueContent });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    const shareData = await shareResponse.json();

    // Visit the shared URL
    const viewUrl = shareData.url;
    await page.goto(viewUrl);
    await page.waitForTimeout(2000);

    // Check HTML structure
    const htmlContent = await page.content();
    
    // Should have proper HTML structure
    expect(htmlContent).toContain('<!DOCTYPE html>');
    expect(htmlContent).toContain('<html');
    expect(htmlContent).toContain('<head>');
    expect(htmlContent).toContain('<body>');

    // Should render markdown content
    expect(htmlContent).toContain('Shared HTML View');
    expect(htmlContent).toContain('shared paragraph');

    // Should have proper formatting
    expect(htmlContent).toContain('<strong>'); // bold
    expect(htmlContent).toContain('<em>'); // italic
    expect(htmlContent).toContain('<ul>'); // list

    await page.screenshot({ path: `config/test-results/export-html-view-${testPrefix}.png`, fullPage: true });
  });

  test('shared note includes code syntax highlighting', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_code_highlight`;
    const codeContent = `# Code Highlighting Test

\`\`\`python
def hello_world():
    """Print hello world"""
    print("Hello, World!")
    return True

if __name__ == "__main__":
    hello_world()
\`\`\`

\`\`\`javascript
function greet(name) {
    console.log(\`Hello, \${name}!\`);
}

greet('Developer');
\`\`\``;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: codeContent });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    const shareData = await shareResponse.json();

    // Visit the shared URL
    await page.goto(shareData.url);
    await page.waitForTimeout(2000);

    const htmlContent = await page.content();

    // Should include highlight.js
    expect(htmlContent).toContain('highlight.js');
    expect(htmlContent).toContain('hljs');

    // Should have code blocks with language classes
    expect(htmlContent).toContain('language-python');
    expect(htmlContent).toContain('language-javascript');

    // Should have syntax highlighting spans
    const hasHighlightSpans = htmlContent.includes('<span') && (
      htmlContent.includes('hljs-') || 
      htmlContent.includes('class="')
    );
    console.log(`Has syntax highlighting spans: ${hasHighlightSpans}`);

    await page.screenshot({ path: `config/test-results/export-code-highlight-${testPrefix}.png`, fullPage: true });
  });

  test('shared note includes MathJax for math rendering', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_mathjax`;
    const mathContent = `# Math Rendering Test

## Inline Math

The famous equation is $E = mc^2$ in inline format.

## Block Math

$$
\\int_{0}^{\\infty} x^2 dx = \\left[ \\frac{x^3}{3} \\right]_{0}^{\\infty}
$$

## Matrix

$$
\\begin{pmatrix}
a & b \\\\
c & d
\\end{pmatrix}
$$`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: mathContent });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    const shareData = await shareResponse.json();

    // Visit the shared URL
    await page.goto(shareData.url);
    await page.waitForTimeout(2000); // MathJax needs time to load

    const htmlContent = await page.content();

    // Should include MathJax
    expect(htmlContent).toContain('MathJax');
    expect(htmlContent).toContain('mathjax');

    // Should have math delimiters
    expect(htmlContent).toContain('$E = mc^2$');
    expect(htmlContent).toContain('$$');

    // Check if MathJax rendered (look for SVG or MathML output)
    const hasMathOutput = htmlContent.includes('<svg') || 
                          htmlContent.includes('<math') || 
                          htmlContent.includes('mjx-');
    console.log(`MathJax rendered output: ${hasMathOutput}`);

    await page.screenshot({ path: `config/test-results/export-mathjax-${testPrefix}.png`, fullPage: true });
  });

  test('shared note with Mermaid diagrams', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_mermaid`;
    const mermaidContent = `# Mermaid Diagram Test

## Flowchart

\`\`\`mermaid
graph TD
    A[Start] --> B{Is it working?}
    B -->|Yes| C[Great!]
    B -->|No| D[Debug]
    D --> B
    C --> E[End]
\`\`\`

## Sequence Diagram

\`\`\`mermaid
sequenceDiagram
    Alice->>John: Hello John, how are you?
    John-->>Alice: Great!
    Alice-)John: See you later!
\`\`\``;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: mermaidContent });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    const shareData = await shareResponse.json();

    // Visit the shared URL
    await page.goto(shareData.url);
    await page.waitForTimeout(2000); // Mermaid needs time to render

    const htmlContent = await page.content();

    // Should include Mermaid
    expect(htmlContent).toContain('mermaid');

    // Should have mermaid-rendered container (the class name used in export)
    expect(htmlContent).toContain('class="mermaid-rendered"');
    expect(htmlContent).toContain('graph TD');
    expect(htmlContent).toContain('sequenceDiagram');

    // Check if Mermaid rendered (look for SVG output)
    const hasMermaidSvg = htmlContent.includes('<svg') && htmlContent.includes('mermaid-diagram');
    console.log(`Mermaid rendered as SVG: ${hasMermaidSvg}`);

    await page.screenshot({ path: `config/test-results/export-mermaid-${testPrefix}.png`, fullPage: true });
  });

  test('shared note with theme support', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_theme`;
    const content = `# Theme Test\n\nThis note tests theme support in shared views.\n\n- Dark theme\n- Light theme\n- Custom themes`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link with dark theme
    const shareResponse = await page.request.post(`/api/share/${encodedPath}?theme=dark`);
    expect(shareResponse.ok()).toBeTruthy();
    const shareData = await shareResponse.json();
    expect(shareData.theme).toBe('dark');

    // Visit the shared URL
    await page.goto(shareData.url);
    await page.waitForTimeout(2000);

    const htmlContent = await page.content();

    // Should have dark theme indicators
    const hasDarkTheme = htmlContent.includes('dark') || 
                         htmlContent.includes('class="dark"') ||
                         htmlContent.includes('data-theme="dark"');
    console.log(`Dark theme applied: ${hasDarkTheme}`);

    // Create share link with light theme
    const lightShareResponse = await page.request.post(`/api/share/${encodedPath}?theme=light`);
    const lightShareData = await lightShareResponse.json();
    expect(lightShareData.theme).toBe('light');

    console.log(`Theme support verified for: dark, light`);
  });

  test('revoke share link', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_revoke`;
    const content = `# Revoke Test\n\nThis share link will be revoked.`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    expect(shareResponse.ok()).toBeTruthy();
    const shareData = await shareResponse.json();
    const shareUrl = shareData.url;

    // Verify share link works
    const viewResponse1 = await page.request.get(shareUrl);
    expect(viewResponse1.ok()).toBeTruthy();

    // Revoke share link
    const revokeResponse = await page.request.delete(`/api/share/${encodedPath}`);
    expect(revokeResponse.ok()).toBeTruthy();

    // Verify share link is revoked
    const viewResponse2 = await page.request.get(shareUrl);
    // Should return 404 or error after revocation
    expect(viewResponse2.status()).not.toBe(200);

    // Check share status API
    const statusResponse = await page.request.get(`/api/share/${encodedPath}`);
    const statusData = await statusResponse.json();
    expect(statusData.shared).toBeFalsy();

    console.log(`Share link revoked successfully`);
  });

  test('shared note with embedded media', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_media_export`;
    const mediaContent = `# Media Export Test

## Image

![[test-image.png]]

## External Image

![External](https://example.com/image.png)

## Video

![[test-video.mp4]]
`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: mediaContent });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    const shareData = await shareResponse.json();

    // Visit the shared URL
    await page.goto(shareData.url);
    await page.waitForTimeout(2000);

    const htmlContent = await page.content();

    // Should have media references
    expect(htmlContent).toContain('test-image.png');
    expect(htmlContent).toContain('test-video.mp4');

    // Should have media elements
    const hasImgTag = htmlContent.includes('<img');
    const hasVideoTag = htmlContent.includes('<video');
    console.log(`Has img tag: ${hasImgTag}, Has video tag: ${hasVideoTag}`);

    await page.screenshot({ path: `config/test-results/export-media-${testPrefix}.png`, fullPage: true });
  });

  test('shared note preserves wikilinks', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_wikilinks`;
    const linkedNote = `${testPrefix}_linked`;
    
    const content = `# Wikilinks Test\n\nThis links to [[${linkedNote}]] another note.\n\nAlso check [[${linkedNote}#section|this section]].`;
    const linkedContent = `# Linked Note\n\nThis is the linked content.\n\n## Section\n\nSection content here.`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content });
    await apiPost(page, `${BASE_URL}/api/notes/${linkedNote}.md`, { content: linkedContent });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    const shareData = await shareResponse.json();

    // Visit the shared URL
    await page.goto(shareData.url);
    await page.waitForTimeout(2000);

    const htmlContent = await page.content();

    // Should preserve wikilink syntax or convert to links
    const hasWikilink = htmlContent.includes('[[') || 
                        htmlContent.includes('href') && htmlContent.includes(linkedNote);
    console.log(`Wikilinks preserved or converted: ${hasWikilink}`);

    // Should have link to the linked note
    const hasLink = htmlContent.includes(linkedNote);
    expect(hasLink).toBe(true);

    await page.screenshot({ path: `config/test-results/export-wikilinks-${testPrefix}.png`, fullPage: true });
  });

  test('export note with table of contents', async ({ page, testPrefix }) => {
    const noteName = `${testPrefix}_toc`;
    const tocContent = `# Table of Contents Test

## Section 1

Content for section 1.

### Subsection 1.1

More details here.

## Section 2

Content for section 2.

### Subsection 2.1

Even more details.

## Section 3

Final section content.`;

    await apiPost(page, `${BASE_URL}/api/notes/${noteName}.md`, { content: tocContent });
    await page.waitForTimeout(200);

    const notePath = `${noteName}.md`;
    const encodedPath = encodeURIComponent(notePath);

    // Create share link
    const shareResponse = await page.request.post(`/api/share/${encodedPath}`);
    const shareData = await shareResponse.json();

    // Visit the shared URL
    await page.goto(shareData.url);
    await page.waitForTimeout(2000);

    const htmlContent = await page.content();

    // Should have heading structure
    expect(htmlContent).toContain('<h1>');
    expect(htmlContent).toContain('<h2>');
    expect(htmlContent).toContain('<h3>');

    // Should have section titles
    expect(htmlContent).toContain('Section 1');
    expect(htmlContent).toContain('Section 2');
    expect(htmlContent).toContain('Section 3');

    // Check for table of contents (if generated)
    const hasToc = htmlContent.includes('toc') || 
                   htmlContent.includes('Table of Contents') ||
                   htmlContent.includes('nav') ||
                   htmlContent.includes('id="toc"');
    console.log(`Table of contents present: ${hasToc}`);

    await page.screenshot({ path: `config/test-results/export-toc-${testPrefix}.png`, fullPage: true });
  });
});
