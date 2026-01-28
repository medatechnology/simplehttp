// framework/fiber/context.go
package fiber

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/medatechnology/goutil/object"
	"github.com/medatechnology/simplehttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type bindingType int

const (
	bindingNone bindingType = iota
	bindingJSON
	bindingForm

	// TODO : write the headers (response/request) in the context using fiber.local
	//        modify the SetRequestHeader and all to modify the one from context above
	// reqHeaderKey = "req_header"
	// resHeaderKey = "res_header"
)

type FiberContext struct {
	ctx         *fiber.Ctx
	userContext context.Context
}

func NewContext(c *fiber.Ctx) *FiberContext {
	return &FiberContext{
		ctx:         c,
		userContext: context.Background(),
	}
}

// Header manipulation methods
func (c *FiberContext) SetRequestHeader(key, value string) {
	c.ctx.Request().Header.Set(key, value)
}

func (c *FiberContext) SetResponseHeader(key, value string) {
	c.ctx.Response().Header.Set(key, value)
}

func (c *FiberContext) SetHeader(key, value string) {
	c.SetRequestHeader(key, value)
	c.SetResponseHeader(key, value)
}

func (c *FiberContext) GetHeader(key string) string {
	// return string(c.ctx.Request().Header.Peek(key))
	return c.ctx.Get(key)
}

func (c *FiberContext) GetHeaders() *simplehttp.RequestHeader {
	// Check if headers are already parsed and stored in context
	if headers, ok := c.ctx.Locals(HEADER_PARSED_KEY).(*simplehttp.RequestHeader); ok {
		return headers
	}
	var headers simplehttp.RequestHeader
	req := http.Request{}
	// This is just to have a http.Request format that are needed by headers.FromHttpHeader
	err := fasthttpadaptor.ConvertRequest(c.ctx.Context(), &req, true)
	if err != nil {
		req = http.Request{
			Header:     make(http.Header),
			RemoteAddr: c.ctx.IP(),
		}
		// Copy headers from Fiber request
		headerMap := c.ctx.GetReqHeaders()
		for key, value := range headerMap {
			for _, each := range value {
				req.Header.Add(key, each)
			}
		}
	}

	// c.ctx.Request().Header.VisitAll(func(key, value []byte) {
	// 	// fmt.Printf("%s,", key)
	// 	r.Header.Add(string(key), string(value))
	// })

	headers.FromHttpRequest(&req)

	// Just in case the IP is not there
	if headers.RemoteIP == "" {
		headers.RemoteIP = c.ctx.IP()
	}
	if headers.RealIP == "" {
		headers.RealIP = c.ctx.Get(simplehttp.HEADER_REAL_IP)
	}
	if headers.ConnectingIP == "" {
		headers.ConnectingIP = c.ctx.Get(simplehttp.HEADER_CONNECTING_IP)
	}
	if headers.TrueIP == "" {
		headers.TrueIP = c.ctx.Get(simplehttp.HEADER_TRUE_CLIENT_IP)
	}
	return &headers
}

// Standard http.Request and http.ResponseWriter implementation
func (c *FiberContext) Request() *http.Request {
	req := &http.Request{
		Method: c.ctx.Method(),
		URL: &url.URL{
			Scheme:   string(c.ctx.Request().URI().Scheme()),
			Host:     string(c.ctx.Request().URI().Host()),
			Path:     string(c.ctx.Request().URI().Path()),
			RawQuery: string(c.ctx.Request().URI().QueryString()),
		},
		Body:   io.NopCloser(bytes.NewReader(c.ctx.Body())),
		Header: make(http.Header),
	}

	// Copy current headers
	c.ctx.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	return req
}

type fiberResponseWriter struct {
	ctx *fiber.Ctx
}

func (w *fiberResponseWriter) Header() http.Header {
	h := make(http.Header)
	w.ctx.Response().Header.VisitAll(func(key, value []byte) {
		h.Add(string(key), string(value))
	})
	return h
}

func (w *fiberResponseWriter) Write(b []byte) (int, error) {
	return w.ctx.Write(b)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.ctx.Status(statusCode)
}

func (c *FiberContext) Response() http.ResponseWriter {
	return &fiberResponseWriter{ctx: c.ctx}
}

// Path and method accessors
func (c *FiberContext) GetPath() string {
	return c.ctx.Path()
}

func (c *FiberContext) GetMethod() string {
	return c.ctx.Method()
}

// Query parameter handling
func (c *FiberContext) GetQueryParam(key string) string {
	return c.ctx.Query(key)
}

