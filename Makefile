.PHONY: help build run test test-junit clean destroy demo show-runs show-latest validate validate-all test-failures test-fail-fast test-continue-on-fail test-artifacts install-deps check-fmt fmt lint gosec codeql-setup codeql-db codeql-analyze codeql-clean codeql-view security security-test-enable security-test-disable security-test-status generate-docs check-docs setup-hooks

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
	@echo "  make test-junit            - Run tests with JUnit XML output (requires gotestsum)"
	@echo "  make clean                 - Remove build artifacts"
	@echo "  make destroy               - Remove build artifacts AND .devpipe directory"
	@echo ""
	@echo "Failure testing (for internal validation of devpipe):"
	@echo "  make test-failures         - Run all failure tests"
	@echo "  make test-fail-fast        - Test --fail-fast stops on first failure"
	@echo "  make test-continue-on-fail - Test pipeline continues without --fail-fast"
	@echo "  make test-artifacts        - Test artifact/metrics validation"
	@echo ""
	@echo "Security Scanning:"
	@echo "  make security              - Run all security scans (gosec + CodeQL)"
	@echo "  make gosec                 - Run gosec security scanner"
	@echo "  make codeql-setup          - Install CodeQL packs (run once)"
	@echo "  make codeql-db             - Create CodeQL database"
	@echo "  make codeql-analyze        - Analyze code with CodeQL (creates SARIF + CSV)"
	@echo "  make codeql-view           - View CodeQL results in readable format"
	@echo "  make codeql-clean          - Remove CodeQL database and results"
	@echo ""
	@echo "Security Testing (for testing scanner detection):"
	@echo "  make security-test-enable  - Enable vulnerable code samples for testing"
	@echo "  make security-test-disable - Disable vulnerable code samples"
	@echo "  make security-test-status  - Check if security test samples are enabled"
	@echo ""
	@echo "Utilities:"
	@echo "  make show-runs             - List all pipeline runs"
	@echo "  make show-latest           - Show latest run.json"
	@echo "  make validate              - Validate default config.toml"
	@echo "  make validate-all          - Validate all config files in config/"
	@echo ""
	@echo "Documentation:"
	@echo "  make generate-docs         - Generate docs from config structs"
	@echo "  make check-docs            - Check if docs are up to date"
	@echo "  make setup-hooks           - Install git hooks (auto-generates docs)"

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
	@command -v gotestsum >/dev/null 2>&1 && echo "  ✓ gotestsum $$(gotestsum --version)" || echo "  ✗ gotestsum"
	@command -v golangci-lint >/dev/null 2>&1 && echo "  ✓ golangci-lint $$(golangci-lint --version | head -1 | awk '{print $$4}')" || echo "  ✗ golangci-lint"
	@command -v gosec >/dev/null 2>&1 && echo "  ✓ gosec $$(gosec -version 2>&1 | head -1)" || echo "  ✗ gosec"
	@command -v codeql >/dev/null 2>&1 && echo "  ✓ codeql $$(codeql version --format=terse)" || echo "  ✗ codeql"
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

test-junit:
	@echo "Running tests with JUnit XML output..."
	@mkdir -p artifacts
	@if ! command -v gotestsum >/dev/null 2>&1; then \
		echo "❌ Error: gotestsum is not installed"; \
		echo "Install: brew install gotestsum"; \
		exit 1; \
	fi
	gotestsum \
		--junitfile artifacts/junit.xml \
		--format testname \
		-- \
		-coverprofile=artifacts/coverage.out \
		-covermode=atomic \
		./...
	@echo "✓ JUnit XML report: artifacts/junit.xml"
	@go tool cover -func=artifacts/coverage.out | grep total: | awk '{print "✓ Total coverage: " $$3}'

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

# ============================================================================
# Security Scanning
# ============================================================================

gosec:
	@echo "Running gosec security scanner..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "❌ Error: gosec is not installed"; \
		echo "Install: brew install gosec"; \
		exit 1; \
	fi
	@gosec ./...

# ============================================================================
# CodeQL Security Scanning
# ============================================================================

