# SimpleHttp Project Context

**Version:** 0.1.0
**Organization:** medatechnology
**Repository:** simplehttp
**Language:** Go (1.23.2+)
**Last Updated:** 2026-01-28

## Project Overview

SimpleHttp is a **framework-agnostic HTTP handler library** for Go that provides a unified abstraction layer across multiple web frameworks (Echo, Fiber, FastHTTP). The core design philosophy is **Write Once, Run Anywhere** - allowing developers to write web applications once and switch between underlying frameworks without code changes.

### Key Capabilities
- Multi-framework support (Echo v5, Fiber v2, FastHTTP)
- Modular middleware system
- **Unified Path Parameter Extraction** (New in v0.1.0)
- **Cookie Management** (New in v0.1.0)
- **Form Data Handling** (New in v0.1.0)
- WebSocket support
- File handling (upload/download)
- Server-Sent Events (SSE)
- HTTP client package
- Comprehensive request/response abstractions

---

## Architecture Overview

### Core Design Pattern: Adapter Pattern

SimpleHttp uses the **Adapter Pattern** to provide a consistent interface (`simplehttp.Context`) across different web frameworks. Each framework has its own adapter implementation.

```
┌─────────────────────────────────────────┐
│         Application Code                │
│    (Framework-agnostic handlers)        │
└──────────────┬──────────────────────────┘
               │ Uses
               ▼
┌─────────────────────────────────────────┐
│      SimpleHttp Interface Layer         │
│  - Context interface                    │
│  - Server interface                     │
│  - Router interface                     │
│  - Middleware interface                 │
└──────────────┬──────────────────────────┘
               │ Implemented by
               ▼
┌─────────────────────────────────────────┐
│         Framework Adapters              │
│  ┌──────────┬──────────┬──────────┐    │
│  │  Echo    │  Fiber   │ FastHTTP │    │
│  │ Adapter  │ Adapter  │  Adapter │    │
│  └──────────┴──────────┴──────────┘    │
└──────────────┬──────────────────────────┘
               │ Wraps
               ▼
┌─────────────────────────────────────────┐
│      Underlying Frameworks              │
│  - Echo v5  - Fiber v2  - FastHTTP     │
└─────────────────────────────────────────┘
```

### File Structure

```
simplehttp/
├── .agent/                    # AI agent context and workflows
│   └── context.md            # This file
├── framework/                # Framework-specific adapters
│   ├── echo/
│   │   ├── server.go         # Echo server implementation
│   │   ├── context.go        # Echo context adapter
│   │   └── adapter.go        # Echo-specific utilities
│   ├── fiber/
│   │   ├── server.go         # Fiber server implementation
│   │   ├── context.go        # Fiber context adapter
│   │   ├── adapter.go        # Fiber-specific utilities
│   │   └── middleware.go     # Fiber-specific middleware
│   └── fasthttp/
│       ├── server.go         # FastHTTP server implementation
│       ├── context.go        # FastHTTP context adapter
│       └── adapter.go        # FastHTTP-specific utilities
├── client/                   # HTTP client package
│   ├── client.go             # HTTP client implementation
│   └── models.go             # Client data models
├── example/                  # Example implementations
│   ├── main.go               # Main example app
│   ├── echo.go               # Echo-specific example
│   ├── fasthttp.go           # FastHTTP-specific example
│   └── .env.example          # Example environment config
├── simplehttp.go             # Core interfaces (Context, Server, Router, etc.)
├── config.go                 # Configuration structures and loaders
├── middleware.go             # Standard middleware implementations
├── logger.go                 # Logging interface and default implementation
├── cache.go                  # Caching interface and memory cache
├── session.go                # Session management
├── security.go               # Security utilities
├── files.go                  # File handling utilities
├── streaming.go              # SSE and streaming support
├── error.go                  # Error definitions
├── helper.go                 # Helper utilities
├── internal_api.go           # Internal debug endpoints
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── README.md                 # Project documentation
├── LICENSE                   # MIT License
├── .version                  # Current version number
└── .changelog                # Version history
```

