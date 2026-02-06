# Troubleshooting

This guide covers common issues and solutions when using Odin.

## Installation Issues

### Binary Not Found

**Problem:**

```bash
odin: command not found
```

**Solutions:**

1. Check if binary is in PATH:

```bash
which odin
```

2. If not, add to PATH or move to a directory in PATH:

```bash
sudo mv odin /usr/local/bin/
```

3. Ensure binary is executable:

```bash
chmod +x odin
```

### Permission Denied

**Problem:**

```bash
bash: ./odin: Permission denied
```

**Solution:**

```bash
chmod +x odin
```

## Bundle Issues

### Bundle Not Found

**Problem:**

```
Error: bundle not found at path "."
```

**Solutions:**

1. Ensure you're in a directory with `bundle.cue`:

```bash
ls bundle.cue
```

2. Or specify the correct path:

```bash
odin template /path/to/bundle
```

3. Check if auto-discovery is working:

```bash
# Should work from any subdirectory
cd components/
odin template .
```

### CUE Syntax Errors

**Problem:**

```
Error: invalid CUE syntax
```

**Solutions:**

1. Validate syntax with CUE:

```bash
cue vet bundle.cue
```

2. Check for common mistakes:
   - Missing commas
   - Unmatched braces
   - Invalid field names

3. Use CUE's formatting:

```bash
cue fmt bundle.cue
```

### Module Not Found

**Problem:**

```
Error: package "go-valkyrie.com/odin/api/v1alpha1" not found
```

**Solutions:**

1. Check `cue.mod/module.cue` has the dependency:

```cue
deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
}
```

2. Run module tidy:

```bash
odin cue mod tidy
```

3. Clear cache and retry:

```bash
odin cache clean
odin cue mod tidy
```

## Template Issues

### Field Not Found

**Problem:**

```
Error: field "replicas" not defined in struct
```

**Solutions:**

1. Check template documentation:

```bash
odin docs webapp
```

2. Verify field exists in config schema

3. Check spelling and case

### Type Mismatch

**Problem:**

```
Error: conflicting values string and int
```

**Solutions:**

1. Check type in component config:

```cue
config: {
    replicas: int  // Not string
}
```

2. Ensure values file has correct type:

```cue
values: {
    components: webapp: {
        replicas: 3  // Not "3"
    }
}
```

### Constraint Violation

**Problem:**

```
Error: invalid value 15 (exceeds maximum 10)
```

**Solution:**

Adjust value to meet constraints:

```cue
config: {
    replicas: int & >=1 & <=10  // Maximum is 10
}

values: {
    components: webapp: {
        replicas: 10  // Was 15, changed to 10
    }
}
```

## Registry Issues

### Registry Not Found

**Problem:**

```
Error: failed to fetch module: 404 Not Found
```

**Solutions:**

1. Check registry configuration:

```bash
odin config eval
```

2. Verify module exists in registry:

```bash
# Check with docker/podman
docker pull ghcr.io/myorg/cue/platform.example.com/webapp/v1alpha1:v0.1.0
```

3. Check registry URL is correct in config

### Authentication Failed

**Problem:**

```
Error: failed to fetch module: 401 Unauthorized
```

**Solutions:**

1. Login to registry:

```bash
echo $TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

2. Check credentials are valid:

```bash
docker logout ghcr.io
docker login ghcr.io
```

3. Set credentials path:

```bash
export CUE_REGISTRY_AUTH=~/.docker/config.json
```

### Registry Timeout

**Problem:**

```
Error: failed to fetch module: connection timeout
```

**Solutions:**

1. Check network connectivity:

```bash
ping ghcr.io
```

2. Check firewall/proxy settings

3. Try with increased timeout

## Values Issues

### Values Not Applied

**Problem:**

Values file isn't being used.

**Solutions:**

1. Check file path is correct:

```bash
odin template . -f environments/prod.cue
```

2. Verify package name matches:

```cue
// Both files must have same package
package main
```

3. Check values are in correct structure:

```cue
values: {
    components: {
        webapp: {
            // Values here
        }
    }
}
```

### Values Override Not Working

**Problem:**

Later values file not overriding earlier one.

**Solution:**

Ensure values are at same depth:

```cue
// base.cue
values: {
    components: webapp: {
        replicas: 1
    }
}

// prod.cue - this overrides
values: {
    components: webapp: {
        replicas: 5
    }
}

