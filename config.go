// config.go
package simplehttp

import (
	"fmt"
	"os"
	"time"

	utils "github.com/medatechnology/goutil"
)

const (
	// in seconds, later converted to time.Duration
	DEFAULT_HTTP_READ_TIMEOUT  = 30
	DEFAULT_HTTP_WRITE_TIMEOUT = 30
	DEFAULT_HTTP_IDLE_TIMEOUT  = 60

	// This was used in fiber
	DEFAULT_HTTP_CONCURRENCY = 512 * 1024

	// environment string
	SIMPLEHTTP_FRAMEWORK                 = "SIMPLEHTTP_FRAMEWORK"
	SIMPLEHTTP_PORT                      = "SIMPLEHTTP_PORT"
	SIMPLEHTTP_APP_NAME                  = "SIMPLEHTTP_APP_NAME"
	SIMPLEHTTP_HOST_NAME                 = "SIMPLEHTTP_HOST_NAME"
	SIMPLEHTTP_READ_TIMEOUT              = "SIMPLEHTTP_READ_TIMEOUT"
	SIMPLEHTTP_WRITE_TIMEOUT             = "SIMPLEHTTP_WRITE_TIMEOUT"
	SIMPLEHTTP_IDLE_TIMEOUT              = "SIMPLEHTTP_IDLE_TIMEOUT"
	SIMPLEHTTP_DEBUG                     = "SIMPLEHTTP_DEBUG"
	SIMPLEHTTP_FRAMEWORK_STARTUP_MESSAGE = "SIMPLEHTTP_FRAMEWORK_STARTUP_MESSAGE"
	SIMPLEHTTP_INTERNAL_API              = "SIMPLEHTTP_INTERNAL_API"
	SIMPLEHTTP_INTERNAL_STATUS           = "SIMPLEHTTP_INTERNAL_STATUS"

	// internal API (if enabled)
	DEFAULT_INTERNAL_API    = "/internal_d" // internal debug
	DEFAULT_INTERNAL_STATUS = "/http_status"
)

type TimeOutConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Used to save all endpoints or routes that the server currently handling!
type Routes struct {
	EndPoint string
	Methods  []string
}

func (r *Routes) Sprint() string {
	methods := ""
	for _, m := range r.Methods {
		methods = methods + "," + m
	}
	if len(methods) > 1 {
		methods = methods[1:]
	}
	return r.EndPoint + "\t [" + methods + "]"
}

// Configuration holds server settings
type Config struct {
	Framework string
	AppName   string
	Hostname  string
	Port      string
	// ReadTimeout    time.Duration
	// WriteTimeout   time.Duration
	// IdleTimeout    time.Duration
	MaxHeaderBytes          int
	MaxRequestSize          int64
	UploadDir               string
	TempDir                 string
	TrustedProxies          []string
	Debug                   bool
	FrameworkStartupMessage bool // true means display the default framework startup message, false: quite mode
	Concurrency             int  // for fiber settings

	// TLS Configuration
	TLSCert   string
	TLSKey    string
	AutoTLS   bool
	TLSDomain string

	// Security
	AllowedHosts []string
	SSLRedirect  bool

	// CORS Configuration
	ConfigCORS    *CORSConfig
	ConfigTimeOut *TimeOutConfig
	// TODO: Do we need to add other config like security, limiter, timeout, etc?

	// Custom error handlers
	ErrorHandler func(error, Context) error

	// Additional components
	Logger Logger // Interface defined in logger.go
	// Cache        Cache   // Interface defined in cache.go
	// SessionStore Session // Interface defined in cache.go (session interface)
}

