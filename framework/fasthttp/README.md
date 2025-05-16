# FastHTTP Implementation Guide for MedaHTTP

This guide explains how to use the FastHTTP implementation of the MedaHTTP framework. The FastHTTP implementation provides high-performance HTTP server capabilities while maintaining compatibility with the MedaHTTP interface.

## Table of Contents
- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Configuration](#configuration)
- [Routing](#routing)
- [Middleware](#middleware)
- [Context Usage](#context-usage)
- [Static Files](#static-files)
- [WebSocket Support](#websocket-support)
- [Examples](#examples)

## Installation

To use the FastHTTP implementation, you need to have the required dependencies:

```bash
go get github.com/valyala/fasthttp
go get github.com/fasthttp/router
go get github.com/fasthttp/websocket
```

## Basic Usage

Here's a basic example of how to create and start a FastHTTP server:

```go
package main

import (
    "log"
    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/fasthttp"
)

func main() {
    // Load configuration
    config := simplehttp.LoadConfig()
    
    // Create new FastHTTP server
    server := fasthttp.NewServer(config)
    
    // Define routes
    server.GET("/", func(c simplehttp.Context) error {
        return c.String(200, "Hello, World!")
    })
    
    // Start server
    if err := server.Start(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

The FastHTTP implementation uses the standard SimpleHTTP configuration. Here are the relevant environment variables:

```bash
SIMPLEHTTP_FRAMEWORK=fasthttp     # Specify FastHTTP as the framework
SIMPLEHTTP_PORT=8080             # Server port
SIMPLEHTTP_READ_TIMEOUT=30s       # Read timeout in seconds
SIMPLEHTTP_WRITE_TIMEOUT=30      # Write timeout in seconds
SIMPLEHTTP_DEBUG=true           # Enable debug mode
```

Custom configuration example:

```go
config := &simplehttp.Config{
    Framework:      "fasthttp",
    Port:           "8080",
    ReadTimeout:    time.Second * 30,
    WriteTimeout:   time.Second * 30,
    MaxRequestSize: 32 << 20, // 32MB
    Debug:          true,
}
```

## Routing

The FastHTTP implementation supports all standard HTTP methods and route grouping:

```go
// Basic routing
server.GET("/users", handleUsers)
server.POST("/users", createUser)
server.PUT("/users/:id", updateUser)
server.DELETE("/users/:id", deleteUser)

// Route grouping
api := server.Group("/api")
{
    v1 := api.Group("/v1")
    {
        v1.GET("/users", handleUsersV1)
        v1.POST("/users", createUserV1)
    }
}
```

### Route Parameters

Access route parameters using the context:

```go
server.GET("/users/:id", func(c simplehttp.Context) error {
    id := c.GetParam("id")
    return c.JSON(200, map[string]string{"id": id})
})
```

## Middleware

Adding middleware to your FastHTTP server:

```go
// Global middleware
server.Use(simplehttp.LoggerMiddleware(logger))
server.Use(simplehttp.RequestID())

// Group middleware
api := server.Group("/api")
api.Use(simplehttp.BasicAuth("username", "password"))

// Route-specific middleware
server.GET("/protected", handler, simplehttp.BasicAuth("user", "pass"))
```

Built-in middleware examples:

```go
// Timeout middleware
server.Use(simplehttp.Timeout(5 * time.Second))

// CORS middleware
corsConfig := &simplehttp.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Origin", "Content-Type"},
}
server.Use(simplehttp.CORS(corsConfig))

// Rate limiter
rateLimitConfig := simplehttp.RateLimitConfig{
    RequestsPerSecond: 10,
    BurstSize:        20,
}
server.Use(simplehttp.RateLimiter(rateLimitConfig))
```

## Context Usage

The FastHTTP context implementation provides access to request and response functionality:

```go
server.POST("/api/data", func(c simplehttp.Context) error {
    // Get request headers
    headers := c.GetHeaders()
    
    // Get query parameters
    query := c.GetQueryParam("filter")
    
    // Get request body
    var data map[string]interface{}
    if err := c.BindJSON(&data); err != nil {
        return err
    }
    
    // Send JSON response
    return c.JSON(200, map[string]interface{}{
        "status": "success",
        "data": data,
    })
})
```

## Static Files

Serving static files with FastHTTP:

```go
// Serve directory
server.Static("/assets", "./public")

// Serve single file
server.StaticFile("/favicon.ico", "./public/favicon.ico")
```

## WebSocket Support

Implementing WebSocket endpoints:

```go
server.WebSocket("/ws", func(ws simplehttp.WebSocket) error {
    for {
        var msg map[string]interface{}
        if err := ws.ReadJSON(&msg); err != nil {
            return err
        }
        
        // Echo the message back
        if err := ws.WriteJSON(msg); err != nil {
            return err
        }
    }
})
```

## Examples

### Complete API Server Example

```go
package main

import (
    "log"
    "time"
    
    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/fasthttp"
)

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func main() {
    // Load configuration
    config := simplehttp.LoadConfig()
    
    // Create server
    server := fasthttp.NewServer(config)
    
    // Add middleware
    server.Use(simplehttp.LoggerMiddleware(simplehttp.NewDefaultLogger()))
    server.Use(simplehttp.RequestID())
    
    // API routes
    api := server.Group("/api")
    {
        api.Use(simplehttp.Timeout(5 * time.Second))
        
        // Users endpoints
        users := api.Group("/users")
        {
            users.GET("", listUsers)
            users.POST("", createUser)
            users.GET("/:id", getUser)
            users.PUT("/:id", updateUser)
            users.DELETE("/:id", deleteUser)
        }
    }
    
    // Start server
    if err := server.Start(config.Port); err != nil {
        log.Fatal(err)
    }
}

func listUsers(c simplehttp.Context) error {
    users := []User{
        {ID: "1", Name: "John"},
        {ID: "2", Name: "Jane"},
    }
    return c.JSON(200, users)
}

func createUser(c simplehttp.Context) error {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return err
    }
    return c.JSON(201, user)
}

func getUser(c simplehttp.Context) error {
    id := c.GetParam("id")
    user := User{ID: id, Name: "John Doe"}
    return c.JSON(200, user)
}

func updateUser(c simplehttp.Context) error {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return err
    }
    return c.JSON(200, user)
}

func deleteUser(c simplehttp.Context) error {
    return c.JSON(204, nil)
}
```

### File Upload Example

```go
func handleFileUpload(c simplehttp.Context) error {
    file, err := c.GetFile("file")
    if err != nil {
        return err
    }
    
    // Save the file
    dst := fmt.Sprintf("./uploads/%s", file.Filename)
    if err := c.SaveFile(file, dst); err != nil {
        return err
    }
    
    return c.JSON(200, map[string]string{
        "message": "File uploaded successfully",
        "path":    dst,
    })
}
```

## Performance Considerations

The FastHTTP implementation is designed for high performance. Here are some tips to maximize performance:

1. Use appropriate buffer sizes for your use case
2. Enable compression only when needed
3. Set reasonable timeouts
4. Use connection pooling when making outbound requests
5. Monitor memory usage with large file uploads

## Error Handling

The FastHTTP implementation includes built-in error handling:

```go
func handler(c simplehttp.SimpleHttpContext) error {
    // Return MedaError for structured error responses
    if someError {
        return simplehttp.NewError(400, "Bad Request", details)
    }
    
    // Regular errors are automatically handled
    return errors.New("something went wrong")
}
```

## Graceful Shutdown

Implementing graceful shutdown:

```go
func main() {
    server := fasthttp.NewServer(config)
    
    // Setup routes...
    
    // Start server in goroutine
    go func() {
        if err := server.Start(":8080"); err != nil {
            log.Printf("Server error: %v\n", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    
    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
}
```

For more information and updates, please visit the project's GitHub repository.