// framework/fasthttp/context.go
package fasthttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"

	"github.com/medatechnology/goutil/filesystem"
	"github.com/medatechnology/goutil/object"
	"github.com/medatechnology/simplehttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type FHContext struct {
	ctx         *fasthttp.RequestCtx
	userContext context.Context
	store       map[string]interface{}
}

func NewContext(ctx *fasthttp.RequestCtx) *FHContext {
	return &FHContext{
		ctx:         ctx,
		userContext: context.Background(),
		store:       make(map[string]interface{}),
	}
}

func (c *FHContext) GetPath() string {
	return string(c.ctx.Path())
}

func (c *FHContext) GetMethod() string {
	return string(c.ctx.Method())
}

// check for both request and response header
func (c *FHContext) GetHeader(key string) string {
	// First check response headers (for headers set by middleware)
	if val := c.ctx.Response.Header.Peek(key); len(val) > 0 {
		return string(val)
	}
	return string(c.ctx.Request.Header.Peek(key))
}

func (c *FHContext) GetHeaders() *simplehttp.RequestHeader {
	var headers simplehttp.RequestHeader
	// Maybe already parsed in header!
	// TODO: this means that if it's already parsed earlier, then already added some key to header
	// then the changes are not seen...
	// if c.Get(simplehttp.HEADER_PARSED_STRING) == nil {
	// Convert fasthttp request to http.Request for header parsing
	r := &http.Request{
		Header: make(http.Header),
	}

	c.ctx.Request.Header.VisitAll(func(key, value []byte) {
		r.Header.Add(string(key), string(value))
	})
	// Add headers from response (which may have been set by middleware)
	c.ctx.Response.Header.VisitAll(func(key, value []byte) {
		r.Header.Add(string(key), string(value))
	})

	// NOTE: do not save any header in the store, this is for parsed header or
	// other info that we want to pass via context
	// Add any headers stored in context
	// for k, v := range c.store {
	// 	if strVal, ok := v.(string); ok {
	// 		r.Header.Set(k, strVal)
	// 	}
	// }
	headers.FromHttpRequest(r)
	// } else {
	// 	headers = c.Get(simplehttp.HEADER_PARSED_STRING).(simplehttp.RequestHeader)
	// }
	return &headers
}

func (c *FHContext) SetRequestHeader(key, value string) {
	c.ctx.Request.Header.Set(key, value)
}

func (c *FHContext) SetResponseHeader(key, value string) {
	c.ctx.Response.Header.Set(key, value)
}

func (c *FHContext) SetHeader(key, value string) {
	c.ctx.Request.Header.Set(key, value)
	c.ctx.Response.Header.Set(key, value)
}

func (c *FHContext) GetQueryParam(key string) string {
	return string(c.ctx.QueryArgs().Peek(key))
}

func (c *FHContext) GetQueryParams() map[string][]string {
	params := make(map[string][]string)
	c.ctx.QueryArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		params[k] = append(params[k], string(value))
	})
	return params
}

func (c *FHContext) GetBody() []byte {
	return c.ctx.Request.Body()
}

func (c *FHContext) Request() *http.Request {
	// Convert fasthttp request to http.Request
	var r http.Request
	fasthttpadaptor.ConvertRequest(c.ctx, &r, true)
	// r := &http.Request{
	// 	Method: string(c.ctx.Method()),
	// 	URL:    URIString2URL(c.ctx.URI().String()),
	// 	Header: make(http.Header),
	// 	Body:   io.NopCloser(c.ctx.Request.BodyStream()),
	// }

	// c.ctx.Request.Header.VisitAll(func(key, value []byte) {
	// 	r.Header.Add(string(key), string(value))
	// })

	return &r
}

func (c *FHContext) Response() http.ResponseWriter {
	// Create a wrapper that implements http.ResponseWriter
	// w := fasthttpadaptor.netHTTPResponseWriter{
	// 	w:   ctx.Response.BodyWriter(),
	// 	ctx: ctx,
	// }
	// return w
	return &responseWriter{ctx: c.ctx}
}

