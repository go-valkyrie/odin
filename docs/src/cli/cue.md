# odin cue

Passthrough to CUE CLI with configured registries.

## Synopsis

```bash
odin cue <cue-command> [cue-flags] [arguments]
```

## Description

The `cue` command is a passthrough to the CUE CLI that automatically configures registry settings from Odin's configuration.

This is useful when you need to run CUE commands (like `cue mod tidy`, `cue eval`, etc.) with the same registry configuration that Odin uses.

## Examples

Tidy module dependencies:

```bash
odin cue mod tidy
```

Evaluate a CUE file:

```bash
odin cue eval bundle.cue
```

Export to JSON:

```bash
odin cue export bundle.cue
```

Format CUE files:

```bash
odin cue fmt bundle.cue
```

Vet a CUE package:

```bash
odin cue vet
```

## How It Works

The command:

1. Loads Odin's configuration
2. Sets up CUE registry environment variables
3. Executes the CUE CLI with your arguments

This ensures CUE commands use the same registry configuration as `odin template`.

## Why Use This?

Instead of:

```bash
# Manual registry configuration
export CUE_REGISTRY=/path/to/registryconfig.cue
cue mod tidy
```

You can simply:

```bash
# Automatic registry configuration
odin cue mod tidy
```

## Common CUE Commands

### Module Management

```bash
# Tidy dependencies
odin cue mod tidy

# Publish module
odin cue mod publish v0.1.0

# Get module
odin cue mod get platform.example.com/webapp@v0.1.0
```

### Evaluation and Export

```bash
# Evaluate CUE
odin cue eval bundle.cue

# Export to JSON
odin cue export bundle.cue -o bundle.json

# Export to YAML
odin cue export bundle.cue --out yaml
```

### Formatting and Validation

```bash
# Format CUE files
odin cue fmt .

# Vet package
odin cue vet

# Validate against schema
odin cue vet data.cue schema.cue
```

## Limitations

This is a pure passthrough - Odin doesn't modify CUE's behavior, only configures registries.

For full CUE CLI documentation, see:

```bash
cue help
```

Or visit: https://cuelang.org/docs/

## Next Steps

- Learn about [Configuration & Registries](../guides/configuration.md)
- See [Publishing Components](../guides/publishing.md)
- Read the [CUE documentation](https://cuelang.org/docs/)

## See Also

- [odin config](./config.md)
- [Configuration Guide](../guides/configuration.md)
