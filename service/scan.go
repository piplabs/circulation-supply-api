package service

import (
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

	blockEventCh = make(chan *Block, 100)
)

// Start starts the main loop
func Start() {
	var err error
	currentBlock, totalBurntBaseFee, totalStakeReward, err = load()
	if err != nil {
		log.Warn("Error loading data", "error", err)
		return
	}

	go watchBlocks()
	handleBlock()

}

// watchBlocks fetches the latest block from the RPC endpoint and sends it to the blockEventCh channel
func watchBlocks() {
	for {
		block, err := getBlockByNumber(fmt.Sprintf("0x%x", currentBlock))
		if err != nil {
			log.Warn("Error fetching block", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}

		if block.Number == "" {
			time.Sleep(time.Second * 2)
			continue
		}
		blockEventCh <- block
	}
}

// handleBlock calculates the burnt IP and stake reward for each block and saves the accumulated fees every 100 blocks
func handleBlock() {
	for block := range blockEventCh {
		burntIP := calculateBurntIP(block.BaseFeePerGas, block.GasUsed)
		stakeReward := calculateStakeReward(block.Withdrawals)

		totalBurntBaseFee.Add(totalBurntBaseFee, burntIP)
		totalStakeReward.Add(totalStakeReward, stakeReward)

		if currentBlock > 0 && currentBlock%100 == 0 {
			err := saveAccumulatedFees(currentBlock, totalBurntBaseFee, totalStakeReward)
			if err != nil {
				log.Warn("Error saving data", "error", err)
			}
			log.Info("BlockNumber", "number", currentBlock, "burntBaseFee", burntIP.Text('f', -1), "totalBurntBaseFee", totalBurntBaseFee.Text('f', -1), "stakeReward", stakeReward.Text('f', -1), "totalStake", totalStakeReward.Text('f', -1))
		}
		currentBlock++
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
		if withdrawal.Validator != withdrawalTypeStakeReward && withdrawal.Address != withdrawalTypeUBI {
			continue
		}
		amount, _ := new(big.Int).SetString(withdrawal.Amount[2:], 16)
		totalReward.Add(totalReward, new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e9)))
	}
	return totalReward
}
