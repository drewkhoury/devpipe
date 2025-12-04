# SARIF Viewer

**Note: This standalone tool has been integrated into devpipe as the `devpipe sarif` subcommand.**

Use `./devpipe sarif` instead of building this separately.

## Usage

### View CodeQL results
```bash
make codeql-view
# or
./devpipe sarif tmp/codeql/results.sarif
```

### View a specific SARIF file
```bash
./devpipe sarif path/to/results.sarif
```

### View multiple SARIF files
```bash
./devpipe sarif file1.sarif file2.sarif file3.sarif
```

### Search a directory for SARIF files
```bash
./devpipe sarif -d tmp/
```

### Show summary grouped by rule
```bash
./devpipe sarif -s tmp/codeql/results.sarif
```

### Verbose output (show rule names)
```bash
./devpipe sarif -v tmp/codeql/results.sarif
```

## Output Format

### Default output (when issues are found)
```
⚠️  Found 2 security issue(s):

⚠️  Rule:    go/path-injection
   File:    internal/foo/bar.go:42
   Message: Untrusted input flows into a file path

❌ Rule:    go/sql-injection
   File:    pkg/db/db.go:88
   Message: User input used to build SQL query
```

### Summary output
```
📊 Security Issues Summary (2 total):

    1  go/sql-injection
    1  go/path-injection
```

## Exit Codes

- `0`: No issues found
- `1`: Issues found or error occurred

## Integration with CI/CD

The tool exits with code 1 if any security issues are found, making it suitable for CI/CD pipelines:

```bash
./devpipe sarif results.sarif || exit 1
```
