# Writing Components

This guide covers best practices and patterns for writing Odin components.

## Component Anatomy

Every component has three parts:

```cue
import odin "go-valkyrie.com/odin/api/v1alpha1"

mycomponent: odin.#Component & {
    // 1. Metadata - identifies the component
    metadata: {
        name: "mycomponent"
        labels?: [string]: string
        annotations?: [string]: string
    }
    
    // 2. Config - schema for configuration
    config: {
        // Your configuration fields
    }
    
    // 3. Resources - Kubernetes manifests
    resources: {
        // Your Kubernetes resources
    }
}
```

## Define a Clear Config Schema

### Use Descriptive Names

```cue
config: {
    // Good: clear and specific
    replicas: int
    containerPort: int
    healthCheckPath: string
    
    // Bad: vague names
    count: int
    port: int
    path: string
}
```

### Provide Defaults

```cue
config: {
    // Optional with default
    replicas: int | *1
    port: int | *8080
    pullPolicy: string | *"IfNotPresent"
    
    // Required (no default)
    image: string
}
```

### Group Related Fields

```cue
config: {
    // Good: grouped by concern
    image: {
        repository: string
        tag: string | *"latest"
        pullPolicy: string | *"IfNotPresent"
    }
    
    resources: {
        requests: {
            memory: string | *"128Mi"
            cpu: string | *"100m"
        }
        limits: {
            memory: string | *"256Mi"
            cpu: string | *"200m"
        }
    }
    
    // Bad: flat structure
    imageRepository: string
    imageTag: string
    imagePullPolicy: string
    requestsMemory: string
    requestsCpu: string
    limitsMemory: string
    limitsCpu: string
}
```

### Add Constraints

```cue
config: {
    // Numeric constraints
    replicas: int & >=1 & <=10
    port: int & >0 & <65536
    
    // String constraints (enums)
    environment: "dev" | "staging" | "prod"
    logLevel: "debug" | "info" | "warn" | "error"
    
    // Pattern matching
    email: =~"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    
    // Conditional requirements
    if ingress.enabled {
        ingress: hostname: string  // Required if enabled
    }
}
```

## Generate Resources

### Use Config Values

```cue
config: {
    image: string
    replicas: int
    port: int
}

resources: {
    deployment: {
        spec: {
            replicas: config.replicas
            template: spec: containers: [{
                image: config.image
                ports: [{
                    containerPort: config.port
                }]
            }]
        }
    }
}
```

### Reference Metadata

```cue
resources: {
    deployment: {
        metadata: {
            name: metadata.name
            labels: metadata.labels
        }
        spec: {
            selector: matchLabels: app: metadata.name
            template: {
                metadata: labels: app: metadata.name
            }
        }
    }
    
    service: {
        metadata: {
            name: metadata.name
            labels: metadata.labels
        }
        spec: selector: app: metadata.name
    }
}
```

### Use List Comprehensions

```cue
config: {
    env: [string]: string
    volumes: [Name=string]: {
        path: string
        size: string
    }
}

resources: {
    deployment: {
        spec: template: spec: {
            // Convert env map to list
            containers: [{
                env: [
                    for k, v in config.env {
                        name: k
                        value: v
                    }
                ]
                
                // Generate volume mounts
                volumeMounts: [
                    for name, vol in config.volumes {
                        name: name
                        mountPath: vol.path
                    }
                ]
            }]
            
            // Generate volumes
            volumes: [
                for name, vol in config.volumes {
                    name: name
                    persistentVolumeClaim: claimName: "\(metadata.name)-\(name)"
                }
            ]
        }
    }
    
    // Generate PVCs for each volume
    for name, vol in config.volumes {
        "pvc-\(name)": {
            apiVersion: "v1"
            kind: "PersistentVolumeClaim"
            metadata: name: "\(metadata.name)-\(name)"
            spec: {
                accessModes: ["ReadWriteOnce"]
                resources: requests: storage: vol.size
            }
        }
    }
}
```

## Conditional Resources

### Boolean Flags

```cue
config: {
    service: {
        enabled: bool | *true
        type: string | *"ClusterIP"
    }
    ingress: {
        enabled: bool | *false
        hostname: string
    }
}

resources: {
    deployment: {...}
    
    if config.service.enabled {
        service: {
            apiVersion: "v1"
            kind: "Service"
            // ...
        }
    }
    
    if config.ingress.enabled {
        ingress: {
            apiVersion: "networking.k8s.io/v1"
            kind: "Ingress"
            // ...
        }
    }
}
```

### Environment-Based

```cue
config: {
    environment: "dev" | "staging" | "prod"
}

resources: {
    deployment: {...}
    
    // Only create HPA in prod
    if config.environment == "prod" {
        hpa: {
            apiVersion: "autoscaling/v2"
            kind: "HorizontalPodAutoscaler"
            // ...
        }
    }
}
```

