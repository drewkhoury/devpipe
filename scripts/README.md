# devpipe Scripts

Utility scripts for testing and development.

## Exploratory Testing

### exploratory-test-realworld.sh

Interactive script that tests devpipe on a real-world Go project.

**Usage:**
```bash
# From repo root (uses ./devpipe)
./scripts/exploratory-test-realworld.sh

# With custom devpipe binary
./scripts/exploratory-test-realworld.sh /path/to/devpipe
```

**What it does:**
1. Clones golang-gin-realworld-example-app to /tmp
2. Creates a realistic devpipe config
3. Runs 6 interactive test scenarios:
   - No changes (all tasks skip)
   - Change articles/*.go (module-specific tasks)
   - Change *.md (docs-check only)
   - Change go.mod (dependency-sensitive tasks)
   - --ignore-watch-paths (force all tasks)
   - --since HEAD~2 (git diff filtering)

**Interactive:**
- Pauses between tests (press Enter to continue)
- Shows expected behavior for each test
- Leaves test environment for manual exploration

**Cleanup:**
```bash
rm -rf /tmp/golang-gin-realworld-example-app
```

## Creating Your Own Exploratory Tests

**Template:**
```bash
#!/bin/bash
set -e

DEVPIPE_BIN="${1:-./devpipe}"
TEST_DIR="/tmp/my-test-project"

# 1. Setup test environment
cd "$TEST_DIR"

# 2. Create config.toml with watchPaths

# 3. Run test scenarios
$DEVPIPE_BIN --config config.toml

# 4. Make changes and test filtering
echo "change" >> some-file.txt
$DEVPIPE_BIN --config config.toml

# 5. Test flags
$DEVPIPE_BIN --config config.toml --ignore-watch-paths
$DEVPIPE_BIN --config config.toml --since main
```

**Tips:**
- Use real projects to test realistic scenarios
- Test edge cases (empty patterns, weird globs, etc.)
- Try different git modes (staged, staged_unstaged, ref)
- Test with many files (performance)
- Test with nested directories (pattern matching)
