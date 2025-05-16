# Framework Implementation Guide

This guide explains how to implement support for a new web framework in SimpleHTTP. The abstraction layer is designed to make it easy to add new frameworks while maintaining a consistent API.

## Implementation Structure

To implement a new framework, you need to create three main components:

1. **Server**: Implements the SimpleHttp `Server` interface
2. **Context**: Implements the SimpleHttp `Context` interface
3. **Adapter**: Converts between SimpleHttpHTTP and the framework's handlers

## Step-by-Step Guide

Let's walk through implementing support for a new framework (e.g., Gin):

### 1. Create the Directory Structure

```
simplehttp/
  ├── framework/
      ├── gin/
          ├── server.go    # SimpleHttp Server implementation
          ├── context.go   # SimpleHttp Context implementation
          ├── adapter.go   # Adapter functions
```

### 2. Implement the Context

The context is the core of your implementation. It wraps the framework's native context and provides the methods defined in the SimpleHttp `Context` interface.

```go
// framework/gin/context.go
package gin

import (
    "context"
    "io"
    "mime/multipart"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/medatechnology/simplehttp"
)

type GinContext struct {
    ctx         *gin.Context
    userContext context.Context
}

func NewContext(c *gin.Context) *GinContext {
    return &GinContext{
        ctx:         c,
        userContext: context.Background(),
    }
}

// Implement all methods from SimpleHttp Context interface
func (c *GinContext) GetPath() string {
    return c.ctx.FullPath()
}

func (c *GinContext) GetMethod() string {
    return c.ctx.Request.Method
}

func (c *GinContext) GetHeader(key string) string {
    return c.ctx.GetHeader(key)
}

func (c *GinContext) GetHeaders() *simplehttp.RequestHeader {
    var headers simplehttp.RequestHeader
    headers.FromHttpHeader(c.ctx.Request)
    return &headers
}

func (c *GinContext) SetRequestHeader(key, value string) {
    c.ctx.Request.Header.Set(key, value)
}

func (c *GinContext) SetResponseHeader(key, value string) {
    c.ctx.Writer.Header().Set(key, value)
}

func (c *GinContext) SetHeader(key, value string) {
    c.SetRequestHeader(key, value)
    c.SetResponseHeader(key, value)
}

// Implement remaining methods...
// JSON, String, Stream responses
// File handling
// WebSocket integration
// Context handling
// Data binding

// Example of a JSON response implementation
func (c *GinContext) JSON(code int, data interface{}) error {
    c.ctx.JSON(code, data)
    return nil
}

// Example of a Get/Set implementation
func (c *GinContext) Set(key string, value interface{}) {
    c.ctx.Set(key, value)
}

func (c *GinContext) Get(key string) interface{} {
    value, _ := c.ctx.Get(key)
    return value
}

// Request/Response standard http objects
func (c *GinContext) Request() *http.Request {
    return c.ctx.Request
}

func (c *GinContext) Response() http.ResponseWriter {
    return c.ctx.Writer
}

// Context implementation
func (c *GinContext) Context() context.Context {
    return c.userContext
}

func (c *GinContext) SetContext(ctx context.Context) {
    c.userContext = ctx
}
```

### 3. Implement the Adapter

The adapter converts between SimpleHTTP handlers and the framework's native handlers.

```go
// framework/gin/adapter.go
package gin

import (
    "github.com/gin-gonic/gin"
    "github.com/medatechnology/simplehttp"
)

// Adapter converts SimpleHTTP HandlerFunc to gin.HandlerFunc
func Adapter(handler simplehttp.HandlerFunc) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := NewContext(c)
        if err := handler(ctx); err != nil {
            handleError(ctx, err)
        }
    }
}

// handleError processes errors and sends appropriate responses
func handleError(c *GinContext, err error) {
    if medaErr, ok := err.(*simplehttp.SimpleHttpError); ok {
        c.ctx.JSON(medaErr.Code, medaErr)
    } else {
        c.ctx.JSON(500, map[string]string{"error": err.Error()})
    }
}

// Optional: middleware adapter
func MiddlewareAdapter(middleware simplehttp.MiddlewareFunc) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := NewContext(c)
        err := middleware(func(medaCtx simplehttp.Context) error {
            c.Next()
            return nil
        })(ctx)
        
        if err != nil {
            handleError(ctx, err)
        }
    }
}
```

### 4. Implement the Server

The server implements the `SimpleHttp Server` interface and manages the lifecycle of the web application.

