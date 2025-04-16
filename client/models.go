package client

import (
	"net"
	"net/http"
	"time"

	utils "github.com/medatechnology/goutil"
)

// HTTP client configuration constants
const (
	// Default HTTP timeouts
	DEFAULT_TIMEOUT                       = 60 * time.Second
	DEFAULT_DIAL_TIMEOUT                  = 30 * time.Second
	DEFAULT_KEEP_ALIVE                    = 30 * time.Second
	DEFAULT_TLS_HANDSHAKE_TIMEOUT         = 10 * time.Second
	DEFAULT_RESPONSE_HEADER_TIMEOUT       = 60 * time.Second
	DEFAULT_EXPECT_CONTINUE_TIMEOUT       = 5 * time.Second
	DEFAULT_MAX_IDLE_CONNECTIONS          = 100
	DEFAULT_MAX_IDLE_CONNECTIONS_PER_HOST = 100
	DEFAULT_MAX_CONNECTIONS_PER_HOST      = 1000
	DEFAULT_IDLE_CONNECTION_TIMEOUT       = 90 * time.Second
	DEFAULT_MAX_RETRIES                   = 3
	DEFAULT_RETRY_DELAY                   = 1 * time.Second

	// Content type constants
	CONTENT_TYPE_JSON         = "application/json"
	CONTENT_TYPE_FORM         = "application/x-www-form-urlencoded"
	CONTENT_TYPE_MULTIPART    = "multipart/form-data"
	CONTENT_TYPE_TEXT         = "text/plain"
	CONTENT_TYPE_XML          = "application/xml"
	CONTENT_TYPE_OCTET_STREAM = "application/octet-stream"
)

// StatusCode represents an HTTP status code
type StatusCode int

// ClientConfig holds configuration options for the HTTP client
type ClientConfig struct {
	// Basic settings
	BaseURL     string
	Headers     map[string][]string
	QueryParams map[string]string
	ContentType string

	// Authentication
	Username  string
	Password  string
	Token     string
	TokenType string

	// Error handling
	ErrorResult interface{}

	// Timeout settings
	Timeout               time.Duration
	DialTimeout           time.Duration
	KeepAlive             time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
	IdleConnectionTimeout time.Duration

	// Connection settings
	MaxIdleConnections  int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int

	// Retry settings
	MaxRetries  int
	RetryDelay  time.Duration
	RetryPolicy RetryPolicy
}

// RetryPolicy determines if a request should be retried
type RetryPolicy func(resp *http.Response, err error) bool

// ClientOption defines a function that modifies client configuration
type ClientOption func(*ClientConfig)

// Client is the main HTTP client interface
type Client struct {
	Config     ClientConfig
	HTTPClient *http.Client
}

