package service

import (
	"circulation-supply-api/dao"
	"circulation-supply-api/metrics"
	"errors"
	"fmt"
	"math/big"
	"time"

	log "log/slog"
)

const (
	withdrawalTypeStakeReward = "0x1"
	withdrawalTypeUBI         = "0x2"
)

var (
	currentBlock      uint64
	totalBurntBaseFee = big.NewFloat(0).SetPrec(64)
	totalStakeReward  = big.NewFloat(0).SetPrec(64)
	totalStakedToken  = big.NewFloat(0).SetPrec(64)

	blockEventCh = make(chan *Block, 100)
	traceEventCh = make(chan *Trace, 100)
)

var accumulatedFeesBatch = []dao.HistoryAccumulatedFees{}

var (
	firstBlockBackwards        int64 = 0
	totalBurntBaseFeeBackwards       = big.NewFloat(0).SetPrec(64)
	totalStakeRewardBackwards        = big.NewFloat(0).SetPrec(64)
	totalStakedTokenBackwards        = big.NewFloat(0).SetPrec(64)

	// should init from history_accumulated_fees while no records found in history_accumulated_fees
	initHistory = false
)

// Start starts the main loop
func Start() {
	var err error
	currentBlock, totalBurntBaseFee, totalStakeReward, totalStakedToken, err = load()
	if err != nil {
		log.Warn("Error loading data", "error", err)
		return
	}
	metrics.CurrentlyIndexed.Set(float64(currentBlock))

	oldestFees, err := GetOldestAccumulatedFees()
	if err != nil {
		if errors.Is(err, dao.ErrRecordNotFound) {
			initHistory = true
		} else {
			log.Warn("Error fetching oldest accumulated fees", "error", err)
			return
		}
	}
	if initHistory {
		firstBlockBackwards = int64(currentBlock) - 1
		totalBurntBaseFeeBackwards.Set(totalBurntBaseFee)
		totalStakeRewardBackwards.Set(totalStakeReward)
		totalStakedTokenBackwards.Set(totalStakedToken)
	} else {
		firstBlockBackwards = int64(oldestFees.BlockNumber)
		totalBurntBaseFeeBackwards, _ = new(big.Float).SetString(oldestFees.TotalBurntBaseFee)
		totalStakeRewardBackwards, _ = new(big.Float).SetString(oldestFees.TotalStakeReward)
		totalStakedTokenBackwards, _ = new(big.Float).SetString(oldestFees.TotalStakedToken)
	}

	// get burnt fees from first block of current round to a configurable block number backwards
	go startBackwardsScan(firstBlockBackwards)

	// forwards scan
	go watchBlocks()
	handleBlock()

}