```go
// framework/gin/server.go
package gin

import (
    "context"
    "sync"
    
    "github.com/gin-gonic/gin"
    "github.com/medatechnology/simplehttp"
)

type Server struct {
    engine     *gin.Engine
    config     *simplehttp.Config
    middleware []simplehttp.Middleware
    mu         sync.RWMutex
}

func NewServer(config *simplehttp.Config) *Server {
    if config == nil {
        config = simplehttp.DefaultConfig
    }
    
    // Set Gin mode based on config
    if config.Debug {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }
    
    engine := gin.New()
    
    return &Server{
        engine: engine,
        config: config,
    }
}

// Apply middleware to a handler
func (s *Server) applyMiddleware(handler simplehttp.HandlerFunc) simplehttp.HandlerFunc {
    for i := len(s.middleware) - 1; i >= 0; i-- {
        handler = s.middleware[i].Handle(handler)
    }
    return handler
}

// Implement all methods from SimpleHTTP Router interface
func (s *Server) GET(path string, handler simplehttp.HandlerFunc) {
    s.engine.GET(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) POST(path string, handler simplehttp.HandlerFunc) {
    s.engine.POST(path, Adapter(s.applyMiddleware(handler)))
}

// Implement remaining HTTP methods...

// Implement Group
func (s *Server) Group(prefix string) simplehttp.Router {
    return &RouterGroup{
        prefix: prefix,
        server: s,
    }
}

// Implement Use
func (s *Server) Use(middleware ...simplehttp.Middleware) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.middleware = append(s.middleware, middleware...)
}

// Implement server lifecycle methods
func (s *Server) Start(address string) error {
    if address == "" {
        address = s.config.Hostname + ":" + s.config.Port
    }
    
    return s.engine.Run(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
    server := &http.Server{
        Addr:    s.config.Hostname + ":" + s.config.Port,
        Handler: s.engine,
    }
    return server.Shutdown(ctx)
}

// Implement static file methods
func (s *Server) Static(prefix, root string) {
    s.engine.Static(prefix, root)
}

func (s *Server) StaticFile(path, filepath string) {
    s.engine.StaticFile(path, filepath)
}

// Implement WebSocket (this would depend on your WebSocket integration)
func (s *Server) WebSocket(path string, handler func(simplehttp.WebSocket) error) {
    // Implementation depends on the WebSocket library you use with Gin
}

// Router Group implementation
type RouterGroup struct {
    prefix string
    server *Server
}

// Implement all SimpleHTTP Router methods for RouterGroup...
func (g *RouterGroup) GET(path string, handler simplehttp.HandlerFunc) {
    g.server.GET(g.prefix+path, handler)
}

// Implement remaining methods...
```

### 5. WebSocket Implementation

If your framework supports WebSockets, you'll need to implement the SimpleHTTP `Websocket` interface as well:

```go
// framework/gin/websocket.go (or include in context.go)
package gin

import (
    "github.com/gorilla/websocket"
    "github.com/medatechnology/simplehttp"
)

// GinWebSocket implements the SimpleHTTP Websocket interface
type GinWebSocket struct {
    conn *websocket.Conn
}

func NewGinWebSocket(conn *websocket.Conn) *GinWebSocket {
    return &GinWebSocket{conn: conn}
}

func (ws *GinWebSocket) WriteJSON(v interface{}) error {
    return ws.conn.WriteJSON(v)
}

func (ws *GinWebSocket) ReadJSON(v interface{}) error {
    return ws.conn.ReadJSON(v)
}

func (ws *GinWebSocket) WriteMessage(messageType int, data []byte) error {
    return ws.conn.WriteMessage(messageType, data)
}

func (ws *GinWebSocket) ReadMessage() (messageType int, p []byte, err error) {
    return ws.conn.ReadMessage()
}

func (ws *GinWebSocket) Close() error {
    return ws.conn.Close()
}
```

## Testing Your Implementation

Create a simple test program to verify your implementation:

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/medatechnology/simplehttp"
    "github.com/medatechnology/simplehttp/framework/gin" // Your new implementation
)

func main() {
    config := simplehttp.LoadConfig()
    server := gin.NewServer(config)
    
    server.Use(
        simplehttp.MiddlewareRequestID(),
        simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
    )
    
    server.GET("/hello", func(c simplehttp.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "message": "Hello from Gin implementation!",
        })
    })
    
    if err := server.Start(""); err != nil {
        log.Fatal(err)
    }
}
```

## Best Practices

1. **Follow Existing Patterns**: Review the existing implementations to maintain consistency
2. **Complete Implementation**: Ensure all methods of the interfaces are implemented
3. **Error Handling**: Properly convert between framework-specific errors and SimpleHTTP errors
4. **Default Configuration**: Provide sensible defaults for all framework-specific settings
5. **Documentation**: Add framework-specific notes to your implementation

## Final Steps

1. Add your implementation to the main README.md
2. Create tests for your implementation
3. Submit a pull request with your changes

By following this guide, you should be able to implement support for any web framework in SimpleHTTP.