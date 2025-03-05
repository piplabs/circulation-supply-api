package service

import (
	"circulation-supply-api/metrics"
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

// Start starts the main loop
func Start() {
	var err error
	currentBlock, totalBurntBaseFee, totalStakeReward, totalStakedToken, err = load()
	if err != nil {
		log.Warn("Error loading data", "error", err)
		return
	}
	metrics.CurrentlyIndexed.Set(float64(currentBlock))

	go BackwardsScanTrace()

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

// handleBlock calculates the burnt IP and stake reward for each block and saves the accumulated fees every 100 blocks
func handleBlock() {
	for block := range blockEventCh {
		burntIP := calculateBurntIP(block.BaseFeePerGas, block.GasUsed)
		stakeReward := calculateStakeReward(block.Withdrawals)

		trace := <-traceEventCh
		stakedToken := ProcessTraces(trace)

		totalBurntBaseFee.Add(totalBurntBaseFee, burntIP)
		totalStakeReward.Add(totalStakeReward, stakeReward)
		totalStakedToken.Add(totalStakedToken, stakedToken)

		if currentBlock > 0 && currentBlock%100 == 0 {
			err := saveAccumulatedFees(currentBlock, totalBurntBaseFee, totalStakeReward, totalStakedToken)
			if err != nil {
				log.Error("Error saving data", "error", err)
			}
			log.Info("[Forwards] BlockNumber", "number", currentBlock, "burntBaseFee", burntIP.Text('f', -1), "totalBurntBaseFee", totalBurntBaseFee.Text('f', -1), "stakeReward", stakeReward.Text('f', -1), "totalStake", totalStakeReward.Text('f', -1), "totalStakedToken", totalStakedToken.Text('f', -1))
		}
		metrics.CurrentlyIndexed.Set(float64(currentBlock))
	}
}

func calculateBurntIP(baseFeePerGasHex, gasUsedHex string) *big.Float {
	baseFeePerGas, _ := new(big.Int).SetString(baseFeePerGasHex[2:], 16)
	gasUsed, _ := new(big.Int).SetString(gasUsedHex[2:], 16)

	burntWei := new(big.Int).Mul(baseFeePerGas, gasUsed)
	burntIP := new(big.Float).Quo(new(big.Float).SetInt(burntWei), big.NewFloat(1e18)).SetPrec(64)
	return burntIP
}

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
