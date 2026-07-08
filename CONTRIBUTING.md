# Contributing to GoNote

Thank you for your interest in contributing to GoNote! This guide will help you get started with contributing to the project, whether it's code, documentation, translations, or bug reports.

---

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting Pull Requests](#submitting-pull-requests)
  - [Translating Documentation](#translating-documentation)
  - [Improving Documentation](#improving-documentation)
- [Development Setup](#development-setup)
- [Code Standards](#code-standards)
- [Commit Guidelines](#commit-guidelines)
- [Testing Requirements](#testing-requirements)
- [Review Process](#review-process)
- [Community](#community)

---

## 📜 Code of Conduct

This project adheres to a code of conduct based on the [Contributor Covenant](https://www.contributor-covenant.org/). By participating, you are expected to:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept responsibility and apologize for mistakes

Please report unacceptable behavior to the project maintainers.

---

## 🤝 How to Contribute

### Reporting Bugs

If you find a bug, please **open an issue** before submitting a pull request, especially for significant changes.

**Before submitting a bug report:**

1. Check if the bug has already been reported in [Issues](../../issues)
2. Ensure you're using the latest version
3. Try to reproduce the issue consistently

**When submitting a bug report, please include:**

- **Clear title** - Summarize the issue in one line
- **Environment** - Version, OS, browser (if applicable), deployment method (Docker/standalone)
- **Steps to reproduce** - Detailed, numbered steps
- **Expected behavior** - What should happen
- **Actual behavior** - What actually happens
- **Screenshots** - If applicable, for visual issues
- **Logs** - Server logs if available (redact sensitive data)
- **Additional context** - Any other relevant information

**Template:**

```markdown
## Bug Description
[Describe the bug]

## Environment
- GoNote Version: [e.g., v0.25.0]
- OS: [e.g., Ubuntu 22.04, macOS 14, Windows 11]
- Browser: [if UI issue, e.g., Chrome 120, Firefox 121]
- Deployment: [Docker / Standalone / Other]

## Steps to Reproduce
1. [First step]
2. [Second step]
3. [Further steps...]

## Expected Behavior
[What you expected to see]

## Actual Behavior
[What actually happened]

## Screenshots / Logs
[Attach screenshots or log excerpts]
```

### Suggesting Features

We welcome feature suggestions! Before proposing a new feature:

1. Check if the feature has already been requested
2. Consider if the feature aligns with GoNote's philosophy (simple, self-hosted, markdown-based)
3. Open an issue with the tag "enhancement"

**Feature request template:**

```markdown
## Feature Description
[Describe the feature you'd like to see]

## Problem / Use Case
[What problem does this solve? Who benefits?]

## Proposed Solution
[How should this work? Be specific]

## Alternatives Considered
[Any alternative approaches?]

## Additional Context
[Screenshots, examples, references to similar features in other apps]
```

### Submitting Pull Requests

**Before starting work:**

1. **Open an issue** to discuss the change, especially for:
   - Major features or architectural changes
   - UI/UX modifications
   - Breaking changes
   
   This prevents wasted effort if the change isn't aligned with project direction.

2. **Fork the repository** and create a branch:
   ```bash
   git checkout -b feature/your-feature-name
   # or for bug fixes:
   git checkout -b fix/issue-description
   ```

3. **Follow the commit guidelines** (see [Commit Guidelines](#commit-guidelines))

**Pull Request Requirements:**

- ✅ Code follows GoNote's style (see [Code Standards](#code-standards))
- ✅ All tests pass (E2E and Go unit tests)
- ✅ New functionality includes tests
- ✅ Documentation updated (if applicable)
  - English docs updated
  - Chinese docs updated (if affecting user-facing features)
- ✅ No sensitive data in code or logs
- ✅ Run `make build` to ensure it compiles
- ✅ Run `make test` and `make test-e2e` locally before submitting

**Pull Request Template:**

When you open a PR, please fill out the template:

```markdown
## Description
[Brief description of changes]

## Related Issues
[Link to related issues, e.g., Closes #123]

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Translation update

## Testing
- [ ] Tests added/updated (if applicable)
- [ ] Manual testing performed
- [ ] E2E tests pass
- [ ] Go unit tests pass

## Checklist
- [ ] Code follows project style
- [ ] Self-review completed
- [ ] Documentation updated (English)
- [ ] Documentation updated (Chinese) - if user-facing
- [ ] No hardcoded secrets or credentials
- [ ] All tests pass locally

## Screenshots (if UI changes)
[Attach screenshots]

## Additional Notes
[Anything else reviewers should know]
```

**Review Process:**

1. **Automated checks**: CI runs tests and linting
2. **Maintainer review**: At least one maintainer will review
3. **Address feedback**: Make requested changes or discuss alternatives
4. **Merge**: Once approved and CI passes, a maintainer will merge

### Translating Documentation

GoNote supports multiple languages and welcomes translation contributions!

**Supported languages:**
- English (en-US) - default
- Simplified Chinese (zh-CN) - partial

**Adding translations:**

1. **For user-facing documentation** (in `project-docs/user-guide/`):
   - Documentation files are in Simplified Chinese (zh-CN)
   - Follow the same structure and formatting
   - Keep technical terms consistent across documents

2. **For developer documentation** (in `project-docs/developer-guide/`):
   - Documentation files are in Simplified Chinese (zh-CN)
   - Check `project-docs/README.md` for current status

3. **For UI translations** (locale files):
   - Locale files are in `go/internal/locale/`
   - Add new language JSON file (e.g., `zh-CN.json`)
   - Follow the existing structure
   - See existing files for reference
   - Submit PR with new locale file

**Translation guidelines:**
- Use formal/mature tone (avoid slang)
- Keep technical terms consistent (e.g., "deployment" → "部署")
- Translate concepts, not literal words
- Maintain formatting (links, code blocks, emphasis)
- Preserve original file structure

### Improving Documentation

Documentation improvements are always welcome:

- Fix typos and grammar
- Clarify unclear explanations
- Update outdated content
- Add examples
- Improve formatting and structure

When updating documentation:

1. Update both English and Chinese versions (if applicable)
2. Ensure code examples are accurate and tested
3. Preserve the original structure
4. Check relative links in `.md` files

---

## 🔧 Development Setup

### Prerequisites

- **Go 1.24+** (latest stable recommended)
- **Node.js 18+** (for CSS build and E2E tests)
- **Docker** (optional, for running via container)
- **Make** (optional, for using Makefile targets)

### Local Development

1. **Clone the repository:**
   ```bash
   git clone https://github.com/gamosoft/gonote.git
   cd gonote
   ```

2. **Checkout a new branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Run the application (from project root):**
   ```bash
   make run
   # or
   go run go/cmd/server/main.go --config go/config.yaml
   ```
   
   ⚠️ **Important:** Always run from the **project root directory**, not from `go/`.

4. **Access the app:**
   Open http://localhost:9000 in your browser

### Building from Source

```bash
# Build the Go binary
make build

# The binary will be at ./bin/gonote
./bin/gonote --config go/config.yaml
```

### Frontend Development

For UI changes:

1. **Edit files in `shared/frontend/`** - they are served directly
2. **If modifying Tailwind CSS classes**, rebuild:
   ```bash
   npm run css-build
   ```
   
3. **For CSS development**, run the watcher in another terminal:
   ```bash
   npm run css-watch
   ```

### Running Tests

GoNote has two test suites:

**E2E Tests (Playwright):**
```bash
# Install browsers (one-time)
npx playwright install

# Run all E2E tests
make test-e2e

# Run with UI
npx playwright test --ui

# Run specific test
npx playwright test tests/e2e/notes/create-note.spec.ts
```

**Go Unit Tests:**
```bash
# Run all tests
make test

# Or manually (handles STORAGE_NOTES_DIR correctly)
STORAGE_NOTES_DIR=../data go test ./...

# With coverage
go test ./... -cover
```

---

## 📏 Code Standards

### Go Code

- Follow standard [Go conventions](https://go.dev/doc/effective_go)
- Use `gofmt` or `go fmt` before committing
- Write clear comments for exported functions and types
- Handle errors explicitly (no `_` discard unless intentional)
- Meaningful variable and function names
- Keep functions small and focused
- Use Go modules for dependency management

### Frontend (JavaScript/HTML/CSS)

- **Vanilla JS + Alpine.js** - no additional frameworks
- Semantic HTML
- Accessible (ARIA labels, keyboard navigation)
- Responsive design (mobile-first)
- Tailwind CSS for styling (when applicable)
- Keep JavaScript modular and event-driven

### Code Organization

- **Backend:** `go/internal/handlers/`, `go/internal/services/`
- **Frontend:** `shared/frontend/`
- **Tests:** `tests/e2e/` (E2E), `go/internal/**/*_test.go` (unit)
- **Config:** `go/config.yaml`
- **Static Assets:** `shared/frontend/libs/`

---

## 📝 Commit Guidelines

We use [Conventional Commits](https://www.conventionalcommits.org/) for consistent, readable commit history.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

| Type | When to Use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Code style changes (formatting, missing semicolons, etc.) |
| `refactor` | Code refactoring (no functional changes) |
| `perf` | Performance improvements |
| `test` | Adding or fixing tests |
| `chore` | Build, tool, or config changes |
| `ci` | CI/CD changes |
| `revert` | Reverting a previous commit |

### Examples

```bash
git commit -m "feat(search): add fuzzy matching support"

git commit -m "fix(auth): correct password validation logic"

git commit -m "docs(README): update deployment instructions"

git commit -m "test(e2e): add tests for theme switching"

git commit -m "chore(deps): update Fiber to v2.52"
```

**Scope:** Optional - indicates the part of codebase (e.g., `search`, `auth`, `handlers`, `frontend`)

**Subject:** Short description (50 chars or less), imperative mood ("add" not "added")

**Body:** Optional - more detailed explanation

**Footer:** Optional - reference issues (`Closes #123`, `Fixes #456`) or BREAKING CHANGE

---

## 🧪 Testing Requirements

### For All Contributions

- **Existing tests must pass** - Run `make test` and `make test-e2e` before submitting
- **New features should include tests** - E2E tests for user-facing features, unit tests for business logic
- **Bug fixes should include regression tests** - Prevent bug from coming back

### E2E Tests (Playwright)

Located in `tests/e2e/`:

- Organize by feature (e.g., `notes/`, `tags/`, `search/`)
- Use `[data-testid]` attributes for selectors (add them when needed)
- Follow the existing pattern (fixtures, page objects if applicable)
- Tests should be independent and idempotent
- Clean up test data (fixtures handle this globally)

Example structure:
```
tests/e2e/
├── auth/           # Authentication tests
├── notes/          # Note creation, editing, deletion
├── search/         # Search functionality
└── ...             # Other features
```

### Go Unit Tests

Located alongside source: `go/internal/**/*_test.go`

- Test handlers, services, and utility functions
- Use table-driven tests for multiple cases
- Mock dependencies where appropriate
- Test edge cases and error conditions

### Test Coverage

Maintain or improve coverage:
- E2E tests cover major user workflows
- Unit tests cover critical business logic

Run coverage:
```bash
cd go && go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 🔍 Review Process

All PRs undergo review:

1. **Automated checks**: CI runs tests, linting, and builds
2. **Maintainer assignment**: One or more maintainers review
3. **Review feedback**: Address comments or discuss alternatives
4. **Approval**: Once approved and CI passes, PR is merged
5. **Squash and merge**: We typically squash to keep history clean

**What reviewers look for:**

- Correctness - Does the code work as intended?
- Tests - Is it tested? Do tests make sense?
- Documentation - Are docs updated? Clear explanations?
- Security - Are there vulnerabilities? Secrets exposure?
- Performance - Any performance regressions?
- Style - Follows project conventions?
- Impact - Is the change scoped appropriately?

---

## 🌍 Community

- **Issues**: For bugs, features, and questions - [GitHub Issues](../../issues)
- **Discussions**: For broader conversations - [GitHub Discussions](../../discussions)
- **Translations**: See [Translating Documentation](#translating-documentation)

---

## 📄 License

By contributing, you agree that your contributions will be licensed under the project's [MIT License](../LICENSE).

---

## 🙏 Thank You!

Every contribution, whether it's a bug report, documentation fix, or code change, makes GoNote better. Thank you for taking the time to contribute!

---

*Need help? Have questions? Open an issue or discussion!*