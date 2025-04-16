// framework/echo/server.go
package echo

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/medatechnology/simplehttp"
)

type EchoServer struct {
	e      *echo.Echo
	config *simplehttp.Config
	// router *EchoGroup
	middleware []simplehttp.MedaMiddleware
	// mu         sync.RWMutex
}

func NewServer(config *simplehttp.Config) simplehttp.MedaServer {
	e := echo.New()

	// Basic middleware setup
	e.Use(middleware.Recover())
	if config.Debug {
		e.Use(middleware.Logger())
	}

	// Set max request size
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.JSONSerializer = echo.DefaultJSONSerializer{}

	return &EchoServer{
		e:      e,
		config: config,
	}
}

func (s *EchoServer) GET(path string, handler simplehttp.MedaHandlerFunc) {
	s.e.GET(path, Adapter(handler))
}

func (s *EchoServer) POST(path string, handler simplehttp.MedaHandlerFunc) {
	s.e.POST(path, Adapter(handler))
}

func (s *EchoServer) PUT(path string, handler simplehttp.MedaHandlerFunc) {
	s.e.PUT(path, Adapter(handler))
}

func (s *EchoServer) DELETE(path string, handler simplehttp.MedaHandlerFunc) {
	s.e.DELETE(path, Adapter(handler))
}

func (s *EchoServer) PATCH(path string, handler simplehttp.MedaHandlerFunc) {
	s.e.PATCH(path, Adapter(handler))
}

func (s *EchoServer) OPTIONS(path string, handler simplehttp.MedaHandlerFunc) {
	s.e.OPTIONS(path, Adapter(handler))
}

func (s *EchoServer) HEAD(path string, handler simplehttp.MedaHandlerFunc) {
	s.e.HEAD(path, Adapter(handler))
}

func (s *EchoServer) Static(prefix, root string) {
	s.e.Static(prefix, root)
}

func (s *EchoServer) StaticFile(path, filepath string) {
	s.e.FileFS(path, filepath, nil)
}

func (s *EchoServer) WebSocket(path string, handler func(simplehttp.MedaWebsocket) error) {
	s.e.GET(path, func(c echo.Context) error {
		echoCtx := NewEchoContext(c, s.config)
		ws, err := echoCtx.Upgrade()
		if err != nil {
			return err
		}
		return handler(ws)
	})
}

func (s *EchoServer) Group(prefix string) simplehttp.MedaRouter {
	group := s.e.Group(prefix)
	return &EchoGroup{group: group, config: s.config}
}

func (s *EchoServer) Use(middleware ...simplehttp.MedaMiddleware) {
	for _, m := range middleware {
		s.e.Use(MiddlewareAdapter(m.Handle))
	}
}

func (s *EchoServer) Start(address string) error {
	return s.e.Start(fmt.Sprintf(":%s", s.config.Port))
}

// Shutdown is a no-op in Echo v5 as it's handled internally
func (s *EchoServer) Shutdown(ctx context.Context) error {
	// Echo v5 handles graceful shutdown internally
	return nil
}

// EchoGroup implements MedaRouter interface for route groups
type EchoGroup struct {
	group  *echo.Group
	config *simplehttp.Config
}

func (g *EchoGroup) GET(path string, handler simplehttp.MedaHandlerFunc) {
	g.group.GET(path, Adapter(handler))
}

func (g *EchoGroup) POST(path string, handler simplehttp.MedaHandlerFunc) {
	g.group.POST(path, Adapter(handler))
}

func (g *EchoGroup) PUT(path string, handler simplehttp.MedaHandlerFunc) {
	g.group.PUT(path, Adapter(handler))
}

func (g *EchoGroup) DELETE(path string, handler simplehttp.MedaHandlerFunc) {
	g.group.DELETE(path, Adapter(handler))
}

func (g *EchoGroup) PATCH(path string, handler simplehttp.MedaHandlerFunc) {
	g.group.PATCH(path, Adapter(handler))
}

func (g *EchoGroup) OPTIONS(path string, handler simplehttp.MedaHandlerFunc) {
	g.group.OPTIONS(path, Adapter(handler))
}

func (g *EchoGroup) HEAD(path string, handler simplehttp.MedaHandlerFunc) {
	g.group.HEAD(path, Adapter(handler))
}

func (g *EchoGroup) Static(prefix, root string) {
	g.group.Static(prefix, root)
}

func (g *EchoGroup) StaticFile(path, filepath string) {
	g.group.FileFS(path, filepath, nil)
}

func (g *EchoGroup) WebSocket(path string, handler func(simplehttp.MedaWebsocket) error) {
	g.group.GET(path, func(c echo.Context) error {
		medaCtx := NewEchoContext(c, g.config)
		ws, err := medaCtx.Upgrade()
		if err != nil {
			return err
		}
		return handler(ws)
	})
}

func (g *EchoGroup) Group(prefix string) simplehttp.MedaRouter {
	subgroup := g.group.Group(prefix)
	return &EchoGroup{group: subgroup, config: g.config}
}

func (g *EchoGroup) Use(middleware ...simplehttp.MedaMiddleware) {
	for _, m := range middleware {
		g.group.Use(MiddlewareAdapter(m.Handle))
	}
}
