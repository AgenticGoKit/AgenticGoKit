# API Versioning Strategy

**How AgenticGoKit manages versions, stability, and breaking changes.**

---

## Version Structure

AgenticGoKit follows semantic versioning with the following structure:

```
v<major>.<minor>.<patch>[-<prerelease>]

Examples:
- v1.0.0        (stable release)
- v1.2.3        (stable with patches)
- v1beta        (beta version)
- v2alpha       (alpha version)
```

### Version Stages

| Stage | Stability | Breaking Changes | Production Use |
|-------|-----------|------------------|----------------|
| **alpha** | Experimental | Frequent | Not recommended |
| **beta** | Stabilizing | Possible | Acceptable |
| **stable** | Production-ready | Rare, with migration path | Recommended |

---

## Current Versions

### v1beta (Current Recommended)

**Status:** Beta - Production-acceptable

**Import Path:**
```go
import "github.com/agenticgokit/agenticgokit/v1beta"
```

**Stability Guarantee:**
- API surface is mostly stable
- Minor breaking changes possible with deprecation notices
- All breaking changes documented in release notes
- Migration guides provided for breaking changes
- Security and bug fixes backported

**Lifecycle:**
- Current: Active development and support
- After v1.0 release: 6 months maintenance
- After maintenance: Legacy status (security fixes only)

### core/vnext (Legacy)

**Status:** Deprecated

**Import Path:**
```go
import "github.com/agenticgokit/agenticgokit/core/vnext"
```

**Stability Guarantee:**
- No new features
- Critical bug fixes only
- Security updates until EOL

**Lifecycle:**
- Current: Maintenance mode
- Support ends: 6 months after v1.0 release
- After EOL: No updates

**Migration:** See [Migration Guide](MIGRATION.md)

---

## Breaking Changes Policy

### What is a Breaking Change?

A breaking change is any modification that requires users to update their code:

- Removing public APIs
- Changing function signatures
- Renaming types or functions
- Changing behavior of existing APIs
- Changing import paths

### How We Handle Breaking Changes

#### Beta Versions (v1beta)

**Process:**
1. **Announcement** - Breaking change announced in GitHub discussions
2. **Deprecation Period** - Old API marked deprecated (minimum 1 release cycle)
3. **Migration Guide** - Step-by-step guide provided
4. **Release** - Breaking change shipped with clear release notes

**Timeline:**
```
Release N:   Deprecation notice + new API introduced
Release N+1: Old API still works with warnings
Release N+2: Old API removed (breaking change)
```

#### Stable Versions (v1.x)

**Process:**
1. **Avoid if possible** - Breaking changes avoided in stable versions
2. **Major version bump** - Breaking changes only in major versions (v1 → v2)
3. **Long deprecation** - Minimum 6-month deprecation period
4. **Parallel support** - Old and new APIs coexist during transition

**Example:**
```
v1.0: Current API
v1.5: New API introduced, old API deprecated
v1.x: Both APIs work for 6+ months
v2.0: Old API removed
```

---

## Deprecation Process

### Marking APIs as Deprecated

```go
// Deprecated: Use NewBuilder instead.
// This function will be removed in v2.0.
func CreateAgent(name string) *Agent {
    // Implementation
}

// NewBuilder creates an agent using the builder pattern.
// Recommended replacement for deprecated CreateAgent.
func NewBuilder(name string) *Builder {
    // Implementation
}
```

### Deprecation Notice Format

```go
// Deprecated: [Reason]
// Use [Replacement] instead.
// This will be removed in [Version].
```

### User Communication

Deprecations are communicated via:

1. **Code Comments** - GoDoc deprecation notices
2. **Release Notes** - Explicit mention in changelog
3. **GitHub Discussions** - Advance notice thread
4. **Migration Guides** - Detailed migration instructions
5. **Runtime Warnings** - Log warnings for deprecated APIs (where feasible)

---

## Compatibility Guarantees

### v1beta Guarantees

✅ **What We Promise:**
- Core functionality remains stable
- Bug fixes don't break existing code
- Performance improvements are transparent
- Security patches are non-breaking when possible

⚠️ **What May Change:**
- New features may be added
- Minor API adjustments with deprecation notices
- Error messages and types may be enhanced
- Internal implementation may change

❌ **No Guarantee:**
- Experimental features (marked explicitly)
- Internal packages (not in public API)
- Undocumented behavior

### v1.0+ Guarantees

✅ **What We Promise:**
- Public API is stable across minor versions (v1.x)
- No breaking changes within major version
- Security and bug fixes for 2+ years after major release
- Clear migration path for major version upgrades

---

## Version Support Timeline

```
v1beta:    Active (Current)
           ↓
v1.0:      Stable Release
           ├─ v1.x: Active development (24+ months)
           ├─ Security fixes (12+ months after v2.0)
           └─ EOL
           
v1beta:    Maintenance (after v1.0)
           ├─ Critical bugs (6 months)
           ├─ Security fixes (6 months)
           └─ EOL (6 months after v1.0)
```

### Support Phases

| Phase | Duration | Updates | Example |
|-------|----------|---------|---------|
| **Active** | Indefinite | Features, bugs, security | v1beta (current) |
| **Maintenance** | 6 months | Critical bugs, security | core/vnext |
| **Security** | 6 months | Security only | - |
| **EOL** | Forever | None | - |

