package main

import (
	"github.com/medatechnology/simplehttp"
)

// Phase 1 Features Examples

// 1. Path Parameters
func featureParamHandler(c simplehttp.Context) error {
	// String param
	id := c.GetParam("id")
	
	// Int param (with validation)
	age, err := c.GetParamInt("age")
	if err != nil {
		// Use new status helper
		return c.BadRequest("Invalid age parameter")
	}

	return c.JSON(200, map[string]interface{}{
		"id":  id,
		"age": age,
		"all": c.GetParams(),
	})
}

// 2. Cookie Handling
func featureCookieHandler(c simplehttp.Context) error {
	// Set cookie
	c.SetCookieSimple("demo_cookie", "hello-world", 3600)
	
	// Get cookie
	val, err := c.GetCookie("demo_cookie")
	status := "found"
	if err != nil {
		status = "not found (set now)"
	}

	return c.JSON(200, map[string]string{
		"cookie_status": status,
		"cookie_value":  val,
	})
}

// 3. Redirects
func featureRedirectHandler(c simplehttp.Context) error {
	return c.RedirectTemporary("/api/features/params/user123/25")
}

// 4. Form Values
func featureFormHandler(c simplehttp.Context) error {
	if c.GetMethod() == "POST" {
		name := c.GetFormValue("name")
		email := c.GetFormValue("email")
		
		return c.JSON(201, map[string]string{
			"status": "created",
			"name":   name,
			"email":  email,
		})
	}
	return c.BadRequest("Only POST allowed")
}

// 5. Status Helpers
func featureStatusHandler(c simplehttp.Context) error {
	action := c.GetParam("action")
	
	switch action {
	case "404":
		return c.NotFound()
	case "400":
		return c.BadRequest("This is a bad request")
	case "401":
		return c.Unauthorized("Login required")
	case "204":
		return c.NoContent()
	default:
		return c.JSON(200, map[string]string{"result": "ok"})
	}
}
