#!/usr/bin/env bash
set -euo pipefail

CMD="${1:-banner}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ART_DIR="${ROOT_DIR}/artifacts"

mkdir -p "${ART_DIR}"

# Support simulated failures via DEVPIPE_TEST_FAIL env var
# Usage: DEVPIPE_TEST_FAIL=lint ./hello-world.sh lint
if [[ "${DEVPIPE_TEST_FAIL:-}" == "$CMD" ]]; then
  echo "[hello-world] ❌ Simulating failure for $CMD"
  exit 1
fi

case "$CMD" in
  lint)
    echo "[hello-world] Linting sources..."
    sleep 0.25
    echo "[hello-world] Linting is happening..."
    sleep 0.25
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    sleep 0.25
    echo "[hello-world] Linting is happening..."
    sleep 0.25
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."
    echo "[hello-world] Linting is happening (lots of lines)..."       
    sleep 0.25
    echo "[hello-world] Linting is happening (1 line)..."       
    sleep 0.25
    echo "[hello-world] Lint OK"
    ;;

  format)
    echo "[hello-world] Checking formatting..."
    sleep 0.25
    echo "[hello-world] check..."
    sleep 0.25
    echo "[hello-world] check..."
    exit 1
    echo "[hello-world] Format OK"
    ;;

  type-check)
    echo "[hello-world] Type checking..."
    sleep 1
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    echo "[hello-world] Type echo..."
    sleep 0.25
    
    echo "[hello-world] Types OK"
    ;;

  contract-tests)
    echo "[hello-world] Contract testing..."
    sleep 1
    echo "[hello-world] Contracts OK"
    ;;

  build)
    echo "[hello-world] Building app..."
    mkdir -p "${ART_DIR}/build"
    echo "hello world app binary" > "${ART_DIR}/build/app.txt"
    sleep 0.5
    echo "[hello-world] Build done, artifact at artifacts/build/app.txt"
    ;;

  unit-tests)
    echo "[hello-world] Running unit tests..."
    mkdir -p "${ART_DIR}/test"
    JUNIT_FILE="${ART_DIR}/test/junit.xml"
    cat > "${JUNIT_FILE}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="hello-world-unit" tests="2" failures="0" errors="0" skipped="0" time="0.01">
  <testcase classname="hello.WorldTest" name="testOne" time="0.005"/>
  <testcase classname="hello.WorldTest" name="testTwo" time="0.005"/>
</testsuite>
EOF
    sleep 1
    echo "[hello-world] Unit tests OK, junit at artifacts/test/junit.xml"
    ;;

  sast-tests)
    echo "[hello-world] SAST testing..."
    sleep 1
    echo "[hello-world] SAST OK"
    ;;

  smoke-tests)
    echo "[hello-world] Running smoke tests (simulated long run)..."
    sleep 1
    echo "[hello-world] Smoke tests OK"
    ;;

  e2e-tests)
    echo "[hello-world] Running e2e tests (simulated long run)..."
    sleep 1
    echo "[hello-world] E2E tests OK"
    ;;

  demo-complete)
    cat << 'EOF'

┌───────────────────────────────────────────────────────────────────┐
│                                                                   │
│  ✨ Demo Complete!                                                │
│                                                                   │
│  You just ran:                                                    │
│    ./devpipe --config config/hello-world.toml                     │
│                                                                   │
│  📁 Explore the Configs:                                          │
│                                                                   │
│  • config/hello-world.toml  - Simple demo (what you just ran)     │
│    Basic task definitions, perfect for learning                   │
│                                                                   │
│  • config.toml              - Real-world setup (devpipe itself)   │
│    Phases, metrics, git integration, auto-fix, and more           │
│                                                                   │
│  • config.example.toml      - Template with all features          │
│    Copy this to start your own project                            │
│                                                                   │
│  🎮 Try These Cool Commands:                                      │
│                                                                   │
│    devpipe --ui full --dashboard     # Live dashboard view        │
│    devpipe --skip lint --skip format # Skip specific tasks        │
│    devpipe --only build              # Run just one task          │
│    devpipe --fast                    # Skip slow tasks            │
│    devpipe --fail-fast               # Stop on first failure      │
│    devpipe --dry-run                 # Preview without running    │
│                                                                   │
│  📊 View Your Results:                                            │
│    open .devpipe/report.html         # Dashboard with run history │
│                                                                   │
│  🚀 Start your devpipe journey today!                             │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘

EOF
    ;;

  banner)
    cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║   ██╗  ██╗███████╗██╗     ██╗      ██████╗                        ║
║   ██║  ██║██╔════╝██║     ██║     ██╔═══██╗                       ║
║   ███████║█████╗  ██║     ██║     ██║   ██║                       ║
║   ██╔══██║██╔══╝  ██║     ██║     ██║   ██║                       ║
║   ██║  ██║███████╗███████╗███████╗╚██████╔╝                       ║
║   ╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝ ╚═════╝                        ║
║                                                                   ║
║   ██╗    ██╗ ██████╗ ██████╗ ██╗     ██████╗                      ║
║   ██║    ██║██╔═══██╗██╔══██╗██║     ██╔══██╗                     ║
║   ██║ █╗ ██║██║   ██║██████╔╝██║     ██║  ██║                     ║
║   ██║███╗██║██║   ██║██╔══██╗██║     ██║  ██║                     ║
║   ╚███╔███╔╝╚██████╔╝██║  ██║███████╗██████╔╝                     ║
║    ╚══╝╚══╝  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═════╝                      ║
║                                                                   ║
║              🎭 Mock Application for devpipe                      ║
║                                                                   ║
║   This is a fake application used to demonstrate devpipe.         ║
║   It simulates common CI/CD tasks.                                ║
║                                                                   ║
║   Available commands:                                             ║
║     • lint        - Simulated linting                             ║
║     • format      - Simulated formatting                          ║
║     • type-check  - Simulated type checking                       ║
║     • build       - Simulated build                               ║
║     • unit-tests  - Simulated tests                               ║
║     • e2e-tests   - Simulated E2E tests                           ║
║     • banner      - Show this banner                              ║
║                                                                   ║
║   Usage: ./hello-world.sh <command>                               ║
║                                                                   ║
║   Try it with devpipe:                                            ║
║     devpipe --config config/hello-world.toml                      ║
║     make hello-demo                                               ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
EOF
    ;;

  *)
    echo "Unknown command: $CMD"
    echo "Run './hello-world.sh banner' for help"
    exit 1
    ;;
esac
