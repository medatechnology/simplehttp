package simplehttp

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
)

// MedaContext represents our framework-agnostic request context
type MedaContext interface {
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
	Upgrade() (MedaWebsocket, error)

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

// MedaWebsocket interface for websocket connections
type MedaWebsocket interface {
	WriteJSON(v interface{}) error
	ReadJSON(v interface{}) error
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// MedaHandlerFunc is our framework-agnostic handler function
type MedaHandlerFunc func(MedaContext) error

// MedaMiddlewareFunc defines the contract for middleware
type MedaMiddlewareFunc func(MedaHandlerFunc) MedaHandlerFunc

// Predefined common MedaMiddleware as global variables
type MedaMiddleware interface {
	Name() string
	Handle(MedaHandlerFunc) MedaHandlerFunc
}

// MedaRouter interface defines common routing operations
type MedaRouter interface {
	GET(path string, handler MedaHandlerFunc)
	POST(path string, handler MedaHandlerFunc)
	PUT(path string, handler MedaHandlerFunc)
	DELETE(path string, handler MedaHandlerFunc)
	PATCH(path string, handler MedaHandlerFunc)
	OPTIONS(path string, handler MedaHandlerFunc)
	HEAD(path string, handler MedaHandlerFunc)

	// Static file serving
	Static(prefix, root string)
	StaticFile(path, filepath string)

	// Websocket
	WebSocket(path string, handler func(MedaWebsocket) error)

	Group(prefix string) MedaRouter
	Use(middleware ...MedaMiddleware)
}

// MedaServer interface defines the contract for our web server
type MedaServer interface {
	MedaRouter
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
