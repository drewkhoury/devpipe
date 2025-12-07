# devpipe

A lightweight, local pipeline runner for development tasks.

---

You know you should run tests before committing, but you don't, because it's slow, scattered across multiple commands, and honestly just annoying. So you push to CI, wait 3-7 minutes, and find out you had a typo, build issue, or worse, a security vulnerability.

`devpipe` is like GitHub Actions for your laptop — a single command that orchestrates your entire build/test pipeline locally, with support for parallel execution, a dashboard that shows you exactly what's happening, and how long it takes.

`devpipe` looks for a config.toml you specify, and runs the commands you tell it to run. It doesn't make any assumptions about how you run your tests. Its power is in the standardization that the flexible configuration file (`config.toml`) provides. With one file you get metrics, history, dashboards, a CLI to your local pipeline, and JUnit/SARIF/artifact parsing that give you feedback without having to leave your local machine.

See [features.md](docs/features.md) for a complete overview of features and screenshots.

<img src="mascot/squirrel.png" alt="devpipe mascot - Flowmunk" width="300">

`config.toml` really does start off as simple as:

```toml
[tasks.your-task]
command = "<your command>"
```

## Quick Start

### Install - Homebrew (macOS/Linux)

```bash
brew install drewkhoury/tap/devpipe
```