func (c *FiberContext) GetQueryParams() map[string][]string {
	params := make(map[string][]string)
	c.ctx.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		params[k] = append(params[k], string(value))
	})
	return params
}

func (c *FiberContext) GetBody() []byte {
	return c.ctx.Body()
}

// Response methods
func (c *FiberContext) JSON(code int, data interface{}) error {
	return c.ctx.Status(code).JSON(data)
}

func (c *FiberContext) String(code int, data string) error {
	return c.ctx.Status(code).SendString(data)
}

func (c *FiberContext) Stream(code int, contentType string, reader io.Reader) error {
	c.ctx.Set("Content-Type", contentType)
	return c.ctx.Status(code).SendStream(reader)
}

// File handling
func (c *FiberContext) GetFile(fieldName string) (*multipart.FileHeader, error) {
	return c.ctx.FormFile(fieldName)
}

func (c *FiberContext) SaveFile(file *multipart.FileHeader, dst string) error {
	return c.ctx.SaveFile(file, dst)
}

func (c *FiberContext) SendFile(filepath string, attachment bool) error {
	if attachment {
		return c.ctx.Download(filepath)
	}
	return c.ctx.SendFile(filepath)
}

// SSE starts a Server-Sent Events stream
func (c *FiberContext) SSE(handler simplehttp.SSEHandler) error {
	writer := simplehttp.NewSSEWriter(c.Response())
	return handler(writer)
}

// WebSocket handling
func (c *FiberContext) Upgrade() (simplehttp.Websocket, error) {
	if !websocket.IsWebSocketUpgrade(c.ctx) {
		return nil, fmt.Errorf("not a websocket upgrade request")
	}
	return nil, fiber.ErrUpgradeRequired
}

// Context handling
func (c *FiberContext) Context() context.Context {
	return c.userContext
}

func (c *FiberContext) SetContext(ctx context.Context) {
	c.userContext = ctx
}

func (c *FiberContext) Set(key string, value interface{}) {
	c.ctx.Locals(key, value)
}

func (c *FiberContext) Get(key string) interface{} {
	return c.ctx.Locals(key)
}

// Binding implementation
func (c *FiberContext) getBindingType() bindingType {
	contentType := string(c.ctx.Request().Header.ContentType())
	switch {
	case strings.Contains(contentType, "application/json"):
		return bindingJSON
	case strings.Contains(contentType, "application/x-www-form-urlencoded"),
		strings.Contains(contentType, "multipart/form-data"):
		return bindingForm
	default:
		return bindingNone
	}
}

func (c *FiberContext) getFormData() (map[string]interface{}, error) {
	formData := make(map[string]interface{})

	// Handle multipart form
	if multipartForm, err := c.ctx.MultipartForm(); err == nil && multipartForm != nil {
		for key, values := range multipartForm.Value {
			if len(values) > 0 {
				formData[key] = values[0]
			}
		}
	}

	// Handle regular form
	c.ctx.Request().PostArgs().VisitAll(func(key, value []byte) {
		formData[string(key)] = string(value)
	})

	return formData, nil
}

func (c *FiberContext) Bind(v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("binding element must be a pointer")
	}

	// Initialize params map with query parameters
	params := make(map[string]interface{})
	c.ctx.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		params[string(key)] = string(value)
	})

	// Handle body based on content type
	switch c.getBindingType() {
	case bindingJSON:
		var jsonData map[string]interface{}
		if err := c.BindJSON(&jsonData); err == nil {
			for k, v := range jsonData {
				params[k] = v
			}
		}
	case bindingForm:
		if formData, err := c.getFormData(); err == nil {
			for k, v := range formData {
				params[k] = v
			}
		}
	}

	result := object.MapToStruct[any](params)
	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(result))
	return nil
}

func (c *FiberContext) BindJSON(v interface{}) error {
	return c.ctx.BodyParser(v)
}

func (c *FiberContext) BindForm(v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("binding element must be a pointer")
	}

	formData, err := c.getFormData()
	if err != nil {
		return err
	}

	// If formData is empty, return immediately
	if len(formData) == 0 {
		return nil
	}
	// TODO : if we have time, improve the reflection and make it dynamic based on tag
	result := object.MapToStruct[any](formData)
	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(result))
	return nil
}

// Path parameters (Phase 1 - v0.1.0)
func (c *FiberContext) GetParam(key string) string {
	return c.ctx.Params(key)
}

