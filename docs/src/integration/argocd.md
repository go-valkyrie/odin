# ArgoCD Integration

Odin provides a ConfigManagementPlugin for ArgoCD that enables GitOps workflows with CUE bundles.

## Quick Start

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/my-bundle
    targetRevision: main
    path: .
    plugin:
      name: odin
      parameters:
      - name: valuesFile
        string: environments/prod.cue
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

## Plugin Configuration

The Odin plugin is defined in `argocd/plugin.yml`:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ConfigManagementPlugin
metadata:
  name: odin
spec:
  version: v0.1.0
  init:
    command: [/home/argocd/bin/init.sh]
  generate:
    command: [/home/argocd/bin/generate.sh]
  discover:
    fileName: "./bundle.cue"
  parameters:
    static:
      - name: "valuesFile"
        itemType: string
      - name: "values"
        collectionType: map
```

### Discovery

The plugin automatically discovers Odin bundles by looking for `bundle.cue` files.

### Parameters

**valuesFile** - Path to a values file:

```yaml
plugin:
  name: odin
  parameters:
  - name: valuesFile
    string: environments/prod.cue
```

**values** - Inline values (as a map):

```yaml
plugin:
  name: odin
  parameters:
  - name: values
    map:
      components.webapp.replicas: "5"
      components.webapp.image.tag: "v1.0.0"
```

## Installation

### Option 1: Sidecar Container

Add the Odin plugin as a sidecar to the ArgoCD repo server:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cmp-odin
  namespace: argocd
data:
  plugin.yaml: |
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: odin
    spec:
      version: v0.1.0
      init:
        command: [sh, -c]
        args:
        - |
          # Init phase (optional)
          echo "Initializing Odin bundle..."
      generate:
        command: [sh, -c]
        args:
        - |
          # Generate manifests
          odin template . ${ARGOCD_ENV_valuesFile:+-f $ARGOCD_ENV_valuesFile}
      discover:
        fileName: "./bundle.cue"
      parameters:
        static:
          - name: "valuesFile"
            itemType: string
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
  namespace: argocd
spec:
  template:
    spec:
      volumes:
      - name: cmp-odin
        configMap:
          name: argocd-cmp-odin
      - name: cmp-tmp
        emptyDir: {}
      containers:
      - name: odin
        image: ghcr.io/go-valkyrie/odin:latest
        command: [/var/run/argocd/argocd-cmp-server]
        volumeMounts:
        - mountPath: /var/run/argocd
          name: var-files
        - mountPath: /home/argocd/cmp-server/config/plugin.yaml
          subPath: plugin.yaml
          name: cmp-odin
        - mountPath: /tmp
          name: cmp-tmp
```

### Option 2: Built-in Plugin

