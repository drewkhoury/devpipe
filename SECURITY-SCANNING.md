# Security Scanning with devpipe

This document describes the security scanning capabilities integrated into devpipe.

## Overview

devpipe includes two complementary security scanning tools:

1. **gosec** - Fast security scanner for Go code that catches common vulnerabilities
2. **CodeQL** - Deep semantic analysis that detects complex security issues and data flow vulnerabilities

## Quick Start

### One-time Setup

```bash
# Install all dependencies (including CodeQL)
make install-deps

# Download CodeQL query packs
make codeql-setup
```

### Running Security Scans

```bash
# Run both gosec and CodeQL
make security

# Run only gosec
make gosec

# Run only CodeQL
make codeql-analyze

# View CodeQL results in readable format
make codeql-view
```

## CodeQL Commands

### Create Database
```bash
make codeql-db
```
Creates a CodeQL database by building your Go code and extracting semantic information.

### Analyze Code
```bash
make codeql-analyze
```
Runs security queries against the database and generates:
- `tmp/codeql/results.sarif` - SARIF format (for tools/IDEs)
- `tmp/codeql/results.csv` - CSV format (for spreadsheets)

**Important**: The command will **exit with error code 1** if security issues are found, causing the task to fail in CI/CD pipelines.

### View Results
```bash
make codeql-view
# or
./devpipe sarif tmp/codeql/results.sarif
```
Displays results in a human-readable format:
```
⚠️  Rule:    go/path-injection
   File:    internal/foo/bar.go:42
   Message: Untrusted input flows into a file path

❌ Rule:    go/sql-injection
   File:    pkg/db/db.go:88
   Message: User input used to build SQL query
```
When no issues are found:
```
✅ No security issues found
```

### Clean Up
```bash
make codeql-clean
```
Removes CodeQL database and results (saves disk space).

## SARIF Viewer

The SARIF viewer is built into devpipe for viewing CodeQL results:

```bash
# View specific file (default format)
./devpipe sarif tmp/codeql/results.sarif

# Verbose output (show detailed metadata)
./devpipe sarif -v tmp/codeql/results.sarif

# Show summary grouped by rule
./devpipe sarif -s tmp/codeql/results.sarif

# View multiple files
./devpipe sarif file1.sarif file2.sarif

# Search directory for SARIF files
./devpipe sarif -d tmp/

# View help
./devpipe sarif -h
```

### Output Formats

**Default format:**
```
⚠️  Rule:    go/command-injection
   File:    internal/sectest/vulnerable.go:18:34
   Message: This command depends on a user-provided value.
```

**Verbose format (`-v`):**
```
⚠️  Rule:    go/command-injection
   File:    internal/sectest/vulnerable.go:18:34
   Message: This command depends on a user-provided value.
   Level:   warning
   Info:    Command built from user-controlled sources
   Precision: high
   Severity:  9.8
   Tags:    security, external/cwe/cwe-078
   Details: Building a system command from user-controlled sources is vulnerable...
   Source:  internal/sectest/vulnerable.go:17:15 - user-provided value
   Data Flow Path (4 steps):
     1. internal/sectest/vulnerable.go:17:15 - selection of URL
     2. internal/sectest/vulnerable.go:17:15 - call to Query
     3. internal/sectest/vulnerable.go:17:15 - call to Get
     4. internal/sectest/vulnerable.go:18:34 - userInput
```

**Summary format (`-s`):**
```
📊 Security Issues Summary (2 total):

    1  go/command-injection
    1  go/path-injection
```

## Integration with devpipe

Security scanning is integrated into the validation phase of `config.toml`:

```toml
[tasks.gosec]
name = "Security Scan (gosec)"
desc = "Scans Go code for security vulnerabilities"
command = "gosec ./..."
type = "check-security"

[tasks.codeql]
name = "Security Scan (CodeQL)"
desc = "Deep security analysis with CodeQL"
command = "make codeql-analyze"
type = "check-security"
```

Run with devpipe:
```bash
./devpipe
```

## What CodeQL Detects

CodeQL runs 32 security queries covering:

### Injection Vulnerabilities
- SQL injection (CWE-089)
- Command injection (CWE-078)
- Path traversal (CWE-022)
- XPath injection (CWE-643)
- Email injection (CWE-640)
- XSS (CWE-079)

### Cryptographic Issues
- Insecure TLS (CWE-327)
- Insufficient key size (CWE-326)
- Insecure randomness (CWE-338)
- Disabled certificate checks (CWE-295)

### Authentication & Authorization
- Missing JWT signature check (CWE-347)
- Constant OAuth2 state (CWE-352)
- Bad redirect check (CWE-601)

### Information Disclosure
- Cleartext logging (CWE-312)
- Stack trace exposure (CWE-209)

### Resource Management
- Allocation size overflow (CWE-190)
- Uncontrolled allocation size (CWE-770)

### Input Validation
- Incomplete hostname regexp (CWE-020)
- Missing regexp anchor (CWE-020)
- Suspicious characters in regexp (CWE-020)

### Other
- Request forgery (CWE-918)
- Incorrect integer conversion (CWE-681)
- Zip slip (CWE-022)

## CI/CD Integration

### Exit Codes
- `make security` exits with code 1 if any issues are found
- `make codeql-view` exits with code 1 if any issues are found

### Example GitHub Actions
```yaml
- name: Security Scan
  run: |
    make codeql-setup
    make security
    make codeql-view
```

### Example GitLab CI
```yaml
security:
  script:
    - make codeql-setup
    - make security
    - make codeql-view
  artifacts:
    reports:
      sast: tmp/codeql/results.sarif
```

## Performance

- **gosec**: Fast (~1-2 seconds for most projects)
- **CodeQL database creation**: Moderate (~10-30 seconds, depends on project size)
- **CodeQL analysis**: Fast (~3-5 seconds, cached after first run)

## Testing Security Scanner Detection

For testing that security scanners correctly detect vulnerabilities, devpipe includes intentionally vulnerable code samples that can be toggled on/off.

### Enable Test Samples
```bash
make security-test-enable
```

This copies vulnerable code to `internal/sectest/` where scanners will find it.

### Run Scans (will find issues)
```bash
make security
# or
make codeql-analyze
./devpipe sarif -v tmp/codeql/results.sarif
```

You should see 2 security issues detected:
- `go/command-injection` - Command injection vulnerability
- `go/path-injection` - Path traversal vulnerability

### Disable Test Samples
```bash
make security-test-disable
```

This removes the vulnerable code so your normal scans are clean.

### Check Status
```bash
make security-test-status
```

Shows whether test samples are currently enabled or disabled.

**⚠️ Important**: Always disable test samples before committing code. The `internal/sectest/` directory is gitignored to prevent accidental commits.

## Troubleshooting

### CodeQL not found
```bash
brew install --cask codeql
make codeql-setup
```

### Database creation fails
```bash
# Clean and retry
make codeql-clean
make codeql-db
```

### No results shown
```bash
# Check if SARIF file exists and has content
ls -lh tmp/codeql/results.sarif
cat tmp/codeql/results.sarif | jq '.runs[].results | length'
```

## Files and Directories

```
tmp/codeql/
├── db-go/              # CodeQL database (can be large)
├── results.sarif       # SARIF format results
└── results.csv         # CSV format results

internal/sarif/         # SARIF parsing library (used by devpipe sarif command)
```

## References

- [CodeQL Documentation](https://codeql.github.com/docs/)
- [SARIF Specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html)
- [gosec Documentation](https://github.com/securego/gosec)