# CodeQL directories
CODEQL_DIR = tmp/codeql
CODEQL_DB = $(CODEQL_DIR)/db-go
CODEQL_RESULTS_SARIF = $(CODEQL_DIR)/results.sarif
CODEQL_RESULTS_CSV = $(CODEQL_DIR)/results.csv

codeql-setup:
	@echo "Setting up CodeQL..."
	@if ! command -v codeql >/dev/null 2>&1; then \
		echo "❌ Error: CodeQL is not installed"; \
		echo "Install: brew install --cask codeql"; \
		exit 1; \
	fi
	@echo "Downloading CodeQL packs..."
	@codeql pack download codeql/go-queries
	@codeql pack download codeql/go-all
	@echo "✓ CodeQL setup complete"

codeql-db:
	@echo "Creating CodeQL database..."
	@if ! command -v codeql >/dev/null 2>&1; then \
		echo "❌ Error: CodeQL is not installed. Run: make codeql-setup"; \
		exit 1; \
	fi
	@mkdir -p $(CODEQL_DIR)
	@rm -rf $(CODEQL_DB)
	@echo "Building Go code and extracting database..."
	@codeql database create $(CODEQL_DB) \
		--language=go \
		--source-root=. \
		--command='go build ./...' \
		--overwrite
	@echo "✓ CodeQL database created: $(CODEQL_DB)"

codeql-analyze: codeql-db
	@echo "Analyzing code with CodeQL..."
	@echo "Generating SARIF report..."
	@codeql database analyze $(CODEQL_DB) \
		codeql/go-queries \
		--format=sarifv2.1.0 \
		--output=$(CODEQL_RESULTS_SARIF)
	@echo "✓ SARIF report: $(CODEQL_RESULTS_SARIF)"
	@echo ""
	@echo "Generating CSV report..."
	@codeql database analyze $(CODEQL_DB) \
		codeql/go-queries \
		--format=csv \
		--output=$(CODEQL_RESULTS_CSV)
	@echo "✓ CSV report: $(CODEQL_RESULTS_CSV)"
	@echo ""
	@echo "✅ CodeQL analysis complete!"
	@echo "View results:"
	@echo "  - SARIF: $(CODEQL_RESULTS_SARIF)"
	@echo "  - CSV:   $(CODEQL_RESULTS_CSV)"
	@echo ""
	@echo "Checking for security issues..."
	@./devpipe sarif $(CODEQL_RESULTS_SARIF) > /dev/null 2>&1 || \
		(echo "❌ Security issues found! Run 'make codeql-view' or './devpipe sarif $(CODEQL_RESULTS_SARIF)' to see details" && exit 1)
	@echo "✅ No security issues found"

codeql-view: build
	@if [ ! -f $(CODEQL_RESULTS_SARIF) ]; then \
		echo "❌ No SARIF results found. Run: make codeql-analyze"; \
		exit 1; \
	fi
	@./devpipe sarif $(CODEQL_RESULTS_SARIF)

codeql-clean:
	@echo "Cleaning CodeQL artifacts..."
	@rm -rf $(CODEQL_DIR)
	@echo "✓ CodeQL artifacts removed"

# Run all security scans
security:
	@echo "=========================================="
	@echo "Running Security Scans"
	@echo "=========================================="
	@echo ""
	@echo "1. Running gosec..."
	@$(MAKE) gosec || true
	@echo ""
	@echo "2. Running CodeQL..."
	@$(MAKE) codeql-analyze
	@echo ""
	@echo "✅ All security scans complete!"

# ============================================================================
# Security Test Samples (for testing scanner detection)
# ============================================================================

SECURITY_SAMPLES_DIR = testdata/security-samples
SECURITY_TEST_DIR = internal/sectest

security-test-enable:
	@if [ -d "$(SECURITY_TEST_DIR)" ]; then \
		echo "⚠️  Security test samples are already enabled"; \
		echo "   Location: $(SECURITY_TEST_DIR)/"; \
	else \
		echo "Enabling security test samples..."; \
		if [ ! -f "$(SECURITY_SAMPLES_DIR)/vulnerable.go.sample" ]; then \
			echo "❌ Error: Sample file not found at $(SECURITY_SAMPLES_DIR)/vulnerable.go.sample"; \
			exit 1; \
		fi; \
		mkdir -p $(SECURITY_TEST_DIR); \
		cp $(SECURITY_SAMPLES_DIR)/vulnerable.go.sample $(SECURITY_TEST_DIR)/vulnerable.go; \
		echo "✅ Security test samples enabled at $(SECURITY_TEST_DIR)/"; \
		echo ""; \
		echo "⚠️  WARNING: This code contains intentional security vulnerabilities!"; \
		echo "   Scanners will now detect issues. Use for testing only."; \
		echo ""; \
		echo "To disable: make security-test-disable"; \
	fi

