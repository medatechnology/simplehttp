# Echo Implementation Guide

This guide demonstrates how to use the Echo framework implementation with MedaHTTP.

## Installation

```bash
go get github.com/medatechnology/simplehttp/framework/echo
```

## Basic Usage

Here's a simple example of using MedaHTTP with Echo:

```go
package main

import (
    "log"
    
    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/echo"
)

func main() {
    // Load configuration
    config := simplehttp.LoadConfig()
    
    // Create Echo server
    server := echo.NewServer(config)
    
    // Add middleware
    server.Use(
        simplehttp.RequestID(),
        simplehttp.LoggerMiddleware(simplehttp.NewDefaultLogger()),
        simplehttp.HeaderParser(),
    )
    
    // Define routes
    server.GET("/", handleHome)
    server.POST("/api/users", handleCreateUser)
    
    // Start server
    if err := server.Start(":8080"); err != nil {
        log.Fatal(err)
    }
}

func handleHome(c simplehttp.Context) error {
    return c.JSON(200, map[string]string{
        "message": "Welcome to SimpleHTTP with Echo!",
    })
}

func handleCreateUser(c simplehttp.Context) error {
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := c.BindJSON(&user); err != nil {
        return err
    }
    
    // Process user...
    
    return c.JSON(201, user)
}
```

## Complete Example

Here's a more comprehensive example showing various features:

```go
package main

import (
    "log"
    "time"
    
    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/echo"
)

type App struct {
    server simplehttp.Server
    logger simplehttp.Logger
}

func NewApp() *App {
    // Load configuration
    config := simplehttp.LoadConfig()
    
    // Create logger
    logger := simplehttp.NewDefaultLogger(&simplehttp.DefaultLoggerConfig{
        Level:      simplehttp.LogLevelDebug,
        TimeFormat: "2006/01/02 15:04:05",
        Prefix:     "[APP] ",
    })
    
    // Create server
    server := echo.NewServer(config)
    
    return &App{
        server: server,
        logger: logger,
    }
}

func (app *App) setupMiddleware() {
    // Basic middleware
    app.server.Use(
        simplehttp.RequestID(),
        simplehttp.LoggerMiddleware(app.logger),
        simplehttp.HeaderParser(),
    )
    
    // Security middleware
    app.server.Use(
        simplehttp.Security(simplehttp.SecurityConfig{
            ContentTypeNosniff: true,
            BrowserXssFilter:  true,
            FrameDeny:        true,
        }),
    )
    
    // Rate limiting
    app.server.Use(
        simplehttp.RateLimiter(simplehttp.RateLimitConfig{
            RequestsPerSecond: 10,
            BurstSize:        20,
            KeyFunc: func(c simplehttp.Context) string {
                return c.GetHeader("X-Real-IP")
            },
        }),
    )
    
    // CORS
    app.server.Use(
        simplehttp.CORS(&simplehttp.CORSConfig{
            AllowOrigins:     []string{"*"},
            AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
            AllowCredentials: true,
            MaxAge:          24 * time.Hour,
        }),
    )
}

func (app *App) setupRoutes() {
    // API routes
    api := app.server.Group("/api")
    {
        // User routes
        users := api.Group("/users")
        users.GET("", app.handleListUsers)
        users.POST("", app.handleCreateUser)
        users.GET("/:id", app.handleGetUser)
        
        // File upload routes
        files := api.Group("/files")
        fileHandler := simplehttp.NewFileHandler("./uploads")
        files.POST("/upload", fileHandler.HandleUpload())
        files.GET("/download/:filename", fileHandler.HandleDownload("./uploads"))
    }
    
    // WebSocket route
    app.server.WebSocket("/ws", app.handleWebSocket)
}

func (app *App) handleListUsers(c simplehttp.Context) error {
    users := []map[string]string{
        {"id": "1", "name": "John Doe"},
        {"id": "2", "name": "Jane Doe"},
    }
    return c.JSON(200, users)
}

func (app *App) handleCreateUser(c simplehttp.Context) error {
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := c.BindJSON(&user); err != nil {
        return err
    }
    
    // Validate headers
    headers := c.GetHeaders()
    if headers.APIKey == "" {
        return simplehttp.NewError(401, "API key required")
    }
    
    return c.JSON(201, user)
}

func (app *App) handleGetUser(c simplehttp.Context) error {
    id := c.GetQueryParam("id")
    user := map[string]string{
        "id":    id,
        "name":  "John Doe",
        "email": "john@example.com",
    }
    return c.JSON(200, user)
}

func (app *App) handleWebSocket(ws simplehttp.WebSocket) error {
    for {
        var msg struct {
            Type    string `json:"type"`
            Content string `json:"content"`
        }
        
        if err := ws.ReadJSON(&msg); err != nil {
            return err
        }
        
        response := map[string]string{
            "type":    "response",
            "content": "Received: " + msg.Content,
        }
        
        if err := ws.WriteJSON(response); err != nil {
            return err
        }
    }
}

func main() {
    app := NewApp()
    
    app.setupMiddleware()
    app.setupRoutes()
    
    if err := app.server.Start(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

## Framework-Specific Notes

When using the Echo implementation:

1. **Context Access**: You can access the underlying Echo context if needed:
   ```go
   echoCtx := c.Get("echo.context").(*echo.Context)
   ```

2. **File Uploads**: Echo's multipart file handling is automatically wrapped:
   ```go
   file, err := c.GetFile("upload")
   if err != nil {
       return err
   }
   ```

3. **WebSocket**: The Echo implementation uses gorilla/websocket internally:
   ```go
   server.WebSocket("/ws", func(ws simplehttp.WebSocket) error {
       // WebSocket connection handler
   })
   ```

## Testing

Here's an example of how to test your Echo implementation:

```go
package main

import (
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/echo"
)

func TestEchoServer(t *testing.T) {
    // Create server
    config := simplehttp.LoadConfig()
    server := echo.NewServer(config)
    
    // Add test route
    server.GET("/test", func(c simplehttp.Context) error {
        return c.JSON(200, map[string]string{"message": "success"})
    })
    
    // Create test request
    req := httptest.NewRequest("GET", "/test", nil)
    rec := httptest.NewRecorder()
    
    // Serve request
    server.ServeHTTP(rec, req)
    
    // Assert response
    assert.Equal(t, 200, rec.Code)
    assert.Contains(t, rec.Body.String(), "success")
}
```

## Best Practices

1. **Configuration**: Always use environment variables for configuration:
   ```bash
   export SIMPLEHTTP_FRAMEWORK=echo
   export SIMPLEHTTP_PORT=8080
   export SIMPLEHTTP_READ_TIMEOUT=30
   ```

2. **Middleware Order**: Consider the order of middleware:
   - RequestID (first to track requests)
   - Logger (early to log all requests)
   - Security middleware
   - CORS
   - Rate limiting
   - Your custom middleware

3. **Error Handling**: Use the built-in error types:
   ```go
   if err != nil {
       return simplehttp.NewError(500, "Internal error", err)
   }
   ```

4. **Graceful Shutdown**: Implement graceful shutdown:
   ```go
   c := make(chan os.Signal, 1)
   signal.Notify(c, os.Interrupt)
   go func() {
       <-c
       ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
       defer cancel()
       if err := server.Shutdown(ctx); err != nil {
           log.Fatal(err)
       }
   }()
   ```