// DefaultRetryPolicy provides a reasonable default retry policy
func DefaultRetryPolicy(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on 5xx server errors
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}

	return false
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig(options ...ClientOption) *ClientConfig {
	config := &ClientConfig{
		Headers:               make(map[string][]string),
		QueryParams:           make(map[string]string),
		ContentType:           CONTENT_TYPE_JSON,
		Timeout:               utils.GetEnvDuration("HTTP_TIMEOUT", DEFAULT_TIMEOUT),
		DialTimeout:           utils.GetEnvDuration("HTTP_DIAL_TIMEOUT", DEFAULT_DIAL_TIMEOUT),
		KeepAlive:             utils.GetEnvDuration("HTTP_KEEP_ALIVE", DEFAULT_KEEP_ALIVE),
		TLSHandshakeTimeout:   utils.GetEnvDuration("HTTP_TLS_TIMEOUT", DEFAULT_TLS_HANDSHAKE_TIMEOUT),
		ResponseHeaderTimeout: utils.GetEnvDuration("HTTP_RESPONSE_TIMEOUT", DEFAULT_RESPONSE_HEADER_TIMEOUT),
		ExpectContinueTimeout: utils.GetEnvDuration("HTTP_CONTINUE_TIMEOUT", DEFAULT_EXPECT_CONTINUE_TIMEOUT),
		IdleConnectionTimeout: utils.GetEnvDuration("HTTP_IDLE_CONN_TIMEOUT", DEFAULT_IDLE_CONNECTION_TIMEOUT),
		MaxIdleConnections:    utils.GetEnvInt("HTTP_MAX_IDLE_CONNS", DEFAULT_MAX_IDLE_CONNECTIONS),
		MaxIdleConnsPerHost:   utils.GetEnvInt("HTTP_MAX_IDLE_CONNS_PER_HOST", DEFAULT_MAX_IDLE_CONNECTIONS_PER_HOST),
		MaxConnsPerHost:       utils.GetEnvInt("HTTP_MAX_CONNS_PER_HOST", DEFAULT_MAX_CONNECTIONS_PER_HOST),
		MaxRetries:            utils.GetEnvInt("HTTP_MAX_RETRIES", DEFAULT_MAX_RETRIES),
		RetryDelay:            utils.GetEnvDuration("HTTP_RETRY_DELAY", DEFAULT_RETRY_DELAY),
		RetryPolicy:           DefaultRetryPolicy,
	}

	// Apply all options
	for _, option := range options {
		option(config)
	}
	return config
}

//-----------------------------------------------------------------------------
// Configuration Options
//-----------------------------------------------------------------------------

// WithBaseURL sets the base URL for the client
func WithBaseURL(url string) ClientOption {
	return func(c *ClientConfig) {
		c.BaseURL = url
	}
}

// WithHeaders sets headers for requests
func WithHeaders(headers map[string][]string) ClientOption {
	return func(c *ClientConfig) {
		for k, v := range headers {
			c.Headers[k] = v
		}
	}
}

// WithHeader adds a single header
func WithHeader(key string, values ...string) ClientOption {
	return func(c *ClientConfig) {
		c.Headers[key] = values
	}
}

// WithQueryParams sets query parameters for requests
func WithQueryParams(params map[string]string) ClientOption {
	return func(c *ClientConfig) {
		for k, v := range params {
			c.QueryParams[k] = v
		}
	}
}

// WithQueryParam adds a single query parameter
func WithQueryParam(key, value string) ClientOption {
	return func(c *ClientConfig) {
		c.QueryParams[key] = value
	}
}

// WithContentType sets the content type for requests
func WithContentType(contentType string) ClientOption {
	return func(c *ClientConfig) {
		c.ContentType = contentType
		c.Headers["Content-Type"] = []string{contentType}
	}
}

// WithJSONContentType sets the content type to application/json
func WithJSONContentType() ClientOption {
	return WithContentType(CONTENT_TYPE_JSON)
}

// WithFormContentType sets the content type to application/x-www-form-urlencoded
func WithFormContentType() ClientOption {
	return WithContentType(CONTENT_TYPE_FORM)
}

// WithMultipartContentType sets the content type to multipart/form-data
func WithMultipartContentType() ClientOption {
	return WithContentType(CONTENT_TYPE_MULTIPART)
}

// WithBasicAuth sets basic authentication for requests
func WithBasicAuth(username, password string) ClientOption {
	return func(c *ClientConfig) {
		c.Username = username
		c.Password = password
	}
}

// WithBearerToken sets bearer token authentication for requests
func WithBearerToken(token string) ClientOption {
	return func(c *ClientConfig) {
		c.Token = token
		c.TokenType = "Bearer"
	}
}

// WithCustomToken sets a custom token authentication for requests
func WithCustomToken(tokenType, token string) ClientOption {
	return func(c *ClientConfig) {
		c.Token = token
		c.TokenType = tokenType
	}
}

// WithErrorResult sets a result object for error responses
func WithErrorResult(result interface{}) ClientOption {
	return func(c *ClientConfig) {
		c.ErrorResult = result
	}
}

