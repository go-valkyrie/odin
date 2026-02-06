# Best Practices

This guide covers best practices for working with Odin.

## Bundle Organization

### Keep Bundles Focused

**Good:** One bundle per application

```
my-webapp/              # One bundle
├── bundle.cue
├── components/
└── cue.mod/
```

**Avoid:** Many unrelated apps in one bundle

```
everything/
├── webapp.cue          # Too many concerns
├── database.cue
├── monitoring.cue
├── logging.cue
└── ...
```

### Use Clear Directory Structure

```
my-bundle/
├── bundle.cue              # Entry point
├── components/             # Component definitions
│   ├── webapp.cue
│   └── database.cue
├── lib/                    # Shared code
│   └── common.cue
├── environments/           # Environment values
│   ├── dev.cue
│   ├── staging.cue
│   └── prod.cue
├── values.cue              # Default values
└── cue.mod/
    └── module.cue
```

### Document Your Bundle

Include a README with:

- Purpose and scope
- How to render manifests
- Environment descriptions
- Required dependencies
- Contact information

## Configuration

### Provide Sensible Defaults

```cue
config: {
    // Good: most values have defaults
    image: string                      // Required
    replicas: int | *1                 // Default: 1
    port: int | *8080                  // Default: 8080
    resources: {
        requests: {
            memory: string | *"128Mi"  // Default
            cpu: string | *"100m"      // Default
        }
    }
}
```

### Use Constraints

```cue
config: {
    // Numeric ranges
    replicas: int & >=1 & <=100
    port: int & >0 & <65536
    
    // Enumerations
    environment: "dev" | "staging" | "prod"
    logLevel: "debug" | "info" | "warn" | "error"
    
    // Pattern matching
    email: =~"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
}
```

### Document Configuration

```cue
config: {
    // Number of pod replicas to run
    // Valid range: 1-100
    // Default: 1
    replicas: int & >=1 & <=100 | *1
    
    // Container image configuration
    image: {
        // Image repository without tag
        // Example: "ghcr.io/myorg/myapp"
        repository: string
        
        // Image tag or digest
        // Default: "latest"
        // Note: Using "latest" in production is not recommended
        tag: string | *"latest"
    }
}
```

## Components

### Keep Resources DRY

```cue
// Good: define once, reference everywhere
#appName: metadata.name

resources: {
    deployment: {
        metadata: name: #appName
        spec: {
            selector: matchLabels: app: #appName
            template: metadata: labels: app: #appName
        }
    }
    service: {
        metadata: name: #appName
        spec: selector: app: #appName
    }
}
```

### Use Meaningful Resource Names

```cue
resources: {
    // Good: descriptive
    deployment: {...}
    service: {...}
    configmap: {...}
    ingress: {...}
    
    // Bad: unclear
    res1: {...}
    thing: {...}
}
```

### Make Features Optional

```cue
config: {
    ingress: {
        enabled: bool | *false
        hostname?: string  // Only required if enabled
    }
}

resources: {
    deployment: {...}
    service: {...}
    
    if config.ingress.enabled {
        ingress: {...}
    }
}
```

## Values

### Layer Values Logically

```bash
# Clear hierarchy
base.cue → environments/prod.cue → regional/us-east.cue
```

### Pin Versions in Production

```cue
// environments/prod.cue
package main

values: {
    components: {
        webapp: {
            // Good: specific version
            image: tag: "v1.2.0"
            
            // Bad: floating tag
            // image: tag: "latest"
        }
    }
}
```

### Use Global Values

```cue
values: {
    // Global configuration
    domain: "example.com"
    environment: "production"
    
    // Components reference globals
    components: {
        webapp: {
            hostname: "\(values.domain)"
            env: {
                ENVIRONMENT: values.environment
            }
        }
    }
}
```

## Templates

### Document Your Templates

```markdown
# My Component Template

A reusable component for deploying web applications.

## Configuration

- `image.repository` (required) - Container image
- `replicas` (optional, default: 1) - Number of replicas
- `port` (optional, default: 8080) - Container port
```

### Provide Examples

Include basic and advanced examples in `examples/` directory.

### Version Carefully

- Use v0 during development
- Release v1.0.0 when API is stable
- Never break v1.x compatibility
- Use v2 for breaking changes

### Test Thoroughly

Create test bundles with expected output:

```
test/
├── basic/
│   ├── bundle.cue
│   └── expected.yaml
└── advanced/
    ├── bundle.cue
    └── expected.yaml
```

## Git & GitOps

### Use Conventional Commits

```
feat: add Redis cache component
fix: correct database connection string
chore: update webapp to v1.2.0
docs: add deployment instructions
```

### Protect Main Branch

Require:
- Pull request reviews
- Passing CI checks
- Up-to-date with base branch

### Automated Validation

```yaml
# CI validation
- name: Validate bundle
  run: |
    odin template . -f environments/prod.cue | \
      kubectl apply --dry-run=server -f -
```

### Clear PR Descriptions

Include:
- What changed
- Why it changed
- How to test
- Deployment notes

## Security

### Don't Commit Secrets

```cue
// Good: reference secrets
config: {
    database: {
        passwordSecret: {
            name: "db-credentials"
            key: "password"
        }
    }
}

// Bad: inline secrets
config: {
    database: {
        password: "super-secret-123"  // Don't do this!
    }
}
```

### Use Read-Only Filesystems

```cue
resources: deployment: {
    spec: template: spec: {
        containers: [{
            securityContext: {
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: false
            }
        }]
    }
}
```

### Run as Non-Root

```cue
resources: deployment: {
    spec: template: spec: {
        securityContext: {
            runAsNonRoot: true
            runAsUser: 1000
            fsGroup: 1000
        }
    }
}
```

## Performance

### Cache Dependencies

Odin caches modules automatically. Don't disable it unless necessary.

### Use Specific Versions

```cue
// Good: fast lookup
deps: {
    "platform.example.com/webapp/v1alpha1@v0": {
        v: "v0.2.5"
    }
}

// Slow: needs to find latest
deps: {
    "platform.example.com/webapp/v1alpha1@v0": {
        v: ">=v0.2.0"
    }
}
```

### Keep Bundles Small

Large bundles take longer to evaluate. Split them if needed.

## Validation

### Validate Early

```bash
# Validate syntax
cue vet bundle.cue

# Validate against schema
odin template . > /dev/null

# Validate with Kubernetes
odin template . | kubectl apply --dry-run=server -f -
```

### Use Type Constraints

```cue
config: {
    // CUE validates types
    replicas: int           // Must be integer
    image: string           // Must be string
    enabled: bool           // Must be boolean
    
    // Add constraints
    port: int & >0 & <65536
    environment: "dev" | "staging" | "prod"
}
```

## Troubleshooting

### Enable Debug Logging

```bash
odin --debug template .
```

### Check Cache

```bash
# View cache location
odin config eval | grep cacheDir

# Clean cache if issues
odin cache clean
```

### Validate Registry Config

```bash
# View effective config
odin config eval

# Check module access
odin cue mod get platform.example.com/webapp/v1alpha1@v0.1.0
```

## Next Steps

- See [Troubleshooting](./troubleshooting.md) for common issues
- Learn [Configuration](../guides/configuration.md) for registry setup
- Explore [Writing Components](../guides/writing-components.md)