---

## Code Conventions

### 1. Naming Conventions

#### Files
- **Lowercase with underscores:** Not used in this project
- **Lowercase:** All Go files use lowercase names (e.g., `middleware.go`, `config.go`)
- **Descriptive names:** File names should clearly indicate their purpose

#### Variables and Functions
- **Exported (Public):** Use PascalCase for exported types, functions, and methods
  ```go
  type Context interface { ... }
  func LoadConfig() *Config { ... }
  func NewDefaultLogger() Logger { ... }
  ```

- **Unexported (Private):** Use camelCase for internal types and functions
  ```go
  func validateBasicAuth(auth, username, password string) bool { ... }
  func getAllowedOrigin(allowedOrigins []string, origin string) string { ... }
  ```

#### Constants
- **ALL_CAPS with underscores** for constants
  ```go
  const (
      DEFAULT_HTTP_READ_TIMEOUT  = 30
      DEFAULT_HTTP_WRITE_TIMEOUT = 30
      HEADER_AUTHORIZATION       = "Authorization"
      HEADER_REQUEST_ID          = "X-Request-ID"
  )
  ```

#### Environment Variables
- Prefix with `SIMPLEHTTP_` for all environment variables
  ```go
  SIMPLEHTTP_FRAMEWORK
  SIMPLEHTTP_PORT
  SIMPLEHTTP_APP_NAME
  SIMPLEHTTP_DEBUG
  ```

### 2. Interface Design

#### Core Pattern: Interface-First Design
All major components are defined as interfaces first, then implemented by concrete types.

**Example: Context Interface**
```go
type Context interface {
    // Request information
    GetPath() string
    GetMethod() string
    GetHeader(key string) string
    GetParam(key string) string              // Path params
    GetQueryParam(key string) string         // Query params
    GetCookie(name string) (string, error)   // Cookies
    GetFormValue(key string) string          // Form values
    
    // Response methods
    JSON(code int, data interface{}) error
    String(code int, data string) error
    Redirect(code int, url string) error     // Redirects
    
    // Status Helpers
    BadRequest(message string) error
    NotFound() error
    
    // Request binding
    BindJSON(interface{}) error
    BindForm(interface{}) error
}
```

**Implementation Pattern:**
Each framework provides its own implementation:
- `framework/echo/context.go` → `echoContext` struct
- `framework/fiber/context.go` → `fiberContext` struct
- `framework/fasthttp/context.go` → `fastContext` struct

### 3. Middleware Pattern

#### Standard Middleware Structure
```go
// Named middleware wrapper
type NamedMiddleware struct {
    name       string
    middleware MiddlewareFunc
}

// Middleware interface
type Middleware interface {
    Name() string
    Handle(HandlerFunc) HandlerFunc
}

// Middleware function signature
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Helper to create named middleware
func WithName(name string, m MiddlewareFunc) NamedMiddleware {
    return NamedMiddleware{
        name:       name,
        middleware: m,
    }
}
```

#### Built-in Middleware Convention
All built-in middleware follows this pattern:
1. Define a public factory function with `Middleware` prefix
2. Return a `Middleware` interface using `WithName`
3. Internal implementation function for the actual logic

**Example:**
```go
// Public factory
func MiddlewareRequestID() Middleware {
    return WithName("request-id", RequestID())
}

// Internal implementation
func RequestID() MiddlewareFunc {
    return func(next HandlerFunc) HandlerFunc {
        return func(c Context) error {
            // Implementation
            return next(c)
        }
    }
}
```

### 4. Configuration Pattern

#### Environment-Based Configuration
- All configuration loaded from environment variables
- `LoadConfig()` reads from ENV with defaults from `DefaultConfig`
- Use `github.com/medatechnology/goutil/utils` for ENV parsing

