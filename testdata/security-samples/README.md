# Security Test Samples

This directory contains **intentionally vulnerable code** used for testing security scanners.

## ⚠️ WARNING

**DO NOT use these code patterns in production!**

These files contain real security vulnerabilities including:
- Command injection
- Path traversal
- Weak cryptography
- And more...

## Purpose

These samples are used to:
1. Test that security scanners (gosec, CodeQL) correctly detect vulnerabilities
2. Verify SARIF parsing and dashboard display
3. Validate the security scanning pipeline

## Usage

### Enable Test Samples (for testing)
```bash
make security-test-enable
```

This copies `vulnerable.go.sample` to `internal/sectest/vulnerable.go` where scanners will find it.

### Run Security Scans (will find issues)
```bash
make security
# or
make codeql-analyze
./devpipe sarif -v tmp/codeql/results.sarif
```

### Disable Test Samples (for clean scans)
```bash
make security-test-disable
```

This removes the vulnerable code so your normal security scans are clean.

### Check Status
```bash
make security-test-status
```

## Files

- `vulnerable.go.sample` - Intentionally vulnerable Go code with:
  - Command injection (CWE-078)
  - Path traversal (CWE-022)
  - Weak crypto (MD5)
  - Data flow from HTTP requests to dangerous sinks

## How It Works

1. **Stored safely**: Files have `.sample` extension and are in `testdata/`
2. **Ignored by default**: Scanners skip `testdata/` and `.sample` files
3. **Toggle on demand**: Use Makefile commands to enable/disable
4. **Gitignored**: `internal/sectest/` is in `.gitignore` so enabled samples won't be committed

## Best Practices

- Always disable after testing: `make security-test-disable`
- Check status before committing: `make security-test-status`
- Never commit files from `internal/sectest/`
- Use only for testing scanner detection