security-test-disable:
	@if [ ! -d "$(SECURITY_TEST_DIR)" ]; then \
		echo "✅ Security test samples are already disabled"; \
	else \
		echo "Disabling security test samples..."; \
		rm -rf $(SECURITY_TEST_DIR); \
		echo "✅ Security test samples disabled"; \
		echo "   Scanners will no longer detect test vulnerabilities"; \
	fi

security-test-status:
	@if [ -d "$(SECURITY_TEST_DIR)" ]; then \
		echo "🔴 Security test samples are ENABLED"; \
		echo "   Location: $(SECURITY_TEST_DIR)/"; \
		echo "   Files:"; \
		ls -la $(SECURITY_TEST_DIR)/ 2>/dev/null | tail -n +2 | sed 's/^/     /' || true; \
		echo ""; \
		echo "⚠️  Scanners will detect intentional vulnerabilities"; \
		echo "   To disable: make security-test-disable"; \
	else \
		echo "🟢 Security test samples are DISABLED"; \
		echo "   Scanners will run clean"; \
		echo ""; \
		echo "To enable for testing: make security-test-enable"; \
	fi

# ============================================================================
# Documentation Generation
# ============================================================================

generate-docs:
	@echo "Generating documentation from config structs and snippets..."
	@go run cmd/generate-docs/main.go
	@echo ""
	@echo "✅ Documentation generated:"
	@echo "   - config.example.toml"
	@echo "   - config.schema.json"
	@echo "   - docs/configuration.md"
	@echo "   - docs/cli-reference.md"
	@echo "   - docs/config-validation.md"

check-docs:
	@echo "Checking if documentation is up to date..."
	@# Check if files exist
	@if [ ! -f config.example.toml ] || [ ! -f config.schema.json ] || [ ! -f docs/configuration.md ] || [ ! -f docs/cli-reference.md ] || [ ! -f docs/config-validation.md ]; then \
		echo "❌ Generated documentation files are missing!"; \
		echo "Run 'make generate-docs' to create them."; \
		exit 1; \
	fi
	@# Save checksums of current files
	@BEFORE=$$(cat config.example.toml config.schema.json docs/configuration.md docs/cli-reference.md docs/config-validation.md 2>/dev/null | shasum -a 256 | cut -d' ' -f1); \
	go run cmd/generate-docs/main.go > /dev/null 2>&1; \
	AFTER=$$(cat config.example.toml config.schema.json docs/configuration.md docs/cli-reference.md docs/config-validation.md 2>/dev/null | shasum -a 256 | cut -d' ' -f1); \
	if [ "$$BEFORE" != "$$AFTER" ]; then \
		git checkout config.example.toml config.schema.json docs/configuration.md docs/cli-reference.md docs/config-validation.md 2>/dev/null || true; \
		echo "❌ Documentation is out of sync!"; \
		echo ""; \
		echo "The generated docs don't match committed files."; \
		echo "Run 'make generate-docs' and commit the changes."; \
		exit 1; \
	else \
		echo "✅ Documentation is up to date"; \
	fi

setup-hooks:
	@echo "Installing git hooks..."
	@if [ -f .git/hooks/pre-commit ]; then \
		echo "⚠️  .git/hooks/pre-commit already exists"; \
		echo "Backing up to .git/hooks/pre-commit.backup"; \
		mv .git/hooks/pre-commit .git/hooks/pre-commit.backup; \
	fi
	@cp hooks/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Git hooks installed"
	@echo ""
	@echo "The pre-commit hook will automatically:"
	@echo "  - Detect changes to internal/config/"
	@echo "  - Run make generate-docs"
	@echo "  - Add generated files to your commit"
