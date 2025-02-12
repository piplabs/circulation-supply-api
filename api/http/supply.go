package http

import (
	"circulation-supply-api/metrics"
	"circulation-supply-api/service"
	log "log/slog"

	"github.com/gin-gonic/gin"
)

func getCirculatingSupply(c *gin.Context) {
	ret, err := service.GetCirculatingSupply()
	if err != nil {
		metrics.MetricApiErrorCount.Add(1)
		log.Error("Error getting circulating supply", "error", err)
		ret = "syncing"
	}
	c.JSON(200, gin.H{
		"result": ret,
	})
}

func getTotalSupply(c *gin.Context) {
	ret, err := service.GetTotalSupply()
	if err != nil {
		metrics.MetricApiErrorCount.Add(1)
		log.Error("Error getting total supply", "error", err)
		ret = "syncing"
	}
	c.JSON(200, gin.H{
		"result": ret,
	})
}
