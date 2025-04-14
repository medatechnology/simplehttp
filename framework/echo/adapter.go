// framework/echo/adapter.go
package echo

import (
	"github.com/labstack/echo/v5"
	"github.com/medatechnology/simplehttp"
)

// Adapter converts MedaHandlerFunc to echo.HandlerFunc
func Adapter(handler simplehttp.MedaHandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handler(NewEchoContext(c))
	}
}

// MiddlewareAdapter converts MedaMiddleware to echo.MiddlewareFunc
func MiddlewareAdapter(middleware simplehttp.MedaMiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			medaNext := func(mc simplehttp.MedaContext) error {
				return next(c)
			}
			return middleware(medaNext)(NewEchoContext(c))
		}
	}
}
