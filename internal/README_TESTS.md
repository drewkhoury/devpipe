# Unit Tests

Comprehensive unit tests for core devpipe components.

## Test Coverage Summary

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/sarif` | **85.9%** | ✅ Excellent |
| `internal/metrics` | **79.7%** | ✅ Excellent |
| `internal/git` | **66.7%** | ✅ Good |
| `internal/config` | **60.2%** | ✅ Good |
| `internal/ui` | 20.0% | 🔴 Low |
| `main` | 3.1% | 🔴 Very low |

**4 out of 6 packages now have 60%+ coverage!** 🎉

## Test Files

- **`main_test.go`** - CLI utility functions (7 tests)
- **`internal/config/config_test.go`** - Config loading, validation, inheritance (17 tests)
- **`internal/git/git_test.go`** - Git repo detection, changed files (3 tests)
- **`internal/metrics/junit_test.go`** - JUnit XML parsing (3 tests)
- **`internal/metrics/sarif_test.go`** - SARIF metrics parsing (2 tests)
- **`internal/model/model_test.go`** - Data structures, status constants (3 tests)
- **`internal/sarif/sarif_test.go`** - SARIF document parsing (6 tests)
- **`internal/sarif/dataflow_test.go`** - Data flow, code flow testing (3 tests)
- **`internal/sarif/comprehensive_test.go`** - Comprehensive SARIF features (2 tests)
- **`internal/sarif/util_test.go`** - Print and utility functions (6 tests)
- **`internal/ui/colors_test.go`** - Colors, symbols, formatting (10 tests)

## Running Tests

```bash
# Run all tests
make test

# Run tests with JUnit XML output (using gotestsum)
make test-junit

# Run tests for a specific package
go test ./internal/config -v

# Run tests with coverage
go test ./internal/... -cover

# Run a specific test
go test ./internal/config -run TestLoadConfig -v
```

### JUnit XML Output

The project uses [gotestsum](https://github.com/gotestyourself/gotestsum) to generate JUnit XML reports from Go tests:

```bash
# Install gotestsum (included in Brewfile)
brew install gotestsum

# Run tests and generate JUnit XML
make test-junit
# Output: artifacts/junit.xml

# Use with devpipe to track test metrics
make run
```

The JUnit XML output can be consumed by devpipe's metrics system, allowing you to:
- Track test counts, failures, and skipped tests
- View test results in the dashboard
- Integrate with CI/CD pipelines
- Monitor test trends over time

**Note:** The main `config.toml` is already configured to use `make test-junit` for the unit-tests task, so running `make run` will automatically generate and consume JUnit XML reports.

## Test Philosophy

These tests focus on:
- ✅ **Core functionality** - Config loading, validation, inheritance
- ✅ **Public APIs** - Functions that are used throughout the codebase
- ✅ **Edge cases** - Invalid inputs, empty values, missing fields
- ✅ **Integration points** - Git detection, file system operations

Tests are intentionally simple and focus on behavior rather than implementation details.

## Future Improvements

Potential areas for expanded test coverage:
- Dashboard HTML generation
- SARIF/JUnit parsing
- Metrics collection
- Task execution logic
- CLI flag parsing
- Error handling edge cases
