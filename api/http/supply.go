package http

import (
	"circulation-supply-api/metrics"
	"circulation-supply-api/service"
	log "log/slog"

	"github.com/gin-gonic/gin"
)

// Returns the circulating supply as a string with two decimal places.
func getCirculatingSupply(c *gin.Context) {
	var ret string
	circulatingSupply, err := service.GetCirculatingSupply()
	if err != nil {
		metrics.MetricApiErrorCount.Add(1)
		log.Error("Error getting circulating supply", "error", err)
		ret = "syncing"
	} else {
		ret = circulatingSupply.Text('f', 2)
	}
	c.JSON(200, gin.H{
		"result": ret,
	})
}

// Returns the total supply as a string with two decimal places.
func getTotalSupply(c *gin.Context) {
	var ret string
	totalSupply, err := service.GetTotalSupply()
	if err != nil {
		metrics.MetricApiErrorCount.Add(1)
		log.Error("Error getting total supply", "error", err)
		ret = "syncing"
	} else {
		ret = totalSupply.Text('f', 2)
	}
	c.JSON(200, gin.H{
		"result": ret,
	})
}

// Returns the circulating supply as an integer (only the whole number part, no decimals).
func getCirculatingSupplyWhole(c *gin.Context) {
	var ret int64
	circulatingSupply, err := service.GetCirculatingSupply()
	if err != nil {
		metrics.MetricApiErrorCount.Add(1)
		log.Error("Error getting circulating supply", "error", err)
	} else {
		ret, _ = circulatingSupply.Int64()
	}
	c.JSON(200, ret)
}

// Returns the total supply as an integer (only the whole number part, no decimals).
func getTotalSupplyWhole(c *gin.Context) {
	var ret int64
	totalSupply, err := service.GetTotalSupply()
	if err != nil {
		metrics.MetricApiErrorCount.Add(1)
		log.Error("Error getting total supply", "error", err)
	} else {
		ret, _ = totalSupply.Int64()
	}
	c.JSON(200, ret)
}
