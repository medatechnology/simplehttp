// framework/fiber/server.go
package fiber

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/medatechnology/simplehttp"
)

const (
	HEADER_PARSED_KEY = "simplehttp.header"
)

type Server struct {
	app        *fiber.App
	config     *simplehttp.Config
	middleware []simplehttp.MedaMiddleware
	mu         sync.RWMutex
}

func NewServer(config *simplehttp.Config) *Server {
	if config == nil {
		config = simplehttp.DefaultConfig
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:           config.ConfigTimeOut.ReadTimeout,
		WriteTimeout:          config.ConfigTimeOut.WriteTimeout,
		IdleTimeout:           config.ConfigTimeOut.IdleTimeout,
		BodyLimit:             int(config.MaxRequestSize),
		DisableStartupMessage: !config.FrameworkStartupMessage,
		AppName:               "MedaHTTP/Fiber",
		Concurrency:           config.Concurrency, // Increase concurrency limit
		// Add explicit H2C configuration if needed
		// EnableH2C:             true,
	})

	return &Server{
		app:    app,
		config: config,
	}
}

func (s *Server) PrintMiddleware(verbose bool) {
	fmt.Printf("Registered Middlewares (%d)\n", len(s.middleware))
	if verbose {
		for i, m := range s.middleware {
			fmt.Printf("- %d:%s\n", i+1, m.Name())
		}
	}
}

func (s *Server) applyMiddleware(handler simplehttp.MedaHandlerFunc) simplehttp.MedaHandlerFunc {
	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i].Handle(handler)
	}
	return handler
}

