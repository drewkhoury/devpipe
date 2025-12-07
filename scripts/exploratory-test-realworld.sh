#!/bin/bash
set -e

# Exploratory Testing Script for devpipe
# Tests devpipe on a real-world Go project (golang-gin-realworld-example-app)

DEVPIPE_BIN="${1:-./devpipe}"
TEST_DIR="/tmp/golang-gin-realworld-example-app"
REPO_URL="https://github.com/gothinkster/golang-gin-realworld-example-app.git"

echo "=========================================="
echo "devpipe Exploratory Testing"
echo "=========================================="
echo ""
echo "Using devpipe binary: $DEVPIPE_BIN"
echo "Test directory: $TEST_DIR"
echo ""

# Check if devpipe binary exists
if [ ! -f "$DEVPIPE_BIN" ]; then
    echo "ERROR: devpipe binary not found at $DEVPIPE_BIN"
    echo "Usage: $0 [path-to-devpipe-binary]"
    exit 1
fi

# Clone or update the test project
echo "Step 1: Setting up test project..."
if [ -d "$TEST_DIR" ]; then
    echo "  → Test directory exists, cleaning up..."
    rm -rf "$TEST_DIR"
fi

echo "  → Cloning $REPO_URL..."
git clone --quiet "$REPO_URL" "$TEST_DIR"
cd "$TEST_DIR"
echo "  ✓ Project cloned"
echo ""

# Create devpipe config
echo "Step 2: Creating devpipe config..."
cat > config.toml << 'EOF'
[defaults]
outputRoot = ".devpipe"

[defaults.git]
mode = "staged_unstaged"

[tasks.go-fmt]
name = "Go Format Check"
command = "gofmt -l ."
type = "check"
watchPaths = ["**/*.go"]

[tasks.go-vet]
name = "Go Vet"
command = "go vet ./..."
type = "check"
watchPaths = ["**/*.go", "go.mod"]

[tasks.go-test]
name = "Go Tests"
command = "go test ./..."
type = "test"
watchPaths = ["**/*.go", "go.mod"]

[tasks.articles-test]
name = "Articles Module Tests"
command = "go test ./articles/..."
type = "test"
workdir = "."
watchPaths = ["articles/**/*.go"]

[tasks.users-test]
name = "Users Module Tests"
command = "go test ./users/..."
type = "test"
workdir = "."
watchPaths = ["users/**/*.go"]

[tasks.common-test]
name = "Common Module Tests"
command = "go test ./common/..."
type = "test"
workdir = "."
watchPaths = ["common/**/*.go"]

[tasks.build]
name = "Build Binary"
command = "go build -o bin/app ."
type = "build"
watchPaths = ["**/*.go", "go.mod"]

[tasks.docs-check]
name = "Documentation Check"
command = "echo 'Checking docs...'"
type = "check"
watchPaths = ["*.md", "**/*.md"]
EOF
echo "  ✓ Config created"
echo ""

# Test 1: No changes
echo "=========================================="
echo "TEST 1: No changes (all tasks should skip)"
echo "=========================================="
$DEVPIPE_BIN --config config.toml
echo ""
read -p "Press Enter to continue to Test 2..."
echo ""

# Test 2: Change articles file
echo "=========================================="
echo "TEST 2: Change articles/*.go"
echo "Expected: articles-test + global tasks run"
echo "=========================================="
echo "// test comment" >> articles/models.go
$DEVPIPE_BIN --config config.toml
git checkout articles/models.go
echo ""
read -p "Press Enter to continue to Test 3..."
echo ""

# Test 3: Change markdown file
echo "=========================================="
echo "TEST 3: Change *.md file"
echo "Expected: Only docs-check runs"
echo "=========================================="
echo "test" >> readme.md
$DEVPIPE_BIN --config config.toml
git checkout readme.md
echo ""
read -p "Press Enter to continue to Test 4..."
echo ""

# Test 4: Change go.mod
echo "=========================================="
echo "TEST 4: Change go.mod"
echo "Expected: Tasks watching go.mod run"
echo "=========================================="
echo "// comment" >> go.mod
$DEVPIPE_BIN --config config.toml
git checkout go.mod
echo ""
read -p "Press Enter to continue to Test 5..."
echo ""

# Test 5: Force all tasks
echo "=========================================="
echo "TEST 5: --ignore-watch-paths flag"
echo "Expected: ALL tasks run"
echo "=========================================="
$DEVPIPE_BIN --config config.toml --ignore-watch-paths
echo ""
read -p "Press Enter to continue to Test 6..."
echo ""

# Test 6: Since flag
echo "=========================================="
echo "TEST 6: --since HEAD~2 flag"
echo "Expected: Tasks run based on git diff"
echo "=========================================="
echo "Recent commits:"
git log --oneline -5
echo ""
$DEVPIPE_BIN --config config.toml --since HEAD~2
echo ""

# Summary
echo "=========================================="
echo "Exploratory Testing Complete!"
echo "=========================================="
echo ""
echo "Test directory: $TEST_DIR"
echo "Config file: $TEST_DIR/config.toml"
echo ""
echo "You can now:"
echo "  cd $TEST_DIR"
echo "  # Make your own changes"
echo "  $DEVPIPE_BIN --config config.toml"
echo "  # Or try other flags:"
echo "  $DEVPIPE_BIN --config config.toml --verbose"
echo "  $DEVPIPE_BIN --config config.toml --only articles-test"
echo "  $DEVPIPE_BIN --config config.toml --dashboard"
echo ""
echo "To clean up:"
echo "  rm -rf $TEST_DIR"
echo ""