func (c *FiberContext) GetParams() map[string]string {
	params := make(map[string]string)
	// Fiber stores params differently - iterate through all params
	if c.ctx.Route() != nil {
		for _, param := range c.ctx.Route().Params {
			params[param] = c.ctx.Params(param)
		}
	}
	return params
}

func (c *FiberContext) GetParamInt(key string) (int, error) {
	value := c.ctx.Params(key)
	if value == "" {
		return 0, simplehttp.ErrParamNotFound
	}
	return strconv.Atoi(value)
}

func (c *FiberContext) GetParamInt64(key string) (int64, error) {
	value := c.ctx.Params(key)
	if value == "" {
		return 0, simplehttp.ErrParamNotFound
	}
	return strconv.ParseInt(value, 10, 64)
}

// Cookie handling (Phase 1 - v0.1.0)
func (c *FiberContext) GetCookie(name string) (string, error) {
	value := c.ctx.Cookies(name)
	if value == "" {
		return "", fmt.Errorf("cookie not found: %s", name)
	}
	return value, nil
}

func (c *FiberContext) SetCookie(cookie *http.Cookie) {
	fiberCookie := &fiber.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		MaxAge:   cookie.MaxAge,
		Expires:  cookie.Expires,
		Secure:   cookie.Secure,
		HTTPOnly: cookie.HttpOnly,
		SameSite: string(cookie.SameSite),
	}
	c.ctx.Cookie(fiberCookie)
}

func (c *FiberContext) SetCookieSimple(name, value string, maxAge int) {
	c.ctx.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HTTPOnly: true,
	})
}

func (c *FiberContext) DeleteCookie(name string) {
	c.ctx.Cookie(&fiber.Cookie{
		Name:    name,
		Value:   "",
		MaxAge:  -1,
		Expires: time.Now().Add(-time.Hour),
		Path:    "/",
	})
}

// Redirect methods (Phase 1 - v0.1.0)
func (c *FiberContext) Redirect(code int, url string) error {
	return c.ctx.Redirect(url, code)
}

func (c *FiberContext) RedirectPermanent(url string) error {
	return c.ctx.Redirect(url, http.StatusMovedPermanently)
}

func (c *FiberContext) RedirectTemporary(url string) error {
	return c.ctx.Redirect(url, http.StatusFound)
}

// Form values (Phase 1 - v0.1.0)
func (c *FiberContext) GetFormValue(key string) string {
	return c.ctx.FormValue(key)
}

func (c *FiberContext) GetFormValues() map[string][]string {
	form := make(map[string][]string)
	c.ctx.Request().PostArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		if _, exists := form[keyStr]; !exists {
			form[keyStr] = []string{}
		}
		form[keyStr] = append(form[keyStr], string(value))
	})
	return form
}

// Response status helpers (Phase 1 - v0.1.0)
func (c *FiberContext) NoContent() error {
	return c.ctx.SendStatus(http.StatusNoContent)
}

func (c *FiberContext) NotFound() error {
	return c.ctx.Status(http.StatusNotFound).JSON(map[string]string{
		"error": "Not Found",
	})
}

func (c *FiberContext) BadRequest(message string) error {
	return c.ctx.Status(http.StatusBadRequest).JSON(map[string]string{
		"error": message,
	})
}

func (c *FiberContext) Unauthorized(message string) error {
	return c.ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
		"error": message,
	})
}

func (c *FiberContext) Forbidden(message string) error {
	return c.ctx.Status(http.StatusForbidden).JSON(map[string]string{
		"error": message,
	})
}

func (c *FiberContext) InternalServerError(message string) error {
	return c.ctx.Status(http.StatusInternalServerError).JSON(map[string]string{
		"error": message,
	})
}

// WebSocket implementation
type FiberWebSocket struct {
	conn *websocket.Conn
}

func NewFiberWebSocket(c *websocket.Conn) *FiberWebSocket {
	return &FiberWebSocket{conn: c}
}

func (ws *FiberWebSocket) WriteJSON(v interface{}) error {
	return ws.conn.WriteJSON(v)
}

func (ws *FiberWebSocket) ReadJSON(v interface{}) error {
	return ws.conn.ReadJSON(v)
}

func (ws *FiberWebSocket) WriteMessage(messageType int, data []byte) error {
	return ws.conn.WriteMessage(messageType, data)
}

func (ws *FiberWebSocket) ReadMessage() (messageType int, p []byte, err error) {
	return ws.conn.ReadMessage()
}

func (ws *FiberWebSocket) Close() error {
	return ws.conn.Close()
}