func (s *Server) GET(path string, handler simplehttp.MedaHandlerFunc) {
	s.app.Get(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) POST(path string, handler simplehttp.MedaHandlerFunc) {
	s.app.Post(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) PUT(path string, handler simplehttp.MedaHandlerFunc) {
	s.app.Put(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) DELETE(path string, handler simplehttp.MedaHandlerFunc) {
	s.app.Delete(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) PATCH(path string, handler simplehttp.MedaHandlerFunc) {
	s.app.Patch(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) OPTIONS(path string, handler simplehttp.MedaHandlerFunc) {
	s.app.Options(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) HEAD(path string, handler simplehttp.MedaHandlerFunc) {
	s.app.Head(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) Static(prefix, root string) {
	s.app.Static(prefix, root, fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        true,
		Index:         "index.html",
		CacheDuration: s.config.ConfigTimeOut.IdleTimeout,
	})
}

func (s *Server) StaticFile(path, filepath string) {
	s.app.Static(path, filepath)
}

func (s *Server) WebSocket(path string, handler func(simplehttp.MedaWebsocket) error) {
	// Configure WebSocket route
	s.app.Use(path, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	s.app.Get(path, websocket.New(func(c *websocket.Conn) {
		wsWrapper := &FiberWebSocket{conn: c}
		if err := handler(wsWrapper); err != nil {
			c.Close()
		}
	}))
}

func (s *Server) Group(prefix string) simplehttp.MedaRouter {
	return &RouterGroup{
		prefix: prefix,
		server: s,
	}
}

func (s *Server) Use(middleware ...simplehttp.MedaMiddleware) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware = append(s.middleware, middleware...)
}

// Usually this is framework.Listen() function
// TODO: use config.Debug to print out or if not silence / minimal
func (s *Server) Start(address string) error {
	if address == "" {
		if s.config != nil {
			address = s.config.Hostname + ":" + s.config.Port
		} else {
			return fmt.Errorf("cannot start server with no address and config")
		}
	} else {
		// address passed is the port number
		if !strings.Contains(address, ":") {
			address = ":" + address
		}
	}

	// Get all routes for logging
	allRoutes := make(map[string]simplehttp.Routes)
	totalRoutes := 0

	// Iterate through the stack to get all routes
	for _, stack := range s.app.Stack() {
		for _, route := range stack {
			if route.Method != "" {
				// allRoutes = append(allRoutes, fmt.Sprintf("Route: %s\t %s", route.Method, route.Path))
				var r simplehttp.Routes
				r.EndPoint = route.Path
				r.Methods = append(r.Methods, route.Method)
				if _, ok := allRoutes[route.Path]; !ok {
					allRoutes[r.EndPoint] = r
				} else {
					tmpRoute := allRoutes[r.EndPoint]
					tmpRoute.Methods = append(tmpRoute.Methods, r.Methods...)
					allRoutes[r.EndPoint] = tmpRoute
				}
				totalRoutes++
			}
		}
	}

	// Print startup information based on debug mode
	if s.config.Debug {
		fmt.Printf("Starting MedaHTTP Fiber server...\n")
		fmt.Printf("With Server Configuration:\n")
		fmt.Printf("- Address: %s\n", address)
		fmt.Printf("- Read Timeout: %v\n", s.config.ConfigTimeOut.ReadTimeout)
		fmt.Printf("- Write Timeout: %v\n", s.config.ConfigTimeOut.WriteTimeout)
		fmt.Printf("- Idle Timeout: %v\n", s.config.ConfigTimeOut.IdleTimeout)
		fmt.Printf("- Max Request Size: %d bytes\n", s.config.MaxRequestSize)

		// Print middleware information
		if len(s.middleware) > 0 {
			s.PrintMiddleware(true)
			// fmt.Printf("Registered Middleware (%d):\n", len(s.middleware))
			// for i, mw := range s.middleware {
			// 	fmt.Printf("- Middleware #%d: %s\n", i+1, mw.Name())
			// }
		}

		// Print registered routes
		fmt.Printf("Registered routes/endpoints (%d/%d):\n", totalRoutes, len(allRoutes))
		for _, r := range allRoutes {
			fmt.Println(r.Sprint())
		}
	} else {
		fmt.Printf("MedaHTTP Fiber server starting on %s\n", address)
		s.PrintMiddleware(false)
		// fmt.Printf("Registered Middleware (%d)\n", len(s.middleware))
		fmt.Printf("Registered routes/endpoints (%d)\n", totalRoutes)
	}

	// Apply TLS if configured
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		return s.app.ListenTLS(address, s.config.TLSCert, s.config.TLSKey)
	}

	return s.app.Listen(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// RouterGroup implements group routing
type RouterGroup struct {
	prefix     string
	server     *Server
	middleware []simplehttp.MedaMiddleware
}

func (g *RouterGroup) applyMiddleware(handler simplehttp.MedaHandlerFunc) simplehttp.MedaHandlerFunc {
	// First apply group-specific middleware (in reverse order)
	for i := len(g.middleware) - 1; i >= 0; i-- {
		handler = g.middleware[i].Handle(handler)
	}

	// Then apply server-level middleware
	for i := len(g.server.middleware) - 1; i >= 0; i-- {
		handler = g.server.middleware[i].Handle(handler)
	}

	return handler
}

func (g *RouterGroup) GET(path string, handler simplehttp.MedaHandlerFunc) {
	g.server.app.Get(g.prefix+path, Adapter(g.applyMiddleware(handler)))
}

func (g *RouterGroup) POST(path string, handler simplehttp.MedaHandlerFunc) {
	g.server.app.Post(g.prefix+path, Adapter(g.applyMiddleware(handler)))
}

func (g *RouterGroup) PUT(path string, handler simplehttp.MedaHandlerFunc) {
	g.server.app.Put(g.prefix+path, Adapter(g.applyMiddleware(handler)))
}

func (g *RouterGroup) DELETE(path string, handler simplehttp.MedaHandlerFunc) {
	g.server.app.Delete(g.prefix+path, Adapter(g.applyMiddleware(handler)))
}

func (g *RouterGroup) PATCH(path string, handler simplehttp.MedaHandlerFunc) {
	g.server.app.Patch(g.prefix+path, Adapter(g.applyMiddleware(handler)))
}

func (g *RouterGroup) OPTIONS(path string, handler simplehttp.MedaHandlerFunc) {
	g.server.app.Options(g.prefix+path, Adapter(g.applyMiddleware(handler)))
}

func (g *RouterGroup) HEAD(path string, handler simplehttp.MedaHandlerFunc) {
	g.server.app.Head(g.prefix+path, Adapter(g.applyMiddleware(handler)))
}

func (g *RouterGroup) Static(prefix, root string) {
	g.server.Static(g.prefix+prefix, root)
}

func (g *RouterGroup) StaticFile(path, filepath string) {
	g.server.StaticFile(g.prefix+path, filepath)
}

func (g *RouterGroup) WebSocket(path string, handler func(simplehttp.MedaWebsocket) error) {
	// Apply middleware to WebSocket handler
	wrappedHandler := func(ws simplehttp.MedaWebsocket) error {
		return handler(ws)
	}

	g.server.WebSocket(g.prefix+path, wrappedHandler)
}

func (g *RouterGroup) Group(prefix string) simplehttp.MedaRouter {
	return &RouterGroup{
		prefix:     g.prefix + prefix,
		server:     g.server,
		middleware: make([]simplehttp.MedaMiddleware, 0),
	}
}

func (g *RouterGroup) Use(middleware ...simplehttp.MedaMiddleware) {
	g.middleware = append(g.middleware, middleware...)
}
