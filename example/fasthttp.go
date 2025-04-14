package main

import (
	"fmt"
	"log"

	utils "github.com/medatechnology/goutil"
	"github.com/medatechnology/simplehttp"
	"github.com/medatechnology/simplehttp/framework/fasthttp"
)

type FastHTTPUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestFastHTTP() {
	// Load configuration
	// Usually is run from the root directory like: go run ./simplehttp/example...
	// so reading the env also like that
	utils.LoadEnv("./simplehttp/example/.env.example")
	config := simplehttp.LoadConfig()

	// Create server
	server := fasthttp.NewServer(config)

	// Add middleware
	server.Use(simplehttp.MiddlewareLogger(simplehttp.NewDefaultLogger()))
	server.Use(simplehttp.MiddlewareRequestID())

	// API routes
	api := server.Group("/api")
	{
		api.Use(simplehttp.MiddlewareTimeout(*config.ConfigTimeOut))

		// Users endpoints
		users := api.Group("/users")
		{
			users.GET("", listUsers)
			users.POST("", createUser)
			users.GET("/:id", getUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}
	}

	// Start server
	if err := server.Start(config.Port); err != nil {
		log.Fatal(err)
	}
}

func listUsers(c simplehttp.MedaContext) error {
	users := []FastHTTPUser{
		{ID: "1", Name: "John"},
		{ID: "2", Name: "Jane"},
	}
	return c.JSON(200, users)
}

func createUser(c simplehttp.MedaContext) error {
	var user FastHTTPUser
	if err := c.BindJSON(&user); err != nil {
		return err
	}
	return c.JSON(201, user)
}

func getUser(c simplehttp.MedaContext) error {
	var user FastHTTPUser
	err := c.Bind(&user)
	if err != nil {
		fmt.Printf("cannot bind parameter")
	}
	user.Name = "John Doe"
	return c.JSON(200, user)
}

func updateUser(c simplehttp.MedaContext) error {
	var user FastHTTPUser
	if err := c.BindJSON(&user); err != nil {
		return err
	}
	return c.JSON(200, user)
}

func deleteUser(c simplehttp.MedaContext) error {
	return c.JSON(204, nil)
}
