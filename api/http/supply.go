package http

import (
	"circulation-supply-api/service"
	log "log/slog"

	"github.com/gin-gonic/gin"
)

func getSupply(c *gin.Context) {
	ret, err := service.GetSupply()
	if err != nil {
		log.Error("Error getting supply", "error", err)
		ret = "syncing"
	}
	c.JSON(200, gin.H{
		"result": ret,
	})
}