> ![info](https://img.shields.io/badge/INFO-blue)  
> In a rush? Skip to [cli-reference](#cli-reference) for more details about runtime options once you have `devpipe` installed, or just run it from your project root with no arguments to auto-generate a config.toml with example tasks.

<details>
<summary>More Install Options</summary>

---

**Direct download:**

```bash
# Set PLATFORM to: darwin-arm64, darwin-amd64, linux-amd64, or linux-arm64
#  darwin-arm64 (macOS Apple Silicon)
#  darwin-amd64 (macOS Intel)
#  linux-amd64 (Linux x86_64)
#  linux-arm64 (Linux ARM64)
PLATFORM=darwin-arm64
curl -L https://github.com/drewkhoury/devpipe/releases/latest/download/devpipe_*_${PLATFORM}.tar.gz | tar xz
chmod +x devpipe
```

**Build from source:**

```bash
git clone https://github.com/drewkhoury/devpipe
cd devpipe
make build
```

</details>

## Configuration

**📖 Configuration Reference**
- [config.example.toml](config.example.toml) - Complete annotated example
- [config.schema.json](config.schema.json) - JSON Schema for IDE support
- [Configuration Reference](docs/configuration.md) - Full documentation (Markdown)

`devpipe` expects `config.toml` from the current directory by default. If no config file is specified and `config.toml` doesn't exist, devpipe will auto-generate one with example tasks.

Example `config.toml`:

```toml
[tasks.lint]
command = "npm run lint"

[tasks.test]
command = "npm test"
metricsFormat = "junit"
metricsPath = "test-results/junit.xml"

[tasks.build]
command = "npm run build"
```

### Order of Precedence

All configuration values in devpipe are resolved in this order:

- **CLI flags** (e.g., `--fix-type`, `--since`, `--ui`) *\<highest priority\>*
- **Task level** (e.g., `[tasks.go-fmt] fixType = "auto"`)
- **Task defaults** (e.g., `[task_defaults] fixType = "helper"`)
- **Built-in defaults** (e.g., `helper`) *\<lowest priority\>*

### Auto-Fix

`devpipe` can automatically fix issues when tasks fail. This is useful for formatting checks, linting, and other fixable issues.

#### Fix Types

- **`helper`** (default): Show a suggestion on how to fix the issue
- **`auto`**: Automatically run the fix command and re-check
- **`none`**: Don't show fix suggestions (useful to override global defaults)

<details>
<summary>More Info</summary>

#### Configuration Example

```toml
[task_defaults]
fixType = "helper"  # Default: show fix suggestions

[tasks.go-fmt]
command = "make check-fmt"
fixType = "auto"           # Override: auto-fix this task
fixCommand = "make fmt"    # Command to fix formatting

[tasks.go-vet]
command = "go vet ./..."
fixCommand = "make vet-fix"  # Uses fixType="helper" from task_defaults

[tasks.security-scan]
command = "make security"
fixType = "none"  # Don't suggest fixes for security issues
```

### CLI Override

Override fix behavior for all tasks:

```bash
./devpipe --fix-type auto    # Auto-fix all tasks with fixCommand
./devpipe --fix-type helper  # Show suggestions only
./devpipe --fix-type none    # Disable all fix suggestions
```

#### Example Output

**Helper mode:**
```
[go-fmt         ] ✗ FAIL (80ms)
[go-fmt         ] 💡 To fix run: make fmt
```

**Auto mode:**
```
[go-fmt         ] ✗ FAIL (77ms)

[go-fmt         ] 🔧 Auto-fixing: make fmt (31ms)
[go-fmt         ] ✅ Fix succeeded, re-checking...
[go-fmt         ] ✅ PASS (48ms)
Summary:
  ✓ go-fmt          PASS       0.16s (156ms) [auto-fixed]
```

The total time includes: initial check + fix + re-check.

</details>

## Modes

### UI Modes

**`basic` mode** provides simple text output with status symbols. It's the default UI mode.

```bash
./devpipe -ui basic
```

Simple text output with status symbols:
- ✓ PASS (green)
- ✗ FAIL (red)
- ⊘ SKIPPED (yellow)

**`full` mode** provides borders, and better visuals, and more detailed formatting.

```bash
./devpipe -ui full
```

### Dashboard & Full UI Modes

Dashboard mode provides a live progress view, with animated progress bars and detailed task information.

This can be used with any UI mode (basic or full).

```bash
./devpipe --dashboard -ui full
```

```
╔═════════════════════════════════════════════════════════╗
║ devpipe run 2025-11-29T05-25-25Z_071352                 ║
║ Repo: /Users/you/project                                ║
║ Git: staged_unstaged | Files: 6                         ║
╚═════════════════════════════════════════════════════════╝

Overall: ████████████████████████████████████████ 100%

┌─ Quality Checks ──────────────────────────────────────┐
│ ✓ lint         1s                         
│ ✓ format       1s                         
└───────────────────────────────────────────────────────┘

┌─ Build ───────────────────────────────────────────────┐
│ ✓ build        3s                         
└───────────────────────────────────────────────────────┘

┌─ Tests ───────────────────────────────────────────────┐
│ ✓ unit-tests   2s                         
│ ✓ e2e-tests    5s                          
└───────────────────────────────────────────────────────┘
```

## Git Modes & Smart Task Filtering

Control which files are in scope for changes and automatically skip tasks that don't need to run:

- **`staged`** - Only staged files (`git diff --cached`)
- **`staged_unstaged`** - Staged + unstaged (`git diff HEAD`)
- **`ref`** - Compare against ref (`git diff <ref>`)

### WatchPaths - Automatic Task Filtering

Tasks can declare which files they care about using `watchPaths`. devpipe will automatically skip tasks when their watched files haven't changed:

```toml
[tasks.frontend-test]
command = "npm test"
workdir = "frontend"
watchPaths = ["src/**/*.ts", "src/**/*.tsx", "package.json"]

[tasks.backend-test]
command = "go test ./..."
workdir = "backend"
watchPaths = ["**/*.go", "go.mod"]

[tasks.security-scan]
command = "trivy fs ."
# No watchPaths = always runs
```

```bash
# Only runs tasks with matching file changes
./devpipe

# Compare against main branch
./devpipe --since main

# Force all tasks to run, ignore watchPaths
./devpipe --ignore-watch-paths
```

### Environment Variables

Git information is available to all tasks via environment variables:

- `DEVPIPE_GIT_MODE` - Git mode (staged, staged_unstaged, ref)
- `DEVPIPE_GIT_REF` - Git ref being compared
- `DEVPIPE_CHANGED_FILES_COUNT` - Number of changed files
- `DEVPIPE_CHANGED_FILES` - Newline-separated list of changed files
- `DEVPIPE_CHANGED_FILES_JSON` - JSON array of changed files

```bash
#!/bin/bash
# Example: Smart test runner
if [[ "$DEVPIPE_CHANGED_FILES" == *"src/api/"* ]]; then
    npm run test:api
fi
```

## Metrics & Dashboard

devpipe can parse test results, SARIF security findings, and artifacts, and generate HTML dashboards with detailed contextual information:

```toml
[tasks.unit-tests]
command = "npm test"
metricsFormat = "junit"
metricsPath = "test-results/junit.xml"

[tasks.security-scan]
command = "make security-scan"
metricsFormat = "sarif"
metricsPath = "tmp/codeql/results.sarif"

[tasks.build]
command = "make build"
metricsFormat = "artifact"
metricsPath = "dist/app.js"
```

View the dashboard:
```bash
open .devpipe/report.html
```

### SARIF Security Scanning

devpipe has built-in support for SARIF (Static Analysis Results Interchange Format) used by security scanners like CodeQL and gosec.

**View SARIF results:**
```bash
./devpipe sarif tmp/codeql/results.sarif           # Default view
./devpipe sarif -v tmp/codeql/results.sarif        # Verbose with data flow
./devpipe sarif -s tmp/codeql/results.sarif        # Summary by rule
```

**In your pipeline:**
```toml
[tasks.security-scan]
name = "Security Scan (CodeQL)"
command = "make codeql-analyze"
type = "check-security"
metricsFormat = "sarif"
metricsPath = "tmp/codeql/results.sarif"
```

The dashboard will display:
- 🔒 Security findings with severity levels
- 📊 Issue counts (errors, warnings, notes)
- 🔍 Data flow visualization (source → sink)
- 🏷️ CWE tags and CVSS scores
- ✅ Task fails if security issues are found

## Output Structure

```
.devpipe/
├── report.html             # HTML dashboard
├── summary.json            # Aggregated metrics
└── runs/
    └── 2025-11-29T05-25-25Z_071352/
        ├── run.json        # Run metadata
        ├── pipeline.log    # Verbose output log
        └── logs/
            ├── lint.log
            ├── build.log
            └── unit-tests.log
```

## Where you can use Devpipe

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
./devpipe --config .devpipe-precommit.toml --fail-fast
```

### CI/CD

```yaml
# .github/workflows/ci.yml
- name: Run devpipe
  run: |
    curl -L https://github.com/drewkhoury/devpipe/releases/latest/download/devpipe_*_linux_amd64.tar.gz | tar xz
    ./devpipe --no-color
```

### Local Development

```bash
# Quick checks before commit
./devpipe --fast

# Full pipeline with dashboard
./devpipe --dashboard -ui full
```

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please open an issue or PR. See [CONTRIBUTING.md](docs-developer/CONTRIBUTING.md) for details.

---

Built by [Andrew Khoury](https://github.com/drewkhoury)
