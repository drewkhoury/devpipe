.PHONY: help build run test clean destroy demo show-runs show-latest validate validate-all test-failures test-fail-fast test-continue-on-fail test-artifacts install-deps check-fmt fmt lint

help:
	@echo "devpipe - Makefile commands"
	@echo ""
	@echo "Get Started:"
	@echo "  make install-deps          - Install development dependencies - ie Golang (requires Homebrew)"
	@echo "  make demo                  - Run devpipe on hello-world example (start here if you're new to devpipe)"
	@echo ""
	@echo "Build & Run devpipe:"
	@echo "  make build                 - Build the devpipe binary"
	@echo "  make run                   - Build and run devpipe (uses config.toml)"
	@echo "  make test                  - Run Go tests"
	@echo "  make clean                 - Remove build artifacts"
	@echo "  make destroy               - Remove build artifacts AND .devpipe directory"
	@echo ""
	@echo "Failure testing (for internal validation of devpipe):"
	@echo "  make test-failures         - Run all failure tests"
	@echo "  make test-fail-fast        - Test --fail-fast stops on first failure"
	@echo "  make test-continue-on-fail - Test pipeline continues without --fail-fast"
	@echo "  make test-artifacts        - Test artifact/metrics validation"
	@echo ""
	@echo "Utilities:"
	@echo "  make show-runs             - List all pipeline runs"
	@echo "  make show-latest           - Show latest run.json"
	@echo "  make validate              - Validate default config.toml"
	@echo "  make validate-all          - Validate all config files in config/"

install-deps:
	@echo "Installing development dependencies..."
	@if ! command -v brew >/dev/null 2>&1; then \
		echo "❌ Error: Homebrew is not installed"; \
		echo "Install from: https://brew.sh"; \
		exit 1; \
	fi
	@echo "Running: brew bundle"
	@brew bundle
	@echo "✓ Dependencies installed"
	@echo ""
	@echo "Installed tools:"
	@command -v go >/dev/null 2>&1 && echo "  ✓ go $$(go version | awk '{print $$3}')" || echo "  ✗ go"
	@command -v golangci-lint >/dev/null 2>&1 && echo "  ✓ golangci-lint $$(golangci-lint --version | head -1 | awk '{print $$4}')" || echo "  ✗ golangci-lint"
	@command -v gosec >/dev/null 2>&1 && echo "  ✓ gosec $$(gosec -version 2>&1 | head -1)" || echo "  ✗ gosec"
	@command -v goreleaser >/dev/null 2>&1 && echo "  ✓ goreleaser $$(goreleaser --version | grep GitVersion | awk '{print $$2}')" || echo "  ✗ goreleaser"

build:
	@echo "Building devpipe..."
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "none"); \
	DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	go build -ldflags "\
		-X main.version=$$VERSION \
		-X main.commit=$$COMMIT \
		-X main.buildDate=$$DATE \
	" -o devpipe .
	@echo "✓ Built: ./devpipe"

run: build
	./devpipe

test:
	@echo "Running tests..."
	go test ./... -v

# Check commands (for devpipe config.toml)
check-fmt:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

fmt:
	@echo "Formatting Go files..."
	@gofmt -w .
	@echo "✓ All files formatted"

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed, skipping advanced linting"; \
		echo "Install: https://golangci-lint.run/usage/install/"; \
	fi

clean:
	@echo "Cleaning build artifacts..."
	rm -rf artifacts devpipe
	@echo "✓ Cleaned (kept .devpipe run history)"

destroy: clean
	@echo "⚠️  WARNING: This will delete ALL run history in .devpipe/"
	@printf "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@echo "Removing .devpipe directory (all run history)..."
	@rm -rf .devpipe
	@echo "✓ Destroyed everything"

# Demo commands
demo: build
	@echo "Running devpipe on hello-world app..."
	@chmod +x hello-world.sh
	./hello-world.sh banner
	./devpipe --config config/hello-world.toml || true
	@./hello-world.sh demo-complete

# Utility commands
show-runs:
	@echo "Pipeline runs:"
	@ls -lt .devpipe/runs/ 2>/dev/null || echo "No runs yet"

show-latest:
	@echo "Latest run.json:"
	@find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1 | xargs cat | jq . 2>/dev/null || \
	find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1 | xargs cat

