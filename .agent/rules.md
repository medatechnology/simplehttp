# SimpleHttp Project Rules

This file defines mandatory coding rules, conventions, and patterns for the SimpleHttp project.

---

## 🚨 CRITICAL RULES - NEVER VIOLATE

### 1. Framework Agnosticism
**RULE:** Never expose framework-specific types in the public API.

❌ **WRONG:**
```go
func MyHandler(c *fiber.Ctx) error { ... }  // Framework-specific!
```

✅ **CORRECT:**
```go
func MyHandler(c simplehttp.Context) error { ... }  // Framework-agnostic!
```

**Why:** The entire purpose of SimpleHttp is framework independence. Breaking this rule breaks the core value proposition.

---

### 2. Interface-First Design
**RULE:** Define interfaces before implementations. All major components must have an interface.

❌ **WRONG:**
```go
// Just a concrete type
type MyCache struct {
    data map[string]interface{}
}
```

✅ **CORRECT:**
```go
// Interface first
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration) error
}

// Then implementation
type memoryCache struct {
    data map[string]interface{}
}
```

**Why:** Enables swappable implementations and framework adapters.

---

### 3. Context Never Stored
**RULE:** Never store `simplehttp.Context` in structs. Always pass as parameters.

❌ **WRONG:**
```go
type MyService struct {
    ctx simplehttp.Context  // Don't store context!
}
```

✅ **CORRECT:**
```go
type MyService struct {
    db *Database
}

func (s *MyService) Handle(c simplehttp.Context) error {
    // Pass context as parameter
}
```

**Why:** Contexts are request-scoped. Storing them causes race conditions and memory leaks.

---

### 4. Middleware Must Be Named
**RULE:** All middleware must use `WithName()` wrapper.

❌ **WRONG:**
```go
func MiddlewareMyFeature() MiddlewareFunc {
    return func(next HandlerFunc) HandlerFunc { ... }
}
```

✅ **CORRECT:**
```go
func MiddlewareMyFeature() Middleware {
    return WithName("my-feature", MyFeature())
}

func MyFeature() MiddlewareFunc {
    return func(next HandlerFunc) HandlerFunc { ... }
}
```

**Why:** Named middleware enables debugging and middleware ordering inspection.

---

### 5. Error Handling - Never Panic
**RULE:** Return errors instead of panicking. Let the framework handle error responses.

❌ **WRONG:**
```go
func MyHandler(c simplehttp.Context) error {
    if err != nil {
        panic(err)  // Never panic!
    }
}
```

✅ **CORRECT:**
```go
func MyHandler(c simplehttp.Context) error {
    if err != nil {
        return err  // Return errors
    }
}
```

**Why:** Panics crash the server. Returning errors allows graceful handling.

---

## 📋 MANDATORY CONVENTIONS

### File Organization

**RULE:** Framework-specific code must stay in `framework/<name>/` directory.

```
✅ framework/echo/server.go     - Echo implementation
✅ framework/fiber/context.go   - Fiber adapter
❌ echo_server.go               - Wrong location
❌ simplehttp.go (with Echo code) - No framework code in root
```

---

### Environment Variables

**RULE:** All ENV variables must be prefixed with `SIMPLEHTTP_`

✅ Correct:
```go
SIMPLEHTTP_FRAMEWORK
SIMPLEHTTP_PORT
SIMPLEHTTP_DEBUG
```

❌ Wrong:
```go
FRAMEWORK      // Missing prefix
PORT           // Missing prefix
APP_DEBUG      // Wrong prefix
```

**Why:** Prevents conflicts with other packages and clearly identifies SimpleHttp configs.

---

### Naming Patterns

**RULE:** Follow these naming patterns strictly:

| Type | Pattern | Example |
|------|---------|---------|
| Middleware factory | `Middleware<Name>()` | `MiddlewareLogger()`, `MiddlewareCORS()` |
| Middleware implementation | `<Name>()` | `Logger()`, `CORS()` |
| Config struct | `<Name>Config` | `CORSConfig`, `SecurityConfig` |
| Interface | `<Feature>` (noun) | `Context`, `Server`, `Logger` |
| Constants | `UPPER_SNAKE_CASE` | `DEFAULT_HTTP_READ_TIMEOUT` |
| Env constants | `SIMPLEHTTP_<NAME>` | `SIMPLEHTTP_FRAMEWORK` |

---

### Documentation

**RULE:** Every exported symbol MUST have a godoc comment.

❌ **WRONG:**
```go
type Server interface {
    Start(address string) error
}
```

