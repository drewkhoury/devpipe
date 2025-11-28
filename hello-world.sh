#!/usr/bin/env bash
set -euo pipefail

CMD="${1:-}"

if [[ -z "$CMD" ]]; then
  echo "Usage: $0 <lint|format|type-check|build|unit-tests|e2e-tests>"
  exit 1
fi

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
    sleep 1
    echo "[hello-world] Linting is happening..."
    sleep 1
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    echo "[hello-world] Linting is happening (6 lines)..."
    sleep 1
    echo "[hello-world] Linting is happening..."
    sleep 1
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
    sleep 1
    echo "[hello-world] Linting is happening (1 line)..."       
    sleep 1
    echo "[hello-world] Lint OK"
    ;;

  format)
    echo "[hello-world] Checking formatting..."
    sleep 3
    echo "[hello-world] Format OK"
    ;;

  type-check)
    echo "[hello-world] Type checking..."
    sleep 3
    echo "[hello-world] Types OK"
    ;;

  build)
    echo "[hello-world] Building app..."
    mkdir -p "${ART_DIR}/build"
    echo "hello world app binary" > "${ART_DIR}/build/app.txt"
    sleep 3
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
    sleep 3
    echo "[hello-world] Unit tests OK, junit at artifacts/test/junit.xml"
    ;;

  e2e-tests)
    echo "[hello-world] Running e2e tests (simulated long run)..."
    sleep 600
    echo "[hello-world] E2E tests OK"
    ;;

  *)
    echo "Unknown command: $CMD"
    exit 1
    ;;
esac