validate: build
	@echo "Validating default config.toml..."
	./devpipe validate

validate-all: build
	@echo "Validating all config files..."
	./devpipe validate config/*.toml

# Failure testing
test-failures: test-fail-fast test-continue-on-fail
	@echo ""
	@echo "✅ All failure tests passed!"

test-fail-fast: build
	@echo "=========================================="
	@echo "TEST: --fail-fast should stop on first failure"
	@echo "=========================================="
	@echo "Running phase-testing config with --fail-fast..."
	@echo ""
	@./devpipe --config config/phase-testing.toml --fail-fast; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -ne 1 ]; then \
		echo ""; \
		echo "❌ FAIL: Expected exit code 1, got $$EXIT_CODE"; \
		exit 1; \
	fi
	@echo ""
	@echo "Checking that phase-should-not-run tasks did NOT run..."
	@LATEST_RUN=$$(find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1); \
	if [ -f "$$LATEST_RUN" ]; then \
		TASKS_RUN=$$(cat "$$LATEST_RUN" | grep -o '"id"' | wc -l | tr -d ' '); \
		if [ $$TASKS_RUN -gt 4 ]; then \
			echo "❌ FAIL: Expected only 4 tasks (phase-pass + phase-fail), but $$TASKS_RUN ran"; \
			cat "$$LATEST_RUN" | grep '"id"'; \
			exit 1; \
		fi; \
		echo "✅ PASS: Only $$TASKS_RUN tasks ran (phase-pass passed, phase-fail failed, phase-should-not-run skipped)"; \
	else \
		echo "❌ FAIL: No run.json found"; \
		exit 1; \
	fi
	@echo ""

test-continue-on-fail: build
	@echo "=========================================="
	@echo "TEST: Without --fail-fast, pipeline should continue"
	@echo "=========================================="
	@echo "Running phase-testing config WITHOUT --fail-fast..."
	@echo ""
	@./devpipe --config config/phase-testing.toml; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -ne 1 ]; then \
		echo ""; \
		echo "❌ FAIL: Expected exit code 1, got $$EXIT_CODE"; \
		exit 1; \
	fi
	@echo ""
	@echo "Checking that all tasks ran despite failure..."
	@LATEST_RUN=$$(find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1); \
	if [ -f "$$LATEST_RUN" ]; then \
		TASKS_RUN=$$(cat "$$LATEST_RUN" | grep -o '"id"' | wc -l | tr -d ' '); \
		if [ $$TASKS_RUN -ne 6 ]; then \
			echo "❌ FAIL: Expected all 6 tasks to run, but only $$TASKS_RUN ran"; \
			cat "$$LATEST_RUN" | grep '"id"'; \
			exit 1; \
		fi; \
		echo "✅ PASS: All $$TASKS_RUN tasks ran (phase-fail failed, but all phases completed)"; \
	else \
		echo "❌ FAIL: No run.json found"; \
		exit 1; \
	fi
	@echo ""

test-artifacts: build
	@echo "=========================================="
	@echo "TEST: Artifact and metrics validation"
	@echo "=========================================="
	@echo "Running artifact validation tests (expects some failures)..."
	@echo ""
	@./devpipe --config config/artifact-testing.toml; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -ne 1 ]; then \
		echo ""; \
		echo "❌ FAIL: Expected exit code 1 (some tasks should fail), got $$EXIT_CODE"; \
		exit 1; \
	fi
	@echo ""
	@echo "Checking results..."
	@LATEST_RUN=$$(find .devpipe/runs -name "run.json" -type f -print0 | xargs -0 ls -t | head -1); \
	if [ -f "$$LATEST_RUN" ]; then \
		PASSED=$$(cat "$$LATEST_RUN" | grep -c '"status": "PASS"' || true); \
		FAILED=$$(cat "$$LATEST_RUN" | grep -c '"status": "FAIL"' || true); \
		echo "✅ Results: $$PASSED passed, $$FAILED failed"; \
		if [ $$PASSED -lt 3 ] || [ $$FAILED -lt 5 ]; then \
			echo "❌ FAIL: Expected at least 3 passes and 5 failures"; \
			exit 1; \
		fi; \
	else \
		echo "❌ FAIL: No run.json found"; \
		exit 1; \
	fi
	@echo "✅ PASS: Artifact validation tests completed successfully"
	@echo ""
