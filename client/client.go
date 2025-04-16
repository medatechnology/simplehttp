package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Get performs an HTTP GET request and returns the result as a JSON map
func (c *Client) Get(endpoint string, options ...ClientOption) (map[string]interface{}, error) {
	return RequestAs[map[string]interface{}](c, "GET", endpoint, nil, options...)
}

// Post performs an HTTP POST request and returns the result as a JSON map
func (c *Client) Post(endpoint string, body interface{}, options ...ClientOption) (map[string]interface{}, error) {
	return RequestAs[map[string]interface{}](c, "POST", endpoint, body, options...)
}

// Put performs an HTTP PUT request and returns the result as a JSON map
func (c *Client) Put(endpoint string, body interface{}, options ...ClientOption) (map[string]interface{}, error) {
	return RequestAs[map[string]interface{}](c, "PUT", endpoint, body, options...)
}

// Delete performs an HTTP DELETE request and returns the result as a JSON map
func (c *Client) Delete(endpoint string, options ...ClientOption) (map[string]interface{}, error) {
	return RequestAs[map[string]interface{}](c, "DELETE", endpoint, nil, options...)
}

// Patch performs an HTTP PATCH request and returns the result as a JSON map
func (c *Client) Patch(endpoint string, body interface{}, options ...ClientOption) (map[string]interface{}, error) {
	return RequestAs[map[string]interface{}](c, "PATCH", endpoint, body, options...)
}

// GetAs performs an HTTP GET request and decodes the response to the specified type
func GetAs[T any](c *Client, endpoint string, options ...ClientOption) (T, error) {
	return RequestAs[T](c, "GET", endpoint, nil, options...)
}

// PostAs performs an HTTP POST request and decodes the response to the specified type
func PostAs[T any](c *Client, endpoint string, body interface{}, options ...ClientOption) (T, error) {
	return RequestAs[T](c, "POST", endpoint, body, options...)
}

// PutAs performs an HTTP PUT request and decodes the response to the specified type
func PutAs[T any](c *Client, endpoint string, body interface{}, options ...ClientOption) (T, error) {
	return RequestAs[T](c, "PUT", endpoint, body, options...)
}

// DeleteAs performs an HTTP DELETE request and decodes the response to the specified type
func DeleteAs[T any](c *Client, endpoint string, options ...ClientOption) (T, error) {
	return RequestAs[T](c, "DELETE", endpoint, nil, options...)
}

// PatchAs performs an HTTP PATCH request and decodes the response to the specified type
func PatchAs[T any](c *Client, endpoint string, body interface{}, options ...ClientOption) (T, error) {
	return RequestAs[T](c, "PATCH", endpoint, body, options...)
}

// RequestAs performs an HTTP request and decodes the response to the specified type
func RequestAs[T any](c *Client, method, endpoint string, body interface{}, options ...ClientOption) (T, error) {
	var result T

	resp, err := c.Request(method, endpoint, body, options...)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	return DecodeResponse[T](resp)
}

// Request performs an HTTP request and returns the raw response
func (c *Client) Request(method, endpoint string, body interface{}, options ...ClientOption) (*http.Response, error) {
	// Create a copy of the client's configuration
	reqConfig := c.Config

	// Apply request-specific options
	for _, option := range options {
		option(&reqConfig)
	}

	// Build the full URL from base URL and endpoint
	fullURL := buildURL(reqConfig.BaseURL, endpoint, reqConfig.QueryParams)

	// Prepare the request body once
	bodyData, contentType, err := prepareRequestBody(body, reqConfig.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request body: %w", err)
	}

	// Use the content type from body preparation if not explicitly set
	if contentType != "" && reqConfig.ContentType == "" {
		reqConfig.ContentType = contentType
	}

	// Create a new request with a fresh body reader
	var bodyReader io.Reader
	if bodyData != nil {
		bodyReader = bytes.NewReader(bodyData)
	}

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt < reqConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(reqConfig.RetryDelay)
		}

		req, err := http.NewRequest(method, fullURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		for key, values := range reqConfig.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		// Set content type if needed
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", reqConfig.ContentType)
		}

		// Apply authentication
		applyAuth(req, &reqConfig)

		// Execute the request
		resp, err = c.HTTPClient.Do(req)
		lastErr = err

		// Check if the request was successful
		if err == nil {
			// Check if we should retry based on response
			if reqConfig.RetryPolicy != nil && reqConfig.RetryPolicy(resp, nil) {
				resp.Body.Close()
				continue
			}
			// No need to retry
			break
		}

		// Check if we should retry, if no retrypolicy then we also do not retry!
		if attempt >= reqConfig.MaxRetries || reqConfig.RetryPolicy == nil || !reqConfig.RetryPolicy(nil, err) {
			return nil, fmt.Errorf("request failed: %w(%d)", err, attempt)
		}
		// Will retry
	}

	if resp == nil {
		return nil, fmt.Errorf("all request attempts failed: %w", lastErr)
	}

	// Check for error status codes
	// if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	// 	errorBody, _ := io.ReadAll(resp.Body)
	// 	resp.Body.Close()
	// 	return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(errorBody))
	// }

	return resp, nil
}

