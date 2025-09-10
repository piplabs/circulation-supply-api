package service

import (
	"fmt"
	log "log/slog"
	"math/big"
	"sync"
	"time"
)

const (
	BlockReward = 1.929
)

var blockTimeLock = &sync.RWMutex{}
var mintPerSec float64

func StartBlockTimeCronjob() {
	CalculateAverageBlockTime()

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			_ = CalculateAverageBlockTime()
		}
	}()
}

func CalculateAverageBlockTime() error {
	latestBlock, err := getBlockByNumber("latest")
	if err != nil {
		log.Error("Error fetching latest block", "error", err)
		return err
	}
	latestBlockNumberStr := latestBlock.Number
	latestBlockNumber, _ := new(big.Int).SetString(latestBlockNumberStr, 0)
	latestBlockTimestamp, _ := new(big.Int).SetString(latestBlock.Time, 0)

	// Get the (latest-1000) block
	olderBlockNumber := new(big.Int).Sub(latestBlockNumber, big.NewInt(1000))
	olderBlock, err := getBlockByNumber(fmt.Sprintf("0x%x", olderBlockNumber.Int64()))
	if err != nil {
		log.Error("Error fetching older block", "error", err)
		return err
	}
	olderBlockTimestamp, _ := new(big.Int).SetString(olderBlock.Time, 0)

	// Calculate average block time
	if latestBlockNumber.Cmp(olderBlockNumber) <= 0 {
		log.Error("Invalid block numbers", "latest", latestBlockNumber, "older", olderBlockNumber)
		return err
	}
	if latestBlockTimestamp.Cmp(olderBlockTimestamp) <= 0 {
		log.Error("Invalid block timestamps", "latest", latestBlockTimestamp, "older", olderBlockTimestamp)
		return err
	}

	var averageBlockTime = new(big.Float).SetPrec(64)
	averageBlockTime.Sub(new(big.Float).SetInt(latestBlockTimestamp), new(big.Float).SetInt(olderBlockTimestamp))
	averageBlockTime.Quo(averageBlockTime, new(big.Float).SetInt(new(big.Int).Sub(latestBlockNumber, olderBlockNumber)))

	log.Info("latestBlockNumber", "number", latestBlockNumber, "timestamp", latestBlockTimestamp)
	log.Info("latestBlockNumber-1000", "number", olderBlockNumber, "timestamp", olderBlockTimestamp)
	log.Info("Average block time calculated", "average_seconds", averageBlockTime)

	blockTimeLock.Lock()
	defer blockTimeLock.Unlock()
	mintPerSec, _ = new(big.Float).SetPrec(64).Quo(new(big.Float).SetFloat64(BlockReward), averageBlockTime).Float64()
	log.Info("Mint per second updated", "mintPerSec", mintPerSec)

	return nil
}