// prod.cue - this DOESN'T override
values: {
    components: {
        replicas: 5  // Wrong depth
    }
}
```

## Cache Issues

### Stale Cache

**Problem:**

Odin using old version of module.

**Solution:**

```bash
odin cache clean
odin template .
```

### Cache Corruption

**Problem:**

```
Error: failed to load cached module
```

**Solution:**

```bash
# Clear cache
odin cache clean

# Verify cache directory
odin config eval | grep cacheDir

# Check permissions
ls -la $(odin config eval | grep cacheDir | awk '{print $2}')
```

### Disk Space

**Problem:**

Cache filling disk.

**Solutions:**

1. Check cache size:

```bash
du -sh $(odin config eval | grep cacheDir | awk '{print $2}')
```

2. Clean cache:

```bash
odin cache clean
```

3. Use custom cache location:

```bash
export XDG_CACHE_HOME=/path/with/more/space
```

## ArgoCD Issues

### Plugin Not Found

**Problem:**

```
Error: plugin 'odin' not found
```

**Solutions:**

1. Verify plugin is registered:

```bash
kubectl get cm argocd-cm -n argocd -o yaml
```

2. Check sidecar is running:

```bash
kubectl get pods -n argocd -l app.kubernetes.io/name=argocd-repo-server
kubectl logs -n argocd <pod-name> -c odin
```

3. Restart repo server:

```bash
kubectl rollout restart deployment argocd-repo-server -n argocd
```

### Bundle Not Discovered

**Problem:**

ArgoCD doesn't detect Odin bundle.

**Solutions:**

1. Ensure `bundle.cue` exists:

```bash
ls bundle.cue
```

2. Check discovery pattern in plugin config:

```yaml
discover:
  fileName: "./bundle.cue"
```

3. Verify file is in repository root (or specified path)

### Generate Failed

**Problem:**

```
Error: failed to generate manifests
```

**Solutions:**

1. Check repo server logs:

```bash
kubectl logs -n argocd <repo-server-pod> -c odin
```

2. Test locally:

```bash
odin template . -f environments/prod.cue
```

3. Check environment variables are set in plugin

## Performance Issues

### Slow Template Rendering

**Problem:**

`odin template` takes a long time.

**Solutions:**

1. Use debug logging to identify bottleneck:

```bash
odin --debug template .
```

2. Check if network is slow (fetching modules)

3. Clean and rebuild cache:

```bash
odin cache clean
```

4. Split large bundle into smaller ones

### High Memory Usage

**Problem:**

Odin using too much memory.

**Solutions:**

1. Reduce bundle complexity

2. Avoid circular references

3. Use definitions to deduplicate code

## Debug Commands

### View Configuration

```bash
# View all configuration
odin config eval

# View cache location
odin config eval | grep cacheDir

# View registries
odin config eval | grep -A 10 registries
```

### Test Bundle Locally

```bash
# Validate syntax
cue vet bundle.cue

# Test template rendering
odin template . > /tmp/test.yaml

# Validate with Kubernetes
odin template . | kubectl apply --dry-run=client -f -

# Check with server validation
odin template . | kubectl apply --dry-run=server -f -
```

### Test Module Access

```bash
# Test module fetch
odin cue mod get platform.example.com/webapp/v1alpha1@v0.1.0

# List available components
odin components

# View component docs
odin docs webapp
```

### Enable Debug Logging

```bash
# Debug logging
odin --debug template .

# Verbose output
odin -v template .

# Both
odin --debug -v template .
```

## Getting Help

### Check Documentation

1. View command help:

```bash
odin template --help
```

2. Check this documentation

3. Review examples in `examples/` directories

### Report Issues

If you find a bug:

1. Check existing issues: https://github.com/go-valkyrie/odin/issues

2. Create new issue with:
   - Odin version: `odin version` (if available)
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Debug output: `odin --debug template .`

3. Include minimal reproduction case

## Common Patterns That Work

### Validate Before Commit

```bash
#!/bin/bash
# pre-commit hook

for env in dev staging prod; do
    odin template . -f "environments/${env}.cue" > /dev/null || exit 1
done
```

### Test Multiple Environments

```bash
#!/bin/bash
# test.sh

for env in dev staging prod; do
    echo "Testing $env..."
    odin template . -f "environments/${env}.cue" | \
        kubectl apply --dry-run=client -f - > /dev/null
done
```

### Clean Build

```bash
#!/bin/bash
# clean-build.sh

odin cache clean
odin cue mod tidy
odin template . -f environments/prod.cue
```

## Next Steps

- Review [Best Practices](./best-practices.md)
- Check [Configuration](../guides/configuration.md)
- See [API Schema](./api-schema.md)
