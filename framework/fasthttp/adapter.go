// framework/fasthttp/adapter.go
package fasthttp

import (
	"net/url"
	"strings"

	"github.com/medatechnology/simplehttp"
	"github.com/valyala/fasthttp"
)

type bindingType int

const (
	bindingNone bindingType = iota
	bindingJSON
	bindingForm
)

// Adapter converts SimpleHttp HandlerFunc to fasthttp.RequestHandler
func Adapter(handler simplehttp.HandlerFunc) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		c := NewContext(ctx)
		if err := handler(c); err != nil {
			handleError(c, err)
		}
		// return handler(NewContext(ctx))
	}
}

// MiddlewareAdapter converts SimpleHttp Middleware to fasthttp middleware
func MiddlewareAdapter(middleware simplehttp.MiddlewareFunc) func(fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return Adapter(middleware(func(c simplehttp.Context) error {
			ctx := c.(*FHContext).ctx
			next(ctx)
			return nil
		}))
	}
}

// handleError processes errors and sends appropriate responses
func handleError(c *FHContext, err error) {
	if medaErr, ok := err.(*simplehttp.SimpleHttpError); ok {
		c.JSON(medaErr.Code, medaErr)
		return
	}

	// Default error response
	c.JSON(500, map[string]string{
		"error": err.Error(),
	})
}

// Convert fasthttp URI to net/url skipping the error!
func URIString2URL(uristring string) *url.URL {
	u, _ := url.ParseRequestURI(uristring)
	return u
}

// getBindingType returns the appropriate binding type based on Content-Type
func (c *FHContext) getBindingType() bindingType {
	contentType := string(c.ctx.Request.Header.ContentType())
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

// Optimized helper function to get form data as map
func (c *FHContext) getFormData() (map[string]interface{}, error) {
	// Pre-allocate map based on PostArgs length
	formData := make(map[string]interface{}, c.ctx.PostArgs().Len())

	// Handle regular form values
	c.ctx.PostArgs().VisitAll(func(key, value []byte) {
		formData[string(key)] = string(value)
	})

	// Handle multipart form if present
	if form, err := c.ctx.MultipartForm(); err == nil && form != nil {
		// If we have multipart form values, create a new map with larger size
		if len(form.Value) > 0 {
			newData := make(map[string]interface{}, len(formData)+len(form.Value))
			// Copy existing data
			for k, v := range formData {
				newData[k] = v
			}
			formData = newData
		}

		// Add form values
		for key, values := range form.Value {
			if len(values) > 0 {
				formData[key] = values[0]
			}
		}
	}

	return formData, nil
}
