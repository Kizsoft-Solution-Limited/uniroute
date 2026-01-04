# 🚀 UniRoute

**One unified gateway for every AI model. Route, secure, and manage traffic to any LLM—cloud or local—with one unified platform.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![Status](https://img.shields.io/badge/Status-In%20Development-orange)](https://github.com/Kizsoft-Solution-Limited/uniroute)

**100% FREE • Open Source • Self-Hostable**

---

## 📖 Overview

UniRoute is a unified gateway platform that routes, secures, and manages traffic to any LLM (cloud or local) with one unified API. Think of it as **ngrok for AI models** - a single entry point that intelligently routes requests to the best available model.

### Why UniRoute?

- ✅ **100% Free** - No pricing, no limits for users
- ✅ **Open Source** - Full transparency, community-driven
- ✅ **Self-Hostable** - Deploy anywhere, full control
- ✅ **Local LLM First** - Priority support for local models (Ollama, vLLM) - Free & Private
- ✅ **Multi-Provider** - Support for OpenAI, Anthropic, Google, Local LLMs, and more
- ✅ **Shareable** - Host and share your gateway with others (ngrok-like tunneling)
- ✅ **Intelligent Routing** - Load balancing, failover, cost-based routing
- ✅ **Enterprise Security** - API keys, rate limiting, Zero Trust support

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+ (or [Supabase](https://supabase.com) free tier)
- Redis 7+ (or [Upstash](https://upstash.com) free tier)
- Docker (optional, for containerized deployment)

### Installation

**Option 1: Download Pre-built Binary (Easiest)** ⭐

```bash
# Download for your platform (see CLI_INSTALLATION.md for all platforms)
# macOS (Apple Silicon):
curl -L https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-arm64 -o uniroute
chmod +x uniroute
sudo mv uniroute /usr/local/bin/

# Verify
uniroute --version

# Login to your account
uniroute auth login

# Start using UniRoute!
uniroute projects list
```

**Option 2: Build from Source**

```bash
# Clone the repository
git clone https://github.com/Kizsoft-Solution-Limited/uniroute.git
cd uniroute

# Install dependencies
go mod download

# Build binaries
make build

# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Run database migrations
make migrate

# Start the server
make dev
```

See [CLI_INSTALLATION.md](./CLI_INSTALLATION.md) for detailed installation instructions.

### Verify Installation

```bash
# Check health endpoint
curl http://localhost:8084/health
```

### First Request (Local LLM)

```bash
# With Ollama running locally
curl -X POST http://localhost:8084/v1/chat \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

### Expose to Internet

**Option 1: Using cloudflared (100% free, no signup) - Recommended**
```bash
cloudflared tunnel --url http://localhost:8084
# Returns: https://random-subdomain.trycloudflare.com
```

**Option 2: Using ngrok (free tier, requires signup)**
```bash
ngrok http 8084
# Returns: https://abc123.ngrok-free.app -> http://localhost:8084
```

**Option 3: Built-in UniRoute tunnel (requires CLI installation)**
```bash
# Download CLI (see CLI_INSTALLATION.md)
# Or build: make build

# Expose your local app (any port, any app)
uniroute tunnel --port 8084
# Returns: http://{subdomain}.uniroute.dev -> http://localhost:8084

# Works with any local application, not just UniRoute!
```

---

## ✨ Features

### Core Features

- **Unified API Interface** - Single endpoint for all LLM providers
- **Intelligent Routing** - Model selection based on cost, latency, availability
- **Load Balancing** - Distribute traffic across multiple instances
- **Automatic Failover** - Seamless switching when providers fail
- **Security & Access Control** - API keys, JWT, rate limiting
- **Monitoring & Analytics** - Usage tracking, cost tracking, performance metrics
- **Multi-Provider Support** - OpenAI, Anthropic, Google, Cohere, Local LLMs
- **Developer Experience** - CLI tool, SDKs, OpenAPI docs

### Supported Providers

- 🏠 **Local LLMs (Ollama, vLLM)** ⭐ Priority - Free, private, self-hosted
- 🤖 OpenAI (GPT-4, GPT-3.5, etc.) - Phase 3
- 🧠 Anthropic (Claude) - Phase 3
- 🔷 Google (Gemini) - Phase 3
- 📊 Cohere - Phase 3
- ➕ Custom providers

---

## 🏗️ Architecture

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
│  │  Security Layer          │ │
│  │  Monitoring & Analytics  │ │
│  └───────────────────────────┘ │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│   Model Providers (Backends)   │
│  OpenAI │ Anthropic │ Local   │
└─────────────────────────────────┘
```

---

## 💻 Technology Stack

- **Backend**: Go 1.21+
- **API Framework**: Gin
- **Database**: PostgreSQL + Redis
- **Authentication**: JWT + API Keys
- **Monitoring**: Prometheus + Grafana
- **Logging**: Structured logging (zerolog)

---

## 📚 Documentation

- **[SECURITY_OVERVIEW.md](SECURITY_OVERVIEW.md)** - 🔐 Complete security documentation and measures

For complete documentation, see **[START_HERE.md](./START_HERE.md)** which includes:

- 📋 Complete project overview
- 🏗️ Detailed architecture
- 🔐 Security requirements
- 🚀 Implementation plan
- 💰 Cost analysis
- 🆚 Competitive analysis
- 📖 API design
- 🐳 Deployment guides

---

## 🔐 Security

**⚠️ CRITICAL: Security First**

UniRoute implements enterprise-grade security:

- ✅ Input validation and sanitization
- ✅ API keys with bcrypt hashing
- ✅ JWT authentication with strong secrets
- ✅ Rate limiting (per-key, per-IP)
- ✅ HTTPS/TLS enforcement
- ✅ Security headers (CSP, HSTS, etc.)
- ✅ Parameterized queries (SQL injection prevention)
- ✅ Encrypted secrets at rest

See [START_HERE.md](./START_HERE.md#security-requirements) for the complete security checklist.

---

## 🐳 Deployment

### Docker Compose

```bash
docker-compose up -d
```

### Coolify

1. Create new application in Coolify
2. Connect GitHub repository
3. Set environment variables
4. Deploy automatically

See [START_HERE.md](./START_HERE.md#deployment) for detailed deployment instructions.

---

## 🧪 Development

### Project Structure

```
uniroute/
├── cmd/
│   └── gateway/          # Application entry point
├── internal/
│   ├── api/              # HTTP handlers & middleware
│   ├── gateway/          # Routing logic
│   ├── providers/        # LLM provider implementations
│   ├── security/         # Auth & rate limiting
│   ├── storage/          # Database clients
│   └── monitoring/       # Metrics & analytics
├── migrations/           # Database migrations
└── pkg/                  # Shared packages (reusable utilities)
```

### Makefile Commands

```bash
make dev          # Start development server
make build        # Build binary
make test         # Run tests
make migrate      # Run database migrations
make lint         # Run linters
make security     # Run security scans
make fmt          # Format code
make vet          # Run go vet
```

---

## 🧹 Code Quality & Best Practices

### Clean Code Principles

We follow **clean code principles** to ensure maintainability, readability, and reusability:

#### 1. **Single Responsibility Principle (SRP)**
Each function, struct, and package should have one clear purpose.

```go
// ✅ Good: Single responsibility
func ValidateAPIKey(key string) error {
    // Only validates API key format
}

// ❌ Bad: Multiple responsibilities
func ProcessRequestAndValidateAndLog(key string, req Request) error {
    // Too many responsibilities
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
Use interfaces for abstraction and testability.

```go
// ✅ Good: Provider interface for reusability
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    HealthCheck(ctx context.Context) error
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
    client *http.Client
    logger *zerolog.Logger
}

type OpenAIProvider struct {
    BaseProvider
    apiKey string
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Reuse BaseProvider's client and logger
    return p.makeRequest(ctx, req)
}
```

#### 5. **Error Handling**
Consistent, informative error handling.

```go
// ✅ Good: Wrapped errors with context
var (
    ErrProviderUnavailable = errors.New("provider unavailable")
    ErrRateLimitExceeded   = errors.New("rate limit exceeded")
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
All LLM providers implement a common interface for easy swapping:

```go
// pkg/providers/interface.go
type Provider interface {
    Name() string
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
    HealthCheck(ctx context.Context) error
    GetModels() []string
}

// Easy to add new providers
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
}

// PostgreSQL implementation
type PostgresAPIKeyRepository struct {
    db *sql.DB
}

// Redis cache implementation
type CachedAPIKeyRepository struct {
    repo APIKeyRepository
    cache *redis.Client
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
type CostBasedStrategy struct {}
type LatencyBasedStrategy struct {}
type RoundRobinStrategy struct {}
type FailoverStrategy struct {}

// Easy to swap strategies
router := NewRouter(CostBasedStrategy{})
```

#### 5. **Factory Pattern**
Centralized creation of providers and services:

```go
// internal/providers/factory.go
type ProviderFactory struct {
    config *Config
    logger *zerolog.Logger
}

func (f *ProviderFactory) CreateProvider(name string) (Provider, error) {
    switch name {
    case "openai":
        return NewOpenAIProvider(f.config.OpenAI, f.logger)
    case "anthropic":
        return NewAnthropicProvider(f.config.Anthropic, f.logger)
    case "local":
        return NewLocalLLMProvider(f.config.Local, f.logger)
    default:
        return nil, ErrUnknownProvider
    }
}
```

### Code Organization Guidelines

#### Package Structure
- **`pkg/`**: Reusable packages that can be imported by other projects
- **`internal/`**: Private application code, not importable by external projects
- **`cmd/`**: Application entry points

#### Naming Conventions
```go
// ✅ Good naming
type APIKeyService struct {}        // Clear, descriptive
func ValidateRequest(req Request)   // Verb + noun
const MaxRetries = 3                // Constants in PascalCase

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
        return 0, err
    }
    return float64(tokens) * rate, nil
}

// ❌ Bad: Too many responsibilities
func ProcessEverything(req Request) (*Response, error) {
    // 200 lines of mixed logic
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
        {"gpt-4", 1000, "gpt-4", 0.03, false},
        {"invalid model", 1000, "invalid", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateCost(tt.tokens, tt.model)
            // assertions...
        })
    }
}
```

### Code Review Checklist

Before submitting code, ensure:

- [ ] **Single Responsibility**: Each function/struct has one clear purpose
- [ ] **DRY**: No code duplication; common logic extracted
- [ ] **Interfaces**: Used for abstraction and testability
- [ ] **Error Handling**: All errors handled appropriately
- [ ] **Tests**: Unit tests for new functionality
- [ ] **Documentation**: Public functions/types documented
- [ ] **Naming**: Clear, descriptive names
- [ ] **Formatting**: Code formatted with `gofmt`
- [ ] **Linting**: No linter warnings
- [ ] **Security**: No security vulnerabilities introduced

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
type BaseProvider struct {
    name     string
    client   *http.Client
    logger   *zerolog.Logger
    metrics  *prometheus.CounterVec
}

func (b *BaseProvider) makeRequest(ctx context.Context, url string, body interface{}) (*http.Response, error) {
    // Reusable HTTP request logic
    // Logging, metrics, retries, etc.
}

// internal/providers/openai.go - Specific implementation
type OpenAIProvider struct {
    BaseProvider
    apiKey string
    baseURL string
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Uses BaseProvider.makeRequest() for HTTP calls
    return p.makeRequest(ctx, p.baseURL+"/chat", req)
}
```

---

## 📝 Code Standards

### Go Style Guide
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (enforced in CI)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Required Tools
```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run before committing
make fmt      # Format code
make lint     # Run linters
make vet      # Run go vet
make security # Security scan
```

---

## 📊 API Examples

### Chat Completion (Local LLM)

```bash
# Using local Ollama instance
curl -X POST http://localhost:8084/v1/chat \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Explain quantum computing"}
    ],
    "temperature": 0.7,
    "max_tokens": 1000
  }'
```

### Response

```json
{
  "id": "chat-123",
  "model": "llama2",
  "provider": "local",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Quantum computing is a type of computation..."
    }
  }],
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 10,
    "total_tokens": 15
  },
  "cost": 0.0,
  "latency_ms": 250
}
```

### Share Your Gateway Server

```bash
# 1. Start Ollama
ollama serve

# 2. Start UniRoute
make dev

# 3. Expose to internet (using cloudflared - 100% free, no signup)
cloudflared tunnel --url http://localhost:8084
# Share: https://random-subdomain.trycloudflare.com

# 4. Others can now use your gateway!
curl -X POST https://random-subdomain.trycloudflare.com/v1/chat \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "messages": [...]}'
```

---

## 🤝 Contributing

Contributions are welcome! We prioritize **clean, reusable code**. Please follow these guidelines:

### Contribution Workflow

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Write clean, reusable code**:
   - Follow Single Responsibility Principle
   - Extract common patterns into reusable functions
   - Use interfaces for abstraction
   - Keep functions small and focused
4. **Write tests** for new features (aim for >80% coverage)
5. **Update documentation** (code comments, README, etc.)
6. **Run quality checks**:
   ```bash
   make fmt      # Format code
   make lint     # Run linters
   make test     # Run tests
   make security # Security scan
   ```
7. **Commit using conventional commits**:
   ```
   feat: add new provider support
   fix: resolve rate limiting bug
   refactor: extract common error handling
   docs: update API documentation
   ```
8. **Push to the branch** (`git push origin feature/amazing-feature`)
9. **Open a Pull Request** with clear description

### Code Contribution Guidelines

#### Before You Start
- Check existing code patterns and follow them
- Look for similar functionality that can be reused
- Discuss major changes in an issue first

#### While Coding
- **Reuse existing code**: Check `pkg/` and `internal/` for reusable utilities
- **Create reusable components**: If you write something that could be reused, put it in `pkg/`
- **Use interfaces**: Abstract dependencies for testability
- **Write small functions**: Max 50 lines, ideally 20-30
- **Document public APIs**: All exported functions/types need godoc comments

#### Code Review Focus
- **Reusability**: Can this code be reused elsewhere?
- **Maintainability**: Is the code easy to understand and modify?
- **Testability**: Can this be easily tested?
- **Performance**: Are there any obvious performance issues?
- **Security**: Any security concerns?

### Example: Good Contribution

```go
// ✅ Good: Reusable, well-documented, testable
// Package providers contains reusable provider implementations.
package providers

// Provider defines the interface for all LLM providers.
// This interface allows easy swapping of providers and testing.
type Provider interface {
    // Chat sends a chat request to the provider.
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    
    // HealthCheck verifies the provider is available.
    HealthCheck(ctx context.Context) error
}

// NewProviderFactory creates a new provider factory with the given config.
// This factory pattern allows centralized provider creation.
func NewProviderFactory(config *Config) *ProviderFactory {
    return &ProviderFactory{config: config}
}
```

### Example: Bad Contribution

```go
// ❌ Bad: Not reusable, no documentation, hard to test
func doStuff(req map[string]interface{}) map[string]interface{} {
    // 200 lines of mixed logic
    // Hardcoded values
    // No error handling
    // Can't be reused
}
```

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- **GitHub**: https://github.com/Kizsoft-Solution-Limited/uniroute
- **Documentation**: See [START_HERE.md](./START_HERE.md)
- **Issues**: https://github.com/Kizsoft-Solution-Limited/uniroute/issues

---

## 🎯 Roadmap

**Current Focus: Phase 1 - Local LLM Support** 🏠

- **Phase 1**: Local LLM provider + Shareable server ✅
- **Phase 2**: Security & Rate Limiting ✅
- **Phase 3**: Cloud providers (OpenAI, Anthropic, Google) ✅
- **Phase 4**: Advanced Routing
- **Phase 5**: Monitoring & Analytics
- **Phase 6**: Developer Experience (CLI, SDKs, Built-in tunneling)

**Development Approach:**
- ✅ Complete, test, and verify each phase before moving forward
- ✅ Iterative development with testing checkpoints
- ✅ Each phase is production-ready before proceeding

See [START_HERE.md](./START_HERE.md#implementation-plan) for detailed implementation plan.

---

## ⭐ Star History

If you find UniRoute useful, please consider giving it a star ⭐ on GitHub!

---

## 🙏 Acknowledgments

- Inspired by [ngrok's AI Gateway](https://ngrok.com/)
- Built with ❤️ by the open-source community

---

## 📋 Quick Reference: Clean Code Checklist

### Before Committing Code

```bash
# Run all checks
make fmt      # Format with gofmt
make lint     # Run linters
make test     # Run tests
make vet      # Run go vet
make security # Security scan
```

### Code Quality Questions

Ask yourself:
- ✅ **Single Purpose?** Does this function/struct do one thing?
- ✅ **Reusable?** Can this be extracted and reused?
- ✅ **Testable?** Can I easily write tests for this?
- ✅ **Documented?** Is the code self-explanatory or documented?
- ✅ **No Duplication?** Is this logic already implemented elsewhere?
- ✅ **Interface-Based?** Am I using interfaces for abstraction?
- ✅ **Error Handling?** Are all errors handled appropriately?
- ✅ **Small Functions?** Is each function < 50 lines?

### Reusability Patterns

| Pattern | Use Case | Location |
|---------|----------|----------|
| **Interface** | Provider abstraction | `internal/providers/interface.go` |
| **Middleware** | Cross-cutting concerns | `internal/api/middleware/` |
| **Repository** | Data access abstraction | `internal/storage/repository.go` |
| **Strategy** | Pluggable algorithms | `internal/gateway/strategy.go` |
| **Factory** | Object creation | `internal/providers/factory.go` |
| **Base Struct** | Shared functionality | `internal/providers/base.go` |

### Common Reusable Components

- **Error Handling**: `pkg/errors/` - Standardized error types
- **Logging**: `pkg/logger/` - Structured logging utilities
- **HTTP Client**: `internal/providers/base.go` - Reusable HTTP client
- **Validation**: `internal/api/validators/` - Request validation
- **Metrics**: `internal/monitoring/metrics.go` - Prometheus metrics

---

**Made with ❤️ by [Kizsoft Solution Limited](https://github.com/Kizsoft-Solution-Limited)**

