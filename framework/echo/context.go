// framework/echo/context.go
package echo

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/medatechnology/simplehttp"
)

// EchoContext implements MedaContext interface using Echo's Context
type EchoContext struct {
	ctx    echo.Context
	config *simplehttp.Config
}

func NewEchoContext(c echo.Context, cfgs ...*simplehttp.Config) simplehttp.Context {
	// to enable optional parameter of configs, but in actual always pass 1
	if len(cfgs) > 0 && cfgs[0] != nil {
		return &EchoContext{ctx: c, config: cfgs[0]}
	}
	return &EchoContext{ctx: c}
}

func (c *EchoContext) GetPath() string {
	return c.ctx.Path()
}

func (c *EchoContext) GetMethod() string {
	return c.ctx.Request().Method
}

func (c *EchoContext) GetHeader(key string) string {
	return c.ctx.Request().Header.Get(key)
}

func (c *EchoContext) GetHeaders() *simplehttp.RequestHeader {
	headers := &simplehttp.RequestHeader{}
	headers.FromHttpRequest(c.ctx.Request())
	return headers
}

func (c *EchoContext) SetRequestHeader(key, value string) {
	c.ctx.Request().Header.Set(key, value)
}

func (c *EchoContext) SetResponseHeader(key, value string) {
	c.ctx.Response().Header().Set(key, value)
}

func (c *EchoContext) SetHeader(key, value string) {
	c.ctx.Request().Header.Set(key, value)
	c.ctx.Response().Header().Set(key, value)
}

func (c *EchoContext) GetQueryParam(key string) string {
	return c.ctx.QueryParam(key)
}

func (c *EchoContext) GetQueryParams() map[string][]string {
	return c.ctx.QueryParams()
}

func (c *EchoContext) GetBody() []byte {
	body, _ := io.ReadAll(c.ctx.Request().Body)
	return body
}

func (c *EchoContext) Request() *http.Request {
	return c.ctx.Request()
}

func (c *EchoContext) Response() http.ResponseWriter {
	return c.ctx.Response().Writer
}

func (c *EchoContext) JSON(code int, data interface{}) error {
	return c.ctx.JSON(code, data)
}

func (c *EchoContext) String(code int, data string) error {
	return c.ctx.String(code, data)
}

func (c *EchoContext) Stream(code int, contentType string, reader io.Reader) error {
	return c.ctx.Stream(code, contentType, reader)
}

func (c *EchoContext) GetFile(fieldName string) (*multipart.FileHeader, error) {
	return c.ctx.FormFile(fieldName)
}

func (c *EchoContext) SaveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	newfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer newfile.Close()

	_, err = io.Copy(newfile, src)
	return err
}

func (c *EchoContext) SendFile(filepath string, attachment bool) error {
	if attachment {
		return c.ctx.Attachment(filepath, filepath)
	}
	return c.ctx.File(filepath)
}

func (c *EchoContext) Upgrade() (simplehttp.Websocket, error) {
	var checkOrigin func(r *http.Request) bool

	// Configure origin checker based on CORS config if available
	if c.config != nil && c.config.ConfigCORS != nil {
		checkOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// Check against allowed origins
			for _, allowed := range c.config.ConfigCORS.AllowOrigins {
				if allowed == "*" || allowed == origin {
					return true
				}
			}
			return false
		}
	} else {
		// Default permissive check - you might want to make this more restrictive
		checkOrigin = func(r *http.Request) bool {
			return true
		}
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil, err
	}

	return &EchoWebSocket{conn: conn}, nil
}

func (c *EchoContext) Context() context.Context {
	return c.ctx.Request().Context()
}

func (c *EchoContext) SetContext(ctx context.Context) {
	c.ctx.SetRequest(c.ctx.Request().WithContext(ctx))
}

func (c *EchoContext) Set(key string, value interface{}) {
	c.ctx.Set(key, value)
}

func (c *EchoContext) Get(key string) interface{} {
	return c.ctx.Get(key)
}

func (c *EchoContext) Bind(v interface{}) error {
	return c.ctx.Bind(v)
}

func (c *EchoContext) BindJSON(i interface{}) error {
	return c.ctx.Bind(i)
}

func (c *EchoContext) BindForm(i interface{}) error {
	return c.ctx.Bind(i)
}

// EchoWebSocket implements MedaWebsocket interface using gorilla
type EchoWebSocket struct {
	conn *websocket.Conn
}

func (ws *EchoWebSocket) WriteJSON(v interface{}) error {
	return ws.conn.WriteJSON(v)
}

func (ws *EchoWebSocket) ReadJSON(v interface{}) error {
	return ws.conn.ReadJSON(v)
}

func (ws *EchoWebSocket) WriteMessage(messageType int, data []byte) error {
	return ws.conn.WriteMessage(messageType, data)
}

func (ws *EchoWebSocket) ReadMessage() (messageType int, p []byte, err error) {
	return ws.conn.ReadMessage()
}

func (ws *EchoWebSocket) Close() error {
	return ws.conn.Close()
}
