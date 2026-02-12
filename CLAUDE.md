# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Odin is a CLI tool for generating Kubernetes manifests using CUE. It is designed to work alongside Valkyrie and Freyr but can be used standalone in any GitOps pipeline. The core concept is "bundles" - sets of components with configurable values that render to Kubernetes manifests.

## Common Commands

### Building and Testing
```bash
# Run all tests
make test
# or
go test ./...

# Build container locally
make container-local

# Build multi-arch containers
make container

# Push container (requires REPOSITORY and TAG)
make push-container
```

### Development Workflow
```bash
# Run odin commands during development
go run cmd/odin/main.go <command>

# Common odin commands
odin template <bundle-path>              # Render manifests from a bundle
odin show values <bundle-path>           # Show the values schema
odin test <module-path>                  # Run testscript integration tests
odin pull <registry>/<bundle>:<version>  # Pull bundle from OCI registry
odin push <registry>/<bundle>:<version>  # Push bundle to OCI registry
odin init <name>                         # Initialize new bundle
odin components                          # List available component templates
odin docs <component>                    # Show component documentation
odin cue <args>                          # Proxy to cue command with bundle context
```

## Architecture

### Core Packages

**`pkg/model`**: Bundle loading and manipulation. The `Bundle` type is the core abstraction representing an Odin bundle with its components, values, and resources. Uses CUE's `cue.Context` for evaluating configurations.

**`pkg/schema`**: Schema extraction and manipulation utilities for working with CUE schemas.

**`pkg/oci`**: OCI registry interactions for pulling/pushing bundles.

**`pkg/odintest`**: Testing utilities including `SetupRegistry()` for spinning up in-process CUE module registries during tests.

**`internal/config`**: Configuration file (`odin.toml`) loading and management. Handles registry configuration.

**`internal/schema`**: Internal schema validation and processing logic.

**`internal/cmd`**: Shared command infrastructure and utilities.

### CUE API Definitions

The `api/v1alpha1/` directory contains CUE schemas that define the Bundle and Component structures. These are the contract that all bundles must conform to.

### CLI Structure

The CLI uses Cobra and is organized under `cmd/odin/cmd/`. Each major command has its own file (e.g., `template.go`, `show.go`, `test.go`). The `root.go` sets up shared context including logger and config manager.

### Testing Approach

Integration tests use the `testscript` framework (see `internal/integration/template_test.go`). Tests run actual odin commands as subprocesses against fixture bundles. Use `go test ./internal/integration -update` to update golden files.

Unit tests follow standard Go conventions. The `pkg/odintest` package provides helpers for setting up test registries.

## Bundle Structure

An Odin bundle is a CUE module containing:
- `cue.mod/module.cue` - CUE module definition
- `odin.toml` - Registry configuration for pulling dependencies
- CUE files defining components conforming to `odin.#Bundle` schema
- Components contain resources (Kubernetes manifests) and a config schema
- Values are provided separately and merged with component configs

## Git Workflow

### Commit Requirements

**IMPORTANT**: Always present commit messages to the user for review before creating commits. Never commit without explicit user confirmation.

All commits must:
1. Use [Conventional Commits](https://www.conventionalcommits.org/) format: `<type>[optional scope]: <description>`
2. Include DCO sign-off: Use `git commit -s` to automatically add `Signed-off-by` line (assumes git user.email is configured)
3. Include AI co-author attribution when AI assistance is used

Valid commit types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

Example commit message:
```
feat(template): add support for custom resource filtering

Adds filtering logic to exclude resources based on labels.

Co-authored-by: Claude Sonnet 4.5 <noreply@anthropic.com>
```

Note: The `Signed-off-by` line will be automatically added by `git commit -s`.

### AI Usage Policy

This project requires disclosure of AI assistance. When using AI tools:
- Add the AI tool as a co-author in commit messages
- Ensure you understand and can explain all contributed code
- Avoid "prompt and submit" - engage in iterative refinement
- Review AI-generated test data and commit messages carefully

See CONTRIBUTING.md for complete details.

## Environment

The project uses `mise` for tooling. Key environment variables:
- `CUE_REGISTRY`: Set in mise.toml to point to registry config
- Cache directory: `~/.cache/odin` by default, configurable via `--cache-dir`

## Development Notes

### License Headers
Use SPDX identifier format: `// SPDX-License-Identifier: MIT`

### Logging
Use structured logging with `log/slog`. The CLI configures a logger with tint formatting and prefix support. Access via context: `logger := cmd.Context().Value(loggerCtxKey).(*slog.Logger)`

### CUE Module Dependencies
Odin bundles can depend on CUE modules from OCI registries. Registry mappings are configured in `odin.toml` using `[[registries]]` sections with `module-prefix` and `registry` fields.