# Release Process

This document describes how to create new releases for UniRoute when there are updates to the tunnel CLI or any other components.

## Quick Start

Use the automated release script:

```bash
./scripts/create-release.sh
```

The script will:
- Show current version
- Ask for release type (patch/minor/major/custom)
- Create and push the tag
- Trigger GitHub Actions workflow

## Manual Process

### 1. Ensure Everything is Committed

```bash
git status
git add .
git commit -m "feat: your changes description"
git push
```

### 2. Create Release Tag

Follow [Semantic Versioning](https://semver.org/):
- **Patch** (v1.0.1): Bug fixes, small changes
- **Minor** (v1.1.0): New features, backward compatible
- **Major** (v2.0.0): Breaking changes

```bash
# Patch release
git tag -a v1.0.1 -m "Release v1.0.1 - Bug fixes and improvements"

# Minor release
git tag -a v1.1.0 -m "Release v1.1.0 - New features"

# Major release
git tag -a v2.0.0 -m "Release v2.0.0 - Breaking changes"
```

### 3. Push Tag

```bash
git push origin v1.0.1
```

### 4. Monitor Workflow

The GitHub Actions workflow will automatically:
1. Build CLI binaries for all platforms
2. Create GitHub release
3. Attach binaries to release

Monitor at: https://github.com/Kizsoft-Solution-Limited/uniroute/actions

## What Happens After Release

Once the workflow completes:

1. **GitHub Release Created**: Available at `/releases/latest`
2. **Download Links Work**: All platform binaries are available
3. **Version Checker Updated**: The tunnel CLI will detect the new version
4. **Users Get Notified**: When users run `uniroute tunnel`, they'll see:
   ```
   📦 Update available: v1.0.1 (current: v1.0.0)
   Press Ctrl+U to upgrade
   ```

## Version Numbering

### Current Version
Check the current version:
```bash
git describe --tags --abbrev=0
```

### Update Version in Code

The version is defined in:
- `cmd/cli/commands/root.go`: `version = "1.0.0"`

**Important**: Update this before creating a new release tag!

```go
var (
    version = "1.0.1"  // Update this
    rootCmd = &cobra.Command{
        // ...
    }
)
```

Then commit and push:
```bash
git add cmd/cli/commands/root.go
git commit -m "chore: bump version to 1.0.1"
git push
```

## Release Checklist

Before creating a release:

- [ ] All changes committed and pushed
- [ ] Version number updated in `cmd/cli/commands/root.go`
- [ ] Tests passing
- [ ] Documentation updated (if needed)
- [ ] Release notes prepared

## Examples

### Example 1: Bug Fix Release

```bash
# 1. Fix the bug and commit
git add .
git commit -m "fix: tunnel connection issue"
git push

# 2. Update version
# Edit cmd/cli/commands/root.go: version = "1.0.1"

# 3. Commit version bump
git add cmd/cli/commands/root.go
git commit -m "chore: bump version to 1.0.1"
git push

# 4. Create release
./scripts/create-release.sh
# Choose option 1 (Patch)
```

### Example 2: New Feature Release

```bash
# 1. Add feature and commit
git add .
git commit -m "feat: add new tunnel feature"
git push

# 2. Update version
# Edit cmd/cli/commands/root.go: version = "1.1.0"

# 3. Commit version bump
git add cmd/cli/commands/root.go
git commit -m "chore: bump version to 1.1.0"
git push

# 4. Create release
./scripts/create-release.sh
# Choose option 2 (Minor)
```

## Troubleshooting

### Workflow Failed

1. Check GitHub Actions logs: https://github.com/Kizsoft-Solution-Limited/uniroute/actions
2. Common issues:
   - Go version mismatch (should be 1.24)
   - Build errors
   - Permission issues

### Download Links Still 404

1. Wait for workflow to complete (usually 5-10 minutes)
2. Check release was created: https://github.com/Kizsoft-Solution-Limited/uniroute/releases
3. Verify binaries are attached to release

### Version Not Updating

1. Ensure version in `cmd/cli/commands/root.go` matches tag
2. Rebuild and test locally:
   ```bash
   go build -o uniroute ./cmd/cli
   ./uniroute --version
   ```

## Automated Release (Future)

For CI/CD, you could automate releases on merge to main:

```yaml
# .github/workflows/auto-release.yml
on:
  push:
    branches:
      - main
    paths:
      - 'cmd/cli/**'
      - 'internal/tunnel/**'
```

But for now, manual releases give you more control.

