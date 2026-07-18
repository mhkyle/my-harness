package route

import "github.com/gin-gonic/gin"

// Register registers all HTTP routes.
func Register(r *gin.Engine) {
	// Basic liveness probe.
	r.GET("/healthcheck", func(c *gin.Context) {
		c.String(200, "OK")
	})
}
