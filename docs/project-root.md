# Repo Root vs Git Root

devpipe uses two distinct concepts for directory resolution:

## Repo Root (repoRoot)

**Purpose:** Base directory for devpipe operations (the project root)

**Used for:**
- ✅ Resolving `outputRoot` (where `.devpipe/` is created)
- ✅ Resolving relative task `workdir` paths
- ✅ Resolving `watchPaths` patterns
- ✅ Safety validation (preventing execution in system directories)
- ✅ Display in UI and run metadata

**How it's determined:**

1. **If `repoRoot` is set in config:** Use that value (can be absolute or relative to config file)
2. **Otherwise, auto-detect:**
   - Resolve config file to absolute path
   - Get config directory: `filepath.Dir(configPath)`
   - Run `git rev-parse --show-toplevel` from config directory
   - If git repo found: use git root
   - If NO git repo: use config directory

## Git Root

**Purpose:** Actual git repository root for version control operations

**Used for:**
- ✅ Detecting changed files (`--since` flag)
- ✅ Running git commands
- ✅ Matching changed files against task `watchPaths`

**How it's determined:**
- Runs `git rev-parse --show-toplevel` from the **repo root** location
- This ensures git operations work correctly even when running devpipe from outside the project
- Falls back to repo root if not in a git repo

## When They Differ

### Scenario 1: Monorepo
```
/monorepo/                    ← Git root
├── .git/
├── services/
│   └── api/                  ← Project root (via config)
│       ├── config.toml
│       └── src/
```

```toml
[defaults]
repoRoot = "/monorepo/services/api"
```

### Scenario 2: Config Outside Project
```
/configs/
│   └── shared.toml           ← Config location
/projects/
    └── myapp/                ← Project root (via config)
        ├── .git/
        └── src/
```

```toml
[defaults]
repoRoot = "/projects/myapp"
```

### Scenario 3: No Git Repo
```
/opt/myapp/
├── config.toml               ← Config location
└── src/
```

- Git root: N/A (not in git repo)
- Project root: `/opt/myapp` (auto-detected from config location)

## Configuration

### Optional: Explicit repoRoot

```toml
[defaults]
repoRoot = "/absolute/path/to/project"  # Absolute path (recommended)
# OR
repoRoot = ".."                          # Relative to config file location
```

**When to set `repoRoot`:**
- Non-git projects
- Configs stored outside the project
- Monorepo scenarios with multiple projects
- When auto-detection doesn't match your needs

**When NOT to set `repoRoot`:**
- Standard single-repo projects (auto-detection works)
- Config is at or near project root
- You're using git

### Auto-Detection (Default)

If `repoRoot` is not set, devpipe automatically determines it:

```bash
# Example 1: Standard usage
cd /Users/drew/repos/myproject
devpipe
# → repoRoot = /Users/drew/repos/myproject (git root)

# Example 2: Absolute config path
cd /
devpipe --config /Users/drew/repos/myproject/config.toml
# → repoRoot = /Users/drew/repos/myproject (git root from config dir)

# Example 3: Config in subdirectory
cd /Users/drew/repos/myproject
devpipe --config configs/prod.toml
# → repoRoot = /Users/drew/repos/myproject (git root from configs/)

# Example 4: No git repo
cd /opt/myapp
devpipe
# → repoRoot = /opt/myapp (config directory, no git)
```

## Verbose Logging

Use `--verbose` to see how repoRoot and gitRoot are resolved:

```bash
devpipe --verbose
```

Output:
```
Config: config.toml
Repo root: /Users/drew/repos/myproject (auto-detected from git)
Git root: /Users/drew/repos/myproject (detected by running git from repo root)
Output directory: /Users/drew/repos/myproject/.devpipe
```

Or with explicit repoRoot:
```
Config: /configs/shared.toml (from --config)
Repo root: /opt/myapp (from config)
Git root: /opt/myapp (no git repo found at repo root)
Output directory: /opt/myapp/.devpipe
```

## Path Resolution Examples

### Example 1: Relative Paths (Auto-Detected)

```toml
[defaults]
outputRoot = ".devpipe"

[tasks.build]
workdir = "./src"
```

```bash
cd /Users/drew/repos/myproject
devpipe
```

**Resolution:**
- Project root: `/Users/drew/repos/myproject` (git root)
- Output: `/Users/drew/repos/myproject/.devpipe`
- Build workdir: `/Users/drew/repos/myproject/src`

### Example 2: Explicit repoRoot

```toml
[defaults]
repoRoot = "/opt/myapp"
outputRoot = ".devpipe"

[tasks.build]
workdir = "./src"
```

```bash
cd /
devpipe --config /configs/myapp.toml
```

**Resolution:**
- Project root: `/opt/myapp` (from config)
- Output: `/opt/myapp/.devpipe`
- Build workdir: `/opt/myapp/src`

### Example 3: Absolute Paths (Override)

```toml
[defaults]
outputRoot = "/var/log/devpipe"

[tasks.build]
workdir = "/absolute/path/to/src"
```

**Resolution:**
- Output: `/var/log/devpipe` (absolute, not relative to repoRoot)
- Build workdir: `/absolute/path/to/src` (absolute, not relative to repoRoot)

## Best Practices

1. **Let auto-detection work** - Don't set `repoRoot` unless you need to
2. **Use absolute paths** - If you do set `repoRoot`, use absolute paths for clarity
3. **Keep configs in the project** - Simplifies path resolution
4. **Use `--verbose`** - When debugging path issues
5. **Document your setup** - If using non-standard `repoRoot`, document why

## Troubleshooting

### Issue: Output directory created in wrong location

**Check:**
```bash
devpipe --verbose
```

Look for the "Project root" line to see where devpipe thinks the project is.

**Solutions:**
1. Set explicit `repoRoot` in config
2. Move config file to project root
3. Use absolute `outputRoot` path

### Issue: Tasks running in wrong directory

**Check:**
```bash
devpipe --verbose
```

Look for task workdir resolution.

**Solutions:**
1. Use absolute `workdir` in task config
2. Set correct `repoRoot`
3. Verify config file location

### Issue: WatchPaths not matching files

**Check:**
```bash
devpipe --verbose
```

Git root and project root should usually be the same for watchPaths to work correctly.

**Solutions:**
1. Ensure config is within git repo
2. Use absolute paths in `watchPaths`
3. Set `repoRoot` to match git root
