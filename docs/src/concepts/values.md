# Values

**Values** are the configuration data that flow into your components. They allow you to separate configuration from definition, making it easy to deploy the same bundle with different settings across environments.

## What are Values?

Values are:

- Configuration data defined in the bundle's `values` section
- Automatically bound to component `config` sections
- Mergeable from external files using `odin template -f`
- Type-checked against component schemas

## Basic Values Structure

The `values` section in a bundle contains configuration:

```cue
odin.#Bundle & {
    components: {
        webapp: odin.#Component & {
            config: {
                image: string
                replicas: int
            }
        }
    }
    
    values: {
        components: {
            webapp: {
                image: "nginx:latest"
                replicas: 3
            }
        }
    }
}
```

## Automatic Binding

Odin automatically binds `values.components.<name>` to `components.<name>.config`:

```cue
components: {
    webapp: odin.#Component & {
        metadata: name: "webapp"
        
        // This config schema...
        config: {
            image: string
            replicas: int | *1
        }
        
        resources: deployment: {
            spec: {
                // ...uses config fields
                replicas: config.replicas
                template: spec: containers: [{
                    image: config.image
                }]
            }
        }
    }
}

values: {
    components: {
        // ...is automatically bound from values
        webapp: {
            image: "nginx:latest"
            replicas: 3
        }
    }
}
```

This bidirectional binding is handled by Odin's internal schema.

## Value Files

You can override values using external files with `odin template -f`:

**values/dev.cue:**

```cue
package main

values: {
    components: {
        webapp: {
            image: "nginx:alpine"
            replicas: 1
        }
    }
}
```

**values/prod.cue:**

```cue
package main

values: {
    components: {
        webapp: {
            image: "nginx:1.25"
            replicas: 5
        }
    }
}
```

Then render with different values:

```bash
# Development
odin template . -f values/dev.cue > manifests/dev.yaml

# Production
odin template . -f values/prod.cue > manifests/prod.yaml
```

## Merging Values

Multiple value files are merged in order:

```bash
odin template . -f base-values.cue -f prod-values.cue -f override.cue
```

Later files override earlier ones. This allows layered configuration:

```
base-values.cue       → Common defaults
↓
env/prod.cue          → Environment-specific
↓
regional/us-west.cue  → Regional overrides
↓
local-override.cue    → Local development tweaks
```

## Value Defaults

Components can specify defaults in their config schema:

```cue
config: {
    // Required field
    image: string
    
    // Optional with default
    replicas: int | *1
    
    // Nested defaults
    resources: {
        requests: {
            memory: string | *"128Mi"
            cpu: string | *"100m"
        }
    }
    
    // Default list
    ports: [...int] | *[8080]
}
```

Values override defaults:

```cue
values: {
    components: myapp: {
        image: "myapp:v1.0"
        // replicas uses default: 1
        // resources.requests uses defaults
        ports: [8080, 8443]  // Overrides default
    }
}
```

## Global Values

You can define global values that aren't component-specific:

```cue
values: {
    // Global values
    domain: "example.com"
    environment: "production"
    region: "us-west-2"
    
    // Component-specific values
    components: {
        webapp: {
            // Can reference global values
            hostname: "\(values.domain)"
            env: {
                ENVIRONMENT: values.environment
                REGION: values.region
            }
        }
    }
}
```

## Structured Values

Values can be deeply nested:

```cue
values: {
    components: {
        webapp: {
            image: {
                repository: "ghcr.io/myorg/webapp"
                tag: "v1.0.0"
                pullPolicy: "IfNotPresent"
            }
            
            database: {
                host: "postgres.default.svc"
                port: 5432
                credentials: {
                    secretName: "db-credentials"
                    usernameKey: "username"
                    passwordKey: "password"
                }
            }
            
            features: {
                auth: enabled: true
                metrics: enabled: true
                tracing: {
                    enabled: true
                    endpoint: "tempo.observability:4317"
                }
            }
        }
    }
}
```

## Type Safety

Values are validated against component config schemas:

```cue
config: {
    replicas: int & >=1 & <=10
    environment: "dev" | "staging" | "prod"
}

values: {
    components: myapp: {
        replicas: 15  // Error: exceeds maximum
        environment: "test"  // Error: not in allowed values
    }
}
```

CUE catches these errors before rendering.

## Value Transformations

You can transform values before using them:

```cue
config: {
    image: {
        repository: string
        tag: string
    }
    
    // Computed field
    fullImage: "\(image.repository):\(image.tag)"
}

resources: {
    deployment: {
        spec: template: spec: containers: [{
            image: config.fullImage
        }]
    }
}
```

## Secret References

Values can reference Kubernetes secrets:

```cue
config: {
    database: {
        host: string
        port: int
        passwordSecret: {
            name: string
            key: string
        }
    }
}

values: {
    components: webapp: {
        database: {
            host: "postgres.default.svc"
            port: 5432
            passwordSecret: {
                name: "db-credentials"
                key: "password"
            }
        }
    }
}

resources: {
    deployment: {
        spec: template: spec: containers: [{
            env: [
                {
                    name: "DB_HOST"
                    value: config.database.host
                },
                {
                    name: "DB_PASSWORD"
                    valueFrom: secretKeyRef: {
                        name: config.database.passwordSecret.name
                        key: config.database.passwordSecret.key
                    }
                },
            ]
        }]
    }
}
```

## Environment-Based Values

A common pattern is environment-specific value files:

```
bundle/
├── bundle.cue
├── values.cue           # Defaults
└── environments/
    ├── dev.cue          # Development overrides
    ├── staging.cue      # Staging overrides
    └── prod.cue         # Production overrides
```

**environments/prod.cue:**

```cue
package main

values: {
    // Production-specific global values
    domain: "app.example.com"
    environment: "production"
    
    components: {
        webapp: {
            replicas: 5
            resources: {
                requests: {
                    memory: "512Mi"
                    cpu: "500m"
                }
                limits: {
                    memory: "1Gi"
                    cpu: "1000m"
                }
            }
        }
        
        database: {
            size: "50Gi"
            instanceType: "db.r5.large"
        }
    }
}
```

## Values in ArgoCD

When using ArgoCD, you can specify value files in the Application spec:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  source:
    plugin:
      name: odin
      parameters:
      - name: valuesFile
        string: environments/prod.cue
```

See [ArgoCD Integration](../integration/argocd.md) for details.

## Best Practices

### 1. Use Defaults Liberally

```cue
config: {
    // Good: sensible defaults
    replicas: int | *1
    image: pullPolicy: string | *"IfNotPresent"
}
```

### 2. Group Related Values

```cue
config: {
    // Good: grouped by concern
    image: {
        repository: string
        tag: string
        pullPolicy: string
    }
    
    resources: {
        requests: {...}
        limits: {...}
    }
}
```

### 3. Document Configuration

```cue
config: {
    // Number of pod replicas to run
    // Valid range: 1-10
    replicas: int & >=1 & <=10 | *1
    
    // Container image configuration
    image: {
        // Image repository (without tag)
        repository: string
        
        // Image tag or digest
        tag: string
    }
}
```

### 4. Validate Early

```cue
config: {
    // Use constraints
    port: int & >0 & <65536
    environment: "dev" | "staging" | "prod"
    
    // Require specific formats
    email: =~"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
}
```

## Next Steps

- Learn about [Component Templates](./templates.md) - reusable components
- See [Working with Values](../guides/values.md) - advanced patterns
- Explore [Configuration](../guides/configuration.md) - registry and cache setup
