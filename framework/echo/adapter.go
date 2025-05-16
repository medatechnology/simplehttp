// framework/echo/adapter.go
package echo

import (
	"github.com/labstack/echo/v5"
	"github.com/medatechnology/simplehttp"
)

// Adapter converts SimpleHttp HandlerFunc to echo.HandlerFunc
func Adapter(handler simplehttp.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handler(NewEchoContext(c))
	}
}

// MiddlewareAdapter converts SimpleHttp Middleware to echo.MiddlewareFunc
func MiddlewareAdapter(middleware simplehttp.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			medaNext := func(mc simplehttp.Context) error {
				return next(c)
			}
			return middleware(medaNext)(NewEchoContext(c))
		}
	}
}
