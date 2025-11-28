.PHONY: help build run test clean demo demo-verbose demo-fast demo-fail-fast demo-only demo-skip demo-dry-run test-failures test-fail-fast test-continue-on-fail

help:
	@echo "devpipe - Makefile commands"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build          - Build the devpipe binary"
	@echo "  make run            - Build and run with default settings"
	@echo "  make test           - Run Go tests (when we add them)"
	@echo "  make clean          - Remove build artifacts and .devpipe directory"
	@echo ""
	@echo "Demo commands (for testing):"
	@echo "  make demo           - Run basic pipeline"
	@echo "  make demo-verbose   - Run with verbose output"
	@echo "  make demo-fast      - Run with --fast (skip long stages)"
	@echo "  make demo-fail-fast - Run with --fail-fast"
	@echo "  make demo-only      - Run only unit-tests stage"
	@echo "  make demo-skip      - Run pipeline, skip lint and format"
	@echo "  make demo-dry-run   - Dry run (don't execute commands)"
	@echo ""
	@echo "Failure testing:"
	@echo "  make test-failures       - Run all failure tests"
	@echo "  make test-fail-fast      - Test --fail-fast stops on first failure"
	@echo "  make test-continue-on-fail - Test pipeline continues without --fail-fast"
	@echo ""
	@echo "Utilities:"
	@echo "  make show-runs      - List all pipeline runs"
	@echo "  make show-latest    - Show latest run.json"
	@echo "  make hello-test     - Test hello-world.sh directly"

build:
	@echo "Building devpipe..."
	go build -o devpipe .
	@echo "✓ Built: ./devpipe"

run: build
	./devpipe

test:
	@echo "Running tests..."
	go test ./... -v

clean:
	@echo "Cleaning up..."
	rm -rf .devpipe artifacts devpipe
	@echo "✓ Cleaned"

# Demo commands
demo: build
	@echo "Running basic pipeline..."
	./devpipe

demo-verbose: build
	@echo "Running pipeline with verbose output..."
	./devpipe --verbose

demo-fast: build
	@echo "Running pipeline with --fast (skips stages >= 300s)..."
	./devpipe --fast --verbose

demo-fail-fast: build
	@echo "Running pipeline with --fail-fast..."
	./devpipe --fail-fast --verbose

demo-only: build
	@echo "Running only unit-tests stage..."
	./devpipe --only unit-tests --verbose

demo-skip: build
	@echo "Running pipeline, skipping lint and format..."
	./devpipe --skip lint --skip format --verbose

demo-dry-run: build
	@echo "Dry run (no actual execution)..."
	./devpipe --dry-run --verbose

# Utility commands
show-runs:
	@echo "Pipeline runs:"
	@ls -lt .devpipe/runs/ 2>/dev/null || echo "No runs yet"

show-latest:
	@echo "Latest run.json:"
	@find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1 | xargs cat | jq . 2>/dev/null || \
	find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1 | xargs cat

hello-test:
	@echo "Testing hello-world.sh commands..."
	@chmod +x hello-world.sh
	@echo ""
	@echo "=== Testing lint ==="
	./hello-world.sh lint
	@echo ""
	@echo "=== Testing format ==="
	./hello-world.sh format
	@echo ""
	@echo "=== Testing type-check ==="
	./hello-world.sh type-check
	@echo ""
	@echo "=== Testing build ==="
	./hello-world.sh build
	@echo ""
	@echo "=== Testing unit-tests ==="
	./hello-world.sh unit-tests
	@echo ""
	@echo "✓ All hello-world.sh commands work!"

# Failure testing
test-failures: test-fail-fast test-continue-on-fail
	@echo ""
	@echo "✅ All failure tests passed!"

test-fail-fast: build
	@echo "=========================================="
	@echo "TEST: --fail-fast should stop on first failure"
	@echo "=========================================="
	@echo "Making format stage fail..."
	@echo ""
	@DEVPIPE_TEST_FAIL=format ./devpipe --fail-fast --verbose; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -ne 1 ]; then \
		echo ""; \
		echo "❌ FAIL: Expected exit code 1, got $$EXIT_CODE"; \
		exit 1; \
	fi
	@echo ""
	@echo "Checking that stages after format did NOT run..."
	@LATEST_RUN=$$(find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1); \
	if [ -f "$$LATEST_RUN" ]; then \
		STAGES_RUN=$$(cat "$$LATEST_RUN" | grep -o '"id"' | wc -l | tr -d ' '); \
		if [ $$STAGES_RUN -gt 2 ]; then \
			echo "❌ FAIL: Expected only 2 stages (lint, format), but $$STAGES_RUN ran"; \
			cat "$$LATEST_RUN" | grep '"id"'; \
			exit 1; \
		fi; \
		echo "✅ PASS: Only $$STAGES_RUN stages ran (lint passed, format failed, rest skipped)"; \
	else \
		echo "❌ FAIL: No run.json found"; \
		exit 1; \
	fi
	@echo ""

test-continue-on-fail: build
	@echo "=========================================="
	@echo "TEST: Without --fail-fast, pipeline should continue"
	@echo "=========================================="
	@echo "Making format stage fail (no --fail-fast)..."
	@echo ""
	@DEVPIPE_TEST_FAIL=format ./devpipe --verbose; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -ne 1 ]; then \
		echo ""; \
		echo "❌ FAIL: Expected exit code 1, got $$EXIT_CODE"; \
		exit 1; \
	fi
	@echo ""
	@echo "Checking that all stages ran despite failure..."
	@LATEST_RUN=$$(find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1); \
	if [ -f "$$LATEST_RUN" ]; then \
		STAGES_RUN=$$(cat "$$LATEST_RUN" | grep -o '"id"' | wc -l | tr -d ' '); \
		if [ $$STAGES_RUN -ne 6 ]; then \
			echo "❌ FAIL: Expected all 6 stages to run, but only $$STAGES_RUN ran"; \
			cat "$$LATEST_RUN" | grep '"id"'; \
			exit 1; \
		fi; \
		echo "✅ PASS: All $$STAGES_RUN stages ran (format failed, but pipeline continued)"; \
	else \
		echo "❌ FAIL: No run.json found"; \
		exit 1; \
	fi
	@echo ""
