// framework/fiber/middleware.go
package fiber

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/medatechnology/simplehttp"
)

const (
	// Logger constants
	defaultLogFormat     = "[${time}] ${status} - ${latency} ${method} ${path}\n"
	defaultLogTimeFormat = "2006/01/02 15:04:05"
	defaultTimeZone      = "Local"

	// Security header constants
	xFrameOptionsDeny       = "DENY"
	xFrameOptionsSameOrigin = "SAMEORIGIN"
	xssProtectionEnabled    = "1; mode=block"
	xssProtectionDisabled   = "0"
	contentTypeNosniffValue = "nosniff"
	defaultReferrerPolicy   = "strict-origin-when-cross-origin"
	defaultHSTSMaxAge       = 31536000
	defaultHSTSPreload      = true

	// CORS constants
	defaultCORSAllowOrigins  = "*"
	defaultCORSAllowMethods  = "GET,POST,HEAD,PUT,DELETE,PATCH"
	defaultCORSAllowHeaders  = ""
	defaultCORSExposeHeaders = ""
	defaultCORSMaxAge        = 24 * time.Hour

	// CSRF constants
	csrfHeaderName     = "X-CSRF-Token"
	csrfHeaderLookup   = "header:" + csrfHeaderName
	csrfCookiePrefix   = "csrf_"
	csrfCookieSameSite = "Strict"
	csrfExpiration     = 1 * time.Hour

	// Context keys
	requestIDContextKey = "requestid"

	// Cache constants
	// defaultCacheEnabled = true

	// Rate Limiter constants
	defaultLimitSkipFailed    = false
	defaultLimitSkipSucceeded = false

	// ETag constants
	defaultETagWeak = true

	// Recovery constants
	defaultStackTraceEnabled = true

	// Compression constants
	defaultCompressLevel = 1

	// Basic Auth constants
	defaultAuthRealm = "Restricted"

	// SimpleHTTP NextHandlerKey
	// simpleHTTPNextHandlerKey = "meda_next_handler"
)

var (
	requestIDHeaderKey = simplehttp.HEADER_REQUEST_ID
)

// namedMiddleware wraps a Fiber middleware with a name
type namedMiddleware struct {
	name       string
	middleware fiber.Handler
}

func (m namedMiddleware) Name() string {
	return m.name
}

func (m namedMiddleware) Handle(next simplehttp.HandlerFunc) simplehttp.HandlerFunc {
	return func(c simplehttp.Context) error {
		fiberCtx := c.(*FiberContext).ctx

		// Create a wrapper handler that will be called after middleware
		wrappedNext := func(c *fiber.Ctx) error {
			return next(&FiberContext{ctx: c})
		}

		// Store the handler
		fiberCtx.Locals("nextHandler", wrappedNext)

		return m.middleware(fiberCtx)
	}
}

// RequestID middleware as an example., TODO: Check and test this, last time it wasn't working!
func MiddlewareRequestID() simplehttp.Middleware {
	return namedMiddleware{
		name: "request ID",
		middleware: func(c *fiber.Ctx) error {
			// Run the actual requestid middleware
			if err := requestid.New(requestid.Config{
				Header:     requestIDHeaderKey,
				Generator:  simplehttp.GenerateRequestID,
				ContextKey: requestIDContextKey,
			})(c); err != nil {
				return err
			}

			// Get and call the next handler
			if next, ok := c.Locals("nextHandler").(func(*fiber.Ctx) error); ok {
				return next(c)
			}
			return c.Next()
		},
	}
}

// Example of another middleware following the same pattern
func MiddlewareCORS(config *simplehttp.CORSConfig) simplehttp.Middleware {
	fiberConfig := cors.Config{
		AllowOrigins:     defaultCORSAllowOrigins,
		AllowMethods:     defaultCORSAllowMethods,
		AllowHeaders:     defaultCORSAllowHeaders,
		ExposeHeaders:    defaultCORSExposeHeaders,
		AllowCredentials: false,
		MaxAge:           int(defaultCORSMaxAge.Seconds()),
		Next: func(c *fiber.Ctx) bool {
			return false // Process all requests
		},
	}

	if config != nil {
		if config.AllowCredentials {
			if len(config.AllowOrigins) > 0 {
				fiberConfig.AllowOrigins = strings.Join(config.AllowOrigins, ",")
			} else {
				fiberConfig.AllowOrigins = ""
			}
		} else if len(config.AllowOrigins) > 0 {
			fiberConfig.AllowOrigins = strings.Join(config.AllowOrigins, ",")
		}

		if len(config.AllowMethods) > 0 {
			fiberConfig.AllowMethods = strings.Join(config.AllowMethods, ",")
		}
		if len(config.AllowHeaders) > 0 {
			fiberConfig.AllowHeaders = strings.Join(config.AllowHeaders, ",")
		}
		fiberConfig.AllowCredentials = config.AllowCredentials
		fiberConfig.MaxAge = int(config.MaxAge.Seconds())
	}

	return namedMiddleware{
		name:       "CORS",
		middleware: cors.New(fiberConfig),
	}
}

