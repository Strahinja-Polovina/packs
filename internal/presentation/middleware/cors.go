package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds configuration for CORS middleware
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a CORS configuration that allows all origins and common headers
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Authorization",
			"Cache-Control",
			"Connection",
			"Cookie",
			"Host",
			"Pragma",
			"Referer",
			"User-Agent",
			"X-Requested-With",
			"X-CSRF-Token",
			"X-API-Key",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Content-Disposition",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a middleware that handles CORS with the provided configuration
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if len(config.AllowOrigins) > 0 {
			if contains(config.AllowOrigins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if contains(config.AllowOrigins, origin) {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", joinStrings(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders, ", "))
		}

		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders, ", "))
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", joinStrings([]string{intToString(config.MaxAge)}, ""))
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func CORSAllowAll() gin.HandlerFunc {
	return CORS(DefaultCORSConfig())
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func joinStrings(slice []string, separator string) string {
	if len(slice) == 0 {
		return ""
	}

	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += separator + slice[i]
	}
	return result
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
