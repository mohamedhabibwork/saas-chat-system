package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper authentication
		// For now, we'll just check for a basic auth header
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// TODO: Validate credentials against database
		// For now, we'll just accept any non-empty credentials
		if username == "" || password == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			c.Abort()
			return
		}

		// Set user info in context for later use
		c.Set("username", username)
		c.Next()
	}
} 