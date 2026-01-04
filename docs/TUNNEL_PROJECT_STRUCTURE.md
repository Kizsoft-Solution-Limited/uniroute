# Tunnel Project Structure Decision

## Question: Separate Project or Same Project?

This document analyzes whether the custom tunnel solution should be:
1. **Same Project** - Part of UniRoute repository
2. **Separate Project** - Independent repository
3. **Monorepo** - Same repository, separate modules

## Option 1: Same Project (Recommended for MVP)

### Structure
```
uniroute/
├── cmd/
│   ├── gateway/          # Main gateway server
│   ├── cli/              # CLI tool
│   └── tunnel-server/    # Tunnel server (NEW)
├── internal/
│   ├── api/              # Gateway API
│   ├── gateway/          # Gateway routing
│   ├── tunnel/           # Tunnel client/server (NEW)
│   └── ...
└── ...
```

### Pros
✅ **Simpler Development**
- Single codebase to manage
- Shared utilities and packages
- Easier code reuse
- Unified testing

✅ **Easier Deployment (MVP)**
- Single binary option
- Shared configuration
- Unified logging
- Same deployment pipeline

✅ **Faster Development**
- No cross-repo dependencies
- Shared types and interfaces
- Easier refactoring
- Single CI/CD pipeline

✅ **Better for MVP**
- Get to market faster
- Simpler for initial users
- Less operational complexity

### Cons
❌ **Tighter Coupling**
- Changes affect both systems
- Harder to scale independently
- Shared dependencies

❌ **Deployment Complexity (Scale)**
- Can't scale tunnel servers independently
- Must deploy both together
- Resource allocation issues

❌ **Code Organization**
- Larger codebase
- More complex structure
- Harder to navigate

## Option 2: Separate Project

### Structure
```
uniroute/                 # Main gateway
├── cmd/gateway/
├── internal/
└── ...

uniroute-tunnel/          # Separate project
├── cmd/tunnel-server/
├── cmd/tunnel-client/
├── internal/
└── ...
```

### Pros
✅ **Independent Scaling**
- Scale tunnel servers separately
- Different resource requirements
- Independent deployments

✅ **Clear Separation**
- Distinct codebases
- Independent versioning
- Separate release cycles

✅ **Team Organization**
- Different teams can own each
- Independent development
- Clear ownership

✅ **Deployment Flexibility**
- Deploy tunnel servers globally
- Gateway can be local/regional
- Different infrastructure

### Cons
❌ **Development Complexity**
- Cross-repo dependencies
- Version management
- Shared code duplication
- More complex testing

❌ **Operational Overhead**
- Multiple deployments
- Separate CI/CD pipelines
- More infrastructure
- Coordination needed

❌ **Slower Initial Development**
- More setup time
- Cross-repo coordination
- More complex for MVP

## Option 3: Monorepo (Best of Both Worlds)

### Structure
```
uniroute/
├── cmd/
│   ├── gateway/
│   ├── cli/
│   └── tunnel-server/
├── internal/
│   ├── gateway/          # Gateway code
│   ├── tunnel/           # Tunnel code
│   └── shared/           # Shared utilities
├── pkg/
│   └── common/           # Shared packages
└── ...
```

### Pros
✅ **Code Sharing**
- Shared utilities
- Common types
- Unified testing
- Single repository

✅ **Independent Deployment**
- Separate binaries
- Independent scaling
- Different release cycles
- Flexible deployment

✅ **Development Benefits**
- Single codebase
- Easier refactoring
- Unified CI/CD
- Shared tooling

### Cons
❌ **Repository Size**
- Larger repository
- More complex structure
- Slower clones

❌ **Build Complexity**
- Multiple build targets
- More complex CI/CD
- Dependency management

## Recommendation: Phased Approach

### Phase 1: Same Project (MVP) ✅

**Start with tunnel in the same project** for:
- Faster development
- Simpler deployment
- Easier testing
- Quicker to market

