package http

import (
	"github.com/gin-gonic/gin"
)

func StartHTTPServer() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/supply", getSupply)
	r.Run()
}
