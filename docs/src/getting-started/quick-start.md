# Quick Start

This quick start guide will walk you through creating and rendering your first Odin bundle in under 5 minutes.

## Prerequisites

- Odin installed (see [Installation](./installation.md))
- Basic familiarity with Kubernetes resources
- A terminal

## Step 1: Initialize a Bundle

Create a new directory and initialize an Odin bundle:

```bash
mkdir my-app
cd my-app
odin init
```

Odin will prompt you for some information and create the basic bundle structure:

```
my-app/
├── bundle.cue          # Main bundle definition
├── cue.mod/
│   └── module.cue      # CUE module declaration
└── values.cue          # Default values (optional)
```

## Step 2: Define a Component

Open `bundle.cue` and you'll see a basic bundle structure. Let's add a simple component:

```cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
)

odin.#Bundle & {
    metadata: name: "my-app"

    components: {
        nginx: odin.#Component & {
            metadata: name: "nginx"

            config: {
                replicas: int | *1
                image: string
            }

            resources: deployment: {
                apiVersion: "apps/v1"
                kind: "Deployment"
                metadata: name: "nginx"
                spec: {
                    replicas: config.replicas
                    selector: matchLabels: app: "nginx"
                    template: {
                        metadata: labels: app: "nginx"
                        spec: containers: [{
                            name: "nginx"
                            image: config.image
                            ports: [{
                                containerPort: 80
                            }]
                        }]
                    }
                }
            }

            resources: service: {
                apiVersion: "v1"
                kind: "Service"
                metadata: name: "nginx"
                spec: {
                    selector: app: "nginx"
                    ports: [{
                        port: 80
                        targetPort: 80
                    }]
                }
            }
        }
    }

    values: {
        components: nginx: {
            image: "nginx:latest"
            replicas: 2
        }
    }
}
```

## Step 3: Generate Manifests

Now render the bundle to YAML:

```bash
odin template .
```

You should see output like:

```yaml
---
# Source: nginx/deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
---
# Source: nginx/service
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
```

## Step 4: Override Values

You can override values using a separate values file. Create `prod-values.cue`:

```cue
package main

values: {
    components: nginx: {
        image: "nginx:1.25"
        replicas: 3
    }
}
```

Then render with the overrides:

```bash
odin template . -f prod-values.cue
```

The output will now show 3 replicas and the specific nginx version.

## Step 5: Save to Files

To save the manifests to a file:

```bash
odin template . > manifests.yaml
```

Or apply directly to a cluster:

```bash
odin template . | kubectl apply -f -
```

## What's Next?

Congratulations! You've created your first Odin bundle. Now you can:

- [Learn more about bundles](../concepts/bundles.md)
- [Understand components and resources](../concepts/components.md)
- [Explore using component templates](../guides/using-templates.md)
- [Set up ArgoCD integration](../integration/argocd.md)

Let's dive deeper into [creating your first real bundle](./first-bundle.md)!
