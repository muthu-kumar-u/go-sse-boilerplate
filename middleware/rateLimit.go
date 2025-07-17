package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nelsonin-research-org/clenz-stream/globals"
	"github.com/nelsonin-research-org/clenz-stream/message"
	appschema "github.com/nelsonin-research-org/clenz-stream/models"
)

// RateLimitMiddleware returns a Gin middleware function that rate limits based on IP and endpoint
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.Request.URL.Path
		now := time.Now()

		globals.RequestStore.Mu.Lock()
		defer globals.RequestStore.Mu.Unlock()

		if globals.RequestStore.Requests == nil {
			globals.RequestStore.Requests = make(map[string]map[string]*appschema.RateLimitEntry)
		}
		if globals.RequestStore.Requests[ip] == nil {
			globals.RequestStore.Requests[ip] = make(map[string]*appschema.RateLimitEntry)
		}

		entry, exists := globals.RequestStore.Requests[ip][endpoint]
		if !exists {
			globals.RequestStore.Requests[ip][endpoint] = &appschema.RateLimitEntry{
				Count:     1,
				Timestamp: now,
			}
		} else {
			if now.Sub(entry.Timestamp) > window {
				entry.Count = 1
				entry.Timestamp = now
			} else {
				if entry.Count >= limit {
					c.JSON(http.StatusTooManyRequests, message.ReturnMessage(http.StatusTooManyRequests))
					c.Abort()
					return
				}
				entry.Count++
			}
		}

		c.Next()
	}
}