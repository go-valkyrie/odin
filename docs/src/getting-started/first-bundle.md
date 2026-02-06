# Your First Bundle

Now that you've seen the quick start, let's build a more realistic bundle that demonstrates Odin's key features.

## What We'll Build

We'll create a bundle that deploys a simple web application with:

- A web server component
- A Redis cache component  
- Configurable environments (dev, staging, prod)
- Health checks and resource limits

## Project Setup

```bash
mkdir web-app
cd web-app
odin init --module example.com/web-app
```

## Bundle Structure

Create the following file structure:

```
web-app/
├── bundle.cue              # Main bundle definition
├── components/
│   ├── webapp.cue          # Web application component
│   └── redis.cue           # Redis component
├── environments/
│   ├── dev.cue             # Development values
│   ├── staging.cue         # Staging values
│   └── prod.cue            # Production values
└── cue.mod/
    └── module.cue
```

## Define the Web App Component

Create `components/webapp.cue`:

```cue
package components

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
)

webapp: odin.#Component & {
    metadata: name: "webapp"

    config: {
        // Image configuration
        image: {
            repository: string
            tag: string
        }

        // Scaling
        replicas: int | *1

        // Environment variables
        env: [string]: string

        // Resource limits
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

        // Health check configuration
        healthCheck: {
            path: string | *"/health"
            port: int | *8080
        }
    }

    resources: {
        deployment: {
            apiVersion: "apps/v1"
            kind: "Deployment"
            metadata: {
                name: "webapp"
                labels: app: "webapp"
            }
            spec: {
                replicas: config.replicas
                selector: matchLabels: app: "webapp"
                template: {
                    metadata: labels: app: "webapp"
                    spec: {
                        containers: [{
                            name: "webapp"
                            image: "\(config.image.repository):\(config.image.tag)"
                            
                            // Convert env map to list of env vars
                            env: [
                                for k, v in config.env {
                                    name: k
                                    value: v
                                }
                            ]

                            ports: [{
                                containerPort: config.healthCheck.port
                                name: "http"
                            }]

                            resources: {
                                requests: config.resources.requests
                                limits: config.resources.limits
                            }

                            livenessProbe: {
                                httpGet: {
                                    path: config.healthCheck.path
                                    port: "http"
                                }
                                initialDelaySeconds: 30
                                periodSeconds: 10
                            }

                            readinessProbe: {
                                httpGet: {
                                    path: config.healthCheck.path
                                    port: "http"
                                }
                                initialDelaySeconds: 5
                                periodSeconds: 5
                            }
                        }]
                    }
                }
            }
        }

        service: {
            apiVersion: "v1"
            kind: "Service"
            metadata: {
                name: "webapp"
                labels: app: "webapp"
            }
            spec: {
                selector: app: "webapp"
                ports: [{
                    port: 80
                    targetPort: "http"
                    name: "http"
                }]
            }
        }
    }
}
```

## Define the Redis Component

Create `components/redis.cue`:

```cue
package components

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
)

redis: odin.#Component & {
    metadata: name: "redis"

    config: {
        // Redis version
        version: string | *"7.2"

        // Persistence
        persistence: {
            enabled: bool | *true
            size: string | *"1Gi"
        }

        // Resource limits
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
    }

    resources: {
        deployment: {
            apiVersion: "apps/v1"
            kind: "Deployment"
            metadata: {
                name: "redis"
                labels: app: "redis"
            }
            spec: {
                replicas: 1
                selector: matchLabels: app: "redis"
                template: {
                    metadata: labels: app: "redis"
                    spec: {
                        containers: [{
                            name: "redis"
                            image: "redis:\(config.version)"
                            ports: [{
                                containerPort: 6379
                                name: "redis"
                            }]
                            resources: {
                                requests: config.resources.requests
                                limits: config.resources.limits
                            }
                            if config.persistence.enabled {
                                volumeMounts: [{
                                    name: "data"
                                    mountPath: "/data"
                                }]
                            }
                        }]
                        if config.persistence.enabled {
                            volumes: [{
                                name: "data"
                                persistentVolumeClaim: {
                                    claimName: "redis-data"
                                }
                            }]
                        }
                    }
                }
            }
        }

        service: {
            apiVersion: "v1"
            kind: "Service"
            metadata: {
                name: "redis"
                labels: app: "redis"
            }
            spec: {
                selector: app: "redis"
                ports: [{
                    port: 6379
                    targetPort: "redis"
                    name: "redis"
                }]
            }
        }

        if config.persistence.enabled {
            pvc: {
                apiVersion: "v1"
                kind: "PersistentVolumeClaim"
                metadata: {
                    name: "redis-data"
                    labels: app: "redis"
                }
                spec: {
                    accessModes: ["ReadWriteOnce"]
                    resources: requests: storage: config.persistence.size
                }
            }
        }
    }
}
```

## Create the Main Bundle

Create `bundle.cue`:

```cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    "example.com/web-app/components"
)

odin.#Bundle & {
    metadata: name: "web-app"

    // Include our components
    components

    // Default values
    values: {
        components: {
            webapp: {
                image: {
                    repository: "nginx"
                    tag: "latest"
                }
                replicas: 1
                env: {
                    REDIS_HOST: "redis"
                    REDIS_PORT: "6379"
                }
            }

            redis: {
                version: "7.2"
                persistence: enabled: false
            }
        }
    }
}
```

## Environment-Specific Values

Create `environments/dev.cue`:

```cue
package main

values: {
    components: {
        webapp: {
            replicas: 1
            resources: {
                requests: {
                    memory: "64Mi"
                    cpu: "50m"
                }
                limits: {
                    memory: "128Mi"
                    cpu: "100m"
                }
            }
        }
        redis: {
            persistence: enabled: false
        }
    }
}
```

Create `environments/prod.cue`:

```cue
package main

values: {
    components: {
        webapp: {
            image: tag: "v1.0.0"  // Use specific version in prod
            replicas: 3
            resources: {
                requests: {
                    memory: "256Mi"
                    cpu: "200m"
                }
                limits: {
                    memory: "512Mi"
                    cpu: "500m"
                }
            }
        }
        redis: {
            version: "7.2.4"  // Pin specific version
            persistence: {
                enabled: true
                size: "10Gi"
            }
            resources: {
                requests: {
                    memory: "512Mi"
                    cpu: "250m"
                }
                limits: {
                    memory: "1Gi"
                    cpu: "500m"
                }
            }
        }
    }
}
```

## Generate Manifests

For development:

```bash
odin template . -f environments/dev.cue > manifests/dev.yaml
```

For production:

```bash
odin template . -f environments/prod.cue > manifests/prod.yaml
```

## Key Takeaways

This example demonstrates several important Odin concepts:

1. **Component organization** - Keep components in separate files for clarity
2. **Configuration schema** - Define clear config structures with defaults
3. **Conditional resources** - Use `if` to conditionally include resources (like PVCs)
4. **CUE features** - String interpolation, list comprehensions, and type constraints
5. **Environment separation** - Use value files to manage different environments
6. **Type safety** - CUE catches configuration errors before rendering

## Next Steps

- Learn about [component templates](../concepts/templates.md) to reuse components
- Explore the [CLI reference](../cli/overview.md) for more template options
- Set up [ArgoCD integration](../integration/argocd.md) for automated deployments
