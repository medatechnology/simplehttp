// framework/fasthttp/server.go
package fasthttp

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/medatechnology/simplehttp"
	"github.com/valyala/fasthttp"
)

type Server struct {
	server     *fasthttp.Server
	config     *simplehttp.Config
	router     *router.Router
	middleware []simplehttp.Middleware
	mu         sync.RWMutex
}

func NewServer(config *simplehttp.Config) *Server {
	r := router.New()
	if config == nil {
		config = simplehttp.DefaultConfig
	}
	s := &Server{
		config: config,
		router: r,
		server: &fasthttp.Server{
			Handler:            r.Handler,
			ReadTimeout:        config.ConfigTimeOut.ReadTimeout,
			WriteTimeout:       config.ConfigTimeOut.WriteTimeout,
			IdleTimeout:        config.ConfigTimeOut.IdleTimeout,
			MaxRequestBodySize: int(config.MaxRequestSize),
			Name:               "MedaHTTP/FastHTTP",
		},
	}
	return s
}

func (s *Server) applyMiddleware(handler simplehttp.HandlerFunc) simplehttp.HandlerFunc {
	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i].Handle(handler)
	}
	return handler
}

func (s *Server) GET(path string, handler simplehttp.HandlerFunc) {
	s.router.GET(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) POST(path string, handler simplehttp.HandlerFunc) {
	s.router.POST(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) PUT(path string, handler simplehttp.HandlerFunc) {
	s.router.PUT(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) DELETE(path string, handler simplehttp.HandlerFunc) {
	s.router.DELETE(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) PATCH(path string, handler simplehttp.HandlerFunc) {
	s.router.PATCH(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) OPTIONS(path string, handler simplehttp.HandlerFunc) {
	s.router.OPTIONS(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) HEAD(path string, handler simplehttp.HandlerFunc) {
	s.router.HEAD(path, Adapter(s.applyMiddleware(handler)))
}

func (s *Server) Static(prefix, root string) {
	fs := &fasthttp.FS{
		Root:               root,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true,
		AcceptByteRange:    true,
		Compress:           true,
		PathRewrite: func(ctx *fasthttp.RequestCtx) []byte {
			path := ctx.Path()
			if len(path) >= len(prefix) {
				path = path[len(prefix):]
			}
			if len(path) == 0 {
				return []byte("/")
			}
			return path
		},
	}
	fsHandler := fs.NewRequestHandler()
	s.router.ANY(prefix+"/{filepath:*}", func(ctx *fasthttp.RequestCtx) {
		fsHandler(ctx)
	})
}

func (s *Server) StaticFile(path, filepath string) {
	s.router.GET(path, fasthttp.FSHandler(filepath, 0))
}

// WebSocket configuration
var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		// TODO: Implement proper origin checking based on config
		return true
	},
}

// FastHTTP WebSocket wrapper
type wsConn struct {
	*websocket.Conn
}

func (w *wsConn) WriteJSON(v interface{}) error {
	return w.Conn.WriteJSON(v)
}

func (w *wsConn) ReadJSON(v interface{}) error {
	return w.Conn.ReadJSON(v)
}

func (s *Server) WebSocket(path string, handler func(simplehttp.Websocket) error) {
	s.router.GET(path, func(ctx *fasthttp.RequestCtx) {
		err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
			wsWrapper := &wsConn{Conn: ws}
			if err := handler(wsWrapper); err != nil {
				ws.Close()
			}
		})
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	})
}

func (s *Server) Group(prefix string) simplehttp.Router {
	return &RouterGroup{
		prefix: prefix,
		server: s,
	}
}

func (s *Server) Use(middleware ...simplehttp.Middleware) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware = append(s.middleware, middleware...)
}

func (s *Server) Start(address string) error {
	if address == "" {
		if s.config != nil {
			address = s.config.Hostname + ":" + s.config.Port
		} else {
			return fmt.Errorf("cannot start server with no address and config")
		}
	} else {
		// this is assuming that address is only port number.
		// NOTE: How about if it's only hostname without port?
		if !strings.Contains(address, ":") {
			address = ":" + address
		}
	}

	// Get All Routes
	var allroutes []string
	totalroutes := 0
	for method, rmap := range s.router.List() {
		for _, endpoint := range rmap {
			allroutes = append(allroutes, fmt.Sprintf("Route: %s\t %s", method, endpoint))
			// fmt.Printf("Route: %s\t %s\n", method, endpoint)
			totalroutes++
		}
		// fmt.Printf("Route: %s %s\n", method, path)
	}

	// Print startup information based on debug mode
	if s.config.Debug {
		fmt.Printf("Starting MedaHTTP FastHTTP server...\n")
		fmt.Printf("With Server Configuration:\n")
		fmt.Printf("- Address: %s\n", address)
		fmt.Printf("- Read Timeout: %v\n", s.config.ConfigTimeOut.ReadTimeout)
		fmt.Printf("- Write Timeout: %v\n", s.config.ConfigTimeOut.WriteTimeout)
		fmt.Printf("- Idle Timeout: %v\n", s.config.ConfigTimeOut.IdleTimeout)
		fmt.Printf("- Max Request Size: %d bytes\n", s.config.MaxRequestSize)

		// Print middleware information
		if len(s.middleware) > 0 {
			fmt.Printf("Registered Middleware (%d):\n", len(s.middleware))
			for i, mw := range s.middleware {
				fmt.Printf("- Middleware #%d:%s\n", i+1, mw.Name())
			}
		}

		// Print registered routes
		fmt.Printf("Registered routes/endpoints (%d):\n", totalroutes)
		for _, r := range allroutes {
			fmt.Println(r)
		}
	} else {
		fmt.Printf("MedaHTTP FastHTTP server starting on %s\n", address)
		fmt.Printf("Registered Middleware (%d)\n", len(s.middleware))
		fmt.Printf("Registered routes/endpoints (%d)\n", totalroutes)
	}
	// Apply TLS if configured
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		return s.server.ListenAndServeTLS(address, s.config.TLSCert, s.config.TLSKey)
	}

	// Start server
	if s.config.Debug {
		fmt.Printf("Server is running on %s\n", address)
	}
	return s.server.ListenAndServe(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.ShutdownWithContext(ctx)
}

// RouterGroup implements group routing
type RouterGroup struct {
	prefix string
	server *Server
}

func (g *RouterGroup) GET(path string, handler simplehttp.HandlerFunc) {
	g.server.GET(g.prefix+path, handler)
}

func (g *RouterGroup) POST(path string, handler simplehttp.HandlerFunc) {
	g.server.POST(g.prefix+path, handler)
}

func (g *RouterGroup) PUT(path string, handler simplehttp.HandlerFunc) {
	g.server.PUT(g.prefix+path, handler)
}

func (g *RouterGroup) DELETE(path string, handler simplehttp.HandlerFunc) {
	g.server.DELETE(g.prefix+path, handler)
}

func (g *RouterGroup) PATCH(path string, handler simplehttp.HandlerFunc) {
	g.server.PATCH(g.prefix+path, handler)
}

func (g *RouterGroup) OPTIONS(path string, handler simplehttp.HandlerFunc) {
	g.server.OPTIONS(g.prefix+path, handler)
}

func (g *RouterGroup) HEAD(path string, handler simplehttp.HandlerFunc) {
	g.server.HEAD(g.prefix+path, handler)
}

func (g *RouterGroup) Static(prefix, root string) {
	g.server.Static(g.prefix+prefix, root)
}

func (g *RouterGroup) StaticFile(path, filepath string) {
	g.server.StaticFile(g.prefix+path, filepath)
}

func (g *RouterGroup) WebSocket(path string, handler func(simplehttp.Websocket) error) {
	g.server.WebSocket(g.prefix+path, handler)
}

func (g *RouterGroup) Group(prefix string) simplehttp.Router {
	return &RouterGroup{
		prefix: g.prefix + prefix,
		server: g.server,
	}
}

func (g *RouterGroup) Use(middleware ...simplehttp.Middleware) {
	g.server.Use(middleware...)
}
