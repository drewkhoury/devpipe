# JUnit XML Test Fixtures

This directory contains sample JUnit XML files for testing the metrics parser.

## Valid Formats

### `junit-single-suite.xml`
Classic JUnit format with a single `<testsuite>` root element.
- 5 tests: 2 passed, 1 failed, 1 error, 1 skipped
- Used by: JUnit (Java), Maven Surefire, Go (go-junit-report)

### `junit-multiple-suites.xml`
Modern format with `<testsuites>` root containing multiple `<testsuite>` elements.
- 8 tests across 2 suites: 5 passed, 2 failed, 1 skipped
- Used by: Gradle, Pytest, Jest, Mocha

### `junit-playwright-style.xml`
Playwright/modern test framework format with aggregated attributes at root.
- 3 tests across 2 suites: all passed
- Includes system-out with attachment references
- Used by: Playwright, Cypress, WebdriverIO

### `junit-no-xml-declaration.xml`
Valid XML without the `<?xml version="1.0"?>` declaration.
- 2 tests: all passed
- Tests parser tolerance for missing declarations

## Invalid Formats (for error handling tests)

### `junit-invalid-malformed.xml`
Malformed XML with unclosed tags.
- Should fail to parse with clear error message

### `junit-invalid-wrong-root.xml`
Valid XML but wrong root element (not testsuite/testsuites).
- Should fail to parse or return empty results

## Testing

Run the parser against these files:

```bash
# Valid files should parse successfully
./devpipe --config testdata/config-test-junit.toml

# Invalid files should show warnings but not crash
```

## Adding New Test Cases

When adding new test fixtures:
1. Name files descriptively: `junit-<format>-<variant>.xml`
2. Include comments explaining the format
3. Add entry to this README
4. Test both success and failure cases