**Example:**
```go
config := &Config{
    Framework: utils.GetEnvString(SIMPLEHTTP_FRAMEWORK, DefaultConfig.Framework),
    Port:      utils.GetEnvString(SIMPLEHTTP_PORT, DefaultConfig.Port),
    Debug:     utils.GetEnvBool(SIMPLEHTTP_DEBUG, DefaultConfig.Debug),
}
```

#### Configuration Validation
- Always validate configuration with `ValidateConfig(config *Config) error`
- Create required directories (upload, temp)
- Set sensible defaults for missing values

### 5. Error Handling

#### Error Definition Pattern
```go
var (
    ErrInvalidConfig     = errors.New("invalid configuration")
    ErrUnsupportedMethod = errors.New("unsupported HTTP method")
)
```

#### Error Responses
- Return errors from handlers; framework adapters handle conversion to HTTP responses
- Use standard `error` interface
- Custom error handler can be set in `Config.ErrorHandler`

### 6. Documentation Standards

#### Go Doc Comments
- Every exported type, function, and constant must have a doc comment
- Doc comments should be complete sentences starting with the item name

**Examples:**
```go
// Context represents our framework-agnostic request context
type Context interface { ... }

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config { ... }

// CORSConfig defines CORS settings
type CORSConfig struct { ... }
```

#### Inline Comments
- Use inline comments sparingly for complex logic
- Prefer self-documenting code with clear variable/function names

---

## Development Patterns

### 1. Adding a New Framework Adapter

To add support for a new web framework:

1. **Create framework directory**: `framework/<framework-name>/`
2. **Implement required files:**
   - `server.go`: Implement `Server` interface
   - `context.go`: Implement `Context` adapter
   - `adapter.go`: Framework-specific utilities

3. **Key interfaces to implement:**
   ```go
   // In server.go
   type <Framework>Server struct {
       app    *<framework>.App
       config *simplehttp.Config
   }

   func NewServer(config *simplehttp.Config) simplehttp.Server {
       // Initialize framework
   }

   // In context.go
   type <framework>Context struct {
       ctx <framework>.Context
   }

   func (c *<framework>Context) GetPath() string {
       // Adapt framework context
   }
   ```

### 2. Adding New Middleware

Pattern for adding new middleware:

1. **Define configuration struct** (if needed)
   ```go
   type MyMiddlewareConfig struct {
       Setting1 string
       Setting2 int
   }
   ```

2. **Create public factory function**
   ```go
   func MiddlewareMyFeature(config MyMiddlewareConfig) Middleware {
       return WithName("my-feature", MyFeature(config))
   }
   ```

3. **Implement middleware logic**
   ```go
   func MyFeature(config MyMiddlewareConfig) MiddlewareFunc {
       return func(next HandlerFunc) HandlerFunc {
           return func(c Context) error {
               // Pre-processing
               
               err := next(c)
               
               // Post-processing
               
               return err
           }
       }
   }
   ```

### 3. Request Header Parsing Pattern

The `RequestHeader` struct centralizes header parsing:

```go
type RequestHeader struct {
    Authorization HeaderAuthorization
    RequestID     string
    UserAgent     string
    Browser       string
    RealIP        string
    RemoteIP      string
    // ... other fields
}

func (h *RequestHeader) FromHttpRequest(stdRequest *http.Request) {
    // Parse all standard headers
}
```

**Usage:**
```go
// In middleware
server.Use(simplehttp.MiddlewareHeaderParser())

// In handler
func myHandler(c simplehttp.Context) error {
    headers := c.GetHeaders()
    ip := headers.RealIP
    browser := headers.Browser
    // ...
}
```

### 4. File Handling Pattern

```go
// Setup
fileHandler := simplehttp.NewFileHandler("./uploads")
fileHandler.MaxFileSize = 50 << 20  // 50MB
fileHandler.AllowedTypes = []string{"image/jpeg", "image/png"}

// Routes
server.POST("/upload", fileHandler.HandleUpload())
server.GET("/files/:filename", fileHandler.HandleDownload("./uploads/{{filename}}"))
```

