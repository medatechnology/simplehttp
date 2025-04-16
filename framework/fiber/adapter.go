// framework/fiber/adapter.go
package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/medatechnology/simplehttp"
)

// type bindingType int

// const (
// 	bindingNone bindingType = iota
// 	bindingJSON
// 	bindingForm
// )

// Adapter converts MedaHandlerFunc to fiber.Handler
func Adapter(handler simplehttp.MedaHandlerFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := NewContext(c)
		if err := handler(ctx); err != nil {
			return handleError(ctx, err)
		}
		return nil
	}
}

// MiddlewareAdapter converts MedaMiddleware to fiber middleware
func MiddlewareAdapter(middleware simplehttp.MedaMiddlewareFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := NewContext(c)
		err := middleware(func(medaCtx simplehttp.MedaContext) error {
			return c.Next()
		})(ctx)
		return err
	}
}

// handleError processes errors and sends appropriate responses
func handleError(c *FiberContext, err error) error {
	if medaErr, ok := err.(*simplehttp.MedaError); ok {
		return c.JSON(medaErr.Code, medaErr)
	}
	return c.JSON(500, map[string]string{"error": err.Error()})
}

// getBindingType returns the appropriate binding type based on Content-Type
// func (c *FiberContext) getBindingType() bindingType {
// 	contentType := string(c.ctx.Request().Header.ContentType())
// 	switch {
// 	case strings.Contains(contentType, "application/json"):
// 		return bindingJSON
// 	case strings.Contains(contentType, "application/x-www-form-urlencoded"),
// 		strings.Contains(contentType, "multipart/form-data"):
// 		return bindingForm
// 	default:
// 		return bindingNone
// 	}
// }

// // getFormData returns form data as a map
// func (c *FiberContext) getFormData() (map[string]interface{}, error) {
// 	formData := make(map[string]interface{})

// 	// Handle multipart form
// 	if multipartForm, err := c.ctx.MultipartForm(); err == nil && multipartForm != nil {
// 		for key, values := range multipartForm.Value {
// 			if len(values) > 0 {
// 				formData[key] = values[0]
// 			}
// 		}
// 	}

// 	// Handle regular form
// 	c.ctx.Request().PostArgs().VisitAll(func(key, value []byte) {
// 		formData[string(key)] = string(value)
// 	})

// 	return formData, nil
// }
