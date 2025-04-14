## Usage example



```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/medatechnology/simplehttp"
	"github.com/medatechnology/simplehttp/framework/fiber"
)

// Define a simple User type for demonstration purposes
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Handler functions
func getUsers(c simplehttp.MedaContext) error {
	users := []User{
		{ID: "1", Name: "John Doe"},
		{ID: "2", Name: "Jane Smith"},
	}
	return c.JSON(http.StatusOK, users)
}

func getUserByID(c simplehttp.MedaContext) error {
	id := c.GetQueryParam("id")
	if id == "" {
		return simplehttp.NewError(http.StatusBadRequest, "Missing ID parameter")
	}

	// In a real app, you would fetch the user from a database
	user := User{ID: id, Name: "Example User"}
	return c.JSON(http.StatusOK, user)
}

func createUser(c simplehttp.MedaContext) error {
	var user User
	if err := c.BindJSON(&user); err != nil {
		return simplehttp.NewError(http.StatusBadRequest, "Invalid user data")
	}

	// In a real app, you would save the user to a database
	return c.JSON(http.StatusCreated, user)
}

func handleWebSocket(ws simplehttp.MedaWebsocket) error {
	// Simple chat message type
	type Message struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	for {
		// Read message from client
		msg := &Message{}
		if err := ws.ReadJSON(msg); err != nil {
			return err
		}

		// Echo the message back with a timestamp
		reply := &Message{
			Type: "reply",
			Text: "Echo: " + msg.Text + " (at " + time.Now().Format(time.RFC3339) + ")",
		}

		if err := ws.WriteJSON(reply); err != nil {
			return err
		}
	}
}

func main() {
	// Load configuration
	config := simplehttp.LoadConfig()
	config.Debug = true // Enable debug mode for verbose output

	// Create a server with the Fiber implementation
	server := fiber.NewServer(config)

	// Add global middleware
	server.Use(
		simplehttp.MiddlewareRequestID(),                             // Add unique request IDs
		simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()),   // Log all requests
		simplehttp.MiddlewareTimeout(*config.ConfigTimeOut),          // Apply timeouts
		simplehttp.MiddlewareHeaderParser(),                          // Parse headers
	)

	// Security configuration
	securityConfig := simplehttp.SecurityConfig{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
	}

	// Rate limiting configuration
	rateLimitConfig := simplehttp.RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
		KeyFunc: func(c simplehttp.MedaContext) string {
			headers := c.GetHeaders()
			return headers.RemoteIP // Rate limit by IP
		},
	}

	// CORS configuration
	corsConfig := &simplehttp.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	}

	// Create API routes
	api := server.Group("/api")
	{
		// Add route-specific middleware
		api.Use(
			simplehttp.MiddlewareSecurity(securityConfig),
			simplehttp.MiddlewareRateLimiter(rateLimitConfig),
			simplehttp.MiddlewareCORS(corsConfig),
		)

		// User endpoints
		users := api.Group("/users")
		{
			users.GET("", getUsers)
			users.GET("/:id", getUserByID)
			users.POST("", createUser)
		}

		// Status endpoint
		api.GET("/status", func(c simplehttp.MedaContext) error {
			headers := c.GetHeaders()
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "OK",
				"time":    time.Now().Format(time.RFC3339),
				"request": headers.RequestID,
			})
		})
	}

	// File handling example
	fileHandler := simplehttp.NewFileHandler("./uploads")
	fileHandler.MaxFileSize = 10 << 20 // 10MB
	fileHandler.AllowedTypes = []string{
		"image/jpeg",
		"image/png",
		"application/pdf",
	}

	files := server.Group("/files")
	{
		files.POST("/upload", fileHandler.HandleUpload())
		files.GET("/download/:filename", fileHandler.HandleDownload("./uploads/{{filename}}"))
	}

	// WebSocket example
	server.WebSocket("/ws/chat", handleWebSocket)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s:%s", config.Hostname, config.Port)
		if err := server.Start(""); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Gracefully shutdown with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server shutdown complete")
}
```