// watchBlocks fetches the latest block from the RPC endpoint and sends it to the blockEventCh channel
func watchBlocks() {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for range ticker.C {
		block, err := getBlockByNumber(fmt.Sprintf("0x%x", currentBlock))
		if err != nil {
			log.Error("Error fetching block", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}
		// current block is not mined yet
		if block == nil {
			time.Sleep(time.Second * 5)
			continue
		}
		trace, err := FetchTraces(fmt.Sprintf("0x%x", currentBlock))
		if err != nil {
			log.Error("Error fetching traces", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}
		if trace == nil {
			time.Sleep(time.Second * 5)
			continue
		}
		blockEventCh <- block
		traceEventCh <- trace
		currentBlock++
	}
}

// handleBlock calculates the burnt IP and stake reward for each block
// and saves the accumulated fees every 100 blocks
func handleBlock() {
	for block := range blockEventCh {
		blockNumber, _ := new(big.Int).SetString(block.Number, 0)
		curBlock := blockNumber.Uint64()

		burntIP := calculateBurntIP(block.BaseFeePerGas, block.GasUsed)
		stakeReward := calculateStakeReward(block.Withdrawals)

		trace := <-traceEventCh
		stakedToken := ProcessTraces(trace)

		totalBurntBaseFee.Add(totalBurntBaseFee, burntIP)
		totalStakeReward.Add(totalStakeReward, stakeReward)
		totalStakedToken.Add(totalStakedToken, stakedToken)

		blockTimestamp, _ := new(big.Int).SetString(block.Time, 0)
		accumulatedFeesBatch = append(accumulatedFeesBatch, dao.HistoryAccumulatedFees{
			BlockNumber:       curBlock,
			TotalBurntBaseFee: totalBurntBaseFee.Text('f', -1),
			TotalStakeReward:  totalStakeReward.Text('f', -1),
			TotalStakedToken:  totalStakedToken.Text('f', -1),
			BlockTimestamp:    blockTimestamp.Uint64(),
		})

		if curBlock > 0 && curBlock%100 == 0 {
			err := saveAccumulatedFees(curBlock, totalBurntBaseFee, totalStakeReward, totalStakedToken)
			if err != nil {
				log.Error("Error saving data", "error", err)
			}
			log.Info("[Forwards] BlockNumber", "number", curBlock, "burntBaseFee", burntIP.Text('f', -1), "totalBurntBaseFee", totalBurntBaseFee.Text('f', -1), "stakeReward", stakeReward.Text('f', -1), "totalStake", totalStakeReward.Text('f', -1), "totalStakedToken", totalStakedToken.Text('f', -1))

			err = BatchAddHistoryAccumulatedFees(accumulatedFeesBatch)
			if err != nil {
				log.Error("Error batch saving accumulated fees", "error", err)
			}
			log.Info("[Forwards] batch saved accumulated fees up from block", "block", accumulatedFeesBatch[0].BlockNumber, "to", accumulatedFeesBatch[len(accumulatedFeesBatch)-1].BlockNumber)
			accumulatedFeesBatch = []dao.HistoryAccumulatedFees{}
		}
		metrics.CurrentlyIndexed.Set(float64(curBlock))

	}
}

func startBackwardsScan(startBlock int64) {
	curBlock := startBlock - 1
	batch := []dao.HistoryAccumulatedFees{}

	// Add first block to batch
	if initHistory {
		block, err := getBlockByNumber(fmt.Sprintf("0x%x", startBlock))
		if err != nil {
			log.Error("Error fetching block", "error", err)
			return
		}
		if block == nil {
			log.Error("First block for backwards scan is nil", "block", startBlock)
			return
		}
		blockTimestamp, _ := new(big.Int).SetString(block.Time, 0)
		batch = append(batch, dao.HistoryAccumulatedFees{
			BlockNumber:       uint64(startBlock),
			TotalBurntBaseFee: totalBurntBaseFeeBackwards.Text('f', -1),
			TotalStakeReward:  totalStakeRewardBackwards.Text('f', -1),
			TotalStakedToken:  totalStakedTokenBackwards.Text('f', -1),
			BlockTimestamp:    blockTimestamp.Uint64(),
		})
		initHistory = false
	}

	for curBlock > 0 {
		block, err := getBlockByNumber(fmt.Sprintf("0x%x", curBlock))
		if err != nil {
			log.Error("Error fetching block", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}
		if block == nil {
			time.Sleep(time.Second * 5)
			continue
		}
		burntIP := calculateBurntIP(block.BaseFeePerGas, block.GasUsed)

		trace, err := FetchTraces(fmt.Sprintf("0x%x", curBlock))
		if err != nil {
			log.Error("Error fetching traces", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}
		if trace == nil {
			time.Sleep(time.Second * 5)
			continue
		}
		stakedToken := ProcessTraces(trace)

		totalBurntBaseFeeBackwards.Sub(totalBurntBaseFeeBackwards, burntIP)
		totalStakeRewardBackwards.Sub(totalStakeRewardBackwards, calculateStakeReward(block.Withdrawals))
		totalStakedTokenBackwards.Sub(totalStakedTokenBackwards, stakedToken)

		blockTimestamp, _ := new(big.Int).SetString(block.Time, 0)
		batch = append(batch, dao.HistoryAccumulatedFees{
			BlockNumber:       uint64(curBlock),
			TotalBurntBaseFee: totalBurntBaseFeeBackwards.Text('f', -1),
			TotalStakeReward:  totalStakeRewardBackwards.Text('f', -1),
			TotalStakedToken:  totalStakedTokenBackwards.Text('f', -1),
			BlockTimestamp:    blockTimestamp.Uint64(),
		})
		curBlock--

		if len(batch) >= 100 || curBlock == 0 {
			err = BatchAddHistoryAccumulatedFees(batch)
			if err != nil {
				log.Error("Error batch saving accumulated fees backwards", "error", err)
				break
			}
			log.Info("[Backwards] batch saved accumulated fees up from block", "block", batch[0].BlockNumber, "to", batch[len(batch)-1].BlockNumber)
			batch = []dao.HistoryAccumulatedFees{}
		}
	}
}

// Burnt IP is calculated by multiplying the base fee per gas by the gas used in the block.
func calculateBurntIP(baseFeePerGasHex, gasUsedHex string) *big.Float {
	baseFeePerGas, _ := new(big.Int).SetString(baseFeePerGasHex[2:], 16)
	gasUsed, _ := new(big.Int).SetString(gasUsedHex[2:], 16)

	burntWei := new(big.Int).Mul(baseFeePerGas, gasUsed)
	burntIP := new(big.Float).Quo(new(big.Float).SetInt(burntWei), big.NewFloat(1e18)).SetPrec(64)
	return burntIP
}

// Stake reward is calculated by summing up the values of specific types of withdrawals
func calculateStakeReward(withdrawals []Withdrawal) *big.Float {
	totalReward := big.NewFloat(0).SetPrec(64)
	for _, withdrawal := range withdrawals {
		if withdrawal.Validator != withdrawalTypeStakeReward &&
			withdrawal.Validator != withdrawalTypeUBI {
			continue
		}
		amount, _ := new(big.Int).SetString(withdrawal.Amount[2:], 16)
		totalReward.Add(totalReward, new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e9)))
	}
	return totalReward
}
