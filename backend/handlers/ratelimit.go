package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	rateMu    sync.Mutex
	rateStore = map[string][]time.Time{}
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateMu.Lock()
		now := time.Now()
		window := now.Add(-15 * time.Minute)

		// Clean old entries
		var valid []time.Time
		for _, t := range rateStore[ip] {
			if t.After(window) {
				valid = append(valid, t)
			}
		}

		if len(valid) >= 5 {
			rateMu.Unlock()
			oldest := valid[0]
			retryAfter := int(oldest.Add(15*time.Minute).Sub(now).Seconds()) + 1
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts, try again later"})
			return
		}

		rateStore[ip] = append(valid, now)
		rateMu.Unlock()
		c.Next()
	}
}