// WithTimeout sets the overall request timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = timeout
	}
}

// WithDialTimeout sets the connection dial timeout
func WithDialTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.DialTimeout = timeout
	}
}

// WithKeepAlive sets the keep-alive duration
func WithKeepAlive(keepAlive time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.KeepAlive = keepAlive
	}
}

// WithTLSHandshakeTimeout sets the TLS handshake timeout
func WithTLSHandshakeTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.TLSHandshakeTimeout = timeout
	}
}

// WithResponseHeaderTimeout sets the response header timeout
func WithResponseHeaderTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.ResponseHeaderTimeout = timeout
	}
}

// WithExpectContinueTimeout sets the expect continue timeout
func WithExpectContinueTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.ExpectContinueTimeout = timeout
	}
}

// WithIdleConnectionTimeout sets the idle connection timeout
func WithIdleConnectionTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.IdleConnectionTimeout = timeout
	}
}

// WithMaxIdleConnections sets the maximum number of idle connections
func WithMaxIdleConnections(max int) ClientOption {
	return func(c *ClientConfig) {
		c.MaxIdleConnections = max
	}
}

// WithMaxIdleConnectionsPerHost sets the maximum number of idle connections per host
func WithMaxIdleConnectionsPerHost(max int) ClientOption {
	return func(c *ClientConfig) {
		c.MaxIdleConnsPerHost = max
	}
}

// WithMaxConnectionsPerHost sets the maximum number of connections per host
func WithMaxConnectionsPerHost(max int) ClientOption {
	return func(c *ClientConfig) {
		c.MaxConnsPerHost = max
	}
}

// WithMaxRetries sets the maximum number of retry attempts
func WithMaxRetries(max int) ClientOption {
	return func(c *ClientConfig) {
		c.MaxRetries = max
	}
}

// WithRetryDelay sets the delay between retry attempts
func WithRetryDelay(delay time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.RetryDelay = delay
	}
}

// WithRetryPolicy sets a custom retry policy
func WithRetryPolicy(policy RetryPolicy) ClientOption {
	return func(c *ClientConfig) {
		c.RetryPolicy = policy
	}
}

// NoRetry disables retries
func NoRetry() ClientOption {
	return func(c *ClientConfig) {
		c.MaxRetries = 0
	}
}

// NewHTTPClient creates and configures a new HTTP client
func NewHTTPClient(config *ClientConfig, options ...ClientOption) *http.Client {
	// Use provided config or create a default one
	if config == nil {
		config = NewDefaultConfig(options...)
	}
	// Ensure timeout has a value. NOTE: maybe this is redundant
	timeout := config.Timeout
	if timeout == 0 {
		timeout = DEFAULT_TIMEOUT
	}

	// Create and return a configured HTTP client
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   config.DialTimeout,
				KeepAlive: config.KeepAlive,
			}).Dial,
			TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
			ResponseHeaderTimeout: config.ResponseHeaderTimeout,
			ExpectContinueTimeout: config.ExpectContinueTimeout,
			MaxIdleConns:          config.MaxIdleConnections,
			MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
			MaxConnsPerHost:       config.MaxConnsPerHost,
			IdleConnTimeout:       config.IdleConnectionTimeout,
		},
	}
}

// NewClient creates a new HTTP client with the provided configuration
// Actually this should be better to have it's own simpleClientOptions instead of ClientOption
// which belongs to http.Client options. Then one of them is just WithConfig(NewDefaultConfig(ClientOption))
func NewClient(options ...ClientOption) *Client {
	// Create default config
	config := *NewDefaultConfig(options...)

	// Create HTTP client with the configuration. No need to pass the options here, because
	// already passed via NewDefaultConfig above.
	httpClient := NewHTTPClient(&config)

	// Create and return the client
	return &Client{
		Config:     config,
		HTTPClient: httpClient,
	}
}
