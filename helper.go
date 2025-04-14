package simplehttp

import (
	"os"
	"strings"
	"time"

	"github.com/medatechnology/goutil/encryption"
)

func GenerateRequestID() string {
	return encryption.NewRandomToken()
}

func getEnvOrDefault(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// NOTE: already moved to validateBasicAuth using encryption package
// func parseBasicAuth(auth string) (username, password string, ok bool) {
// 	if auth == "" {
// 		return
// 	}

// 	const prefix = "Basic "
// 	if !strings.HasPrefix(auth, prefix) {
// 		return
// 	}

// 	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
// 	if err != nil {
// 		return
// 	}

// 	cs := string(c)
// 	s := strings.IndexByte(cs, ':')
// 	if s < 0 {
// 		return
// 	}

// 	return cs[:s], cs[s+1:], true
// }