func (c *FHContext) JSON(code int, data interface{}) error {
	c.ctx.Response.Header.SetContentType("application/json")
	c.ctx.Response.SetStatusCode(code)
	return json.NewEncoder(c.ctx).Encode(data)
}

func (c *FHContext) String(code int, data string) error {
	c.ctx.Response.Header.SetContentType("text/plain")
	c.ctx.Response.SetStatusCode(code)
	c.ctx.WriteString(data)
	return nil
}

func (c *FHContext) Stream(code int, contentType string, reader io.Reader) error {
	c.ctx.Response.Header.SetContentType(contentType)
	c.ctx.Response.SetStatusCode(code)
	_, err := io.Copy(c.ctx, reader)
	return err
}

func (c *FHContext) GetFile(fieldName string) (*multipart.FileHeader, error) {
	form, err := c.ctx.MultipartForm()
	if err != nil {
		return nil, err
	}

	files := form.File[fieldName]
	if len(files) == 0 {
		return nil, fasthttp.ErrMissingFile
	}

	return files[0], nil
}

func (c *FHContext) SaveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func (c *FHContext) SendFile(filepath string, attachment bool) error {
	if attachment {
		c.ctx.Response.Header.Set("Content-Disposition", "attachment; filename="+filesystem.FileName(filepath))
	}
	c.ctx.SendFile(filepath)
	if c.ctx.Response.Header.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("error in SendFile %s", filepath)
	}
	return nil
}

func (c *FHContext) Upgrade() (simplehttp.MedaWebsocket, error) {
	// TODO: Implement WebSocket upgrade using fasthttp.Upgrader
	return nil, fmt.Errorf("websocket not implemented for fasthttp")
}

func (c *FHContext) Context() context.Context {
	return c.userContext
}

func (c *FHContext) SetContext(ctx context.Context) {
	c.userContext = ctx
}

func (c *FHContext) Set(key string, value interface{}) {
	c.store[key] = value
}

func (c *FHContext) Get(key string) interface{} {
	return c.store[key]
}

// Basic binding that supports query params, form data, and JSON body
func (c *FHContext) Bind(v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("binding element must be a pointer")
	}

	// Initialize params map with query parameters
	params := make(map[string]interface{}, c.ctx.QueryArgs().Len()) // Pre-allocate with known size
	c.ctx.QueryArgs().VisitAll(func(key, value []byte) {
		params[string(key)] = string(value)
	})

	// Handle body based on content type
	switch c.getBindingType() {
	case bindingJSON:
		var jsonData map[string]interface{}
		if err := c.BindJSON(&jsonData); err == nil {
			// Merge JSON data into params, pre-allocate the map if needed
			if len(jsonData) > len(params) {
				newParams := make(map[string]interface{}, len(jsonData))
				for k, v := range params {
					newParams[k] = v
				}
				params = newParams
			}
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

	// Get concrete type and convert
	result := object.MapToStruct[any](params)

	// Set the result back
	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(result))
	return nil
}

// func (c *FHContext) BindJSON(v interface{}) error {
// 	return json.Unmarshal(c.ctx.Request.Body(), v)
// }

// JSON-specific binding with memory optimization. This is only to get payload from Request Body!
func (c *FHContext) BindJSON(v interface{}) error {
	body := c.ctx.Request.Body()
	if len(body) == 0 {
		return fmt.Errorf("empty request body")
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber() // For better number handling
	return decoder.Decode(v)
}

// Form-specific binding with optimized map allocation
func (c *FHContext) BindForm(v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("binding element must be a pointer")
	}

	formData, err := c.getFormData()
	if err != nil {
		return err
	}

	result := object.MapToStruct[any](formData)
	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(result))
	return nil
}

// responseWriter implements http.ResponseWriter for fasthttp
type responseWriter struct {
	ctx *fasthttp.RequestCtx
}

func (w *responseWriter) Header() http.Header {
	h := make(http.Header)
	w.ctx.Response.Header.VisitAll(func(key, value []byte) {
		h.Add(string(key), string(value))
	})
	return h
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.ctx.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.ctx.Response.SetStatusCode(statusCode)
}
