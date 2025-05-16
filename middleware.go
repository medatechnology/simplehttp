package simplehttp

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/medatechnology/goutil/encryption"
	"github.com/mileusna/useragent"
)

var (
	REQUEST_HEADER_PARSED_STRING string = "request_header"

	HEADER_AUTHORIZATION  string = "authorization"
	HEADER_MEDA_API_KEY   string = "MEDA_API_KEY"
	HEADER_API_KEY        string = "API_KEY"
	HEADER_PRIVATE_TOKEN  string = "PRIVATE_TOKEN"
	HEADER_CODE           string = "code"
	HEADER_CONNECTING_IP  string = "CF-Connecting-IP"
	HEADER_FORWARDED_FOR  string = "X-Forwarded-For"
	HEADER_REAL_IP        string = "X-Real-IP"
	HEADER_TRUE_CLIENT_IP string = "True-Client-IP"
	HEADER_USER_AGENT     string = "User-Agent"
	HEADER_ACCEPT_TYPE    string = "Accept"
	HEADER_TRACE_ID       string = "X-Trace-ID"
	HEADER_REQUEST_ID     string = "X-Request-ID"
	HEADER_ORIGIN         string = "Origin"
)

// NamedMiddleware wraps a middleware with a name for debugging
type NamedMiddleware struct {
	name       string
	middleware MiddlewareFunc
}

// GetMiddlewareName returns the name of the middleware if it's a NamedMiddleware,
// or "unnamed" if it's a regular middleware
//
//	func GetMiddlewareName(m MiddlewareFunc) string {
//		if named, ok := m.(*NamedMiddleware); ok {
//			return named.name
//		}
//		return "unnamed"
//	}
func (m NamedMiddleware) Name() string {
	return m.name
}

// WithName adds a name to a middleware
func WithName(name string, m MiddlewareFunc) NamedMiddleware {
	return NamedMiddleware{
		name:       name,
		middleware: m,
	}
}

// Implement the SimpleHttpMiddleware interface
func (n NamedMiddleware) Handle(next HandlerFunc) HandlerFunc {
	return n.middleware(next)
}

type HeaderAuthorization struct {
	Raw   string `db:"authorization"            json:"authorization,omitempty"`
	Type  string `db:"authorization_type"       json:"authorization_type,omitempty"`
	Token string `db:"authorization_token"      json:"authorization_token,omitempty"`
}

// Please change this if RequestHeader struct is changed
type RequestHeader struct {
	Authorization HeaderAuthorization `db:"header_authorization"             json:"header_authorization,omitempty"`
	// below are specific to some Meda lib, in this case auth-lib
	MedaAPIKey   string `db:"meda_api_key"        json:"MEDA_API_KEY,omitempty"`
	APIKey       string `db:"api_key"             json:"API_KEY,omitempty"`
	PrivateToken string `db:"private_token"       json:"PRIVATE_TOKEN,omitempty"`
	Code         string `db:"code"                json:"code,omitempty"`
	// standard header
	UserAgent         string `db:"user_agent"          json:"User-Agent,omitempty"`
	AcceptType        string `db:"accept_type"         json:"Accept,omitempty"`
	TraceID           string `db:"trace_id"            json:"X-Trace-ID,omitempty"`
	RequestID         string `db:"request_id"          json:"X-Request-ID,omitempty"`
	Origin            string `db:"origin"              json:"Origin,omitempty"`
	ForwardedFor      string `db:"forwarded_for"       json:"X-Forwarded-For,omitempty"`
	RealIP            string `db:"real_ip"             json:"X-Real-IP,omitempty"`
	ConnectingIP      string `db:"connecting_ip"       json:"CF-Connecting-IP,omitempty"`
	TrueIP            string `db:"true_ip"             json:"true-client-ip,omitempty"`
	RemoteIP          string `db:"remote_ip"           json:"remote-address,omitempty"`
	Browser           string `db:"browser"             json:"browser,omitempty"`
	BrowserVersion    string `db:"browser_version"     json:"browser_version,omitempty"`
	PlatformOS        string `db:"platform_os"         json:"platform_os,omitempty"`
	PlatformOSVersion string `db:"platform_os_version" json:"platform_os_version,omitempty"`
	Platform          string `db:"platform"            json:"platform,omitempty"` // mobile, desktop, unknown
	Device            string `db:"device"              json:"device,omitempty"`   // usually if mobile, this one has value
}