✅ **CORRECT:**
```go
// Server interface defines the contract for our web server
type Server interface {
    Start(address string) error
}
```

**Format:**
- Start with the item name
- Be a complete sentence
- Explain purpose, not just restating the signature

---

## 🛠️ DEVELOPMENT RULES

### Adding New Features

**RULE:** When adding new features, follow this order:

1. **Define interface** in `simplehttp.go` or relevant file
2. **Create configuration** struct if needed
3. **Implement for all frameworks** (or mark as framework-specific)
4. **Add middleware wrapper** if it's middleware
5. **Document in README.md**
6. **Add example** in `example/` directory
7. **Update `.agent/context.md`**

---

### Backward Compatibility

**RULE:** When updating existing features:

✅ **Allowed:**
- Adding new optional fields to config structs
- Adding new methods to interfaces (consider carefully)
- Adding new middleware
- Internal refactoring

❌ **Not Allowed Without Version Bump:**
- Removing public methods
- Changing function signatures
- Renaming exported symbols
- Changing struct field types

**If breaking change needed:**
1. Bump major version
2. Document in `.changelog`
3. Update README.md with migration guide

---

### Configuration Defaults

**RULE:** All configuration must have sensible defaults in `DefaultConfig`.

```go
var DefaultConfig = &Config{
    Framework: "fiber",           // Default framework
    Port:      "8080",            // Standard dev port
    Debug:     false,             // Safe default
    ConfigTimeOut: &TimeOutConfig{
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    },
}
```

**Why:** Enables "zero-config" startup for development.

---

### Security Rules

**RULE:** Security-first defaults:

```go
✅ Debug:     false          // Safe default
✅ CORS:      restrictive    // Don't allow all origins by default
❌ Debug:     true           // Dangerous default
❌ CORS:      "*"            // Too permissive
```

**File Handling:**
- Always validate file size limits
- Always check file types
- Sanitize file paths to prevent directory traversal

---

## 🧪 TESTING RULES

### Example-Driven Development

**RULE:** Each feature should have a working example in `example/` directory.

**Pattern:**
```go
// In example/main.go
func ExampleMyFeature() {
    config := simplehttp.LoadConfig()
    server := fiber.NewServer(config)
    
    // Show the feature usage
    server.Use(simplehttp.MiddlewareMyFeature(config))
    
    server.Start("")
}
```

---

### Cross-Framework Testing

**RULE:** Major features should be tested with multiple frameworks.

**Pattern in `example/main.go`:**
```go
// Test with multiple frameworks
func TestWithEcho() { ... }
func TestWithFiber() { ... }
func TestWithFastHTTP() { ... }
```

---

## 🔧 IMPLEMENTATION PATTERNS

### Adapter Pattern

**RULE:** Each framework adapter must implement all required interfaces.

**Required files:**
```
framework/<name>/
├── server.go    - Implements Server interface
├── context.go   - Implements Context interface
└── adapter.go   - Framework-specific utilities
```

**Standard structure:**
```go
// server.go
type <framework>Server struct {
    app    *<framework>.App
    config *simplehttp.Config
}

func NewServer(config *simplehttp.Config) simplehttp.Server {
    // Validate config
    if err := simplehttp.ValidateConfig(config); err != nil {
        panic(err)
    }
    
    // Initialize framework
    // Return adapter
}
```

---

### Header Parsing Pattern

**RULE:** Use `RequestHeader` struct for all header access.

❌ **WRONG:**
```go
func MyHandler(c simplehttp.Context) error {
    ip := c.GetHeader("X-Real-IP")  // Direct header access
}
```

✅ **CORRECT:**
```go
func MyHandler(c simplehttp.Context) error {
    headers := c.GetHeaders()
    ip := headers.RealIP  // Use structured access
}
```

**Why:** `RequestHeader` handles multiple header variations and provides consistent access.

---

### Middleware Ordering

**RULE:** Apply middleware in this order:

1. **RequestID** - Add unique ID first
2. **Logger** - Log with the ID
3. **Security** - Apply security headers
4. **CORS** - Handle cross-origin
5. **RateLimiter** - Protect from abuse
6. **HeaderParser** - Parse request headers
7. **Timeout** - Set timeout last

```go
server.Use(
    simplehttp.MiddlewareRequestID(),      // 1
    simplehttp.MiddlewareLogger(logger),    // 2
    simplehttp.MiddlewareSecurity(sec),     // 3
    simplehttp.MiddlewareCORS(cors),        // 4
    simplehttp.MiddlewareRateLimiter(rate), // 5
    simplehttp.MiddlewareHeaderParser(),    // 6
    simplehttp.MiddlewareTimeout(timeout),  // 7
)
```

