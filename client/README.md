# Universal HTTP Client for Go

A lightweight, flexible HTTP client with sensible defaults and powerful customization options.

## Features

- **Simple API**: Clean, intuitive methods for common HTTP operations
- **Type-Safe Responses**: Generic functions for decoding responses to any struct type
- **Flexible Configuration**: Extensive options for timeouts, retries, and connections
- **Authentication Support**: Built-in methods for Basic Auth, Bearer tokens, and custom authentication
- **Automatic Retries**: Configurable retry policies with exponential backoff
- **Environment-Aware**: Reads configuration from environment variables when available
- **Fluent Interface**: Chainable configuration methods

## Installation

```bash
go get github.com/medatechnology/goutil/httpclient
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/medatechnology/goutil/httpclient"
)

func main() {
    // Create a new client with default configuration
    client := httpclient.NewClient()
    
    // Make a GET request and decode to a map
    response, err := client.Get("https://api.example.com/users")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Response: %v\n", response)
}
```

### Configuring the Client
It is best practice to just use one (1) instance of client. Because inside basically it's just http.Client and http package have channels for Client, which can make many multiple API calls from single instance. If you create multiple simplehttp.Client the overhead will be high and less efficient.  If you are planning to call large number of http calls, then better to setup pools of simplehttp.Client, but even 10 instances already can handle near 100K calls.
```go
// Create a client with custom configuration
client := httpclient.NewClient(
    httpclient.WithBaseURL("https://api.example.com"),
    httpclient.WithTimeout(30 * time.Second),
    httpclient.WithMaxRetries(5),
    httpclient.WithBearerToken("your-token-here"),
    httpclient.WithHeader("X-API-Key", "your-api-key"),
)
```

### POST Request with JSON Body

```go
// Create a POST request with a JSON body
data := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
}

response, err := client.Post("/users", data)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Created user with ID: %v\n", response["id"])
```

### Decoding to a Struct

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Get a user and decode directly to struct
user, err := httpclient.GetAs[User](client, "/users/42")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
```

### Handling Form Data

```go
// Send form data
formData := map[string]string{
    "username": "johndoe",
    "password": "secret",
}

response, err := client.Post("/login", formData, httpclient.WithFormContentType())
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Auth token: %v\n", response["token"])
```

## Advanced Usage

### Custom Retry Policy

```go
// Create a custom retry policy
customRetryPolicy := func(resp *http.Response, err error) bool {
    // Retry on network errors
    if err != nil {
        return true
    }
    
    // Retry on 429 Too Many Requests
    if resp != nil && resp.StatusCode == 429 {
        return true
    }
    
    // Retry on 5xx server errors
    if resp != nil && resp.StatusCode >= 500 {
        return true
    }
    
    return false
}

// Use the custom retry policy
client := httpclient.NewClient(
    httpclient.WithRetryPolicy(customRetryPolicy),
    httpclient.WithMaxRetries(3),
    httpclient.WithRetryDelay(2 * time.Second),
)
```

### Handling Raw Responses

```go
// Get raw bytes
bytes, err := client.GetBytes("/files/document.pdf")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

// Save to file
ioutil.WriteFile("document.pdf", bytes, 0644)
```

### Batch Requests

```go
// Define request functions
requests := []func() (interface{}, error){
    func() (interface{}, error) {
        return httpclient.GetAs[User](client, "/users/1")
    },
    func() (interface{}, error) {
        return httpclient.GetAs[User](client, "/users/2")
    },
    func() (interface{}, error) {
        return httpclient.GetAs[User](client, "/users/3")
    },
}

// Execute requests concurrently
results := make([]interface{}, len(requests))
errors := make([]error, len(requests))
var wg sync.WaitGroup

for i, request := range requests {
    wg.Add(1)
    go func(i int, req func() (interface{}, error)) {
        defer wg.Add(-1)
        results[i], errors[i] = req()
    }(i, request)
}

wg.Wait()

