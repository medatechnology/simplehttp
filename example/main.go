// examples/echo/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	utils "github.com/medatechnology/goutil"
	"github.com/medatechnology/simplehttp"
	"github.com/medatechnology/simplehttp/framework/fiber"
)

func main() {

	// Test FastHTTP
	// TestFastHTTP()

	// Test Echo Simple
	// SimpleEchoExample()

	// Test Echo Full
	// fmt.Println("Testing Echo Full Usage")
	// ExampleFullUsage()

	// Test SimpleHTTP
	utils.LoadEnv("./example/.env.example")
	config := simplehttp.LoadConfig()
	// var server simplehttp.MedaServer
	// server := echo.NewServer(config)
	server := fiber.NewServer(config)

	TestSimpleHTTP(server, config)
	// TestFullUsage(server, config)

}

func TestSimpleHTTP(server simplehttp.MedaServer, config *simplehttp.Config) {
	// Load configuration
	// Usually is run from the root directory like: go run ./simplehttp/example...
	// so reading the env also like that
	// goutil.LoadEnv("./simplehttp/example/.env.example")
	// config := simplehttp.LoadConfig()
	// var server simplehttp.MedaServer

	// Create server
	// server = echo.NewServer(config)

	// Add middleware
	// server.Use(simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()))
	// server.Use(fiber.MiddlewareRequestID())
	server.Use(
		simplehttp.MiddlewareRequestID(),
		simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
		// fiber.MiddlewareRequestID(),
	)

	// API routes
	api := server.Group("/api")
	{
		api.Use(simplehttp.MiddlewareTimeout(*config.ConfigTimeOut),
			simplehttp.MiddlewareBasicAuth("yudi", "yudiyudi"),
		)

		// Users endpoints
		users := api.Group("/users")
		{
			users.GET("", listUsers)
			users.POST("", createUser)
			users.GET("/:id", getUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}

		api.GET("/status", func(c simplehttp.MedaContext) error {
			headers := c.GetHeaders()
			rid := c.GetHeader(simplehttp.HEADER_REQUEST_ID)
			fmt.Println("--API - get rid = [", rid, "], from Headers=[", headers.RequestID, "]")
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": "Service OK",
				"headers": map[string]interface{}{
					"RequestID": headers.RequestID,
					"UserAgent": headers.UserAgent,
				},
			})
		})
	}

	// Start server
	if err := server.Start(""); err != nil {
		log.Fatal(err)
	}
}

func TestFullUsage(server simplehttp.MedaServer, config *simplehttp.Config) {
	// Test stop watch as well
	// swatch := metrics.StartTimeIt("Loading", 50)

	CORSConfig := &simplehttp.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	}

	secConfig := simplehttp.SecurityConfig{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
	}

	rateConfig := simplehttp.RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
		KeyFunc: func(c simplehttp.MedaContext) string {
			header := c.GetHeaders()
			if header.RealIP != "" {
				return header.RealIP
			}
			// fmt.Println("IP=", header.RealIP, "- ", header.ConnectingIP, "- ", header.RemoteIP, "- ", header.RemoteIP)
			return header.RemoteIP
			// return c.GetHeader("X-Real-IP")
		},
	}

	// cacheConfig := simplehttp.CacheConfig{
	// 	TTL:       time.Minute * 5,
	// 	KeyPrefix: "cache:",
	// 	Store:     simplehttp.NewMemoryCache(),
	// 	KeyFunc: func(c simplehttp.MedaContext) string {
	// 		return c.GetPath() + c.GetHeader("Authorization")
	// 	},
	// }
	// Setup logging
	logger := log.New(os.Stdout, "MEDA", log.LstdFlags)

	server.Use(
		// simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
		simplehttp.MiddlewareRequestID(),
		simplehttp.MiddlewareSecurity(secConfig),
		simplehttp.MiddlewareRateLimiter(rateConfig),
		// simplehttp.Cache(cacheConfig),
		simplehttp.MiddlewareCORS(CORSConfig),
		simplehttp.MiddlewareTimeout(*config.ConfigTimeOut),
		// simplehttp.MiddlewareHeaderParser(),
		simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
	)

	// File handling setup
	fileHandler := simplehttp.NewFileHandler(config.UploadDir)
	fileHandler.MaxFileSize = 50 << 20 // 50MB
	fileHandler.AllowedTypes = []string{
		"image/jpeg",
		"image/png",
		"application/pdf",
	}

	// testing stop watch / timeit
	// time.Sleep(1200 * time.Millisecond) // Simulate loading
	// metrics.StopTimeItPrint(swatch, "Done")
	// fmt.Printf("Manual print == %s\n", l)

	api := server.Group("/api")
	{
		api.GET("/users", func(c simplehttp.MedaContext) error {
			return c.JSON(200, []map[string]string{
				{"id": "1", "name": "John"},
				{"id": "2", "name": "Jane"},
			})
		})

		api.GET("/status", func(c simplehttp.MedaContext) error {
			headers := c.GetHeaders()
			// rid := c.GetHeader(simplehttp.HEADER_REQUEST_ID)
			// fmt.Println("--API - get rid = ", rid)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": "Service OK",
				"headers": map[string]interface{}{
					"RequestID": headers.RequestID,
					"UserAgent": headers.UserAgent,
				},
			})
		})

		api.GET("/header", func(c simplehttp.MedaContext) error {
			headers := c.GetHeaders()
			return c.JSON(http.StatusOK, headers)
		})

		api.POST("/users", func(c simplehttp.MedaContext) error {
			var user struct {
				Name string `json:"name"`
			}
			if err := c.BindJSON(&user); err != nil {
				return err
			}
			return c.JSON(201, user)
		})
	}

	fileAPI := server.Group("/attach")
	// File operations
	fileAPI.POST("/upload", fileHandler.HandleUpload())
	fileAPI.GET("/files/:filename", fileHandler.HandleDownload("./uploads/{{filename}}"))

	secure := server.Group("/auth")
	{
		secure.Use(simplehttp.MiddlewareBasicAuth("yudi", "topsecret"))
		secure.GET("/hello", listUsers)
		secure.GET("/hello/:id", getUser)
	}

	// Websocket chat example
	server.WebSocket("/ws/chat", func(ws simplehttp.MedaWebsocket) error {
		for {
			msg := &Message{}
			if err := ws.ReadJSON(msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Printf("websocket error: %v", err)
				}
				return err
			}

			// Echo the message back
			response := &Message{
				Type: "response",
				Data: msg.Data,
			}

			if err := ws.WriteJSON(response); err != nil {
				logger.Printf("websocket write error: %v", err)
				return err
			}
		}
	})

	// Start server
	go func() {
		if err := server.Start(""); err != nil {
			logger.Printf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal(err)
	}
}
