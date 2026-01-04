# 🚀 UniRoute - Complete Project Guide

**One unified gateway for every AI model. Route, secure, and manage traffic to any LLM—cloud or local—with one unified platform.**

**100% FREE • Open Source • Self-Hostable**

---

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Quick Start](#quick-start)
3. [Architecture](#architecture)
4. [Technology Stack](#technology-stack)
5. [Code Quality & Best Practices](#code-quality--best-practices)
6. [Security Requirements](#security-requirements)
7. [Implementation Plan](#implementation-plan)
8. [Cost Analysis](#cost-analysis)
9. [Competitive Analysis](#competitive-analysis)
10. [Features & Roadmap](#features--roadmap)
11. [Database Schema](#database-schema)
12. [Deployment](#deployment)
13. [API Design](#api-design)
14. [Testing](#testing)
15. [Documentation](#documentation)
16. [Pre-Launch Checklist](#pre-launch-checklist)
17. [Success Metrics](#success-metrics)
18. [Contributing](#contributing)
19. [Support & Resources](#support--resources)
20. [Next Steps](#next-steps)
21. [Quick Reference](#quick-reference)

---

## 🎯 Project Overview

### What is UniRoute?

**UniRoute** is a unified gateway platform that routes, secures, and manages traffic to any LLM (cloud or local) with one unified API. Think of it as **ngrok for AI models** - a single entry point that intelligently routes requests to the best available model.

**Inspired by [ngrok's AI Gateway](https://ngrok.com/)** - ngrok just launched "One gateway for every AI model" (available in early access at [ngrok.ai](https://ngrok.ai)), proving there's massive market demand for this solution.

### Project Details

- **Project Name**: UniRoute
- **Domain**: `api.uniroute.pages.dev` (Cloudflare Pages - Free)
- **Documentation**: `uniroute.pages.dev`
- **Status**: `status.uniroute.pages.dev`
- **License**: MIT (Open Source)
- **Pricing**: **100% FREE** - No pricing, no limits for users

### Our Mission

Make AI gateway technology accessible to everyone, not just enterprises with budgets.

### Competitive Advantage vs ngrok

| Feature | ngrok AI Gateway | UniRoute |
|---------|------------------|----------|
| **Pricing** | Paid (usage-based) | **100% FREE** |
| **Open Source** | ❌ Closed source | ✅ **Open source** |
| **Self-hosted** | Limited | ✅ **Fully self-hostable** |
| **Community** | Enterprise-focused | ✅ **Community-driven** |
| **Customization** | Limited | ✅ **Fully customizable** |
| **Local LLM Support** | Limited | ✅ **Native support** |
| **Transparency** | Limited | ✅ **Full code visibility** |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+ (or Supabase free tier)
- Redis 7+ (or Upstash free tier)
- Docker (optional, for Coolify deployment)

### 5-Minute Setup

```bash
# 1. Clone repository
git clone https://github.com/uniroute/ai-gateway
cd ai-gateway

# 2. Install dependencies
go mod download

# 3. Set up environment
cp .env.example .env
# Edit .env with your configuration

# 4. Run migrations
make migrate

# 5. Start server
make dev

# 6. Test it
curl http://localhost:8080/health
```

### First Request

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

---

## 🏗️ Architecture

### System Architecture

```
┌─────────────────┐
│   Client Apps   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│      AI Gateway (API)          │
│  ┌───────────────────────────┐ │
│  │  Request Router          │ │
│  │  - Model Selection       │ │
│  │  - Load Balancing        │ │
│  │  - Failover              │ │
│  └───────────────────────────┘ │
│  ┌───────────────────────────┐ │
│  │  Security Layer          │ │
│  │  - Authentication         │ │
│  │  - Rate Limiting         │ │
│  │  - API Key Management    │ │
│  │  - Zero Trust            │ │
│  └───────────────────────────┘ │
│  ┌───────────────────────────┐ │
│  │  Monitoring & Analytics  │ │
│  │  - Usage Tracking        │ │
│  │  - Cost Tracking         │ │
│  │  - Performance Metrics   │ │
│  │  - Live Traffic View     │ │
│  └───────────────────────────┘ │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│   Model Providers (Backends)   │
│  ┌──────┐  ┌──────┐  ┌──────┐ │
│  │OpenAI│  │Anthropic│Local│ │
│  │      │  │ Claude │ LLM  │ │
│  └──────┘  └──────┘  └──────┘ │
│  ┌──────┐  ┌──────┐  ┌──────┐ │
│  │Google│  │Cohere│  │...   │ │
│  │Gemini│  │      │  │      │ │
│  └──────┘  └──────┘  └──────┘ │
└─────────────────────────────────┘
```

### Project Structure

```
ai-gateway/
├── cmd/
│   └── gateway/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/             # HTTP handlers
│   │   ├── middleware/          # Middleware (auth, rate limit)
│   │   └── router.go           # Route definitions
│   ├── gateway/
│   │   ├── router.go           # Request routing logic
│   │   ├── loadbalancer.go     # Load balancing
│   │   ├── failover.go         # Failover logic
│   │   └── model_selector.go   # Model selection strategy
│   ├── providers/
│   │   ├── openai.go           # OpenAI provider
│   │   ├── anthropic.go        # Anthropic provider
│   │   ├── local.go            # Local LLM provider
│   │   └── interface.go        # Provider interface
│   ├── security/
│   │   ├── auth.go             # Authentication
│   │   ├── apikey.go           # API key management
│   │   └── ratelimit.go        # Rate limiting
│   ├── storage/
│   │   ├── postgres.go         # PostgreSQL client
│   │   ├── redis.go            # Redis client
│   │   └── models.go            # Database models
│   ├── monitoring/
│   │   ├── metrics.go          # Prometheus metrics
│   │   ├── analytics.go        # Usage analytics
│   │   └── cost_tracking.go    # Cost tracking
│   └── config/
│       └── config.go           # Configuration management
├── pkg/
│   ├── logger/                 # Logging utilities
│   └── errors/                 # Error handling
├── migrations/
│   └── *.sql                   # Database migrations
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── coolify/
│   └── docker-compose.yml      # Coolify deployment config
├── .env.example
├── go.mod
├── Makefile
└── README.md
```

---

## 💻 Technology Stack

### Primary Language: Go (Golang) ⭐ Recommended

**Why Go?**
- ✅ High Performance: Excellent for concurrent request handling
- ✅ Low Memory Footprint: Perfect for containerized deployments
- ✅ Built-in Concurrency: Goroutines handle thousands of connections
- ✅ Fast Compilation: Quick iteration during development
- ✅ Strong Standard Library: HTTP, JSON, crypto built-in
- ✅ Great for Gateways: Used by major API gateways (Kong, Traefik)

### Complete Stack

```
Backend:        Go 1.21+
API Framework:  Gin (lightweight, fast)
Database:       PostgreSQL (primary) + Redis (caching)
Message Queue:  Redis Streams (for async processing)
Auth:           JWT + API Keys (stored in PostgreSQL)
Monitoring:     Prometheus + Grafana
Logging:        Structured logging (zerolog)
CLI:            Cobra
SDKs:           Go, Python, JavaScript
```

### Dependencies

```go
// go.mod
module github.com/uniroute/ai-gateway

require (
    // Web Framework
    github.com/gin-gonic/gin v1.9.1
    
    // Database
    github.com/go-redis/redis/v8 v8.11.5
    github.com/jackc/pgx/v5 v5.4.3
    
    // Authentication
    github.com/golang-jwt/jwt/v5 v5.0.0
    
    // Monitoring
    github.com/prometheus/client_golang v1.17.0
    
    // Configuration
    github.com/spf13/viper v1.17.0
    
    // Logging
    github.com/rs/zerolog v1.31.0
    
    // Rate Limiting
    golang.org/x/time v0.4.0
    
    // CLI
    github.com/spf13/cobra v1.8.0
    
    // OpenAPI/Swagger
    github.com/swaggo/swag v1.16.2
    github.com/swaggo/gin-swagger v1.6.0
    
    // WebSocket
    github.com/gorilla/websocket v1.5.1
    
    // OAuth/OIDC
    golang.org/x/oauth2 v0.15.0
)
```

---

## 🧹 Code Quality & Best Practices

### Clean Code Principles

We follow **clean code principles** to ensure maintainability, readability, and reusability throughout the codebase. This is critical for an open-source project that will be maintained by the community.

#### 1. **Single Responsibility Principle (SRP)**
Each function, struct, and package should have one clear purpose.

```go
// ✅ Good: Single responsibility
func ValidateAPIKey(key string) error {
    // Only validates API key format
    if len(key) < 32 {
        return ErrInvalidAPIKey
    }
    return nil
}

// ❌ Bad: Multiple responsibilities
func ProcessRequestAndValidateAndLog(key string, req Request) error {
    // Too many responsibilities - validation, processing, logging
}
```

#### 2. **DRY (Don't Repeat Yourself)**
Extract common patterns into reusable functions and packages.

```go
// ✅ Good: Reusable error handling
func HandleProviderError(err error, provider string) error {
    return fmt.Errorf("provider %s: %w", provider, err)
}

// ❌ Bad: Repeated error handling
func CallOpenAI() error {
    if err != nil {
        return fmt.Errorf("provider openai: %v", err)
    }
}
func CallAnthropic() error {
    if err != nil {
        return fmt.Errorf("provider anthropic: %v", err)
    }
}
```

#### 3. **Interface-Based Design**
Use interfaces for abstraction and testability. This is especially important for provider implementations.

```go
// ✅ Good: Provider interface for reusability
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
    HealthCheck(ctx context.Context) error
    GetModels() []string
}

// All providers implement the same interface
type OpenAIProvider struct {}
type AnthropicProvider struct {}
type LocalLLMProvider struct {}
```

#### 4. **Composition Over Inheritance**
Use composition and embedding for code reuse.

```go
// ✅ Good: Composition with base functionality
type BaseProvider struct {
    client   *http.Client
    logger   *zerolog.Logger
    metrics  *prometheus.CounterVec
}

type OpenAIProvider struct {
    BaseProvider  // Embed base functionality
    apiKey        string
    baseURL       string
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Reuse BaseProvider's client and logger
    return p.makeRequest(ctx, req)
}
```

#### 5. **Error Handling**
Consistent, informative error handling with error wrapping.

```go
// ✅ Good: Wrapped errors with context
var (
    ErrProviderUnavailable = errors.New("provider unavailable")
    ErrRateLimitExceeded   = errors.New("rate limit exceeded")
    ErrInvalidRequest      = errors.New("invalid request")
)

func CallProvider(ctx context.Context, req Request) error {
    if err := provider.Call(ctx, req); err != nil {
        return fmt.Errorf("failed to call provider: %w", err)
    }
    return nil
}
```

### Reusable Code Patterns

#### 1. **Provider Interface Pattern**
All LLM providers implement a common interface for easy swapping and testing:

```go
// internal/providers/interface.go
package providers

// Provider defines the interface for all LLM providers.
// This interface allows easy swapping of providers and testing.
type Provider interface {
    // Name returns the provider's name (e.g., "openai", "anthropic")
    Name() string
    
    // Chat sends a chat request to the provider
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    
    // Stream sends a streaming chat request
    Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
    
    // HealthCheck verifies the provider is available
    HealthCheck(ctx context.Context) error
    
    // GetModels returns list of available models
    GetModels() []string
}

// Easy to add new providers - just implement the interface
type CustomProvider struct {
    // implements Provider interface
}
```

#### 2. **Middleware Pattern**
Reusable middleware for cross-cutting concerns:

```go
// internal/api/middleware/auth.go
func AuthMiddleware(apiKeyService *APIKeyService) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := extractAPIKey(c)
        if err := apiKeyService.Validate(key); err != nil {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// Reusable across all routes
router.Use(AuthMiddleware(apiKeyService))
router.Use(RateLimitMiddleware(rateLimiter))
router.Use(LoggingMiddleware(logger))
router.Use(RecoveryMiddleware())
```

#### 3. **Repository Pattern**
Abstract data access for testability and reusability:

```go
// internal/storage/repository.go
type APIKeyRepository interface {
    FindByKeyHash(ctx context.Context, hash string) (*APIKey, error)
    Create(ctx context.Context, key *APIKey) error
    Update(ctx context.Context, key *APIKey) error
    Delete(ctx context.Context, id string) error
    ListByUserID(ctx context.Context, userID string) ([]*APIKey, error)
}

// PostgreSQL implementation
type PostgresAPIKeyRepository struct {
    db *sql.DB
}

// Redis cache implementation (decorator pattern)
type CachedAPIKeyRepository struct {
    repo  APIKeyRepository
    cache *redis.Client
    ttl   time.Duration
}
```

#### 4. **Strategy Pattern**
Pluggable algorithms for routing, load balancing, etc.:

```go
// internal/gateway/strategy.go
type RoutingStrategy interface {
    SelectProvider(ctx context.Context, req Request, providers []Provider) (Provider, error)
}

// Different strategies
type CostBasedStrategy struct {
    costCalculator *CostCalculator
}

type LatencyBasedStrategy struct {
    latencyTracker *LatencyTracker
}

type RoundRobinStrategy struct {
    current int
    mu      sync.Mutex
}

type FailoverStrategy struct {
    primary   Provider
    fallbacks []Provider
}

// Easy to swap strategies
router := NewRouter(CostBasedStrategy{})
// or
router := NewRouter(LatencyBasedStrategy{})
```

#### 5. **Factory Pattern**
Centralized creation of providers and services:

```go
// internal/providers/factory.go
type ProviderFactory struct {
    config *Config
    logger *zerolog.Logger
    metrics *prometheus.Registry
}

func NewProviderFactory(config *Config, logger *zerolog.Logger) *ProviderFactory {
    return &ProviderFactory{
        config:  config,
        logger:   logger,
        metrics: prometheus.NewRegistry(),
    }
}

func (f *ProviderFactory) CreateProvider(name string) (Provider, error) {
    switch name {
    case "openai":
        return NewOpenAIProvider(f.config.OpenAI, f.logger, f.metrics)
    case "anthropic":
        return NewAnthropicProvider(f.config.Anthropic, f.logger, f.metrics)
    case "local":
        return NewLocalLLMProvider(f.config.Local, f.logger, f.metrics)
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}
```

### Code Organization Guidelines

#### Package Structure
- **`pkg/`**: Reusable packages that can be imported by other projects
  - `pkg/logger/` - Logging utilities
  - `pkg/errors/` - Error handling utilities
  - `pkg/validator/` - Validation utilities
- **`internal/`**: Private application code, not importable by external projects
  - `internal/api/` - HTTP handlers and middleware
  - `internal/gateway/` - Routing logic
  - `internal/providers/` - Provider implementations
- **`cmd/`**: Application entry points
  - `cmd/gateway/` - Main gateway application

#### Naming Conventions
```go
// ✅ Good naming
type APIKeyService struct {}        // Clear, descriptive
func ValidateRequest(req Request)   // Verb + noun
const MaxRetries = 3                // Constants in PascalCase
var defaultTimeout = 30 * time.Second  // Unexported constants in camelCase

// ❌ Bad naming
type Svc struct {}                  // Abbreviation unclear
func Do(req Request)                // Too generic
const max = 3                       // Should be exported if used elsewhere
```

#### Function Guidelines
- **Keep functions small**: Max 50 lines, ideally 20-30
- **One level of abstraction**: Don't mix high-level and low-level logic
- **Descriptive names**: Function name should describe what it does
- **Avoid side effects**: Functions should do one thing and do it well

```go
// ✅ Good: Small, focused function
func CalculateCost(tokens int, model string) (float64, error) {
    rate, err := GetModelRate(model)
    if err != nil {
        return 0, fmt.Errorf("failed to get rate for model %s: %w", model, err)
    }
    return float64(tokens) * rate, nil
}

// ❌ Bad: Too many responsibilities
func ProcessEverything(req Request) (*Response, error) {
    // 200 lines of mixed logic
    // Hard to test
    // Hard to maintain
}
```

### Testing Best Practices

#### Unit Tests
- Test one thing at a time
- Use table-driven tests for multiple scenarios
- Mock external dependencies

```go
// ✅ Good: Table-driven test
func TestCalculateCost(t *testing.T) {
    tests := []struct {
        name     string
        tokens   int
        model    string
        want     float64
        wantErr  bool
    }{
        {
            name:    "gpt-4 calculation",
            tokens:  1000,
            model:   "gpt-4",
            want:    0.03,
            wantErr: false,
        },
        {
            name:    "invalid model",
            tokens:  1000,
            model:   "invalid",
            want:    0,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateCost(tt.tokens, tt.model)
            if (err != nil) != tt.wantErr {
                t.Errorf("CalculateCost() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("CalculateCost() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Code Review Checklist

Before submitting code, ensure:

- [ ] **Single Responsibility**: Each function/struct has one clear purpose
- [ ] **DRY**: No code duplication; common logic extracted
- [ ] **Interfaces**: Used for abstraction and testability
- [ ] **Error Handling**: All errors handled appropriately with context
- [ ] **Tests**: Unit tests for new functionality (>80% coverage goal)
- [ ] **Documentation**: Public functions/types documented with godoc
- [ ] **Naming**: Clear, descriptive names following Go conventions
- [ ] **Formatting**: Code formatted with `gofmt`
- [ ] **Linting**: No linter warnings (`golangci-lint`)
- [ ] **Security**: No security vulnerabilities introduced
- [ ] **Performance**: No obvious performance issues

### Reusability Checklist

When writing code, ask:

1. **Can this be reused?** - Extract to `pkg/` if reusable across projects
2. **Is this abstracted?** - Use interfaces for flexibility
3. **Is this configurable?** - Avoid hardcoded values
4. **Is this testable?** - Dependencies injected, not hardcoded
5. **Is this documented?** - Clear documentation for future developers

### Example: Reusable Provider Pattern

```go
// internal/providers/base.go - Reusable base provider
package providers

type BaseProvider struct {
    name     string
    client   *http.Client
    logger   *zerolog.Logger
    metrics  *prometheus.CounterVec
}

// NewBaseProvider creates a new base provider with common functionality
func NewBaseProvider(name string, logger *zerolog.Logger) *BaseProvider {
    return &BaseProvider{
        name:   name,
        client: &http.Client{Timeout: 30 * time.Second},
        logger: logger,
        metrics: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "provider_requests_total",
                Help: "Total number of provider requests",
            },
            []string{"provider", "status"},
        ),
    }
}

// makeRequest is a reusable HTTP request method
func (b *BaseProvider) makeRequest(ctx context.Context, url string, body interface{}) (*http.Response, error) {
    // Reusable HTTP request logic
    // Logging, metrics, retries, etc.
    b.logger.Info().
        Str("provider", b.name).
        Str("url", url).
        Msg("making request")
    
    // ... request implementation
    
    return resp, nil
}

// internal/providers/openai.go - Specific implementation
type OpenAIProvider struct {
    *BaseProvider  // Embed base functionality
    apiKey         string
    baseURL        string
}

func NewOpenAIProvider(config Config, logger *zerolog.Logger) *OpenAIProvider {
    return &OpenAIProvider{
        BaseProvider: NewBaseProvider("openai", logger),
        apiKey:       config.APIKey,
        baseURL:      config.BaseURL,
    }
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Uses BaseProvider.makeRequest() for HTTP calls
    // Reuses logging, metrics, error handling
    return p.makeRequest(ctx, p.baseURL+"/chat", req)
}
```

### Code Standards

#### Go Style Guide
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (enforced in CI)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

#### Required Development Tools
```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run before committing
make fmt      # Format code with gofmt
make lint      # Run linters (golangci-lint)
make vet       # Run go vet
make test      # Run tests
make security  # Security scan (gosec)
```

### Reusability Patterns Reference

| Pattern | Use Case | Location | Example |
|---------|----------|----------|---------|
| **Interface** | Provider abstraction | `internal/providers/interface.go` | All providers implement `Provider` interface |
| **Middleware** | Cross-cutting concerns | `internal/api/middleware/` | Auth, rate limiting, logging |
| **Repository** | Data access abstraction | `internal/storage/repository.go` | Database operations |
| **Strategy** | Pluggable algorithms | `internal/gateway/strategy.go` | Routing strategies |
| **Factory** | Object creation | `internal/providers/factory.go` | Provider creation |
| **Base Struct** | Shared functionality | `internal/providers/base.go` | Common provider code |

### Common Reusable Components

- **Error Handling**: `pkg/errors/` - Standardized error types and wrapping
- **Logging**: `pkg/logger/` - Structured logging utilities
- **HTTP Client**: `internal/providers/base.go` - Reusable HTTP client with retries
- **Validation**: `internal/api/validators/` - Request validation utilities
- **Metrics**: `internal/monitoring/metrics.go` - Prometheus metrics helpers
- **Config**: `internal/config/` - Configuration management

---

## 🔐 Security Requirements

### ⚠️ CRITICAL: Security First

**DO NOT DEPLOY without completing these security measures. Hackers will find vulnerabilities within hours.**

### Day 1 Security Checklist

- [ ] **Input Validation**
  - All API endpoints validate input
  - SQL injection prevention (parameterized queries only)
  - XSS prevention (input sanitization)
  - Request size limits enforced

- [ ] **Authentication**
  - API keys properly hashed (bcrypt, not plaintext)
  - JWT tokens signed with strong secret (min 32 chars)
  - Token expiration enforced

- [ ] **Rate Limiting**
  - Per-API-key rate limits
  - Per-IP rate limits
  - Redis-based implementation

- [ ] **Secrets Management**
  - No secrets in code or git
  - Environment variables for all secrets
  - Provider API keys encrypted at rest
  - `.env` in `.gitignore`

- [ ] **HTTPS/TLS**
  - HTTPS enforced (no HTTP in production)
  - TLS 1.3 minimum
  - Valid SSL certificate
  - HSTS header enabled

- [ ] **Security Headers**
  - X-Frame-Options: DENY
  - X-Content-Type-Options: nosniff
  - Content-Security-Policy configured
  - Strict-Transport-Security header

- [ ] **Error Handling**
  - No sensitive data in error messages
  - Generic errors to users
  - Detailed errors logged server-side only

- [ ] **Database Security**
  - Parameterized queries only
  - Database credentials secured
  - Connection encryption enabled

**Full security guide**: See `AI_GATEWAY_SECURITY.md` for complete implementation.

---

## 🚀 Implementation Plan

**⚠️ Important: Complete, Test, and Verify Each Phase Before Moving Forward**

### Development Approach: Iterative & Test-Driven

**Follow this process for each phase:**

1. ✅ **Implement** - Complete all tasks in the phase
2. ✅ **Test** - Write and run tests (unit, integration, manual)
3. ✅ **Verify** - Confirm all functionality works end-to-end
4. ✅ **Document** - Update code comments and documentation
5. ✅ **Review** - Code review and quality checks
6. ✅ **Checklist** - Complete phase testing checklist
7. ✅ **Only then** - Proceed to next phase

**Why this approach?**
- 🐛 **Catch bugs early** - Issues found before they compound
- 🔒 **Stable foundation** - Each phase builds on tested code
- 📊 **Clear progress** - Know exactly what's working
- 🚀 **Confident deployment** - Each phase is production-ready
- 🔄 **Easy rollback** - Can revert to last working phase if needed

**Testing Requirements:**
- Unit tests for all new functions
- Integration tests for component interactions
- Manual testing of all features
- Code coverage >80% for new code
- All tests must pass before moving forward

---

### Phase 1: Core Gateway (Week 1-2)

**Priority: Local LLM Support + Shareable Server** 🏠🌐

Phase 1 includes **TWO key capabilities**:

1. **UniRoute as a Shareable Server** - Users can host and share their gateway
2. **Local LLM Provider** - Connect to locally hosted models (Ollama)

We prioritize local LLM provider implementation because:
- ✅ **No API costs** - Users can test without spending money
- ✅ **Privacy** - Data stays on user's infrastructure
- ✅ **Self-hostable** - Aligns with UniRoute's core mission
- ✅ **Easy testing** - Developers can test without external API keys
- ✅ **Key differentiator** - Best-in-class local model support
- ✅ **Shareable** - Users can host UniRoute and share access with others

**Tasks:**

**A. Core Server Infrastructure**
- [ ] Set up Go project structure
- [ ] **Implement HTTP server with Gin** (shareable server)
  - Server runs on configurable host/port (default: `:8080`)
  - Accessible over network (not just localhost)
  - Users can share their UniRoute instance with others
  - Support for remote access (with proper authentication)
- [ ] Create provider interface
- [ ] Basic request routing
- [ ] Simple authentication (API keys)
  - API key generation and validation
  - Users can create API keys to share access

**B. Local LLM Provider** ⭐ Priority
- [ ] **Implement Local LLM provider (Ollama)**
  - Support for locally hosted models (Ollama, vLLM)
  - Connect to `http://localhost:11434` (default Ollama port)
  - Support configurable Ollama host/port
  - Support chat completions and streaming
  - Users can host their own LLM server
  - No API keys required for local hosting
  - Enable self-hosting from day one

**Use Case:**
1. User runs Ollama locally: `ollama serve` (on port 11434)
2. User runs UniRoute: `uniroute start` (on port 8080)
3. User shares UniRoute URL: `http://their-server:8080`
4. Others can access UniRoute with API key
5. UniRoute routes requests to user's local Ollama instance
6. **Result**: User shares their local LLM through UniRoute gateway

**Note:** OpenAI and other cloud providers will be implemented in Phase 3, allowing users to use UniRoute with local models first, then add cloud providers as needed.

#### ✅ Phase 1 Testing & Verification Checklist

**Before moving to Phase 2, verify:**

- [ ] **Server starts successfully**
  ```bash
  uniroute start
  # Should start on :8080 without errors
  ```

- [ ] **Health endpoint works**
  ```bash
  curl http://localhost:8080/health
  # Should return: {"status": "ok"}
  ```

- [ ] **Local LLM provider connects**
  ```bash
  # With Ollama running
  curl -X POST http://localhost:8080/v1/chat \
    -H "Authorization: Bearer YOUR_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"model": "llama2", "messages": [{"role": "user", "content": "Hello"}]}'
  # Should return valid response from Ollama
  ```

- [ ] **API key authentication works**
  - Valid API key → Request succeeds
  - Invalid API key → Request fails with 401
  - No API key → Request fails with 401

- [ ] **Request routing works**
  - Request routed to correct provider
  - Response format is correct
  - Errors handled gracefully

- [ ] **Server accessible on network**
  - Can access from another machine on same network
  - Can access via ngrok/cloudflared tunnel

- [ ] **Unit tests pass**
  ```bash
  make test
  # All tests should pass
  ```

- [ ] **Integration tests pass**
  ```bash
  make test-integration
  # Integration tests should pass
  ```

- [ ] **Code quality checks pass**
  ```bash
  make fmt lint vet security
  # No errors or warnings
  ```

**Only proceed to Phase 2 when all Phase 1 items are ✅ complete and tested.**

#### What Phase 1 Includes

Phase 1 delivers **two critical capabilities**:

### 1. UniRoute as a Shareable Server 🌐

**What it means:**
- UniRoute runs as an HTTP server that others can access
- Not just localhost - accessible over network
- Users can host UniRoute and share it with team/clients
- API key authentication for access control

**Implementation:**
```go
// cmd/gateway/main.go
func main() {
    router := gin.Default()
    
    // API endpoints
    router.POST("/v1/chat", chatHandler)
    router.GET("/health", healthHandler)
    
    // Start server (accessible on network)
    router.Run(":8080")  // or "0.0.0.0:8080" for all interfaces
}
```

**Use Case:**
```
User A runs UniRoute on their server (192.168.1.100:8080)
User A creates API key: "ur_abc123..."
User A shares: "Use my UniRoute: http://192.168.1.100:8080"
User B can now access: curl -H "Authorization: Bearer ur_abc123..." http://192.168.1.100:8080/v1/chat
```

### 2. Local LLM Provider (Ollama) 🏠

**What it means:**
- UniRoute connects to locally hosted LLM servers (Ollama)
- Routes requests from UniRoute API to local Ollama instance
- Translates between UniRoute format and Ollama format

**Implementation:**
```go
// internal/providers/local.go
type LocalLLMProvider struct {
    baseURL string  // e.g., "http://localhost:11434"
    client  *http.Client
    logger  *zerolog.Logger
}

func (p *LocalLLMProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Convert UniRoute format to Ollama format
    ollamaReq := convertToOllamaFormat(req)
    
    // Make HTTP request to local Ollama server
    resp, err := p.client.Post(
        p.baseURL + "/api/chat",
        ollamaReq,
    )
    
    // Convert Ollama response back to UniRoute format
    return convertFromOllamaFormat(resp), err
}
```

**Features:**
- Connect to `http://localhost:11434` (Ollama's default port)
- Support configurable host/port for user's local server
- List available models from local server
- Support model selection (llama2, mistral, etc.)
- Handle connection errors gracefully

### Combined Use Case: Sharing Local LLM via UniRoute

**Complete Flow:**
1. **User runs Ollama locally:**
   ```bash
   ollama serve  # Runs on localhost:11434
   ```

2. **User runs UniRoute:**
   ```bash
   uniroute start  # Runs on :8080 (accessible on network)
   ```

3. **User exposes UniRoute to internet (optional):**
   ```bash
   # Option A: Using ngrok (until built-in tunneling is ready)
   ngrok http 8080
   # Returns: https://abc123.ngrok-free.app -> http://localhost:8080
   
   # Option B: Using cloudflared (free alternative)
   cloudflared tunnel --url http://localhost:8080
   # Returns: https://random-subdomain.trycloudflare.com
   
   # Option C: Built-in tunneling (Phase 6)
   uniroute tunnel --port 8080
   # Returns: https://your-instance.uniroute.dev
   ```

4. **User creates API key:**
   ```bash
   uniroute keys create --name "Team Access"
   # Returns: ur_abc123...
   ```

5. **User shares UniRoute:**
   - URL: `https://abc123.ngrok-free.app` (or local IP if on same network)
   - API Key: `ur_abc123...`

6. **Others access shared gateway:**
   ```bash
   curl -X POST https://abc123.ngrok-free.app/v1/chat \
     -H "Authorization: Bearer ur_abc123..." \
     -H "Content-Type: application/json" \
     -d '{"model": "llama2", "messages": [...]}'
   ```

7. **UniRoute routes to local Ollama:**
   - Receives request → Validates API key → Routes to Ollama → Returns response

**Benefits:**
- ✅ **Free** - No API costs
- ✅ **Private** - Data stays on user's infrastructure
- ✅ **Shareable** - Users can share their local LLM with others
- ✅ **Fast** - No network latency to cloud
- ✅ **Self-hosted** - Full control over infrastructure
- ✅ **Testable** - Developers can test without external dependencies

### Phase 2: Security & Rate Limiting (Week 3)

**⚠️ Do not start Phase 2 until Phase 1 is complete, tested, and verified.**

**Tasks:**
- [ ] JWT authentication
- [ ] API key management (CRUD)
- [ ] Rate limiting (Redis-based)
- [ ] Request validation
- [ ] IP whitelisting
- [ ] Security headers

#### ✅ Phase 2 Testing & Verification Checklist

**Before moving to Phase 3, verify:**

- [ ] **JWT authentication works**
  - Valid JWT token → Request succeeds
  - Expired token → Request fails with 401
  - Invalid token → Request fails with 401

- [ ] **API key CRUD works**
  - Create API key → Key created successfully
  - List API keys → Keys displayed
  - Revoke API key → Key revoked, requests fail
  - Update API key → Changes applied

- [ ] **Rate limiting works**
  - Within limit → Request succeeds
  - Exceeds limit → Request fails with 429
  - Per-key limits enforced
  - Per-IP limits enforced

- [ ] **Request validation works**
  - Valid request → Processed
  - Invalid request → Returns 400 with error message
  - Malformed JSON → Returns 400

- [ ] **Security headers present**
  - X-Frame-Options: DENY
  - X-Content-Type-Options: nosniff
  - Content-Security-Policy set
  - HSTS header (if HTTPS)

- [ ] **All Phase 1 functionality still works**
  - Regression testing passed

**Only proceed to Phase 3 when all Phase 2 items are ✅ complete and tested.**

### Phase 3: Multi-Provider Support (Week 4)

**⚠️ Do not start Phase 3 until Phase 2 is complete, tested, and verified.**

**Cloud Provider Support** ☁️

Now that local hosting works, add cloud providers:
- [ ] **OpenAI provider** (moved from Phase 1)
- [ ] Anthropic provider
- [ ] Google provider
- [ ] Provider health checks
- [ ] Failover logic

**Why OpenAI is Phase 3:**
- Requires API keys (users need to sign up and pay)
- Local hosting is free and more accessible
- Users can test UniRoute fully with local models first
- Cloud providers can be added incrementally based on demand

#### ✅ Phase 3 Testing & Verification Checklist

**Before moving to Phase 4, verify:**

- [ ] **OpenAI provider works**
  - Valid API key → Requests succeed
  - Invalid API key → Requests fail with proper error
  - Response format matches UniRoute standard

- [ ] **Anthropic provider works**
  - Same verification as OpenAI

- [ ] **Google provider works**
  - Same verification as OpenAI

- [ ] **Provider health checks work**
  - Healthy provider → Returns available
  - Unhealthy provider → Returns unavailable
  - Health check endpoint functional

- [ ] **Failover logic works**
  - Primary provider fails → Fails over to backup
  - All providers fail → Returns error
  - Failover happens automatically

- [ ] **All previous phases still work**
  - Phase 1 & 2 functionality intact
  - No regressions introduced

**Only proceed to Phase 4 when all Phase 3 items are ✅ complete and tested.**

### Phase 4: Advanced Routing (Week 5)
- [ ] Load balancing
- [ ] Model selection strategies
- [ ] Cost-based routing
- [ ] Latency-based routing
- [ ] Custom routing rules

### Phase 5: Monitoring & Analytics (Week 6)

**⚠️ Do not start Phase 5 until Phase 4 is complete, tested, and verified.**

**Tasks:**
- [ ] Prometheus metrics
- [ ] Usage tracking
- [ ] Cost calculation
- [ ] Analytics API
- [ ] Basic dashboard / Advanced

#### ✅ Phase 5 Testing & Verification Checklist

**Before moving to Phase 6, verify:**

- [ ] **Prometheus metrics exposed**
  - Metrics endpoint accessible
  - Metrics updated correctly
  - Can scrape with Prometheus

- [ ] **Usage tracking works**
  - Requests tracked in database
  - Usage stats accurate
  - Per-user usage tracked

- [ ] **Cost calculation accurate**
  - Costs calculated correctly
  - Per-provider costs tracked
  - Total costs accurate

- [ ] **Analytics API works**
  - Endpoints return correct data
  - Filtering works
  - Aggregation correct

**Only proceed to Phase 6 when all Phase 5 items are ✅ complete and tested.**

### Phase 6: Developer Experience (Week 7-8)
- [ ] CLI tool (`uniroute` command)
  - `uniroute start` - Start gateway server
  - `uniroute tunnel` - Expose local server to internet (ngrok-like)
    - Built-in tunneling with public URL
    - `uniroute tunnel --port 8080` → `https://your-instance.uniroute.dev`
    - Free public URLs for sharing
    - Web interface for monitoring (like ngrok's http://127.0.0.1:4040)
    - Session management and connection stats
  - `uniroute keys create` - Create API keys
  - `uniroute status` - Check server status
  - `uniroute logs` - View live logs
- [ ] SDKs (Go, Python, JavaScript)
- [ ] OpenAPI/Swagger documentation
- [ ] One-line setup script
- [ ] Example applications

#### Tunneling Feature (ngrok-like)

**What it does:**
Expose your local UniRoute server to the internet with a public URL, just like ngrok.

**Usage:**
```bash
# Start UniRoute locally
uniroute start  # Runs on localhost:8080

# In another terminal, expose it to internet
uniroute tunnel --port 8080

# Output:
# Session Status                online
# Account                       your-email@example.com (Plan: Free)
# Version                       1.0.0
# Region                        Europe (eu)
# Latency                       45ms
# Web Interface                 http://127.0.0.1:4040
# Forwarding                    https://abc123.uniroute.dev -> http://localhost:8080
```

**Features:**
- ✅ **Free public URLs** - Get a public HTTPS URL instantly
- ✅ **Web interface** - Monitor requests at `http://127.0.0.1:4040`
- ✅ **Connection stats** - See latency, requests, connections
- ✅ **Secure** - HTTPS by default
- ✅ **No port forwarding** - Works behind NAT/firewall

**Until Phase 6 (Built-in Tunneling):**

Users can use existing tools:
- **ngrok**: `ngrok http 8080` (free tier available)
- **cloudflared**: `cloudflared tunnel --url http://localhost:8080` (completely free, no signup)
- **Local network**: Share local IP if on same network

### Phase 7: Production Ready (Week 9-10)
- [ ] Database migrations
- [ ] Error handling improvements
- [ ] Logging & observability
- [ ] Performance optimization
- [ ] Security audit
- [ ] Documentation
- [ ] Coolify deployment

### Phase 8: Advanced Features (Week 11-12)
- [ ] Zero Trust (SSO, OAuth, SAML, OIDC)
- [ ] Mutual TLS (mTLS)
- [ ] WebSocket streaming
- [ ] Traffic Policy UI
- [ ] Live traffic viewer
- [ ] Global CDN integration

---

## 💰 Cost Analysis

### Total Cost to Build & Run

**Development Cost**: **$0** (if you're building it yourself)  
**Monthly Operating Cost**: **$0 - $50** (depending on scale)  
**One-Time Costs**: **$0 - $50** (API testing during development)

### Free Infrastructure Stack

```
Cloudflare Pages (API):     $0/month
Railway.app (Backend):      $0/month (free tier)
Supabase (Database):        $0/month (free tier)
Upstash (Redis):            $0/month (free tier)
─────────────────────────────────────────
Total:                      $0/month
```

### Low-Cost Production Stack

```
VPS (Hetzner):              $5/month
Self-hosted PostgreSQL:     $0/month
Self-hosted Redis:          $0/month
Cloudflare (CDN/DDoS):      $0/month
─────────────────────────────────────────
Total:                      $5-10/month
```

### What You Don't Pay For

- ✅ Development tools (all free)
- ✅ Code hosting (GitHub free)
- ✅ CI/CD (GitHub Actions free)
- ✅ Monitoring (Prometheus free)
- ✅ SSL certificates (Let's Encrypt free)
- ✅ DDoS protection (Cloudflare free tier)

**The only "cost" is your time**, which is an investment in learning and building.

---

## 🆚 Competitive Analysis

### vs ngrok AI Gateway

[ngrok](https://ngrok.com/) just launched their AI Gateway - **"One gateway for every AI model"** - available in early access at [ngrok.ai](https://ngrok.ai). This validates our project concept.

### Feature Comparison

| Feature | ngrok AI Gateway | UniRoute |
|---------|------------------|----------|
| **Unified API for all LLMs** | ✅ | ✅ |
| **Route to cloud models** | ✅ | ✅ |
| **Route to local models** | ✅ | ✅ |
| **Load balancing** | ✅ | ✅ |
| **Rate limiting** | ✅ | ✅ |
| **Authentication** | ✅ | ✅ |
| **Monitoring** | ✅ | ✅ |
| **Traffic Policy** | ✅ | ⏳ Planned |
| **Instant Observability** | ✅ | ⏳ Planned |
| **CLI Tool** | ✅ | ⏳ Planned |
| **SDKs** | ✅ | ⏳ Planned |
| **Pricing** | Paid | **FREE** |
| **Open Source** | ❌ | ✅ |
| **Self-hosted** | Limited | ✅ |

### Our Unique Advantages

1. **100% Free** - No usage limits, no credit card required
2. **Open Source** - Full transparency, community-driven
3. **Self-Hostable** - Deploy anywhere, full control
4. **Local LLM Focus** - Best-in-class local model support
5. **Community-Driven** - Users shape features

---

## 🎯 Features & Roadmap

### Core Features (MVP)

1. **Unified API Interface**
   - Single endpoint for all LLM providers
   - Standardized request/response format
   - Provider abstraction layer

2. **Intelligent Routing**
   - Model selection based on cost, latency, availability
   - Load balancing across instances
   - Automatic failover

3. **Security & Access Control**
   - API key management
   - JWT authentication
   - Rate limiting (per-key, per-IP)
   - Zero Trust security (SSO, OAuth, SAML, OIDC)

4. **Monitoring & Analytics**
   - Usage tracking
   - Cost tracking
   - Performance metrics
   - Real-time dashboard
   - Live traffic viewer

5. **Multi-Provider Support**
   - OpenAI (GPT-4, GPT-3.5)
   - Anthropic (Claude)
   - Google (Gemini)
   - Cohere
   - Local LLMs (Ollama, vLLM)
   - Custom providers

6. **Developer Experience**
   - CLI tool (`uniroute` command)
   - SDKs (Go, Python, JavaScript)
   - OpenAPI/Swagger docs
   - One-line setup

7. **Protocol Support**
   - HTTP/HTTPS
   - TLS
   - TCP
   - WebSocket (streaming)
   - gRPC (future)

### Roadmap

**Q1**: Core functionality + Developer experience (CLI, SDKs)  
**Q2**: Enterprise features (Zero Trust, SSO, OAuth)  
**Q3**: Scale & performance (Global CDN, edge locations)

---

## 🗄️ Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### API Keys Table
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    rate_limit_per_minute INTEGER DEFAULT 60,
    rate_limit_per_day INTEGER DEFAULT 10000,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);
```

### Providers Table
```sql
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    api_key_encrypted TEXT,
    base_url VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    config JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Requests Table (Analytics)
```sql
CREATE TABLE requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID REFERENCES api_keys(id),
    provider_id UUID REFERENCES providers(id),
    model VARCHAR(100),
    request_type VARCHAR(50),
    input_tokens INTEGER,
    output_tokens INTEGER,
    cost DECIMAL(10, 6),
    latency_ms INTEGER,
    status_code INTEGER,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_requests_created_at ON requests(created_at);
CREATE INDEX idx_requests_api_key ON requests(api_key_id);
```

---

## 🐳 Deployment

### Coolify Deployment

1. **Create New Application** in Coolify
2. **Connect Repository** (GitHub/GitLab)
3. **Set Environment Variables**:
   ```
   DATABASE_URL=postgres://user:pass@postgres:5432/ai_gateway
   REDIS_URL=redis://redis:6379
   JWT_SECRET=<64-char-random-string>
   ENCRYPTION_KEY=<32-byte-key>
   ```
4. **Deploy** - Coolify will build and deploy automatically

### Docker Compose

```yaml
version: '3.8'

services:
  gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=ai_gateway
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

---

## 📖 API Design

### Unified Chat Endpoint

```http
POST /v1/chat
Authorization: Bearer <api_key>

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000
}
```

### Response

```json
{
  "id": "chat-123",
  "model": "gpt-4",
  "provider": "openai",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you?"
    }
  }],
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 10,
    "total_tokens": 15
  },
  "cost": 0.0003,
  "latency_ms": 250
}
```

### Admin Endpoints

```http
# List providers
GET /admin/providers

# Add provider
POST /admin/providers
{
  "name": "openai",
  "api_key": "sk-...",
  "base_url": "https://api.openai.com/v1"
}

# Create API key
POST /admin/api-keys
{
  "name": "Production Key",
  "rate_limit_per_minute": 100
}
```

---

## 🧪 Testing

Comprehensive testing is critical for maintaining code quality and ensuring reliability. We follow a multi-layered testing strategy.

### Testing Strategy

```
┌─────────────────────────────────────┐
│     End-to-End Tests (E2E)          │  ← Full system integration
├─────────────────────────────────────┤
│     Integration Tests                │  ← Component integration
├─────────────────────────────────────┤
│     Unit Tests                       │  ← Individual functions/units
└─────────────────────────────────────┘
```

### 1. Unit Testing

Unit tests test individual functions, methods, and components in isolation.

#### Best Practices

- **Test one thing at a time**: Each test should verify one behavior
- **Use table-driven tests**: For testing multiple scenarios
- **Mock external dependencies**: Use interfaces and mocks
- **Fast execution**: Unit tests should run in milliseconds
- **No external dependencies**: No database, network, or file system access

#### Example: Unit Test for Cost Calculation

```go
// internal/monitoring/cost_test.go
package monitoring

import (
    "testing"
)

func TestCalculateCost(t *testing.T) {
    tests := []struct {
        name     string
        tokens   int
        model    string
        want     float64
        wantErr  bool
        errMsg   string
    }{
        {
            name:    "gpt-4 calculation",
            tokens:  1000,
            model:   "gpt-4",
            want:    0.03,
            wantErr: false,
        },
        {
            name:    "gpt-3.5-turbo calculation",
            tokens:  1000,
            model:   "gpt-3.5-turbo",
            want:    0.002,
            wantErr: false,
        },
        {
            name:    "invalid model",
            tokens:  1000,
            model:   "invalid-model",
            want:    0,
            wantErr: true,
            errMsg:  "unknown model",
        },
        {
            name:    "zero tokens",
            tokens:  0,
            model:   "gpt-4",
            want:    0,
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateCost(tt.tokens, tt.model)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CalculateCost() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if err != nil && tt.errMsg != "" {
                if !contains(err.Error(), tt.errMsg) {
                    t.Errorf("CalculateCost() error = %v, want error containing %v", err, tt.errMsg)
                }
            }
            
            if got != tt.want {
                t.Errorf("CalculateCost() = %v, want %v", got, tt.want)
            }
        })
    }
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
        (len(s) > len(substr) && contains(s[1:], substr)))
}
```

#### Example: Unit Test with Mocking

```go
// internal/providers/openai_test.go
package providers

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
    mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    args := m.Called(req)
    return args.Get(0).(*http.Response), args.Error(1)
}

func TestOpenAIProvider_Chat(t *testing.T) {
    // Create mock server
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/v1/chat/completions", r.URL.Path)
        assert.Equal(t, "POST", r.Method)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"choices":[{"message":{"content":"Hello!"}}]}`))
    }))
    defer mockServer.Close()
    
    // Create provider with test server
    provider := &OpenAIProvider{
        BaseProvider: NewBaseProvider("openai", logger),
        apiKey:       "test-key",
        baseURL:      mockServer.URL,
    }
    
    // Test
    ctx := context.Background()
    req := ChatRequest{
        Model: "gpt-4",
        Messages: []Message{
            {Role: "user", Content: "Hello"},
        },
    }
    
    resp, err := provider.Chat(ctx, req)
    
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Equal(t, "Hello!", resp.Choices[0].Message.Content)
}
```

### 2. Integration Testing

Integration tests verify that multiple components work together correctly.

#### Best Practices

- **Test component interactions**: Verify components integrate correctly
- **Use test databases**: Separate test database for integration tests
- **Clean up after tests**: Ensure tests don't leave state
- **Test real scenarios**: Use realistic data and scenarios

#### Example: Integration Test for API Handler

```go
// internal/api/handlers/chat_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestChatHandler_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Setup test Redis
    redis := setupTestRedis(t)
    defer redis.Close()
    
    // Create handler with real dependencies
    handler := NewChatHandler(db, redis, logger)
    
    // Setup router
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/v1/chat", handler.HandleChat)
    
    // Create request
    reqBody := ChatRequest{
        Model: "gpt-4",
        Messages: []Message{
            {Role: "user", Content: "Hello"},
        },
    }
    body, _ := json.Marshal(reqBody)
    
    req := httptest.NewRequest("POST", "/v1/chat", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-api-key")
    
    // Execute request
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(t, http.StatusOK, w.Code)
    
    var resp ChatResponse
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)
    assert.NotEmpty(t, resp.ID)
    assert.NotEmpty(t, resp.Choices)
}
```

### 3. End-to-End (E2E) Testing

E2E tests verify the entire system works from user perspective.

#### Example: E2E Test

```go
// tests/e2e/chat_test.go
package e2e

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestChatE2E(t *testing.T) {
    // Start test server (or use test environment)
    baseURL := "http://localhost:8080"
    
    // 1. Create API key
    apiKey := createAPIKey(t, baseURL)
    
    // 2. Send chat request
    reqBody := map[string]interface{}{
        "model": "gpt-4",
        "messages": []map[string]string{
            {"role": "user", "content": "Hello, world!"},
        },
    }
    
    body, _ := json.Marshal(reqBody)
    req, _ := http.NewRequest("POST", baseURL+"/v1/chat", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    assert.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 3. Verify response
    var chatResp ChatResponse
    json.NewDecoder(resp.Body).Decode(&chatResp)
    assert.NotEmpty(t, chatResp.ID)
    assert.NotEmpty(t, chatResp.Choices)
}
```

### 4. Test Organization

```
uniroute/
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── chat.go
│   │   │   └── chat_test.go        # Unit tests
│   │   └── handlers_test.go        # Integration tests
│   ├── providers/
│   │   ├── openai.go
│   │   └── openai_test.go          # Unit tests with mocks
│   └── gateway/
│       ├── router.go
│       └── router_test.go          # Unit tests
├── tests/
│   ├── integration/                # Integration tests
│   │   ├── api_test.go
│   │   └── providers_test.go
│   ├── e2e/                        # End-to-end tests
│   │   ├── chat_test.go
│   │   └── auth_test.go
│   └── fixtures/                   # Test data
│       ├── requests.json
│       └── responses.json
└── testdata/                       # Test data files
    ├── sample_requests.json
    └── sample_responses.json
```

### 5. Test Utilities and Helpers

#### Test Helpers

```go
// internal/testutil/helpers.go
package testutil

import (
    "database/sql"
    "testing"
    
    _ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgres://test:test@localhost/testdb?sslmode=disable")
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    
    // Run migrations
    runMigrations(t, db)
    
    return db
}

// CleanupTestDB cleans up test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
    // Truncate tables
    tables := []string{"requests", "api_keys", "providers", "users"}
    for _, table := range tables {
        db.Exec("TRUNCATE TABLE " + table + " CASCADE")
    }
    db.Close()
}

// CreateTestAPIKey creates a test API key
func CreateTestAPIKey(t *testing.T, db *sql.DB, userID string) string {
    // Implementation
    return "test-api-key"
}
```

### 6. Running Tests

#### Run All Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests in verbose mode
go test -v ./...

# Run specific test
go test -v ./internal/providers -run TestOpenAIProvider_Chat

# Run tests with race detector
go test -race ./...
```

#### Test Coverage Goals

- **Unit Tests**: >80% coverage
- **Integration Tests**: >60% coverage of integration paths
- **Critical Paths**: 100% coverage (auth, routing, cost calculation)

#### Makefile Test Commands

```bash
make test          # Run all tests
make test-unit     # Run unit tests only
make test-integration  # Run integration tests
make test-e2e      # Run E2E tests
make test-coverage  # Generate coverage report
make test-race     # Run tests with race detector
```

### 7. Mocking and Test Doubles

#### Using Interfaces for Testability

```go
// internal/providers/interface.go
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// internal/providers/mock.go (for testing)
type MockProvider struct {
    ChatFunc func(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

func (m *MockProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    if m.ChatFunc != nil {
        return m.ChatFunc(ctx, req)
    }
    return nil, errors.New("not implemented")
}

// Usage in tests
func TestRouter_WithMockProvider(t *testing.T) {
    mockProvider := &MockProvider{
        ChatFunc: func(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
            return &ChatResponse{
                ID: "test-123",
                Choices: []Choice{{Message: Message{Content: "Mock response"}}},
            }, nil
        },
    }
    
    router := NewRouter(mockProvider)
    // Test router with mock
}
```

#### Using testify/mock

```go
// Install: go get github.com/stretchr/testify/mock

import (
    "github.com/stretchr/testify/mock"
)

type MockProvider struct {
    mock.Mock
}

func (m *MockProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*ChatResponse), args.Error(1)
}

// Usage
func TestSomething(t *testing.T) {
    mockProvider := new(MockProvider)
    mockProvider.On("Chat", mock.Anything, mock.Anything).
        Return(&ChatResponse{ID: "test"}, nil)
    
    // Use mockProvider in test
    mockProvider.AssertExpectations(t)
}
```

### 8. Test Data Management

#### Using Test Fixtures

```go
// tests/fixtures/requests.go
package fixtures

func GetChatRequest() ChatRequest {
    return ChatRequest{
        Model: "gpt-4",
        Messages: []Message{
            {Role: "user", Content: "Hello"},
        },
    }
}

// Usage in tests
func TestHandler(t *testing.T) {
    req := fixtures.GetChatRequest()
    // Use req in test
}
```

#### Using testdata Directory

```go
// testdata/sample_request.json
{
  "model": "gpt-4",
  "messages": [{"role": "user", "content": "Hello"}]
}

// In test
func TestWithTestData(t *testing.T) {
    data, err := os.ReadFile("testdata/sample_request.json")
    require.NoError(t, err)
    
    var req ChatRequest
    json.Unmarshal(data, &req)
    // Use req in test
}
```

### 9. CI/CD Testing

#### GitHub Actions Example

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: make test-unit
      
      - name: Run integration tests
        run: make test-integration
        env:
          DATABASE_URL: postgres://test:test@localhost/test
          REDIS_URL: redis://localhost:6379
      
      - name: Generate coverage
        run: make test-coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

### 10. Security Testing

```bash
# Run security scans
gosec ./...
govulncheck ./...

# Check for secrets in code
git secrets --scan

# Dependency vulnerability scan
go list -json -deps | nancy sleuth

# Run security tests
go test -tags=security ./internal/security/...
```

### 11. Load Testing

```bash
# Test with Apache Bench
ab -n 1000 -c 10 -H "Authorization: Bearer YOUR_API_KEY" \
   -H "Content-Type: application/json" \
   -p testdata/request.json \
   https://api.uniroute.pages.dev/v1/chat

# Test with k6
k6 run loadtest.js

# Test with wrk
wrk -t4 -c100 -d30s -s script.lua http://localhost:8080/v1/chat
```

### 12. Performance Testing

```go
// internal/gateway/benchmark_test.go
func BenchmarkRouter_SelectProvider(b *testing.B) {
    router := setupRouter()
    req := createTestRequest()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        router.SelectProvider(context.Background(), req)
    }
}

// Run: go test -bench=. -benchmem
```

### Testing Checklist

Before submitting code, ensure:

- [ ] **Unit tests** written for new functions/methods
- [ ] **Integration tests** for component interactions
- [ ] **E2E tests** for critical user flows
- [ ] **Test coverage** >80% for new code
- [ ] **Mocks** used for external dependencies
- [ ] **Test data** properly managed (fixtures/testdata)
- [ ] **Tests pass** locally before committing
- [ ] **CI/CD** tests pass
- [ ] **Performance tests** for critical paths
- [ ] **Security tests** run and pass

### Test Commands Reference

```bash
# Unit tests
go test ./internal/... -short

# Integration tests
go test ./tests/integration/...

# E2E tests
go test ./tests/e2e/...

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detector
go test -race ./...

# Benchmarks
go test -bench=. -benchmem ./...

# All tests
make test
```

---

## 📚 Documentation

### Essential Documents

1. **START_HERE.md** (this file) - Complete project guide
2. **AI_GATEWAY_PROJECT_PLAN.md** - Detailed technical plan
3. **AI_GATEWAY_SECURITY.md** - Complete security guide
4. **AI_GATEWAY_QUICKSTART.md** - Quick start with code examples
5. **COMPETITIVE_ANALYSIS.md** - Market analysis
6. **COST_ANALYSIS.md** - Cost breakdown
7. **NGROK_FEATURE_COMPARISON.md** - Feature comparison
8. **SECURITY_FIRST.md** - Security checklist

---

## ✅ Pre-Launch Checklist

### Security
- [ ] All Day 1 security measures implemented
- [ ] Security audit completed
- [ ] Penetration testing done
- [ ] Dependencies scanned for vulnerabilities

### Functionality
- [ ] Core routing working
- [ ] Multiple providers integrated
- [ ] Authentication working
- [ ] Rate limiting working
- [ ] Monitoring configured

### Developer Experience
- [ ] CLI tool functional
- [ ] SDKs available (at least Go)
- [ ] Documentation complete
- [ ] Examples provided

### Infrastructure
- [ ] HTTPS configured
- [ ] Database migrations run
- [ ] Monitoring set up
- [ ] Backup plan in place

---

## 🎯 Success Metrics

- **Latency**: < 100ms gateway overhead
- **Throughput**: 1000+ requests/second
- **Uptime**: 99.9% availability
- **Cost Tracking**: Accurate to $0.01
- **Provider Failover**: < 1 second

---

## 🤝 Contributing

We prioritize **clean, reusable code**. Please follow these guidelines:

### Contribution Guidelines

1. **Follow Go code style**:
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
   - Run `golangci-lint` before committing

2. **Write clean, reusable code**:
   - Follow Single Responsibility Principle
   - Extract common patterns into reusable functions
   - Use interfaces for abstraction
   - Keep functions small and focused (< 50 lines)
   - See [Code Quality & Best Practices](#code-quality--best-practices) section

3. **Write tests**:
   - Unit tests for new features
   - Aim for >80% code coverage
   - Use table-driven tests where appropriate

4. **Update documentation**:
   - Document public APIs with godoc comments
   - Update README if needed
   - Add examples for new features

5. **Use conventional commits**:
   ```
   feat: add new provider support
   fix: resolve rate limiting bug
   refactor: extract common error handling
   docs: update API documentation
   test: add tests for provider factory
   ```

### Code Review Focus

When reviewing code, we focus on:
- **Reusability**: Can this code be reused elsewhere?
- **Maintainability**: Is the code easy to understand and modify?
- **Testability**: Can this be easily tested?
- **Performance**: Are there any obvious performance issues?
- **Security**: Any security concerns?

### Before Submitting

Run these commands before submitting a PR:

```bash
make fmt      # Format code
make lint      # Run linters
make test      # Run tests
make vet       # Run go vet
make security  # Security scan
```

See [Code Quality & Best Practices](#code-quality--best-practices) for detailed guidelines.

---

## 📞 Support & Resources

- **GitHub**: https://github.com/uniroute/ai-gateway
- **Documentation**: https://uniroute.pages.dev
- **Status**: https://status.uniroute.pages.dev
- **Discord**: [Join our community]

---

## 🚦 Next Steps

1. **Review this document** - Understand the full scope
2. **Set up development environment** - Go, PostgreSQL, Redis
3. **Clone and initialize** - Follow Quick Start section
4. **Start with Phase 1** - Core gateway implementation
5. **Follow security checklist** - Never skip security
6. **Deploy to Coolify** - Test in production-like environment

---

## 📝 Quick Reference

### Environment Variables

```env
# Server
PORT=8080

# Database
DATABASE_URL=postgres://user:pass@localhost/ai_gateway

# Redis
REDIS_URL=redis://localhost:6379

# Security
JWT_SECRET=<64-char-random-string>
ENCRYPTION_KEY=<32-byte-key>

# Rate Limiting
RATE_LIMIT_PER_MINUTE=60
RATE_LIMIT_PER_DAY=10000

# Provider API Keys (users provide their own)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
```

### Makefile Commands

```bash
make dev          # Start development server
make build        # Build binary
make test         # Run tests
make migrate      # Run database migrations
make lint         # Run linters
make security     # Run security scans
```

### CLI Commands (Future)

```bash
# Server Management
uniroute start              # Start gateway server
uniroute status             # Check server status
uniroute logs --follow      # View live logs

# Tunneling (ngrok-like)
uniroute tunnel --port 8080  # Expose local server to internet
# Returns: https://abc123.uniroute.dev -> http://localhost:8080
# Web UI: http://127.0.0.1:4040

# API Key Management
uniroute keys create        # Create API key
uniroute keys list          # List API keys
uniroute keys revoke <key>  # Revoke API key

# Provider Management
uniroute providers list     # List available providers
uniroute providers add      # Add new provider
```

### Tunneling with External Tools (Until Phase 6)

```bash
# Using ngrok
ngrok http 8080
# Returns: https://abc123.ngrok-free.app -> http://localhost:8080

# Using cloudflared (free, no signup)
cloudflared tunnel --url http://localhost:8080
# Returns: https://random-subdomain.trycloudflare.com
```

### Clean Code Quick Checklist

Before committing code, verify:

- ✅ **Single Purpose?** Does this function/struct do one thing?
- ✅ **Reusable?** Can this be extracted and reused?
- ✅ **Testable?** Can I easily write tests for this?
- ✅ **Documented?** Is the code self-explanatory or documented?
- ✅ **No Duplication?** Is this logic already implemented elsewhere?
- ✅ **Interface-Based?** Am I using interfaces for abstraction?
- ✅ **Error Handling?** Are all errors handled appropriately?
- ✅ **Small Functions?** Is each function < 50 lines?

Run: `make fmt lint test vet security` before committing.

---

## 🎉 Ready to Build!

**You now have everything you need to start building UniRoute!**

- ✅ Complete architecture
- ✅ Technology stack defined
- ✅ Security requirements
- ✅ Implementation plan
- ✅ Cost analysis
- ✅ Competitive positioning

**Remember**: 
- Security first - Never deploy without security measures
- Start simple - MVP first, then iterate
- Community matters - Open source thrives on contributions
- Free forever - Keep it accessible to everyone

**Let's build the future of AI routing! 🚀**

---

**Last Updated**: 2025-01-XX  
**Version**: 1.0  
**Status**: Ready for Development

