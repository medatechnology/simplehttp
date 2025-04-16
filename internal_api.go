package simplehttp

import (
	"net/http"
)

var (
	// Use this so we can change it on by reading from environment
	PathInternalAPI    string = DEFAULT_INTERNAL_API
	PathInternalStatus string = DEFAULT_INTERNAL_STATUS
)

func CreateInternalAPI(s MedaServer) MedaRouter {
	// API routes
	internalAPI := s.Group(PathInternalAPI)
	{
		s.Use(
			MiddlewareHeaderParser(),
		)

		internalAPI.GET(PathInternalStatus, func(c MedaContext) error {
			headers := c.GetHeaders()
			// rid := c.GetHeader(HEADER_REQUEST_ID)
			// fmt.Println("--API - get rid = [", rid, "], from Headers=[", headers.RequestID, "]")
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": "Service OK",
				"headers": map[string]interface{}{
					"RequestID": headers.RequestID,
					"UserAgent": headers.UserAgent,
				},
			})
		})
	}
	return internalAPI
}
