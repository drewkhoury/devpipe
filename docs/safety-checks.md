# Safety Checks

devpipe includes safety checks to prevent accidental execution in dangerous system directories.

## Directory Safety Validation

devpipe will refuse to run in the following directories:

### Strictly Blocked (including all subdirectories)
- `/usr/bin` - System binaries
- `/usr/sbin` - System administration binaries  
- `/usr/lib` - System libraries
- `/etc` - System configuration
- `/bin` - Essential system binaries
- `/sbin` - System administration binaries
- `/boot` - Boot loader files
- `/System` - macOS system files
- `/Library` - macOS system libraries
- `/Applications` - macOS applications directory
- `/dev` - Device files
- `/proc` - Process information (Linux)
- `/sys` - System information (Linux)

### Top-Level Only Blocked (subdirectories are allowed)
- `/usr` - Blocked, but `/usr/local/*` and `/usr/src/*` are allowed
- `/var` - Blocked, but `/var/lib/myapp` is allowed
- `/tmp` - Blocked, but `/tmp/myproject` is allowed
- `/Volumes` - Blocked, but `/Volumes/MyDrive/project` is allowed

### Always Safe
- `/usr/local/*` - User-installed software (Homebrew, custom apps)
- `/usr/src/*` - Source code directories
- `/opt/*` - Optional software packages
- User home directories (e.g., `/Users/username`, `/home/username`)
- External drives and network shares (e.g., `/Volumes/ExternalDrive/project`)
- Relative paths (e.g., `.`, `./project`)

## Error Messages

### Unsafe Project Root

If the detected project root is in an unsafe directory:

```
ERROR: Refusing to run devpipe in system directory: /usr
This safety check prevents accidental execution in critical system paths.
Please run devpipe from your project directory, or set projectRoot in your config.
```

### Unsafe Output Directory

If an absolute `outputRoot` resolves to an unsafe directory:

```
ERROR: Output directory resolves to dangerous location: /etc/devpipe
This safety check prevents accidental execution in critical system paths.
Use a safe location like /tmp/devpipe or a relative path within your project.
```

**Note:** Absolute paths are validated for safety. Symlinks are resolved to check the actual destination.

## Workarounds

If you need to run devpipe tasks in a specific directory:

### 1. Change to Your Project Directory
```bash
cd /path/to/your/project
devpipe
```

### 2. Set `projectRoot` in Config
Override the auto-detected project root:

```toml
[defaults]
projectRoot = "/path/to/your/project"
```

This tells devpipe where your project is, regardless of where you run it from.

### 3. Use Absolute `workdir` in Config
Set an absolute path in your `config.toml`:

```toml
[task_defaults]
workdir = "/path/to/your/project"
```

Or per-task:

```toml
[tasks.build]
command = "make build"
workdir = "/path/to/your/project"
```

### 4. Use Absolute Config Path
When using an absolute config path, devpipe auto-detects the project root from the config location:

```bash
cd /
devpipe --config /path/to/your/project/config.toml
# → Project root auto-detected as /path/to/your/project
```

## Design Rationale

The safety check prevents:
1. **Accidental system damage** - Running tasks that modify files in system directories
2. **Permission errors** - Attempting to create `.devpipe` output in read-only locations
3. **Confusion** - Making it clear where devpipe is operating

The check is intentionally conservative - it's better to block a legitimate use case (which can be worked around) than to allow accidental system modification.
