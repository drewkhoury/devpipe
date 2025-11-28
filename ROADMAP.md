# devpipe Development Roadmap

## Current Status: Iteration 1 ✅ COMPLETE

---

## 🔧 Refactoring Recommendations (Before Iteration 2)

### Quick Wins (30 min)
1. **Fix `multiWriter`** - Use `io.MultiWriter` from stdlib instead of custom implementation
2. **Fix `rand.Seed`** - Use `os.Getpid()` instead of deprecated `rand.Seed()`
3. **Fix git diff bug** - Handle empty output correctly (returns `[""]` instead of `[]`)

### Medium Effort (2-3 hours)
4. **Add context/cancellation** - Handle Ctrl+C gracefully with `context.Context`
5. **Add --version flag** - Show version info
6. **Validate --skip IDs** - Check that stage IDs exist before running

### Large Refactor (1-2 days)
7. **Restructure into packages**:
   ```
   cmd/devpipe/main.go           # CLI only
   internal/config/              # TOML loading (Iteration 2)
   internal/git/                 # Git operations
   internal/runner/              # Pipeline execution
   internal/model/               # Data structures
   internal/output/              # Console/logs/UI
   ```

**Recommendation:** Do quick wins now, defer large refactor until after Iteration 2 config is working.

---

## 📋 The 5 Iterations

### **Iteration 1 - Minimal Pipeline Runner** ✅ COMPLETE
- Hardcoded 6 stages
- CLI flags: --only, --skip, --fast, --fail-fast, --dry-run, --verbose
- Git integration (changed files from HEAD)
- run.json + per-stage logs
- Plain text output
- Comprehensive tests

**What you can do:** Use as a "make all checks" script with proper logging

---

### **Iteration 2 - Config + TOML + Git Modes** ✅ COMPLETE

**Goal:** Make it project-configurable

**Key Features:**
- TOML config file (`config.toml`)
- Define custom stages per project
- Git modes: `staged`, `staged_unstaged`, `ref`
- CLI: `--config <path>`, `--since <ref>`

**Example config.toml:**
```toml
[defaults]
outputRoot = ".devpipe"

[defaults.git]
mode = "staged_unstaged"

[stages.lint]
command = "npm run lint"
estimatedSeconds = 5

[stages.build]
command = "npm run build"
estimatedSeconds = 30
```

**Deliverables:**
- [x] TOML parser
- [x] Config loading with defaults
- [x] Git mode implementations
- [x] Example config.toml
- [x] Documentation
- [x] Package structure (internal/config, internal/git, internal/model)
- [x] All tests passing

**Timeline:** ✅ Completed in ~1 hour

**What you can do now:** Define your own stages per project with config.toml!

---

### **Iteration 3 - TUI + Colors + Progress** ✅ COMPLETE

**Goal:** Beautiful live UI

**Key Features:**
- TTY detection
- UI modes: `no-ui`, `minimal`, `full`
- Live progress bars (per-stage + overall)
- Colors (green=pass, red=fail, yellow=skip)
- CLI: `--ui=full`, `--no-color`

**Minimal UI:**
```
████████████████░░░░░░░░░░░░░░░░ 45% (3/6 stages)

✓ lint          1.2s
✓ format        1.1s
⚙ type-check    running... 2.3s
⋯ build         pending
```

**Full UI:**
```
╔══════════════════════════════════════════════════╗
║  devpipe run 2025-11-28T21-13-26Z_692593        ║
╚══════════════════════════════════════════════════╝

Overall: ████████████████░░░░░░░░░░░░░░░░ 45%

┌─ Quality ────────────────────────────────────┐
│ ✓ lint       1.2s / 5s   ████████████ 100%  │
│ ✓ format     1.1s / 5s   ████████████ 100%  │
└──────────────────────────────────────────────┘
```

**Deliverables:**
- [x] TTY detection and terminal width
- [x] Minimal UI implementation
- [x] Full UI implementation  
- [x] Color support with --no-color flag
- [x] Progress calculation logic
- [x] Status symbols (✓✗⊘⚙⋯)
- [x] UI package (internal/ui)
- [x] All tests passing

**Timeline:** ✅ Completed in ~30 minutes

**What you can do now:** Enjoy beautiful colored output with status symbols!

---

### **Iteration 4 - Metrics + summary.json + HTML Dashboard** 📊

**Goal:** Historical analytics

**Key Features:**
- Parse metrics: JUnit XML, ESLint JSON, SARIF
- Build artifact validation
- `summary.json` with aggregated stats
- HTML dashboard (`.devpipe/report.html`)
- Trends: pass rates, durations, test counts

**Metrics in run.json:**
```json
{
  "stages": [{
    "id": "unit-tests",
    "metrics": {
      "kind": "test",
      "summaryFormat": "junit",
      "data": {
        "tests": 42,
        "failures": 0,
        "time": 2.5
      }
    }
  }]
}
```

**HTML Dashboard:**
- Overview: total runs, pass rate, avg duration
- Recent runs list
- Per-stage stats and trends
- Charts for durations, test counts, lint issues

**Deliverables:**
- [ ] JUnit XML parser
- [ ] ESLint JSON parser
- [ ] summary.json generation
- [ ] HTML dashboard template
- [ ] JavaScript SPA with charts

**Timeline:** 5-7 days

---

### **Iteration 5 - Autofix + AI + Polish** 🤖

**Goal:** Make it feel like an assistant

**Key Features:**

**1. Autofix:**
```toml
[stages.lint]
command = "npm run lint"
autoFixEnabled = true
autoFixCommand = "npm run lint -- --fix"
```
- Run → fail → autofix → re-run
- Track success rate

**2. Build Contracts:**
```toml
[stages.build]
artifacts = ["dist/app.js", "dist/app.css"]
artifactMode = "fail"  # fail if missing
```

**3. AI Stubs (local, no API):**
- `ai-commit-msg`: Generate commit message from changed files
- `ai-pr-review`: Generate PR review suggestions

**4. Polish:**
- Better error messages
- Performance optimizations
- Enhanced HTML dashboard

**Deliverables:**
- [ ] Autofix execution logic
- [ ] Artifact contract validation
- [ ] AI commit message generator
- [ ] AI PR review generator
- [ ] Updated dashboard

**Timeline:** 5-7 days

---

## 📅 Total Timeline

| Iteration | Duration | Cumulative |
|-----------|----------|------------|
| 1 ✅ | 3-4 days | 3-4 days |
| 2 | 3-5 days | 6-9 days |
| 3 | 4-6 days | 10-15 days |
| 4 | 5-7 days | 15-22 days |
| 5 | 5-7 days | 20-29 days |

**Total: ~4-6 weeks of development**

---

## 🎯 Next Steps

### Option A: Start Iteration 2 Immediately
Jump into TOML config support without refactoring.

**Pros:** Fastest path to project-ready tool  
**Cons:** Technical debt accumulates

### Option B: Refactor First (Recommended)
Do quick wins + restructure into packages, then start Iteration 2.

**Pros:** Cleaner codebase, easier to maintain  
**Cons:** 1-2 extra days before new features

### Option C: Quick Wins Only
Fix the 3 small bugs, then start Iteration 2.

**Pros:** Balance of cleanup + progress  
**Cons:** Large refactor still needed later

---

## ❓ Decision Points

**Please confirm:**

1. **Which option?** A, B, or C?
2. **Iteration 2 scope:** Full TOML support or minimal config first?
3. **Testing approach:** Keep comprehensive tests or lighter coverage?
4. **Timeline:** Aggressive (4 weeks) or comfortable (6 weeks)?

Let me know your preferences and I'll create a detailed plan for the next iteration!
