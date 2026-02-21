package middleware

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger logs detailed information about every incoming request
// for debugging purposes. Includes method, path, IP, headers, query
// params, request body, response status, and latency.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Capture request details before processing
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")
		contentType := c.Get("Content-Type")
		queryParams := c.Request().URI().QueryString()

		// Capture request body (limit to 2KB to avoid huge payloads flooding logs)
		var bodySnippet string
		body := c.Body()
		if len(body) > 0 {
			if len(body) > 2048 {
				bodySnippet = string(body[:2048]) + "... (truncated)"
			} else {
				bodySnippet = string(body)
			}
		}

		// Skip logging body for multipart (file uploads)
		if strings.HasPrefix(contentType, "multipart/form-data") {
			bodySnippet = fmt.Sprintf("[multipart form-data, %d bytes]", len(body))
		}

		// Log request info
		log.Printf("─── REQUEST ───────────────────────────────────────")
		log.Printf("  %s %s", method, path)
		log.Printf("  IP:           %s", ip)
		log.Printf("  User-Agent:   %s", userAgent)
		log.Printf("  Content-Type: %s", contentType)
		if len(queryParams) > 0 {
			log.Printf("  Query:        %s", string(queryParams))
		}
		if bodySnippet != "" {
			log.Printf("  Body:         %s", bodySnippet)
		}

		// Collect relevant request headers (skip common noise)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			// Mask the token, show only first 20 chars
			if len(authHeader) > 20 {
				authHeader = authHeader[:20] + "..."
			}
			log.Printf("  Authorization: %s", authHeader)
		}

		// Process the request
		err := c.Next()

		// Log response info
		latency := time.Since(start)
		status := c.Response().StatusCode()

		log.Printf("  ← %d | %v", status, latency)
		log.Printf("───────────────────────────────────────────────────")

		return err
	}
}
