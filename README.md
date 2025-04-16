# SimpleHttp

SimpleHttp is a flexible HTTP handler interface in Go that provides a unified abstraction layer for multiple web frameworks. It allows you to write your web applications once and switch between different underlying frameworks (Echo, Fiber, etc.) without changing your application code.

## Features

- **Framework Agnostic**: Supports multiple frameworks (Echo, Fiber, FastHTTP, Gin) through a consistent interface
- **Modular Middleware**: Built-in support for logging, authentication, rate limiting, CORS, and more
- **Standardized Header Parsing**: Easily extract auth tokens, API keys, IP addresses, and browser info
- **Websocket Support**: Integrated WebSocket handling with the same consistent API
- **File Handling**: Upload and download files with size limits and content validation
- **Configuration via Environment**: Seamlessly switch between frameworks using environment variables
- **Graceful Shutdown**: Handle shutdowns gracefully for better user experience

## Installation

```bash
go get github.com/medatechnology/simplehttp
```

## Quick Start

### 1. Set Up Your Project

Create a new Go project and add SimpleHttp as a dependency:

```bash
mkdir myapi
cd myapi
go mod init github.com/yourusername/myapi
go get github.com/medatechnology/simplehttp
```

Create a `.env` file in your project root:

```
# Framework to use (echo, fiber, etc.)
MEDA_FRAMEWORK=echo

# Application settings
MEDA_APP_NAME=MyAPIService
MEDA_PORT=8080
MEDA_HOST_NAME=localhost

# Timeouts in seconds
MEDA_READ_TIMEOUT=30
MEDA_WRITE_TIMEOUT=30
MEDA_IDLE_TIMEOUT=60

# Debug mode
MEDA_DEBUG=false

# Display startup message
FRAMEWORK_STARTUP_MESSAGE=true
```

### 2. Create Your First API

Create a `main.go` file:

```go
package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "context"
    "time"

    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/echo" // Or fiber
)

func main() {
    // Load configuration from .env
    config := simplehttp.LoadConfig()

    // Create server instance
    server := echo.NewServer(config)

    // Add middleware
    server.Use(
        simplehttp.MiddlewareRequestID(),
        simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
    )

    // Define routes
    server.GET("/hello", func(c simplehttp.MedaContext) error {
        return c.String(http.StatusOK, "Hello, World!")
    })

    api := server.Group("/api")
    {
        api.GET("/status", func(c simplehttp.MedaContext) error {
            return c.JSON(http.StatusOK, map[string]string{
                "status": "running",
                "version": "1.0.0",
            })
        })
    }

    // Start server in a goroutine
    go func() {
        if err := server.Start(""); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit

    // Gracefully shutdown
    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exiting")
}
```

### 3. Run Your API

Build and run your application:

```bash
go build
./myapi
```

Your API is now running! Try these endpoints:

- GET http://localhost:8080/hello
- GET http://localhost:8080/api/status

## Middleware

SimpleHttp comes with several built-in middleware components that you can use to enhance your application:

### RequestID Middleware

Adds a unique identifier to each request:

```go
server.Use(simplehttp.MiddlewareRequestID())
```

### Logger Middleware

Logs incoming requests and outgoing responses:

```go
// Use default logger
server.Use(simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()))

// Use custom logger with configuration
logConfig := &simplehttp.DefaultLoggerConfig{
    Level:      simplehttp.LogLevelDebug,
    TimeFormat: "2006/01/02 15:04:05",
    Output:     os.Stdout,
    Prefix:     "[MyAPI] ",
}
logger := simplehttp.NewDefaultLogger(logConfig)
server.Use(simplehttp.MiddlewareLogger(logger))
```

### Timeout Middleware

Sets a maximum duration for request handling:

```go
timeoutConfig := simplehttp.TimeOutConfig{
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
}
server.Use(simplehttp.MiddlewareTimeout(timeoutConfig))
```

### HeaderParser Middleware

Parses common HTTP headers into a structured object:

```go
server.Use(simplehttp.MiddlewareHeaderParser())

// Later in your handler
func myHandler(c simplehttp.MedaContext) error {
    headers := c.GetHeaders()
    
    // Access parsed header data
    userAgent := headers.UserAgent
    browserName := headers.Browser
    clientIP := headers.RealIP
    requestID := headers.RequestID
    
    return c.JSON(http.StatusOK, map[string]string{
        "browser": browserName,
        "ip": clientIP,
    })
}
```

### CORS Middleware

Handles Cross-Origin Resource Sharing:

```go
corsConfig := &simplehttp.CORSConfig{
    AllowOrigins:     []string{"https://example.com", "https://api.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           24 * time.Hour,
}
server.Use(simplehttp.MiddlewareCORS(corsConfig))
```

### RateLimiter Middleware

Prevents abuse by limiting request frequency:

```go
rateConfig := simplehttp.RateLimitConfig{
    RequestsPerSecond: 10,         // Allow 10 requests per second
    BurstSize:         20,         // Allow bursts up to 20 requests
    KeyFunc: func(c simplehttp.MedaContext) string {
        // Rate limit by IP
        headers := c.GetHeaders()
        if headers.RealIP != "" {
            return headers.RealIP
        }
        return headers.RemoteIP
    },
}
server.Use(simplehttp.MiddlewareRateLimiter(rateConfig))
```