// Process results
for i, result := range results {
    if errors[i] != nil {
        fmt.Printf("Request %d failed: %v\n", i, errors[i])
        continue
    }
    
    user := result.(User)
    fmt.Printf("User %d: %s\n", i, user.Name)
}
```

### Custom Authentication

```go
// Create client with API key authentication
client := httpclient.NewClient(
    httpclient.WithHeader("X-API-Key", "your-api-key"),
)

// Make authenticated request
response, err := client.Get("/protected-resource")
```

## Configuration Options

### Client Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithBaseURL` | Sets the base URL for all requests | "" |
| `WithTimeout` | Sets the overall request timeout | 60s |
| `WithHeader` | Adds a single header to all requests | - |
| `WithHeaders` | Sets multiple headers at once | - |
| `WithQueryParam` | Adds a query parameter to all requests | - |
| `WithQueryParams` | Sets multiple query parameters | - |
| `WithBasicAuth` | Sets basic authentication credentials | - |
| `WithBearerToken` | Sets bearer token authentication | - |
| `WithContentType` | Sets the default content type | application/json |
| `WithMaxRetries` | Sets the number of retry attempts | 3 |
| `WithRetryDelay` | Sets the delay between retries | 1s |
| `WithRetryPolicy` | Sets a custom retry policy | DefaultRetryPolicy |

### Connection Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithDialTimeout` | Sets the connection dial timeout | 30s |
| `WithKeepAlive` | Sets the keep-alive duration | 30s |
| `WithTLSHandshakeTimeout` | Sets the TLS handshake timeout | 10s |
| `WithResponseHeaderTimeout` | Sets the response header timeout | 60s |
| `WithExpectContinueTimeout` | Sets the expect-continue timeout | 5s |
| `WithIdleConnectionTimeout` | Sets the idle connection timeout | 90s |
| `WithMaxIdleConnections` | Sets the maximum idle connections | 100 |
| `WithMaxIdleConnectionsPerHost` | Sets the maximum idle connections per host | 100 |
| `WithMaxConnectionsPerHost` | Sets the maximum connections per host | 1000 |

## Environment Variables

The client can be configured using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `HTTP_TIMEOUT` | Overall request timeout | 60s |
| `HTTP_DIAL_TIMEOUT` | Connection dial timeout | 30s |
| `HTTP_KEEP_ALIVE` | Keep-alive duration | 30s |
| `HTTP_TLS_TIMEOUT` | TLS handshake timeout | 10s |
| `HTTP_RESPONSE_TIMEOUT` | Response header timeout | 60s |
| `HTTP_CONTINUE_TIMEOUT` | Expect-continue timeout | 5s |
| `HTTP_IDLE_CONN_TIMEOUT` | Idle connection timeout | 90s |
| `HTTP_MAX_IDLE_CONNS` | Maximum idle connections | 100 |
| `HTTP_MAX_IDLE_CONNS_PER_HOST` | Maximum idle connections per host | 100 |
| `HTTP_MAX_CONNS_PER_HOST` | Maximum connections per host | 1000 |
| `HTTP_MAX_RETRIES` | Maximum retry attempts | 3 |
| `HTTP_RETRY_DELAY` | Delay between retries | 1s |
| `HTTP_BASE_URL` | Base URL for all requests | "" |

## Example Response Handling

### Map Response

```go
response, err := client.Get("/users")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

// Example response:
// {
//   "users": [
//     {"id": 1, "name": "John"},
//     {"id": 2, "name": "Jane"}
//   ],
//   "total": 2
// }

users := response["users"].([]interface{})
for _, user := range users {
    userData := user.(map[string]interface{})
    fmt.Printf("User %v: %s\n", userData["id"], userData["name"])
}
```

### Struct Response

```go
type UserResponse struct {
    Users []User `json:"users"`
    Total int    `json:"total"`
}

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

response, err := httpclient.GetAs[UserResponse](client, "/users")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

// Example response structure:
// UserResponse{
//   Users: []User{
//     {ID: 1, Name: "John"},
//     {ID: 2, Name: "Jane"},
//   },
//   Total: 2,
// }

for _, user := range response.Users {
    fmt.Printf("User %d: %s\n", user.ID, user.Name)
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.