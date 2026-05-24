package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WriteFreeze returns a middleware that rejects mutating requests (POST/PUT/DELETE)
// with 503 when write freeze is enabled. Used during DB migration cutover.
func WriteFreeze(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}
		method := c.Request.Method
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "service is in maintenance mode, please retry later",
			})
			return
		}
		c.Next()
	}
}
