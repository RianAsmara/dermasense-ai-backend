package middleware

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ResponseLogger logs detailed information about every outgoing response
// for debugging purposes. Includes status code, content-type, response body
// (truncated), and latency.
func ResponseLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		method := c.Method()
		path := c.Path()

		// Process the request
		err := c.Next()

		// Capture response details after processing
		latency := time.Since(start)
		status := c.Response().StatusCode()
		respContentType := string(c.Response().Header.ContentType())

		// Capture response body (limit to 2KB to avoid huge payloads flooding logs)
		var respBodySnippet string
		respBody := c.Response().Body()
		if len(respBody) > 0 {
			// Only log text-based responses (JSON, XML, HTML, plain text)
			if isTextContentType(respContentType) {
				if len(respBody) > 2048 {
					respBodySnippet = string(respBody[:2048]) + "... (truncated)"
				} else {
					respBodySnippet = string(respBody)
				}
			} else {
				respBodySnippet = "[binary data, " + formatBytes(len(respBody)) + "]"
			}
		}

		// Log response info
		log.Printf("─── RESPONSE ──────────────────────────────────────")
		log.Printf("  %s %s → %d", method, path, status)
		log.Printf("  Latency:      %v", latency)
		log.Printf("  Content-Type: %s", respContentType)
		if len(respBody) > 0 {
			log.Printf("  Size:         %s", formatBytes(len(respBody)))
		}
		if respBodySnippet != "" {
			log.Printf("  Body:         %s", respBodySnippet)
		}
		log.Printf("───────────────────────────────────────────────────")

		return err
	}
}

// isTextContentType checks if a content type is text-based and safe to log.
func isTextContentType(ct string) bool {
	ct = strings.ToLower(ct)
	textTypes := []string{
		"application/json",
		"application/xml",
		"text/",
		"application/javascript",
		"application/x-www-form-urlencoded",
	}
	for _, t := range textTypes {
		if strings.Contains(ct, t) {
			return true
		}
	}
	return false
}

// formatBytes returns a human-readable byte size string.
func formatBytes(b int) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(b)/(1024*1024))
	case b >= 1024:
		return fmt.Sprintf("%.2f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
