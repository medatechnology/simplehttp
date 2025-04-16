package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/medatechnology/simplehttp"
	"github.com/medatechnology/simplehttp/framework/echo"
)

type Message struct {
	Type string
	Data interface{}
}

func SimpleEchoExample() {
	config := simplehttp.LoadConfig()
	server := echo.NewServer(config)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Use middleware and routes
	server.Use(simplehttp.MiddlewareLogger(config.Logger))
	server.GET("/", func(c simplehttp.MedaContext) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Start the server
	server.Start(":8080")
}

// Example usage with all features
func ExampleFullUsage() {
	config := &simplehttp.Config{
		Framework: "echo",
		Port:      "8080",
		ConfigTimeOut: &simplehttp.TimeOutConfig{
			ReadTimeout:  time.Second * 30,
			WriteTimeout: time.Second * 30,
			IdleTimeout:  time.Second * 30,
		},
		UploadDir: "./uploads",
		ConfigCORS: &simplehttp.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           24 * time.Hour,
		},
	}

	server := echo.NewServer(config)

	// Setup logging
	logger := log.New(os.Stdout, "[MEDA] ", log.LstdFlags)

	// Add standard middleware
	server.Use(
		simplehttp.MiddlewareHeaderParser(),
		simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
		simplehttp.MiddlewareBasicAuth("admin", "secret"), // Optional
	)

	// File handling setup
	fileHandler := simplehttp.NewFileHandler(config.UploadDir)
	fileHandler.MaxFileSize = 50 << 20 // 50MB
	fileHandler.AllowedTypes = []string{
		"image/jpeg",
		"image/png",
		"application/pdf",
	}

	// API routes
	api := server.Group("/api")

	// File operations
	api.POST("/upload", fileHandler.HandleUpload())
	api.GET("/files/:filename", fileHandler.HandleDownload("./uploads/{{filename}}"))

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
		if err := server.Start(":" + config.Port); err != nil {
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

func SimpleEchoMain() {
	// Example 1: Basic usage without middleware
	config := &simplehttp.Config{
		Port: "8080",
	}

	server := echo.NewServer(config)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	server.GET("/hello", func(c simplehttp.MedaContext) error {
		return c.JSON(200, map[string]string{"message": "Hello World!"})
	})

	// Start server
	server.Start(":8080")

	// Example 2: Using middleware
	configWithMiddleware := &simplehttp.Config{
		Port: "8081",
		// Logger: simplehttp.NewDefaultLogger(),
		// CORSConfig: &simplehttp.CORSConfig{
		// 	AllowOrigins: []string{"*"},
		// 	AllowMethods: []string{"GET", "POST"},
		// 	MaxAge:       time.Hour * 24,
		// },
	}

	serverWithMiddleware := echo.NewServer(configWithMiddleware)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Add middleware
	serverWithMiddleware.Use(
		simplehttp.MiddlewareHeaderParser(),
		simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),
		simplehttp.MiddlewareSecurity(simplehttp.SecurityConfig{
			FrameDeny:          true,
			ContentTypeNosniff: true,
			BrowserXssFilter:   true,
		}),
		simplehttp.MiddlewareRateLimiter(simplehttp.RateLimitConfig{
			RequestsPerSecond: 10,
			BurstSize:         20,
			KeyFunc: func(c simplehttp.MedaContext) string {
				return c.GetHeader("X-Real-IP")
			},
		}),
		simplehttp.MiddlewareCache(simplehttp.CacheConfig{
			TTL:       time.Minute * 5,
			KeyPrefix: "cache:",
			Store:     simplehttp.NewMemoryCache(),
			KeyFunc: func(c simplehttp.MedaContext) string {
				return c.GetPath() + c.GetHeader("Authorization")
			},
		}),
	)

	// API routes with middleware applied
	api := serverWithMiddleware.Group("/api")
	{
		api.GET("/users", func(c simplehttp.MedaContext) error {
			return c.JSON(200, []map[string]string{
				{"id": "1", "name": "John"},
				{"id": "2", "name": "Jane"},
			})
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

	// Start server with middleware
	serverWithMiddleware.Start(":8081")

}