func (mh *RequestHeader) FromHttpRequest(stdRequest *http.Request) {
	// auth header
	// fmt.Println("fromHTTPHeader Request header = ", stdRequest.Header)
	mh.Authorization.Raw = stdRequest.Header.Get(HEADER_AUTHORIZATION)
	// NOTE: the function will return string "empty" if .Raw is null
	mh.Authorization.Type, mh.Authorization.Token = encryption.GetAuthorizationFromHeader(mh.Authorization.Raw)
	// if mh.Authorization.Raw != "" {
	// 	parts := strings.Split(mh.Authorization.Raw, " ")
	// 	if len(parts) == 2 {
	// 		mh.Authorization.Type = parts[0]
	// 		mh.Authorization.Token = parts[1]
	//  } // else we cannot parse!
	// }

	// Meda specific lib, auth-lib
	mh.MedaAPIKey = stdRequest.Header.Get(HEADER_MEDA_API_KEY)
	mh.APIKey = stdRequest.Header.Get(HEADER_API_KEY)
	mh.PrivateToken = stdRequest.Header.Get(HEADER_PRIVATE_TOKEN)
	mh.Code = stdRequest.Header.Get(HEADER_CODE)

	// standard header
	mh.UserAgent = stdRequest.Header.Get(HEADER_USER_AGENT)
	mh.AcceptType = stdRequest.Header.Get(HEADER_ACCEPT_TYPE)
	mh.TraceID = stdRequest.Header.Get(HEADER_TRACE_ID)
	mh.RequestID = stdRequest.Header.Get(HEADER_REQUEST_ID)
	mh.Origin = stdRequest.Header.Get(HEADER_ORIGIN)
	mh.ForwardedFor = stdRequest.Header.Get(HEADER_FORWARDED_FOR)
	mh.RealIP = stdRequest.Header.Get(HEADER_REAL_IP)
	mh.ConnectingIP = stdRequest.Header.Get(HEADER_CONNECTING_IP)
	mh.TrueIP = stdRequest.Header.Get(HEADER_TRUE_CLIENT_IP)
	mh.RemoteIP = stdRequest.RemoteAddr
	// mh.RemoteIP, _, _ = net.SplitHostPort(mh.RemoteIP) // is this necessary to split?
	agent := useragent.Parse(stdRequest.UserAgent())
	// agent := useragent.Parse(stdRequest.Header.Get("User-Agent"))
	mh.Device = agent.Device
	mh.Browser = agent.Name
	mh.BrowserVersion = agent.VersionNoFull()
	mh.PlatformOS = agent.OS
	mh.PlatformOSVersion = agent.OSVersion
	if agent.Mobile {
		mh.Platform = "mobile"
	}
	if agent.Tablet {
		mh.Platform = "tablet"
	}
	if agent.Desktop {
		mh.Platform = "desktop"
	}
	if agent.Bot {
		mh.Platform = "bot"
	}
}

func (mh *RequestHeader) IP() string {
	// Are these IP variables more accurate? if not just return remoteIP which usually
	// are not empty.
	if mh.ConnectingIP != "" {
		return mh.ConnectingIP
	} else if mh.RealIP != "" {
		return mh.RealIP
	} else if mh.TrueIP != "" {
		return mh.TrueIP
	}
	return mh.RemoteIP
}

func MiddlewareHeaderParser() Middleware {
	return WithName("header parser", HeaderParser())
}