---

## Choosing a Version

### For New Projects

**Recommendation: v1beta**

```go
import "github.com/agenticgokit/agenticgokit/v1beta"
```

**Reasons:**
- Modern, clean API
- Active development
- Best documentation
- Production-acceptable stability
- Will become v1.0 (seamless upgrade)

### For Existing Projects (Legacy)

**Option 1: Migrate to v1beta (Recommended)**

See [Migration Guide](MIGRATION.md) for step-by-step instructions.

**Benefits:**
- Modern features
- Better performance
- Long-term support
- Continued updates

**Option 2: Stay on Legacy**

Acceptable if:
- Project is in maintenance mode
- No new features needed
- Will be decommissioned within 6 months

---

## API Stability Markers

### Stable

```go
// Stable: This API is production-ready and will not change
// in breaking ways within the current major version.
func NewBuilder(name string) *Builder
```

### Beta

```go
// Beta: This API is mostly stable but may have minor changes
// with deprecation notices. Production-acceptable.
func WithExperimentalFeature() *Builder
```

### Experimental

```go
// Experimental: This API is under active development and may
// change significantly. Not recommended for production use.
func ExperimentalFeature() error
```

### Internal

```go
// Internal: This package is for internal use only and may
// change without notice. Do not import.
package internal
```

---

## Release Process

### Version Bumping

**Patch Release (v1.0.x):**
- Bug fixes
- Security patches
- Documentation improvements
- No API changes

**Minor Release (v1.x.0):**
- New features (backward compatible)
- Deprecations (with notices)
- Performance improvements
- API additions (no removals)

**Major Release (vX.0.0):**
- Breaking changes
- API removals (after deprecation)
- Major architectural changes
- Requires migration

### Release Cadence

- **Patch releases:** As needed (bug fixes)
- **Minor releases:** Monthly or as needed
- **Major releases:** Yearly or as needed

---

## Backward Compatibility

### What We Maintain

✅ **Guaranteed Compatibility:**
- Function signatures
- Type definitions
- Package imports
- Documented behavior
- Configuration formats

⚠️ **Best Effort:**
- Error messages
- Log output
- Performance characteristics
- Internal implementation

❌ **No Guarantee:**
- Undocumented features
- Internal packages
- Test helpers
- Examples (may be updated)

### Testing Compatibility

We test backward compatibility via:

1. **Regression Tests** - Existing tests must pass
2. **API Compatibility Checker** - Tools detect breaking changes
3. **Integration Tests** - Real-world usage patterns validated
4. **Community Feedback** - Beta releases for early testing

---

## Experimental Features

### Definition

Experimental features are:
- New functionality under evaluation
- May change based on feedback
- Not subject to normal stability guarantees
- Clearly marked in documentation

### Usage Guidelines

**Acceptable for experimentation:**
```go
// Try new features in development
agent, _ := v1beta.NewBuilder("Agent").
    WithExperimentalFeature(). // OK for testing
    Build()
```

**Not recommended for production:**
```go
// Avoid in production systems
agent, _ := v1beta.NewBuilder("ProductionAgent").
    WithExperimentalFeature(). // Risky - may change
    Build()
```

### Graduation Process

Experimental → Beta → Stable

1. **Experimental:** Initial release, gather feedback
2. **Beta:** API stabilized, minor changes possible
3. **Stable:** API locked, stability guaranteed

---

## Migration Support

### Resources

- **[Migration Guide](MIGRATION.md)** - Complete migration instructions
- **[Migration Tutorials](tutorials/v1beta-migration/)** - Step-by-step guides
- **[GitHub Discussions](https://github.com/agenticgokit/agenticgokit/discussions)** - Community help
- **[Release Notes](../RELEASE.md)** - Detailed changelogs

### Migration Timeline

**Standard Timeline:**
```
Deprecation Notice → 6 months → Removal
```

**Emergency Timeline (security):**
```
Security Issue → Patch Release → 3 months → Removal
```

---

## Security Updates

### Policy

- Security issues receive immediate attention
- Patches released for supported versions
- Coordinated disclosure process
- Security advisories published

### Supported Versions for Security Patches

| Version | Security Support |
|---------|------------------|
| v1beta (current) | ✅ Active |
| core/vnext | ✅ Until v1.0 + 6 months |
| v1.0+ | ✅ 12 months after next major |

### Reporting Security Issues

**Do not open public issues for security vulnerabilities.**

Email: security@agenticgokit.dev

---

## Questions & Support

- **Version Questions:** [GitHub Discussions](https://github.com/agenticgokit/agenticgokit/discussions)
- **Bug Reports:** [GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues)
- **Feature Requests:** [GitHub Discussions](https://github.com/agenticgokit/agenticgokit/discussions)
- **Migration Help:** [Migration Guide](MIGRATION.md)

---

## Next Steps

- **[Getting Started](v1beta/getting-started.md)** - Start building with v1beta
- **[Migration Guide](MIGRATION.md)** - Migrate from legacy APIs
- **[Roadmap](ROADMAP.md)** - Planned features and timeline