**Structure:**
```
uniroute/
├── cmd/
│   ├── gateway/          # Main gateway
│   ├── cli/              # CLI (includes tunnel client)
│   └── tunnel-server/    # Tunnel server (optional, can run separately)
├── internal/
│   ├── tunnel/           # Tunnel client & server code
│   └── ...
```

**Benefits:**
- Single codebase
- Shared utilities
- Unified deployment
- Faster iteration

### Phase 2: Separate Deployment (Scale)

**When you need to scale**, deploy tunnel server separately:

```bash
# Deploy gateway
./bin/uniroute-gateway

# Deploy tunnel server (same binary, different config)
./bin/uniroute-tunnel-server
```

**Or separate binaries:**
```bash
# Build both
go build -o bin/gateway cmd/gateway/main.go
go build -o bin/tunnel-server cmd/tunnel-server/main.go
```

### Phase 3: Separate Project (If Needed)

**Only if you need:**
- Completely independent scaling
- Different teams
- Different release cycles
- Global tunnel infrastructure

## Recommended Structure (Same Project)

```
uniroute/
├── cmd/
│   ├── gateway/              # Main gateway server
│   │   └── main.go
│   ├── cli/                   # CLI tool
│   │   ├── main.go
│   │   └── commands/
│   │       └── tunnel.go      # Tunnel client command
│   └── tunnel-server/         # Tunnel server (optional separate binary)
│       └── main.go
├── internal/
│   ├── tunnel/
│   │   ├── server.go          # Tunnel server implementation
│   │   ├── client.go          # Tunnel client implementation
│   │   ├── protocol.go        # WebSocket protocol
│   │   └── manager.go         # Tunnel management
│   ├── gateway/               # Gateway code
│   ├── api/                   # API handlers
│   └── ...
├── pkg/
│   └── tunnel/                # Shared tunnel utilities (if needed)
│       └── types.go
└── ...
```

## Deployment Options

### Option A: Single Binary (Simple)
```bash
# Gateway with built-in tunnel server
./bin/uniroute-gateway --enable-tunnel-server
```

### Option B: Separate Binaries (Flexible)
```bash
# Gateway
./bin/uniroute-gateway

# Tunnel server (separate)
./bin/uniroute-tunnel-server
```

### Option C: Docker Compose (Production)
```yaml
services:
  gateway:
    build: .
    command: ./bin/uniroute-gateway
    
  tunnel-server:
    build: .
    command: ./bin/uniroute-tunnel-server
    scale: 3  # Scale independently
```

## Decision Matrix

| Factor | Same Project | Separate Project | Monorepo |
|--------|-------------|------------------|----------|
| **Development Speed** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Deployment Flexibility** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Code Sharing** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Scaling** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Complexity** | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| **MVP Suitability** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ |

## Final Recommendation

### ✅ **Start: Same Project**

**Reasons:**
1. Faster development for MVP
2. Easier code sharing
3. Simpler deployment initially
4. Can separate later if needed

**Structure:**
- Tunnel code in `internal/tunnel/`
- Tunnel server in `cmd/tunnel-server/` (optional)
- Tunnel client in CLI (`cmd/cli/commands/tunnel.go`)
- Shared utilities in `internal/`

### 🔄 **Later: Separate if Needed**

**When to separate:**
- Need independent scaling
- Different teams
- Global tunnel infrastructure
- Different release cycles

**Migration path:**
- Extract to separate repo
- Use shared packages
- Independent deployments

## Implementation Plan

### Phase 1: Same Project (Current)
- ✅ Tunnel code in `internal/tunnel/`
- ✅ CLI command in `cmd/cli/commands/tunnel.go`
- ✅ Optional tunnel server in `cmd/tunnel-server/`

### Phase 2: Flexible Deployment
- Build separate binaries
- Deploy independently
- Share code via packages

### Phase 3: Separate if Needed
- Extract to separate repo
- Use Go modules for dependencies
- Independent versioning

## Conclusion

**Start with tunnel in the same project** for faster development and simpler deployment. You can always separate later if scaling requirements demand it.

The current structure (`internal/tunnel/`) is perfect for this approach!