---

## 📦 DEPENDENCY RULES

### Minimal Dependencies

**RULE:** Only add dependencies that are truly necessary.

**Ask before adding:**
- Can we implement this ourselves simply?
- Is this a core dependency or convenience?
- Does it add significant value?

**Excluded dependencies:**
- Utility libraries we can implement (e.g., simple string manipulation)
- Heavy frameworks that conflict with our abstraction
- Deprecated packages

---

### External Package Usage

**RULE:** Prefer `medatechnology/goutil` for common utilities.

✅ **Use:**
```go
import "github.com/medatechnology/goutil/utils"

value := utils.GetEnvString("KEY", "default")
```

❌ **Don't:**
```go
// Don't add another ENV parsing library
import "some-other-env-lib"
```

---

## 🔍 CODE REVIEW CHECKLIST

Before committing, verify:

### Interface Compliance
- [ ] No framework-specific types in public API
- [ ] All adapters implement required interfaces
- [ ] Interfaces are minimal and focused

### Documentation
- [ ] All exported symbols have godoc comments
- [ ] README.md updated if public API changed
- [ ] Examples are working and clear

### Configuration
- [ ] Uses environment variables with `SIMPLEHTTP_` prefix
- [ ] Has defaults in `DefaultConfig`
- [ ] Validates configuration

### Error Handling
- [ ] Returns errors, never panics
- [ ] Errors are meaningful
- [ ] Nil checks where needed

### Middleware
- [ ] Uses `WithName()` wrapper
- [ ] Follows standard pattern
- [ ] Can be applied globally or per-route

### Security
- [ ] No hardcoded credentials
- [ ] File paths are sanitized
- [ ] Input is validated
- [ ] Safe defaults used

---

## 🚀 RELEASE CHECKLIST

Before tagging a release:

- [ ] All examples work
- [ ] README.md is current
- [ ] `.agent/context.md` is updated
- [ ] `.version` is bumped
- [ ] `.changelog` has entry
- [ ] No breaking changes without major version bump
- [ ] Documentation complete
- [ ] Dependencies are clean (`go mod tidy`)

---

## ⚡ PERFORMANCE RULES

### Avoid Allocations in Hot Paths

**RULE:** Minimize allocations in middleware and handlers.

✅ **Good:**
```go
// Reuse buffers
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}
```

❌ **Bad:**
```go
// Allocating on every request
func (c *Context) GetBody() []byte {
    buf := make([]byte, 1024)  // Allocation on every call
}
```

---

### Cache When Possible

**RULE:** Use the caching middleware for expensive operations.

```go
cacheConfig := simplehttp.CacheConfig{
    TTL: 5 * time.Minute,
    KeyFunc: func(c simplehttp.Context) string {
        return c.GetPath()
    },
}
server.Use(simplehttp.MiddlewareCache(cacheConfig))
```

---

## 🎯 COMMON MISTAKES TO AVOID

### 1. Storing Context
```go
❌ type Service struct { ctx simplehttp.Context }
✅ func (s *Service) Process(c simplehttp.Context) error
```

### 2. Framework Leakage
```go
❌ func Handler(c *fiber.Ctx) error
✅ func Handler(c simplehttp.Context) error
```

### 3. Missing Error Handling
```go
❌ file, _ := os.Open("file.txt")  // Ignoring error
✅ file, err := os.Open("file.txt")
   if err != nil { return err }
```

### 4. Hardcoded Values
```go
❌ port := "8080"  // Hardcoded
✅ port := utils.GetEnvString("SIMPLEHTTP_PORT", "8080")
```

### 5. Panicking Instead of Returning Errors
```go
❌ panic(err)
✅ return err
```

---

## 📝 AGENT-SPECIFIC RULES

### Before Ending Session

**MANDATORY:** Run `/update-context` workflow before:
- Session ends
- Token limit approaches
- Major changes committed

### When Making Changes

1. **Understand the architecture** - Read `.agent/context.md` first
2. **Follow patterns** - Use existing patterns, don't invent new ones
3. **Document changes** - Update context if patterns change
4. **Test with examples** - Verify changes work with `example/main.go`

### When Updating Dependencies

1. Run `go mod tidy` after changes
2. Document new dependencies in context.md
3. Explain why they were needed
4. Check for version conflicts

---

## 🔄 WORKFLOW INTEGRATION

These workflows MUST be used:

- `/update-context` - Before session end or token limits
- Add more workflows as needed

---

**Last Updated:** 2026-01-21  
**Enforcement:** These rules are mandatory for all contributions