### 5. WebSocket Pattern

```go
server.WebSocket("/ws/chat", func(ws simplehttp.Websocket) error {
    for {
        msg := &Message{}
        if err := ws.ReadJSON(msg); err != nil {
            return err
        }
        
        response := &Message{
            Type: "response",
            Data: msg.Data,
        }
        
        if err := ws.WriteJSON(response); err != nil {
            return err
        }
    }
})
```

---

## Testing Conventions

### 1. Example-Based Testing
- Use `example/` directory for integration examples
- Each framework should have a working example
- Examples serve as both documentation and manual tests

### 2. Framework Switching Test
- Test the same application code with different frameworks
- Verify consistent behavior across adapters

---

## Dependency Management

### Core Dependencies
```go
require (
    github.com/labstack/echo/v5
    github.com/gofiber/fiber/v2
    github.com/valyala/fasthttp
    github.com/medatechnology/goutil
    github.com/mileusna/useragent
    github.com/gorilla/websocket
    golang.org/x/time
)
```

### Dependency Guidelines
- **Minimize dependencies**: Only add truly necessary packages
- **Version pinning**: Use specific versions in `go.mod`
- **Shared utilities**: Use `medatechnology/goutil` for common utilities (ENV parsing, etc.)

---

## Version Management

### Versioning Strategy
- **Semantic Versioning**: `major.minor.patch`
- Version stored in `.version` file
- Changelog maintained in `.changelog` file

### Version Update Pattern
```bash
# Format: <branch>.<version>\tCommit: <description>
main.0.0.3  Commit: Renaming everything Meda into SimpleHTTP
main.0.0.2  Commit: Added http client package
main.0.0.1  Commit: Init simplehttp
```

---

## Security Considerations

### 1. Input Validation
- Always validate file uploads (size, type)
- Sanitize file paths to prevent directory traversal
- Validate configuration on load

### 2. Security Middleware
```go
secConfig := simplehttp.SecurityConfig{
    FrameDeny:             true,   // Prevent clickjacking
    ContentTypeNosniff:    true,   // Prevent MIME sniffing
    BrowserXssFilter:      true,   // Enable XSS filter
    ContentSecurityPolicy: "default-src 'self'",
}
server.Use(simplehttp.MiddlewareSecurity(secConfig))
```

### 3. Authentication Patterns
- Use `MiddlewareBasicAuth` for simple auth
- Support for custom auth via middleware
- Extract auth tokens from `RequestHeader.Authorization`

---

## Performance Guidelines

### 1. Middleware Ordering
Order matters for performance:
```go
server.Use(
    simplehttp.MiddlewareRequestID(),      // 1. Fast, add ID
    simplehttp.MiddlewareLogger(logger),    // 2. Logging
    simplehttp.MiddlewareSecurity(sec),     // 3. Security headers
    simplehttp.MiddlewareCORS(cors),        // 4. CORS handling
    simplehttp.MiddlewareRateLimiter(rate), // 5. Rate limiting
    simplehttp.MiddlewareHeaderParser(),    // 6. Parse headers
    simplehttp.MiddlewareTimeout(timeout),  // 7. Set timeout
)
```

### 2. Caching Strategy
- Use `MiddlewareCache` for expensive operations
- Implement custom `KeyFunc` for cache keys
- Set appropriate TTL values