### BasicAuth Middleware

Adds HTTP Basic Authentication:

```go
// Global basic auth
server.Use(simplehttp.MiddlewareBasicAuth("admin", "secret"))

// Or apply to specific route groups
adminAPI := server.Group("/admin")
adminAPI.Use(simplehttp.MiddlewareBasicAuth("admin", "admin_password"))
```

### Security Middleware

Adds security-related headers:

```go
secConfig := simplehttp.SecurityConfig{
    FrameDeny:             true,   // Add X-Frame-Options: DENY
    ContentTypeNosniff:    true,   // Add X-Content-Type-Options: nosniff
    BrowserXssFilter:      true,   // Add X-XSS-Protection: 1; mode=block
    ContentSecurityPolicy: "default-src 'self'",
}
server.Use(simplehttp.MiddlewareSecurity(secConfig))
```

### Cache Middleware

Caches responses to improve performance:

```go
cacheConfig := simplehttp.CacheConfig{
    TTL:       5 * time.Minute,
    KeyPrefix: "api:",
    Store:     simplehttp.NewMemoryCache(),
    KeyFunc: func(c simplehttp.MedaContext) string {
        // Cache key based on path and auth
        return c.GetPath() + ":" + c.GetHeader("Authorization")
    },
}
server.Use(simplehttp.MiddlewareCache(cacheConfig))
```

## Creating Custom Middleware

You can create your own middleware to extend SimpleHttp's functionality:

### Example: Request Timer Middleware

```go
func RequestTimer() simplehttp.MedaMiddleware {
    return simplehttp.WithName("request-timer", func(next simplehttp.MedaHandlerFunc) simplehttp.MedaHandlerFunc {
        return func(c simplehttp.MedaContext) error {
            // Start timing
            start := time.Now()
            
            // Process request
            err := next(c)
            
            // Calculate duration
            duration := time.Since(start)
            
            // Add duration header
            c.SetResponseHeader("X-Request-Duration", duration.String())
            
            return err
        }
    })
}

// Usage
server.Use(RequestTimer())
```

## Route-Specific Middleware

You can apply middleware to specific route groups:

```go
// Global middleware
server.Use(
    simplehttp.MiddlewareRequestID(),
    simplehttp.MiddlewareLogger(logger),
)

// API routes with rate limiting
api := server.Group("/api")
api.Use(simplehttp.MiddlewareRateLimiter(rateConfig))

// Admin routes with authentication
admin := server.Group("/admin")
admin.Use(simplehttp.MiddlewareBasicAuth("admin", "password"))

// Public routes with caching
public := server.Group("/public")
public.Use(simplehttp.MiddlewareCache(cacheConfig))
```

## File Handling

SimpleHttp provides built-in file handling capabilities:

```go
// Setup file handler
fileHandler := simplehttp.NewFileHandler("./uploads")
fileHandler.MaxFileSize = 10 << 20  // 10MB
fileHandler.AllowedTypes = []string{"image/jpeg", "image/png", "application/pdf"}

// File upload endpoint
server.POST("/upload", fileHandler.HandleUpload())

// File download endpoint
server.GET("/files/:filename", fileHandler.HandleDownload("./uploads/{{filename}}"))
```

## WebSockets

SimpleHttp has built-in WebSocket support:

```go
server.WebSocket("/ws/chat", func(ws simplehttp.MedaWebsocket) error {
    for {
        // Read message
        msg := &Message{}
        if err := ws.ReadJSON(msg); err != nil {
            return err
        }
        
        // Process message...
        
        // Send response
        response := &Message{
            Type: "response",
            Data: "Received: " + msg.Data,
        }
        
        if err := ws.WriteJSON(response); err != nil {
            return err
        }
    }
})
```

## Complete Example with Middleware and Route Groups

Here's a more complete example that demonstrates how to use SimpleHttp with various middleware and route groups:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/echo"
)

