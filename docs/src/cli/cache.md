# odin cache

Utilities for working with Odin's cache.

## Synopsis

```bash
odin cache <subcommand> [flags]
```

## Description

The `cache` command provides utilities for managing Odin's module cache.

## Subcommands

### clean

Clean the cache directory:

```bash
odin cache clean
```

This removes all cached modules, forcing Odin to re-download them on next use.

## Cache Directory

Odin caches CUE modules at:

```
$USER_CACHE_DIR/odin/
```

Typically:
- **Linux:** `~/.cache/odin/`
- **macOS:** `~/Library/Caches/odin/`
- **Windows:** `%LocalAppData%\odin\`

## Cache Structure

The cache is organized by registry configuration hash:

```
odin/
├── abc123def456/
│   └── modules/
│       ├── platform.example.com/
│       └── go-valkyrie.com/
└── 789ghi012jkl/
    └── modules/
```

Different registry configurations use separate cache directories to avoid conflicts.

## When to Clean Cache

Clean the cache when:

- You've updated registry configuration
- You suspect cached modules are corrupted
- You want to free up disk space
- You're troubleshooting module loading issues

## Cache Size

Check cache size:

```bash
du -sh $(odin config eval | grep cacheDir | awk '{print $2}')
```

Or on macOS/Linux:

```bash
du -sh ~/Library/Caches/odin  # macOS
du -sh ~/.cache/odin          # Linux
```

## Selective Cache Cleaning

To clean cache for specific modules:

```bash
# Find cache directory
cache_dir=$(odin config eval | grep cacheDir | awk '{print $2}')

# Remove specific module
rm -rf "$cache_dir"/*/modules/platform.example.com
```

## Cache Behavior

Odin caches:
- Downloaded CUE modules
- Module metadata
- Registry responses

The cache is automatically populated on first use and doesn't expire unless you clean it manually.

## Examples

Clean entire cache:

```bash
odin cache clean
```

## Next Steps

- Learn about [Configuration](../guides/configuration.md)
- See [odin config](./config.md) for cache location

## See Also

- [odin config](./config.md)
- [Configuration Guide](../guides/configuration.md)
