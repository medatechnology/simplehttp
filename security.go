package simplehttp

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Security middleware configuration
type SecurityConfig struct {
	AllowedHosts          []string
	SSLRedirect           bool
	SSLHost               string
	STSSeconds            int64
	STSIncludeSubdomains  bool
	FrameDeny             bool
	ContentTypeNosniff    bool
	BrowserXssFilter      bool
	ContentSecurityPolicy string
}

func MiddlewareSecurity(config SecurityConfig) MedaMiddleware {
	return WithName("basic security", Security(config))
}

// Security returns security middleware
func Security(config SecurityConfig) MedaMiddlewareFunc {
	return func(next MedaHandlerFunc) MedaHandlerFunc {
		return func(c MedaContext) error {
			// fmt.Println("--- security middleware")
			if config.FrameDeny {
				c.Response().Header().Set("X-Frame-Options", "DENY")
			}
			if config.ContentTypeNosniff {
				c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			}
			if config.BrowserXssFilter {
				c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			}
			if config.ContentSecurityPolicy != "" {
				c.Response().Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}
			return next(c)
		}
	}
}

// Rate limit, remember burst is usually the one that taking effects (as maximum)
// Tested OK, it works fine.
// NOTE: make sure the cache middleware is not interfeering, because that can
// effect the rateLimit. When it is returned from cache, it doesn't hit the
// rate limit at all.
type RateLimit struct {
	requestsPerSecond int
	burstSize         int
	store             map[string]*rate.Limiter
	mu                sync.RWMutex
}

// RateLimiter middleware configuration
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
	ClientTimeout     time.Duration
	KeyFunc           func(MedaContext) string // Function to generate rate limit key
}

func MiddlewareRateLimiter(config RateLimitConfig) MedaMiddleware {
	return WithName("rate limiter", RateLimiter(config))
}

// RateLimiter returns a rate limiting middleware
func RateLimiter(config RateLimitConfig) MedaMiddlewareFunc {
	limiter := newRateLimiter(config)
	return func(next MedaHandlerFunc) MedaHandlerFunc {
		return func(c MedaContext) error {
			key := config.KeyFunc(c)
			if err := limiter.Allow(key); err != nil {
				return NewError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}

func newRateLimiter(config RateLimitConfig) *RateLimit {
	return &RateLimit{
		requestsPerSecond: config.RequestsPerSecond,
		burstSize:         config.BurstSize,
		store:             make(map[string]*rate.Limiter),
	}
}

func (rl *RateLimit) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.store[key]

	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.requestsPerSecond), rl.burstSize)
		rl.store[key] = limiter
	}

	return limiter
}

func (rl *RateLimit) Allow(key string) error {
	limiter := rl.getLimiter(key)
	if !limiter.Allow() {
		return ErrRateLimitExceeded
	}
	return nil
}