func main() {
    // Load config
    config := simplehttp.LoadConfig()
    
    // Create server
    server := echo.NewServer(config)
    
    // Global middleware
    server.Use(
        simplehttp.MiddlewareRequestID(),
        simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
        simplehttp.MiddlewareHeaderParser(),
    )
    
    // CORS configuration
    corsConfig := &simplehttp.CORSConfig{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           24 * time.Hour,
    }
    
    // Security configuration
    secConfig := simplehttp.SecurityConfig{
        FrameDeny:          true,
        ContentTypeNosniff: true,
        BrowserXssFilter:   true,
    }
    
    // Rate limit configuration
    rateConfig := simplehttp.RateLimitConfig{
        RequestsPerSecond: 10,
        BurstSize:         20,
        KeyFunc: func(c simplehttp.MedaContext) string {
            headers := c.GetHeaders()
            return headers.RealIP
        },
    }
    
    // Public routes
    server.GET("/", func(c simplehttp.MedaContext) error {
        return c.String(http.StatusOK, "Welcome to SimpleHttp API")
    })
    
    // API routes with additional middleware
    api := server.Group("/api")
    api.Use(
        simplehttp.MiddlewareCORS(corsConfig),
        simplehttp.MiddlewareSecurity(secConfig),
        simplehttp.MiddlewareRateLimiter(rateConfig),
    )
    
    // API endpoints
    api.GET("/status", getStatus)
    
    // Users endpoints
    users := api.Group("/users")
    users.GET("", listUsers)
    users.POST("", createUser)
    users.GET("/:id", getUser)
    users.PUT("/:id", updateUser)
    users.DELETE("/:id", deleteUser)
    
    // Admin endpoints with basic auth
    admin := server.Group("/admin")
    admin.Use(simplehttp.MiddlewareBasicAuth("admin", "secret"))
    admin.GET("/dashboard", adminDashboard)
    
    // File handling
    fileHandler := simplehttp.NewFileHandler("./uploads")
    fileHandler.MaxFileSize = 10 << 20
    server.POST("/upload", fileHandler.HandleUpload())
    server.GET("/files/:filename", fileHandler.HandleDownload("./uploads/{{filename}}"))
    
    // WebSocket endpoint
    server.WebSocket("/ws/chat", handleChat)
    
    // Start server
    go func() {
        if err := server.Start(""); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
}

// Handler functions
func getStatus(c simplehttp.MedaContext) error {
    return c.JSON(http.StatusOK, map[string]string{
        "status": "running",
        "version": "1.0.0",
    })
}

func listUsers(c simplehttp.MedaContext) error {
    return c.JSON(http.StatusOK, []map[string]string{
        {"id": "1", "name": "John"},
        {"id": "2", "name": "Jane"},
    })
}

func createUser(c simplehttp.MedaContext) error {
    var user struct {
        Name string `json:"name"`
    }
    if err := c.BindJSON(&user); err != nil {
        return err
    }
    return c.JSON(http.StatusCreated, user)
}

func getUser(c simplehttp.MedaContext) error {
    id := c.GetQueryParam("id")
    return c.JSON(http.StatusOK, map[string]string{
        "id": id,
        "name": "John Doe",
    })
}

func updateUser(c simplehttp.MedaContext) error {
    id := c.GetQueryParam("id")
    var user struct {
        Name string `json:"name"`
    }
    if err := c.BindJSON(&user); err != nil {
        return err
    }
    return c.JSON(http.StatusOK, map[string]string{
        "id": id,
        "name": user.Name,
    })
}

func deleteUser(c simplehttp.MedaContext) error {
    id := c.GetQueryParam("id")
    return c.JSON(http.StatusOK, map[string]string{
        "message": "User " + id + " deleted",
    })
}

func adminDashboard(c simplehttp.MedaContext) error {
    return c.JSON(http.StatusOK, map[string]string{
        "message": "Admin dashboard",
    })
}

func handleChat(ws simplehttp.MedaWebsocket) error {
    for {
        msg := struct {
            Type string `json:"type"`
            Text string `json:"text"`
        }{}
        
        if err := ws.ReadJSON(&msg); err != nil {
            return err
        }
        
        response := struct {
            Type string `json:"type"`
            Text string `json:"text"`
        }{
            Type: "response",
            Text: "Echo: " + msg.Text,
        }
        
        if err := ws.WriteJSON(response); err != nil {
            return err
        }
    }
}
```

## Environment Configuration

SimpleHttp can be configured using environment variables. Here's a complete example of what you can set in your `.env` file:

```
# Framework configuration
MEDA_FRAMEWORK=echo              # Framework to use (echo, fiber)
MEDA_APP_NAME=SimpleHttp-App     # Application name
MEDA_HOST_NAME=localhost         # Host name
MEDA_PORT=8080                   # Port to listen on

# Timeout configuration (in seconds)
MEDA_READ_TIMEOUT=30             # HTTP read timeout
MEDA_WRITE_TIMEOUT=30            # HTTP write timeout
MEDA_IDLE_TIMEOUT=60             # HTTP idle timeout

# Debug and logging
MEDA_DEBUG=false                 # Debug mode
FRAMEWORK_STARTUP_MESSAGE=true   # Display startup message
```

## Middleware Order

The order in which middleware is applied is important. Middleware is executed in the order it's added:

```go
server.Use(
    // 1. First, add request ID
    simplehttp.MiddlewareRequestID(),
    
    // 2. Then log the incoming request with the ID
    simplehttp.MiddlewareLogger(logger),
    
    // 3. Apply security headers
    simplehttp.MiddlewareSecurity(secConfig),
    
    // 4. Handle CORS
    simplehttp.MiddlewareCORS(corsConfig),
    
    // 5. Apply rate limiting
    simplehttp.MiddlewareRateLimiter(rateConfig),
    
    // 6. Parse headers
    simplehttp.MiddlewareHeaderParser(),
    
    // 7. Finally, set timeout
    simplehttp.MiddlewareTimeout(timeoutConfig),
)
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.