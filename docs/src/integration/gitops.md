# GitOps Workflows

This guide covers GitOps workflows with Odin bundles.

## What is GitOps?

GitOps uses Git as the single source of truth for declarative infrastructure and applications. Changes to infrastructure are made via Git commits, and automated processes ensure the cluster matches the desired state in Git.

## Repository Structure

### Single Application Bundle

```
my-app/
├── bundle.cue
├── components/
├── environments/
│   ├── dev.cue
│   ├── staging.cue
│   └── prod.cue
├── manifests/           # Rendered manifests (optional)
│   ├── dev.yaml
│   ├── staging.yaml
│   └── prod.yaml
└── .argocd/
    └── applications/
        ├── dev.yaml
        ├── staging.yaml
        └── prod.yaml
```

### Multi-Application Repository

```
platform/
├── apps/
│   ├── frontend/
│   │   ├── bundle.cue
│   │   └── environments/
│   ├── backend/
│   │   ├── bundle.cue
│   │   └── environments/
│   └── worker/
│       ├── bundle.cue
│       └── environments/
└── .argocd/
    └── applications/
        ├── frontend-prod.yaml
        ├── backend-prod.yaml
        └── worker-prod.yaml
```

## Workflow Patterns

### Pattern 1: Direct Rendering (ArgoCD Plugin)

ArgoCD renders manifests on-the-fly using the Odin plugin.

**Advantages:**
- No committed manifests
- Single source of truth (CUE files)
- Automatic rendering on sync

**Disadvantages:**
- Requires ArgoCD plugin
- Can't preview manifests in PR
- Harder to debug rendering issues

**Application:**

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

### Pattern 2: Pre-rendered Manifests

Render manifests in CI and commit to Git.

**Advantages:**
- Works with any GitOps tool
- Manifest preview in PRs
- Easy to debug
- Audit trail of all changes

**Disadvantages:**
- Duplicate files (CUE + YAML)
- Need CI to keep in sync

**CI Script:**

```bash
#!/bin/bash
# .github/workflows/render.yml

set -euo pipefail

for env in dev staging prod; do
    echo "Rendering $env..."
    odin template . -f "environments/${env}.cue" > "manifests/${env}.yaml"
done

# Commit if changed
if ! git diff --quiet manifests/; then
    git config user.name "GitHub Actions"
    git config user.email "actions@github.com"
    git add manifests/
    git commit -m "chore: update rendered manifests"
    git push
fi
```

**ArgoCD Application:**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  source:
    path: manifests
    targetRevision: main
```

### Pattern 3: Hybrid (PR Preview + Direct Rendering)

Render in CI for PR preview, but use plugin in ArgoCD.

**CI for PRs:**

```yaml
name: Preview Manifests
on: pull_request

jobs:
  render:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Render manifests
      run: |
        for env in dev staging prod; do
          odin template . -f "environments/${env}.cue" > "/tmp/${env}.yaml"
        done
    - name: Comment PR
      uses: actions/github-script@v7
      with:
        script: |
          const fs = require('fs');
          const dev = fs.readFileSync('/tmp/dev.yaml', 'utf8');
          const body = `## Rendered Manifests\n\n<details><summary>Development</summary>\n\n\`\`\`yaml\n${dev}\n\`\`\`\n</details>`;
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body
          });
```

## Development Workflow

### Local Development

```bash
# 1. Create feature branch
git checkout -b feature/add-cache

# 2. Edit bundle
vim components/redis.cue

# 3. Test locally
odin template . -f environments/dev.cue | kubectl apply --dry-run=client -f -

# 4. Apply to dev cluster (optional)
odin template . -f environments/dev.cue | kubectl apply -f -

# 5. Commit changes
git add .
git commit -m "feat: add Redis cache"
git push origin feature/add-cache
```

### Pull Request Review

```bash
# 1. Reviewer checks out PR
git fetch origin pull/123/head:pr-123
git checkout pr-123

# 2. Review CUE changes
git diff main bundle.cue components/

# 3. Render and review manifests
odin template . -f environments/prod.cue > /tmp/prod.yaml
kubectl diff -f /tmp/prod.yaml  # Compare with cluster

# 4. Approve or request changes
```

### Deployment

```bash
# 1. Merge PR
git checkout main
git merge feature/add-cache

# 2. Tag release (optional)
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0

# 3. ArgoCD automatically syncs
# Or manually sync:
argocd app sync my-app-prod
```

## Environment Promotion

### Manual Promotion

```bash
# 1. Test in dev
odin template . -f environments/dev.cue | kubectl apply -f -

# 2. Promote to staging
git checkout -b promote/v1.2.0-staging
# Update staging values to match tested config
vim environments/staging.cue
git commit -am "chore: promote v1.2.0 to staging"

# 3. After testing, promote to prod
git checkout -b promote/v1.2.0-prod
vim environments/prod.cue
git commit -am "chore: promote v1.2.0 to prod"
```

### Automatic Promotion

Use CI to promote after tests pass:

```yaml
name: Promote to Production
on:
  push:
    tags:
      - 'v*'