// Standard middleware implementations
func HeaderParser() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			// fmt.Println("--- header middleware")
			// var header RequestHeader
			// header.FromHttpHeader(c.Request())
			header := c.GetHeaders()

			// Parse Authorization header
			// if auth := c.GetHeader("Authorization"); auth != "" {
			// 	parts := strings.Split(auth, " ")
			// 	if len(parts) == 2 {
			// 		header.Authorization.Type = parts[0]
			// 		header.Authorization.Token = parts[1]
			// 	}
			// }

			c.Set(REQUEST_HEADER_PARSED_STRING, header)
			// c.Set("headers", header)
			return next(c)
		}
	}
}

// CORSConfig defines CORS settings
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func MiddlewareCORS(config *CORSConfig) Middleware {
	return WithName("CORS bypass", CORS(config))
}

// CORS middleware returns a Middleware that adds CORS headers to the response
func CORS(config *CORSConfig) MiddlewareFunc {
	// Use default config if nil
	if config == nil {
		config = &CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "HEAD", "PUT", "POST", "DELETE", "PATCH"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
			ExposeHeaders:    []string{},
			AllowCredentials: false,
			MaxAge:           24 * time.Hour,
		}
	}

	// Convert methods and headers to uppercase
	for i, m := range config.AllowMethods {
		config.AllowMethods[i] = strings.ToUpper(m)
	}
	for i, h := range config.AllowHeaders {
		config.AllowHeaders[i] = http.CanonicalHeaderKey(h)
	}

	// IMPORTANT: Remember to use context SetResponseHeader or SetRequestHeader!!!
	//            doing c.Response().Header().Set() -- doesn't work, it will set a copy of header and won't persist
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			// fmt.Println("--- cors middleware")
			req := c.Request()
			// res := c.Response()

			// Set CORS headers
			c.SetResponseHeader("Access-Control-Allow-Origin", getAllowedOrigin(config.AllowOrigins, req.Header.Get("Origin")))
			// res.Header().Set("Access-Control-Allow-Origin", getAllowedOrigin(config.AllowOrigins, req.Header.Get("Origin")))

			if config.AllowCredentials {
				c.SetResponseHeader("Access-Control-Allow-Credentials", "true")
				// res.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if req.Method == http.MethodOptions {
				// Handle preflight request
				c.SetResponseHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ","))
				c.SetResponseHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))
				// res.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ","))
				// res.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))

				if len(config.ExposeHeaders) > 0 {
					c.SetResponseHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ","))
					// res.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ","))
				}

				if config.MaxAge > 0 {
					c.SetResponseHeader("Access-Control-Max-Age", strconv.FormatInt(int64(config.MaxAge.Seconds()), 10))
					// res.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(config.MaxAge.Seconds()), 10))
				}

				return c.String(http.StatusNoContent, "")
			}

			return next(c)
		}
	}
}

// Helper function for CORS
func getAllowedOrigin(allowedOrigins []string, origin string) string {
	if len(allowedOrigins) == 0 {
		return "*"
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return origin
		}
	}

	return allowedOrigins[0]
}

// Compression middleware configuration
type CompressionConfig struct {
	Level   int      // Compression level (1-9)
	MinSize int64    // Minimum size to compress
	Types   []string // Content types to compress
}

func MiddlewareCompress(config CompressionConfig) Middleware {
	return WithName("compression", Compress(config))
}

// Compress returns a compression middleware
func Compress(config CompressionConfig) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			// Implementation details for compression
			return next(c)
		}
	}
}

func MiddlewareBasicAuth(username, password string) Middleware {
	return WithName("basic auth", BasicAuth(username, password))
}

func BasicAuth(username, password string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			// fmt.Println("--- auth middleware")

			auth := c.GetHeader("Authorization")
			if !validateBasicAuth(auth, username, password) {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "unauthorized",
				})
			}
			return next(c)
		}
	}
}