### 3. Timeout Configuration
```go
config.ConfigTimeOut = &TimeOutConfig{
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

---

## Common Pitfalls & Best Practices

### ✅ DO
- Use interface types in function signatures
- Return errors; don't panic
- Use `context.Context` for cancellation and timeouts
- Call `ValidateConfig()` on configuration
- Use middleware for cross-cutting concerns
- Document all exported symbols
- Use constants for header names and ENV variables

### ❌ DON'T
- Don't access framework-specific context directly in handlers
- Don't hardcode configuration values
- Don't ignore errors
- Don't use framework-specific middleware in application code
- Don't bypass the Context interface
- Don't modify global state in middleware

---

## Integration Guidelines

### Using SimpleHttp in a New Project

1. **Install dependency**
   ```bash
   go get github.com/medatechnology/simplehttp
   ```

2. **Create `.env` file**
   ```env
   SIMPLEHTTP_FRAMEWORK=echo
   SIMPLEHTTP_PORT=8080
   SIMPLEHTTP_APP_NAME=MyApp
   SIMPLEHTTP_DEBUG=false
   ```

3. **Initialize server**
   ```go
   import (
       "github.com/medatechnology/simplehttp"
       "github.com/medatechnology/simplehttp/framework/echo"
   )

   config := simplehttp.LoadConfig()
   server := echo.NewServer(config)
   ```

4. **Add middleware and routes**
   ```go
   server.Use(
       simplehttp.MiddlewareRequestID(),
       simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
   )

   server.GET("/", homeHandler)
   ```

5. **Start server**
   ```go
   if err := server.Start(""); err != nil {
       log.Fatal(err)
   }
   ```

---

## Maintenance Guidelines

### Before Each Session Ends
- [ ] Update `.version` if version changed
- [ ] Update `.changelog` with significant changes
- [ ] Update this `context.md` if patterns changed
- [ ] Run `go mod tidy` to clean dependencies
- [ ] Verify examples still work
- [ ] Update README.md if public API changed

### Code Review Checklist
- [ ] All exported symbols have doc comments
- [ ] Error handling is consistent
- [ ] No framework-specific leaks in interfaces
- [ ] Configuration uses environment variables
- [ ] Tests/examples updated
- [ ] No breaking changes without version bump

---

## Future Considerations

### Planned Features (Check Issues/Roadmap)
- Additional framework adapters (Gin, Chi, etc.)
- Enhanced caching strategies (Redis, etc.)
- Metrics and observability middleware
- GraphQL support
- gRPC gateway support

### Extension Points
- Custom `Logger` implementations
- Custom `Cache` implementations
- Custom error handlers
- Custom middleware

---

## Quick Reference

### Essential Commands
```bash
# Run example
go run ./example

# Build library
go build

# Run tests
go test ./...

# Update dependencies
go get -u ./...
go mod tidy

# Format code
go fmt ./...

# Lint (if golangci-lint installed)
golangci-lint run
```

### Key Files for Common Tasks

| Task | Primary Files |
|------|--------------|
| Add new framework | `framework/<name>/server.go`, `context.go`, `adapter.go` |
| Add middleware | `middleware.go` |
| Modify interfaces | `simplehttp.go` |
| Change configuration | `config.go` |
| Add utilities | `helper.go`, `security.go` |
| File handling | `files.go` |
| Logging | `logger.go` |
| Caching | `cache.go` |
| HTTP client | `client/client.go` |

---

## Agent Instructions

### Context Management
> **CRITICAL:** This context file must be updated before:
> - Session ends
> - Token limits are approaching
> - Major architectural changes are made
> - New patterns are introduced
> - Breaking changes are committed

### What to Update
- **Version history**: Add to `.changelog`
- **Patterns**: Document new coding patterns here
- **Dependencies**: Update if new packages added
- **Conventions**: Add new conventions as they emerge
- **Architecture**: Update diagrams if structure changes
- **Examples**: Reference new example code

### Update Workflow
1. Review changes made during session
2. Identify new patterns or conventions
3. Update relevant sections in this file
4. Update `.version` and `.changelog` if needed
5. Commit context changes separately with clear message

### Context Preservation Pattern
```bash
# Before ending session or hitting limits
1. Review session changes
2. Update .agent/context.md
3. Update .version and .changelog if needed
4. Commit: "docs: update project context [session end]"
```

---

**Last reviewed:** 2026-01-21  
**Next review:** Before session end or at token limit warning