jobs:
  promote:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Update prod values
      run: |
        TAG=${GITHUB_REF#refs/tags/}
        # Update image tag in prod values
        sed -i "s/tag: \".*\"/tag: \"$TAG\"/" environments/prod.cue
    
    - name: Commit changes
      run: |
        git config user.name "GitHub Actions"
        git config user.email "actions@github.com"
        git add environments/prod.cue
        git commit -m "chore: promote $TAG to production"
        git push
```

## Rollback

### Via Git

```bash
# 1. Find previous working commit
git log --oneline environments/prod.cue

# 2. Revert to previous version
git revert HEAD
# Or cherry-pick specific change
git cherry-pick abc123

# 3. Push
git push origin main

# 4. ArgoCD syncs automatically
```

### Via ArgoCD

```bash
# Rollback to previous sync
argocd app rollback my-app-prod

# Rollback to specific revision
argocd app rollback my-app-prod 5
```

## Multi-Cluster Deployment

### Cluster-Specific Values

```
environments/
├── dev.cue
├── staging.cue
├── prod-us-east.cue
└── prod-eu-west.cue
```

### ApplicationSet for Multi-Cluster

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-app-prod
spec:
  generators:
  - list:
      elements:
      - cluster: prod-us-east
        server: https://us-east.k8s.example.com
        valuesFile: environments/prod-us-east.cue
      - cluster: prod-eu-west
        server: https://eu-west.k8s.example.com
        valuesFile: environments/prod-eu-west.cue
  template:
    metadata:
      name: 'my-app-{{cluster}}'
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
        server: '{{server}}'
        namespace: my-app
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

## CI/CD Integration

### GitHub Actions

```yaml
name: CI
on:
  pull_request:
  push:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Install Odin
      run: |
        curl -LO https://github.com/go-valkyrie/odin/releases/latest/download/odin-linux-amd64
        chmod +x odin-linux-amd64
        sudo mv odin-linux-amd64 /usr/local/bin/odin
    
    - name: Validate bundle
      run: |
        for env in dev staging prod; do
          echo "Validating $env..."
          odin template . -f "environments/${env}.cue" > /dev/null
        done
    
    - name: Render manifests (on main)
      if: github.ref == 'refs/heads/main'
      run: |
        for env in dev staging prod; do
          odin template . -f "environments/${env}.cue" > "manifests/${env}.yaml"
        done
    
    - name: Commit manifests
      if: github.ref == 'refs/heads/main'
      run: |
        git config user.name "GitHub Actions"
        git config user.email "actions@github.com"
        git add manifests/
        if ! git diff --quiet --staged; then
          git commit -m "chore: update rendered manifests [skip ci]"
          git push
        fi
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - render

validate:
  stage: validate
  image: ghcr.io/go-valkyrie/odin:latest
  script:
    - |
      for env in dev staging prod; do
        echo "Validating $env..."
        odin template . -f "environments/${env}.cue" > /dev/null
      done

render:
  stage: render
  image: ghcr.io/go-valkyrie/odin:latest
  only:
    - main
  script:
    - |
      for env in dev staging prod; do
        odin template . -f "environments/${env}.cue" > "manifests/${env}.yaml"
      done
    - |
      git config user.name "GitLab CI"
      git config user.email "ci@gitlab.com"
      git add manifests/
      if ! git diff --quiet --staged; then
        git commit -m "chore: update rendered manifests [skip ci]"
        git push https://oauth2:${CI_JOB_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.git HEAD:${CI_COMMIT_REF_NAME}
      fi
```

## Best Practices

### 1. One Bundle Per Application

Each logical application gets its own bundle and repository (or directory).

### 2. Environment Branches (Optional)

Some teams use branches for environments:

```
main      → development
staging   → staging environment
prod      → production environment
```

But environment-specific values files are usually simpler.

### 3. Automated Testing

```yaml
- name: Test manifests
  run: |
    # Apply with dry-run
    odin template . -f environments/prod.cue | \
      kubectl apply --dry-run=server -f -
    
    # Validate with kubeval
    odin template . -f environments/prod.cue | \
      kubeval --strict
    
    # Policy checks with OPA/Gatekeeper
    odin template . -f environments/prod.cue | \
      conftest test -p policy.rego -
```

### 4. Clear Commit Messages

Use conventional commits:

```
feat: add Redis cache component
fix: correct database connection string
chore: update webapp to v1.2.0
docs: add deployment instructions
```

### 5. Protected Branches

Protect `main` and require:
- PR reviews
- Passing CI checks
- Up-to-date with base branch

## Next Steps

- See [ArgoCD Integration](./argocd.md) for setup details
- Learn [Best Practices](../reference/best-practices.md)
- Explore [Bundle Structure](../guides/bundle-structure.md)
