package http

import (
	"circulation-supply-api/dao"
	"circulation-supply-api/metrics"
	"circulation-supply-api/service"
	"errors"
	log "log/slog"
	"math/big"
	"net/http"
	"time"

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

func EstimateFutureCirculatingSupply(c *gin.Context) {
	date := c.Query("date")
	layout := "2006-01-02"
	t, err := time.ParseInLocation(layout, date, time.UTC)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid date format, must be YYYY-MM-DD",
		})
		return
	}

	timestamp := t.Unix()
	if timestamp < time.Now().UTC().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "date must be in the future",
		})
		return
	}
	circulatingSupply, err := service.EstimateFutureCirculatingSupply(timestamp)
	if err != nil {
		log.Error("Error estimating future circulating supply", "error", err)
		c.JSON(500, gin.H{
			"error": "internal server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"result": circulatingSupply.Text('f', 2),
	})
}

func getSupplyDelta(c *gin.Context) {
	startDate := c.Query("from")
	endDate := c.Query("to")
	layout := "2006-01-02"
	startTime, err := time.ParseInLocation(layout, startDate, time.UTC)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid start date format, must be YYYY-MM-DD",
		})
		return
	}
	endTime, err := time.ParseInLocation(layout, endDate, time.UTC)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid end date format, must be YYYY-MM-DD",
		})
		return
	}
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "end date must be after start date",
		})
		return
	}

	// TODO: remove this check when historical data is available
	// start time is yyyy-mm-dd, check if start time is in the current day, if in current day but before current timestamp, use start time. if start time is before current day, return error
	isCurDay := startTime.Year() == time.Now().UTC().Year() && startTime.YearDay() == time.Now().UTC().YearDay()
	if startTime.Before(time.Now().UTC()) && !isCurDay {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start date must not be in the past",
		})
		return
	}

	supplyDelta, err := service.GetSupplyDelta(startTime.Unix(), endTime.Unix())
	if err != nil {
		log.Error("Error getting supply delta", "error", err)
		c.JSON(500, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"result": supplyDelta,
	})
}

func getHistoryTotalSupply(c *gin.Context) {
	blockStr := c.Query("block")
	blockNumber, ok := new(big.Int).SetString(blockStr, 10)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid block number",
		})
		return
	}

	totalSupply, err := service.GetHistoryTotalSupply(blockNumber.Uint64())
	if err != nil {
		if errors.Is(err, dao.ErrRecordNotFound) {
			oldestBlock, latestBlock, err := service.GetHistoryRange()
			if err != nil {
				log.Error("Error getting history range", "error", err)
				c.JSON(500, gin.H{
					"error": "internal server error",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"result":         "syncing",
				"available_from": oldestBlock,
				"available_to":   latestBlock,
			})
			return
		}
		log.Error("Error getting history total supply", "error", err)
		c.JSON(500, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"result": totalSupply.Text('f', 2),
	})
}