// MiddlewareLogger returns Fiber's logger middleware
func MiddlewareLogger(log simplehttp.Logger) simplehttp.Middleware {
	return namedMiddleware{
		name: "logger",
		middleware: logger.New(logger.Config{
			Format:     defaultLogFormat,
			TimeFormat: defaultLogTimeFormat,
			TimeZone:   defaultTimeZone,
		}),
	}
}

// MiddlewareCompress returns Fiber's compression middleware
func MiddlewareCompress(config simplehttp.CompressionConfig) simplehttp.Middleware {
	level := defaultCompressLevel
	if config.Level != 0 {
		level = config.Level
	}

	return namedMiddleware{
		name: "compress",
		middleware: compress.New(compress.Config{
			Level: compress.Level(level),
		}),
	}
}

// MiddlewareBasicAuth returns Fiber's basic auth middleware
func MiddlewareBasicAuth(username, password string) simplehttp.Middleware {
	return namedMiddleware{
		name: "basic auth",
		middleware: basicauth.New(basicauth.Config{
			Users: map[string]string{
				username: password,
			},
			Realm: defaultAuthRealm,
		}),
	}
}

// MiddlewareRateLimiter returns Fiber's rate limiter middleware
func MiddlewareRateLimiter(config simplehttp.RateLimitConfig) simplehttp.Middleware {
	return namedMiddleware{
		name: "rate limiter",
		middleware: limiter.New(limiter.Config{
			Max:                    config.RequestsPerSecond,
			Expiration:             config.ClientTimeout,
			KeyGenerator:           func(c *fiber.Ctx) string { return c.IP() },
			LimitReached:           func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusTooManyRequests) },
			SkipFailedRequests:     defaultLimitSkipFailed,
			SkipSuccessfulRequests: defaultLimitSkipSucceeded,
		}),
	}
}

// MiddlewareSecurity returns Fiber's security middleware (Helmet)
func MiddlewareSecurity(config simplehttp.SecurityConfig) simplehttp.Middleware {
	xFrameOptions := xFrameOptionsSameOrigin
	if config.FrameDeny {
		xFrameOptions = xFrameOptionsDeny
	}

	xssProtection := xssProtectionDisabled
	if config.BrowserXssFilter {
		xssProtection = xssProtectionEnabled
	}

	nosniff := ""
	if config.ContentTypeNosniff {
		nosniff = contentTypeNosniffValue
	}

	return namedMiddleware{
		name: "security",
		middleware: helmet.New(helmet.Config{
			ContentSecurityPolicy: config.ContentSecurityPolicy,
			XSSProtection:         xssProtection,
			ContentTypeNosniff:    nosniff,
			XFrameOptions:         xFrameOptions,
			ReferrerPolicy:        defaultReferrerPolicy,
			HSTSMaxAge:            defaultHSTSMaxAge,
			HSTSExcludeSubdomains: !config.STSIncludeSubdomains,
			HSTSPreloadEnabled:    defaultHSTSPreload,
		}),
	}
}

// MiddlewareCache returns Fiber's cache middleware
func MiddlewareCache(config simplehttp.CacheConfig) simplehttp.Middleware {
	return namedMiddleware{
		name: "cache",
		middleware: cache.New(cache.Config{
			Expiration: config.TTL,
			KeyGenerator: func(c *fiber.Ctx) string {
				ctx := &FiberContext{ctx: c}
				return config.KeyFunc(ctx)
			},
			Next: func(c *fiber.Ctx) bool {
				for _, header := range config.IgnoreHeaders {
					if c.Get(header) != "" {
						return true
					}
				}
				return false
			},
		}),
	}
}

// MiddlewareRecover returns Fiber's recover middleware
func MiddlewareRecover() simplehttp.Middleware {
	return namedMiddleware{
		name: "recover",
		middleware: recover.New(recover.Config{
			EnableStackTrace: defaultStackTraceEnabled,
		}),
	}
}

// MiddlewareCSRF returns Fiber's CSRF middleware
func MiddlewareCSRF() simplehttp.Middleware {
	return namedMiddleware{
		name: "csrf",
		middleware: csrf.New(csrf.Config{
			KeyLookup:      csrfHeaderLookup,
			CookieName:     csrfCookiePrefix,
			CookieSameSite: csrfCookieSameSite,
			Expiration:     csrfExpiration,
			KeyGenerator:   simplehttp.GenerateRequestID,
		}),
	}
}

// MiddlewareETag returns Fiber's ETag middleware
func MiddlewareETag() simplehttp.Middleware {
	return namedMiddleware{
		name: "etag",
		middleware: etag.New(etag.Config{
			Weak: defaultETagWeak,
		}),
	}
}

// MiddlewareMonitor returns Fiber's monitor middleware
func MiddlewareMonitor() simplehttp.Middleware {
	return namedMiddleware{
		name:       "monitor",
		middleware: monitor.New(),
	}
}
