package service

import (
	"circulation-supply-api/config"
	"circulation-supply-api/dao"
	"circulation-supply-api/metrics"
	"errors"
	"fmt"
	log "log/slog"
	"math/big"
	"time"

	"gorm.io/gorm"
)

var (
	BackwardsBlockNumber uint64
	BackwardsStakedToken *big.Float
)

// Iterate from current block/last block to a configurable block number
func BackwardsScanTrace() {
	lastBlock, stakedToken, err := dao.GetBackwardsStakedToken()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(err)
		}
		if currentBlock >= 1 {
			BackwardsBlockNumber = currentBlock - 1
		}
		BackwardsStakedToken = big.NewFloat(0).SetPrec(64)
	} else {
		BackwardsBlockNumber = lastBlock - 1
		t, b := new(big.Float).SetString(stakedToken)
		if !b {
			panic("error parsing total staked token")
		}
		BackwardsStakedToken = t
	}

	metrics.BackwardsIndexed.Set(float64(BackwardsBlockNumber))

	for BackwardsBlockNumber > uint64(config.Conf.BackwardsStartBlock) {
		traces, err := FetchTraces(fmt.Sprintf("0x%x", BackwardsBlockNumber))
		if err != nil {
			log.Error("Failed to fetch traces", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}
		stakedToken := ProcessTraces(traces)
		BackwardsStakedToken.Add(BackwardsStakedToken, stakedToken)

		if BackwardsBlockNumber%10 == 0 || BackwardsBlockNumber == uint64(config.Conf.BackwardsStartBlock)+1 {
			log.Info("[Backwards] Saving staked token", "block", BackwardsBlockNumber, "stakedToken", BackwardsStakedToken.Text('f', -1))
			err = dao.SaveBackwardsStakedToken(BackwardsBlockNumber, BackwardsStakedToken.Text('f', -1))
			if err != nil {
				log.Error("Failed to save backwards staked token", "error", err)
				time.Sleep(time.Second * 5)
			}
		}
		BackwardsBlockNumber--
		metrics.BackwardsIndexed.Set(float64(BackwardsBlockNumber))
	}
}