// DecodeResponse is a generic function to decode an HTTP response into the specified type
func DecodeResponse[T any](resp *http.Response) (T, error) {
	var result T

	if resp == nil {
		return result, fmt.Errorf("nil response")
	}

	// Handle byte slice result
	if _, ok := any(&result).(*[]byte); ok {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, fmt.Errorf("failed to read response body: %w", err)
		}

		// This is a bit of a hack but works for the byte slice case
		*any(&result).(*[]byte) = data
		return result, nil
	}

	// Handle string result
	if _, ok := any(&result).(*string); ok {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, fmt.Errorf("failed to read response body: %w", err)
		}

		// Another hack for string pointers
		*any(&result).(*string) = string(data)
		return result, nil
	}

	// For all other types, try to decode as JSON
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetBytes gets the raw bytes from an HTTP response
func (c *Client) GetBytes(endpoint string, options ...ClientOption) ([]byte, error) {
	return RequestAs[[]byte](c, "GET", endpoint, nil, options...)
}

// GetString gets the response as a string
func (c *Client) GetString(endpoint string, options ...ClientOption) (string, error) {
	return RequestAs[string](c, "GET", endpoint, nil, options...)
}

// buildURL combines the base URL with the endpoint and adds query parameters
func buildURL(baseURL, endpoint string, queryParams map[string]string) string {
	// If endpoint is already a full URL, use it directly
	if strings.HasPrefix(strings.ToLower(endpoint), "http") {
		return endpoint
	}

	// Ensure base URL doesn't end with slash if endpoint starts with slash
	if baseURL != "" {
		baseURL = strings.TrimSuffix(baseURL, "/")
		if !strings.HasPrefix(endpoint, "/") {
			baseURL += "/"
		}
		endpoint = baseURL + endpoint
	}

	// Parse the URL
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return endpoint // Return as is if it can't be parsed
	}

	// Get existing query values
	values := parsedURL.Query()

	// Add query parameters
	for key, value := range queryParams {
		values.Set(key, value)
	}

	// Update the URL with new query values
	parsedURL.RawQuery = values.Encode()
	return parsedURL.String()
}

// prepareRequestBody prepares the request body and returns the byte data and appropriate content type
func prepareRequestBody(body interface{}, contentType string) ([]byte, string, error) {
	if body == nil {
		return nil, contentType, nil
	}

	// Handle byte slice and string directly
	switch v := body.(type) {
	case []byte:
		return v, contentType, nil
	case string:
		return []byte(v), contentType, nil
	case url.Values:
		return []byte(v.Encode()), CONTENT_TYPE_FORM, nil
	}

	// Handle form data
	if contentType == CONTENT_TYPE_FORM {
		// Try to convert to a map for form encoding
		var formData map[string]interface{}

		switch v := body.(type) {
		case map[string]string:
			// Convert string map to interface map
			formData = make(map[string]interface{}, len(v))
			for key, value := range v {
				formData[key] = value
			}
		case map[string]interface{}:
			formData = v
		default:
			// Try to marshal to JSON and unmarshal to map
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, contentType, fmt.Errorf("failed to marshal data: %w", err)
			}

			if err := json.Unmarshal(jsonData, &formData); err != nil {
				return nil, contentType, fmt.Errorf("failed to convert to form data: %w", err)
			}
		}

		// Convert the map to URL values
		values := url.Values{}
		for key, value := range formData {
			values.Set(key, fmt.Sprintf("%v", value))
		}

		return []byte(values.Encode()), CONTENT_TYPE_FORM, nil
	}

	// Default to JSON
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, contentType, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, CONTENT_TYPE_JSON, nil
}

// applyAuth applies authentication to the request
func applyAuth(req *http.Request, config *ClientConfig) {
	// Apply basic auth
	if config.Username != "" && config.Password != "" {
		req.SetBasicAuth(config.Username, config.Password)
		return
	}

	// Apply token auth
	if config.Token != "" {
		tokenType := config.TokenType
		if tokenType == "" {
			tokenType = "Bearer"
		}
		req.Header.Set("Authorization", tokenType+" "+config.Token)
	}
}

// SetHeader sets or adds a header to the client's default headers
func (c *Client) SetHeader(key string, values ...string) *Client {
	if c.Config.Headers == nil {
		c.Config.Headers = make(map[string][]string)
	}
	c.Config.Headers[key] = values
	return c
}

// SetQueryParam sets a query parameter to the client's default parameters
func (c *Client) SetQueryParam(key, value string) *Client {
	if c.Config.QueryParams == nil {
		c.Config.QueryParams = make(map[string]string)
	}
	c.Config.QueryParams[key] = value
	return c
}

// SetBasicAuth sets basic authentication credentials
func (c *Client) SetBasicAuth(username, password string) *Client {
	c.Config.Username = username
	c.Config.Password = password
	return c
}

// SetBearerToken sets a bearer token for authentication
func (c *Client) SetBearerToken(token string) *Client {
	c.Config.Token = token
	c.Config.TokenType = "Bearer"
	return c
}

// SetContentType sets the content type for requests
func (c *Client) SetContentType(contentType string) *Client {
	c.Config.ContentType = contentType
	c.SetHeader("Content-Type", contentType)
	return c
}