The Odin container image includes the plugin configuration. Mount it into the ArgoCD repo server:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      initContainers:
      - name: install-odin
        image: ghcr.io/go-valkyrie/odin:latest
        command: [sh, -c]
        args:
        - |
          cp /usr/local/bin/odin /custom-tools/
          cp -r /etc/odin/argocd/* /custom-tools/odin-plugin/
        volumeMounts:
        - mountPath: /custom-tools
          name: custom-tools
      volumes:
      - name: custom-tools
        emptyDir: {}
      containers:
      - name: argocd-repo-server
        volumeMounts:
        - mountPath: /usr/local/bin/odin
          name: custom-tools
          subPath: odin
        env:
        - name: PATH
          value: /usr/local/bin:/usr/bin:/bin
```

## Hook System

The Odin ArgoCD plugin supports extensibility via hooks:

### Init Hooks

Scripts in `/etc/odin/argocd/init.d/` run during the init phase:

```bash
#!/bin/bash
# /etc/odin/argocd/init.d/01-setup-credentials

# Example: Set up credentials
export CUE_REGISTRY_AUTH=/path/to/credentials
```

### Generate Hooks

Scripts in `/etc/odin/argocd/generate.d/` run during manifest generation:

```bash
#!/bin/bash
# /etc/odin/argocd/generate.d/01-validate

# Example: Validate bundle before generating
odin template . --dry-run
```

Hooks must be:
- Executable regular files
- Run in lexical order
- Exit 0 for success

## Environment Variables

The plugin sets up the following environment:

- `HOME` - Set correctly for CUE registry operations
- `PATH` - Includes Odin binary
- `ARGOCD_ENV_*` - ArgoCD parameter values

## Bundle Structure for ArgoCD

Recommended structure:

```
my-bundle/
├── bundle.cue
├── components/
├── environments/
│   ├── dev.cue
│   ├── staging.cue
│   └── prod.cue
└── .argocd/
    └── applications/
        ├── dev.yaml
        ├── staging.yaml
        └── prod.yaml
```

### Application Manifests

**.argocd/applications/prod.yaml:**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app-prod
  namespace: argocd
  finalizers:
  - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/my-bundle
    targetRevision: main
    path: .
    plugin:
      name: odin
      parameters:
      - name: valuesFile
        string: environments/prod.cue
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-prod
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
```

## Multi-Environment Setup

### App of Apps Pattern

Create a parent app that manages environment-specific apps:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/my-bundle
    targetRevision: main
    path: .argocd/applications
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### ApplicationSet

Use ApplicationSet for dynamic environment management:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-app
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - env: dev
        valuesFile: environments/dev.cue
      - env: staging
        valuesFile: environments/staging.cue
      - env: prod
        valuesFile: environments/prod.cue
  template:
    metadata:
      name: 'my-app-{{env}}'
    spec:
      project: default
      source:
        repoURL: https://github.com/myorg/my-bundle
        targetRevision: main
        path: .
        plugin:
          name: odin
          parameters:
          - name: valuesFile
            string: '{{valuesFile}}'
      destination:
        server: https://kubernetes.default.svc
        namespace: 'my-app-{{env}}'
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
        - CreateNamespace=true
```

## Troubleshooting

### Plugin Not Found

**Error:** `plugin 'odin' not found`

**Solution:** Ensure the plugin is properly installed in the repo server.

### Bundle Not Discovered

**Error:** Bundle not detected

**Solution:** Ensure `bundle.cue` exists at the specified path.

### Values File Not Found

**Error:** `failed to read values file`

**Solution:** Check the valuesFile path is relative to the bundle root.

### Registry Access

If using private registries, configure credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: odin-registry-credentials
  namespace: argocd
type: Opaque
data:
  config.json: <base64-encoded-docker-config>
---
# Mount in repo server
spec:
  template:
    spec:
      containers:
      - name: odin
        env:
        - name: CUE_REGISTRY_AUTH
          value: /etc/odin/registry-credentials/config.json
        volumeMounts:
        - name: registry-credentials
          mountPath: /etc/odin/registry-credentials
      volumes:
      - name: registry-credentials
        secret:
          secretName: odin-registry-credentials
```

## Best Practices

### 1. Use Separate Environments

Create separate Applications for each environment:

```
.argocd/applications/
├── dev.yaml
├── staging.yaml
└── prod.yaml
```

### 2. Version Lock in Production

```cue
// environments/prod.cue
package main

values: {
    components: {
        webapp: {
            image: tag: "v1.0.0"  // Pin specific version
        }
    }
}
```

### 3. Enable Auto-Sync Carefully

Development:
```yaml
syncPolicy:
  automated:
    prune: true
    selfHeal: true
```

Production:
```yaml
syncPolicy:
  automated:
    prune: false  # Manual approval for deletions
    selfHeal: true
```

### 4. Use Health Checks

Ensure your components define proper health checks so ArgoCD can track readiness.

## Next Steps

- See [GitOps Workflows](./gitops.md) for complete workflows
- Learn [Bundle Structure](../guides/bundle-structure.md) for organization
- Explore [Best Practices](../reference/best-practices.md)
