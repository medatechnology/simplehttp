package simplehttp

import (
	"github.com/medatechnology/goutil/encryption"
)

func GenerateRequestID() string {
	return encryption.NewRandomToken()
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
