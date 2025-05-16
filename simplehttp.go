package simplehttp

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
)

// Context represents our framework-agnostic request context
type Context interface {
	// Request information
	GetPath() string
	GetMethod() string
	GetHeader(key string) string
	GetHeaders() *RequestHeader
	SetRequestHeader(key, value string)
	SetResponseHeader(key, value string)
	SetHeader(key, value string)
	GetQueryParam(key string) string
	GetQueryParams() map[string][]string
	GetBody() []byte

	// Added these two methods
	Request() *http.Request
	Response() http.ResponseWriter

	// Response methods
	JSON(code int, data interface{}) error
	String(code int, data string) error
	Stream(code int, contentType string, reader io.Reader) error

	// File handling
	GetFile(fieldName string) (*multipart.FileHeader, error)
	SaveFile(file *multipart.FileHeader, dst string) error
	SendFile(filepath string, attachment bool) error

	// Websocket
	Upgrade() (Websocket, error)

	// Context handling
	Context() context.Context
	SetContext(ctx context.Context)
	Set(key string, value interface{})
	Get(key string) interface{}

	// Request binding
	Bind(interface{}) error // Generic binding based on Content-Type
	BindJSON(interface{}) error
	BindForm(interface{}) error
}

// Websocket interface for websocket connections
type Websocket interface {
	WriteJSON(v interface{}) error
	ReadJSON(v interface{}) error
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// HandlerFunc is our framework-agnostic handler function
type HandlerFunc func(Context) error

// MiddlewareFunc defines the contract for middleware
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Predefined common Middleware as global variables
type Middleware interface {
	Name() string
	Handle(HandlerFunc) HandlerFunc
}

// Router interface defines common routing operations
type Router interface {
	GET(path string, handler HandlerFunc)
	POST(path string, handler HandlerFunc)
	PUT(path string, handler HandlerFunc)
	DELETE(path string, handler HandlerFunc)
	PATCH(path string, handler HandlerFunc)
	OPTIONS(path string, handler HandlerFunc)
	HEAD(path string, handler HandlerFunc)

	// Static file serving
	Static(prefix, root string)
	StaticFile(path, filepath string)

	// Websocket
	WebSocket(path string, handler func(Websocket) error)

	Group(prefix string) Router
	Use(middleware ...Middleware)
}

// Server interface defines the contract for our web server
type Server interface {
	Router
	Start(address string) error
	Shutdown(ctx context.Context) error
}

// type newServerFunc func (*MedaConfig) (MedaServer, error)
// // Server factory function
// func NewMedaServer(config *MedaConfig, framework func(*MedaConfig) MedaServer) (MedaServer, error) {
// 	// Validate configuration
// 	if err := validateConfig(config); err != nil {
// 		return nil, err
// 	}

// 	// Create server based on framework
// 	return framework(config)
// 	// switch strings.ToLower(config.Framework) {
// 	// case "echo":
// 	// 	return newEchoServer(config), nil
// 	// case "fiber":
// 	// 	return newFiberServer(config), nil
// 	// case "gin":
// 	// 	return newGinServer(config), nil
// 	// default:
// 	// 	return nil, fmt.Errorf("unsupported framework: %s", config.Framework)
// 	// }
// }
