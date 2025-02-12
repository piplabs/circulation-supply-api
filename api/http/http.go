package http

import (
	"circulation-supply-api/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

func StartHTTPServer() {
	r := gin.Default()

	setupProbeRoutes(r)

	r.Use(requestCounterMiddleware(), latencyMiddleware())

	r.GET("/circulating-supply", getCirculatingSupply)
	r.GET("/total-supply", getTotalSupply)
	err := r.Run()
	if err != nil {
		panic(err)
	}
}

func requestCounterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.MetricApiCallCount.Add(1)
		c.Next()
	}
}

func latencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		metrics.MetricApiUsedTime.Add(float64(time.Since(start).Nanoseconds()))
	}
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

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "circulation-supply-api",
			"version": "0.1.0",
			"status":  "healthy",
		})
	})
}