## Reusable Patterns

### Define Helpers

```cue
// Helper definitions (prefixed with #)
#containerSpec: {
    name: string
    image: string
    port: int
    
    _out: {
        name: name
        image: image
        ports: [{containerPort: port}]
    }
}

config: {
    containers: {
        app: #containerSpec & {
            name: "app"
            image: "myapp:latest"
            port: 8080
        }
        sidecar: #containerSpec & {
            name: "sidecar"
            image: "envoy:latest"
            port: 9901
        }
    }
}

resources: {
    deployment: {
        spec: template: spec: {
            containers: [
                for _, c in config.containers {
                    c._out
                }
            ]
        }
    }
}
```

### Common Labels

```cue
#labels: {
    app: metadata.name
    version: config.version
    environment: config.environment
}

resources: {
    deployment: {
        metadata: labels: #labels
        spec: {
            selector: matchLabels: app: metadata.name
            template: metadata: labels: #labels
        }
    }
    
    service: {
        metadata: labels: #labels
        spec: selector: app: metadata.name
    }
}
```

## Error Handling

### Validate Input

```cue
config: {
    replicas: int & >=1 & <=100
    
    database: {
        host: string
        port: int & >0 & <65536
        
        // At least one must be provided
        _: {password: string} | {passwordSecretName: string}
    }
}
```

### Provide Clear Messages

Use CUE's error messages:

```cue
config: {
    environment: "dev" | "staging" | "prod"
    
    // Clear constraint
    replicas: int & >=1 & <=10 | *1
    
    if environment == "prod" {
        // Require specific version in prod
        image: tag: !="latest"
    }
}
```

## Documentation

### Document Config Fields

```cue
config: {
    // Number of pod replicas to run
    // Valid range: 1-10
    // Default: 1
    replicas: int & >=1 & <=10 | *1
    
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

### Add Examples

Create `examples/` directory:

```
my-component/
├── v1alpha1/
│   └── component.cue
└── examples/
    ├── basic.cue
    ├── with-ingress.cue
    └── production.cue
```

**examples/basic.cue:**

```cue
package examples

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    mycomp "example.com/my-component/v1alpha1"
)

odin.#Bundle & {
    metadata: name: "basic-example"
    
    components: {
        app: mycomp.#Component & {
            metadata: name: "app"
        }
    }
    
    values: {
        components: app: {
            image: "nginx:latest"
            replicas: 2
        }
    }
}
```

## Testing Components

### Create Test Bundles

```
my-component/
├── v1alpha1/
│   └── component.cue
└── test/
    ├── basic/
    │   ├── bundle.cue
    │   └── expected.yaml
    ├── with-ingress/
    │   ├── bundle.cue
    │   └── expected.yaml
    └── production/
        ├── bundle.cue
        └── expected.yaml
```

### Validate Output

```bash
#!/bin/bash
# test.sh

for test in test/*/; do
    echo "Testing $(basename $test)..."
    
    # Render bundle
    actual=$(odin template "$test")
    
    # Compare with expected
    expected=$(cat "${test}/expected.yaml")
    
    if [ "$actual" == "$expected" ]; then
        echo "✓ $(basename $test) passed"
    else
        echo "✗ $(basename $test) failed"
        diff -u <(echo "$expected") <(echo "$actual")
    fi
done
```

## Best Practices

### 1. Make Everything Configurable

```cue
// Good: configurable
config: {
    probe: {
        path: string | *"/health"
        port: int | *8080
        initialDelay: int | *30
        period: int | *10
    }
}

// Bad: hardcoded
resources: deployment: {
    spec: template: spec: containers: [{
        livenessProbe: {
            httpGet: {
                path: "/health"
                port: 8080
            }
            initialDelaySeconds: 30
            periodSeconds: 10
        }
    }]
}
```

### 2. Use Sensible Defaults

Minimize required configuration:

```cue
config: {
    // Only image is required
    image: string
    
    // Everything else has defaults
    replicas: int | *1
    port: int | *8080
    resources: {
        requests: {
            memory: string | *"128Mi"
            cpu: string | *"100m"
        }
    }
}
```

### 3. Keep Resources DRY

```cue
// Good: single source of truth
#appName: metadata.name

resources: {
    deployment: {
        metadata: name: #appName
        spec: {
            selector: matchLabels: app: #appName
            template: {
                metadata: labels: app: #appName
            }
        }
    }
    
    service: {
        metadata: name: #appName
        spec: selector: app: #appName
    }
}
```

### 4. Validate Early

```cue
config: {
    // Validate at config time, not apply time
    replicas: int & >=1
    port: int & >0 & <65536
    environment: "dev" | "staging" | "prod"
}
```

## Next Steps

- See [Using Templates](./using-templates.md) to use components
- Learn [Publishing Components](./publishing.md) to share them
- Explore [Best Practices](../reference/best-practices.md)