func validateBasicAuth(auth, username, password string) bool {
	// TODO: implement simple basic auth
	var authHeader HeaderAuthorization
	authHeader.Raw = auth
	authHeader.Type, authHeader.Token = encryption.GetAuthorizationFromHeader(authHeader.Raw)

	if auth == "" || authHeader.Type != "Basic" {
		return false
	}
	authUser, authPass, err := encryption.GetClientIDSecretFromTokenString(authHeader.Token)
	// authUser, authPass, ok := parseBasicAuth(auth)
	if err == nil && authUser == username && authPass == password {
		return true
	} else {
		return false
	}
}

func MiddlewareRequestID() Middleware {
	return WithName("request ID", RequestID())
}

// RequestID middleware adds a unique ID to each request
func RequestID() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			// fmt.Println("--- requestID middleware")

			rid := c.GetHeader(HEADER_REQUEST_ID)
			if rid == "" {
				rid = GenerateRequestID()
				// fmt.Println("request id generated = ", rid)
				c.SetHeader(HEADER_REQUEST_ID, rid)
				// c.Response().Header().Set(HEADER_REQUEST_ID, rid)
				// c.Request().Header.Set(HEADER_REQUEST_ID, rid)

			}
			// Test if requestID is actually there.
			// testRID := c.GetHeader(HEADER_REQUEST_ID)
			// testGet := c.Request().Header.Get(HEADER_REQUEST_ID)
			// fmt.Println("======== TEST ID = [", testRID, "] ========== ", testGet)
			return next(c)
		}
	}
}

func MiddlewareTimeout(config TimeOutConfig) Middleware {
	return WithName("timeout middleware", Timeout(config))
}

// Timeout middleware adds a timeout to the request context
func Timeout(config TimeOutConfig) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			// fmt.Println("--- timeout middleware")

			ctx, cancel := context.WithTimeout(c.Context(), config.ReadTimeout)
			defer cancel()

			c.SetContext(ctx)

			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return NewError(http.StatusGatewayTimeout, "request timeout")
			}
		}
	}
}

// RecoverConfig holds configuration for the Recover middleware
type RecoverConfig struct {
	// StackTrace determines whether to include stack traces in error responses
	StackTrace bool
	// LogStackTrace determines whether to log stack traces
	LogStackTrace bool
	// ErrorHandler is a custom handler for recovered panics
	ErrorHandler func(c Context, err interface{}, stack []byte) error
	// Logger for recording panic information
	Logger Logger
}

func MiddlewareRecover(config ...RecoverConfig) Middleware {
	return WithName("recover", Recover(config...))
}

// Recover returns a middleware that recovers from panics
func Recover(config ...RecoverConfig) MiddlewareFunc {
	// Setup config
	cfg := RecoverConfig{
		// DefaultRecoverConfig provides sensible defaults for Recover middleware
		StackTrace:    false, // Don't expose stack traces in production by default
		LogStackTrace: true,  // But do log them
		Logger:        nil,   // Will use DefaultLogger if not provided
	}
	if len(config) > 0 {
		cfg = config[0]
	}

	// Use default logger if none provided
	if cfg.Logger == nil {
		cfg.Logger = NewDefaultLogger()
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			defer func() {
				if r := recover(); r != nil {
					var err error
					stack := debug.Stack()

					// Log the panic
					cfg.Logger.Errorf("[PANIC RECOVERED] %v\n%s", r, string(stack))

					// Use custom error handler if provided
					if cfg.ErrorHandler != nil {
						err = cfg.ErrorHandler(c, r, stack)
					} else {
						// Default error response
						errMsg := fmt.Sprintf("Internal Server Error: %v", r)
						errResp := map[string]interface{}{
							"error": errMsg,
						}

						// Include stack trace if configured
						if cfg.StackTrace {
							errResp["stack"] = string(stack)
						}

						err = c.JSON(http.StatusInternalServerError, errResp)
					}

					// If response couldn't be sent, log it
					if err != nil {
						cfg.Logger.Errorf("Failed to send error response: %v", err)
					}
				}
			}()

			return next(c)
		}
	}
}
