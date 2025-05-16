# SimpleHTTP

SimpleHTTP is a flexible HTTP handler interface for Go that provides a unified abstraction layer for multiple web frameworks. It allows you to write web applications that can seamlessly switch between different web framework implementations (Fiber, Echo, Gin, etc.) without changing your application code.

## Features

- **Framework Agnostic**: Write your code once and switch between different web frameworks
- **Consistent Interface**: Unified API for handling HTTP requests and responses
- **Middleware Support**: Rich collection of built-in middleware for common web application needs
- **Extensible**: Easy to add support for additional web frameworks
- **Configurable**: Simple configuration through environment variables
- **WebSocket Support**: Standardized API for WebSocket connections

## Installation

```bash
go get github.com/medatechnology/simplehttp
```

## Quick Start

```go
package main

import (
	"log"
	"net/http"

	"github.com/medatechnology/simplehttp"
	"github.com/medatechnology/simplehttp/framework/fiber" // Import implementation of your choice
)

func main() {
	// Load configuration from environment variables
	config := simplehttp.LoadConfig()
	
	// Create server with Fiber implementation
	server := fiber.NewServer(config)
	
	// Add middleware
	server.Use(
		simplehttp.MiddlewareRequestID(),
		simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
	)
	
	// Define routes
	server.GET("/hello", func(c simplehttp.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Hello, World!",
		})
	})
	
	// Start server
	if err := server.Start(""); err != nil {
		log.Fatal(err)
	}
}
```

## Configuration

SimpleHTTP can be configured through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| SIMPLEHTTP_FRAMEWORK | The web framework to use (e.g., "fiber", "echo") | "echo" |
| SIMPLEHTTP_PORT | The port to listen on | "8080" |
| SIMPLEHTTP_APP_NAME | Application name | "MedaHTTP" |
| SIMPLEHTTP_HOST_NAME | Hostname to bind to | "localhost" |
| SIMPLEHTTP_READ_TIMEOUT | Read timeout in seconds | 30s |
| SIMPLEHTTP_WRITE_TIMEOUT | Write timeout in seconds | 30s |
| SIMPLEHTTP_IDLE_TIMEOUT | Idle timeout in seconds | 60s |
| SIMPLEHTTP_DEBUG | Enable debug mode | false |
| SIMPLEHTTP_FRAMEWORK_STARTUP_MESSAGE | Show startup message | true |

## Core Components

### MedaContext

`MedaContext` provides a unified interface for handling HTTP requests and responses:

```go
// Get request information
path := c.GetPath()
method := c.GetMethod()
header := c.GetHeader("Content-Type")
headers := c.GetHeaders() // Get all parsed headers
queryParam := c.GetQueryParam("filter")
body := c.GetBody()

// Set headers
c.SetHeader("X-Custom-Header", "value")
c.SetRequestHeader("Authorization", "Bearer token")
c.SetResponseHeader("Cache-Control", "no-cache")

// Response methods
c.JSON(http.StatusOK, data)
c.String(http.StatusOK, "Hello World")
c.Stream(http.StatusOK, "text/plain", reader)

// File handling
file, err := c.GetFile("upload")
c.SaveFile(file, "/path/to/save")
c.SendFile("/path/to/file", true) // Download as attachment

// Context values
c.Set("user", user)
user := c.Get("user")

// Data binding
var user User
c.BindJSON(&user)
c.BindForm(&user)
```

### MedaRouter

`MedaRouter` provides routing capabilities:

```go
// HTTP methods
server.GET("/users", listUsers)
server.POST("/users", createUser)
server.PUT("/users/:id", updateUser)
server.DELETE("/users/:id", deleteUser)

// Route groups
api := server.Group("/api")
{
    api.GET("/status", getStatus)
    
    // Nested groups
    users := api.Group("/users")
    {
        users.GET("", listUsers)
        users.POST("", createUser)
    }
}

// Static file serving
server.Static("/assets", "./public")
server.StaticFile("/favicon.ico", "./public/favicon.ico")

// Websockets
server.WebSocket("/ws", handleWebSocket)
```

### Middleware

MedaHTTP includes a variety of built-in middleware:

```go
// Basic middleware
server.Use(
    simplehttp.MiddlewareRequestID(),
    simplehttp.MiddlewareLogger(logger),
    simplehttp.MiddlewareHeaderParser(),
)

// Rate limiting
server.Use(simplehttp.MiddlewareRateLimiter(simplehttp.RateLimitConfig{
    RequestsPerSecond: 10,
    BurstSize: 20,
    KeyFunc: func(c simplehttp.Context) string {
        headers := c.GetHeaders()
        return headers.RealIP // Rate limit by IP
    },
}))

// Security middleware
server.Use(simplehttp.MiddlewareSecurity(simplehttp.SecurityConfig{
    FrameDeny: true,
    ContentTypeNosniff: true,
    BrowserXssFilter: true,
}))

// CORS middleware
server.Use(simplehttp.MiddlewareCORS(&simplehttp.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Origin", "Content-Type"},
}))

// Timeout middleware
server.Use(simplehttp.MiddlewareTimeout(simplehttp.TimeOutConfig{
    ReadTimeout: 30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout: 60 * time.Second,
}))

// Basic Auth middleware
server.Use(simplehttp.MiddlewareBasicAuth("username", "password"))

// Cache middleware
server.Use(simplehttp.MiddlewareCache(simplehttp.CacheConfig{
    TTL: 5 * time.Minute,
    Store: simplehttp.NewMemoryCache(),
    KeyFunc: func(c simplehttp.Context) string {
        return c.GetPath() + c.GetHeader("Authorization")
    },
}))
```

## Advanced Usage

### File Handling

```go
// Create a file handler
fileHandler := simplehttp.NewFileHandler("./uploads")
fileHandler.MaxFileSize = 50 << 20 // 50MB
fileHandler.AllowedTypes = []string{
    "image/jpeg",
    "image/png",
    "application/pdf",
}

// Register file routes
server.POST("/upload", fileHandler.HandleUpload())
server.GET("/files/:filename", fileHandler.HandleDownload("./uploads/{{filename}}"))
```

### WebSockets

```go
type Message struct {
    Type string `json:"type"`
    Data string `json:"data"`
}

server.WebSocket("/ws/chat", func(ws simplehttp.WebSocket) error {
    for {
        msg := &Message{}
        if err := ws.ReadJSON(msg); err != nil {
            return err
        }
        
        // Echo the message back
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

### Graceful Shutdown

```go
// Start server in a goroutine
go func() {
    if err := server.Start(""); err != nil {
        log.Printf("server error: %v", err)
    }
}()

// Wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
<-quit

// Gracefully shutdown
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Fatal(err)
}
```

## Framework Implementations

SimpleHTTP currently includes the following framework implementations:

- Fiber (`github.com/medatechnology/simplehttp/framework/fiber`)

## License

[MIT License](LICENSE)