# CLI Overview

Odin provides a command-line interface for working with bundles, templates, and manifests.

## Command Structure

```bash
odin [global-flags] <command> [command-flags] [arguments]
```

## Global Flags

These flags work with all commands:

- `--debug` - Enable debug logging
- `-v, --verbose` - Enable verbose output
- `-h, --help` - Show help for any command

## Commands

### Core Commands

- [`odin init`](./init.md) - Initialize a new bundle
- [`odin template`](./template.md) - Render manifests from a bundle
- [`odin components`](./components.md) - List available component templates
- [`odin docs`](./docs.md) - Show documentation for templates

### Configuration Commands

- [`odin config`](./config.md) - Manage Odin configuration
- [`odin cache`](./cache.md) - Manage the module cache

### Utility Commands

- [`odin cue`](./cue.md) - Passthrough to CUE CLI with configured registries
- `odin completion` - Generate shell completion scripts

## Getting Help

### Command Help

Show help for any command:

```bash
odin template --help
odin docs --help
```

### Examples

Most commands include examples in their help output:

```bash
odin template --help
```

```
Examples:
  # Render bundle in current directory
  odin template .

  # Render with custom values
  odin template . -f values/prod.cue

  # Render remote bundle
  odin template oci://ghcr.io/myorg/bundle:v1.0.0
```

## Common Workflows

### Initialize and Render

```bash
# Create a new bundle
mkdir my-app && cd my-app
odin init

# Edit bundle.cue to define components
vim bundle.cue

# Render manifests
odin template .
```

### Explore Templates

```bash
# List available templates
odin components

# View template documentation
odin docs webapp

# Use template in your bundle
# (edit bundle.cue to import and use)
```

### Environment Management

```bash
# Render for different environments
odin template . -f environments/dev.cue > manifests/dev.yaml
odin template . -f environments/staging.cue > manifests/staging.yaml
odin template . -f environments/prod.cue > manifests/prod.yaml
```

### Configuration

```bash
# View current configuration
odin config eval

# Clean cache
odin cache clean
```

## Exit Codes

Odin uses standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Misuse of command (invalid flags, etc.)

## Shell Completion

Generate completion scripts for your shell:

```bash
# Bash
odin completion bash > /etc/bash_completion.d/odin

# Zsh
odin completion zsh > "${fpath[1]}/_odin"

# Fish
odin completion fish > ~/.config/fish/completions/odin.fish

# PowerShell
odin completion powershell > odin.ps1
```

## Environment Variables

Odin respects these environment variables:

- `CUE_REGISTRY` - Override CUE registry configuration
- `XDG_CACHE_HOME` - Override cache directory location
- `HOME` - Used to find config directory (`~/.config/odin/`)

## Debug Mode

Enable debug logging to troubleshoot issues:

```bash
odin --debug template .
```

This shows:

- CUE module loading
- Registry operations
- Cache hits/misses
- Template resolution

## Verbose Output

Enable verbose output for more details:

```bash
odin -v template .
```

This shows:

- Components being processed
- Resources being generated
- Value merging operations

## Next Steps

Learn about specific commands:

- [odin init](./init.md) - Create new bundles
- [odin template](./template.md) - Render manifests
- [odin components](./components.md) - Discover templates
- [odin docs](./docs.md) - View documentation
