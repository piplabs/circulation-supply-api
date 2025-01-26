package http

import (
	"circulation-supply-api/service"

	"github.com/gin-gonic/gin"
)

func getSupply(c *gin.Context) {
	ret, err := service.GetSupply()
	if err != nil {
		ret = "syncing"
	}
	c.JSON(200, gin.H{
		"result": ret,
	})
}