// Default configuration values
var DefaultConfig = &Config{
	Framework: "fiber",
	AppName:   "MedaHTTP",
	Hostname:  "localhost",
	Port:      "8080",
	ConfigTimeOut: &TimeOutConfig{
		ReadTimeout:  time.Second * DEFAULT_HTTP_READ_TIMEOUT,
		WriteTimeout: time.Second * DEFAULT_HTTP_WRITE_TIMEOUT,
		IdleTimeout:  time.Second * DEFAULT_HTTP_IDLE_TIMEOUT,
	},
	MaxHeaderBytes:          1 << 20,  // 1MB
	MaxRequestSize:          32 << 20, // 32MB
	Debug:                   false,
	FrameworkStartupMessage: true,
	Logger:                  NewDefaultLogger(),
	Concurrency:             DEFAULT_HTTP_CONCURRENCY,
	// Cache:          NewMemoryCache(),
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		Framework: utils.GetEnvString(SIMPLEHTTP_FRAMEWORK, DefaultConfig.Framework),
		Port:      utils.GetEnvString(SIMPLEHTTP_PORT, DefaultConfig.Port),
		AppName:   utils.GetEnvString(SIMPLEHTTP_APP_NAME, DefaultConfig.AppName),
		Hostname:  utils.GetEnvString(SIMPLEHTTP_HOST_NAME, DefaultConfig.Hostname),
		ConfigTimeOut: &TimeOutConfig{
			ReadTimeout:  utils.GetEnvDuration(SIMPLEHTTP_READ_TIMEOUT, DefaultConfig.ConfigTimeOut.ReadTimeout),
			WriteTimeout: utils.GetEnvDuration(SIMPLEHTTP_WRITE_TIMEOUT, DefaultConfig.ConfigTimeOut.WriteTimeout),
			IdleTimeout:  utils.GetEnvDuration(SIMPLEHTTP_IDLE_TIMEOUT, DefaultConfig.ConfigTimeOut.IdleTimeout),
		},
		Debug:                   utils.GetEnvBool(SIMPLEHTTP_DEBUG, DefaultConfig.Debug),
		FrameworkStartupMessage: utils.GetEnvBool(SIMPLEHTTP_FRAMEWORK_STARTUP_MESSAGE, DefaultConfig.FrameworkStartupMessage),
		Logger:                  NewDefaultLogger(),
	}
	PathInternalAPI = utils.GetEnvString(SIMPLEHTTP_INTERNAL_API, DEFAULT_INTERNAL_API)
	PathInternalStatus = utils.GetEnvString(SIMPLEHTTP_INTERNAL_STATUS, DEFAULT_INTERNAL_STATUS)
	// Set default components if not provided
	// if config.Logger == nil {
	// 	config.Logger = NewDefaultLogger()
	// }
	// if config.Cache == nil {
	// 	config.Cache = NewMemoryCache()
	// }
	// if config.SessionStore == nil {
	// 	config.SessionStore = NewMemorySession(GenerateRequestID())
	// }

	return config
}

// Configuration validation
func ValidateConfig(config *Config) error {
	if config == nil {
		return ErrInvalidConfig
	}

	// Set defaults if not provided
	if config.Port == "" {
		config.Port = DefaultConfig.Port
	}

	if config.ConfigTimeOut.ReadTimeout == 0 {
		config.ConfigTimeOut.ReadTimeout = DefaultConfig.ConfigTimeOut.ReadTimeout
	}

	if config.ConfigTimeOut.WriteTimeout == 0 {
		config.ConfigTimeOut.WriteTimeout = DefaultConfig.ConfigTimeOut.WriteTimeout
	}

	if config.ConfigTimeOut.IdleTimeout == 0 {
		config.ConfigTimeOut.IdleTimeout = DefaultConfig.ConfigTimeOut.IdleTimeout
	}

	if config.MaxHeaderBytes == 0 {
		config.MaxHeaderBytes = DefaultConfig.MaxHeaderBytes
	}

	if config.MaxRequestSize == 0 {
		config.MaxRequestSize = DefaultConfig.MaxRequestSize
	}

	// Validate file upload directories
	if config.UploadDir != "" {
		if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory: %v", err)
		}
	}

	if config.TempDir != "" {
		if err := os.MkdirAll(config.TempDir, 0755); err != nil {
			return fmt.Errorf("failed to create temp directory: %v", err)
		}
	}

	// Validate TLS configuration
	if config.AutoTLS && config.TLSDomain == "" {
		return fmt.Errorf("TLS domain required when AutoTLS is enabled")
	}

	return nil
}
