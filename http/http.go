package http

import (
	"github.com/gin-gonic/gin"
)

func StartHTTPServer() {
	r := gin.Default()
	r.GET("/circulating-supply", getSupply)
	r.Run()
}
