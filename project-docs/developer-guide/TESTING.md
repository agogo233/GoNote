# GoNote Tests

This directory contains all test suites for GoNote.

## Directory Structure

```
tests/
├── e2e/                    # Playwright E2E tests
│   ├── auth/               # Authentication tests
│   ├── bugs/               # Bug regression tests
│   ├── encoding-fix/       # Encoding tests
│   ├── export/             # Export functionality tests
│   ├── fixtures/           # Test fixtures and data
│   ├── folders/            # Folder navigation tests
│   ├── graph/              # Graph view tests
│   ├── i18n/               # Internationalization tests
│   ├── media/              # Media embedding tests
│   ├── mobile/             # Mobile responsiveness tests
│   ├── notes/              # Note CRUD tests
│   ├── outline/            # Outline/navigation tests
│   ├── search/             # Search functionality tests
│   ├── security/           # Security tests
│   ├── share/              # Sharing functionality tests
│   ├── shortcuts/          # Keyboard shortcut tests
│   ├── statistics/         # Statistics tests
│   ├── tags/               # Tag functionality tests
│   ├── templates/          # Template tests
│   ├── themes/             # Theme tests
│   ├── view-modes/         # View mode tests
│   └── homepage-fix.spec.ts # Homepage tests
└── README.md               # This file
```

## Test Types

### E2E Tests (Playwright)

Located in `e2e/`, these tests verify the complete application flow:

- **Browser automation** using Playwright
- **Full-stack testing** including frontend and backend
- **Real user scenarios** and workflows
- **Cross-browser testing** (Chromium, Firefox, WebKit)

### Unit Tests (Go)

Located alongside Go source code in `go/internal/**/*_test.go`:

- **Handler tests** - HTTP request/response
- **Service tests** - Business logic
- **Model tests** - Data structures
- **Utility tests** - Helper functions

## Running Tests

### E2E Tests

```bash
# Install Playwright browsers
npx playwright install

# Run all E2E tests
npx playwright test

# Run with UI
npx playwright test --ui

# Run specific test file
npx playwright test tests/e2e/notes/create-note.spec.ts

# Run tests by tag
npx playwright test --grep @smoke

# Run with specific browser
npx playwright test --project=chromium
```

### Go Unit Tests

```bash
# Run all Go tests
cd go && go test ./...

# Run with race detector
cd go && go test ./... -race

# Run with coverage
cd go && go test ./... -cover

# Run specific package tests
cd go && go test ./internal/handlers/...

# Run benchmarks
cd go && go test -bench=. ./...
```

## Test Configuration

### Playwright Config

See `playwright.config.ts` in the project root for:

- Browser configurations
- Test timeouts
- Reporter settings
- Parallel execution options

### Go Test Config

Go tests use standard Go testing package with no additional configuration.

## Writing Tests

### E2E Test Example

```typescript
import { test, expect } from '@playwright/test';

test('should create a new note', async ({ page }) => {
  await page.goto('http://localhost:9000');
  
  // Click new note button
  await page.click('[data-testid="new-note-btn"]');
  
  // Enter title
  await page.fill('#note-title', 'Test Note');
  
  // Verify note was created
  await expect(page.locator('.note-title')).toHaveText('Test Note');
});
```

### Go Test Example

```go
func TestNoteHandler_GetNote(t *testing.T) {
    handler := NewNoteHandler(service)
    
    req := httptest.NewRequest("GET", "/notes/test.md", nil)
    w := httptest.NewRecorder()
    
    handler.GetNote(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

## Test Coverage

### E2E Coverage

| Feature | Coverage |
|---------|----------|
| Authentication | ✅ |
| Note CRUD | ✅ |
| Search | ✅ |
| Tags | ✅ |
| Templates | ✅ |
| Graph View | ✅ |
| Sharing | ✅ |
| Themes | ✅ |
| Mobile | ✅ |
| Security | ✅ |

### Go Coverage

Run to see detailed coverage:

```bash
cd go && go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Continuous Integration

Tests run automatically on:

- Every push to main branch
- Every pull request
- Every release tag

See `.github/workflows/` for CI configuration.

## Fixtures

Test fixtures are stored in `e2e/fixtures/`:

- Sample notes
- Test images
- Configuration files
- Mock data

## Reporting

### Playwright Reports

```bash
# Generate HTML report
npx playwright test --reporter=html

# View report
npx playwright show-report
```

### Go Reports

```bash
# Generate coverage report
cd go && go test ./... -coverprofile=coverage.out

# View in browser
go tool cover -html=coverage.out
```

## Troubleshooting

### Flaky Tests

If a test is flaky:

1. Check for timing issues (add proper waits)
2. Ensure test isolation (clean state)
3. Check for external dependencies
4. Add retry logic if appropriate

### Debugging E2E Tests

```bash
# Run with slow motion
npx playwright test --debug

# Run specific test with logs
npx playwright test tests/e2e/notes/test.spec.ts --debug

# Use Playwright Inspector
PWDEBUG=1 npx playwright test
```

### Debugging Go Tests

```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./...

# Use delve debugger
dlv test ./internal/handlers
```

## Best Practices

1. **Test isolation** - Each test should be independent
2. **Descriptive names** - Test names should describe the behavior
3. **Arrange-Act-Assert** - Follow AAA pattern
4. **Clean up** - Remove test data after tests
5. **Meaningful assertions** - Test actual behavior, not implementation
6. **Page objects** - Use page object pattern for E2E tests
7. **Test data** - Use fixtures for consistent test data

## Related Documentation

- [Contributing Guidelines](../../CONTRIBUTING.md)
- [Security Tests](../../tests/e2e/security/)
- [API Documentation](./API.md)
