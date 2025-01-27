package http

import (
	"github.com/gin-gonic/gin"
)

func StartHTTPServer() {
	r := gin.Default()

	setupProbeRoutes(r)

	r.GET("/circulating-supply", getSupply)
	r.Run()
}

func setupProbeRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "circulation-supply",
		})
	})

	r.GET("/ready", func(c *gin.Context) {
    // TODO: check DB connection
		c.JSON(200, gin.H{
			"ready":   true,
			"service": "circulation-supply",
		})
	})
}
